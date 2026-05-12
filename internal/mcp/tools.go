// Copyright (c) 2026 John Dewey

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER
// DEALINGS IN THE SOFTWARE.

package mcp

import (
	"context"
	"fmt"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/retr0h/mlb-sdk/pkg/mlb"
)

// registerTools wires every MLB MCP tool onto s.mcp.
func (s *Server) registerTools() {
	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "schedule",
		Description: "Fetch MLB game schedule. Filter by team ID, a single date (YYYY-MM-DD), or a date range. All filters are optional; omitting them returns the full league schedule for today.",
	}, s.toolSchedule)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "standings",
		Description: "Fetch division standings for a league. league_id is required: 103 = American League, 104 = National League. Optionally filter by season year, standings type (e.g. regularSeason, wildCard), or a specific date.",
	}, s.toolStandings)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "person",
		Description: "Fetch a single player or person by their MLB person ID (e.g. 660271 = Shohei Ohtani). Returns biographical info, primary position, bat side, and pitch hand.",
	}, s.toolPerson)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "team",
		Description: "Fetch team metadata by MLB team ID (e.g. 119 = Los Angeles Dodgers). Includes venue, league, division, and sport. Pass hydrate='league,division,sport,venue' to expand sub-objects beyond id/name/link.",
	}, s.toolTeam)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "stats_leaders",
		Description: "Fetch league stat leaders. leader_categories is required (e.g. 'homeRuns', 'battingAverage', 'strikeOuts'). Optionally filter by season, sport_id (1=MLB), league_id, stat_group, player_pool, and limit.",
	}, s.toolStatsLeaders)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name:        "linescore",
		Description: "Fetch the inning-by-inning linescore for a game by its gamePk. Returns per-inning runs/hits/errors for both teams, game totals, current count (balls/strikes/outs), and the active defense and offense lineups.",
	}, s.toolLinescore)
}

// --- args structs ---

// All date fields are YYYY-MM-DD strings; the handler converts them to
// time.Time before forwarding to the SDK so the JSON schema stays
// serialisable.

type scheduleArgs struct {
	TeamID string `json:"team_id,omitempty" jsonschema:"MLB team ID (integer), e.g. 119 for the Dodgers. Omit for all teams."`
	On     string `json:"on,omitempty"      jsonschema:"Single date filter in YYYY-MM-DD format. Mutually exclusive with from/to."`
	From   string `json:"from,omitempty"    jsonschema:"Start of date range in YYYY-MM-DD format. Requires 'to'."`
	To     string `json:"to,omitempty"      jsonschema:"End of date range in YYYY-MM-DD format. Requires 'from'."`
}

type standingsArgs struct {
	LeagueID       int    `json:"league_id"                 jsonschema:"MLB league ID (required): 103 = American League, 104 = National League."`
	Season         int    `json:"season,omitempty"          jsonschema:"Season year, e.g. 2026. Defaults to the current season."`
	StandingsTypes string `json:"standings_types,omitempty" jsonschema:"Standings type: regularSeason, wildCard, divisionLeaders, etc."`
	On             string `json:"on,omitempty"              jsonschema:"View standings as of this date (YYYY-MM-DD). Defaults to today."`
	Hydrate        string `json:"hydrate,omitempty"         jsonschema:"Comma-separated hydrate string to expand sub-objects."`
}

type personArgs struct {
	PersonID int    `json:"person_id"         jsonschema:"MLB person ID (required), e.g. 660271 for Shohei Ohtani."`
	Hydrate  string `json:"hydrate,omitempty" jsonschema:"Comma-separated hydrate string, e.g. 'stats'."`
	Fields   string `json:"fields,omitempty"  jsonschema:"Comma-separated field projection to restrict the response."`
}

type teamArgs struct {
	TeamID  int    `json:"team_id"           jsonschema:"MLB team ID (required), e.g. 119 for the Los Angeles Dodgers."`
	Season  int    `json:"season,omitempty"  jsonschema:"Season year, e.g. 2026. Constrains metadata to that season."`
	Hydrate string `json:"hydrate,omitempty" jsonschema:"Comma-separated hydrate string, e.g. 'league,division,sport,venue'."`
	Fields  string `json:"fields,omitempty"  jsonschema:"Comma-separated field projection to restrict the response."`
}

type statsLeadersArgs struct {
	LeaderCategories string `json:"leader_categories"           jsonschema:"Stat category (required), e.g. 'homeRuns', 'battingAverage', 'strikeOuts'."`
	Season           int    `json:"season,omitempty"            jsonschema:"Season year, e.g. 2026. Defaults to current season."`
	SportID          int    `json:"sport_id,omitempty"          jsonschema:"Sport ID: 1 = MLB. Defaults to MLB."`
	LeagueID         int    `json:"league_id,omitempty"         jsonschema:"MLB league ID: 103 = AL, 104 = NL. Omit for both leagues."`
	StatGroup        string `json:"stat_group,omitempty"        jsonschema:"Stat group: hitting, pitching, fielding, catching."`
	PlayerPool       string `json:"player_pool,omitempty"       jsonschema:"Player pool: All, Qualified, Rookies."`
	LeaderGameTypes  string `json:"leader_game_types,omitempty" jsonschema:"Game type filter, e.g. 'R' for regular season."`
	StatType         string `json:"stat_type,omitempty"         jsonschema:"Stat type override."`
	Hydrate          string `json:"hydrate,omitempty"           jsonschema:"Comma-separated hydrate string."`
	Limit            int    `json:"limit,omitempty"             jsonschema:"Maximum number of leaders to return per category. Defaults to API default."`
	Fields           string `json:"fields,omitempty"            jsonschema:"Comma-separated field projection to restrict the response."`
}

type linescoreArgs struct {
	GamePk   int    `json:"game_pk"            jsonschema:"MLB game PK (required). Obtain from the schedule tool."`
	Timecode string `json:"timecode,omitempty" jsonschema:"Point-in-time timecode YYYYMMDD_HHmmss for historical linescores."`
	Fields   string `json:"fields,omitempty"   jsonschema:"Comma-separated field projection to restrict the response."`
}

// --- tool handlers ---

func (s *Server) toolSchedule(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args scheduleArgs,
) (*mcpsdk.CallToolResult, any, error) {
	q := mlb.ScheduleQuery{}
	if args.TeamID != "" {
		var id int
		if _, err := fmt.Sscanf(args.TeamID, "%d", &id); err != nil {
			return nil, nil, fmt.Errorf("mlb-mcp: schedule: team_id must be an integer: %w", err)
		}
		q.Team = mlb.TeamID(id)
	}
	if args.On != "" {
		t, err := time.Parse("2006-01-02", args.On)
		if err != nil {
			return nil, nil, fmt.Errorf("mlb-mcp: schedule: on must be YYYY-MM-DD: %w", err)
		}
		q.On = t
	}
	if args.From != "" && args.To != "" {
		from, err := time.Parse("2006-01-02", args.From)
		if err != nil {
			return nil, nil, fmt.Errorf("mlb-mcp: schedule: from must be YYYY-MM-DD: %w", err)
		}
		to, err := time.Parse("2006-01-02", args.To)
		if err != nil {
			return nil, nil, fmt.Errorf("mlb-mcp: schedule: to must be YYYY-MM-DD: %w", err)
		}
		q.From = from
		q.To = to
	}

	games, err := s.client.Schedule(ctx, q)
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: schedule: %w", err)
	}
	return textResult(jsonOrErr(games)), nil, nil
}

func (s *Server) toolStandings(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args standingsArgs,
) (*mcpsdk.CallToolResult, any, error) {
	q := mlb.StandingsQuery{
		League:         mlb.LeagueID(args.LeagueID),
		Season:         args.Season,
		StandingsTypes: args.StandingsTypes,
		Hydrate:        args.Hydrate,
	}
	if args.On != "" {
		t, err := time.Parse("2006-01-02", args.On)
		if err != nil {
			return nil, nil, fmt.Errorf("mlb-mcp: standings: on must be YYYY-MM-DD: %w", err)
		}
		q.On = t
	}

	st, err := s.client.Standings(ctx, q)
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: standings: %w", err)
	}
	return textResult(jsonOrErr(st)), nil, nil
}

func (s *Server) toolPerson(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args personArgs,
) (*mcpsdk.CallToolResult, any, error) {
	q := mlb.PersonQuery{
		Hydrate: args.Hydrate,
		Fields:  args.Fields,
	}
	p, err := s.client.Person(ctx, args.PersonID, q)
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: person: %w", err)
	}
	return textResult(jsonOrErr(p)), nil, nil
}

func (s *Server) toolTeam(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args teamArgs,
) (*mcpsdk.CallToolResult, any, error) {
	q := mlb.TeamQuery{
		Season:  args.Season,
		Hydrate: args.Hydrate,
		Fields:  args.Fields,
	}
	t, err := s.client.Team(ctx, args.TeamID, q)
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: team: %w", err)
	}
	return textResult(jsonOrErr(t)), nil, nil
}

func (s *Server) toolStatsLeaders(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args statsLeadersArgs,
) (*mcpsdk.CallToolResult, any, error) {
	q := mlb.StatsLeadersQuery{
		LeaderCategories: args.LeaderCategories,
		Season:           args.Season,
		SportID:          args.SportID,
		LeagueID:         args.LeagueID,
		StatGroup:        args.StatGroup,
		PlayerPool:       args.PlayerPool,
		LeaderGameTypes:  args.LeaderGameTypes,
		StatType:         args.StatType,
		Hydrate:          args.Hydrate,
		Limit:            args.Limit,
		Fields:           args.Fields,
	}
	sl, err := s.client.StatsLeaders(ctx, q)
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: stats_leaders: %w", err)
	}
	return textResult(jsonOrErr(sl)), nil, nil
}

func (s *Server) toolLinescore(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args linescoreArgs,
) (*mcpsdk.CallToolResult, any, error) {
	q := mlb.LinescoreQuery{
		Timecode: args.Timecode,
		Fields:   args.Fields,
	}
	ls, err := s.client.Linescore(ctx, args.GamePk, q)
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: linescore: %w", err)
	}
	return textResult(jsonOrErr(ls)), nil, nil
}

// --- helpers ---

// textResult wraps a string as an MCP CallToolResult with a single
// TextContent block — the canonical response shape for every JSON-returning
// tool.
func textResult(text string) *mcpsdk.CallToolResult {
	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{
			&mcpsdk.TextContent{Text: text},
		},
	}
}

// jsonOrErr renders a value as pretty JSON for an MCP TextContent response,
// or returns an mcp-typed error string if marshaling fails.
func jsonOrErr(v any) string {
	b, err := jsonMarshalIndent(v)
	if err != nil {
		return "error: marshal response: " + err.Error()
	}
	return string(b)
}

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
		Name: "today_scores",
		Description: "Use this when the user asks who won today, what are today's scores, " +
			"or what games are being played today. Returns every MLB game scheduled for today " +
			"with team names, scores, and game status (Final/Live/Scheduled).",
	}, s.toolTodayScores)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name: "standings",
		Description: "Use this when the user asks about standings, division standings, or " +
			"how teams are doing in the standings. Fetches AL (103) and/or NL (104) standings " +
			"with wins, losses, winning percentage, and games back. Optionally filter by " +
			"league ('AL' or 'NL') and season year.",
	}, s.toolStandings)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name: "player_bio",
		Description: "Use this when the user asks about a specific player — 'tell me about " +
			"Ohtani', 'who is Mike Trout', 'what position does Freddie Freeman play'. " +
			"Returns full biographical info including position, bat side, pitch hand, " +
			"height, weight, birth date, and active status. Requires the player's MLB person ID.",
	}, s.toolPlayerBio)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name: "team_info",
		Description: "Use this when the user asks about a specific team — 'tell me about the " +
			"Dodgers', 'where do the Yankees play', 'what league is Houston in'. Returns rich " +
			"team metadata including venue, league, division, abbreviation, and founding year. " +
			"Requires the team's MLB team ID (e.g. 119 = Los Angeles Dodgers).",
	}, s.toolTeamInfo)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name: "team_roster",
		Description: "Use this when the user asks who is on a team — 'who's on the Dodgers', " +
			"'show me the Red Sox roster', 'list the Yankees players'. Returns the active " +
			"roster with player names, jersey numbers, and positions. Requires the team's " +
			"MLB team ID. Season defaults to the current year.",
	}, s.toolTeamRoster)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name: "league_leaders",
		Description: "Use this when the user asks about stat leaders — 'who leads in home runs', " +
			"'who has the best batting average', 'top ERA pitchers'. Requires a stat category " +
			"(e.g. 'homeRuns', 'battingAverage', 'strikeOuts', 'era'). Season defaults to the " +
			"current year. Limit defaults to 10.",
	}, s.toolLeagueLeaders)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name: "game_detail",
		Description: "Use this when the user asks what happened in a specific game — 'what " +
			"were the stats in game 745455', 'show me the boxscore'. Returns team batting and " +
			"pitching stats for both home and away teams. Requires the game's MLB gamePk " +
			"(obtainable from today_scores or postseason_schedule).",
	}, s.toolGameDetail)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name: "game_linescore",
		Description: "Use this when the user asks for an inning-by-inning breakdown — 'show " +
			"me the linescore for game X', 'what happened each inning'. Returns per-inning " +
			"runs/hits/errors for both teams, game totals, and current count. Requires the " +
			"game's MLB gamePk.",
	}, s.toolGameLinescore)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name: "recent_transactions",
		Description: "Use this when the user asks about recent roster moves — 'any trades " +
			"today', 'what moves happened this week', 'recent DFA transactions'. Returns " +
			"signings, trades, designations, and other transactions. Days defaults to 1 " +
			"(today). Optionally filter to a specific team by MLB team ID.",
	}, s.toolRecentTransactions)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name: "free_agents",
		Description: "Use this when the user asks about free agents — 'who are the free " +
			"agents', 'which players are unsigned', 'show me free agent signings'. Returns " +
			"players who declared free agency with their original team, new team (if signed), " +
			"and signing date. Season defaults to the current year.",
	}, s.toolFreeAgents)

	mcpsdk.AddTool(s.mcp, &mcpsdk.Tool{
		Name: "postseason_schedule",
		Description: "Use this when the user asks about the playoffs or postseason — 'what's " +
			"the postseason schedule', 'when are the playoffs', 'show me the World Series games'. " +
			"Returns all postseason games with teams, scores, and status. Season defaults to " +
			"the current year.",
	}, s.toolPostseasonSchedule)
}

// --- args structs ---

// todayScoresArgs has no fields — today_scores always operates on today.
type todayScoresArgs struct{}

type standingsArgs struct {
	League string `json:"league,omitempty" jsonschema:"Filter to a single league: 'AL' or 'NL'. Omit for both leagues."`
	Season int    `json:"season,omitempty" jsonschema:"Season year, e.g. 2026. Defaults to the current season."`
}

type playerBioArgs struct {
	PersonID int `json:"person_id" jsonschema:"MLB person ID (required), e.g. 660271 for Shohei Ohtani."`
}

type teamInfoArgs struct {
	TeamID int `json:"team_id" jsonschema:"MLB team ID (required), e.g. 119 for the Los Angeles Dodgers."`
}

type teamRosterArgs struct {
	TeamID int `json:"team_id"          jsonschema:"MLB team ID (required), e.g. 119 for the Los Angeles Dodgers."`
	Season int `json:"season,omitempty" jsonschema:"Season year, e.g. 2026. Defaults to the current year."`
}

type leagueLeadersArgs struct {
	Category string `json:"category"         jsonschema:"Stat category (required), e.g. 'homeRuns', 'battingAverage', 'strikeOuts', 'era'."`
	Season   int    `json:"season,omitempty" jsonschema:"Season year, e.g. 2026. Defaults to the current season."`
	Limit    int    `json:"limit,omitempty"  jsonschema:"Number of leaders to return. Defaults to 10."`
}

type gameDetailArgs struct {
	GamePk int `json:"game_pk" jsonschema:"MLB game PK (required). Obtain from today_scores or postseason_schedule."`
}

type gameLinescoreArgs struct {
	GamePk int `json:"game_pk" jsonschema:"MLB game PK (required). Obtain from today_scores or postseason_schedule."`
}

type recentTransactionsArgs struct {
	Days   int `json:"days,omitempty"    jsonschema:"Number of days to look back (and forward) from today. Defaults to 1."`
	TeamID int `json:"team_id,omitempty" jsonschema:"Filter to a single MLB team ID, e.g. 119 for the Dodgers. Omit for all teams."`
}

type freeAgentsArgs struct {
	Season int `json:"season,omitempty" jsonschema:"Season year, e.g. 2026. Defaults to the current year."`
}

type postseasonScheduleArgs struct {
	Season int `json:"season,omitempty" jsonschema:"Season year, e.g. 2026. Defaults to the current year."`
}

// --- tool handlers ---

func (s *Server) toolTodayScores(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	_ todayScoresArgs,
) (*mcpsdk.CallToolResult, any, error) {
	today := time.Now().UTC().Truncate(24 * time.Hour)
	games, err := s.client.Schedule(ctx, mlb.ScheduleQuery{
		On: today,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: today_scores: %w", err)
	}
	return textResult(jsonOrErr(games)), nil, nil
}

func (s *Server) toolStandings(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args standingsArgs,
) (*mcpsdk.CallToolResult, any, error) {
	type leagueFetch struct {
		id   mlb.LeagueID
		name string
	}

	fetches := []leagueFetch{
		{id: 103, name: "AL"},
		{id: 104, name: "NL"},
	}

	switch args.League {
	case "AL":
		fetches = []leagueFetch{{id: 103, name: "AL"}}
	case "NL":
		fetches = []leagueFetch{{id: 104, name: "NL"}}
	}

	type leagueStandings struct {
		League    string         `json:"league"`
		Standings *mlb.Standings `json:"standings"`
	}

	results := make([]leagueStandings, 0, len(fetches))
	for _, f := range fetches {
		st, err := s.client.Standings(ctx, mlb.StandingsQuery{
			League: f.id,
			Season: args.Season,
		})
		if err != nil {
			return nil, nil, fmt.Errorf("mlb-mcp: standings: %w", err)
		}
		results = append(results, leagueStandings{League: f.name, Standings: st})
	}
	return textResult(jsonOrErr(results)), nil, nil
}

func (s *Server) toolPlayerBio(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args playerBioArgs,
) (*mcpsdk.CallToolResult, any, error) {
	p, err := s.client.Person(ctx, args.PersonID, mlb.PersonQuery{})
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: player_bio: %w", err)
	}
	return textResult(jsonOrErr(p)), nil, nil
}

func (s *Server) toolTeamInfo(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args teamInfoArgs,
) (*mcpsdk.CallToolResult, any, error) {
	t, err := s.client.Team(ctx, args.TeamID, mlb.TeamQuery{
		Hydrate: "league,division,sport,venue",
	})
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: team_info: %w", err)
	}
	return textResult(jsonOrErr(t)), nil, nil
}

func (s *Server) toolTeamRoster(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args teamRosterArgs,
) (*mcpsdk.CallToolResult, any, error) {
	season := args.Season
	if season == 0 {
		season = time.Now().Year()
	}
	r, err := s.client.Roster(ctx, args.TeamID, mlb.RosterQuery{
		RosterType: "active",
		Season:     season,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: team_roster: %w", err)
	}
	return textResult(jsonOrErr(r)), nil, nil
}

func (s *Server) toolLeagueLeaders(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args leagueLeadersArgs,
) (*mcpsdk.CallToolResult, any, error) {
	limit := args.Limit
	if limit == 0 {
		limit = 10
	}
	sl, err := s.client.StatsLeaders(ctx, mlb.StatsLeadersQuery{
		LeaderCategories: args.Category,
		Season:           args.Season,
		SportID:          1,
		Limit:            limit,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: league_leaders: %w", err)
	}
	return textResult(jsonOrErr(sl)), nil, nil
}

func (s *Server) toolGameDetail(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args gameDetailArgs,
) (*mcpsdk.CallToolResult, any, error) {
	bs, err := s.client.Boxscore(ctx, args.GamePk)
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: game_detail: %w", err)
	}
	return textResult(jsonOrErr(bs)), nil, nil
}

func (s *Server) toolGameLinescore(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args gameLinescoreArgs,
) (*mcpsdk.CallToolResult, any, error) {
	ls, err := s.client.Linescore(ctx, args.GamePk, mlb.LinescoreQuery{})
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: game_linescore: %w", err)
	}
	return textResult(jsonOrErr(ls)), nil, nil
}

func (s *Server) toolRecentTransactions(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args recentTransactionsArgs,
) (*mcpsdk.CallToolResult, any, error) {
	days := args.Days
	if days == 0 {
		days = 1
	}

	now := time.Now().UTC().Truncate(24 * time.Hour)
	start := now.AddDate(0, 0, -(days - 1))
	end := now

	q := mlb.TransactionsQuery{
		StartDate: start,
		EndDate:   end,
	}
	if args.TeamID != 0 {
		// When a team is specified, use TeamID alone (valid per SDK validation).
		q = mlb.TransactionsQuery{
			TeamID: args.TeamID,
		}
	}

	tx, err := s.client.Transactions(ctx, q)
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: recent_transactions: %w", err)
	}
	return textResult(jsonOrErr(tx)), nil, nil
}

func (s *Server) toolFreeAgents(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args freeAgentsArgs,
) (*mcpsdk.CallToolResult, any, error) {
	season := args.Season
	if season == 0 {
		season = time.Now().Year()
	}
	fa, err := s.client.FreeAgents(ctx, mlb.FreeAgentsQuery{
		Season: season,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: free_agents: %w", err)
	}
	return textResult(jsonOrErr(fa)), nil, nil
}

func (s *Server) toolPostseasonSchedule(
	ctx context.Context,
	_ *mcpsdk.CallToolRequest,
	args postseasonScheduleArgs,
) (*mcpsdk.CallToolResult, any, error) {
	season := args.Season
	if season == 0 {
		season = time.Now().Year()
	}
	games, err := s.client.SchedulePostseason(ctx, mlb.SchedulePostseasonQuery{
		Season: season,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("mlb-mcp: postseason_schedule: %w", err)
	}
	return textResult(jsonOrErr(games)), nil, nil
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

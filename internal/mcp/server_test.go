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

// In-process round-trip tests for the MCP server. Each row drives the server
// through the SDK's InMemoryTransport pair — no stdio fork, no network — and
// asserts on the tool-call result.
//
// The mcpServerHarness builder wires:
//
//	fakeDriver (configurable per test row)
//	  ↓ Driver interface
//	*Server (real tool registration via New)
//	  ↓ MCP-over-pipe
//	*mcpsdk.ClientSession (SDK test client)
//
// so every assertion exercises the full request → handler → textResult path.

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/retr0h/mlb-sdk/pkg/mlb"
)

// mcpServerHarness spins up a *Server with the supplied driver, pairs it with
// an in-memory transport, runs the server in a background goroutine, and
// connects a test client. Returns the connected *mcpsdk.ClientSession.
func mcpServerHarness(t *testing.T, driver Driver) *mcpsdk.ClientSession {
	t.Helper()

	srv := New(Config{Driver: driver})

	clientT, serverT := mcpsdk.NewInMemoryTransports()

	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)

	// Run the MCP server in a goroutine. cancel (registered first, fires last)
	// terminates Run via ctx; cs.Close (registered last, fires first) tears
	// down the client session first so the server unblocks cleanly.
	go func() { _ = srv.mcp.Run(ctx, serverT) }()

	client := mcpsdk.NewClient(
		&mcpsdk.Implementation{Name: "test-client"},
		nil,
	)
	cs, err := client.Connect(ctx, clientT, nil)
	if err != nil {
		t.Fatalf("connect mcp client: %v", err)
	}
	t.Cleanup(func() { _ = cs.Close() })
	return cs
}

// TestNew_NilDriver verifies that New creates a real mlb.Client when
// Config.Driver is nil — covers the nil-guard branch in New.
func TestNew_NilDriver(t *testing.T) {
	s := New(Config{}) // Driver is nil → must not panic
	if s == nil {
		t.Fatal("New returned nil")
	}
	if s.client == nil {
		t.Fatal("New did not set a default client")
	}
}

// TestJsonOrErr_MarshalError covers the error branch in jsonOrErr — a value
// that json.MarshalIndent cannot encode (a channel) must produce the error
// string prefix rather than panicking.
func TestJsonOrErr_MarshalError(t *testing.T) {
	got := jsonOrErr(make(chan int)) // channels are not JSON-serialisable
	if !strings.HasPrefix(got, "error: marshal response:") {
		t.Errorf(
			"jsonOrErr with unmarshalable value = %q, want prefix %q",
			got,
			"error: marshal response:",
		)
	}
}

// TestServer_ListTools verifies every expected tool name is registered and
// that no extra tools have been silently added. Treat the catalog as a frozen
// surface — accidental removal or addition trips this test.
//
// The server exposes two tiers:
//   - Tier 1: 11 hand-written composed tools (friendly names).
//   - Tier 2: 54 auto-generated raw spec tools (mlb_* prefix).
func TestServer_ListTools(t *testing.T) {
	cs := mcpServerHarness(t, &fakeDriver{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := cs.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}

	got := map[string]bool{}
	for _, tool := range resp.Tools {
		got[tool.Name] = true
	}

	// Tier 2 — raw spec tools auto-generated from the OpenAPI spec.
	rawTools := []string{
		"mlb_get_all_seasons",
		"mlb_get_all_star_ballot",
		"mlb_get_all_star_final_vote",
		"mlb_get_all_star_write_ins",
		"mlb_get_attendance",
		"mlb_get_award_recipients",
		"mlb_get_conferences",
		"mlb_get_context_metrics",
		"mlb_get_divisions",
		"mlb_get_game_changes",
		"mlb_get_game_color",
		"mlb_get_game_color_diff",
		"mlb_get_game_color_timestamps",
		"mlb_get_game_content",
		"mlb_get_game_diff",
		"mlb_get_game_pace",
		"mlb_get_game_timestamps",
		"mlb_get_game_uniforms",
		"mlb_get_game_win_probability",
		"mlb_get_high_low",
		"mlb_get_home_run_derby",
		"mlb_get_jobs",
		"mlb_get_jobs_datacasters",
		"mlb_get_jobs_official_scorers",
		"mlb_get_jobs_umpires",
		"mlb_get_leagues",
		"mlb_get_live_feed",
		"mlb_get_meta",
		"mlb_get_people",
		"mlb_get_people_changes",
		"mlb_get_person_game_stats",
		"mlb_get_play_by_play",
		"mlb_get_schedule_postseason_series",
		"mlb_get_schedule_postseason_tune_in",
		"mlb_get_schedule_tied",
		"mlb_get_season",
		"mlb_get_seasons",
		"mlb_get_sports",
		"mlb_get_sports_players",
		"mlb_get_stats_streaks",
		"mlb_get_team_alumni",
		"mlb_get_team_coaches",
		"mlb_get_team_leaders",
		"mlb_get_team_personnel",
		"mlb_get_team_stats",
		"mlb_get_team_uniforms",
		"mlb_get_teams",
		"mlb_get_teams_affiliates",
		"mlb_get_teams_history",
		"mlb_get_teams_stats",
		"mlb_get_umpire_games",
		"mlb_get_venue",
	}

	// Tier 1 — composed tools.
	want := make([]string, 0, 11+len(rawTools))
	want = append(
		want,
		"scores",
		"standings",
		"player_bio",
		"team_info",
		"team_roster",
		"league_leaders",
		"game_detail",
		"game_linescore",
		"recent_transactions",
		"free_agents",
		"postseason_schedule",
		"player_stats",
		"draft",
	)
	want = append(want, rawTools...)

	for _, name := range want {
		if !got[name] {
			t.Errorf("expected tool %q not registered", name)
		}
	}
	if len(resp.Tools) != len(want) {
		t.Errorf("len(tools) = %d, want %d", len(resp.Tools), len(want))
	}
}

// TestServer_CallTool is the table-driven round-trip battery — one row per
// tool happy path plus error paths. The fakeDriver's fn fields are set per
// row so only the relevant method fires.
func TestServer_CallTool(t *testing.T) {
	errBoom := errors.New("boom")

	cases := []struct {
		name        string
		driver      *fakeDriver
		tool        string
		args        map[string]any
		wantSubstrs []string // substrings expected in the first TextContent block
		wantErr     bool     // true when we expect a non-nil error or isError result
	}{
		// ── scores ──────────────────────────────────────────────────────
		{
			name: "scores-returns-two-games",
			driver: &fakeDriver{
				scheduleFn: func(_ context.Context, _ mlb.ScheduleQuery) ([]mlb.Game, error) {
					return []mlb.Game{
						{
							GamePk: 745455,
							Status: mlb.StatusFinal,
							Home:   mlb.TeamScore{ID: 119, Name: "Los Angeles Dodgers", Score: 5},
							Away:   mlb.TeamScore{ID: 147, Name: "New York Yankees", Score: 3},
						},
						{
							GamePk: 745456,
							Status: mlb.StatusLive,
							Home:   mlb.TeamScore{ID: 111, Name: "Boston Red Sox", Score: 2},
							Away:   mlb.TeamScore{ID: 110, Name: "Baltimore Orioles", Score: 1},
						},
					}, nil
				},
			},
			tool: "scores",
			args: nil,
			wantSubstrs: []string{
				"745455",
				"Los Angeles Dodgers",
				"New York Yankees",
				"Final",
				"745456",
				"Boston Red Sox",
				"Live",
			},
		},
		{
			name:        "scores-with-date",
			driver:      &fakeDriver{},
			tool:        "scores",
			args:        map[string]any{"date": "2024-09-07"},
			wantSubstrs: []string{},
		},
		{
			name:        "scores-with-date-range",
			driver:      &fakeDriver{},
			tool:        "scores",
			args:        map[string]any{"from": "2024-09-01", "to": "2024-09-07"},
			wantSubstrs: []string{},
		},
		{
			name:        "scores-with-team-filter",
			driver:      &fakeDriver{},
			tool:        "scores",
			args:        map[string]any{"team_id": 119},
			wantSubstrs: []string{},
		},
		{
			name:    "scores-bad-date-rejects",
			driver:  &fakeDriver{},
			tool:    "scores",
			args:    map[string]any{"date": "not-a-date"},
			wantErr: true,
		},
		{
			name:    "scores-bad-from-date-rejects",
			driver:  &fakeDriver{},
			tool:    "scores",
			args:    map[string]any{"from": "bad", "to": "2024-09-07"},
			wantErr: true,
		},
		{
			name:    "scores-bad-to-date-rejects",
			driver:  &fakeDriver{},
			tool:    "scores",
			args:    map[string]any{"from": "2024-09-01", "to": "bad"},
			wantErr: true,
		},
		{
			name: "scores-driver-error-propagates",
			driver: &fakeDriver{
				scheduleFn: func(_ context.Context, _ mlb.ScheduleQuery) ([]mlb.Game, error) {
					return nil, errBoom
				},
			},
			tool:    "scores",
			args:    nil,
			wantErr: true,
		},

		// ── standings ─────────────────────────────────────────────────────────
		{
			name: "standings-AL-filter",
			driver: &fakeDriver{
				standingsFn: func(_ context.Context, _ mlb.StandingsQuery) (*mlb.Standings, error) {
					return &mlb.Standings{
						Records: []mlb.DivisionStandings{
							{
								Division: mlb.Ref{ID: 201},
								TeamRecords: []mlb.TeamRecord{
									{
										Team: mlb.TeamRef{
											ID:   147,
											Name: "New York Yankees",
										},
										Wins:              90,
										Losses:            40,
										WinningPercentage: ".692",
										GamesBack:         "-",
									},
								},
							},
						},
					}, nil
				},
			},
			tool: "standings",
			args: map[string]any{"league": "AL"},
			wantSubstrs: []string{
				`"league"`,
				"New York Yankees",
				".692",
			},
		},
		{
			name: "standings-both-leagues-no-filter",
			driver: &fakeDriver{
				standingsFn: func(_ context.Context, q mlb.StandingsQuery) (*mlb.Standings, error) { //nolint:revive
					switch q.League {
					case 103:
						return &mlb.Standings{
							Records: []mlb.DivisionStandings{
								{
									Division: mlb.Ref{ID: 201},
									TeamRecords: []mlb.TeamRecord{
										{
											Team:   mlb.TeamRef{Name: "Houston Astros"},
											Wins:   85,
											Losses: 55,
										},
									},
								},
							},
						}, nil
					case 104:
						return &mlb.Standings{
							Records: []mlb.DivisionStandings{
								{
									Division: mlb.Ref{ID: 204},
									TeamRecords: []mlb.TeamRecord{
										{
											Team:   mlb.TeamRef{Name: "Los Angeles Dodgers"},
											Wins:   95,
											Losses: 45,
										},
									},
								},
							},
						}, nil
					default:
						return &mlb.Standings{}, nil
					}
				},
			},
			tool: "standings",
			args: nil,
			wantSubstrs: []string{
				"Houston Astros",
				"Los Angeles Dodgers",
			},
		},
		{
			name: "standings-NL-filter",
			driver: &fakeDriver{
				standingsFn: func(_ context.Context, _ mlb.StandingsQuery) (*mlb.Standings, error) {
					return &mlb.Standings{
						Records: []mlb.DivisionStandings{
							{
								Division: mlb.Ref{ID: 204},
								TeamRecords: []mlb.TeamRecord{
									{
										Team: mlb.TeamRef{
											ID:   119,
											Name: "Los Angeles Dodgers",
										},
										Wins:              95,
										Losses:            45,
										WinningPercentage: ".679",
										GamesBack:         "-",
									},
								},
							},
						},
					}, nil
				},
			},
			tool: "standings",
			args: map[string]any{"league": "NL"},
			wantSubstrs: []string{
				"Los Angeles Dodgers",
				".679",
			},
		},
		{
			name: "standings-driver-error-propagates",
			driver: &fakeDriver{
				standingsFn: func(_ context.Context, _ mlb.StandingsQuery) (*mlb.Standings, error) {
					return nil, errBoom
				},
			},
			tool:    "standings",
			args:    map[string]any{"league": "AL"},
			wantErr: true,
		},

		// ── player_bio ────────────────────────────────────────────────────────
		{
			name: "player_bio-happy-path",
			driver: &fakeDriver{
				personFn: func(_ context.Context, _ int, _ mlb.PersonQuery) (*mlb.PersonDetail, error) {
					return &mlb.PersonDetail{
						ID:            660271,
						FullName:      "Shohei Ohtani",
						PrimaryNumber: "17",
						Active:        true,
						BatSide:       mlb.HandSide{Code: "L", Description: "Left"},
						PitchHand:     mlb.HandSide{Code: "R", Description: "Right"},
						Height:        "6' 4\"",
						Weight:        210,
					}, nil
				},
			},
			tool: "player_bio",
			args: map[string]any{"person_id": 660271},
			wantSubstrs: []string{
				"Shohei Ohtani",
				"660271",
				`"Active": true`,
			},
		},
		{
			name: "player_bio-driver-error-propagates",
			driver: &fakeDriver{
				personFn: func(_ context.Context, _ int, _ mlb.PersonQuery) (*mlb.PersonDetail, error) {
					return nil, errBoom
				},
			},
			tool:    "player_bio",
			args:    map[string]any{"person_id": 660271},
			wantErr: true,
		},

		// ── team_info ─────────────────────────────────────────────────────────
		{
			name: "team_info-happy-path",
			driver: &fakeDriver{
				teamFn: func(_ context.Context, _ int, _ mlb.TeamQuery) (*mlb.TeamInfo, error) {
					return &mlb.TeamInfo{
						ID:           119,
						Name:         "Los Angeles Dodgers",
						Abbreviation: "LAD",
						LocationName: "Los Angeles",
						Active:       true,
						League:       mlb.LeagueInfo{ID: 104, Name: "National League"},
					}, nil
				},
			},
			tool: "team_info",
			args: map[string]any{"team_id": 119},
			wantSubstrs: []string{
				"Los Angeles Dodgers",
				"LAD",
				"National League",
			},
		},
		{
			name: "team_info-driver-error-propagates",
			driver: &fakeDriver{
				teamFn: func(_ context.Context, _ int, _ mlb.TeamQuery) (*mlb.TeamInfo, error) {
					return nil, errBoom
				},
			},
			tool:    "team_info",
			args:    map[string]any{"team_id": 119},
			wantErr: true,
		},

		// ── team_roster ───────────────────────────────────────────────────────
		{
			name: "team_roster-happy-path",
			driver: &fakeDriver{
				rosterFn: func(_ context.Context, _ int, _ mlb.RosterQuery) (*mlb.Roster, error) {
					return &mlb.Roster{
						Roster: []mlb.RosterEntry{
							{
								Person:       mlb.Person{ID: 660271, FullName: "Shohei Ohtani"},
								JerseyNumber: "17",
								Position: mlb.PrimaryPosition{
									Abbreviation: "DH",
									Name:         "Designated Hitter",
								},
							},
							{
								Person:       mlb.Person{ID: 605141, FullName: "Mookie Betts"},
								JerseyNumber: "50",
								Position: mlb.PrimaryPosition{
									Abbreviation: "SS",
									Name:         "Shortstop",
								},
							},
						},
					}, nil
				},
			},
			tool: "team_roster",
			args: map[string]any{"team_id": 119},
			wantSubstrs: []string{
				"Shohei Ohtani",
				"Mookie Betts",
				`"JerseyNumber": "17"`,
				`"JerseyNumber": "50"`,
			},
		},
		{
			name: "team_roster-driver-error-propagates",
			driver: &fakeDriver{
				rosterFn: func(_ context.Context, _ int, _ mlb.RosterQuery) (*mlb.Roster, error) {
					return nil, errBoom
				},
			},
			tool:    "team_roster",
			args:    map[string]any{"team_id": 119},
			wantErr: true,
		},

		// ── league_leaders ────────────────────────────────────────────────────
		{
			name: "league_leaders-happy-path",
			driver: &fakeDriver{
				statsLeadersFn: func(_ context.Context, _ mlb.StatsLeadersQuery) (*mlb.StatsLeaders, error) {
					return &mlb.StatsLeaders{
						LeagueLeaders: []mlb.LeaderCategory{
							{
								LeaderCategory: "homeRuns",
								Leaders: []mlb.LeaderEntry{
									{
										Rank:   1,
										Value:  "52",
										Player: mlb.Person{ID: 660271, FullName: "Shohei Ohtani"},
										Team:   mlb.TeamRef{ID: 119, Name: "Los Angeles Dodgers"},
									},
								},
							},
						},
					}, nil
				},
			},
			tool: "league_leaders",
			args: map[string]any{"category": "homeRuns"},
			wantSubstrs: []string{
				"homeRuns",
				"Shohei Ohtani",
				`"Value": "52"`,
			},
		},
		{
			name: "league_leaders-driver-error-propagates",
			driver: &fakeDriver{
				statsLeadersFn: func(_ context.Context, _ mlb.StatsLeadersQuery) (*mlb.StatsLeaders, error) {
					return nil, errBoom
				},
			},
			tool:    "league_leaders",
			args:    map[string]any{"category": "homeRuns"},
			wantErr: true,
		},

		// ── game_detail ───────────────────────────────────────────────────────
		{
			name: "game_detail-happy-path",
			driver: &fakeDriver{
				boxscoreFn: func(_ context.Context, _ int) (*mlb.Boxscore, error) {
					return &mlb.Boxscore{
						Home: &mlb.BoxscoreTeam{
							ID:   119,
							Name: "Los Angeles Dodgers",
							Batting: mlb.BattingStats{
								Runs: 5, Hits: 10, HomeRuns: 2, RBI: 5,
							},
							Pitching: mlb.PitchingStats{
								Strikeouts: 9, Hits: 6, Runs: 3,
							},
						},
						Away: &mlb.BoxscoreTeam{
							ID:   147,
							Name: "New York Yankees",
							Batting: mlb.BattingStats{
								Runs: 3, Hits: 6, HomeRuns: 1, RBI: 3,
							},
							Pitching: mlb.PitchingStats{
								Strikeouts: 7, Hits: 10, Runs: 5,
							},
						},
					}, nil
				},
			},
			tool: "game_detail",
			args: map[string]any{"game_pk": 745455},
			wantSubstrs: []string{
				"Los Angeles Dodgers",
				"New York Yankees",
				`"Strikeouts": 9`,
				`"HomeRuns": 2`,
			},
		},
		{
			name: "game_detail-driver-error-propagates",
			driver: &fakeDriver{
				boxscoreFn: func(_ context.Context, _ int) (*mlb.Boxscore, error) {
					return nil, errBoom
				},
			},
			tool:    "game_detail",
			args:    map[string]any{"game_pk": 745455},
			wantErr: true,
		},

		// ── game_linescore ────────────────────────────────────────────────────
		{
			name: "game_linescore-happy-path",
			driver: &fakeDriver{
				linescoreFn: func(_ context.Context, _ int, _ mlb.LinescoreQuery) (*mlb.Linescore, error) {
					return &mlb.Linescore{
						CurrentInning:        9,
						CurrentInningOrdinal: "9th",
						ScheduledInnings:     9,
						Teams: mlb.LinescoreTeams{
							Home: mlb.LinescoreTeamTotals{
								Runs:     5,
								Hits:     10,
								Errors:   0,
								IsWinner: true,
							},
							Away: mlb.LinescoreTeamTotals{Runs: 3, Hits: 7, Errors: 1},
						},
						Innings: []mlb.LinescoreInning{
							{
								Num:        1,
								OrdinalNum: "1st",
								Home:       mlb.LinescoreInningHalf{Runs: 2, Hits: 3},
								Away:       mlb.LinescoreInningHalf{Runs: 0, Hits: 1},
							},
						},
					}, nil
				},
			},
			tool: "game_linescore",
			args: map[string]any{"game_pk": 745455},
			wantSubstrs: []string{
				`"CurrentInning": 9`,
				`"CurrentInningOrdinal": "9th"`,
				`"IsWinner": true`,
			},
		},
		{
			name: "game_linescore-driver-error-propagates",
			driver: &fakeDriver{
				linescoreFn: func(_ context.Context, _ int, _ mlb.LinescoreQuery) (*mlb.Linescore, error) {
					return nil, errBoom
				},
			},
			tool:    "game_linescore",
			args:    map[string]any{"game_pk": 745455},
			wantErr: true,
		},

		// ── recent_transactions ───────────────────────────────────────────────
		{
			name: "recent_transactions-happy-path",
			driver: &fakeDriver{
				transactionsFn: func(_ context.Context, _ mlb.TransactionsQuery) (*mlb.Transactions, error) {
					return &mlb.Transactions{
						Transactions: []mlb.Transaction{
							{
								ID:          1001,
								TypeCode:    "TR",
								TypeDesc:    "Trade",
								Description: "Player traded from NYY to LAD",
								Person:      mlb.Person{ID: 99999, FullName: "Trade Bait"},
								FromTeam:    mlb.TeamRef{ID: 147, Name: "New York Yankees"},
								ToTeam:      mlb.TeamRef{ID: 119, Name: "Los Angeles Dodgers"},
							},
						},
					}, nil
				},
			},
			tool: "recent_transactions",
			args: map[string]any{"days": 1},
			wantSubstrs: []string{
				"Trade Bait",
				`"TypeCode": "TR"`,
				"New York Yankees",
				"Los Angeles Dodgers",
			},
		},
		{
			name: "recent_transactions-with-team-filter",
			driver: &fakeDriver{
				transactionsFn: func(_ context.Context, _ mlb.TransactionsQuery) (*mlb.Transactions, error) {
					return &mlb.Transactions{
						Transactions: []mlb.Transaction{
							{
								ID:       1002,
								TypeCode: "DFA",
								TypeDesc: "Designated for Assignment",
								Person:   mlb.Person{ID: 88888, FullName: "DFA Guy"},
								ToTeam:   mlb.TeamRef{ID: 119, Name: "Los Angeles Dodgers"},
							},
						},
					}, nil
				},
			},
			tool: "recent_transactions",
			args: map[string]any{"team_id": 119},
			wantSubstrs: []string{
				"DFA Guy",
				`"TypeCode": "DFA"`,
			},
		},
		{
			name: "recent_transactions-driver-error-propagates",
			driver: &fakeDriver{
				transactionsFn: func(_ context.Context, _ mlb.TransactionsQuery) (*mlb.Transactions, error) {
					return nil, errBoom
				},
			},
			tool:    "recent_transactions",
			args:    map[string]any{"days": 1},
			wantErr: true,
		},

		// ── free_agents ───────────────────────────────────────────────────────
		{
			name: "free_agents-happy-path",
			driver: &fakeDriver{
				freeAgentsFn: func(_ context.Context, _ mlb.FreeAgentsQuery) (*mlb.FreeAgents, error) {
					return &mlb.FreeAgents{
						FreeAgents: []mlb.FreeAgent{
							{
								Player:       mlb.Person{ID: 502671, FullName: "Bryce Harper"},
								OriginalTeam: mlb.TeamRef{ID: 143, Name: "Philadelphia Phillies"},
								NewTeam:      mlb.TeamRef{},
								DateDeclared: "2026-11-01",
							},
						},
					}, nil
				},
			},
			tool: "free_agents",
			args: map[string]any{"season": 2026},
			wantSubstrs: []string{
				"Bryce Harper",
				"Philadelphia Phillies",
				"2026-11-01",
			},
		},
		{
			name: "free_agents-driver-error-propagates",
			driver: &fakeDriver{
				freeAgentsFn: func(_ context.Context, _ mlb.FreeAgentsQuery) (*mlb.FreeAgents, error) {
					return nil, errBoom
				},
			},
			tool:    "free_agents",
			args:    nil,
			wantErr: true,
		},

		// ── postseason_schedule ───────────────────────────────────────────────
		{
			name: "postseason_schedule-happy-path",
			driver: &fakeDriver{
				schedulePostseasonFn: func(_ context.Context, _ mlb.SchedulePostseasonQuery) ([]mlb.Game, error) {
					return []mlb.Game{
						{
							GamePk: 800001,
							Status: mlb.StatusFinal,
							Home:   mlb.TeamScore{ID: 119, Name: "Los Angeles Dodgers", Score: 4},
							Away:   mlb.TeamScore{ID: 147, Name: "New York Yankees", Score: 2},
						},
					}, nil
				},
			},
			tool: "postseason_schedule",
			args: map[string]any{"season": 2026},
			wantSubstrs: []string{
				"800001",
				"Los Angeles Dodgers",
				"New York Yankees",
				"Final",
			},
		},
		{
			name: "postseason_schedule-driver-error-propagates",
			driver: &fakeDriver{
				schedulePostseasonFn: func(_ context.Context, _ mlb.SchedulePostseasonQuery) ([]mlb.Game, error) {
					return nil, errBoom
				},
			},
			tool:    "postseason_schedule",
			args:    nil,
			wantErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cs := mcpServerHarness(t, tc.driver)

			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			argBytes, _ := json.Marshal(tc.args)
			resp, err := cs.CallTool(ctx, &mcpsdk.CallToolParams{
				Name:      tc.tool,
				Arguments: json.RawMessage(argBytes),
			})

			if tc.wantErr {
				// The tool may surface the error either as an SDK-level error
				// (err != nil) or as an MCP isError result (resp.IsError).
				if err != nil {
					return // error propagated as expected
				}
				if resp != nil && resp.IsError {
					return // tool handler surfaced isError as expected
				}
				t.Fatalf("tool %s: expected error or isError result, got %+v", tc.tool, resp)
				return
			}

			if err != nil {
				t.Fatalf("call %s: %v", tc.tool, err)
			}
			if resp.IsError {
				t.Fatalf("tool %s returned isError: %v", tc.tool, resp.Content)
			}
			if len(resp.Content) == 0 {
				t.Fatalf("tool %s returned no content", tc.tool)
			}
			text, ok := resp.Content[0].(*mcpsdk.TextContent)
			if !ok {
				t.Fatalf(
					"tool %s: first content type = %T, want *TextContent",
					tc.tool,
					resp.Content[0],
				)
			}
			for _, want := range tc.wantSubstrs {
				if !strings.Contains(text.Text, want) {
					t.Errorf(
						"tool %s response missing %q\n--- got ---\n%s",
						tc.tool,
						want,
						text.Text,
					)
				}
			}
		})
	}
}

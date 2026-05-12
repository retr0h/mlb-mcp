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

	"github.com/retr0h/mlb-sdk/pkg/mlb"
)

// Driver is the narrow consumer-surface the MCP server requires of its MLB
// SDK client, declared at the consumer seam per the osapi-io pattern.
// Concrete *mlb.Client satisfies it structurally — the compiler verifies at
// the assignment site in New().
//
// New methods get added here first (declare what we need) then the concrete
// *mlb.Client already satisfies them. Keeping the interface narrow lets tests
// slot in a fake without importing the full SDK client.
type Driver interface {
	// Schedule returns games matching the query. SportId is always pinned to
	// 1 (MLB).
	Schedule(ctx context.Context, q mlb.ScheduleQuery) ([]mlb.Game, error)

	// Standings fetches league standings. q.League is required.
	Standings(ctx context.Context, q mlb.StandingsQuery) (*mlb.Standings, error)

	// Person fetches a single person by id. Returns ErrNotFound for 404.
	Person(ctx context.Context, personID int, q mlb.PersonQuery) (*mlb.PersonDetail, error)

	// Team fetches a single team by id. Returns ErrNotFound for 404.
	Team(ctx context.Context, teamID int, q mlb.TeamQuery) (*mlb.TeamInfo, error)

	// StatsLeaders fetches league stat leaders. q.LeaderCategories is required.
	StatsLeaders(ctx context.Context, q mlb.StatsLeadersQuery) (*mlb.StatsLeaders, error)

	// Linescore fetches the inning-by-inning linescore for a game.
	Linescore(ctx context.Context, gamePk int, q mlb.LinescoreQuery) (*mlb.Linescore, error)

	// Boxscore fetches the team-stats boxscore for a game by gamePk.
	Boxscore(ctx context.Context, gamePk int) (*mlb.Boxscore, error)

	// Roster fetches a team's roster filtered by the query.
	Roster(ctx context.Context, teamID int, q mlb.RosterQuery) (*mlb.Roster, error)

	// Transactions fetches roster/assignment transactions.
	Transactions(ctx context.Context, q mlb.TransactionsQuery) (*mlb.Transactions, error)

	// FreeAgents fetches free-agent declarations and signings.
	FreeAgents(ctx context.Context, q mlb.FreeAgentsQuery) (*mlb.FreeAgents, error)

	// SchedulePostseason fetches the postseason schedule.
	SchedulePostseason(ctx context.Context, q mlb.SchedulePostseasonQuery) ([]mlb.Game, error)

	// Stats fetches individual player stats (season, career, etc.).
	Stats(ctx context.Context, q mlb.StatsQuery) (*mlb.TeamStats, error)

	// Draft fetches draft data for a given year.
	Draft(ctx context.Context, year int, q mlb.DraftQuery) (*mlb.DraftData, error)
}

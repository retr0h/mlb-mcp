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

// fakeDriver satisfies the Driver interface for unit tests. Each method is a
// func field so individual test rows can override exactly the path under test
// without touching unrelated methods. When a field is nil the method returns
// the zero value (empty slice / nil pointer) with no error.
type fakeDriver struct {
	scheduleFn           func(ctx context.Context, q mlb.ScheduleQuery) ([]mlb.Game, error)
	standingsFn          func(ctx context.Context, q mlb.StandingsQuery) (*mlb.Standings, error)
	personFn             func(ctx context.Context, personID int, q mlb.PersonQuery) (*mlb.PersonDetail, error)
	teamFn               func(ctx context.Context, teamID int, q mlb.TeamQuery) (*mlb.TeamInfo, error)
	statsLeadersFn       func(ctx context.Context, q mlb.StatsLeadersQuery) (*mlb.StatsLeaders, error)
	linescoreFn          func(ctx context.Context, gamePk int, q mlb.LinescoreQuery) (*mlb.Linescore, error)
	boxscoreFn           func(ctx context.Context, gamePk int) (*mlb.Boxscore, error)
	rosterFn             func(ctx context.Context, teamID int, q mlb.RosterQuery) (*mlb.Roster, error)
	transactionsFn       func(ctx context.Context, q mlb.TransactionsQuery) (*mlb.Transactions, error)
	freeAgentsFn         func(ctx context.Context, q mlb.FreeAgentsQuery) (*mlb.FreeAgents, error)
	schedulePostseasonFn func(ctx context.Context, q mlb.SchedulePostseasonQuery) ([]mlb.Game, error)
	statsFn              func(ctx context.Context, q mlb.StatsQuery) (*mlb.TeamStats, error)
	draftFn              func(ctx context.Context, year int, q mlb.DraftQuery) (*mlb.DraftData, error)
}

func (f *fakeDriver) Schedule(ctx context.Context, q mlb.ScheduleQuery) ([]mlb.Game, error) {
	if f.scheduleFn != nil {
		return f.scheduleFn(ctx, q)
	}
	return nil, nil
}

func (f *fakeDriver) Standings(ctx context.Context, q mlb.StandingsQuery) (*mlb.Standings, error) {
	if f.standingsFn != nil {
		return f.standingsFn(ctx, q)
	}
	return nil, nil
}

func (f *fakeDriver) Person(
	ctx context.Context,
	personID int,
	q mlb.PersonQuery,
) (*mlb.PersonDetail, error) {
	if f.personFn != nil {
		return f.personFn(ctx, personID, q)
	}
	return nil, nil
}

func (f *fakeDriver) Team(ctx context.Context, teamID int, q mlb.TeamQuery) (*mlb.TeamInfo, error) {
	if f.teamFn != nil {
		return f.teamFn(ctx, teamID, q)
	}
	return nil, nil
}

func (f *fakeDriver) StatsLeaders(
	ctx context.Context,
	q mlb.StatsLeadersQuery,
) (*mlb.StatsLeaders, error) {
	if f.statsLeadersFn != nil {
		return f.statsLeadersFn(ctx, q)
	}
	return nil, nil
}

func (f *fakeDriver) Linescore(
	ctx context.Context,
	gamePk int,
	q mlb.LinescoreQuery,
) (*mlb.Linescore, error) {
	if f.linescoreFn != nil {
		return f.linescoreFn(ctx, gamePk, q)
	}
	return nil, nil
}

func (f *fakeDriver) Boxscore(ctx context.Context, gamePk int) (*mlb.Boxscore, error) {
	if f.boxscoreFn != nil {
		return f.boxscoreFn(ctx, gamePk)
	}
	return nil, nil
}

func (f *fakeDriver) Roster(
	ctx context.Context,
	teamID int,
	q mlb.RosterQuery,
) (*mlb.Roster, error) {
	if f.rosterFn != nil {
		return f.rosterFn(ctx, teamID, q)
	}
	return nil, nil
}

func (f *fakeDriver) Transactions(
	ctx context.Context,
	q mlb.TransactionsQuery,
) (*mlb.Transactions, error) {
	if f.transactionsFn != nil {
		return f.transactionsFn(ctx, q)
	}
	return nil, nil
}

func (f *fakeDriver) FreeAgents(
	ctx context.Context,
	q mlb.FreeAgentsQuery,
) (*mlb.FreeAgents, error) {
	if f.freeAgentsFn != nil {
		return f.freeAgentsFn(ctx, q)
	}
	return nil, nil
}

func (f *fakeDriver) SchedulePostseason(
	ctx context.Context,
	q mlb.SchedulePostseasonQuery,
) ([]mlb.Game, error) {
	if f.schedulePostseasonFn != nil {
		return f.schedulePostseasonFn(ctx, q)
	}
	return nil, nil
}

func (f *fakeDriver) Stats(ctx context.Context, q mlb.StatsQuery) (*mlb.TeamStats, error) {
	if f.statsFn != nil {
		return f.statsFn(ctx, q)
	}
	return &mlb.TeamStats{}, nil
}

func (f *fakeDriver) Draft(
	ctx context.Context,
	year int,
	q mlb.DraftQuery,
) (*mlb.DraftData, error) {
	if f.draftFn != nil {
		return f.draftFn(ctx, year, q)
	}
	return &mlb.DraftData{}, nil
}

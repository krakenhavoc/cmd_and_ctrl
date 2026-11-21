package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b137_helpers.go — Dominaria Remastered creatures, GAP batch 137
// (#5948).

// gapB137LifeLeader is the player who has more life than each other
// player (Wild Dogs): the unique highest life total among the players still
// in the game. Two players tied at the top mean nobody has "more life than
// each other player", so the condition is false (uuid.Nil, false).
func GapB137LifeLeader(g *game.Game) (uuid.UUID, bool) {
	var best uuid.UUID
	bestLife, tied, found := 0, false, false
	for _, p := range g.Seats {
		if p == nil || p.Eliminated {
			continue
		}
		switch {
		case !found || p.Life > bestLife:
			best, bestLife, tied, found = p.ID, p.Life, false, true
		case p.Life == bestLife:
			tied = true
		}
	}
	if !found || tied {
		return uuid.Nil, false
	}
	return best, true
}

// gapB137ALifeLeaderExists is the trigger-time half of Wild Dogs'
// intervening "if a player has more life than each other player".
func GapB137ALifeLeaderExists(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	_, ok := GapB137LifeLeader(g)
	return ok
}

// gapB137LifeLeaderCondition is the same condition as the trigger resolves
// (CR 603.4).
func GapB137LifeLeaderCondition() Condition {
	return CondBoard{
		Name: "a player has more life than each other player",
		Test: func(ctx *Context) bool {
			_, ok := GapB137LifeLeader(ctx.Game)
			return ok
		},
	}
}

// GapB137LandCardsInAllGraveyards is Terravore's count: the land cards in
// every player's graveyard, read from the printed card (a card in a
// graveyard has no layer cache).
func GapB137LandCardsInAllGraveyards(g *game.Game) int {
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.Graveyard == nil {
			continue
		}
		for _, c := range p.Graveyard.Cards {
			if c.IsLand() {
				n++
			}
		}
	}
	return n
}

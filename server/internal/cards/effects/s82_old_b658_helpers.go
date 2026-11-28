package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b658_helpers.go — a narrower SeparateIntoPiles for Bend or Break
// (Invasion, #5616): "Each player separates all nontoken lands they control
// into two piles."
//
// The set a player separates is part of the card's registered rest (its
// PilesThen key), so it is on disk with the rest of the prompt and a table
// waiting on a split is still a restore point.

// pilesOfNontokenLands holds the PilesThen keys whose split covers a
// player's nontoken lands rather than all their permanents. Written only
// while the package initialises (RegisterPilesThenOfNontokenLands is called
// from a package-level var), read afterwards.
var pilesOfNontokenLands = map[string]bool{}

// RegisterPilesThenOfNontokenLands is RegisterPilesThen for a card whose
// players separate their nontoken lands.
func RegisterPilesThenOfNontokenLands(key string, body func(ctx *Context, piles []PlayerPiles) error) PilesThen {
	pilesOfNontokenLands[key] = true
	return RegisterPilesThen(key, body)
}

// pilesCandidates is what `who` separates for the split whose registered
// rest is `key`: every permanent they control, or only their nontoken lands.
func pilesCandidates(g *game.Game, who uuid.UUID, key string) []uuid.UUID {
	if !pilesOfNontokenLands[key] {
		return PermanentsControlledByPlayer(g, who)
	}
	var out []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == who && c.IsLand() && !c.IsToken() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

// pilesNoun is the word a split prompt uses for what is being separated.
func pilesNoun(key string) string {
	if pilesOfNontokenLands[key] {
		return "nontoken lands"
	}
	return "permanents"
}

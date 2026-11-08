package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nightwhorl Hermit — Creature — Rat Rogue {2}{U}, 1/4 (Bloomburrow,
// #4467):
//
//	"Vigilance
//	 Threshold — As long as there are seven or more cards in your
//	 graveyard, this creature gets +1/+0 and can't be blocked."
//
// Vigilance is a printed keyword. Threshold is two statics read live:
// the layer 7c self-modifier Mind Drill Assailant uses, and a block rule
// whose condition is the same seven-card count, read as blockers are
// declared. A graveyard that drops below seven after blocks are declared
// does not undo a block that was already legal.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "99007c40-1eb5-4c3a-9889-6d12417d8191",
		Name:            "Nightwhorl Hermit",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static:          []game.StaticAbility{fraThresholdSelfPT(1, 0)},
		BlockRules: []game.BlockRule{CantBeBlockedWhile(OnSelf(), func(g *game.Game, _ uuid.UUID, c game.Card) bool {
			return fraThresholdMet(g, c.Controller)
		})},
	})
}

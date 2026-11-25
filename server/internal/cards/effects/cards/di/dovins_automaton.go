package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dovin's Automaton — Artifact Creature — Homunculus {4}, 3/3 (Ravnica
// Allegiance):
//
//	"As long as you control a Dovin planeswalker, this creature gets +2/+2
//	 and has vigilance."
//
// Both halves are statics gated on the controller holding a planeswalker with
// the Dovin subtype, re-read on every layer pass.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d133b18-f1c7-48de-82a6-affb0514b96b",
		Name:         "Dovin's Automaton",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.B16Anthem(controlsDovin, 2, 2),
			effects.KeywordGrant(controlsDovin, "vigilance"),
		},
	})
}

// controlsDovin scopes a static to this creature while its controller
// controls a Dovin planeswalker.
func controlsDovin(target *game.Card, g *game.Game, source *game.Card) bool {
	return effects.SelfOnly(target, g, source) &&
		len(effects.PermanentsControlledByMatching(g, source.Controller, effects.And(effects.Planeswalker(), effects.HasSubtype("Dovin")))) > 0
}

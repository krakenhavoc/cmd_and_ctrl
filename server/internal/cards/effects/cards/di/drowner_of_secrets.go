package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drowner of Secrets — Creature — Merfolk Wizard {2}{U}, 1/3 (Lorwyn, S82
// OLD b547, #5593):
//
//	"Tap an untapped Merfolk you control: Target player mills a card."
//
// The cost is Zombie Trailblazer's "Tap an untapped <type> you control":
// not the {T} symbol (CR 302.6), so a summoning-sick Merfolk may pay it, the
// Drowner itself included, and it does not target. A changeling is a Merfolk.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5ce3fb6-9c70-4433-a079-467c911bfd4d",
		Name:         "Drowner of Secrets",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Tap an untapped Merfolk you control: Target player mills a card.",
			Cost:    effects.TapAnUntapped("an untapped Merfolk you control", effects.HasSubtype("Merfolk")),
			Targets: effects.TargetPlayer("target player"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				for _, t := range effects.NewContext(g, item).LegalTargets() {
					if t.Kind == game.TargetPlayer {
						return g.MillNForEffect(t.ID, 1)
					}
				}
				return nil
			},
		}},
	})
}

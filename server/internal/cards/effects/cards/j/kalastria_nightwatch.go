package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kalastria Nightwatch — Creature — Vampire Warrior Ally {4}{B}, 4/5
// (Jumpstart 2022, #5948):
//
//	"Whenever you gain life, this creature gains flying until end of turn."
//
// Ajani's Pridemate's lifegain trigger (one per gain event, lifelink
// included) with the shared "this creature gains <keyword> until end of
// turn" body; a Nightwatch that has left the battlefield gains nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f610748a-8b2c-4c7d-b495-b4677ff4c423",
		Name:         "Kalastria Nightwatch",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouGainLife("Kalastria Nightwatch — gains flying until end of turn", func(g *game.Game, item *game.StackItem) error {
				if !effects.OnBattlefield(g, item.SourceCardID) {
					return nil
				}
				return effects.ThisGainsKeywordUntilEndOfTurn("flying", "Kalastria Nightwatch — flying until end of turn")(g, item)
			}),
		},
	})
}

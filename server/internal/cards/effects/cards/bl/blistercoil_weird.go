package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blistercoil Weird — Creature — Weird {U/R}, 1/1 (Return to Ravnica, S82
// OLD):
//
//	"Whenever you cast an instant or sorcery spell, this creature gets +1/+1
//	 until end of turn. Untap it."
//
// Wee Dragonauts' cast trigger with a +1/+1 and an untap, both for the
// Weird only while it is still the permanent the trigger came from (CR
// 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bf9b24e1-60d4-4739-808a-535336e191c2",
		Name:         "Blistercoil Weird",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Or(effects.Instant(), effects.Sorcery()),
				"Blistercoil Weird — gets +1/+1 until end of turn. Untap it",
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.ThisGetsUntilEndOfTurn(1, 1, "Blistercoil Weird — +1/+1")(g, item); err != nil {
						return err
					}
					return effects.UntapThisPermanentEffect(g, item)
				}),
		},
	})
}

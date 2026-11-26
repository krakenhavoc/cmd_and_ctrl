package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Impetuous Protege — Creature — Human Warrior {2}{R}, 0/4 (Battlebond, #5529):
//
//	"Partner with Proud Mentor (When this creature enters, target player may put Proud Mentor into their hand from their library, then shuffle.)"
//	"Whenever this creature attacks, it gets +X/+0 until end of turn, where X is the greatest power among tapped creatures your opponents control."
//
// X is read as the trigger resolves, through the layers (CurrentPower); with no tapped creature of theirs it is 0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a03059dd-f499-407a-9346-6710b0468202",
		Name:         "Impetuous Protege",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Impetuous Protege", "Proud Mentor"),
			effects.WheneverThisAttacks("Impetuous Protege — it gets +X/+0 until end of turn, X the greatest power among tapped creatures your opponents control",
				func(g *game.Game, item *game.StackItem) error {
					x := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller != item.Controller && c.IsCreature() && c.Tapped {
							x = max(x, c.CurrentPower())
						}
					}
					return effects.ThisGetsUntilEndOfTurn(x, 0, "Impetuous Protege — +X/+0")(g, item)
				}),
		},
	})
}

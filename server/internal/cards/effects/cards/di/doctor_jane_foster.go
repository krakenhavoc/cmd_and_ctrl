package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Doctor Jane Foster — Legendary Creature — Human Doctor {3}{W}, 2/3
// (Marvel Super Heroes Commander):
//
//	"Vigilance
//	 When Doctor Jane Foster enters, return target creature card with
//	 mana value 3 or less from your graveyard to your hand. If you gained
//	 life this turn, return that card to the battlefield instead."
//
// The life-gained tally is read when the trigger resolves (CR 608.2c),
// not when it is put on the stack, so lifegain in response counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b1c8fddb-d686-4b4a-8aff-d79ea747801c",
		Name:            "Doctor Jane Foster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Doctor Jane Foster — return a creature card with mana value 3 or less (to the battlefield if you gained life this turn)",
					func(g *game.Game, item *game.StackItem) error {
						if effects.B15LifeGainedThisTurn(g, item.Controller) > 0 {
							return effects.ReturnFirstLegalGraveyardTargetToBattlefield(g, item)
						}
						return effects.ReturnFirstLegalGraveyardTargetToHand(g, item)
					}),
				effects.TargetCardInGraveyard("target creature card with mana value 3 or less in your graveyard", effects.YouOwn(), effects.Creature(), effects.ManaValueLE(3))),
		},
	})
}

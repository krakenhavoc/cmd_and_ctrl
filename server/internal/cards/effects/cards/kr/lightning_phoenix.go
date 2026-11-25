package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lightning Phoenix — Creature — Phoenix {2}{R}, 2/2 (Jumpstart, #5510):
//
//	"Flying, haste
//	 This creature can't block.
//	 At the beginning of your end step, if an opponent was dealt 3 or more
//	 damage this turn, you may pay {R}. If you do, return this card from your
//	 graveyard to the battlefield."
//
// Bloodfeather Phoenix's graveyard return, on the end step. The condition is
// an intervening "if" (CR 603.4): it is read when the step begins and again
// as the trigger resolves, against the damage any single opponent took.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e4e0e23b-f172-4e1c-99d0-bb6f66afbc1c",
		Name:            "Lightning Phoenix",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Static:          []game.StaticAbility{effects.RestrictSelf(game.CantBlock)},
		Triggered: []game.TriggeredAbility{
			effects.InGraveyard(effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, oldB101AnOpponentWasDealt3), "Lightning Phoenix — you may pay {R} to return it from your graveyard to the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					if !oldB101OpponentDealt3(g, item.Controller) {
						return nil
					}
					return effects.MayPayForGraveyardReturn(g, item, "{R}", "Pay {R} to return Lightning Phoenix from your graveyard to the battlefield?")
				})),
		},
	})
}

func oldB101AnOpponentWasDealt3(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return oldB101OpponentDealt3(g, source.Controller)
}

func oldB101OpponentDealt3(g *game.Game, you uuid.UUID) bool {
	for _, s := range g.Seats {
		if s.ID != you && !s.Eliminated && g.DamageTakenThisTurn(s.ID) >= 3 {
			return true
		}
	}
	return false
}

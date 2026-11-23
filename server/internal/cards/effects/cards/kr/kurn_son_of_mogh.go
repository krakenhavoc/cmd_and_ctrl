package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kurn, Son of Mogh — Legendary Creature — Klingon Warrior Officer
// {2}{B}, 2/3 (Star Trek Commander, #4974):
//
//	"Deathtouch
//	 Battlefront — Whenever Kurn attacks, you draw X cards and lose X
//	 life, where X is the number of opponents you attacked this turn."
//
// Battlefront is an ability word. X is counted as the ability resolves
// (OpponentsYouAttackedThisTurn): each opponent still in the game that a
// creature of yours was declared attacking this turn counts once,
// whichever combat it was; an attack on a planeswalker or a battle is not
// an attack on its player (CR 508.3e's distinction). One X is used for
// both halves, the draw first, then the life loss, as printed. With X 0
// (Kurn attacked only a planeswalker or battle) nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2a078705-4afe-4887-94cc-f12f4690fb01",
		Name:            "Kurn, Son of Mogh",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Kurn — battlefront: draw X cards and lose X life, where X is the number of opponents you attacked this turn",
				func(g *game.Game, item *game.StackItem) error {
					x := effects.OpponentsYouAttackedThisTurn(g, item.Controller)
					if x <= 0 {
						return nil
					}
					ctx := effects.NewContext(g, item)
					if err := (effects.DrawCards{Player: item.Controller, N: x}).Apply(ctx); err != nil {
						return err
					}
					return effects.GainLife{Player: item.Controller, Amount: -x}.Apply(ctx)
				}),
		},
	})
}

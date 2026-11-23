package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trynn, Champion of Freedom — Legendary Creature — Human Soldier {3}{W},
// 3/3 (Commander 2020, #5513):
//
//	"Partner with Silvar, Devourer of the Free (When this creature enters,
//	 target player may put Silvar into their hand from their library, then
//	 shuffle.)
//	 At the beginning of your end step, if you attacked this turn, create a
//	 1/1 white Human Soldier creature token."
//
// Partner with is the CR 702.124j entry search (partner_with.go). The end-step
// trigger is Midnight Snack's raid shape: an intervening "if" (CR 603.4),
// checked as the end step begins and again as the trigger resolves; any
// attack you declared this turn counts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "76771556-3eb7-4f87-aaeb-32ea81105aa2",
		Name:         "Trynn, Champion of Freedom",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.PartnerWith("Trynn, Champion of Freedom", "Silvar, Devourer of the Free"),
			effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, effects.TmcB6YouAttackedThisTurn),
				"Trynn, Champion of Freedom — create a 1/1 white Human Soldier",
				effects.Do(effects.If{Cond: effects.AvatarB11CondYouAttacked{}, Then: []effects.Applier{
					effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1},
				}})),
		},
	})
}

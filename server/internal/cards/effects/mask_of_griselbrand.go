package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mask of Griselbrand — Legendary Artifact — Equipment {1}{B}{B} (Duskmourn
// Commander):
//
//	"Equipped creature has flying and lifelink.
//	 Whenever equipped creature dies, you may pay X life, where X is its
//	 power. If you do, draw X cards.
//	 Equip {3}"
//
// The grants are shared keyword statics. The dies trigger is the Skullclamp
// condition (equippedCreatureDied: the Equipment still names the creature when
// the leaves-the-battlefield event fires). X is the dead creature's power as it
// last existed (CR 608.2h, TriggeringPermanent), and the payment is the shared
// "pay N life, draw N" prompt (mayPayLifeDraw), re-checked when the answer
// arrives (CR 119.4). A creature with power 0 or less asks nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b46da6bd-578c-4040-97c3-0005b78e31be",
		Name:         "Mask of Griselbrand",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("flying"),
			GrantToAttached("lifelink"),
		},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return equippedCreatureDied(ev, source)
			}, "Mask of Griselbrand — you may pay X life, where X is its power; if you do, draw X cards",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, ok := ctx.TriggeringPermanent()
					if !ok || info.Power <= 0 {
						return nil
					}
					x := info.Power
					return MayChoice{
						Question: "Mask of Griselbrand — pay " + strconv.Itoa(x) + " life to draw " + strconv.Itoa(x) + " cards?",
						YesLabel: "Pay " + strconv.Itoa(x) + " life",
						NoLabel:  "Decline",
						LifeCost: x,
						Draws:    x,
						Then:     mayPayLifeDraw,
						Word:     strconv.Itoa(x),
					}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{3}")},
	})
}

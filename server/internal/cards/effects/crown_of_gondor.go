package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crown of Gondor — Legendary Artifact — Equipment {3} (Tales of Middle-earth
// Commander, #5948):
//
//	"Equipped creature gets +1/+1 for each creature you control.
//	 When a legendary creature you control enters, if there is no monarch,
//	 you become the monarch.
//	 Equip {4}. This ability costs {3} less to activate if you're the
//	 monarch."
//
// The pump counts the creatures its controller controls on every layer pass
// (PumpAttachedPer, Adaptive Omnitool's shape). The trigger is The Spear of
// Bashenga's intervening "if" (CR 603.4): there is no monarch as the legendary
// creature enters and again as the trigger resolves, so a crown handed out in
// response is not taken. The discount is the equip ability's own cost clause,
// read when the cost is determined (CR 601.2f): {4} while you are not the
// monarch, {1} while you are.
//
// No simplification.
func init() {
	equip := EquipAbility("{4}")
	equip.Label = "Equip {4}. This ability costs {3} less to activate if you're the monarch."
	equip.CostModifiers = []game.CostModifier{
		CostsLess(3, "This ability costs {3} less to activate if you're the monarch.",
			func(q game.CostQuery) bool { return q.Game != nil && YoureTheMonarch(q.Game, q.Controller) }),
	}
	Register(Spec{
		OracleID:     "f95021cf-14c4-4792-8746-432bf9e7c1a9",
		Name:         "Crown of Gondor",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 1, func(g *game.Game, source *game.Card) int {
				return permanentsYouControl(g, source, func(c *game.Card) bool { return c.IsCreature() })
			}),
		},
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "there is no monarch",
				Test: func(ctx *Context) bool { return ctx.Game.Monarch == uuid.Nil },
			}, On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.IsLegendary() && g.Monarch == uuid.Nil
			}, "Crown of Gondor — you become the monarch", Do(BecomeTheMonarch{}))),
		},
		Activated: []ActivatedAbility{equip},
	})
}

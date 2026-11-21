package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// General Marhault Elsdragon — Legendary Creature — Elf Warrior
// {2}{R}{G}, 4/4:
//
//	"Whenever a creature you control becomes blocked, it gets +3/+3 until
//	 end of turn for each creature blocking it."
//
// "Becomes blocked" is the once-per-attacker EventBecomesBlocked (CR
// 509.1h), so a double block is one trigger. The attacker is any creature
// the General's controller controls, the General included. The blockers are
// counted as the ability resolves, the way Berserk Murlodont and rampage
// count them (game.CreaturesBlockingForEffect): a blocker removed from
// combat in response no longer counts, and the bonus is then fixed. Unlike
// rampage there is no "beyond the first": one blocker is already +3/+3. The
// creature is the object that became blocked: one that left the battlefield
// in response gets nothing (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f49b31ce-910e-446a-b7a7-4274d71e961d",
		Name:         "General Marhault Elsdragon",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.Controller == source.Controller
			}, "General Marhault Elsdragon — it gets +3/+3 until end of turn for each creature blocking it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
						return nil
					}
					id := ctx.Trigger().Object.Ref().ID
					n := 3 * g.CreaturesBlockingForEffect(id)
					return effects.BoostUntilEOT{Target: id, Power: n, Toughness: n,
						Label: "General Marhault Elsdragon — +3/+3 for each creature blocking it"}.Apply(ctx)
				}),
		},
	})
}

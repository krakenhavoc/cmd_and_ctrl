package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kragma Warcaller — Creature — Minotaur Warrior {3}{B}{R}, 2/3 (Theros, S82 OLD b389, #5561):
//
//	"Minotaur creatures you control have haste.
//	 Whenever a Minotaur you control attacks, it gets +2/+0 until end of
//	 turn."
//
// Blur Sliver's tribal haste grant (the Warcaller included) and Miriam's
// attack trigger over a creature type: the pump goes on the attacker that
// triggered it, recorded as the object it was, so one that left the
// battlefield since gets nothing.
//
// No simplification.
func init() {
	const label = "Kragma Warcaller — the attacking Minotaur gets +2/+0 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     "da04b438-aa49-4736-b0ed-131bbe70b566",
		Name:         "Kragma Warcaller",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Minotaur"}, YoursOnly: true}, "haste"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			Key:     label,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if !effects.AttackDeclaredByYou(ev, source.Controller) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.HasSubtype(source.TextSubtype("Minotaur"))
			},
			Build: effects.EventCardTriggerItem(label),
			Effect: func(g *game.Game, item *game.StackItem) error {
				o := item.Params.Object
				c, ok := g.LookupCardForEffect(o.ID)
				if !ok || c.ObjectEpoch != o.Epoch || !effects.OnBattlefield(g, o.ID) {
					return nil
				}
				return effects.BoostUntilEOT{Target: o.ID, Power: 2, Label: "Kragma Warcaller — +2/+0"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

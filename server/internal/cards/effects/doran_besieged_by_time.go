package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Doran, Besieged by Time — Legendary Creature — Treefolk Druid {1}{W}{B}{G}
// (Lorwyn Eclipsed, #4458):
//
//	"Each creature spell you cast with toughness greater than its power
//	 costs {1} less to cast.
//	 Whenever a creature you control attacks or blocks, it gets +X/+X until
//	 end of turn, where X is the difference between its power and toughness."
//
// The reduction takes off generic mana only (the ruling). X is counted once,
// as the trigger resolves, from the creature's power and toughness then
// (CR 608.2h if it has left: nothing to pump). A creature blocking several
// attackers triggers once (CR 509.3a).
func init() {
	Register(Spec{
		OracleID:     "dd35a377-0865-4b25-bd0f-aaead29a9f0e",
		Name:         "Doran, Besieged by Time",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Each creature spell you cast with toughness greater than its power costs {1} less to cast.",
				YourSpell(), CreatureSpell(), toughnessGreaterThanPower),
		},
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventAttack, game.EventBlock},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					if ev.Kind == game.EventAttack {
						return attackDeclaredByYou(ev, source.Controller)
					}
					return ev.Kind == game.EventBlock && ev.Actor == source.Controller && ev.Amount <= 1
				},
				"Doran, Besieged by Time — that creature gets +X/+X until end of turn, X the difference between its power and toughness",
				doranPump),
		},
	})
}

// toughnessGreaterThanPower is the spell's own power and toughness, as it
// stands on the stack.
func toughnessGreaterThanPower(q game.CostQuery) bool {
	return q.Card.CurrentToughness() > q.Card.CurrentPower()
}

func doranPump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	obj := ctx.Trigger().Object
	info, ok := ctx.TriggeringPermanent()
	if obj == nil || !ok || info.Left {
		return nil
	}
	x := info.Power - info.Toughness
	if x < 0 {
		x = -x
	}
	if x == 0 {
		return nil
	}
	return BoostUntilEOT{Target: obj.ID, Power: x, Toughness: x, Label: "Doran — +X/+X until end of turn"}.Apply(ctx)
}

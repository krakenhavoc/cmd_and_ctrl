package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wrathful Red Dragon — Creature — Dragon {3}{R}{R}, 5/5 (EDHREC
// rank 2262):
//
//	"Flying
//	 Whenever a Dragon you control is dealt damage, it deals that much
//	 damage to any target that isn't a Dragon."
//
// The Dragon deck's Stuffy Doll. The trigger watches EventDealDamage
// for a Dragon the controller controls in the Target slot — the
// Wrathful itself included, combat damage and burn alike — and is a
// targeted ability: the controller picks any target that is not a
// Dragon (b21TargetAnyNonDragon: TargetAny minus the subtype) as it
// goes on the stack, and the damaged Dragon deals the event's amount
// to it on resolution. The SOURCE of the reflected damage is the
// damaged Dragon, as printed, so its colour and keywords are what a
// prevention or lifelink check sees; a Dragon that died to the
// damage still deals it (the damage path does not need the source on
// the battlefield).
//
// Sandbox simplification, declared, the Phyrexian Obliterator
// posture: the engine emits one damage event per SOURCE, so a Dragon
// damaged by two blockers at once fires twice — once per blocker's
// damage, each with its own target — where the printed card fires
// once for the total. The same damage is reflected either way; only
// the split differs.
func init() {
	Register(Spec{
		OracleID:        "17c2a962-994a-49f1-8c4e-ebcc43c3c92a",
		Name:            "Wrathful Red Dragon",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"If two or more sources damage a Dragon at the same time, it reflects each source's damage separately instead of the total in one go."},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b21DragonYouControlDealtDamage(ev, source, g)
			},
			Targets: b21TargetAnyNonDragon(),
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				dragon, amount := ev.Target, ev.Amount
				return game.NewTriggeredItem(source, "Wrathful Red Dragon — the damaged Dragon deals that much damage to any non-Dragon target",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 {
							return nil
						}
						return DealDamage{Source: dragon, Target: item.Targets[0].ID, Amount: amount}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}

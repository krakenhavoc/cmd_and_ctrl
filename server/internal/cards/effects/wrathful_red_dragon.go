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
// The engine emits one damage event per SOURCE, so the trigger waits for
// the damage to finish (AtBatchEnd, ThatMuchDamage) and reflects the
// TOTAL the Dragon was dealt at the same time, as printed (CR 120.3, CR 510.2).
func init() {
	Register(Spec{
		OracleID:        "17c2a962-994a-49f1-8c4e-ebcc43c3c92a",
		Name:            "Wrathful Red Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Key:        "Wrathful Red Dragon — the damaged Dragon deals that much damage to any non-Dragon target",
			AtBatchEnd: ThatMuchDamage(),
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b21DragonYouControlDealtDamage(ev, source, g)
			},
			Targets: b21TargetAnyNonDragon(),
			Effect:  damagedCreatureReflectsChosenAmount,
		}},
	})
}

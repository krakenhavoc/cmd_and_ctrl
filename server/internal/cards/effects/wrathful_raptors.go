package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wrathful Raptors — Creature — Dinosaur {4}{R}, 5/5 (EDHREC rank
// 2629):
//
//	"Trample
//	 Whenever a Dinosaur you control is dealt damage, it deals that
//	 much damage to any target that isn't a Dinosaur."
//
// Wrathful Red Dragon for the Dinosaur deck — the same card with the
// type swapped, and the same shape here: the trigger watches
// EventDealDamage for a Dinosaur the controller controls in the
// Target slot (b24SubtypeYouControlDealtDamage — the Raptors
// included, combat damage and burn alike), the controller picks any
// target that is not a Dinosaur (b24TargetAnyNotSubtype) as it goes
// on the stack, and the damaged Dinosaur deals the event's amount to
// it on resolution. The SOURCE of the reflected damage is the
// damaged Dinosaur, as printed, so its colour and keywords are what
// a prevention or lifelink check sees; a Dinosaur that died to the
// damage still deals it. Enrage payoffs fire off the same event.
//
// The engine emits one damage event per SOURCE, so the trigger waits for
// the damage to finish (AtBatchEnd, ThatMuchDamage) and reflects the
// TOTAL the Dinosaur was dealt at the same time, as printed (CR 120.3, CR 510.2).
func init() {
	Register(Spec{
		OracleID:        "1ba92b64-d821-4c24-aed6-fc90d0c63c10",
		Name:            "Wrathful Raptors",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{{
			Watches:    []game.EventKind{game.EventDealDamage},
			Key:        "Wrathful Raptors — the damaged Dinosaur deals that much damage to any non-Dinosaur target",
			AtBatchEnd: ThatMuchDamage(),
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b24SubtypeYouControlDealtDamage(ev, source, g, "Dinosaur")
			},
			Targets: b24TargetAnyNotSubtype("Dinosaur"),
			Effect:  damagedCreatureReflectsChosenAmount,
		}},
	})
}

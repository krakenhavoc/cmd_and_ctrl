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
// Sandbox simplification, declared, the Wrathful Red Dragon posture:
// the engine emits one damage event per SOURCE, so a Dinosaur
// damaged by two blockers at once fires twice — once per blocker's
// damage, each with its own target — where the printed card fires
// once for the total. The same damage is reflected either way; only
// the split differs.
func init() {
	Register(Spec{
		OracleID:        "1ba92b64-d821-4c24-aed6-fc90d0c63c10",
		Name:            "Wrathful Raptors",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"If two or more sources damage a Dinosaur at the same time, it reflects each source's damage separately instead of the total in one go."},
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b24SubtypeYouControlDealtDamage(ev, source, g, "Dinosaur")
			},
			Targets: b24TargetAnyNotSubtype("Dinosaur"),
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				dinosaur, amount := ev.Target, ev.Amount
				return game.NewTriggeredItem(source, "Wrathful Raptors — the damaged Dinosaur deals that much damage to any non-Dinosaur target",
					func(g *game.Game, item *game.StackItem) error {
						if len(item.Targets) == 0 {
							return nil
						}
						return DealDamage{Source: dinosaur, Target: item.Targets[0].ID, Amount: amount}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Javelineer — Creature — Goblin Warrior {R}, 1/1 (AFR):
//
//	"Haste
//	 Whenever this creature becomes blocked, it deals 1 damage to target
//	 creature blocking it."
//
// The trigger fires once however many creatures block (CR 509.1h makes the
// attacker "blocked" once); its target is one of the blockers.
func init() {
	Register(Spec{
		OracleID:        "e35c8206-98d0-44ab-8770-823a5ff7ab72",
		Name:            "Goblin Javelineer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBecomesBlocked},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: BlockingSource(TargetCreature("target creature blocking it")),
			Key:     "Goblin Javelineer — 1 damage to target creature blocking it",
			Effect:  arcSpitterPing,
		}},
	})
}

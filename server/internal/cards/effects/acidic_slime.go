package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Acidic Slime — 2/2 Ooze for {3}{G}{G}:
//
//	"Deathtouch. When Acidic Slime enters the battlefield, destroy
//	target artifact, enchantment, or land."
//
// S19 sub-PR 3 shipped this with a sandbox auto-picker; S20 sub-PR
// 2 gives it a real target clause. Mandatory: no "you may", so the
// pick_target prompt queues straight away when the Slime enters
// (and the trigger is removed if nothing qualifies, CR 603.3d).
func init() {
	Register(Spec{
		OracleID:        "21f45043-5419-4019-8b6c-e5294bd5f549",
		Name:            "Acidic Slime",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetPermanent("target artifact, enchantment, or land", Or(Artifact(), Enchantment(), Land())),
			Key:     "Acidic Slime — destroy target permanent",
			Effect:  destroyChosenPermanent,
		}},
	})
}

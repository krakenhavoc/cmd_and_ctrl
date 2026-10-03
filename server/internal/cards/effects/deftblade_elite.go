package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deftblade Elite — Creature — Human Soldier {W}, 1/1:
//
//	"Provoke
//	 {1}{W}: Prevent all combat damage that would be dealt to and dealt by
//	 this creature this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): provoke is the catalog's trigger
// (Goblin Grappler's), and the ability is one to-and-by record
// (Mod.AndDealtBy) pinned to this creature.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "9a3e3a8f-7c19-4b52-a1b4-08eb09f6127a",
		Name:         "Deftblade Elite",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Provoke()},
		Activated: []ActivatedAbility{toAndByThisRow(
			"{1}{W}: Prevent all combat damage that would be dealt to and dealt by this creature this turn.",
			ManaCost("{1}{W}"))},
	})
}

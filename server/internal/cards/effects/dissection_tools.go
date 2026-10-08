package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dissection Tools — Artifact — Equipment {5}:
//
//	"When this Equipment enters, manifest dread, then attach this
//	 Equipment to that creature.
//	 Equipped creature gets +2/+2 and has deathtouch and lifelink.
//	 Equip—Sacrifice a creature"
//
// "Attach this Equipment to that creature" is the continuation of the
// manifest dread, so it names the permanent that entered face down
// rather than the top of the library.
//
// The equip cost is a sacrifice, not mana: it is EquipAbility with its
// cost replaced, so it keeps the sorcery-speed gate, the "creature you
// control" target and the Equip marker that Leonin Shikari reads.
//
// The attach is skipped quietly when the Equipment is no longer the
// permanent that asked by the time the manifest settles (CR 701.3b),
// and when nothing was manifested (an empty library).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "50de5faa-d22e-4909-a1ff-3810cfcfcb1a",
		Name:         "Dissection Tools",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 2), GrantToAttached("deathtouch", "lifelink")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self, "Dissection Tools — manifest dread, then attach this Equipment to that creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ManifestDread{Then: AttachSourceToManifested(ctx)}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{
			dissectionToolsEquip(),
		},
	})
}

// dissectionToolsEquip is "Equip—Sacrifice a creature".
func dissectionToolsEquip() ActivatedAbility {
	ab := EquipAbility("{0}")
	ab.Label = "Equip—Sacrifice a creature"
	ab.Cost = SacrificeACreature()
	return ab
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Killer's Mask — Artifact — Equipment {2}{B}:
//
//	"When this Equipment enters, manifest dread, then attach this
//	 Equipment to that creature.
//	 Equipped creature has menace.
//	 Equip {2}"
//
// "Attach this Equipment to that creature" is the continuation of the
// manifest dread, so it names the permanent that entered face down
// rather than the top of the library.
//
// The attach is skipped quietly when the Equipment is no longer the
// permanent that asked by the time the manifest settles (CR 701.3b),
// and when nothing was manifested (an empty library).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "93a04ce0-9b8e-4d69-82bf-d4c89a9856a9",
		Name:         "Killer's Mask",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{GrantToAttached("menace")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self, "Killer's Mask — manifest dread, then attach this Equipment to that creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ManifestDread{Then: AttachSourceToManifested(ctx)}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}

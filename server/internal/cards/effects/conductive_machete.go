package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Conductive Machete — Artifact — Equipment {4}:
//
//	"When this Equipment enters, manifest dread, then attach this
//	 Equipment to that creature.
//	 Equipped creature gets +2/+1.
//	 Equip {4}"
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
		OracleID:     "513dbf8a-a250-4fd9-a5ae-2fb9c4810ecc",
		Name:         "Conductive Machete",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(2, 1)},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self, "Conductive Machete — manifest dread, then attach this Equipment to that creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ManifestDread{Then: AttachSourceToManifested(ctx)}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{4}"),
		},
	})
}

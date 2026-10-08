package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cursed Windbreaker — Artifact — Equipment {2}{U}:
//
//	"When this Equipment enters, manifest dread, then attach this
//	 Equipment to that creature.
//	 Equipped creature has flying.
//	 Equip {3}"
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
		OracleID:     "8a265910-2e10-4e97-88ab-c0cb1a712eb8",
		Name:         "Cursed Windbreaker",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{GrantToAttached("flying")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, Self, "Cursed Windbreaker — manifest dread, then attach this Equipment to that creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return ManifestDread{Then: AttachSourceToManifested(ctx)}.Apply(ctx)
				}),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{3}"),
		},
	})
}

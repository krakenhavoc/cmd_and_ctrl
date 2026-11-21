package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "8a265910-2e10-4e97-88ab-c0cb1a712eb8",
		Name:         "Cursed Windbreaker",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.GrantToAttached("flying")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.Self, "Cursed Windbreaker — manifest dread, then attach this Equipment to that creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.ManifestDread{ThenTail: effects.AttachSourceToManifested(ctx)}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{3}"),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skullclamp — Artifact — Equipment for {1} (top-100 Commander
// staples rank 40):
//
//	"Equipped creature gets +1/-1.
//	 Whenever equipped creature dies, draw two cards.
//	 Equip {1}"
//
// The card that got banned in Mirrodin block, and the one that makes
// the attachment relation worth building: the -1 toughness is not a
// drawback, it is the engine. Clamp a 1/1 token, the lethal-damage
// state-based action kills it, draw two, re-equip.
//
// That loop only works if three separate pieces of ordering are
// right, and all three are:
//
//   - The 7c static has to reach the toughness the SBA reads.
//     CurrentToughness reads Effective().Toughness, and the SBA runs
//     RecomputeLayersIfStaleLocked before its destruction pre-pass,
//     so a 1/1 under a Skullclamp is already a 2/0 by the time
//     704.5f looks at it.
//   - The dies trigger has to still see the attachment. The trigger
//     harvester runs synchronously inside the LTB emit, and the
//     CR 704.5m unattach is a state-based action that has not run
//     yet — so AttachedTo still names the creature that just died.
//     See equippedCreatureDied for the full ordering note.
//   - Re-equipping has to move the bonus. AttachedToSource is
//     re-evaluated on every recompute rather than captured, and
//     EventAttach bumps the layer version, so it does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "65986c1b-8e51-4604-b685-d82fa7d1263a",
		Name:     "Skullclamp",
		Static:   []game.StaticAbility{PumpAttached(1, -1)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return equippedCreatureDied(ev, source)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Skullclamp — draw two cards",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 2}.Apply(NewContext(g, item))
					})
			},
		}},
		Activated: []ActivatedAbility{
			EquipAbility("{1}"),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ori, Plate Stacker — Legendary Creature — Dwarf Bard {5}{W}{W}, 3/3
// (The Hobbit):
//
//	"When Ori enters, destroy all artifacts and enchantments your opponents
//	 control. You gain 1 life for each permanent destroyed this way."
//
// Ob Nixilis, the Ascended's shape: one sweep over the opponents' permanents
// (an artifact enchantment is destroyed once and counted once) whose keyed
// tail gains 1 life for each permanent that actually went (massGainLifeForEach),
// so a permanent that was indestructible, regenerated or exiled instead is not
// counted (CR 701.7a). Ori's own sweep never touches his own side.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7fc9f7e3-5347-4ea1-9bec-9e04ed78d16a",
		Name:         "Ori, Plate Stacker",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifactsAndEnchantments, How: game.SweepDestroy, OpponentsOnly: true}},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ori, Plate Stacker — destroy all artifacts and enchantments your opponents control; gain 1 life for each destroyed",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return DestroyAllMatching{
						Match:    And(Or(Artifact(), Enchantment()), OpponentControls()),
						ThenTail: ctx.Tail(massGainLifeForEach),
					}.Apply(ctx)
				}),
		},
	})
}

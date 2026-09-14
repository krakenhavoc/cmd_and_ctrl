package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bane of Progress — Creature — Elemental, {4}{G}{G}, 2/2:
//
//	"When this creature enters, destroy all artifacts and
//	 enchantments. Put a +1/+1 counter on this creature for each
//	 permanent destroyed this way."
//
// Green's answer to the entire artifact-and-enchantment half of the
// format, on a body that grows to match the problem it solved. At a
// Commander table with four mana bases' worth of rocks and signets
// it lands as a 12/12.
//
// # "For each permanent destroyed this way" is the count
//
// This is the second reason DestroyAllMatching returns a number
// (Fumigate is the first). The counter total is not "artifacts and
// enchantments on the battlefield" and not "cards in graveyards" —
// it is what this sweep actually destroyed, which is exactly what
// the Then callback is handed.
//
// # It survives its own trigger
//
// The Bane is a creature, not an artifact or enchantment, so the
// sweep never touches it and the counters have somewhere to go. A
// Bane that somehow WAS an artifact (a Layer-4 effect, Mycosynth
// Lattice) would destroy itself and then have nothing to put
// counters on — AddCounter on a card that has left the battlefield
// is a silent no-op, which is the right outcome and needs no guard.
//
// One sweep over Or(Artifact(), Enchantment()) rather than two, so
// an artifact enchantment is destroyed once and counted once.
func init() {
	Register(Spec{
		OracleID:     "51f9a6cc-8eb2-44ed-a2d9-913ac514ad67",
		Name:         "Bane of Progress",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				self := source.InstanceID
				return game.NewTriggeredItem(source, "Bane of Progress — destroy all artifacts and enchantments",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						return DestroyAllMatching{
							Match: Or(Artifact(), Enchantment()),
							Then: func(ctx *Context, _ []game.Card, destroyed int) error {
								if destroyed <= 0 {
									return nil
								}
								return AddCounter{
									Target: self,
									Kind:   game.CounterPlusOne,
									N:      destroyed,
								}.Apply(ctx)
							},
						}.Apply(ctx)
					})
			},
		}},
	})
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Phyrexian Devourer — Artifact Creature — Phyrexian Construct {6}, 1/1:
//
//	"When this creature's power is 7 or greater, sacrifice it.
//	 Exile the top card of your library: Put X +1/+1 counters on this
//	 creature, where X is the exiled card's mana value."
//
// ADR 0109 §7 (#1902): "Exile the top N cards of your library" is a
// cost with nothing to choose. A library of fewer than N cards can't pay
// it (CR 118.3), and it is paid after every other cost (CR 601.2h).
// X is the exiled card's mana value, read off the payment record
// (Context.Exiled, CR 400.7j) as the ability resolves.
//
// The sacrifice is ADR 0107's CR 603.8 state trigger: it triggers as
// soon as the Devourer's power is 7 or more, counters and layers
// included, and not again while it waits or is on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "28b1ed69-784c-4642-86c7-104792d5afc2",
		Name:         "Phyrexian Devourer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenState("Phyrexian Devourer — sacrifice it", func(_ *game.Game, source *game.Card, _ uuid.UUID) bool {
				return source.IsCreature() && source.PowerForComparison() >= 7
			}, SacrificeThisIfStillOnBattlefield),
		},
		Activated: []ActivatedAbility{{
			Label: "Exile the top card of your library: Put X +1/+1 counters on this creature, where X is the exiled card's mana value.",
			Cost:  ExileTopOfLibrary(1),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				x := 0
				for _, id := range ctx.Exiled() {
					if c, ok := g.LookupCardForEffect(id); ok {
						x, _ = g.ManaValueForEffect(c)
					}
				}
				if x <= 0 {
					return nil
				}
				return AddCounter{Target: item.SourceCardID, Kind: "+1/+1", N: x}.Apply(ctx)
			},
		}},
	})
}

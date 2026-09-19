package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wolfbriar Elemental — {2}{G}{G} 4/4 Elemental:
//
//	"Multikicker {G} (You may pay an additional {G} any number of
//	 times as you cast this spell.)
//	 When this creature enters, create a 2/2 green Wolf creature token
//	 for each time it was kicked."
//
// The multikicker proof (CR 702.33d), and the harder half of ADR 0073
// §5: the COUNT is read by a trigger on the permanent, not by the
// spell's own resolution.
//
// That is why the record is carried onto the permanent rather than
// looked up on the stack. A TriggeredAbility's Build receives the
// game, the source card and the event — the spell's StackItem is out
// of StackMeta before the ETB event is even emitted — so if the
// resolution path did not stamp Card.PaidOptionalCosts before
// emitting it, this trigger would have nothing to count and every
// Wolfbriar would make zero Wolves.
//
// game.CardKickedTimes reads that stamp. It is per-instance and
// cleared on the way out (CR 400.7), so a Wolfbriar that dies and is
// reanimated makes no Wolves: the spell that returned it was not the
// spell that was kicked.
//
// The multikicker cap is the engine's, not the card's — paper
// multikicker is unbounded, but an announcement has to be finite and
// a bot's expansion has to terminate. 20 is far above what any board
// in this format can pay for.
func init() {
	Register(Spec{
		OracleID:     "2f8872fe-84dc-4cda-a253-e7503a5c96a3",
		Name:         "Wolfbriar Elemental",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			Multikicker("{G}", 20),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Wolfbriar Elemental — a Wolf for each time it was kicked",
				func(g *game.Game, item *game.StackItem) error {
					src, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok {
						return nil
					}
					n := game.CardKickedTimes(src)
					if n <= 0 {
						return nil
					}
					return CreateToken{
						Controller: item.Controller,
						Template:   TokenCard("2/2 green Wolf"),
						N:          n,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

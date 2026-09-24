package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spark Double — "You may have this creature enter as a copy of a
// creature or planeswalker you control, except it enters with an
// additional +1/+1 counter on it if it's a creature, it enters with
// an additional loyalty counter on it if it's a planeswalker, and it
// isn't legendary."
//
// Three exceptions, and each one lands somewhere different:
//
//   - "isn't legendary" REMOVES a supertype, which is what makes
//     copying your own commander legal (CR 704.5j would otherwise
//     put one of them in the graveyard immediately). This is the
//     card's whole reason for existing in the format.
//   - the additional +1/+1 counter is not a copiable value at all —
//     it changes how the permanent ENTERS, so it goes on the entry
//     event via AddCounterAtETB and arrives before the ETB trigger,
//     the way Hangarback Walker's do.
//   - the additional LOYALTY counter goes through the copied
//     starting loyalty rather than the same route, because the
//     CR 306.5b stamp refuses to run on a walker that already has
//     loyalty counters (effect_hooks.go) and would otherwise leave
//     a copied Teferi on 1. Bumping the printed value it stamps
//     from gets the same total, and keeps the CR 122.6 counter
//     doubling (Doubling Season) applying to the whole of it,
//     which is the ruling.
func init() {
	Register(Spec{
		OracleID:     "8dcb35e5-ae44-455f-86e3-4a77d496ff34",
		Name:         "Spark Double",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOf(
				"Spark Double",
				func(g *game.Game, controller uuid.UUID, self uuid.UUID) []uuid.UUID {
					return copyCandidates(g, self, func(c game.Card) bool {
						if c.Controller != controller {
							return false
						}
						return c.IsCreature() || c.IsPlaneswalker()
					})
				},
				func(ev *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.RemoveSupertype("Legendary")
					if v.HasCardType("Creature") {
						ev.AddCounterAtETB("+1/+1", 1)
					}
					if v.HasCardType("Planeswalker") {
						v.StartingLoyalty++
					}
				},
			),
		},
	})
}

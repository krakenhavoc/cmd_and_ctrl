package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pharika's Spawn — Creature — Gorgon {3}{B}, 3/4:
//
//	"Escape—{5}{B}, Exile three other cards from your graveyard.
//	 This creature escapes with two +1/+1 counters on it. When it
//	 enters this way, each opponent sacrifices a non-Gorgon creature
//	 of their choice."
//
// The other half of #653, and the reason the provenance record is
// about the ENTRY rather than about the cast: "when it enters THIS
// WAY" is CR 400.7d asked of an ordinary enters trigger. A hard-cast
// Pharika's Spawn is a 3/4 that does nothing; an escaped one arrives
// as a 5/6 and takes a creature off every opponent.
//
// The counters and the trigger reach the permanent by different
// routes, and both are right. The two +1/+1 counters ride the escape
// COST (EscapeWithCounters → AlternativeCost.EntersWithCounterName),
// so they go on through the CR 614 entry pipeline and Doubling Season
// doubles them. The sacrifice is a printed triggered ability that
// happens to be conditional, so it is an ordinary WhenThisEnters whose
// body asks ctx.Escaped() — the permanent's own record, written before
// EventETB fired.
//
// "of their choice" is the default: EachPlayerSacrifices queues one
// picker per opponent and each of them chooses their own (CR 701.21a).
// "non-Gorgon" narrows what may be picked rather than what may be
// sacrificed, so an opponent whose only creature is a Gorgon
// sacrifices nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "a44955e7-f1ad-41b9-b93d-0a980ac341d8",
		Name:             "Pharika's Spawn",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{EscapeWithCounters("{5}{B}", 3, 2)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pharika's Spawn — each opponent sacrifices a non-Gorgon creature", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if !ctx.Escaped() {
					return nil
				}
				return EachPlayerSacrifices{
					ExceptController: true,
					Match:            And(Creature(), Not(OfSubtype("Gorgon"))),
					Label:            "a non-Gorgon creature",
				}.Apply(ctx)
			}),
		},
	})
}

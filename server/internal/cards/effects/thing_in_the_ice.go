package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thing in the Ice // Awoken Horror — {1}{U} Creature — Horror 0/4 //
// Creature — Kraken Horror 7/8 (Innistrad Remastered, S82 GAP, #5948):
//
//	Front: "Defender
//	        This creature enters with four ice counters on it.
//	        Whenever you cast an instant or sorcery spell, remove an ice
//	        counter from this creature. Then if it has no ice counters on
//	        it, transform it."
//	Back:  "When this creature transforms into Awoken Horror, return all
//	        non-Horror creatures to their owners' hands."
//
// Dark Depths's ice counters as an entry replacement (CR 614.1c). The cast
// trigger removes one counter and then, as a separate instruction rather
// than an intervening "if", looks at whether it has none left (so a Thing
// that somehow has none still turns over); the transform is the permanent
// turning over in place. The bounce is one simultaneous return of every
// creature on the battlefield that is not a Horror, the Awoken Horror
// itself excluded (Sunderflock's shape).
//
// No simplification.
func init() {
	const oracle = "f2c10591-4c0c-4286-af91-e8d8d10a4c9d"
	const ice = "ice"
	Register(Spec{
		OracleID:        oracle,
		Name:            "Thing in the Ice",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Replacements:    []game.ReplacementEffect{b10EntersWithCounters(ice, 4, "Enters with four ice counters on it.")},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()), "Thing in the Ice — remove an ice counter; then if it has no ice counters, transform it",
				func(g *game.Game, item *game.StackItem) error {
					if err := removeACounterFromThis(ice)(g, item); err != nil {
						return err
					}
					c, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || !onBattlefield(g, item.SourceCardID) || c.Counters[ice] > 0 {
						return nil
					}
					return TransformThis{}.Apply(NewContext(g, item))
				}),
		},
	})
	Register(Spec{
		OracleID:     oracle + "#1",
		Name:         "Awoken Horror",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(On(game.EventTransform, ThisTransformedInto(1),
				"Awoken Horror — return all non-Horror creatures to their owners' hands",
				Do(BounceAllMatching{Match: And(Creature(), Not(HasSubtype("Horror")))})),
				game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}}),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skycoach Conductor // All Aboard — Creature — Bird Pilot {2}{U}, 2/3
// // Instant {U} (preparation card, CR 722):
//
//	"Flash
//	 Flying, vigilance
//	 This creature enters prepared. (While it's prepared, you may cast a
//	 copy of its spell. Doing so unprepares it.)"
//
//	All Aboard — "Exile target non-Pilot creature you control, then
//	return that card to the battlefield under its owner's control."
//
// The card #1328 was filed for (deck tracker #1306, "Aang is so
// flashy"), and the first preparation card in the catalog (ADR 0090).
//
// Two registrations, the adventure shape. Face 0 is the creature:
// flash, flying and vigilance are printed keywords, and "enters
// prepared" is SelfEntersPrepared — a CR 614.1d replacement on its own
// entry, so the landing makes it prepared and puts the copy of All
// Aboard into exile before anything sees it arrive. Face 1 is All
// Aboard, registered under "#1" and never cast from the card
// (CR 722.3): the engine casts the COPY out of exile, which resolves
// through this entry, ceases to exist as it leaves the stack, and
// unprepares the Conductor as it becomes cast (CR 601.2i).
//
// The blink is Flicker with no controller named, which is "under its
// owner's control", and it returns a new object (CR 400.7) — the
// point of the card, re-buying an ETB. "Non-Pilot" is a creature-type
// predicate, so the Conductor, a Pilot, cannot blink itself to prepare
// again; a changeling is every creature type and is refused too.
//
// No simplification.
const skycoachConductorOracleID = "788d3cfa-7706-4728-9c48-cf7bc963d002"

func init() {
	Register(Spec{
		OracleID:        skycoachConductorOracleID,
		Name:            "Skycoach Conductor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying", "vigilance"},
		Replacements:    []game.ReplacementEffect{SelfEntersPrepared()},
	})
	Register(Spec{
		OracleID:     skycoachConductorOracleID + "#1",
		Name:         "All Aboard",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target non-Pilot creature you control", YouControl(), Not(OfCreatureType("Pilot"))),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				return Flicker{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}

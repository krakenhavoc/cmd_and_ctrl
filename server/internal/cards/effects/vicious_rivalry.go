package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vicious Rivalry — Sorcery {2}{B}{G} (Secrets of Strixhaven, #4456):
//
//	"As an additional cost to cast this spell, pay X life.
//	 Destroy all artifacts and creatures with mana value X or less."
//
// Toxic Deluge's additional cost (PayXLifeCost, CR 601.2f: paid with the spell
// on the stack, and paid even if it is countered), then Gaze of Granite's
// sweep over mana value X or less. The X of the life and of the sweep are the
// same announced number. The sweep goes through DestroyAllMatching, so the
// deaths are simultaneous and an indestructible permanent survives. It reads
// each permanent's mana value on the battlefield, where an {X} in a cost
// counts as 0 (CR 202.3e) and a token has mana value 0, so X = 0 still
// destroys tokens and free permanents. It hits the caster's own permanents
// too, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "90898c76-21f9-4597-88fe-dc9179911cde",
		Name:     "Vicious Rivalry",
		// ADR 0126 §6: only artifacts and creatures with mana value X or less.
		Purpose:        game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepDestroy, Partial: true}},
		Completeness:   CompletenessFull,
		XMatters:       true,
		AdditionalCost: PayXLifeCost(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Or(Artifact(), Creature()), ManaValueLE(ctx.X()))}.Apply(ctx)
		},
	})
}

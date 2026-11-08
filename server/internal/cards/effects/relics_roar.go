package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Relic's Roar — Instant {U} (The Lost Caverns of Ixalan, #4470):
//
//	"Until end of turn, target artifact or creature becomes a Dinosaur
//	 artifact creature with base power and toughness 4/3 in addition to
//	 its other types."
//
// Behind the Mask's animation (BecomeCreatureUntilEOT) with the Dinosaur
// creature type added beside the artifact and creature types at layer 4,
// and the base 4/3 set at layer 7b, pinned to the object, so counters and
// pumps still apply on top. It keeps its other types and subtypes.
// "Dinosaur" is a word of the card's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7887f060-3024-4dc4-aeb5-fb5546db6b48",
		Name:         "Relic's Roar",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return BecomeCreatureUntilEOT{
				Target:       t.ID,
				Subtypes:     ctx.TextSubtypes("Dinosaur"),
				SetPower:     4,
				SetToughness: 3,
				Label:        "Relic's Roar — a 4/3 Dinosaur artifact creature until end of turn",
			}.Apply(ctx)
		},
	})
}

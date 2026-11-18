package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Suit Up — Instant {2}{U} (Neon Dynasty):
//
//	"Until end of turn, target creature or Vehicle becomes an artifact
//	 creature with base power and toughness 4/5.
//	 Draw a card."
//
// One layer-4 type SET (Kitesail Larcenist's SetTypesMod, CR 205.1a: a
// target that is a land or enchantment creature stops being one, and the
// subtypes that belonged to the old types go with it; a Vehicle keeps
// Vehicle and a creature keeps its creature types) and the layer-7b 4/5 at
// one timestamp. Then you draw, unless the target was illegal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "29bbbbfe-77cb-411f-9c1b-f7e4724685c6",
		Name:         "Suit Up",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      TargetPermanent("target creature or Vehicle", Or(Creature(), Subtype("Vehicle"))),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				mods := append([]game.Mod{game.SetTypesMod([]string{"Artifact", "Creature"}, nil)}, game.SetBasePTMods(4, 5)...)
				if err := (ScopedEffectFor{
					Target:   t.ID,
					Mods:     mods,
					Duration: DurationUntilEndOfTurn(ctx),
					Label:    "Suit Up — becomes a 4/5 artifact creature until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				break
			}
			return DrawCards{N: 1}.Apply(ctx)
		},
	})
}

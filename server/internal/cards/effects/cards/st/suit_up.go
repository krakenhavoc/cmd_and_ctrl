package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "29bbbbfe-77cb-411f-9c1b-f7e4724685c6",
		Name:         "Suit Up",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetPermanent("target creature or Vehicle", effects.Or(effects.Creature(), effects.Subtype("Vehicle"))),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				mods := append([]game.Mod{game.SetTypesMod([]string{"Artifact", "Creature"}, nil)}, game.SetBasePTMods(4, 5)...)
				if err := (effects.ScopedEffectFor{
					Target:   t.ID,
					Mods:     mods,
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Suit Up — becomes a 4/5 artifact creature until end of turn",
				}).Apply(ctx); err != nil {
					return err
				}
				break
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}

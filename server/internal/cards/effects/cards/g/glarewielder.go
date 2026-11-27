package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glarewielder — Creature — Elemental Shaman {4}{R}, 3/1 (Lorwyn, S82 OLD,
// #5593):
//
//	"Haste
//	 When this creature enters, up to two target creatures can't block this turn.
//	 Evoke {1}{R}"
//
// Markov Warlord's enters trigger (each target still legal as it resolves can't
// block until end of turn) under Mulldrifter's evoke.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "9d3136da-3efa-4103-9d35-8da87d318544",
		Name:             "Glarewielder",
		Completeness:     effects.CompletenessFull,
		PrintedKeywords:  []string{"haste"},
		AlternativeCosts: []game.AlternativeCost{effects.Evoke("{1}{R}")},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.Targeting(
				effects.WhenThisEnters("Glarewielder — up to two target creatures can't block this turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							if err := (effects.RestrictUntilEOT{Target: t.ID, Restrictions: game.CantBlock, Label: "Glarewielder — can't block"}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					}),
				effects.TargetCreature("up to two target creatures").WithCount(0, 2)),
				effects.ForTargets(effects.CantBlockUntilEOT(0, 0, 0))),
		},
	})
}

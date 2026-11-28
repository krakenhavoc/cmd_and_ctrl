package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Xanthic Statue — Artifact {8} (Weatherlight):
//
//	"{5}: Until end of turn, this artifact becomes an 8/8 Golem artifact
//	 creature with trample."
//
// The animation adds the Creature type and the Golem subtype (layer 4), trample
// (layer 6) and a base 8/8 (layer 7b); it stays an artifact.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a8de92c4-84cb-40e7-8e34-428380726ed4",
		Name:         "Xanthic Statue",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}: Until end of turn, this artifact becomes an 8/8 Golem artifact creature with trample.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{5}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Artifact", "Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Golem")...),
						game.AddKeywordsMod("trample"),
					}, game.SetBasePTMods(8, 8)...),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Xanthic Statue — an 8/8 Golem artifact creature with trample until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}

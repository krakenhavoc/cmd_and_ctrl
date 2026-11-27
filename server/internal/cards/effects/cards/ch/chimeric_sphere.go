package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chimeric Sphere — Artifact {3} (Weatherlight):
//
//	"{2}: Until end of turn, this artifact becomes a 2/1 Construct artifact
//	 creature with flying.
//	 {2}: Until end of turn, this artifact becomes a 3/2 Construct artifact
//	 creature and loses flying."
//
// Each activation is its own effect with its own timestamp (CR 613.7), so the
// later one wins where they disagree: the base P/T is the last one set, and
// "loses flying" removes flying granted by an earlier activation while a flying
// grant activated after it gives it back.
//
// No simplification.
func init() {
	animate := func(label string, power, toughness int, mods ...game.Mod) func(*game.Game, *game.StackItem) error {
		return func(g *game.Game, item *game.StackItem) error {
			if !effects.SourceIsStillThisPermanent(g, item) {
				return nil
			}
			ctx := effects.NewContext(g, item)
			base := []game.Mod{
				game.AddTypesMod("Artifact", "Creature"),
				game.AddSubtypesMod(ctx.TextSubtypes("Construct")...),
			}
			base = append(base, mods...)
			return effects.ScopedEffectFor{
				Target:   item.SourceCardID,
				Mods:     append(base, game.SetBasePTMods(power, toughness)...),
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    label,
			}.Apply(ctx)
		}
	}
	effects.Register(effects.Spec{
		OracleID:     "b4aae26a-44a4-4ef0-9a1d-14a07a0865a3",
		Name:         "Chimeric Sphere",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}: Until end of turn, this artifact becomes a 2/1 Construct artifact creature with flying.",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    effects.ManaCost("{2}"),
				Effect: animate("Chimeric Sphere — a 2/1 Construct artifact creature with flying until end of turn", 2, 1,
					game.AddKeywordsMod("flying")),
			},
			{
				Label:   "{2}: Until end of turn, this artifact becomes a 3/2 Construct artifact creature and loses flying.",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    effects.ManaCost("{2}"),
				Effect: animate("Chimeric Sphere — a 3/2 Construct artifact creature that loses flying until end of turn", 3, 2,
					game.RemoveKeywordsMod("flying")),
			},
		},
	})
}

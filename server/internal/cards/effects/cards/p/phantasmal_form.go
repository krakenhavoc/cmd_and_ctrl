package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phantasmal Form — Instant {2}{U} (Foundations Jumpstart):
//
//	"Until end of turn, up to two target creatures each have base power
//	 and toughness 3/3, gain flying, and become blue Illusions in addition
//	 to their other colors and types.
//	 Draw a card."
//
// One continuous effect per target, all at one timestamp: the Illusion
// subtype (layer 4), blue (layer 5), flying (layer 6) and base 3/3
// (layer 7b), the way Ghouls' Night Out builds its Zombie. "In addition
// to" keeps each creature's own colours and types (CR 205.1b). It ends in
// the cleanup step (CR 514.2). The draw happens even with no target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d21facb6-27c2-49cb-9f13-7d6079096eda",
		Name:         "Phantasmal Form",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets:      effects.TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					mods := append(game.SetBasePTMods(3, 3),
						game.AddSubtypesMod(ctx.TextSubtypes("Illusion")...),
						game.AddColorsMod(ctx.TextColors("U")...),
						game.AddKeywordsMod("flying"))
					if err := (effects.ScopedEffectFor{
						Target:   t.ID,
						Mods:     mods,
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Phantasmal Form — a blue Illusion with base 3/3 and flying",
					}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			}),
			effects.DrawCards{N: 1},
		),
	})
}

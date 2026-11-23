package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardian Idol — Artifact {2} (Foundations Jumpstart):
//
//	"This artifact enters tapped.
//	 {T}: Add {C}.
//	 {2}: This artifact becomes a 2/2 Golem artifact creature until end of
//	 turn."
//
// Mishra's Factory's animation (CR 611.2a, layers 4 and 7b) on a rock that
// enters tapped (Abraded Bluffs' SelfEntersTapped).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "829d2e11-192f-4093-99f8-e13767b321a2",
		Name:         "Guardian Idol",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}: This artifact becomes a 2/2 Golem artifact creature until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{2}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				return effects.ScopedEffectFor{
					Target: item.SourceCardID,
					Mods: append([]game.Mod{
						game.AddTypesMod("Creature"),
						game.AddSubtypesMod(ctx.TextSubtypes("Golem")...),
					}, game.SetBasePTMods(2, 2)...),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Guardian Idol — a 2/2 Golem artifact creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}

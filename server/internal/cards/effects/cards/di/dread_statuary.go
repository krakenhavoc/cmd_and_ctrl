package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dread Statuary — Land (Conspiracy: Take the Crown, S82 OLD, #5540):
//
//	"{T}: Add {C}.
//	 {4}: This land becomes a 4/2 Golem artifact creature until end of
//	 turn. It's still a land."
//
// Mishra's Factory's animation (CR 611.2a, layers 4 and 7b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a9789ce4-69cf-435c-b99a-78a21609830c",
		Name:         "Dread Statuary",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}: This land becomes a 4/2 Golem artifact creature until end of turn. It's still a land.",
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Cost:    effects.ManaCost("{4}"),
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
					}, game.SetBasePTMods(4, 2)...),
					Duration: effects.DurationUntilEndOfTurn(ctx),
					Label:    "Dread Statuary — a 4/2 Golem artifact creature until end of turn",
				}.Apply(ctx)
			},
		}},
	})
}

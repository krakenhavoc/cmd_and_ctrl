package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mishra's Factory — Land:
//
//	"{T}: Add {C}.
//	 {1}: This land becomes a 2/2 Assembly-Worker artifact creature until
//	 end of turn. It's still a land.
//	 {T}: Target Assembly-Worker creature gets +1/+1 until end of turn."
//
// Mishra's Foundry's animation (CR 611.2a, layers 4 and 7b). The pump
// targets any Assembly-Worker creature, itself included once animated; it is
// checked as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5963e0ef-e0bc-4611-ad4f-813a4c0eacfb",
		Name:         "Mishra's Factory",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{1}: This land becomes a 2/2 Assembly-Worker artifact creature until end of turn. It's still a land.",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    effects.ManaCost("{1}"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !effects.SourceIsStillThisPermanent(g, item) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					return effects.ScopedEffectFor{
						Target: item.SourceCardID,
						Mods: append([]game.Mod{
							game.AddTypesMod("Artifact", "Creature"),
							game.AddSubtypesMod(ctx.TextSubtypes("Assembly-Worker")...),
						}, game.SetBasePTMods(2, 2)...),
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Mishra's Factory — a 2/2 Assembly-Worker artifact creature until end of turn",
					}.Apply(ctx)
				},
			},
			{
				Label:   "{T}: Target Assembly-Worker creature gets +1/+1 until end of turn.",
				Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)), game.AnswerPump),
				Cost:    game.AbilityCost{Tap: true},
				Targets: effects.TargetCreature("target Assembly-Worker creature", effects.Subtype("Assembly-Worker")),
				Effect:  effects.BoostTargetUntilEOT("Mishra's Factory — +1/+1", 1, 1),
			},
		},
	})
}

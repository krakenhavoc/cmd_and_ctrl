package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mishra's Foundry — Land (The Brothers' War, #5482):
//
//	"{T}: Add {C}.
//	 {2}: This land becomes a 2/2 Assembly-Worker artifact creature until
//	 end of turn. It's still a land.
//	 {1}, {T}: Target attacking Assembly-Worker gets +2/+2 until end of
//	 turn."
//
// A manland built like Mutavault: the animation adds the Artifact and
// Creature types and the Assembly-Worker subtype (layer 4) and sets the
// base power and toughness (layer 7b) until end of turn (CR 611.2a). The
// land stays a land with its mana ability. The pump may target the
// Foundry itself once it is an attacking Assembly-Worker, tapping it as
// a cost (an attacking creature that taps stays attacking, CR 506.4).
// The target is checked as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b43e9772-6ad4-49c7-9557-b18ee1e4587d",
		Name:         "Mishra's Foundry",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{2}: This land becomes a 2/2 Assembly-Worker artifact creature until end of turn. It's still a land.",
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
							game.AddTypesMod("Artifact", "Creature"),
							game.AddSubtypesMod(ctx.TextSubtypes("Assembly-Worker")...),
						}, game.SetBasePTMods(2, 2)...),
						Duration: effects.DurationUntilEndOfTurn(ctx),
						Label:    "Mishra's Foundry — a 2/2 Assembly-Worker artifact creature until end of turn",
					}.Apply(ctx)
				},
			},
			{
				Label:   "{1}, {T}: Target attacking Assembly-Worker gets +2/+2 until end of turn.",
				Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)), game.AnswerPump),
				Cost:    game.AbilityCost{Mana: "{1}", Tap: true},
				Targets: effects.TargetCreature("target attacking Assembly-Worker", effects.AttackingCreature(), effects.Subtype("Assembly-Worker")),
				Effect:  effects.BoostTargetUntilEOT("Mishra's Foundry — +2/+2", 2, 2),
			},
		},
	})
}

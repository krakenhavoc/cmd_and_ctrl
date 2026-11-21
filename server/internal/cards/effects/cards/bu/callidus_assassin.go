package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Callidus Assassin — Creature — Human Shapeshifter Assassin {4}{U}{B}, 3/3:
//
//	"Flash
//	 Polymorphine — You may have this creature enter tapped as a copy of any
//	 creature on the battlefield, except it has "When this creature enters,
//	 destroy up to one other target creature with the same name as this
//	 creature.""
//
// The "enter tapped" belongs to the copy choice: declining the copy leaves
// the Assassin a plain untapped 3/3 with no trigger. The granted trigger's
// "same name" is judged against the copy's own name (SameNameAsSource), live
// at announce and again at resolution.
const callidusAssassinGrant = "callidus-assassin/destroy-same-name"

func init() {
	effects.Register(effects.Spec{
		OracleID:        "4240dab9-ace7-490d-97c3-489a9fd52038",
		Name:            "Callidus Assassin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Grants: []effects.AbilityGrant{{
			Key:  callidusAssassinGrant,
			Text: "When this creature enters, destroy up to one other target creature with the same name as this creature.",
			Triggered: []game.TriggeredAbility{
				effects.Targeting(
					effects.WhenThisEnters("Callidus Assassin — destroy up to one other target creature with the same name", effects.DestroyFirstLegalTarget),
					effects.RelativeToSource(
						effects.Another(effects.TargetCreature("up to one other target creature with the same name as this creature").WithCount(0, 1)),
						effects.SameNameAsSource())),
			},
		}},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("Callidus Assassin", effects.AnyCreatureOnBattlefield,
				func(ev *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					ev.EntersTapped = true
					v.GrantAbility(callidusAssassinGrant)
				}),
		},
	})
}

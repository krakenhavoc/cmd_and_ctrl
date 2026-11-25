package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Evil Twin — Creature — Shapeshifter {2}{U}{B}, 0/0 (Battlebond):
//
//	"You may have this creature enter as a copy of any creature on the
//	 battlefield, except it has "{U}{B}, {T}: Destroy target creature with the
//	 same name as this creature.""
//
// Callidus Assassin's copy choice (EntersAsCopyOf, CR 707.2 and 614.1c) with a
// granted activated ability instead of a trigger. The quoted ability is part of
// the copiable values (CR 707.9a), so a Clone of an Evil Twin has it too. "The
// same name as this creature" is judged against the copy's own name, at
// activation and again as it resolves, and may name the copy itself: a copy of
// a creature can destroy that creature or itself. Declined, it is a 0/0 and
// dies (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da17da9b-f300-4613-978b-ba3026d4e275",
		Name:         "Evil Twin",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:  evilTwinGrant,
			Text: "{U}{B}, {T}: Destroy target creature with the same name as this creature.",
			Activated: []effects.ActivatedAbility{{
				Label:   "{U}{B}, {T}: Destroy target creature with the same name as this creature.",
				Cost:    effects.Plus(effects.ManaCost("{U}{B}"), effects.TapCost()),
				Purpose: game.Purpose{Answers: game.AnswerRemove},
				Targets: effects.RelativeToSource(
					effects.TargetCreature("target creature with the same name as this creature"),
					effects.SameNameAsSource()),
				Effect: effects.DestroyFirstLegalTarget,
			}},
		}},
		Replacements: []game.ReplacementEffect{
			effects.EntersAsCopyOf("Evil Twin", effects.AnyCreatureOnBattlefield,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.GrantAbility(evilTwinGrant)
				}),
		},
	})
}

// evilTwinGrant names the granted ability's bundle: an on-disk identity.
const evilTwinGrant = "evil-twin/destroy-same-name"

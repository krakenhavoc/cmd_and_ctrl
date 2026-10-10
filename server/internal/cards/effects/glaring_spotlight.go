package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glaring Spotlight — Artifact {1}:
//
//	"Creatures your opponents control with hexproof can be the targets
//	 of spells and abilities you control as though they didn't have
//	 hexproof.
//	 {3}, Sacrifice this artifact: Creatures you control gain hexproof
//	 until end of turn and can't be blocked this turn."
//
// Skipped from #1649 and built with #1650. Each half uses a different
// piece of machinery:
//
//   - The static is #1560's hexproof waiver (Spec.HexproofBypasses),
//     read live at the targeting choke point. It is
//     BySpellsAndAbilitiesYouControl, so only YOUR spells and abilities
//     ignore the hexproof; another opponent of the hexproof creature
//     gets nothing from it. That is the printed difference from
//     Nowhere to Run. The creature keeps its hexproof.
//   - The activated ability's two clauses lock their sets differently,
//     and that is why the card had to wait:
//     "gain hexproof until end of turn" changes a characteristic, so
//     CR 611.2c locks it to the creatures you control as it resolves
//     (GrantKeywordUntilEOT's snapshot). "Can't be blocked this turn"
//     changes the rules, so it also covers a creature you cast or
//     flash in afterwards (the card's ruling). RestrictUntilEOT reads
//     game.ScopeYourCreatures live until cleanup.
//
// The Spotlight is sacrificed as a cost, so its static is gone by the
// time the ability resolves, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "21570481-190e-4cfa-a5fd-a0642e2af569",
		Name:         "Glaring Spotlight",
		Completeness: CompletenessFull,
		HexproofBypasses: []game.HexproofBypass{
			AsThoughNoHexproof(BySpellsAndAbilitiesYouControl, Creature(), OpponentControls()),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}, Sacrifice this artifact: Creatures you control gain hexproof until end of turn and can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerProtect | game.AnswerCombatGrant},
			Cost:    Plus(ManaCost("{3}"), SacrificeThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"hexproof"},
					Label:    "Glaring Spotlight — hexproof",
				}).Apply(ctx); err != nil {
					return err
				}
				return RestrictUntilEOT{
					Scope:        game.ScopeYourCreatures,
					Restrictions: game.CantBeBlocked,
					Label:        "Glaring Spotlight — can't be blocked",
				}.Apply(ctx)
			},
		}},
	})
}

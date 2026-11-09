package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Diamond Weapon — Legendary Artifact Creature — Elemental {7}{G}{G}, 8/8
// (Final Fantasy, #4462):
//
//	"This spell costs {1} less to cast for each permanent card in your
//	 graveyard.
//	 Reach
//	 Immune — Prevent all combat damage that would be dealt to Diamond
//	 Weapon."
//
// Furygale Flocking's self cost reduction, counting the permanent cards in
// your graveyard (generic mana only, floored at zero, read as the spell is
// cast). Immune is an ability word with no rules meaning; the ability is a
// static prevention of combat damage dealt to this creature (CR 615.1a),
// Crystal Barricade's shape with the recipient "this creature". Prevented
// damage is not dealt, so no damage marks it and lifelink gains nothing
// from it; trample assignment ignores the prevention, as the rulings say.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3beba02f-4d50-4a93-a13b-78d4595bfbfc",
		Name:            "Diamond Weapon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		SelfCostModifiers: []game.CostModifier{
			CostsLessEach(CountCardsInYourGraveyard(Permanent()),
				"This spell costs {1} less to cast for each permanent card in your graveyard."),
		},
		Replacements: []game.ReplacementEffect{
			PreventAllDamageDealtTo(PreventionStatic{
				To:     ToThisCreature,
				Damage: CombatDamage,
				Label:  "Diamond Weapon — prevent all combat damage that would be dealt to it",
			}),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sapling Nursery — Enchantment {6}{G}{G} (Lorwyn Eclipsed, #4458):
//
//	"Affinity for Forests (This spell costs {1} less to cast for each
//	 Forest you control.)
//	 Landfall — Whenever a land you control enters, create a 3/4 green
//	 Treefolk creature token with reach.
//	 {1}{G}, Exile this enchantment: Treefolk and Forests you control gain
//	 indestructible until end of turn."
//
// Affinity is Cantankerous Keepers' self cost modifier (CR 702.41a) over
// Forests, read by subtype, so a nonbasic Forest counts. The last
// ability's set is fixed as it resolves (CR 611.2c): Treefolk and
// Forest permanents you control then, a Treefolk token and a land that
// is a Forest alike. Exiling the enchantment is part of the cost, so
// the ability resolves without it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8bcbae66-1854-480e-a449-33e5389610ab",
		Name:         "Sapling Nursery",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for Forests", HasSubtype("Forest")),
		},
		Triggered: []game.TriggeredAbility{
			LandfallPaying("Sapling Nursery — create a 3/4 green Treefolk creature token with reach (landfall)",
				game.LandfallPayoff{Tokens: 1},
				Do(CreateToken{Template: TokenCard("3/4 green Treefolk with reach"), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{G}, Exile this enchantment: Treefolk and Forests you control gain indestructible until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    Plus(ManaCost("{1}{G}"), ExileThis()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Match:    And(YouControl(), Or(HasSubtype("Treefolk"), HasSubtype("Forest"))),
					Keywords: []string{"indestructible"},
					Label:    "Sapling Nursery — indestructible",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

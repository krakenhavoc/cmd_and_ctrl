package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Wind Crystal — Legendary Artifact {2}{W}{W}:
//
//	"White spells you cast cost {1} less to cast.
//	 If you would gain life, you gain twice that much life instead.
//	 {4}{W}{W}, {T}: Creatures you control gain flying and lifelink
//	 until end of turn."
//
// The white Crystal, and three lines that are each already a shared
// shape in the catalog:
//
//   - the discount is Pearl Medallion's (CostsLess with a colour
//     predicate, generic mana only per CR 601.2f);
//   - the life-gain doubler is Rhox Faithmender's and Alhammarret's
//     Archive's replacement (YouGainTwiceThatMuchLife), which since
//     #482 sees every writer of a life total, lifelink included — so
//     the activated ability's lifelink feeds the doubler, which is
//     the card's own combo;
//   - the activated ability is Akroma's Will's grant: the set of
//     creatures it affects is locked when it resolves (CR 611.2c), so
//     a creature that enters afterwards does not fly.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "240f0835-36af-4ad8-9336-d6d3d816d293",
		Name:         "The Wind Crystal",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "White spells you cast cost {1} less to cast.",
				YourSpell(), ColoredSpell("W")),
		},
		Replacements: []game.ReplacementEffect{
			YouGainTwiceThatMuchLife("The Wind Crystal: gain twice that much life"),
		},
		Activated: []ActivatedAbility{{
			Label: "{4}{W}{W}, {T}: Creatures you control gain flying and lifelink until end of turn.",
			Cost:  Plus(ManaCost("{4}{W}{W}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return GrantKeywordUntilEOT{
					Match:    And(Creature(), YouControl()),
					Keywords: []string{"flying", "lifelink"},
					Label:    "The Wind Crystal — flying and lifelink",
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

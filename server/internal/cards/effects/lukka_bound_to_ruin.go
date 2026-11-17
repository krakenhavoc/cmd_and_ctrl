package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lukka, Bound to Ruin — Legendary Planeswalker — Lukka {2}{R}{R/G/P}{G},
// starting loyalty 5 (Phyrexia: All Will Be One, #5479):
//
//	"Compleated ({R/G/P} can be paid with {R}, {G}, or 2 life. If life was
//	 paid, this planeswalker enters with two fewer loyalty counters.)
//	 +1: Add {R}{G}. Spend this mana only to cast creature spells or
//	     activate abilities of creatures.
//	 −1: Create a 3/3 green Phyrexian Beast creature token with toxic 1.
//	 −4: Lukka deals X damage divided as you choose among any number of
//	     target creatures and/or planeswalkers, where X is the greatest
//	     power among creatures you control as you activate this ability."
//
// Compleated is ADR 0183's replacement (effects.Compleated). The +1 is a
// loyalty ability, so it uses the stack and its mana empties with the step
// (CR 605.1a); the restriction is Castle Garenbrig's two clauses. The −4's X
// is read once, as the ability is announced (CR 601.2d, 602.2b), and divided
// then: DivideBy(s82OneB20DivideGreatestPower). A target that has become
// illegal takes nothing and its share is not redistributed (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8e68eba1-65ac-4c03-956f-45e7b5695e2a",
		Name:            "Lukka, Bound to Ruin",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		PrintedKeywords: []string{"compleated"},
		Replacements:    []game.ReplacementEffect{Compleated()},
		Activated: []ActivatedAbility{
			{
				Label: "+1: Add {R}{G}. Spend this mana only to cast creature spells or activate abilities of creatures.",
				Cost:  LoyaltyCost(1),
				Effect: Do(AddMana{
					Produced: "{R}{G}",
					Restrictions: []string{game.ManaRestrictAnyOf(
						[]string{game.ManaRestrictCast, game.ManaRestrictType("Creature")},
						[]string{game.ManaRestrictActivate, game.ManaRestrictType("Creature")},
					)},
				}),
			},
			{
				Label:   "−1: Create a 3/3 green Phyrexian Beast creature token with toxic 1.",
				Cost:    LoyaltyCost(-1),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return s82OneB20CreateBeasts(g, item, 1)
				},
			},
			{
				Label: "−4: Lukka deals X damage divided as you choose among any number of target creatures and/or planeswalkers, where X is the greatest power among creatures you control as you activate this ability.",
				Cost:  LoyaltyCost(-4),
				Targets: TargetPermanent("any number of target creatures and/or planeswalkers",
					Or(Creature(), Planeswalker())).WithCount(0, 0).Dividing(DivideBy(s82OneB20DivideGreatestPower)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DealDividedDamage(NewContext(g, item))
				},
			},
		},
	})
}

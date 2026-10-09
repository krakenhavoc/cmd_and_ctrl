package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chandra, Chill of Compliance — Legendary Planeswalker — Chandra
// {1}{U}{U}, starting loyalty 3:
//
//	"+1: Surveil 1. If you put a noncreature, nonland card into your
//	 graveyard this way, put that card into your hand.
//	 +1: Add {U}. Spend this mana only to cast a noncreature spell.
//	 −X: Tap target artifact or creature. Put X stun counters on it.
//	 −6: You get an emblem with "Whenever you cast a spell, draw a card.""
//
// The first +1 is a surveil whose graveyard is compared before and after
// the answer (surveilKeepNoncreatureNonland): the noncreature, nonland
// card it just binned comes back to hand. The second +1 is a loyalty
// ability, so it uses the stack, unlike a mana ability; the mana carries
// "only to cast a noncreature spell" (it cannot pay for a creature spell
// or an activated ability) and empties with the step. The emblem triggers
// on every spell its owner casts, one draw each.
//
// The −X is NOT offered: a loyalty ability whose cost is X has no shape
// yet (the loyalty-cost-x seam, #1944), the same gap Ugin, the Spirit
// Dragon ships with. Printing it with a fixed cost would be wrong either
// way, so the card is a planeswalker with three of its four abilities.
func init() {
	Register(Spec{
		OracleID:     "c3dfa1e2-6785-49a0-a194-fb842a8eb63c",
		Name:         "Chandra, Chill of Compliance",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The -X ability isn't offered — an ability whose loyalty cost is X has no shape yet.",
		},
		StartingLoyalty: 3,
		Emblem: &EmblemSpec{
			Label: "Chandra, Chill of Compliance emblem",
			Text:  "Whenever you cast a spell, draw a card.",
			Triggered: []game.TriggeredAbility{
				WheneverYouCast(nil, "Chandra, Chill of Compliance emblem — draw a card", Do(DrawCards{N: 1})),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:  "+1: Surveil 1. If you put a noncreature, nonland card into your graveyard this way, put that card into your hand.",
				Cost:   LoyaltyCost(1),
				Effect: surveilKeepNoncreatureNonland,
			},
			{
				Label: "+1: Add {U}. Spend this mana only to cast a noncreature spell.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return AddMana{
						Produced:     "{U}",
						Restrictions: []string{ManaRestrictCast, ManaRestrictNotType("Creature")},
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "−6: You get an emblem with \"Whenever you cast a spell, draw a card.\"",
				Cost:  LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

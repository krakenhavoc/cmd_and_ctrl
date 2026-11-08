package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Baylen, the Haymaker — Legendary Creature — Rabbit Warrior {R}{G}{W},
// 4/3 (Bloomburrow, #4467):
//
//	"Tap two untapped tokens you control: Add one mana of any color.
//	 Tap three untapped tokens you control: Draw a card.
//	 Tap four untapped tokens you control: Put three +1/+1 counters on
//	 Baylen. It gains trample until end of turn."
//
// Each cost is #758's TapOthers with a count, Heritage Druid's shape: any
// tokens you control, creature or not (a Food or a Treasure pays it too).
// It is not the {T} symbol (CR 302.6), so tokens that arrived this turn may
// be tapped for it, Baylen himself is never one of them (he is not a token),
// and Baylen needs no haste to use any of the three. The first is a mana
// ability (CR 605.1a), so it does not use the stack. The third does nothing
// to a new object (CR 400.7) if Baylen left and came back in response, as
// Rough Rhino Cavalry's exhaust does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "73dc28a7-37ac-4b16-aa00-967c9c44d979",
		Name:         "Baylen, the Haymaker",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{TapOthers: baylenTapTokens(2, "two untapped tokens you control")},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "Tap three untapped tokens you control: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Cost:    game.AbilityCost{TapOthers: baylenTapTokens(3, "three untapped tokens you control")},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return DrawCards{N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "Tap four untapped tokens you control: Put three +1/+1 counters on Baylen. It gains trample until end of turn.",
				Purpose: game.Purpose{
					Answers: game.AnswerPump | game.AnswerCombatGrant,
					Pump:    &game.Pump{Power: 3, Toughness: 3, Keywords: []string{"trample"}},
				},
				Cost: game.AbilityCost{TapOthers: baylenTapTokens(4, "four untapped tokens you control")},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if !sourceIsStillThisPermanent(g, item) {
						return nil
					}
					if err := plusOneCountersOnThis(3)(g, item); err != nil {
						return err
					}
					return thisGainsKeywordUntilEndOfTurn("trample", "Baylen, the Haymaker — trample until end of turn")(g, item)
				},
			},
		},
	})
}

// baylenTapTokens is "Tap <n> untapped tokens you control" as a cost.
func baylenTapTokens(n int, label string) *game.TapOthersCost {
	return &game.TapOthersCost{
		Count:  n,
		Filter: TargetPermanent(label, IsTokenPredicate()),
		Label:  label,
	}
}

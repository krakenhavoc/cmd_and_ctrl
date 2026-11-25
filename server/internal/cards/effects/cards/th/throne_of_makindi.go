package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Throne of Makindi — Land (Zendikar Rising, #5508):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Put a charge counter on this land.
//	 {T}, Remove a charge counter from this land: Add two mana of any one
//	 color. Spend this mana only to cast kicked spells."
//
// Two mana abilities and an ordinary activated ability. The counter comes
// off as part of the third's cost (Vivid Creek's shape), the two mana are of
// one colour chosen as it is produced (Rootcoil Creeper's), and the
// restriction admits only a cast of a spell whose kicker was paid
// (ManaRestrictKickedSpell, CR 702.33a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7e8198e9-0f3b-420b-ab09-74f13f4fd548",
		Name:         "Throne of Makindi",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost: effects.ManaAbilityCost{
					Tap:            true,
					RemoveCounters: effects.RemoveCountersFromThis(game.CounterCharge, 1).RemoveCounters,
				},
				Produced:     effects.OneColorOfAmount(2),
				Label:        "Remove a charge counter: Add two mana of any one color. Spend this mana only to cast kicked spells",
				Restrictions: []string{game.ManaRestrictCast, game.ManaRestrictKickedSpell},
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Put a charge counter on this land.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect:  effects.Do(effects.CounterOnThis{Kind: game.CounterCharge, N: 1}),
		}},
	})
}

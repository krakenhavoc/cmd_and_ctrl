package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crucible of the Spirit Dragon — Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}: Put a storage counter on this land.
//	 {T}, Remove X storage counters from this land: Add X mana in any
//	 combination of colors. Spend this mana only to cast Dragon spells or
//	 activate abilities of Dragons."
//
// Mage-Ring Network's storage shape (see that card): the counter is a
// plain activated ability, and the payout is a mana ability whose amount is
// the number of counters actually removed (X may be 0, which the land
// still taps for). The mana is X independent picks of any colour, each
// restricted to Dragon spells and to Dragons' activated abilities.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ecfbebc9-7fc7-474e-8c59-8ede800e082e",
		Name:         "Crucible of the Spirit Dragon",
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
					RemoveCounters: effects.RemoveCountersXFromThis(game.CounterStorage, 0).RemoveCounters,
				},
				ProducedForPaid: crucibleOfTheSpiritDragonMana,
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{game.ManaRestrictCast, game.ManaRestrictSubtype("Dragon")},
					[]string{game.ManaRestrictActivate, game.ManaRestrictSubtype("Dragon")},
				)},
				Label: "{T}, Remove X storage counters: Add X mana in any combination of colors. Spend this mana only to cast Dragon spells or activate abilities of Dragons",
			},
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Put a storage counter on this land.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect:  effects.PutStorageCounterOnThis,
		}},
	})
}

// crucibleOfTheSpiritDragonMana is X mana, each of any colour, X the
// counters removed to pay.
func crucibleOfTheSpiritDragonMana(_ *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
	return game.ManaOutput{Amount: game.ManaAmount{Options: game.AllColors, N: paid.CountersRemoved}}
}

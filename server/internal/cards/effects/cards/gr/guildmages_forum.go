package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Guildmages' Forum — Land (Guilds of Ravnica, #5525):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color. If that mana is spent on a
//	 multicolored creature spell, that creature enters with an additional +1/+1
//	 counter on it."
//
// Biophagus' entry spend rider (#1547) with the spend filter narrowed to a
// multicolored creature spell (cast, creature, multicolored: every tag must
// hold): when the token pays for such a spell the rider is stamped on the
// spell's payment record, and the entry pipeline seeds one additional +1/+1
// counter onto the creature's entry event beside any the card prints, so a
// counter doubler sees it. The mana pays for anything; only the counter depends
// on the spell. It is one activation's mana, so a doubler that makes two tokens
// is still one additional counter. The {1} is part of the cost, paid with other
// mana, and the {C} ability is the plain land.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "ace6403d-9fac-4d0f-a6ea-eb2ff3da259d",
		Name:           "Guildmages' Forum",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:     effects.ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color",
				SpendRiders: []game.ManaSpendRider{
					effects.SpentEntersWithCounters("+1/+1", 1, game.ManaRestrictCast, effects.ManaRestrictType("Creature"), game.ManaRestrictMulticolored),
				},
			},
		},
	})
}

package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defiler of Vigor — Creature — Phyrexian Wurm {3}{G}{G}, 6/6 (Dominaria United,
// #5659):
//
//	"Trample
//	 As an additional cost to cast green permanent spells, you may pay 2
//	 life. Those spells cost {G} less to cast if you paid life this way.
//	 This effect reduces only the amount of green mana you pay.
//	 Whenever you cast a green permanent spell, put a +1/+1 counter on each
//	 creature you control."
//
// The offer and its discount are the cycle's linked pair
// (pay_life_for_discount.go): each green permanent spell its controller
// casts may be announced with "pay 2 life", and one that is costs {G}
// less, never less generic mana. The counters go
// on the creatures its controller controls as the trigger resolves, so
// the creature spell that caused it gets none.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:             "0a1dc797-7531-4494-afe6-add6ac3f74c0",
		Name:                 "Defiler of Vigor",
		Completeness:         effects.CompletenessFull,
		PrintedKeywords:      []string{"trample"},
		GrantedOptionalCosts: []game.GrantedOptionalCost{effects.MayPayLifeToCast("Defiler of Vigor", "G", 2)},
		CostModifiers:        []game.CostModifier{effects.CostsColorLessIfLifePaid("G")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCastAPermanentSpellOfColor("G", "Defiler of Vigor — put a +1/+1 counter on each creature you control",
				effects.B11PutCounterOnEachCreatureYouControl),
		},
	})
}

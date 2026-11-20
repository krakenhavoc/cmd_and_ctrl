package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Defiler of Dreams — Creature — Phyrexian Sphinx {3}{U}{U}, 4/3 (Dominaria United,
// #5659):
//
//	"Flying
//	 As an additional cost to cast blue permanent spells, you may pay 2
//	 life. Those spells cost {U} less to cast if you paid life this way.
//	 This effect reduces only the amount of blue mana you pay.
//	 Whenever you cast a blue permanent spell, draw a card."
//
// The offer and its discount are the cycle's linked pair
// (pay_life_for_discount.go): each blue permanent spell its controller
// casts may be announced with "pay 2 life", and one that is costs {U}
// less, never less generic mana. The trigger
// reads the spell as it was cast, so the Defiler's own cast never
// reaches it and an instant or sorcery never does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:             "9d4ec5e5-5520-45c3-8071-53191811117d",
		Name:                 "Defiler of Dreams",
		Completeness:         effects.CompletenessFull,
		PrintedKeywords:      []string{"flying"},
		GrantedOptionalCosts: []game.GrantedOptionalCost{effects.MayPayLifeToCast("Defiler of Dreams", "U", 2)},
		CostModifiers:        []game.CostModifier{effects.CostsColorLessIfLifePaid("U")},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCastAPermanentSpellOfColor("U", "Defiler of Dreams — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}

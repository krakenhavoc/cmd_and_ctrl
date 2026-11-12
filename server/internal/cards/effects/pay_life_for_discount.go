package effects

import (
	"fmt"
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pay_life_for_discount.go — #5659, the Defiler cycle of Dominaria
// United:
//
//	"As an additional cost to cast blue permanent spells, you may pay 2
//	 life. Those spells cost {U} less to cast if you paid life this way.
//	 This effect reduces only the amount of blue mana you pay."
//
// Two linked halves of one static ability, declared side by side:
//
//	GrantedOptionalCosts: []game.GrantedOptionalCost{MayPayLifeToCast("Defiler of Dreams", "U", 2)},
//	CostModifiers:        []game.CostModifier{CostsColorLessIfLifePaid("U")},
//
// The offer is a GrantedOptionalCost (game/granted_optional_cost.go)
// priced in life, one per permanent that grants it (EachSource), so two
// Defilers of one colour are two offers of 2 life each. The intention
// to pay is announced with the spell's other optional costs (CR 601.2b)
// and the life is paid with the total cost (CR 601.2h), so a player who
// can't pay life (CR 119.4, Karn's Sylex) can't announce it. The
// discount reads the announcement of its own permanent's offer
// (CostQuery.GrantedOptionalTimesFromSource) as the total cost is
// determined (CR 601.2f).

var payLifeColorWords = map[string]string{"W": "white", "U": "blue", "B": "black", "R": "red", "G": "green"}

func payLifeColorWord(color string) string {
	w, ok := payLifeColorWords[color]
	if !ok {
		panic(fmt.Sprintf("effects: %q is not a colour letter", color))
	}
	return w
}

// MayPayLifeToCast is "As an additional cost to cast <colour> permanent
// spells, you may pay N life." `from` is the permanent's name, which the
// cost picker shows beside the amount so two such offers on one spell
// can be told apart. The colour is a word of the permanent's text, read
// through its changed words (CR 612.2), and a permanent spell is an
// artifact, battle, creature, enchantment or planeswalker spell (CR
// 110.4b).
func MayPayLifeToCast(from, color string, life int) game.GrantedOptionalCost {
	word := payLifeColorWord(color)
	gr := SpellsYouCastHave(
		"As an additional cost to cast "+word+" permanent spells, you may pay "+strconv.Itoa(life)+" life.",
		And(Permanent(), OfColor(color)),
		game.AdditionalCost{
			Optional: true,
			Key:      game.PaidLifeForDiscountKey,
			PayLife:  life,
			Label:    "Pay " + strconv.Itoa(life) + " life (" + from + ")",
		})
	gr.EachSource = true
	return gr
}

// CostsColorLessIfLifePaid is the linked "Those spells cost {C} less to
// cast if you paid life this way. This effect reduces only the amount
// of <colour> mana you pay.": one mana of the colour off the spell's
// cost when this permanent's own MayPayLifeToCast offer was announced,
// and nothing off the generic part (game.CostModifier.UnitColorsOnly).
func CostsColorLessIfLifePaid(color string) game.CostModifier {
	word := payLifeColorWord(color)
	unit, err := game.ParseCost("{" + color + "}")
	if err != nil {
		panic(fmt.Sprintf("effects.CostsColorLessIfLifePaid: %v", err))
	}
	m := CostsLessEach(func(q game.CostQuery) int {
		return q.GrantedOptionalTimesFromSource(game.PaidLifeForDiscountKey)
	}, "Those spells cost {"+color+"} less to cast if you paid life this way. This effect reduces only the amount of "+word+" mana you pay.")
	m.Unit = &unit
	m.UnitColorsOnly = true
	return m
}

// WheneverYouCastAPermanentSpellOfColor is the cycle's "Whenever you
// cast a <colour> permanent spell".
func WheneverYouCastAPermanentSpellOfColor(color, label string, effect Effect) game.TriggeredAbility {
	return WheneverYouCast(And(Permanent(), OfColor(color)), label, effect)
}

package game

import (
	"fmt"
	"reflect"
	"strings"
)

// ability_row_labels_slots.go — #3133: the wording of the structured
// slots #2410 left on their generic fallbacks. Each reads the entry's
// own fields, so every card that declares one gains wording at once.

// legendRuleExemptionRows words CardDef.LegendRuleExemptions.
func legendRuleExemptionRows(v reflect.Value) []string {
	exemptions, _ := v.Interface().([]LegendRuleExemption)
	out := make([]string, 0, len(exemptions))
	for _, e := range exemptions {
		switch e.Whose {
		case LegendRuleExemptEveryone:
			out = append(out, "The legend rule doesn't apply.")
		case LegendRuleExemptYourCreatures:
			out = append(out, "The legend rule doesn't apply to creatures you control.")
		default:
			out = append(out, "The legend rule doesn't apply to permanents you control.")
		}
	}
	return out
}

// unpreventableDamageRows words CardDef.DamageCantBePrevented.
func unpreventableDamageRows(v reflect.Value) []string {
	statics, _ := v.Interface().([]UnpreventableDamageStatic)
	out := make([]string, 0, len(statics))
	for _, s := range statics {
		switch s.Scope {
		case UnpreventableByThis:
			out = append(out, "Damage dealt by this creature can't be prevented.")
		case UnpreventableCombat:
			out = append(out, "Combat damage can't be prevented.")
		case UnpreventableCombatByYourCreatures:
			out = append(out, "Combat damage dealt by creatures you control can't be prevented.")
		default:
			out = append(out, "Damage can't be prevented.")
		}
	}
	return out
}

// anyColorSpendRows words CardDef.AnyColorSpend. A grant narrowed to
// some payments (Covers is a closure the row cannot read) keeps the
// clause the card printed.
func anyColorSpendRows(v reflect.Value) []string {
	statics, _ := v.Interface().([]AnyColorSpendStatic)
	out := make([]string, 0, len(statics))
	for _, s := range statics {
		who := "You may"
		if s.Whose == AnyColorSpendEveryPlayer {
			who = "Players may"
		}
		switch {
		case s.Covers != nil && strings.TrimSpace(s.Label) != "":
			out = append(out, strings.TrimSpace(s.Label))
		case s.Covers != nil:
			out = append(out, who+" spend mana as though it were mana of any color on some costs.")
		default:
			out = append(out, who+" spend mana as though it were mana of any color.")
		}
	}
	return out
}

var manaColorWords = map[string]string{
	"W": "white", "U": "blue", "B": "black", "R": "red", "G": "green", "C": "colorless",
}

// manaPoolRows words CardDef.ManaPool.
func manaPoolRows(v reflect.Value) []string {
	statics, _ := v.Interface().([]ManaPoolStatic)
	out := make([]string, 0, len(statics))
	for _, s := range statics {
		who, would := "You", "you would"
		if s.Players == ManaPoolEachPlayer {
			who, would = "Players", "a player would"
		}
		if s.Kind == ManaPoolBecomesColorless {
			out = append(out, fmt.Sprintf("If %s lose unspent mana as steps and phases end, it becomes colorless instead.", would))
			continue
		}
		var words []string
		for _, c := range s.Colors {
			if w, ok := manaColorWords[c]; ok {
				words = append(words, w)
			}
		}
		kind := "mana"
		if len(words) > 0 {
			kind = strings.Join(words, " and ") + " mana"
		}
		out = append(out, fmt.Sprintf("%s don't lose unspent %s as steps and phases end.", who, kind))
	}
	return out
}

// grantedAlternativeCostRow words one CardDef.GrantedAlternativeCosts
// entry; the row walk applies its door gate first.
func grantedAlternativeCostRow(e reflect.Value) string {
	gr, _ := e.Interface().(GrantedAlternativeCost)
	return gr.rowLabel()
}

// rowLabel reads the offer's price, the spells it reaches and the zones
// it is limited to.
func (gr GrantedAlternativeCost) rowLabel() string {
	subject := "spells"
	switch {
	case gr.Spells.CreatureOnly:
		subject = "creature spells"
	case gr.Spells.NonLandPermanentOnly:
		subject = "permanent spells"
	}
	if gr.SpellsText != "" {
		subject = gr.SpellsText
	}
	o := gr.Offer
	if o.Key == AltCostKeyGrantedMiracle {
		// Molecule Man: the granted keyword, in the printed words.
		return "Nonland cards in your hand have miracle " + o.ManaCost + "."
	}
	if gr.PricedAtManaCost || o.Sacrifice != nil {
		return "Each " + strings.TrimSuffix(subject, "s") + " you cast can be cast by sacrificing a creature, reducing its cost by that creature's mana value."
	}
	fromHand := len(gr.Zones) == 1 && gr.Zones[0] == ZoneHand
	fromExile := len(gr.Zones) == 1 && gr.Zones[0] == ZoneExile
	var limit string
	switch {
	case gr.MaxManaValue != nil:
		limit = fmt.Sprintf(" with mana value %d or less", *gr.MaxManaValue)
	case gr.MaxManaValueFunc != nil && gr.MaxManaValueText != "":
		limit = " with mana value less than or equal to " + gr.MaxManaValueText
	}
	if fromHand {
		limit += " from your hand"
	}
	if fromExile {
		limit += " from exile"
	}
	var price string
	switch {
	case o.ManaCost != "":
		price = "pay " + o.ManaCost
	case o.Energy == 1:
		price = "pay {E}"
	case o.Energy > 1:
		price = fmt.Sprintf("pay %d {E}", o.Energy)
	}
	var s string
	switch {
	case gr.LifeEqualToManaValue:
		s = "You may cast " + subject + limit + " by paying life equal to their mana value rather than paying their mana costs."
	case price == "":
		s = "You may cast " + subject + limit + " without paying their mana costs."
	default:
		s = "You may " + price + " rather than pay the mana cost for " + subject + " you cast" + limit + "."
	}
	if o.AsThoughFlash {
		s += " Spells cast this way can be cast as though they had flash."
	}
	// #3174: the once-each-turn limit and the "your turn" clause lead
	// the sentence, as printed.
	switch {
	case gr.OnceEachTurn && gr.YourTurnOnly:
		s = "Once during each of your turns, you" + strings.TrimPrefix(s, "You")
	case gr.OnceEachTurn:
		s = "Once each turn, you" + strings.TrimPrefix(s, "You")
	case gr.YourTurnOnly:
		s = "During your turn, you" + strings.TrimPrefix(s, "You")
	}
	return s
}

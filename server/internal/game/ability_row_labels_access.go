package game

import (
	"reflect"
	"strconv"
)

// ability_row_labels_access.go — #3133: the wording of the hexproof
// waiver and special-action-grant slots.

// hexproofBypassRows words CardDef.HexproofBypasses. Which permanents
// it reaches is a closure the row cannot read, so the wording says whose
// spells and abilities ignore hexproof and leaves the rest to the card.
func hexproofBypassRows(v reflect.Value) []string {
	bypasses, _ := v.Interface().([]HexproofBypass)
	out := make([]string, 0, len(bypasses))
	for _, b := range bypasses {
		who := "Spells and abilities"
		if b.YoursOnly {
			who = "Spells and abilities you control"
		}
		what := "hexproof"
		if b.Player != nil && b.Permanent == nil {
			what = "a player's hexproof"
		}
		out = append(out, who+" can ignore "+what+".")
	}
	return out
}

// specialActionGrantRows words CardDef.SpecialActionGrants.
func specialActionGrantRows(v reflect.Value) []string {
	grants, _ := v.Interface().([]SpecialActionGrant)
	out := make([]string, 0, len(grants))
	for _, gr := range grants {
		if gr.Kind == SpecialActionForetell && gr.Zone == ZoneHand {
			s := "Each nonland card in your hand without foretell has foretell."
			if gr.CostIsManaCost {
				s += " Its foretell cost is equal to its mana cost"
				if gr.CostReducedBy > 0 {
					s += " reduced by {" + strconv.Itoa(gr.CostReducedBy) + "}"
				}
				s += "."
			}
			out = append(out, s)
			continue
		}
		if gr.Kind != SpecialActionPlot || gr.Zone != ZoneLibrary {
			out = append(out, "Grants a special action.")
			continue
		}
		cards := "cards"
		if gr.Nonland {
			cards = "nonland cards"
		}
		s := "You may plot " + cards + " from the top of your library."
		if gr.CostIsManaCost {
			s += " Their plot cost is equal to their mana cost."
		}
		out = append(out, s)
	}
	return out
}

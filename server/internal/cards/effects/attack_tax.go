package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attack_tax.go — constructors for Spec.AttackTaxes, the CR 508.1a
// attack tax: "Creatures can't attack you unless their controller pays
// {2} for each creature they control that's attacking you". The engine
// half lives in game/attack_tax.go; this file is the vocabulary a card
// file writes in. ADR 0080, issue #1063.
//
// The shape to copy:
//
//	AttackTaxes: []game.AttackTax{
//	    AttackTax("{2}", "Creatures can't attack you unless their controller pays {2} for each creature they control that's attacking you."),
//	},
//
// and, for the ones that count something:
//
//	AttackTaxes: []game.AttackTax{
//	    AttackTaxCounting(func(q game.AttackTaxQuery) int {
//	        return enchantmentsControlledBy(q.Game, q.Defender)
//	    }, "Creatures can't attack you unless their controller pays {X} …"),
//	},
//
// Every constructor takes the printed clause as its label rather than
// generating one, for cost_modifier.go's reason: a generated label is
// right until two taxes on the board disagree about what they charge,
// and then the log says the same thing twice about different cards.
//
// THE "YOU" IS NOT YOURS TO SET. A tax protects its own controller,
// structurally, in the engine's pricer — see game.AttackTax. Nothing
// here can widen it to another seat, which is the point.
//
// What a card DOES choose is whether the clause covers the
// controller's planeswalkers as well ("creatures can't attack you OR
// PLANESWALKERS YOU CONTROL" — Sphere of Safety, Norn's Annex).
// That is ProtectingPlaneswalkers, and the default without it is the
// narrower "you", so a card file that forgets under-taxes.

// AttackTax is the flat clause: every creature attacking this
// permanent's controller costs `mana` — Propaganda's and Ghostly
// Prison's {2}, Windborn Muse's {2}, Norn's Annex's {W/P}.
//
// `mana` is a cost string in Scryfall brace notation and it is the
// price for ONE attacking creature; the engine concatenates it once
// per attacker, so a three-creature swing into Propaganda is
// "{2}{2}{2}", which ParseCost reads as six generic.
func AttackTax(mana, label string) game.AttackTax {
	return game.AttackTax{
		Label:    label,
		ManaCost: func(game.AttackTaxQuery) string { return mana },
	}
}

// AttackTaxCounting is the clause whose price is a generic amount
// counted off the board — Sphere of Safety's "{X}, where X is the
// number of enchantments you control", Collective Restraint's "{X},
// where X is the number of basic land types among lands you control".
//
// `n` is answered per attacking creature, with q.Defender as the "you"
// it counts for. A count of zero or less is a tax of nothing, which is
// Sphere of Safety alone on an otherwise empty board: X is 1 there
// (the Sphere is itself an enchantment), but a card whose count really
// can reach zero should read as free rather than as an error.
func AttackTaxCounting(n func(q game.AttackTaxQuery) int, label string) game.AttackTax {
	return game.AttackTax{
		Label: label,
		ManaCost: func(q game.AttackTaxQuery) string {
			c := n(q)
			if c <= 0 {
				return ""
			}
			return "{" + strconv.Itoa(c) + "}"
		},
	}
}

// ProtectingPlaneswalkers widens a tax from "creatures can't attack
// you" to "creatures can't attack you or planeswalkers you control" —
// Sphere of Safety, Norn's Annex.
//
// A wrapper rather than a parameter on every constructor so the
// narrower reading stays the default: a card file that does not write
// this taxes only direct attacks on its controller, which errs weaker
// than printed rather than stronger. Battles are NOT covered; see
// game.AttackTaxOnPlayerOrPlaneswalkers.
func ProtectingPlaneswalkers(t game.AttackTax) game.AttackTax {
	t.Scope = game.AttackTaxOnPlayerOrPlaneswalkers
	return t
}

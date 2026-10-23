package game

import "github.com/google/uuid"

// land_mana_color.go — #1997: "lands tapped for mana produce mana of the
// chosen color instead of any other color", Hall of Gemstone's until end
// of turn:
//
//	"At the beginning of each player's upkeep, that player chooses a
//	 color. Until end of turn, lands tapped for mana produce mana of the
//	 chosen color instead of any other color."
//
// A CR 614.1a replacement on the mana production event, and CR 106.12b
// says which event: the one a land's mana ability makes while it is
// resolving and producing the mana, for an ability with {T} in its cost
// (CR 106.12). It is a resolved ability's effect (CR 611.2a), so it is a
// ScopedEffect record (ADR 0041 P8) of its own kind, landManaIsColor,
// whose one parameter is the colour, and it lasts the turn whatever
// happens to the Hall.
//
// "Instead of any other color" rewrites only COLOURED mana: colourless
// is a type of mana but not a colour (CR 106.1a, 106.1b), so a Wastes
// still makes {C}, and mana already of the chosen colour is unchanged.
// Every coloured mana of the production becomes one mana of the chosen
// colour, so the amount is kept: a Karoo's {W}{U} under red is {R}{R}.
//
// The auto-tap planner prices the rewrite (priceProducedSlotsLocked): a
// slot whose every option comes out as one colour is booked as that
// colour, so a Forest under red pays {R}.

// ModLandManaIsColor is "lands tapped for mana produce mana of the
// chosen color instead of any other color this turn" (#1997, Hall of
// Gemstone). Reads Colors[0], one of W, U, B, R, G. Scope ScopeGame. A
// replacement kind (readerReplacement): the replacement gather reads it,
// never the layer pass.
const ModLandManaIsColor ModKind = "landManaIsColor"

var watchManaAdded = []EventKind{EventManaAdded}

// LandsTappedProduceColorThisTurnForEffect registers Hall of Gemstone's
// replacement until end of turn, controlled by `controller` (the player
// who chose the colour). Registers nothing for a colour that is not one
// of the five. Reports whether a record was written.
//
// Caller must hold g.mu (write).
func (g *Game) LandsTappedProduceColorThisTurnForEffect(sourceID, controller uuid.UUID, color, label string) bool {
	if !isFiveColor(color) {
		return false
	}
	return g.RegisterScopedRuleEffectForEffect(sourceID, ScopeGame, controller,
		[]Mod{{Kind: ModLandManaIsColor, Colors: []string{color}}}, g.UntilEndOfTurnDuration(), label)
}

func isFiveColor(c string) bool {
	switch c {
	case "W", "U", "B", "R", "G":
		return true
	}
	return false
}

// landManaColorProblem is registration's check on the mod: one of the
// five colours.
func landManaColorProblem(m Mod) string {
	if len(m.Colors) != 1 || !isFiveColor(m.Colors[0]) {
		return "a landManaIsColor replacement needs exactly one of W, U, B, R, G"
	}
	return ""
}

// landManaColorAppliesLocked is the kind's AppliesTo: a land tapped for
// mana (CR 106.12) that is producing some coloured mana other than the
// chosen colour. A production already all of that colour, or all
// colourless, is left alone: there is nothing to replace.
func landManaColorAppliesLocked(g *Game, m Mod, ev *ReplacementEvent) bool {
	if ev.Kind != RepEventProduceMana || !ev.ManaFromTap || len(m.Colors) != 1 {
		return false
	}
	want := m.Colors[0]
	differs := false
	for _, c := range ev.ManaColors {
		if isFiveColor(c) && c != want {
			differs = true
			break
		}
	}
	if !differs {
		return false
	}
	land, ok := g.LookupCardForEffect(ev.ManaSource)
	return ok && land.IsLand()
}

// applyLandManaColorLocked rewrites every coloured mana of the
// production to the chosen colour; colourless stays.
func applyLandManaColorLocked(m Mod, ev *ReplacementEvent) {
	if len(m.Colors) != 1 {
		return
	}
	out := make([]string, len(ev.ManaColors))
	for i, c := range ev.ManaColors {
		if isFiveColor(c) {
			c = m.Colors[0]
		}
		out[i] = c
	}
	ev.ManaColors = out
}

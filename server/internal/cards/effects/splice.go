package effects

import (
	"fmt"
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// splice.go — CR 702.47, splice (#1989, ADR 0202). The card side is one
// field, Spec.Splice; the engine reveals, prices, targets and runs the
// text (game/splice.go), and pays a cost that is not mana (game/
// splice_cost.go, #4055).

// SpliceOntoArcane is "Splice onto Arcane [cost]".
func SpliceOntoArcane(manaCost string) *game.SpliceAbility {
	return &game.SpliceAbility{Onto: game.SpliceOntoArcane, ManaCost: manaCost}
}

// SpliceOntoInstantOrSorcery is "Splice onto instant or sorcery [cost]".
func SpliceOntoInstantOrSorcery(manaCost string) *game.SpliceAbility {
	return &game.SpliceAbility{Onto: game.SpliceOntoInstantOrSorcery, ManaCost: manaCost}
}

// SpliceOntoArcaneCost is "Splice onto Arcane—[cost]" for a cost that is not
// mana (#4055): Veil of Secrecy, Horobi's Whisper, Roar of Jukai,
// Hundred-Talon Strike and Torrent of Stone.
//
//	Splice: SpliceOntoArcaneCost(SpliceSacrifice(2, "two Mountains", Land(), Subtype("Mountain"))),
func SpliceOntoArcaneCost(cost *game.SpliceCost) *game.SpliceAbility {
	return &game.SpliceAbility{Onto: game.SpliceOntoArcane, Cost: cost}
}

// SpliceSacrifice is "Sacrifice [N permanents]" as a splice cost: Torrent
// of Stone's "Sacrifice two Mountains". `label` is the clause without the
// verb, as printed; `preds` say what may pay, and "you control" is the
// engine's. Nothing is targeted, so a hexproof permanent pays.
func SpliceSacrifice(n int, label string, preds ...CardPredicate) *game.SpliceCost {
	return &game.SpliceCost{
		Label:     "Sacrifice " + label,
		Sacrifice: sacrificeSpec(label, preds...).WithCount(n, n),
	}
}

// SpliceExileFromGraveyard is "Exile N cards from your graveyard" as a
// splice cost: Horobi's Whisper's "Exile four cards from your graveyard".
// `label` is the clause without the verb or the zone ("four cards").
func SpliceExileFromGraveyard(n int, label string) *game.SpliceCost {
	return &game.SpliceCost{
		Label:      "Exile " + label + " from your graveyard",
		ExileCards: ExileCardsFromGraveyard(n, label, nil),
	}
}

// SpliceTapUntapped is "Tap an untapped [permanent] you control" as a
// splice cost: Hundred-Talon Strike's "Tap an untapped white creature you
// control". `label` is the clause without the verb, as printed ("an
// untapped white creature you control"). It is not the {T} symbol, so a
// creature that came under the caster's control this turn may pay (CR
// 302.6), and paying does not target.
func SpliceTapUntapped(label string, preds ...CardPredicate) *game.SpliceCost {
	return &game.SpliceCost{
		Label:     "Tap " + label,
		TapOthers: tapOthersPrice(1, label, preds...),
	}
}

// SpliceReturnToHand is "Return [a permanent] you control to its owner's
// hand" as a splice cost: Veil of Secrecy's "Return a blue creature you
// control to its owner's hand". `label` is the clause without the verb and
// the destination ("a blue creature you control").
func SpliceReturnToHand(label string, preds ...CardPredicate) *game.SpliceCost {
	return &game.SpliceCost{
		Label:        "Return " + label + " to its owner's hand",
		ReturnToHand: ReturnNToHand(1, label, preds...).ReturnToHand,
	}
}

// SpliceOpponentGainsLife is "An opponent gains N life" as a splice cost:
// Roar of Jukai's. The caster chooses the opponent, as gift's recipient is
// chosen, and the life is that opponent's gain through the CR 614 window.
func SpliceOpponentGainsLife(n int) *game.SpliceCost {
	return &game.SpliceCost{
		Label:             "An opponent gains " + strconv.Itoa(n) + " life",
		OpponentGainsLife: n,
	}
}

// checkSplice refuses a splice declaration the engine cannot carry out
// as printed:
//
//   - no cost, both a mana and a non-mana cost (no printed card has both),
//     or a non-mana cost made of anything but the one component a splice
//     cost uses at a time: a card without its cost would be stronger than
//     printed;
//   - an unknown quality;
//   - a text the engine could not add to another spell whole: no body,
//     a body split into modes, an announced X, a division, or a cost of
//     its own besides the splice cost. CR 702.47c adds the RULES TEXT,
//     and none of those is read off a spell's spliced text.
func checkSplice(spec Spec) {
	sa := spec.Splice
	if sa == nil {
		return
	}
	switch sa.Onto {
	case game.SpliceOntoArcane, game.SpliceOntoInstantOrSorcery:
	default:
		panic(fmt.Sprintf("effects.Register: %q splices onto an unknown quality %q", spec.Name, sa.Onto))
	}
	switch {
	case sa.ManaCost == "" && sa.Cost.Empty():
		panic(fmt.Sprintf("effects.Register: %q declares a splice cost that demands nothing — a card without its cost would be stronger than printed (ADR 0202)", spec.Name))
	case sa.ManaCost != "" && sa.Cost != nil:
		panic(fmt.Sprintf("effects.Register: %q declares a splice cost of mana and of something else — no printed splice card has both (ADR 0202)", spec.Name))
	}
	if sa.ManaCost != "" {
		if _, err := game.ParseCost(sa.ManaCost); err != nil {
			panic(fmt.Sprintf("effects.Register: %q splice cost %q: %v", spec.Name, sa.ManaCost, err))
		}
	}
	if sa.Cost != nil {
		checkSpliceCost(spec.Name, sa.Cost)
	}
	if spec.OnResolve == nil {
		panic(fmt.Sprintf("effects.Register: %q declares splice but no OnResolve — its rules text is what a spliced spell gains", spec.Name))
	}
	if spec.Modes != nil || spec.AdditionalCost != nil || len(spec.OptionalCosts) > 0 || spec.XCeiling != nil {
		panic(fmt.Sprintf("effects.Register: %q declares splice beside modes or a cost of its own — a spliced text carries neither (ADR 0202)", spec.Name))
	}
	if spec.Targets != nil {
		for _, c := range spec.Targets.Clauses() {
			if c.CountFromX || c.Divide != nil {
				panic(fmt.Sprintf("effects.Register: %q declares splice with an X- or division-bound target — the spliced spell announces neither for it (ADR 0202)", spec.Name))
			}
		}
	}
}

// checkSpliceCost refuses a non-mana splice cost that would compile and
// then pay something other than what the card prints: no label, no
// component, more than one component (each printed card has exactly one,
// and the payment names one list of objects per component), a variable
// count, or a graveyard cost that is not a fixed count from the
// graveyard.
func checkSpliceCost(name string, c *game.SpliceCost) {
	if c.Label == "" {
		panic(fmt.Sprintf("effects.Register: %q declares a splice cost with no label — the client shows it as printed", name))
	}
	n := 0
	if c.Sacrifice != nil {
		n++
		if c.Sacrifice.CountFromX || game.SacrificeCostVariable(c.Sacrifice) || game.SacrificeCostCount(c.Sacrifice) < 1 {
			panic(fmt.Sprintf("effects.Register: %q splice cost sacrifices a variable number — only a printed count is paid", name))
		}
	}
	if ec := c.ExileCards; ec != nil {
		n++
		if ec.CountFromX || ec.N < 1 || ec.Zone() != game.ZoneGraveyard {
			panic(fmt.Sprintf("effects.Register: %q splice cost must exile a fixed count from the graveyard", name))
		}
	}
	if tc := c.TapOthers; tc != nil {
		n++
		if tc.Empty() || tc.AnyNumber || tc.Count < 1 {
			panic(fmt.Sprintf("effects.Register: %q splice cost must tap a fixed count of permanents", name))
		}
	}
	if rc := c.ReturnToHand; rc != nil {
		n++
		if rc.Empty() {
			panic(fmt.Sprintf("effects.Register: %q splice cost returns nothing", name))
		}
	}
	if c.OpponentGainsLife != 0 {
		n++
		if c.OpponentGainsLife < 1 {
			panic(fmt.Sprintf("effects.Register: %q splice cost gives an opponent no life", name))
		}
	}
	if n != 1 {
		panic(fmt.Sprintf("effects.Register: %q splice cost has %d components; a printed splice cost has exactly one", name, n))
	}
}

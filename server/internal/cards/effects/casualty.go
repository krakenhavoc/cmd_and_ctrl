package effects

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// casualty.go — CR 702.153, casualty, and optional costs a permanent
// grants to spells (#3245). Append-only.
//
// The cost is an optional cost keyed game.CasualtyKey; the copy is the
// engine's trigger (game/casualty.go), read off the paid record, so a
// card declares only the cost:
//
//	OptionalCosts: []game.AdditionalCost{Casualty(2)},                       // a printed "Casualty 2"
//	OptionalCosts: []game.AdditionalCost{CasualtyX()},                       // Ob Nixilis, the Adversary
//	GrantedOptionalCosts: []game.GrantedOptionalCost{SpellsYouCastHave(       // Silverquill, the Disputant
//	    "Each instant and sorcery spell you cast has casualty 1.", Or(Instant(), Sorcery()), Casualty(1))},

// Casualty is "Casualty N": "As an additional cost to cast this spell,
// you may sacrifice a creature with power N or greater."
func Casualty(n int) game.AdditionalCost {
	return game.AdditionalCost{
		Optional:  true,
		Key:       game.CasualtyKey,
		Sacrifice: sacrificeSpec("a creature with power "+strconv.Itoa(n)+" or greater", Creature(), PowerGE(n)),
		Label:     "Casualty " + strconv.Itoa(n),
	}
}

// CasualtyX is "Casualty X. The copy isn't legendary and has starting
// loyalty X." (Ob Nixilis, the Adversary, #5668): casualty with a
// variable N (CR 702.153a, 107.3a). The caster may sacrifice any
// creature with power 0 or greater, and X is that creature's power as
// it last existed on the battlefield, the greatest X it can pay for.
// The engine's casualty trigger makes the copy with both exceptions
// (game/casualty.go), and CR 608.3f makes it a token as it resolves.
//
// Both exceptions are part of the helper, because every printed
// casualty X carries them. effects.Register allows it once, on a card
// with no other sacrifice among its additional costs.
func CasualtyX() game.AdditionalCost {
	return game.AdditionalCost{
		Optional:  true,
		Key:       game.CasualtyKey,
		Sacrifice: sacrificeSpec("a creature with power X", Creature(), PowerGE(0)),
		CasualtyX: true,
		Label:     "Casualty X",
	}
}

// checkCasualtyX is Register's rule for AdditionalCost.CasualtyX: only
// the optional cost CasualtyX builds, and the only sacrifice in the
// card's plan, because the casualty trigger reads X off the first
// permanent the cast sacrificed.
func checkCasualtyX(spec Spec) {
	if ac := spec.AdditionalCost; ac != nil {
		if ac.CasualtyX {
			panic(fmt.Sprintf("effects.Register: %q marks its mandatory additional cost CasualtyX — casualty is optional, build it with CasualtyX()", spec.Name))
		}
		for _, br := range ac.Either {
			if br.CasualtyX {
				panic(fmt.Sprintf("effects.Register: %q marks an either/or branch CasualtyX — build the optional cost with CasualtyX()", spec.Name))
			}
		}
	}
	x := -1
	for i, oc := range spec.OptionalCosts {
		if !oc.CasualtyX {
			continue
		}
		if oc.Key != game.CasualtyKey || oc.Sacrifice == nil || oc.Repeat > 1 {
			panic(fmt.Sprintf("effects.Register: %q declares a CasualtyX cost by hand — build it with CasualtyX()", spec.Name))
		}
		x = i
	}
	if x < 0 {
		return
	}
	if ac := spec.AdditionalCost; ac != nil {
		sacrifices := ac.Sacrifice != nil
		for _, br := range ac.Either {
			sacrifices = sacrifices || br.Sacrifice != nil
		}
		if sacrifices {
			panic(fmt.Sprintf("effects.Register: %q pairs casualty X with a mandatory sacrifice — X is read off the first permanent the cast sacrificed", spec.Name))
		}
	}
	for i, oc := range spec.OptionalCosts {
		if i != x && oc.Sacrifice != nil {
			panic(fmt.Sprintf("effects.Register: %q pairs casualty X with optional cost %q's sacrifice — X is read off the first permanent the cast sacrificed", spec.Name, oc.Key))
		}
	}
}

// SpellsYouCastHave is a permanent's "<spells> you cast have <optional
// cost>": the grant game.OptionalCostsForCastLocked appends to the
// caster's offer while the permanent is on the battlefield under their
// control. `spells` is read against the card being cast; nil is every
// spell.
func SpellsYouCastHave(label string, spells CardPredicate, cost game.AdditionalCost) game.GrantedOptionalCost {
	if cost.CasualtyX {
		panic("effects.SpellsYouCastHave: casualty X is a card's own cost, never a granted one")
	}
	return game.GrantedOptionalCost{
		Spells: func(g *game.Game, caster uuid.UUID, spell game.Card) bool {
			return spells == nil || spells(g, caster, spell)
		},
		Cost:  cost,
		Label: label,
	}
}

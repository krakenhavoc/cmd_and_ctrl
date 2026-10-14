package effects

import (
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

// SpellsYouCastHave is a permanent's "<spells> you cast have <optional
// cost>": the grant game.OptionalCostsForCastLocked appends to the
// caster's offer while the permanent is on the battlefield under their
// control. `spells` is read against the card being cast; nil is every
// spell.
func SpellsYouCastHave(label string, spells CardPredicate, cost game.AdditionalCost) game.GrantedOptionalCost {
	return game.GrantedOptionalCost{
		Spells: func(g *game.Game, caster uuid.UUID, spell game.Card) bool {
			return spells == nil || spells(g, caster, spell)
		},
		Cost:  cost,
		Label: label,
	}
}

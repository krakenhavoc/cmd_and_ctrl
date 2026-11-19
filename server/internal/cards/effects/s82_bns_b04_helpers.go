package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b04_helpers.go — shared pieces of S82 set BNS, batch 4 (#5929).

// s82Bns04DerivedFromOwnBasicLands is Star Compass's DerivedMatch: "a basic
// land you control" — Reflecting Pool's DerivedFromOwnLands narrowed to
// basics. Pair it with DerivedColorsOnly: "any color" cannot make {C}.
func s82Bns04DerivedFromOwnBasicLands() func(game.Card, uuid.UUID) bool {
	return func(c game.Card, controller uuid.UUID) bool {
		return c.Controller == controller && c.IsLand() && c.HasSupertype("Basic")
	}
}

// s82Bns04HandSizeExactly is "Activate only if you have exactly n cards in
// your hand" (Magus of the Library). A hand's size is public; its contents
// are not read.
func s82Bns04HandSizeExactly(n int) ActivationCondition {
	return func(g *game.Game, controller, _ uuid.UUID) bool {
		p := g.PlayerByIDForEffect(controller)
		return p != nil && p.Hand != nil && len(p.Hand.Cards) == n
	}
}

// s82Bns04PrintedVanishing is "that creature has vanishing" for the copiable
// values being copied: the card's catalog row carries the vanishing upkeep
// trigger, or its printed keywords name it.
func s82Bns04PrintedVanishing(v *game.PrintedValues) bool {
	for _, k := range v.Keywords {
		if strings.EqualFold(k, "vanishing") {
			return true
		}
	}
	spec, ok := Lookup(v.OracleID)
	if !ok {
		return false
	}
	for _, t := range spec.Triggered {
		if t.Keyword == cmdr2B17VanishingKeyword {
			return true
		}
	}
	return false
}

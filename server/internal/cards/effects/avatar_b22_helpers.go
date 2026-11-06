package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b22_helpers.go — shared pieces for Avatar: The Last Airbender
// batches 22 and 23 (#4459).

// avatarB22Earthbend is "earthbend N" as a triggered or activated
// ability's whole effect: the first still-legal "target land you
// control" (EarthbendTargets) becomes the 0/0 haste land creature with N
// counters. Pair it with Targeting(..., EarthbendTargets()) on a trigger
// or Targets: EarthbendTargets() on an activated ability.
func avatarB22Earthbend(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		return EarthbendFirstTarget(EarthbendCount(n))(item, NewContext(g, item))
	}
}

// avatarB22ColorsAmongSubtypeYouControl is "each color among Allies you
// control": the distinct colours of the permanents with the subtype the
// player controls, read through the layered view, so at most five
// ("multicolor" and "colorless" are not colours — Earthen Ally's
// ruling). The caller reads the subtype through its own text
// (source.TextSubtype, CR 612.2).
func avatarB22ColorsAmongSubtypeYouControl(g *game.Game, controller uuid.UUID, subtype string) int {
	seen := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !c.HasSubtype(subtype) {
			continue
		}
		for _, col := range c.EffectiveColors() {
			seen[strings.ToUpper(col)] = true
		}
	}
	n := 0
	for _, col := range []string{"W", "U", "B", "R", "G"} {
		if seen[col] {
			n++
		}
	}
	return n
}

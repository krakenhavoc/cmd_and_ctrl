package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// exile_grantor_check.go — the boot-time rules for
// game.AbilityCost.ExileGrantor (#4000, ExileTheGrantor).

// checkExileGrantor refuses an exile-the-grantor cost anywhere but a
// granted row that can pay it:
//
//   - on a card's own ability (granted false): the row has no grantor;
//   - in a bundle given to cards off the battlefield (Zone): only the
//     layer-6 battlefield grant records a grantor on the row;
//   - beside a component that picks permanents (sacrifice another,
//     return, exile, tap another, crew, saddle, waterbend): the grantor
//     would be a legal pick of that component too, and one permanent
//     pays one component (CR 118.3). No printed card does it.
func checkExileGrantor(card, where string, ab ActivatedAbility, granted bool, zone game.ZoneKind) {
	c := ab.Cost
	if !c.ExileGrantor {
		return
	}
	switch {
	case !granted:
		panic(fmt.Sprintf("effects.Register: %q %s exiles its grantor but is not a granted ability — only a granted row has a grantor (#4000)", card, where))
	case zone != "" && zone != game.ZoneBattlefield:
		panic(fmt.Sprintf("effects.Register: %q %s exiles its grantor from a bundle given to cards in %s — only a battlefield grant records one (#4000)", card, where, zone))
	case c.SacrificeOther != nil || !c.ReturnToHand.Empty() || !c.ExilePermanents.Empty() || !c.TapOthers.Empty() ||
		c.Crew > 0 || c.Saddle > 0 || !c.Waterbend.Empty():
		panic(fmt.Sprintf("effects.Register: %q %s exiles its grantor beside a component that picks permanents — not modelled (#4000, CR 118.3)", card, where))
	}
}

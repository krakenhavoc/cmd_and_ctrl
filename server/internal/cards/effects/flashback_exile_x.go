package effects

import (
	"fmt"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// flashback_exile_x.go — #6122: a flashback cost whose non-mana part
// fixes X. Summons of Saruman ({X}{U}{R}):
//
//	"Flashback—{3}{U}{R}, Exile X cards from your graveyard."
//
// The flashback cost is an alternative cost that includes X (CR 107.3a:
// "If a spell … has a mana cost, alternative cost, additional cost, and/or
// activation cost with an {X}, [-X], or X in it … the controller …
// chooses and announces the value of X"), so CR 107.3b does not lock X at
// 0: the caster announces X with the cast (CR 601.2b) and names exactly X
// cards in their graveyard on `alt_cost_ids`, which are exiled as the cost
// is paid (CR 601.2h). The spell reads that X with ctx.X() everywhere its
// text says X. X = 0 exiles nothing.

// FlashbackExileX is "Flashback—<mana>, Exile X cards from your
// graveyard" (#6122, CR 702.34a): Flashback's cast from the graveyard and
// its exile on leaving the stack, with a graveyard exile of X cards beside
// the mana, X being the spell's X. The card file lists ZoneGraveyard in
// CastableZones, as for Flashback, and declares XMatters.
//
//	AlternativeCosts: []game.AlternativeCost{FlashbackExileX("{3}{U}{R}")},
func FlashbackExileX(mana string) game.AlternativeCost {
	ac := Flashback(mana)
	ac.Label = "Flashback—" + mana + ", Exile X cards from your graveyard"
	spec := CardsInYourGraveyard(0, "X cards from your graveyard")
	spec.CountFromX = true
	ac.ExileFromGraveyard = spec
	ac.PayLabel = "X cards from your graveyard"
	return ac
}

// checkAltCostExileX is the boot-time rule for an alternative cost whose
// graveyard exile counts X (#6122): the spell must read X (XMatters), and
// nothing else in the offer may claim the same announced X — no {X} in
// its mana, no counter removal of X — since one X cannot pay two printed
// costs. Every other card component counts a printed number.
func checkAltCostExileX(spec Spec, ac game.AlternativeCost) {
	for _, s := range []*game.TargetSpec{ac.ExileFromHand, ac.ReturnToHand, ac.Sacrifice, ac.DiscardFromHand} {
		if s != nil && s.CountFromX {
			panic(fmt.Sprintf("effects.Register: %q alternative cost %q counts a component other than a graveyard exile by X — only FlashbackExileX does", spec.Name, ac.Key))
		}
	}
	if !ac.CardCountFromX() {
		return
	}
	switch {
	case !spec.XMatters:
		panic(fmt.Sprintf("effects.Register: %q alternative cost %q exiles X cards but the card does not read X (XMatters)", spec.Name, ac.Key))
	case strings.Contains(ac.ManaCost, "{X}"):
		panic(fmt.Sprintf("effects.Register: %q alternative cost %q exiles X cards AND has {X} in its mana — one announced X cannot pay both", spec.Name, ac.Key))
	case ac.CountersDefineX():
		panic(fmt.Sprintf("effects.Register: %q alternative cost %q exiles X cards AND removes X counters — one announced X cannot pay both", spec.Name, ac.Key))
	case ac.GraveyardTopOnly || len(ac.ExileFromGraveyard.EachOf) > 0 || ac.ExileFromGraveyard.ManaValueEqualsX:
		panic(fmt.Sprintf("effects.Register: %q alternative cost %q exiles X cards with a top-only, set or mana-value rule — no printed card has one", spec.Name, ac.Key))
	}
}

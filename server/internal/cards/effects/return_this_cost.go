package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// return_this_cost.go — #2028: activated rows whose cost returns the
// permanent itself to its owner's hand (ReturnThis, game.AbilityCost's
// ReturnSelf). The cost is paid at announce (CR 602.2b, CR 601.2h), so
// each ability resolves with its source already in a hand.
//
// Append-only, mechanic-named: the clone gate sees one body per shape.

// returnThisRegenerateRow is "Return this enchantment to its owner's
// hand: Regenerate target creature." — Broken Fall and Molting Skin,
// which print the same sentence.
func returnThisRegenerateRow() ActivatedAbility {
	return ActivatedAbility{
		Label:   "Return this enchantment to its owner's hand: Regenerate target creature.",
		Cost:    ReturnThis(),
		Targets: TargetCreature("target creature"),
		Effect:  regenerateTheTargetPermanent,
	}
}

// returnEachTargetCardToHand returns every still-legal graveyard card
// target to its owner's hand. A target that has gone is skipped
// (CR 608.2b).
func returnEachTargetCardToHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// shigekiDig is Shigeki, Jukai Visionary's first ability: "Reveal the
// top four cards of your library. You may put a land card from among
// them onto the battlefield tapped. Put the rest into your graveyard."
func shigekiDig(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	revealed := g.RevealTopOfLibraryForEffect(item.Controller, ctx.Source(), 4,
		"Shigeki, Jukai Visionary — revealed from the top of the library")
	return PutFromLibraryOntoBattlefield{
		Player:   item.Controller,
		Cards:    revealed,
		Match:    Land(),
		Max:      1,
		Optional: true,
		Tapped:   true,
		Label:    "Shigeki, Jukai Visionary — put a land card onto the battlefield tapped",
		Then:     PutRestIntoGraveyard,
	}.Apply(ctx)
}

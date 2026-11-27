package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b645_helpers.go — Invasion, "Activated abilities" (S82 OLD batch
// 645, #5616). Append-only.

// S82OldB645TargetGainsRow is "<cost>: Target <what> gains <keyword> until end
// of turn" where what is the whole printed target phrase after "Target"
// ("white or black creature"). preds narrow the creature target.
func S82OldB645TargetGainsRow(cost game.AbilityCost, costText, what, keyword string, preds ...CardPredicate) ActivatedAbility {
	label := costText + ": Target " + what + " gains " + keyword + " until end of turn."
	return ActivatedAbility{
		Label:   label,
		Cost:    cost,
		Targets: TargetCreature("target "+what, preds...),
		Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, keyword)),
		Effect:  EbTargetCreatureGainsUntilEOT(keyword, label),
	}
}

// S82OldB645ThisBecomesColorless is "<this> becomes colorless until end of
// turn" (Ancient Kavu): the layer-5 colour change with no colour (CR 105.2c).
func S82OldB645ThisBecomesColorless(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !SourceIsStillThisPermanent(g, item) {
			return nil
		}
		return UntilEndOfTurn(NewContext(g, item), item.SourceCardID, nil,
			name+" — colorless until end of turn", game.SetColorsMod())
	}
}

// S82OldB645TargetToTopOfLibrary is "Put target creature you control on top
// of its owner's library" (Nightscape Apprentice).
func S82OldB645TargetToTopOfLibrary(g *game.Game, item *game.StackItem) error {
	id, ok := B16FirstLegalTargetCard(NewContext(g, item))
	if !ok {
		return nil
	}
	return g.TuckToLibraryForEffect(id, false)
}

// S82OldB645ProtectionFromChosenColor is "Target creature you control gains
// protection from the color of your choice until end of turn": the target is
// chosen on activation, the colour as the ability resolves (Mother of Runes).
func S82OldB645ProtectionFromChosenColor(name string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		id, ok := B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		ChooseProtectionColorFor(ctx, item.Controller, id, name, name+" — choose a color")
		return nil
	}
}

// s82OldB645AerostatReturn is the rider of Metathran Aerostat's put: if a
// creature card entered, return the Aerostat to its owner's hand, as long as
// it is still the permanent that was activated (CR 400.7).
var s82OldB645AerostatReturn = PutFromHandTail("s82-old-b645/aerostat-return-this", func(ctx *Context, res PutFromHandResult) error {
	if res.Entered == uuid.Nil {
		return nil
	}
	return StdB9ReturnThisToHand(ctx)
})

// S82OldB645AerostatEffect is "You may put a creature card with mana value X
// from your hand onto the battlefield. If you do, return this creature to its
// owner's hand." The card file reads the announced X and passes it in.
func S82OldB645AerostatEffect(g *game.Game, item *game.StackItem, x int) error {
	ctx := NewContext(g, item)
	return PutFromHandOntoBattlefield{
		Match:    And(Creature(), ManaValueIs(x)),
		Optional: true,
		Label:    "Metathran Aerostat — you may put a creature card with mana value X from your hand onto the battlefield",
		ThenTail: s82OldB645AerostatReturn,
		ThenItem: true,
	}.Apply(ctx)
}

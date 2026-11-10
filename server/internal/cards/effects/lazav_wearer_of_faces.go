package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lazav, Wearer of Faces — Legendary Creature — Shapeshifter Detective {U}{B}
// (Murders at Karlov Manor, #4469):
//
//	"Whenever Lazav attacks, exile target card from a graveyard, then
//	 investigate.
//	 Whenever you sacrifice a Clue, you may have Lazav become a copy of a
//	 creature card exiled with it until end of turn."
//
// The exile is linked to this Lazav object (ExileTarget.Linked, CR 607.2a), and
// the Clue comes after it as the exile's keyed tail; a target that is no longer
// legal fizzles the whole ability, so no Clue (the ruling). The second ability
// is a "may" that is asked first; with creature cards exiled with Lazav it asks
// which (none asks nothing, one is not asked) and Lazav becomes a copy of that
// card's printed values until end of turn (BecomeCopy: copiable values read as
// the effect begins, CR 707.2). It sees any Clue you sacrifice, for any reason
// (the ruling), as the sacrifice happens, before the Clue leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ec0b64d-c793-47a6-9c26-9ad4d9e9c182",
		Name:         "Lazav, Wearer of Faces",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverThisAttacks("Lazav — exile target card from a graveyard, then investigate",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						t, ok := ctx.ClauseTarget(0)
						if !ok {
							return nil
						}
						return ExileTarget{Target: t.ID, Linked: true, ThenTail: ctx.ItemTail(lazavInvestigate)}.Apply(ctx)
					}),
				TargetCardInGraveyard("target card in a graveyard")),
			Optional(On(game.EventSacrifice, lonisYouSacrificedAClue,
				"Lazav — you may have Lazav become a copy of a creature card exiled with it until end of turn",
				lazavBecomeACopy),
				"Lazav, Wearer of Faces — become a copy of a creature card exiled with it until end of turn?"),
		},
	})
}

// lazavInvestigate is "then investigate", once the exile is done.
var lazavInvestigate = CardTail("tail/lazav-investigate", func(ctx *Context, _ game.TailOutcome) error {
	return Investigate{N: 1}.Apply(ctx)
})

// lazavCreatureCardsExiledWithIt is the creature cards in exile linked to this
// Lazav object.
func lazavCreatureCardsExiledWithIt(ctx *Context) []uuid.UUID {
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var out []uuid.UUID
	for _, c := range ctx.Game.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsCreature() {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func lazavBecomeACopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cards := lazavCreatureCardsExiledWithIt(ctx)
	switch len(cards) {
	case 0:
		return nil
	case 1:
		return lazavCopy(ctx, cards[0])
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Lazav — choose a creature card exiled with it to become a copy of",
		Cards:    cards,
		Min:      1,
		Max:      1,
		Asked:    game.NewAskedItem(item),
		ThenKey:  lazavCopyChosen,
	})
	return nil
}

var lazavCopyChosen = game.RegisterChooseCardsThen("choose-cards/lazav-wearer-of-faces-copy", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) != 1 || r.Asked == nil {
		return nil
	}
	return lazavCopy(NewContext(g, r.Asked), r.Picked[0])
})

func lazavCopy(ctx *Context, card uuid.UUID) error {
	if z := ctx.Game.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	self := ctx.Source()
	if !onBattlefield(ctx.Game, self) {
		return nil
	}
	return BecomeCopy{
		Targets:  []uuid.UUID{self},
		Of:       card,
		Duration: CopyUntilEndOfTurn,
		Label:    "Lazav, Wearer of Faces — a copy of a creature card exiled with it",
	}.Apply(ctx)
}

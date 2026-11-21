package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yggdrasil, Rebirth Engine — Legendary Artifact {3} (Assassin's Creed,
// #5948):
//
//	"When Yggdrasil enters, exile all creature cards from your graveyard.
//	 {T}: Exile the top three cards of your library.
//	 {4}, {T}: Put a creature card exiled with Yggdrasil onto the battlefield
//	 under your control. It gains haste until end of turn. Activate only as a
//	 sorcery."
//
// Both exiling abilities are linked to this Yggdrasil object (CR 607.2a): the
// cards land in exile with Card.ExiledWith naming it, so a second Yggdrasil's
// exiles, and those of an earlier object that left and came back (CR 400.7),
// are not "exiled with" this one. The entry trigger exiles every creature
// card in the controller's graveyard together; the tap ability exiles the top
// three cards of the library (fewer if there are fewer). The last ability
// chooses among the creature cards exiled with this object as it resolves,
// puts the one picked onto the battlefield under the controller's control,
// whoever owns it, and gives it haste until end of turn once it has entered.
// With none exiled it does nothing. Sorcery speed only.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7617a6f7-c83b-4ee5-a681-c7e164456526",
		Name:         "Yggdrasil, Rebirth Engine",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Yggdrasil — exile all creature cards from your graveyard", yggdrasilExileGraveyard),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Exile the top three cards of your library.",
				Cost:    TapCost(),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect:  yggdrasilExileTop,
			},
			{
				Label:        "{4}, {T}: Put a creature card exiled with Yggdrasil onto the battlefield under your control. It gains haste until end of turn. Activate only as a sorcery.",
				Cost:         Plus(ManaCost("{4}"), TapCost()),
				SorcerySpeed: true,
				Effect:       yggdrasilOffer,
			},
		},
	})
}

func yggdrasilExileGraveyard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	ids := graveyardIDs(g, ctx.Controller(), game.Card.IsCreature)
	if len(ids) == 0 {
		return nil
	}
	return g.ExileLinkedKeyedForEffect(ref, ids, game.KeyedTail{})
}

func yggdrasilExileTop(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	top := g.LookAtTopOfLibraryForEffect(ctx.Controller(), 3)
	if len(top) == 0 {
		return nil
	}
	return g.ExileLinkedKeyedForEffect(ref, top, game.KeyedTail{})
}

func yggdrasilOffer(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var cands []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if c.IsCreature() {
			cands = append(cands, c.InstanceID)
		}
	}
	if len(cands) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Yggdrasil — put a creature card exiled with it onto the battlefield under your control",
		Cards:    cands,
		Min:      1,
		Max:      1,
		ThenKey:  yggdrasilPicked,
		LiveZone: game.ZoneExile,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// yggdrasilPicked returns the chosen card under the controller's control and
// gives it haste once it has entered. An on-disk key.
var yggdrasilPicked = itemChoice("choose-cards/yggdrasil-return", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return ReturnFromExile{Target: r.Picked[0], Controller: ctx.Controller(), ThenTail: ctx.Tail(yggdrasilHaste)}.Apply(ctx)
})

// yggdrasilHaste is "it gains haste until end of turn", told the new object.
var yggdrasilHaste = EnteredTail("yggdrasil/haste", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil {
		return nil
	}
	return GrantKeywordUntilEOT{Target: entered, Keywords: []string{"haste"}, Label: "Yggdrasil — haste"}.Apply(ctx)
})

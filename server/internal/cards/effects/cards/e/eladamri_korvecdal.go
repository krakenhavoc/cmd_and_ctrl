package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eladamri, Korvecdal — Legendary Creature — Elf Warrior {1}{G}{G}, 3/3
// (Modern Horizons 3):
//
//	"You may look at the top card of your library any time.
//	 You may cast creature spells from the top of your library.
//	 {G}, {T}, Tap two untapped creatures you control: Reveal a card from your
//	 hand or the top card of your library. If you reveal a creature card this
//	 way, put it onto the battlefield. Activate only during your turn."
//
// The reveal is the player's choice of the hand or the library's top; a card
// from the hand is picked, then revealed, and only a creature card enters.
func init() {
	const label = "{G}, {T}, Tap two untapped creatures you control: Reveal a card from your hand or the top card of your library. If you reveal a creature card this way, put it onto the battlefield. Activate only during your turn."
	effects.Register(effects.Spec{
		OracleID:          "9aff3d55-e25d-4f06-a783-c3aa9eb9fd50",
		Name:              "Eladamri, Korvecdal",
		Completeness:      effects.CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			effects.PlayFromTopOfYourLibrary(game.PermissionFilter{CreatureOnly: true},
				"Cast a creature spell from the top of your library (Eladamri, Korvecdal)"),
		},
		Activated: []effects.ActivatedAbility{{
			Label: label,
			Cost: effects.Plus(effects.ManaCost("{G}"), effects.TapCost(), game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:         2,
				Filter:        effects.TargetPermanent("two untapped creatures you control", effects.Creature()),
				ExcludeSource: true,
				Label:         "two untapped creatures you control",
			}}),
			Condition: effects.DuringYourTurn(),
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Effect:    gapB51EladamriReveal,
		}},
	})
}

func gapB51EladamriReveal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	p := ctx.PlayerByID(item.Controller)
	if p == nil {
		return nil
	}
	var options []game.ChoiceOption
	hasHand := p.Hand != nil && p.Hand.Size() > 0
	hasTop := p.Library != nil && p.Library.Size() > 0
	if hasHand {
		options = append(options, game.ChoiceOption{Label: "Reveal a card from your hand"})
	}
	if hasTop {
		options = append(options, game.ChoiceOption{Label: "Reveal the top card of your library"})
	}
	switch len(options) {
	case 0:
		return nil
	case 1:
		return gapB51EladamriChoose(ctx, hasHand)
	}
	return effects.PickOption{
		Question: "Eladamri, Korvecdal — reveal a card from your hand or the top card of your library",
		Options:  options,
		ThenKey:  gapB51EladamriSource,
		WithItem: true,
	}.Apply(ctx)
}

// gapB51EladamriSource is the player's pick between the hand (0) and the top of
// the library (1); an on-disk key, never renamed, never reused.
var gapB51EladamriSource = effects.OptionPickThen("option-pick/eladamri-korvecdal-source",
	func(ctx *effects.Context, r game.OptionPicked) error {
		return gapB51EladamriChoose(ctx, r.Index == 0)
	})

func gapB51EladamriChoose(ctx *effects.Context, fromHand bool) error {
	p := ctx.PlayerByID(ctx.Controller())
	if p == nil {
		return nil
	}
	if !fromHand {
		if p.Library == nil || p.Library.Size() == 0 {
			return nil
		}
		top := p.Library.Cards[len(p.Library.Cards)-1]
		return gapB51EladamriRevealAndPut(ctx, top.InstanceID, game.ZoneLibrary)
	}
	var cands []uuid.UUID
	for _, c := range p.Hand.Cards {
		cands = append(cands, c.InstanceID)
	}
	if len(cands) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Eladamri, Korvecdal — reveal a card from your hand",
		Cards:    cands,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneHand,
		ThenKey:  gapB51EladamriPick,
		Asked:    game.NewAskedItem(ctx.Item),
	})
	return nil
}

var gapB51EladamriPick = effects.AskedPick("choose-cards/eladamri-korvecdal-reveal", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	return gapB51EladamriRevealAndPut(ctx, r.Picked[0], game.ZoneHand)
})

func gapB51EladamriRevealAndPut(ctx *effects.Context, id uuid.UUID, from game.ZoneKind) error {
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != from {
		return nil
	}
	ctx.Game.RevealForEffect(game.RevealSpec{Player: ctx.Controller(), Source: ctx.Source(), Cards: []uuid.UUID{id},
		Reason: "Eladamri, Korvecdal — reveal a card"})
	c, ok := ctx.Game.LookupCardForEffect(id)
	if !ok || !c.IsCreature() {
		return nil
	}
	return effects.PutCardsFromZoneTogether(ctx, from, []uuid.UUID{id}, ctx.Controller(), false)
}

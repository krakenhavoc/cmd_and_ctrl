package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2S44GontiLook is Gonti, Lord of Luxury's enters trigger: look at the
// top four cards of the target opponent's library, exile one face down, put
// the rest on the bottom in a random order. The looker alone may play the
// exiled card for as long as it stays exiled, spending mana of any type.
func cmdr2S44GontiLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var opp uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			opp = t.ID
			break
		}
	}
	if opp == uuid.Nil {
		return nil
	}
	controller, source := item.Controller, item.SourceCardID
	looked := g.LookAtTopOfPlayersLibraryForEffect(controller, opp, 4)
	if len(looked) == 0 {
		return nil
	}
	if len(looked) == 1 {
		return cmdr2S44GontiFinish(g, controller, source, looked, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    controller,
		FromPlayer: opp,
		Source:     source,
		Question:   "Gonti, Lord of Luxury — exile one of them face down",
		Cards:      looked,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneLibrary,
		ThenKey:    cmdr2S44GontiPicked,
		Carry:      looked,
	})
	return nil
}

// cmdr2S44GontiPicked is the exiled card chosen; the carry is the four looked
// at. An on-disk key: never renamed, never reused.
var cmdr2S44GontiPicked = game.RegisterChooseCardsThen("choose-cards/otc-gonti-lord-of-luxury-exile", func(g *game.Game, r game.CardsChosen) error {
	return cmdr2S44GontiFinish(g, r.Chooser, r.Source, r.Carry, r.Picked)
})

func cmdr2S44GontiFinish(g *game.Game, controller, source uuid.UUID, looked, picked []uuid.UUID) error {
	var firstErr error
	if len(picked) > 0 {
		firstErr = g.ExileCardsFaceDownWithPermissionForEffect(picked, controller, game.CastPermission{
			AnyColor:   true,
			AnyType:    true,
			Duration:   game.WhileInZoneDuration(),
			Source:     source,
			SourceName: "Gonti, Lord of Luxury",
			Label:      "You may cast that card for as long as it remains exiled; mana of any type can be spent to cast that spell",
			CastOnly:   true,
		})
	}
	took := make(map[uuid.UUID]bool, len(picked))
	for _, id := range picked {
		took[id] = true
	}
	var rest []uuid.UUID
	for _, id := range cardsStillInALibrary(g, looked) {
		if !took[id] {
			rest = append(rest, id)
		}
	}
	if err := g.PutOnBottomInRandomOrderForEffect(controller, game.ZoneLibrary, rest); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

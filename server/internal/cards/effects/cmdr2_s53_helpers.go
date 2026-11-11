package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s53_helpers.go — card-side helpers for the six Commander-deck
// cards of batch 53 (#5368).

// cmdr2S53TargetPlayerBottomsTheirGraveyard is "up to one target player
// puts all the cards from their graveyard on the bottom of their library
// in a random order" (Endurance). With no target chosen, or a target
// that is illegal as the ability resolves, nothing moves (CR 608.2b).
// The cards are the ones in the graveyard as the ability resolves; each
// one is a zone change of its own, in an order nobody chooses or sees.
func cmdr2S53TargetPlayerBottomsTheirGraveyard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetPlayer {
			continue
		}
		p := g.PlayerByIDForEffect(t.ID)
		if p == nil {
			return nil
		}
		ids := make([]uuid.UUID, 0, len(p.Graveyard.Cards))
		for _, c := range p.Graveyard.Cards {
			ids = append(ids, c.InstanceID)
		}
		return g.PutOnBottomInRandomOrderForEffect(t.ID, game.ZoneGraveyard, ids)
	}
	return nil
}

// cmdr2S53SiphonInsightLook is Siphon Insight's resolution: look at the
// top two cards of the target opponent's library, exile one face down
// and put the other on the bottom of that library. The looker alone may
// play the exiled card for as long as it stays exiled, spending mana as
// though it were mana of any color.
func cmdr2S53SiphonInsightLook(g *game.Game, item *game.StackItem) error {
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
	controller, source := item.Controller, ctx.Source()
	looked := g.LookAtTopOfPlayersLibraryForEffect(controller, opp, 2)
	if len(looked) == 0 {
		return nil
	}
	if len(looked) == 1 {
		return cmdr2S53SiphonInsightFinish(g, controller, source, looked, looked)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    controller,
		FromPlayer: opp,
		Source:     source,
		Question:   "Siphon Insight — exile one of them face down",
		Cards:      looked,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneLibrary,
		ThenKey:    cmdr2S53SiphonInsightPicked,
		Carry:      looked,
	})
	return nil
}

// cmdr2S53SiphonInsightPicked is the exiled card chosen; the carry is the
// two looked at. An on-disk key: never renamed, never reused.
var cmdr2S53SiphonInsightPicked = game.RegisterChooseCardsThen("choose-cards/cmdr2-s53-siphon-insight-exile", func(g *game.Game, r game.CardsChosen) error {
	return cmdr2S53SiphonInsightFinish(g, r.Chooser, r.Source, r.Carry, r.Picked)
})

func cmdr2S53SiphonInsightFinish(g *game.Game, controller, source uuid.UUID, looked, picked []uuid.UUID) error {
	var firstErr error
	if len(picked) > 0 {
		firstErr = g.ExileCardsFaceDownWithPermissionForEffect(picked, controller, game.CastPermission{
			AnyColor:   true,
			Duration:   game.WhileInZoneDuration(),
			Source:     source,
			SourceName: "Siphon Insight",
			Label:      "You may play that card for as long as it remains exiled; you may spend mana as though it were mana of any color to cast that spell",
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

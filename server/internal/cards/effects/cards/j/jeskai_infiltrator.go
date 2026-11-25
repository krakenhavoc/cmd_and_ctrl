package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Jeskai Infiltrator — {2}{U} Creature — Human Monk 2/3:
//
//	"This creature can't be blocked as long as you control no other
//	 creatures.
//	 When this creature deals combat damage to a player, exile it and the
//	 top card of your library in a face-down pile, shuffle that pile, then
//	 manifest those cards."
//
// The evasion is Watertight Gondola's CantBeBlockedWhile over the controller's
// other creatures, read as blockers are declared. The trigger exiles the
// Infiltrator (when it is still the same object on the battlefield, CR 400.7)
// and the top card of the library together, then manifests both as one entry
// (Ghastly Conscription's tail). The manifested cards enter face down and
// their controller alone knows which is which (CR 708.5), so shuffling the
// pile changes nothing anyone can tell and the entry order is not one. A
// manifested card need not be a permanent card (CR 701.40a), so an instant
// or sorcery is manifested too. The face-down pile groups cards into a pile
// (CR 700.3), so the ability notes the turn's pile-or-guess tally (#6175).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "39b359c1-48b8-43a7-b972-54764eadd2c7",
		Name:         "Jeskai Infiltrator",
		Completeness: effects.CompletenessFull,
		BlockRules:   []game.BlockRule{effects.CantBeBlockedWhile(effects.OnSelf(), jeskaiInfiltratorAlone)},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(
				"Jeskai Infiltrator — exile it and the top card of your library face down, shuffle, then manifest those cards",
				jeskaiInfiltratorPile),
		},
	})
}

// jeskaiInfiltratorAlone is "you control no other creatures".
func jeskaiInfiltratorAlone(g *game.Game, you uuid.UUID, self game.Card) bool {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == you && c.InstanceID != self.InstanceID && c.IsCreature() {
			return false
		}
	}
	return true
}

func jeskaiInfiltratorPile(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	effects.NotePileOrGuess(ctx)
	var ids []uuid.UUID
	if !g.AbilitySourceGoneForEffect(item) {
		ids = append(ids, item.SourceCardID)
	}
	if p := g.PlayerByIDForEffect(item.Controller); p != nil && len(p.Library.Cards) > 0 {
		ids = append(ids, p.Library.Cards[len(p.Library.Cards)-1].InstanceID)
	}
	if len(ids) == 0 {
		return nil
	}
	return g.ExileCardsKeyedForEffect(ids, ctx.ItemTail(jeskaiInfiltratorExiled))
}

// jeskaiInfiltratorExiled manifests the exiled cards, all at once. An
// on-disk key: never renamed, never reused.
var jeskaiInfiltratorExiled = effects.CardTail("then/jeskai-infiltrator-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	var entries []game.BatchEntry
	for _, id := range o.IDs {
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
			entries = append(entries, game.BatchEntry{CardID: id, From: game.ZoneExile})
		}
	}
	if len(entries) == 0 {
		return nil
	}
	return ctx.Game.PutOntoBattlefieldTogetherKeyedForEffect(entries, game.ZoneEntryOptions{
		Controller: ctx.Controller(),
		FaceDown:   game.FaceDownManifested,
	}, game.KeyedTail{})
})

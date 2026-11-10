package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b43_cloak.go — cloak as an effect (CR 701.58, #4452): the doors in
// game/cloak.go as primitives for a card file. CloakTopOfLibrary is
// cloak_action.go's.

// CloakACardFromYourHand is "cloak a card from your hand" (Vannifar,
// Evolved Enigma): the controller names one card in their hand and it enters
// face down as a 2/2 with ward {2} (CR 701.58a). A hand of one card is not
// asked; an empty hand does nothing. The pick is a keyed choose_cards prompt,
// so a table waiting on it is a restore point.
type CloakACardFromYourHand struct{}

func (CloakACardFromYourHand) Apply(ctx *Context) error {
	g := ctx.Game
	hand := handCardIDs(g, ctx.Controller())
	switch len(hand) {
	case 0:
		return nil
	case 1:
		return cloakFromHand(g, ctx.Controller(), hand)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: ctx.Controller(),
		Source:     ctx.Source(),
		Question:   "Choose a card from your hand to cloak",
		Cards:      hand,
		Min:        1,
		Max:        1,
		Zone:       game.ZoneHand,
		LiveZone:   game.ZoneHand,
		ThenKey:    cloakFromHandPicked,
	})
	return nil
}

var cloakFromHandPicked = game.RegisterChooseCardsThen("choose-cards/cloak-a-card-from-your-hand", func(g *game.Game, r game.CardsChosen) error {
	return cloakFromHand(g, r.Chooser, r.Picked)
})

// cloakFromHand puts the named cards of `player`'s hand onto the battlefield
// face down, cloaked, under their control.
func cloakFromHand(g *game.Game, player uuid.UUID, ids []uuid.UUID) error {
	var live []uuid.UUID
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneHand {
			live = append(live, id)
		}
	}
	if len(live) == 0 {
		return nil
	}
	return g.PutCardsFromHandOntoBattlefieldKeyedForEffect(live,
		game.HandEntryOptions{Controller: player, FaceDown: game.FaceDownCloaked}, game.KeyedTail{})
}

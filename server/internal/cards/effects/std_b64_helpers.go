package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// std_b64_helpers.go — shared bodies for Lorwyn Eclipsed leftovers, batch 64
// (#4458).

// stdB64YouCastDuringAnOpponentsTurn is "whenever you cast a spell during an
// opponent's turn": the caster is the source's controller and is not the
// active player (Katara, Waterbending Master's condition).
func stdB64YouCastDuringAnOpponentsTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return ev.Actor == source.Controller && !isActivePlayer(g, source.Controller)
}

// stdB64PermanentCardsToYourGraveyard is "one or more permanent cards are
// put into your graveyard from anywhere" (Moonshadow): every route a card
// takes into a graveyard, a battlefield death included, judged as the card
// it is in the graveyard (CR 603.6c). A token is never one (CR 108.2b).
var stdB64PermanentCardsToYourGraveyard = graveyardArrivals{
	Match: func(c game.Card) bool {
		return c.IsCreature() || c.IsArtifact() || c.IsEnchantment() || c.IsLand() || c.IsPlaneswalker() || c.IsBattle()
	},
}

// stdB64RemoveAMinusCounterFromThis is the step "remove a -1/-1 counter from
// this creature": nothing when it has none, or when the creature the
// ability came from is no longer on the battlefield (CR 400.7).
func stdB64RemoveAMinusCounterFromThis(ctx *Context) error {
	self, ok := ctx.Game.LookupCardForEffect(ctx.Source())
	if !ok || self.Counters[game.CounterMinusOne] <= 0 {
		return nil
	}
	return AddCounter{Target: self.InstanceID, Kind: game.CounterMinusOne, N: -1}.Apply(ctx)
}

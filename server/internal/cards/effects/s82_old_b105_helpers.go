package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b105_helpers.go — shared pieces of Core Set 2021 batch 105 (#5511).

// S82OldB105ShrinesYouControl is "the number of Shrines you control": the
// battlefield permanents `player` controls with the Shrine subtype, read
// through the layered view.
func S82OldB105ShrinesYouControl(g *game.Game, player uuid.UUID) int {
	return CountControlled(g, player, func(c game.Card) bool { return c.HasSubtype("Shrine") })
}

// S82OldB105MayFetchNamed is "You may search your library and/or graveyard
// for a card named <named>, reveal it, and put it into your hand. If you
// search your library this way, shuffle." as a spell's last step (Liliana's
// Scorn, Basri's Aegis): the same keyed answer the enters-trigger cycle uses.
func S82OldB105MayFetchNamed(self, named string) func(ctx *Context) error {
	return func(ctx *Context) error {
		return MayChoice{
			Question: self + " — search your library and/or graveyard for a card named " + named + "?",
			Then:     MaySearchNamedLibraryOrGraveyard,
			Word:     named,
		}.Apply(ctx)
	}
}

// S82OldB105SacrificeCreatureOrDiscardCreatureCard is the payment "sacrifice
// a creature or discard a creature card" (Crypt Lurker): an either/or payment
// whose branches are offered only to a player who can make them (CR 118.3).
func S82OldB105SacrificeCreatureOrDiscardCreatureCard() UpkeepPayment {
	discard := DiscardPayment(1)
	discard.Action.CardType = "creature"
	discard.Words = "Discard a creature card"
	return EitherPayment(SacrificePayment(1, "creature", "creatures", game.PermanentQuery{Types: []string{"creature"}}), discard)
}

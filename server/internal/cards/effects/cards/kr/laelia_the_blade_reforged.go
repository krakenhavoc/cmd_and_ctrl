package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Laelia, the Blade Reforged — Legendary Creature — Spirit Warrior
// {2}{R}, 2/2 (EDHREC rank 1321):
//
//	"Haste
//	 Whenever Laelia attacks, exile the top card of your library. You
//	 may play that card this turn.
//	 Whenever one or more cards are put into exile from your library
//	 and/or your graveyard, put a +1/+1 counter on Laelia."
//
// Mono-red card advantage on a body that grows every time it draws.
// The attack trigger is the impulse-exile primitive with a "play"
// grant (a land off the top can be the turn's land drop). The growth
// trigger watches the zone-move event every exile path emits and
// reads where the card came FROM: Laelia's own attack exile counts,
// as printed, so an attack is +1 card and +1/+1 in one go.
//
// "One or more" is one trigger per batch, and the engine emits one
// event per card — so the counter trigger declines every later event
// of the same batch, matched by label so the attack trigger is not
// mistaken for it (OncePerBatch; see docs/adding-cards.md). A Bonehoard
// Dracosaur upkeep that exiles two cards grows Laelia once, as
// printed; a second batch that arrives after the first trigger has
// resolved is a second trigger, which is also what paper does.
//
// No simplification.
func init() {
	const grow = "Laelia, the Blade Reforged — a +1/+1 counter"
	effects.Register(effects.Spec{
		OracleID:        "a0be9bb2-3234-4c6c-b8ce-0879b1f43003",
		Name:            "Laelia, the Blade Reforged",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Laelia, the Blade Reforged — exile the top card, play it this turn", func(g *game.Game, item *game.StackItem) error {
				_, err := effects.B12ImpulseExileForTurn(g, item, 1)
				return err
			}),
			effects.OncePerBatch(effects.On(game.EventZoneMove, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B12CardExiledFromYourLibraryOrGraveyard(ev, source, g)
			}, grow, effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}

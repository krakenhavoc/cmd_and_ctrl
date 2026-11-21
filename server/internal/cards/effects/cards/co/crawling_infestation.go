package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crawling Infestation — {2}{G} Enchantment (Innistrad: Crimson Vow, S82):
//
//	"At the beginning of your upkeep, you may mill two cards.
//	 Whenever one or more creature cards are put into your graveyard
//	 from anywhere during your turn, create a 1/1 green Insect creature
//	 token. This ability triggers only once each turn."
//
// Crawling Sensation's two abilities (an optional upkeep mill and a
// once-a-turn batch trigger over every zone move into the graveyard)
// with creature cards instead of lands and "during your turn": the
// event must happen while its controller is the active player. A token
// is not a card (CR 108.2), so a dying Insect token does not count.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a42262b8-8a26-4570-a6ee-24b239d095fc",
		Name:         "Crawling Infestation",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Crawling Infestation — you may mill two cards", effects.Do(effects.MillCards{N: 2})),
				"Crawling Infestation — mill two cards?"),
			effects.TriggersOnlyOnceEachTurn(effects.OncePerBatch(effects.OnAny([]game.EventKind{game.EventZoneMove, game.EventLTB, game.EventMill}, vowCreatureCardToYourGraveyardOnYourTurn,
				"Crawling Infestation — create a 1/1 green Insect creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Insect"), N: 1})))),
		},
	})
}

// vowCreatureCardToYourGraveyardOnYourTurn is "a creature card is put
// into your graveyard from anywhere during your turn" (no OldZone check).
func vowCreatureCardToYourGraveyardOnYourTurn(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if source == nil || ev.NewZone != game.ZoneGraveyard || !effects.IsYourTurn(g, source.Controller) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && !effects.IsToken(c) && c.IsCreature() && c.Owner == source.Controller
}

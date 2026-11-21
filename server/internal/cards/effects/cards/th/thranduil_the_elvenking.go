package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thranduil, the Elvenking — Legendary Creature — Elf Noble
// {2}{B}{G}{U}, 5/6:
//
//	"Thranduil has all activated abilities of all Elf cards in your
//	 graveyard.
//	 Whenever another legendary Elf you control enters, draw two cards,
//	 then discard a card."
//
// The borrow is a layer-6 lend over the controller's graveyard (ADR
// 0155); Elf is read off the card's printed subtypes. The trigger is one
// per legendary Elf entering under your control, the source excluded, and
// loots through the shared draw-then-discard continuation, so the discard
// is the player's own choice from the post-draw hand.
//
// No simplification.
func init() {
	const clause = "Thranduil has all activated abilities of all Elf cards in your graveyard."
	effects.Register(effects.Spec{
		OracleID:     "4a5f76e7-40be-4b06-9935-4a3b2672e1c2",
		Name:         "Thranduil, the Elvenking",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{effects.HasActivatedAbilitiesOf(game.AbilityBorrow{
			Label:   clause,
			Lenders: effects.GraveyardLenders(effects.YourGraveyard, effects.ByController(effects.HasSubtype("Elf"))),
		})},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
				// "Elf" is a word of the source's text (CR 612.2, #3984).
				return ok && c.IsLegendary() && c.HasSubtype(source.TextSubtype("Elf"))
			}, "Thranduil — draw two cards, then discard a card", func(g *game.Game, item *game.StackItem) error {
				return effects.B16DrawThenDiscard(g, item, 2)
			}),
		},
	})
}

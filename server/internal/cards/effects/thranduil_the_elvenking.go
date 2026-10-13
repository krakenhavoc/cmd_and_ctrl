package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "4a5f76e7-40be-4b06-9935-4a3b2672e1c2",
		Name:         "Thranduil, the Elvenking",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{HasActivatedAbilitiesOf(game.AbilityBorrow{
			Label:   clause,
			Lenders: GraveyardLenders(YourGraveyard, ByController(HasSubtype("Elf"))),
		})},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && c.IsLegendary() && c.HasSubtype("Elf")
			}, "Thranduil — draw two cards, then discard a card", func(g *game.Game, item *game.StackItem) error {
				return b16DrawThenDiscard(g, item, 2)
			}),
		},
	})
}

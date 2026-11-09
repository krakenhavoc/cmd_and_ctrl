package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sarkhan, Dragon Ascendant — Legendary Creature — Human Druid {1}{R}, 2/2:
//
//	"When Sarkhan enters, you may behold a Dragon. If you do, create a
//	 Treasure token. (To behold a Dragon, choose a Dragon you control or
//	 reveal a Dragon card from your hand.)
//	 Whenever a Dragon you control enters, put a +1/+1 counter on
//	 Sarkhan. Until end of turn, Sarkhan becomes a Dragon in addition to
//	 his other types and gains flying."
//
// Behold is asked as the enters trigger resolves: one pick across the
// Dragons you control and the Dragon cards in your hand (the cost walk's
// candidates), optional; a card from your hand is revealed, a permanent is
// only chosen, and the Treasure follows either. With nothing to behold
// nothing is asked. The second ability counts Sarkhan himself entering as a
// Dragon, since "a Dragon you control" is not "another"; it does nothing
// to a Sarkhan that has left, and "becomes a Dragon" adds the type until
// end of turn without losing the others.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "71cb48e2-2328-471c-a04a-7b6461f378f4",
		Name:         "Sarkhan, Dragon Ascendant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sarkhan, Dragon Ascendant — you may behold a Dragon; if you do, create a Treasure token",
				func(g *game.Game, item *game.StackItem) error {
					cmdrB5OfferBehold(g, item, "Dragon", sarkhanBeholdPicked,
						"Sarkhan, Dragon Ascendant — behold a Dragon (choose a Dragon you control or reveal a Dragon card from your hand) to create a Treasure token?")
					return nil
				}),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Dragon"))
			}, "Sarkhan, Dragon Ascendant — put a +1/+1 counter on Sarkhan; he becomes a Dragon and gains flying until end of turn",
				Do(
					CounterOnThis{Kind: game.CounterPlusOne, N: 1},
					Instruction(func(ctx *Context) error {
						if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
							return nil
						}
						return untilEndOfTurn(ctx, ctx.Source(), nil,
							"Sarkhan, Dragon Ascendant — a Dragon with flying until end of turn",
							game.AddSubtypesMod(ctx.TextSubtype("Dragon")), game.AddKeywordsMod("flying"))
					}),
				)),
		},
	})
}

var sarkhanBeholdPicked = game.RegisterChooseCardsThen("choose-cards/sarkhan-dragon-ascendant-behold",
	func(g *game.Game, r game.CardsChosen) error {
		if !cmdrB5RevealBeheld(g, r, "Dragon") {
			return nil
		}
		return CreateToken{Controller: r.Chooser, Template: TreasureToken(), N: 1}.Apply(keyedContext(g, r.Chooser, r.Source))
	})

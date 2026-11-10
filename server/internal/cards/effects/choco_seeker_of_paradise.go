package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Choco, Seeker of Paradise — Legendary Creature — Bird {1}{G}{W}{U},
// 3/5 (Final Fantasy, #4462):
//
//	"Whenever one or more Birds you control attack, look at that many
//	 cards from the top of your library. You may put one of them into
//	 your hand. Then put any number of land cards from among them onto
//	 the battlefield tapped and the rest into your graveyard.
//	 Landfall — Whenever a land you control enters, Choco gets +1/+0
//	 until end of turn."
//
// The attack trigger is Poetic Ingenuity's: one per declaration
// (OncePerBatch), "that many" the Birds declared in that batch, counted
// from the events so one that left combat by resolution still counts.
// The look is Commune with Evil's keyed take, with "you may" (up to one);
// its rest is the cards still in the library, of which any number of
// lands enter tapped as one batch (Famished Worldsire's put, and the card
// taken is no longer among them), and whatever then remains goes to the
// graveyard (PutRestIntoGraveyard, a routed move rather than a mill). The
// landfall is the Chocobo token's pump on Choco herself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7979bf83-5091-4e62-8885-ca1833a97306",
		Name:         "Choco, Seeker of Paradise",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && attackerHasSubtype(g, ev, source.TextSubtype("Bird"))
			}, "Choco, Seeker of Paradise — look at that many cards; you may put one into your hand, then lands onto the battlefield tapped and the rest into your graveyard",
				chocoLookAtBirdsManyCards)),
			Landfall("Choco, Seeker of Paradise — landfall: +1/+0 until end of turn",
				thisGetsUntilEndOfTurn(1, 0, "Choco, Seeker of Paradise — +1/+0 until end of turn")),
		},
	})
}

func chocoLookAtBirdsManyCards(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := attackersOfSubtypeInTheSameBatch(g, item.Trigger.Event, item.Controller, g.TextReaderForEffect(item).TextSubtype("Bird"))
	if n <= 0 {
		return nil
	}
	player := ctx.Controller()
	return TakeFromLibraryToHand{
		Player:   player,
		Cards:    g.LookAtTopOfLibraryForEffect(player, n),
		Max:      1,
		Optional: true,
		Label:    "Choco, Seeker of Paradise — you may put one of them into your hand",
		Key:      chocoTake,
	}.Apply(ctx)
}

// chocoTake is the take's continuation: from the cards still in the
// library, any number of lands enter tapped and the rest go to the
// graveyard. An on-disk key.
var chocoTake = RegisterTake("choose-cards/choco-seeker-of-paradise-take", TakeDef{
	Then: func(g *game.Game, res TakeFromLibraryResult) error {
		return PutFromLibraryOntoBattlefield{
			Player:   res.Player,
			Cards:    res.Rest,
			Match:    Land(),
			Max:      0,
			Optional: true,
			Tapped:   true,
			Label:    "Choco, Seeker of Paradise — put any number of land cards from among them onto the battlefield tapped",
			ThenTail: chocoRestToGraveyard,
		}.Apply(keyedContext(g, res.Player, res.Source))
	},
})

// chocoRestToGraveyard is "and the rest into your graveyard". An on-disk key.
var chocoRestToGraveyard = PutFromLibraryTail("choco-seeker-of-paradise/rest-into-graveyard",
	func(ctx *Context, res PutFromLibraryResult) error { return PutRestIntoGraveyard(ctx.Game, res) })

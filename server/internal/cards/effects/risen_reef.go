package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Risen Reef — Creature — Elemental {1}{G}{U}, 1/1:
//
//	"Whenever this creature or another Elemental you control enters,
//	 look at the top card of your library. If it's a land card, you may
//	 put it onto the battlefield tapped. If you don't put the card onto
//	 the battlefield, put it into your hand."
//
// Every Elemental is a card, and some are a land as well. The look is a
// LOOK: only the controller learns the card (the choose_cards prompt is
// withheld from every other seat), and the table finds out what it was
// only when it lands on the battlefield or in hand. The land half is
// #745's library-to-battlefield move with the printed "tapped" seeded
// onto the entry, so the land arrives tapped rather than being tapped
// after it enters, and it is a put, not a land drop (CR 305.4).
//
// "If you don't put the card onto the battlefield" covers three cases
// and all three send the card to hand: it is not a land, it is a land
// and the controller declined, or it is a land a replacement kept off
// the battlefield. That is why the hand move reads Rest rather than
// the card's type.
//
// "Another Elemental you control" reads the entering permanent's
// effective subtypes, so a changeling counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2ae71e86-4400-4a30-9077-4d57a43e7395",
		Name:         "Risen Reef",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, risenReefElementalEntered,
				"Risen Reef — look at the top card; you may put a land onto the battlefield tapped, otherwise put it into your hand",
				risenReefLook),
		},
	})
}

// risenReefElementalEntered is "this creature or another Elemental you
// control enters".
func risenReefElementalEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, false)
	return ok && (c.InstanceID == source.InstanceID || c.HasSubtype("Elemental"))
}

func risenReefLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	looked := g.LookAtTopOfLibraryForEffect(item.Controller, 1)
	return PutFromLibraryOntoBattlefield{
		Player:   item.Controller,
		Cards:    looked,
		Match:    Land(),
		Max:      1,
		Optional: true,
		Tapped:   true,
		Label:    "Risen Reef — you may put the land onto the battlefield tapped (otherwise it goes into your hand)",
		Then: func(g *game.Game, res PutFromLibraryResult) error {
			for _, id := range res.Rest {
				if err := g.BounceToHandForEffect(id); err != nil {
					return err
				}
			}
			return nil
		},
	}.Apply(ctx)
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Acclaimed Contender — Creature — Human Knight {2}{W}, 3/3:
//
//	"When this creature enters, if you control another Knight, look at the
//	 top five cards of your library. You may reveal a Knight, Aura,
//	 Equipment, or legendary artifact card from among them and put it into
//	 your hand. Put the rest on the bottom of your library in a random
//	 order."
//
// An intervening "if" (CR 603.4): another Knight is checked as the creature
// enters and again as the trigger resolves, through the layered view, so a
// changeling counts. The look is Horn of the Mark's
// (LookAtTopThenMayTakeToHand): only the controller sees the five, the card
// taken is revealed, and the rest go to the bottom in a random order.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "35df179a-c0e6-4ac1-a861-e6e9b4d1614d",
		Name:         "Acclaimed Contender",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "you control another Knight",
				Test: func(ctx *Context) bool {
					return mshB4ControlsAnotherHero(ctx.Game, ctx.Controller(), ctx.Source(), ctx.TextSubtype("Knight"))
				},
			}, On(game.EventETB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return Self(ev, source, lki, g) && mshB4ControlsAnotherHero(g, source.Controller, source.InstanceID, source.TextSubtype("Knight"))
			}, "Acclaimed Contender — look at the top five cards of your library; you may reveal a Knight, Aura, Equipment, or legendary artifact card from among them and put it into your hand",
				LookAtTopThenMayTakeToHand(5,
					Or(Subtype("Knight"), Subtype("Aura"), Subtype("Equipment"), And(Legendary(), Artifact())), 1,
					"Acclaimed Contender — you may reveal a Knight, Aura, Equipment, or legendary artifact card from among them and put it into your hand"))),
		},
	})
}

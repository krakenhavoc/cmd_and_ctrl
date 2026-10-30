package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ingenious Smith — {1}{W} Creature — Human Artificer 1/1 (Marvel Super
// Heroes Commander, #4455):
//
//	"When this creature enters, look at the top four cards of your library.
//	 You may reveal an artifact card from among them and put it into your
//	 hand. Put the rest on the bottom of your library in a random order.
//	 Whenever one or more artifacts you control enter, put a +1/+1 counter on
//	 this creature. This ability triggers only once each turn."
//
// The entry trigger is Bucky Barnes' LookAtTopThenMayTakeToHand. The second
// ability is Moon Girl and Devil Dinosaur's: a trigger that fires for the
// first artifact of the turn, so a batch of artifacts entering together is
// one trigger.
//
// No simplification.
func init() {
	const lookLabel = "Ingenious Smith — look at the top four cards; you may reveal an artifact card and put it into your hand"
	const counterLabel = "Ingenious Smith — put a +1/+1 counter on this creature"
	Register(Spec{
		OracleID:     "f3988b96-9e00-445c-b3e1-6b2df9b13dac",
		Name:         "Ingenious Smith",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters(lookLabel,
				LookAtTopThenMayTakeToHand(4, Artifact(), 1, "Ingenious Smith — reveal an artifact card and put it into your hand")),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsArtifact() && !b11TriggeredThisTurn(g, source.InstanceID, counterLabel)
			}, counterLabel, Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

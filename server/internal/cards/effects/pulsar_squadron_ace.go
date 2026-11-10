package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pulsar Squadron Ace — Creature — Human Pilot {1}{W}, 1/2 (Edge of
// Eternities, #4461):
//
//	"When this creature enters, look at the top five cards of your library.
//	 You may reveal a Spacecraft card from among them and put it into your
//	 hand. Put the rest on the bottom of your library in a random order. If
//	 you didn't put a card into your hand this way, put a +1/+1 counter on
//	 this creature."
//
// Monumental Henge's look-reveal-and-take over the top five with the
// rest on the bottom in a random order, narrowed to Spacecraft cards. The
// "if you didn't" rider is the take's continuation, run after the rest has
// been placed and told what was taken: with no card taken (a declined
// reveal, or no Spacecraft among the five) the Ace gets a +1/+1 counter,
// while it is still on the battlefield. The prompt holds the stack, so the
// Ace cannot have left and come back in between.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8019e879-cfea-40a2-ac0d-86b47ec02b48",
		Name:         "Pulsar Squadron Ace",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Pulsar Squadron Ace — look at the top five cards of your library for a Spacecraft card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 5),
						Match:    HasSubtype("Spacecraft"),
						Max:      1,
						Optional: true,
						Label:    "Pulsar Squadron Ace — you may reveal a Spacecraft card and put it into your hand",
						Key:      pulsarSquadronAceTake,
					}.Apply(ctx)
				}),
		},
	})
}

// pulsarSquadronAceTake is the take: the rest on the bottom in a random
// order, then the counter if no card went to the hand. An on-disk key.
var pulsarSquadronAceTake = RegisterTake("choose-cards/pulsar-squadron-ace-take",
	TakeDef{Reveal: true, Then: func(g *game.Game, res TakeFromLibraryResult) error {
		if err := TakeRestOnBottomInRandomOrder(g, res); err != nil {
			return err
		}
		if len(res.Taken) > 0 || !onBattlefield(g, res.Source) {
			return nil
		}
		return g.AddCounterForEffect(res.Source, game.CounterPlusOne, 1)
	}})

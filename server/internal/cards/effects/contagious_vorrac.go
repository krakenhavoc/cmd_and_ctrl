package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Contagious Vorrac — Creature — Phyrexian Boar Beast {2}{G}, 3/3 (Phyrexia:
// All Will Be One, S82):
//
//	"When this creature enters, look at the top four cards of your library.
//	 You may reveal a land card from among them and put it into your hand.
//	 Put the rest on the bottom of your library in a random order. If you
//	 didn't put a card into your hand this way, proliferate."
//
// Pulsar Squadron Ace's look-reveal-and-take, over four cards and lands. The
// "if you didn't" rider is the take's continuation, run after the rest is on
// the bottom: with no card taken (a declined reveal, or no land among the
// four) the controller proliferates.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1b6da140-0d91-4fae-a3ee-33e6808acf34",
		Name:         "Contagious Vorrac",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Contagious Vorrac — look at the top four cards of your library for a land card",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 4),
						Match:    Land(),
						Max:      1,
						Optional: true,
						Label:    "Contagious Vorrac — you may reveal a land card and put it into your hand",
						Key:      contagiousVorracTake,
					}.Apply(ctx)
				}),
		},
	})
}

// contagiousVorracTake is the take: the rest on the bottom in a random
// order, then a proliferate if no card went to the hand. An on-disk key.
var contagiousVorracTake = RegisterTake("choose-cards/contagious-vorrac-take",
	TakeDef{Reveal: true, Then: func(g *game.Game, res TakeFromLibraryResult) error {
		if err := TakeRestOnBottomInRandomOrder(g, res); err != nil {
			return err
		}
		if len(res.Taken) > 0 {
			return nil
		}
		return g.ProliferateChoosingKeyedForEffect(res.Player, res.Source, game.KeyedTail{})
	}})

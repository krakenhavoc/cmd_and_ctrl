package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Connecting the Dots — Enchantment {1}{R} (Murders at Karlov Manor,
// #4469):
//
//	"Whenever a creature you control attacks, exile the top card of your
//	 library face down. (You can't look at it.)
//	 {1}{R}, Discard your hand, Sacrifice this enchantment: Put all cards
//	 exiled with this enchantment into their owners' hands."
//
// Bag of Holding's machine with a different feed. The attack trigger
// fires once per attacking creature and exiles the top card face down
// so that nobody, its owner included, may look at it
// (ExileTopFaceDownForEffect, CR 406.3). The record of what was exiled
// with this enchantment is b27ExiledWith over the trigger's label, so
// the trigger and the sacrifice ability agree by construction, and a
// card that has since left exile some other way is not pulled back.
// Discarding the hand and sacrificing the enchantment are costs (CR
// 601.2h), paid before the ability resolves, so the returning cards
// are not discarded with the hand. Each goes to its owner's hand and
// turns face up as it leaves exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a29e2c70-8909-465b-9cec-1a7037250313",
		Name:         "Connecting the Dots",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller)
			}, connectingTheDotsExileLabel, func(g *game.Game, item *game.StackItem) error {
				_, err := g.ExileTopFaceDownForEffect(item.Controller, 1)
				return err
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}{R}, Discard your hand, Sacrifice this enchantment: Put all cards exiled with this enchantment into their owners' hands.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}{R}"), DiscardYourHand(), SacrificeThis()),
			Effect:  returnCardsExiledWithToOwnersHands(connectingTheDotsExileLabel),
		}},
	})
}

// connectingTheDotsExileLabel is the attack trigger's stack label, and
// the key b27ExiledWith reads "exiled with this enchantment" under.
const connectingTheDotsExileLabel = "Connecting the Dots — exile the top card of your library face down"

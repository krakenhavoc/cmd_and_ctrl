package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nico Minoru, Runaway — Legendary Creature — Human Warlock Hero {3}{R},
// 2/4 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever you cast a spell from anywhere other than your hand, Nico
//	 Minoru deals 2 damage to each opponent.
//	 {2}{R}, {T}, Discard a card: Exile cards from the top of your library
//	 until you exile a nonland card. You may cast that card without paying
//	 its mana cost."
//
// The trigger is Vega, the Watcher's "cast from anywhere other than your
// hand" (the cast event's old zone) with DamageEachOpponent. The activated
// ability is The Infamous Cruelclaw's exile loop (exileUntilNonlandGrantingCast)
// with the price {0} in place of a discard: a flash-speed, cast-only
// permission over the nonland card exiled, made DURING the resolution
// (CR 608.2g), whose lapse leaves an uncast card in exile and an {X} in
// its cost at 0 (CR 107.3b). The lands exiled on the way stay in exile.
// The discard is a cost, so it is paid before the ability is on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5303b3f6-970c-47d8-977e-2d6679206ba5",
		Name:         "Nico Minoru, Runaway",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b25CastFromNotHand(ev, source.Controller)
			}, "Nico Minoru, Runaway — 2 damage to each opponent", Do(DamageEachOpponent{N: 2})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}, {T}, Discard a card: Exile cards from the top of your library until you exile a nonland card. You may cast that card without paying its mana cost.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{R}"), TapCost(), DiscardACard()),
			Effect:  nicoMinoruExile,
		}},
	})
}

func nicoMinoruExile(g *game.Game, item *game.StackItem) error {
	return exileUntilNonlandCastFree(NewContext(g, item), "Nico Minoru, Runaway")
}

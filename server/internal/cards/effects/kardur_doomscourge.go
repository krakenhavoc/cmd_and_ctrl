package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kardur, Doomscourge — Legendary Creature — Demon Berserker, {2}{B}{R}, 4/3:
//
//	"When Kardur enters, until your next turn, creatures your
//	 opponents control attack each combat if able and attack a
//	 player other than you if able.
//	 Whenever an attacking creature dies, each opponent loses 1 life
//	 and you gain 1 life."
//
// #1599: the ETB half is The Akroan War's chapter II shape —
// OpponentsCreaturesAttackIfAble with OtherThanYou set (CR 701.15b's
// second requirement, "a player other than you"), for
// DurationUntilYourNextTurn. It is a requirement, not a
// characteristic change, so CR 611.2c does not lock the affected set:
// a creature an opponent casts during that turn cycle has to attack
// too, and the engine judges it with every other requirement at the
// table.
//
// #1661: the drain is diedWhileAttacking — the dying creature's
// attack as it last existed, carried on its leaves-the-battlefield
// event (CR 603.10a), because the exit has already cleared it from
// the card. ANY attacking creature: an opponent's creature Kardur
// forced in, your own, and Kardur himself if he dies attacking (the
// LTB harvest reads his own event the same way). A blocker dying is
// not an attacking creature dying, and nor is a creature that was
// removed from combat before it died.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bc14356c-3a1a-47af-9a6e-2b449de0331f",
		Name:         "Kardur, Doomscourge",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Kardur, Doomscourge — creatures your opponents control attack each combat if able and attack a player other than you if able",
				kardurDoomscourgeRequirement),
			On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := diedWhileAttacking(ev, g)
				return ok
			}, "Kardur, Doomscourge — each opponent loses 1 life, you gain 1 life", drainEachOpponent),
		},
	})
}

func kardurDoomscourgeRequirement(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	return OpponentsCreaturesAttackIfAble{
		OtherThanYou: true,
		Duration:     DurationUntilYourNextTurn(ctx, item.Controller),
		Label:        "Kardur, Doomscourge — creatures your opponents control attack each combat if able and attack a player other than you if able",
	}.Apply(ctx)
}

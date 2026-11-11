package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hapatra, Vizier of Poisons — Legendary Creature — Human Cleric {B}{G},
// 2/2 (Lorwyn Eclipsed Commander, #5368):
//
//	"Whenever Hapatra deals combat damage to a player, you may put a -1/-1
//	 counter on target creature.
//	 Whenever you put one or more -1/-1 counters on a creature, create a
//	 1/1 green Snake creature token with deathtouch."
//
// The first trigger has a target chosen as it goes on the stack and a "you
// may" asked as it resolves (Optional): a "no", or a target gone by then
// (CR 608.2b), puts nothing. The second is "you put", the placement's own
// actor (Nest of Scarabs' b33YouPutMinusCountersOnACreature): Hapatra's own
// counter makes a Snake, an opponent's counters on their own creatures make
// none for this controller, and several counters in one event are one Snake.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8ddea1c-8d80-49c1-a5b4-630d5e51d66e",
		Name:         "Hapatra, Vizier of Poisons",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(Targeting(
				On(game.EventDealDamage, ThisDealtCombatDamageToAPlayer,
					"Hapatra — you may put a -1/-1 counter on target creature", putACounterOnTheTarget(game.CounterMinusOne)),
				Harmful(TargetCreature("target creature"))),
				"Hapatra — put a -1/-1 counter on the target creature?"),
			On(game.EventCounterPlaced, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b33YouPutMinusCountersOnACreature(ev, source, g)
			}, "Hapatra — create a 1/1 green Snake creature token with deathtouch",
				Do(CreateToken{Template: TokenCard("1/1 green Snake with deathtouch"), N: 1})),
		},
	})
}

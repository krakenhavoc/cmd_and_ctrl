package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Beast, Erudite Aerialist — Legendary Creature — Mutant Scientist Hero
// {3}{G/U}, 3/3 (Marvel Super Heroes, #4455):
//
//	"As long as you've put one or more +1/+1 counters on Beast this turn,
//	 he has flying.
//	 Whenever Beast deals combat damage to a player, draw a card."
//
// The flying is a self-only layer-6 grant whose condition reads this turn's
// event log for a +1/+1 placement on Beast by his controller
// (mshB2YouPutPlusOneOnThisTurn). Counters an opponent puts on him, or that
// he enters with, are not "you put" and do not count; the keyword goes away
// at the next turn. The draw is the ordinary combat-damage trigger,
// mandatory.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "191ba5f0-45f4-4b7f-910d-4bd4d2d28b33",
		Name:         "Beast, Erudite Aerialist",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && mshB2YouPutPlusOneOnThisTurn(g, source)
			}, "flying"),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Beast, Erudite Aerialist — draw a card", Do(DrawCards{N: 1})),
		},
	})
}

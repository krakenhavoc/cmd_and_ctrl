package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tenacious Hunter — Creature — Crocodile {2}{G}{G}, 4/4 (Hour of
// Devastation, #5534):
//
//	"As long as a creature has a -1/-1 counter on it, this creature has
//	 vigilance and deathtouch."
//
// Two layer-6 keyword grants to this creature, both on while any creature on
// the battlefield, whoever controls it and the Hunter itself included, has a
// -1/-1 counter. They are read live through the layers, so the keywords come
// and go as the counters do.
//
// No simplification.
func init() {
	whileACreatureHasOne := func(target *game.Card, g *game.Game, source *game.Card) bool {
		if target.InstanceID != source.InstanceID {
			return false
		}
		for _, c := range g.BattlefieldCardsForEffect() {
			if c.IsCreature() && c.Counters[game.CounterMinusOne] > 0 {
				return true
			}
		}
		return false
	}
	effects.Register(effects.Spec{
		OracleID:     "8c7436fa-4802-405b-ba45-4e14f14a3d8e",
		Name:         "Tenacious Hunter",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.KeywordGrant(whileACreatureHasOne, "vigilance"),
			effects.KeywordGrant(whileACreatureHasOne, "deathtouch"),
		},
	})
}

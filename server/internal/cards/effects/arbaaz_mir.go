package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arbaaz Mir — Legendary Creature — Human Assassin {R}{W}, 2/2
// (EDHREC rank 3631):
//
//	"Whenever Arbaaz Mir or another nontoken historic permanent you
//	 control enters, Arbaaz Mir deals 1 damage to each opponent and
//	 you gain 1 life. (Artifacts, legendaries, and Sagas are
//	 historic.)"
//
// The Assassin's Creed historic pinger. One printed ability with two
// conditions on one declaration: Arbaaz's own entry, or another
// nontoken permanent entering under his controller's control that is
// historic — an artifact, a legendary, a Saga — read post-layer, so
// a Treasure does not count (a token) and an animated artifact does.
// The damage is dealt by Arbaaz (colourless noncombat, one event per
// opponent), then the life is gained; both happen even when Arbaaz
// has left in response.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a383ef16-1af8-4b3a-956c-c10a93768617",
		Name:         "Arbaaz Mir",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b34SelfOrAnotherNontokenHistoricYouControlEntered(ev, source, g)
			}, "Arbaaz Mir — 1 damage to each opponent; gain 1 life", b34DamageEachOpponentAndGainLife(1, 1)),
		},
	})
}

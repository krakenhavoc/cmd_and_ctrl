package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// HYDRA Infiltration — Enchantment {3}{B} (Marvel Super Heroes, #4455):
//
//	"When this enchantment enters, target opponent discards two cards.
//	 Whenever a creature you control attacks alone, target opponent loses
//	 1 life and you gain 1 life."
//
// The enter trigger is Thunderscape Battlemage's "target player
// discards N", aimed at an opponent: the opponent picks the cards
// (CR 701.9a). The attack trigger is Agents of S.H.I.E.L.D.'s "alone"
// (game.AttackedAlone, read when the trigger is harvested) with Down,
// Down to Goblin-town's one-point drain, a target chosen when the
// trigger goes on the stack (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "27692079-a517-4d38-b452-7409980b5912",
		Name:         "HYDRA Infiltration",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("HYDRA Infiltration — target opponent discards two cards", targetPlayerDiscards(2)),
				TargetPlayer("target opponent", Opponent())),
			Targeting(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return attackDeclaredByYou(ev, source.Controller) && game.AttackedAlone(g)
			}, "HYDRA Infiltration — target opponent loses 1 life and you gain 1 life", drainTargetOpponentOne),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

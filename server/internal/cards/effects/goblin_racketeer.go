package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Racketeer — Creature — Goblin Rogue, {3}{R}, 4/2:
//
//	"Whenever this creature attacks, you may goad target creature
//	defending player controls. (Until your next turn, that creature
//	attacks each combat if able and attacks a player other than you
//	if able.)"
//
// #1599: GoadTarget (goad.go) over
// TargetCreatureDefendingPlayerControls (defending_player_targets.go)
// — the defending player's own back row is the whole reason to swing
// with this rather than something bigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9336a62c-f2f9-45a8-bf69-86060aa0ce59",
		Name:         "Goblin Racketeer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(game.TriggeredAbility{
				Watches:     []game.EventKind{game.EventAttack},
				AppliesTo:   ThisAttacked,
				TargetsFrom: TargetCreatureDefendingPlayerControls,
				Key:         "Goblin Racketeer — goad target creature defending player controls",
				Effect:      GoadTarget,
			}, "Goblin Racketeer — goad target creature defending player controls?"),
		},
	})
}

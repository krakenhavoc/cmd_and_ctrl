package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Coveted Peacock — Creature — Bird, {3}{U}{U}, 3/4:
//
//	"Flying
//	 Whenever this creature attacks, you may goad target creature
//	 defending player controls. (Until your next turn, that creature
//	 attacks each combat if able and attacks a player other than you
//	 if able.)"
//
// #1599: flying rides PrintedKeywords; the attack trigger is Goblin
// Racketeer's exact shape, GoadTarget (goad.go) over
// TargetCreatureDefendingPlayerControls
// (defending_player_targets.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e6ad5e92-c1ab-4c91-95ca-1af295e71b23",
		Name:            "Coveted Peacock",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(game.TriggeredAbility{
				Watches:     []game.EventKind{game.EventAttack},
				AppliesTo:   ThisAttacked,
				TargetsFrom: TargetCreatureDefendingPlayerControls,
				Key:         "Coveted Peacock — goad target creature defending player controls",
				Effect:      GoadTarget,
			}, "Coveted Peacock — goad target creature defending player controls?"),
		},
	})
}

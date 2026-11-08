package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kitesail Corsair — Creature — Human Pirate {1}{U}, 2/1 (Foundations,
// #4465):
//
//	"This creature has flying as long as it's attacking."
//
// A layer 6 grant on this permanent while it is an attacker
// (Card.AttackingTarget), declared DependsOnAttackingStatus as The
// Masamune's attacking grant is, so the layer cache is rebuilt as
// attackers are declared and when combat ends. It has flying when
// blockers are declared (CR 509.1b) and never while blocking.
//
// No simplification.
func init() {
	flying := KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.AttackingTarget != uuid.Nil
	}, "flying")
	flying.Label = "This creature has flying as long as it's attacking."
	flying.DependsOnAttackingStatus = true
	Register(Spec{
		OracleID:     "6103f294-f920-4307-af80-d44cc71167fd",
		Name:         "Kitesail Corsair",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{flying},
	})
}

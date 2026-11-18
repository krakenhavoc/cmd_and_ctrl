package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blood Petal Celebrant — Creature — Vampire {1}{R}, 2/1 (Innistrad: Crimson Vow):
//
//	"This creature has first strike as long as it's attacking.
//	 When this creature dies, create a Blood token."
func init() {
	firstStrike := KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.AttackingTarget != uuid.Nil
	}, "first strike")
	firstStrike.Label = "This creature has first strike as long as it's attacking."
	firstStrike.DependsOnAttackingStatus = true
	Register(Spec{
		OracleID:     "219cf182-0942-4569-91e9-9bc9bb1d06d0",
		Name:         "Blood Petal Celebrant",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{firstStrike},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Blood Petal Celebrant — create a Blood token", Do(s82Vow03CreateBlood(1))),
		},
	})
}

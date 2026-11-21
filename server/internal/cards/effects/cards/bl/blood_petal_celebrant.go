package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blood Petal Celebrant — Creature — Vampire {1}{R}, 2/1 (Innistrad: Crimson Vow):
//
//	"This creature has first strike as long as it's attacking.
//	 When this creature dies, create a Blood token."
func init() {
	firstStrike := effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.AttackingTarget != uuid.Nil
	}, "first strike")
	firstStrike.Label = "This creature has first strike as long as it's attacking."
	firstStrike.DependsOnAttackingStatus = true
	effects.Register(effects.Spec{
		OracleID:     "219cf182-0942-4569-91e9-9bc9bb1d06d0",
		Name:         "Blood Petal Celebrant",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{firstStrike},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Blood Petal Celebrant — create a Blood token", effects.Do(effects.S82Vow03CreateBlood(1))),
		},
	})
}

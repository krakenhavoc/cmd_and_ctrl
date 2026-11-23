package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thorned Moloch — Creature — Lizard {2}{R}, 2/2 (Hour of Devastation,
// #5534):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 This creature has first strike as long as it's attacking."
//
// Prowess is a printed keyword (CR 702.108a). The first strike is Blood
// Petal Celebrant's self grant, read off the creature's own attack, which
// ends when it is removed from combat.
//
// No simplification.
func init() {
	firstStrike := effects.KeywordGrant(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && target.AttackingTarget != uuid.Nil
	}, "first strike")
	firstStrike.Label = "This creature has first strike as long as it's attacking."
	firstStrike.DependsOnAttackingStatus = true
	effects.Register(effects.Spec{
		OracleID:        "d2a49109-6db8-4f08-828c-da7a456a78a7",
		Name:            "Thorned Moloch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"prowess"},
		Static:          []game.StaticAbility{firstStrike},
	})
}

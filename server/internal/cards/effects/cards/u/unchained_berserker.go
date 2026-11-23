package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unchained Berserker — Creature — Human Berserker {1}{R}, 1/1 (Core Set 2020, #5518):
//
//	"Protection from white (This creature can't be blocked, targeted, dealt
//	 damage, enchanted, or equipped by anything white.)
//	 This creature gets +2/+0 as long as it's attacking."
//
// The pump is a layer 7c modification on itself while it is an attacker
// (Card.AttackingTarget), declared DependsOnAttackingStatus as Kitesail
// Corsair's grant is, so the layers are rebuilt as attackers are declared
// and as combat ends.
//
// No simplification.
func init() {
	pump := game.StaticAbility{
		Label:    "This creature gets +2/+0 as long as it's attacking.",
		Layer:    game.Layer7PT,
		SubLayer: game.SubLayer7C_Modify,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID && target.AttackingTarget != uuid.Nil
		},
		Apply:                    effects.PumpBy(2, 0),
		DependsOnAttackingStatus: true,
	}
	effects.Register(effects.Spec{
		OracleID:        "32e7befb-c083-4e17-aaeb-7247f317bbcf",
		Name:            "Unchained Berserker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from white"},
		Static:          []game.StaticAbility{pump},
	})
}

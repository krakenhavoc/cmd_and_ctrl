package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kor Blademaster — Creature — Kor Warrior {1}{W}, 1/1 (Zendikar Rising,
// #5508):
//
//	"Double strike
//	 Equipped Warriors you control have double strike."
//
// A layer 6 grant over the Warriors its controller controls that have an
// Equipment attached (CR 301.5, read live, so it follows an attach or an
// unattach). The Blademaster has double strike itself, and grants it to
// itself too when it is equipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "194260af-aba5-4d18-a69b-eedb97ed4a4b",
		Name:            "Kor Blademaster",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"double strike"},
		Static: []game.StaticAbility{effects.KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
			return target.IsCreature() && target.Controller == source.Controller &&
				target.HasSubtype(source.TextSubtype("Warrior")) && effects.IsEquippedCreature(g, *target)
		}, "double strike")},
	})
}

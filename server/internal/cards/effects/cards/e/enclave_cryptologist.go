package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enclave Cryptologist — Creature — Merfolk Wizard {U}, 0/1 (Rise of the
// Eldrazi, #5581):
//
//	"Level up {1}{U} ({1}{U}: Put a level counter on this. Level up only as
//	 a sorcery.)
//	 LEVEL 1-2
//	 0/1
//	 {T}: Draw a card, then discard a card.
//	 LEVEL 3+
//	 0/1
//	 {T}: Draw a card."
//
// A leveler (leveler.go, CR 702.87a, 711.2a, 711.2b). Each band's tap
// ability is a bundle the Cryptologist grants itself behind that band's
// level check.
//
// No simplification.
const (
	EnclaveCryptologistLoot = "enclave-cryptologist/loot"
	EnclaveCryptologistDraw = "enclave-cryptologist/draw"
)

func init() {
	effects.Register(effects.Spec{
		OracleID:     "bc0fc535-357a-4d83-8d64-34337222cc4b",
		Name:         "Enclave Cryptologist",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{
			{
				Key: EnclaveCryptologistLoot,
				Activated: []effects.ActivatedAbility{{
					Label:   "{T}: Draw a card, then discard a card.",
					Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1, Discards: 1},
					Cost:    effects.TapCost(),
					Effect:  func(g *game.Game, item *game.StackItem) error { return effects.LootOne(g, item, 1) },
				}},
				Text: "{T}: Draw a card, then discard a card.",
			},
			{
				Key: EnclaveCryptologistDraw,
				Activated: []effects.ActivatedAbility{{
					Label:   "{T}: Draw a card.",
					Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
					Cost:    effects.TapCost(),
					Effect:  effects.Do(effects.DrawCards{N: 1}),
				}},
				Text: "{T}: Draw a card.",
			},
		},
		Activated: []effects.ActivatedAbility{effects.LevelUpKeyword("{1}{U}")},
		Static: effects.LevelBands(
			effects.LevelBand{From: 1, To: 2, Power: 0, Toughness: 1, Statics: []game.StaticAbility{
				effects.Labelled(effects.GrantAbilities(effects.SelfOnly, EnclaveCryptologistLoot), `This creature has "{T}: Draw a card, then discard a card."`),
			}},
			effects.LevelBand{From: 3, Power: 0, Toughness: 1, Statics: []game.StaticAbility{
				effects.Labelled(effects.GrantAbilities(effects.SelfOnly, EnclaveCryptologistDraw), `This creature has "{T}: Draw a card."`),
			}},
		),
	})
}

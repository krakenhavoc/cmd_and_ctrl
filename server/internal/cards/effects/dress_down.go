package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dress Down — Enchantment {1}{U} (Modern Horizons 3):
//
//	"Flash
//	 When this enchantment enters, draw a card.
//	 Creatures lose all abilities.
//	 At the beginning of the end step, sacrifice this enchantment."
//
// The loss is Spark Rupture's layer 6 removal (RemovesAbilities, so a
// creature's catalogued triggered, activated, mana and static abilities stop
// answering and its keyword badges go), over every creature of every
// player; an ability granted by a later effect (CR 613.6) survives. The
// sacrifice is Ball Lightning's: at the beginning of every end step, whose
// turn it is, taking the enchantment out if it is still on the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c1cbbda0-e02b-4b9c-8669-ea87ca629520",
		Name:            "Dress Down",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Draws: 1},
		Static: []game.StaticAbility{{
			Label:            "Creatures lose all abilities.",
			Layer:            game.Layer6Ability,
			RemovesAbilities: true,
			AppliesTo:        func(target *game.Card, _ *game.Game, _ *game.Card) bool { return target.IsCreature() },
			Apply:            func(*game.Characteristic, *game.Card, *game.Game, *game.Card) {},
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dress Down — draw a card", Do(DrawCards{N: 1})),
			On(game.EventBeginEndStep, AnyPlayer, "Dress Down — sacrifice this enchantment", SacrificeThisIfStillOnBattlefield),
		},
	})
}

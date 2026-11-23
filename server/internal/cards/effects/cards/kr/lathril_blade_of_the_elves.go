package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lathril, Blade of the Elves — Legendary Creature — Elf Noble {2}{B}{G},
// 2/3 (Foundations, #4465):
//
//	"Menace (This creature can't be blocked except by two or more
//	 creatures.)
//	 Whenever Lathril deals combat damage to a player, create that many
//	 1/1 green Elf Warrior creature tokens.
//	 {T}, Tap ten untapped Elves you control: Each opponent loses 10 life
//	 and you gain 10 life."
//
// The trigger is Ant-Man's "that many", read off the damage event, with
// the Elf Warrior token. The cost is Harmonized Trio's {T} plus other
// untapped creatures (Lathril is already tapped by {T}, so she is not one
// of the ten), counted as High Perfect Morcant counts Elves; those ten may
// have arrived this turn, and the cost does not target. The drain is a
// footprinted life loss for each opponent, then the gain, once.
//
// No simplification.
func init() {
	const tapTen = "ten untapped Elves you control"
	effects.Register(effects.Spec{
		OracleID:        "9b568c34-ba0a-44a7-a0ee-bee3a37b3396",
		Name:            "Lathril, Blade of the Elves",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Lathril, Blade of the Elves — create that many 1/1 green Elf Warrior tokens",
				effects.Do(effects.CreateTokensForEach{
					Template: effects.TokenCard("1/1 green Elf Warrior"),
					Name:     "the combat damage dealt",
					Count:    func(ctx *effects.Context) int { return ctx.Trigger().Event.Amount },
				})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Tap ten untapped Elves you control: Each opponent loses 10 life and you gain 10 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost: effects.Plus(effects.TapCost(), game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:         10,
				Filter:        effects.TargetPermanent(tapTen, effects.Creature(), effects.OfCreatureType("Elf")),
				ExcludeSource: true,
				Label:         tapTen,
			}}),
			Effect: effects.Do(effects.EachOpponentLosesLife{N: 10}, effects.GainLife{Amount: 10}),
		}},
	})
}

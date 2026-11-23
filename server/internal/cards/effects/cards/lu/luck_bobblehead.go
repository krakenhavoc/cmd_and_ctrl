package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Luck Bobblehead — Artifact — Bobblehead {3} (Fallout, #5948):
//
//	"{T}: Add one mana of any color.
//	 {1}, {T}: Roll X six-sided dice, where X is the number of Bobbleheads
//	 you control. Create a tapped Treasure token for each even result. If
//	 you rolled 6 exactly seven times, you win the game."
//
// X is counted as the ability resolves (CR 608.2h); with none, nothing is
// rolled. The rest of the clause is the tail told the kept results
// (Clown Car's shape): a tapped Treasure for each even die, and a win
// (WinTheGame, ADR 0057) when exactly seven of them are sixes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d870a93a-8399-4004-9b45-2b49b001f165",
		Name:         "Luck Bobblehead",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Roll X six-sided dice, where X is the number of Bobbleheads you control. Create a tapped Treasure token for each even result. If you rolled 6 exactly seven times, you win the game.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				x := effects.PipBobbleheadsYouControl(ctx.Game, ctx.Controller())
				if x <= 0 {
					return nil
				}
				return effects.RollDice{Sides: 6, N: x, Then: luckBobbleheadResults}.Apply(ctx)
			})),
		}},
	})
}

// luckBobbleheadResults is the rest of the ability, told the kept d6
// results. An on-disk key: never renamed, never reused.
var luckBobbleheadResults = effects.DiceTail("dice/luck-bobblehead-results", LuckBobbleheadApply)

// LuckBobbleheadApply is the tail's body: a tapped Treasure for each even
// result, and the win when exactly seven of the dice came up 6.
func LuckBobbleheadApply(ctx *effects.Context, rolls []int) error {
	even, sixes := 0, 0
	for _, r := range rolls {
		if r%2 == 0 {
			even++
		}
		if r == 6 {
			sixes++
		}
	}
	if even > 0 {
		if err := (effects.CreateTokenAdvanced{Spec: effects.Token(effects.TreasureToken()).EntersTapped(), N: even}).Apply(ctx); err != nil {
			return err
		}
	}
	if sixes == 7 {
		return effects.WinTheGame{Player: ctx.Controller()}.Apply(ctx)
	}
	return nil
}

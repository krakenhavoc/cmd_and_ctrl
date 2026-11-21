package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d870a93a-8399-4004-9b45-2b49b001f165",
		Name:         "Luck Bobblehead",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Roll X six-sided dice, where X is the number of Bobbleheads you control. Create a tapped Treasure token for each even result. If you rolled 6 exactly seven times, you win the game.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect: Do(Instruction(func(ctx *Context) error {
				x := pipBobbleheadsYouControl(ctx.Game, ctx.Controller())
				if x <= 0 {
					return nil
				}
				return RollDice{Sides: 6, N: x, Then: luckBobbleheadResults}.Apply(ctx)
			})),
		}},
	})
}

// luckBobbleheadResults is the rest of the ability, told the kept d6
// results. An on-disk key: never renamed, never reused.
var luckBobbleheadResults = DiceTail("dice/luck-bobblehead-results", luckBobbleheadApply)

// luckBobbleheadApply is the tail's body: a tapped Treasure for each even
// result, and the win when exactly seven of the dice came up 6.
func luckBobbleheadApply(ctx *Context, rolls []int) error {
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
		if err := (CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: even}).Apply(ctx); err != nil {
			return err
		}
	}
	if sixes == 7 {
		return WinTheGame{Player: ctx.Controller()}.Apply(ctx)
	}
	return nil
}

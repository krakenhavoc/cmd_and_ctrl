package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bucknard's Everfull Purse — Artifact {2} (Adventures in the Forgotten Realms):
//
//	"{1}, {T}: Roll a d4 and create a number of Treasure tokens equal to
//	 the result. The player to your right gains control of this artifact."
//
// The Treasures come first, then the hand-off, both after the roll (a
// replaced roll waits for its chooser first, #4307). "The player to your
// right" is the nearest player still in the game walking right from the
// ability's controller (game.NearestOpponentForEffect); the control change
// has no duration (CR 611.2a) and does nothing if the Purse has left the
// battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "f70b9eee-3874-4391-8df8-e10bbd19307e", Name: "Bucknard's Everfull Purse", Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{1}, {T}: Roll a d4 and create that many Treasure tokens. The player to your right gains control of this artifact.",
			Cost:  effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			// Treasures answer nothing; the hand-off is a drawback.
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.RollDice{Sides: 4, N: 1, Then: bucknardsPurseRolled}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}

var bucknardsPurseRolled = effects.DiceTail("dice/bucknards-everfull-purse", func(ctx *effects.Context, rolls []int) error {
	if err := effects.TreasureTokens(ctx.Game, ctx.Controller(), rolls[0]); err != nil {
		return err
	}
	purse := ctx.Source()
	right := ctx.Game.NearestOpponentForEffect(ctx.Controller(), game.AttackDirectionRight)
	if right == uuid.Nil || !effects.OnBattlefield(ctx.Game, purse) {
		return nil
	}
	return effects.GainControl{
		Target:     purse,
		Controller: right,
		Duration:   game.IndefiniteDuration(),
		Label:      "Bucknard's Everfull Purse — the player to your right gains control",
	}.Apply(ctx)
})

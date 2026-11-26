package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Horribly Awry — Instant {1}{U} (Battle for Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Counter target creature spell with mana value 4 or less. If that spell is
//	 countered this way, exile it instead of putting it into its owner's
//	 graveyard."
//
// Reject's target and Defabricate's exile destination: the spell goes to
// exile only if it really is countered, and a spell that can't be countered
// stays on the stack. Mana value 4 or less is read off the spell on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9cac23a1-a0b3-490c-aea2-dc2928f6dc9e",
		Name:            "Horribly Awry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Purpose:         game.Purpose{Answers: game.AnswerRemove},
		Targets:         effects.TargetSpell("target creature spell with mana value 4 or less", effects.And(effects.Creature(), effects.ManaValueLE(4))),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			return effects.CounterTarget{StackID: targets[0].ID, Dest: game.ZoneRef{Kind: game.ZoneExile}}.Apply(ctx)
		},
	})
}

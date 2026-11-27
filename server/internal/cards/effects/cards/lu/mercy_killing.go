package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mercy Killing — Instant {2}{G/W} (Shadowmoor, #5591):
//
//	"Target creature's controller sacrifices it, then creates X 1/1 green
//	 and white Elf Warrior creature tokens, where X is that creature's
//	 power."
//
// The target's controller and its power are read as the spell resolves,
// before the sacrifice, through the layers (CR 608.2h, so an anthem or a
// counter counts). The controller of the creature, not the caster, does the
// sacrificing and gets the tokens; a sacrifice is not a destroy, so
// indestructible and protection don't save it. The tokens follow in the
// sacrifice's own tail, so they come after it has settled. A creature that
// is no longer a legal target does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4d80bac5-ce62-4c5c-873a-ce81d2697842",
		Name:         "Mercy Killing",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			c, found := ctx.Game.LookupCardForEffect(t.ID)
			if !found {
				return nil
			}
			info, live := ctx.Game.PermanentForEffect(game.ObjectRef{ID: t.ID, Epoch: c.ObjectEpoch})
			if !live {
				return nil
			}
			power := info.Power
			if power < 0 {
				power = 0
			}
			then := ctx.ItemTail(mercyKillingTokens)
			then.Player = info.Controller
			then.N = power
			return effects.SacrificePermanent{Target: t.ID, ThenTail: then}.Apply(ctx)
		},
	})
}

// mercyKillingTokens is the "then creates X tokens" half: the player is the
// creature's controller and X its power, both fixed before the sacrifice.
// An on-disk key.
var mercyKillingTokens = effects.CardTail("mercy-killing/elf-warriors", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 || o.Tail.Player == uuid.Nil {
		return nil
	}
	return effects.CreateToken{Controller: o.Tail.Player, Template: effects.TokenCard("1/1 green and white Elf Warrior"), N: o.Tail.N}.Apply(ctx)
})

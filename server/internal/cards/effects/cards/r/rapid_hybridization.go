package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rapid Hybridization — Instant {U} (EDHREC rank 207):
//
//	"Destroy target creature. It can't be regenerated. That
//	 creature's controller creates a 3/3 green Frog Lizard creature
//	 token."
//
// Pongify's twin — the same one-mana blue answer with a different
// token name, and the same shape: destroy first, then hand the
// VICTIM the vanilla 3/3. The controller is read BEFORE the destroy,
// for the reason beast_within.go gives: afterwards the card is in a
// graveyard and its controller field is stale for a creature that
// had changed hands.
//
// "It can't be regenerated" is ENFORCED as of #667 (CR 701.19c):
// the rider rides the destroy route onto the CR 614 event and the
// regeneration built-in declines. The shield is not spent
// (CR 701.19c).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "06692cd9-ac2f-4a32-8fd1-043ba3c0fe71",
		Name:         "Rapid Hybridization",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		// #2679: what the target's controller is given back.
		Purpose: game.Purpose{Targets: game.ForTargets(effects.RemovalReturning(0, game.TargetReturn{CreatureTokens: 1, TokenPower: 3, TokenToughness: 3})), Answers: game.AnswerRemove}, // #4734: destroys a permanent
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			controller, ok := effects.ControllerOfTarget(ctx, target)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: target, CantBeRegenerated: true}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{
				Controller: controller,
				Template:   effects.TokenCard("3/3 green Frog Lizard"),
				N:          1,
			}.Apply(ctx)
		},
	})
}

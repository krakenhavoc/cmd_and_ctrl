package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burn Away — Instant {4}{R} (Conspiracy: Take the Crown, S82 older sets
// batch 270, #5540):
//
//	"Burn Away deals 6 damage to target creature. When that creature dies
//	 this turn, exile its controller's graveyard."
//
// The damage is the plain single-target kind. The second sentence is a
// delayed triggered ability the spell creates about that creature as the
// object it is now (CR 603.7a, 603.7c; WhenThatCreatureDiesThisTurn): it
// fires if that creature dies this turn, whether or not Burn Away's damage
// killed it, and a creature that left and came back is a new object. "Its
// controller" is the player who controlled it as it last existed on the
// battlefield (CR 608.2h), read from the payload's last-known record, so a
// creature stolen before it died empties the thief's graveyard, not the
// owner's. The graveyard is whatever it holds as the trigger resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65dcad6b-9a87-4313-bc71-682e1de262d0",
		Name:         "Burn Away",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 6)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DamageToFirstTargetStep(6)(item, ctx); err != nil {
				return err
			}
			return effects.WhenThatCreatureDiesThisTurn(ctx, effects.FirstLegalCardTarget(ctx),
				"Burn Away — exile that creature's controller's graveyard", burnAwayExileGraveyardBody)
		},
	})
}

// burnAwayExileGraveyardBody is the fired trigger's body. An on-disk key:
// never renamed, never reused.
var burnAwayExileGraveyardBody = game.DelayedBody("old-b270/burn-away-exile-graveyard",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		obj, ok := effects.PayloadObject(ctx)
		if !ok {
			return nil
		}
		info, ok := g.PermanentForEffect(obj)
		if !ok {
			return nil
		}
		return effects.ExileGraveyardForEffect(g, item, info.Controller)
	})

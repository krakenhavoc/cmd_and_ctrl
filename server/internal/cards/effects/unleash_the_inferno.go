package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Unleash the Inferno — Instant {1}{B}{R}{G} (Streets of New Capenna, #5491):
//
//	"Unleash the Inferno deals 7 damage to target creature or planeswalker.
//	 When it deals excess damage this way, destroy target artifact or
//	 enchantment an opponent controls with mana value less than or equal
//	 to that amount of excess damage."
//
// The damage is Hell to Pay's, widened to a planeswalker
// (s82MomB12DealDamageWithExcess): excess is what landed beyond the damage
// the creature still needed, or beyond the planeswalker's loyalty (CR 120.10),
// read after the damage so a prevention shield lowers it. Any excess raises
// the printed "when it deals excess damage this way" as a reflexive trigger
// (CR 603.12) with the excess as its amount; it chooses its own target as it
// goes on the stack, so the artifact or enchantment is picked after the
// damage, and a trigger with no legal target is simply removed (CR 603.3d).
// A target that is gone on resolution leaves the spell with nothing to do.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d305d2dd-578c-4964-a7c2-26490bb2ef2e",
		Name:         "Unleash the Inferno",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		Purpose:      ForTargets(DamageToTarget(0, 7)),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			target := FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			excess, err := s82MomB12DealDamageWithExcess(ctx, target, 7)
			if err != nil || excess <= 0 {
				return err
			}
			return ReflexiveTrigger{
				Label:  "Unleash the Inferno — destroy target artifact or enchantment an opponent controls with mana value " + strconv.Itoa(excess) + " or less",
				Body:   unleashTheInfernoDestroyBody,
				Params: game.EffectParams{Amount: excess},
			}.Apply(ctx)
		},
	})
}

// unleashTheInfernoDestroyBody is the reflexive trigger's registration. The
// clause needs the excess to build, so it rides Params.Amount (Teferi Akosa's
// shape).
var unleashTheInfernoDestroyBody = game.ReflexiveBody("unleash-the-inferno/destroy", simpleBody(destroyFirstLegalCardTarget),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetPermanent(
			"target artifact or enchantment an opponent controls with mana value "+strconv.Itoa(p.Amount)+" or less",
			Or(Artifact(), Enchantment()), OpponentControls(), ManaValueLE(p.Amount))
	})

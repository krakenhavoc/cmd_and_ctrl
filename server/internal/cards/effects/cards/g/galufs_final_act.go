package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galuf's Final Act — Instant {1}{G} (Final Fantasy, #4462):
//
//	"Until end of turn, target creature gets +1/+0 and gains "When this
//	 creature dies, put a number of +1/+1 counters equal to its power on
//	 up to one target creature.""
//
// Malakir Rebirth's duration grant of a dies trigger (ADR 0093), with
// Feral Ghoul's last-known power (CR 603.10a) on the trigger: the power
// the creature had as it last existed on the battlefield, the +1/+0
// included. The granted trigger targets up to one creature as it goes on
// the stack (CR 603.3d), so it may pick the dying creature's ally or an
// opponent's creature, or nothing; a target that has gone gets nothing
// (CR 608.2b). A creature with no power puts no counters. The pump and the
// grant are locked in as the spell resolves (CR 611.2c) and a target that
// is no longer legal gets neither.
//
// No simplification.
func init() {
	const grantKey = "galufs-final-act/dies-counters"
	dies := effects.EerieDiesWithPower("Galuf's Final Act — put +1/+1 counters equal to its power on up to one target creature",
		effects.UpToOneTargetCreature("up to one target creature"),
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok || item.Params.Amount <= 0 {
				return nil
			}
			return effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: item.Params.Amount}.Apply(ctx)
		})
	effects.Register(effects.Spec{
		OracleID:     "9268d369-173e-4325-98f6-348beeb8ba62",
		Name:         "Galuf's Final Act",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)),
		Grants: []effects.AbilityGrant{{
			Key:       grantKey,
			Triggered: []game.TriggeredAbility{dies},
			Text:      "When this creature dies, put a number of +1/+1 counters equal to its power on up to one target creature.",
		}},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				target, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: target, Power: 1, Label: "Galuf's Final Act — +1/+0"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				target, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantAbilitiesFor{
					Target: target,
					Keys:   []string{grantKey},
					Label:  "Galuf's Final Act — until end of turn, it puts counters when it dies",
				}.Apply(ctx)
			}),
		),
	})
}

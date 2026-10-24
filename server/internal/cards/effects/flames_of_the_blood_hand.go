package effects

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Flames of the Blood Hand — Instant {2}{R}:
//
//	"Flames of the Blood Hand deals 4 damage to target player or
//	 planeswalker. The damage can't be prevented. If that player or that
//	 planeswalker's controller would gain life this turn, that player
//	 gains no life instead."
//
// Two of ADR 0107 §5's pieces. The damage's "can't be prevented" is the
// spell's own rider. The last sentence is a REPLACEMENT ("instead"),
// not "can't gain life": a ModGainNoLife record on the life window for
// the rest of the turn, so another "if you would gain life" replacement
// may be ordered against it (CR 616.1). The player is the target, or
// the planeswalker's controller as the spell resolves.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:                   "4e9df979-c1c2-4de1-944e-c5e2d782e66e",
		Name:                       "Flames of the Blood Hand",
		Completeness:               CompletenessFull,
		SpellDamageCantBePrevented: Always(),
		Targets:                    targetPlayerOrPlaneswalker(),
		Purpose:                    ForTargets(DamageToTarget(0, 4)),
		// The no-lifegain replacement waits for a prompt the damage asks (a CR 616 order;
		// CR 608.2c, #4155). The player is read before the damage, since a planeswalker
		// it destroys is no longer there to say who controlled it.
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				targets := ctx.LegalTargets()
				if len(targets) == 0 {
					return nil
				}
				t := targets[0]
				player := t.ID
				if t.Kind == game.TargetCard {
					c, ok := ctx.Game.LookupCardForEffect(t.ID)
					if !ok {
						return nil
					}
					player = c.Controller
				}
				ctx.RecordStep(flamesOfTheBloodHandPlayer, []uuid.UUID{player})
				return DealDamage{Source: ctx.Item.SourceCardID, Target: t.ID, Amount: 4}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				ids := ctx.StepRecord(flamesOfTheBloodHandPlayer).IDs
				if len(ids) == 0 {
					return nil
				}
				ctx.Game.GainNoLifeThisTurnForEffect(ctx.Item.SourceCardID, ids[0], "Flames of the Blood Hand — gains no life this turn")
				return nil
			}),
		),
	})
}

// flamesOfTheBloodHandPlayer files the player the damage went to (or whose
// planeswalker it went to), for the life they gain none of (#4155).
const flamesOfTheBloodHandPlayer = "flames-of-the-blood-hand-player"

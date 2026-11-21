package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Violent Urge — Instant {R} (Duskmourn: House of Horror, #4466):
//
//	"Target creature gets +1/+0 and gains first strike until end of turn.
//	 Delirium — If there are four or more card types among cards in your
//	 graveyard, that creature gains double strike until end of turn."
//
// Kindled Heroism's pump and grant (pumpAndGrantFirstTarget) on the
// target still legal as the spell resolves (CR 608.2b), then the
// delirium check, read as the spell resolves (drcDeliriumMet, the
// caster's own graveyard; the spell itself is still on the stack, so it
// does not count), adding double strike to the same creature. Delirium
// is an ability word (CR 207.2c).
//
// No simplification.
func init() {
	pumpAndGrant := effects.PumpAndGrantFirstTarget(1, 0, "Violent Urge — +1/+0 and first strike", "first strike")
	effects.Register(effects.Spec{
		OracleID:     "0ef80a38-f464-4d1f-9e81-d4087d6ecc9b",
		Name:         "Violent Urge",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "first strike")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := pumpAndGrant(ctx.Game, item); err != nil {
				return err
			}
			target := effects.FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil || !effects.DrcDeliriumMet(ctx.Game, ctx.Controller()) {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"double strike"},
				Label: "Violent Urge — double strike until end of turn"}.Apply(ctx)
		},
	})
}

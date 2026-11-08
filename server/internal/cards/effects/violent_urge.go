package effects

import (
	"github.com/google/uuid"

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
	pumpAndGrant := pumpAndGrantFirstTarget(1, 0, "Violent Urge — +1/+0 and first strike", "first strike")
	Register(Spec{
		OracleID:     "0ef80a38-f464-4d1f-9e81-d4087d6ecc9b",
		Name:         "Violent Urge",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0, "first strike")),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := pumpAndGrant(ctx.Game, item); err != nil {
				return err
			}
			target := FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil || !drcDeliriumMet(ctx.Game, ctx.Controller()) {
				return nil
			}
			return GrantKeywordUntilEOT{Target: target, Keywords: []string{"double strike"},
				Label: "Violent Urge — double strike until end of turn"}.Apply(ctx)
		},
	})
}

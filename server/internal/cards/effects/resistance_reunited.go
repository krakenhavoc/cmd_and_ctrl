package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Resistance Reunited — Instant {1}{W} (Phyrexia: All Will Be One):
//
//	"Target creature gets +2/+2 until end of turn.
//	 Equipped creatures you control gain indestructible until end of turn."
//
// #5479, S82. Giant Growth's pump on the target, then one until-end-of-turn
// keyword grant whose Match is read once as the spell resolves (CR 611.2c):
// the creatures you control that are equipped then, the target included,
// keep indestructible even if the Equipment moves later.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e68fa0e6-9f55-498b-a13e-197cb30ca4aa",
		Name:         "Resistance Reunited",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (BoostUntilEOT{
				Target:    item.Targets[0].ID,
				Power:     2,
				Toughness: 2,
				Label:     "Resistance Reunited — +2/+2",
			}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl(), s82OneB14Equipped()),
				Keywords: []string{"indestructible"},
				Label:    "Resistance Reunited — equipped creatures gain indestructible",
			}.Apply(ctx)
		},
	})
}

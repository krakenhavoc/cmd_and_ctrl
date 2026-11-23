package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cheeky House-Mouse // Squeak By — Creature — Mouse {W}, 2/1 // Sorcery —
// Adventure {W} (Wilds of Eldraine, #4471):
//
//	Cheeky House-Mouse — (no rules text)
//	Squeak By — "Target creature you control gets +1/+1 until end of turn. It
//	 can't be blocked by creatures with power 3 or greater this turn.
//	 (Then exile this card. You may cast the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID and
// "#1"). Squeak By is two until-end-of-turn effects on the one target: the pump
// and the evasion, a block restriction that reads each blocker's power live as
// blockers are declared (CantBeBlockedThisTurnByPowerAtLeast), so raising a
// blocker's power afterwards bars it but does not remove it from a block it is
// already in (the ruling).
//
// No simplification.
const cheekyHouseMouseOracleID = "db70ddae-0ac1-47f2-9158-8448e337a08f"

func init() {
	effects.Register(effects.Spec{
		OracleID:     cheekyHouseMouseOracleID,
		Name:         "Cheeky House-Mouse",
		Completeness: effects.CompletenessFull,
	})
	effects.Register(effects.Spec{
		OracleID:     cheekyHouseMouseOracleID + "#1",
		Name:         "Squeak By",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)),
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Squeak By — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			return effects.CantBeBlockedThisTurnByPowerAtLeast{Target: id, AtLeast: 3, Label: "Squeak By — can't be blocked by creatures with power 3 or greater"}.Apply(ctx)
		},
	})
}

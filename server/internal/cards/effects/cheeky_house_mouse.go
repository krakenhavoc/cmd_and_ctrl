package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     cheekyHouseMouseOracleID,
		Name:         "Cheeky House-Mouse",
		Completeness: CompletenessFull,
	})
	Register(Spec{
		OracleID:     cheekyHouseMouseOracleID + "#1",
		Name:         "Squeak By",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 1)),
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Squeak By — +1/+1"}).Apply(ctx); err != nil {
				return err
			}
			return CantBeBlockedThisTurnByPowerAtLeast{Target: id, AtLeast: 3, Label: "Squeak By — can't be blocked by creatures with power 3 or greater"}.Apply(ctx)
		},
	})
}

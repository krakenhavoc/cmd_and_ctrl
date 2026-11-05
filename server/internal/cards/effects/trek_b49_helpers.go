package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// trekB49AttackerPumpThreeTrample is "Target attacking creature gets
// +3/+3 and gains trample until end of turn.", printed on Smashing Spree
// and Run Amok. The target clause is AttackingCreature; the body is a
// pump and a keyword grant on the one target, both until end of turn.
// If the target is no longer attacking at resolution it is illegal and
// the spell does nothing (CR 608.2b).
func trekB49AttackerPumpThreeTrample(oracleID, name string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking creature", AttackingCreature()),
		Purpose:      ForTargets(BenefitUntilEOT(0, 3, 3, "trample")),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (BoostUntilEOT{Target: id, Power: 3, Toughness: 3, Label: name + " — +3/+3"}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: name + " — trample"}.Apply(ctx)
		},
	}
}

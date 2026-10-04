package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bridgeworks Battle // Tanglespan Bridgeworks — modal double-faced
// card. This file is the FRONT face, Sorcery {2}{G}:
//
//	"Target creature you control gets +2/+2 until end of turn. It
//	 fights up to one target creature you don't control."
//
// The back face, Tanglespan Bridgeworks, is registered with the MDFC
// land cycle in mdfc_lands.go under "<oracle>#1".
//
// Two target clauses, each with its own predicate (CR 601.2c, re-checked
// per slot at resolution, CR 608.2b): slot 0 is your creature, slot 1 is
// "up to one" creature you don't control. The pump comes first and stays
// even when the fight does not happen: a victim declined or gone only
// turns the fight off. If your creature is gone, neither half happens.
// The fight is b10Fight, which reads both powers after the +2/+2.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d581188-ce80-494e-bd38-f411e1f4efb5",
		Name:         "Bridgeworks Battle",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("up to one target creature you don't control", OpponentControls()).WithCount(0, 1),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || mine.Kind != game.TargetCard {
				return nil
			}
			if err := (BoostUntilEOT{
				Target: mine.ID, Power: 2, Toughness: 2,
				Label: "Bridgeworks Battle — +2/+2",
			}).Apply(ctx); err != nil {
				return err
			}
			theirs, ok := ctx.ClauseTarget(1)
			if !ok || theirs.Kind != game.TargetCard {
				return nil
			}
			return b10Fight(ctx, mine.ID, theirs.ID)
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rally the Monastery — Instant {3}{W} (Tarkir: Dragonstorm, #4463):
//
//	"This spell costs {2} less to cast if you've cast another spell this
//	 turn.
//	 Choose one —
//	 • Create two 1/1 white Monk creature tokens with prowess.
//	 • Up to two target creatures you control each get +2/+2 until end of
//	   turn.
//	 • Destroy target creature with power 4 or greater."
//
// The reduction is Focus the Mind's (tdmCastAnotherSpellThisTurn). The
// pump gives +2/+2 to each of its targets still legal as the spell
// resolves (CR 608.2b), and "power 4 or greater" is judged at announce and
// again then, after every pump and counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3f6101a5-3038-45b0-8302-b2e4127b9510",
		Name:         "Rally the Monastery",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if you've cast another spell this turn.",
				tdmCastAnotherSpellThisTurn),
		},
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Create two 1/1 white Monk creature tokens with prowess.", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return CreateToken{Template: TokenCard("1/1 white Monk with prowess"), N: 2}.Apply(ctx)
				}), game.Purpose{Tokens: 2}),
			ModeWithPurpose(ModeDoing("Up to two target creatures you control each get +2/+2 until end of turn.",
				TargetCreature("up to two target creatures you control", YouControl()).WithCount(0, 2),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Rally the Monastery — +2/+2"}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}), ForTargets(BenefitUntilEOT(0, 2, 2))),
			ModeDoing("Destroy target creature with power 4 or greater.",
				TargetCreature("target creature with power 4 or greater", PowerGE(4)), DestroyTheModesTarget),
		),
	})
}

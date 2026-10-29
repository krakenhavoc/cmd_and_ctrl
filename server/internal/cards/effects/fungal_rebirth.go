package effects

// Fungal Rebirth — #4380, S76. Uses the existing catalog seams.
func init() {
	Register(Spec{
		OracleID: "b917f39a-3176-47ad-afda-08d09b873c02", Name: "Fungal Rebirth", Completeness: CompletenessFull,
		Targets: TargetCardInGraveyard("target permanent card from your graveyard", YouOwn(), Permanent()),
		OnResolve: ResolveSteps(Instruction(func(ctx *Context) error { return returnTargetFromGraveyardToHand(ctx.Item, ctx) }), Instruction(func(ctx *Context) error {
			if b11CreaturesDiedThisTurn(ctx.Game) == 0 {
				return nil
			}
			return CreateToken{Template: TokenCard("1/1 green Saproling"), N: 2}.Apply(ctx)
		})),
	})
}

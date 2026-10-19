package effects

// Legion Leadership // Legion Stronghold — Modal DFC. Front: Instant
// {1}{R/W}: "Until end of turn, double target creature's power and it
// gains first strike." Back: Land (mdfc_lands.go).
//
// Doubling is a +N/+0 boost where N is the creature's power as the spell
// resolves, counters and other effects included (CR 608.2h), the same way
// Bulk Up does it. The first strike follows as its own step. A target that
// is no longer legal means neither half happens. No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad225ec2-ff3a-48f6-81a7-dfdd1b75e1f7",
		Name:         "Legion Leadership",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				c, ok := ctx.Game.LookupCardForEffect(id)
				if !ok {
					return nil
				}
				return BoostUntilEOT{
					Target: id,
					Power:  c.CurrentPower(),
					Label:  "Legion Leadership — double power",
				}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return GrantKeywordUntilEOT{
					Target:   id,
					Keywords: []string{"first strike"},
					Label:    "Legion Leadership — first strike",
				}.Apply(ctx)
			}),
		),
	})
}

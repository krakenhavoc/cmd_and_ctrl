package effects

// Mutant Chain Reaction — {2}{G} Sorcery (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"Destroy up to one target artifact, enchantment, or creature with
//	 flying. Create a Mutagen token. (It's an artifact with
//	 "{1}, {T}, Sacrifice this token: Put a +1/+1 counter on target
//	 creature. Activate only as a sorcery.")"
//
// Armaggon's "up to" destroy with one pick, then the token. With no target
// chosen the spell still resolves and makes the token; with a target that
// has become illegal it does not resolve at all (CR 608.2b), so no token.
// The Mutagen token is the catalog's CR 111.10v template (MutagenToken, #4676).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ea9fc868-374b-4acc-be0d-4283907b4524",
		Name:         "Mutant Chain Reaction",
		Completeness: CompletenessFull,
		Targets: TargetPermanent("up to one target artifact, enchantment, or creature with flying",
			Or(Artifact(), Enchantment(), And(Creature(), HasKeyword("flying")))).WithCount(0, 1),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error { return destroyEachLegalTarget(ctx.Item, ctx) }),
			CreateToken{Template: MutagenToken(), N: 1},
		),
	})
}

package effects

// Finishing Move — Sorcery {2}{G}:
//
//	"You get {TK}{TK}, then you may put a sticker on a nonland permanent
//	 you own. Target creature you control deals damage equal to its power
//	 to target creature you don't control."
//
// Three steps in printed order (CR 608.2c), written as ResolveSteps so the
// damage waits for the sticker questions (ADR 0208 §4): a power and
// toughness sticker put on the biter changes the damage it deals. The
// bite is Bite Down's (oneSidedBite), its power read as it deals.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d41dc3a5-58e4-444a-b388-5a32ccbed0eb",
		Name:         "Finishing Move",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature you don't control", OpponentControls()),
		),
		OnResolve: ResolveSteps(
			GetTickets{N: 2},
			PutASticker{},
			Instruction(func(ctx *Context) error { return oneSidedBite(ctx.Item, ctx) }),
		),
	})
}

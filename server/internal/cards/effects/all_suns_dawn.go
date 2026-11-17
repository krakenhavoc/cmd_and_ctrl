package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// All Suns' Dawn — Sorcery {4}{G} (Fifth Dawn, #5604):
//
//	"For each color, return up to one target card of that color from
//	 your graveyard to your hand. Exile All Suns' Dawn."
//
// Five "up to one" clauses, one per color in WUBRG order (ADR 0065),
// so the spell has from zero to five targets. The word "target" appears
// once, so no card can be chosen for two colors (CR 115.3, CR 601.2c):
// every clause after the first is Distinct. A multicolored card fills
// any one of its colors' slots, and two white-and-blue cards can fill
// the white slot and the blue slot. Each clause re-checks its own color
// at resolution, so a card chosen as blue that is no longer blue stays
// in the graveyard while the others still return.
//
// "Exile All Suns' Dawn" is done as the spell finishes resolving
// (Eldritch Evolution's tail), so it goes to exile rather than the
// graveyard. If every chosen target is illegal, the spell does not
// resolve (CR 608.2b) and goes to the graveyard instead.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a97d9a72-2e47-491b-b1db-9829a9abbbd7",
		Name:         "All Suns' Dawn",
		Completeness: CompletenessFull,
		Targets: Clauses(
			allSunsDawnClause("white", "W"),
			Distinct(allSunsDawnClause("blue", "U")),
			Distinct(allSunsDawnClause("black", "B")),
			Distinct(allSunsDawnClause("red", "R")),
			Distinct(allSunsDawnClause("green", "G")),
		),
		OnResolve: ResolveSteps(
			Instruction(returnLegalGraveyardTargetsToHand),
			Instruction(func(ctx *Context) error { return ExileTarget{Target: ctx.Source()}.Apply(ctx) }),
		),
	})
}

// allSunsDawnClause is one color's "up to one target card of that color
// from your graveyard".
func allSunsDawnClause(word, letter string) *game.TargetSpec {
	return TargetCardInGraveyard("up to one target "+word+" card from your graveyard",
		OfColor(letter), YouOwn()).WithCount(0, 1)
}

package effects

// Transmutation Font — Artifact {5} (EDHREC rank 3414):
//
//	"{T}: Create your choice of a Blood token, a Clue token, or a
//	 Food token.
//	 {3}, {T}, Sacrifice three artifact tokens with different names:
//	 Search your library for an artifact card, put it onto the
//	 battlefield, then shuffle. Activate only as a sorcery."
//
// The token-maker whose payoff is a tutor. "Your choice of" is three
// tap abilities with the same cost, one per token — Insidious
// Fungus's shape for a modal activated ability (the catalog has no
// mode picker on a CR 602 ability, and the client's ability menu is
// one): the choice is made at activation either way, and the shared
// tap means exactly one of them per untap, as printed. Each token is
// the real thing, with its own sacrifice ability.
//
// Sandbox simplification, declared (the Magda posture: one whole
// ability omitted): the tutor is not implemented. "Sacrifice three
// artifact tokens with different names" is a three-permanent
// sacrifice cost with a distinctness clause. Since #747 a cost can
// sacrifice three (SacrificeN), but "with different names" is a
// restriction on the SET, which no per-permanent predicate can say,
// and #747 left set-level restrictions out of scope (decided
// 2026-09-17; the "Set-level restriction on a sacrifice cost" row in
// docs/engine-seams.md). Shipping the tutor for any three artifact
// tokens would be stronger than printed (#259), so the ability stays
// out rather than shipping a cheaper one. The Font is still
// recognisably itself: a colourless Blood / Clue / Food engine.
// Weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "a68e57f3-f138-4c26-b8d5-55d1adec44a8",
		Name:         "Transmutation Font",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The tutor isn't implemented — sacrificing three differently named artifact tokens to search for an artifact isn't offered; the Font only makes Blood, Clue and Food tokens."},
		Activated: []ActivatedAbility{
			{
				Label:  "{T}: Create a Blood token.",
				Cost:   TapCost(),
				Effect: b32CreateTokenBody(BloodToken),
			},
			{
				Label:  "{T}: Create a Clue token.",
				Cost:   TapCost(),
				Effect: b32CreateTokenBody(ClueToken),
			},
			{
				Label:  "{T}: Create a Food token.",
				Cost:   TapCost(),
				Effect: b32CreateTokenBody(FoodToken),
			},
		},
	})
}

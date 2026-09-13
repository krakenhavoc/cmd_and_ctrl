package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eerie Ultimatum — Sorcery {W}{W}{B}{B}{B}{G}{G} (EDHREC rank
// 1534):
//
//	"Return any number of permanent cards with different names from
//	 your graveyard to the battlefield."
//
// The seven-mana mass reanimation. The choice is the caster's, so it
// is asked through the graveyard picker as an "any number" clause
// over permanent cards the caster owns, and the cards come back
// under their owner's control — the caster's — in announce order,
// each entering through the ordinary reanimation path so its own
// enters-tapped clause and every ETB trigger fire.
//
// Sandbox simplifications, declared, both weaker than printed:
//
//   - The printed text does not target; this clause does, because a
//     pick-from-graveyard prompt that is not a target clause does
//     not exist. The difference is that an opponent can exile a
//     chosen card in response and that card then stays put — the
//     rest still return.
//   - "With different names" cannot be enforced at announce, so it
//     is enforced at resolution: of two same-named picks only the
//     first returns.
//
// "Any number" includes none, so the spell can be cast with an empty
// graveyard and do nothing.
func init() {
	Register(Spec{
		OracleID:     "5674f6ae-ed5d-441e-a534-b5dd415165fd",
		Name:         "Eerie Ultimatum",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The cards are chosen as targets, so an opponent can exile one in response and it stays in the graveyard.",
			"If you pick two cards with the same name, only the first one returns.",
		},
		Targets: TargetCardInGraveyard("any number of permanent cards with different names from your graveyard",
			YouOwn(), Permanent()).WithCount(0, 0),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b14ReturnDistinctNamesFromGraveyard(ctx)
		},
	})
}

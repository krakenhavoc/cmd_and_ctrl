package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mockingbird — Creature — Bird Bard {X}{U}, 1/1:
//
//	"Flying
//	 You may have this creature enter as a copy of any creature on the
//	 battlefield with mana value less than or equal to the amount of
//	 mana spent to cast this creature, except it's a Bird in addition
//	 to its other types and it has flying."
//
// A Clone whose ceiling is what you paid for it. The copy choice is the
// shared EntersAsCopyOf machinery (CR 707.2 + CR 614.1c, ADR 0043); the
// ceiling is the one thing a plain Clone does not need, and it is why
// this card shipped as a vanilla flier until #1735.
//
// "The amount of mana spent to cast this creature" (CR 601.2h) is the
// spell's payment, not its mana value: {X}{U} cast for X=3 spent four,
// and a commander tax or a Thalia on top spent more. The engine reads
// it off the resolving spell's payment record inside the CR 614 entry
// window — game.CastCounts.ManaSpent, handed to the candidate filter by
// EntersAsCopyOfFromCast through game.EntryCastCountsForEffect. The
// candidate list is evaluated twice (when the prompt is built and when
// the answer arrives) and both see the same figure.
//
// What the printed card says, and the engine does:
//
//   - Mana value is the copied creature's own, read where it sits:
//     {X} counts as 0 on the battlefield (CR 202.3e), a token that
//     copies nothing is 0, and a creature whose cost the engine can't
//     read is never offered (rather than offered as 0).
//   - A Mockingbird that was not cast — reanimated, flickered, put
//     onto the battlefield — spent nothing, so it can copy only a
//     creature with mana value 0. The same holds for one cast without
//     paying its mana cost, and for the token a copy of the Mockingbird
//     spell becomes (CR 707.10: nothing was spent to cast a copy).
//   - Not targeting (CR 115.10a): hexproof and shroud don't narrow the
//     choice, and any controller's creature is fair game.
//   - "Except it's a Bird in addition to its other types and it has
//     flying" is part of the copiable values (CR 707.9b, CR 707.9a), so
//     a Clone that later copies this Mockingbird is a flying Bird too.
//   - Declining leaves the printed 1/1 Bird Bard flier.
//
// Declared caveat, the one every mana-spent reader in the catalog
// carries: with strict mana off the engine does not charge the cast
// (PaidCost.OnPaper), so it does not know what was spent and answers
// zero — the weaker-than-printed direction ADR 0068 §3 requires. The
// copy is then limited to mana value 0.
func init() {
	Register(Spec{
		OracleID:        "b9df2cdf-397c-458d-89cf-911568737ffa",
		Name:            "Mockingbird",
		Completeness:    CompletenessCaveats,
		PrintedKeywords: []string{"flying"},
		Caveats: []string{
			"With strict mana off, the game doesn't track how much mana you spent, so Mockingbird can only copy a creature with mana value 0 — turn strict mana on for it to count.",
		},
		Replacements: []game.ReplacementEffect{
			EntersAsCopyOfFromCast(
				"Mockingbird",
				mockingbirdCandidates,
				func(_ *game.ReplacementEvent, v *game.PrintedValues, _ *game.Game, _ *game.Card) {
					v.AddSubtype("Bird")
					v.AddKeyword("flying")
				},
			),
		},
	})
}

// mockingbirdCandidates is "any creature on the battlefield with mana
// value less than or equal to the amount of mana spent to cast this
// creature" — any controller, never Mockingbird itself.
func mockingbirdCandidates(g *game.Game, _ uuid.UUID, self uuid.UUID, cast game.CastCounts) []uuid.UUID {
	return copyCandidates(g, self, func(c game.Card) bool {
		if !c.IsCreature() {
			return false
		}
		mv, ok := g.ManaValueForEffect(c)
		return ok && mv <= cast.ManaSpent
	})
}

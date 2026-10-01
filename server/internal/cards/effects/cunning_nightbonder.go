package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cunning Nightbonder — Creature — Human Rogue {U/B}{U/B}, 2/2:
//
//	"Flash
//	 Spells with flash you cast cost {1} less to cast and can't be
//	 countered."
//
// Thryx's sentence with a different filter. The discount is a
// CostModifier; "can't be countered" is ADR 0106 §4's battlefield static
// (#1806) in its "you cast" form, so a copy (not cast, CR 707.10) is
// never covered. "With flash" is the keyword on the spell itself
// (SpellWithKeyword / HasKeyword): a spell cast at instant speed only
// because something lets you cast it "as though it had flash" does not
// have flash and gets neither half.
//
// No simplification.
func init() {
	const clause = "Spells with flash you cast cost {1} less to cast and can't be countered."
	Register(Spec{
		OracleID:        "7351317d-62a1-4894-9811-aa950c49eece",
		Name:            "Cunning Nightbonder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, clause, YourSpell(), SpellWithKeyword("flash")),
		},
		SpellsCantBeCountered: []game.CounterShieldStatic{
			SpellsYouCastCantBeCountered(clause, HasKeyword("flash")),
		},
	})
}

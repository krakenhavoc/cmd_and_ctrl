package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warden of Evos Isle — Creature — Bird Wizard {2}{U}, 2/2 (EDHREC
// rank 3854):
//
//	"Flying
//	 Creature spells with flying you cast cost {1} less to cast."
//
// The card this batch was worth re-triaging for. Its batch issue
// files it under "cost modification, alternative casts and costs
// computed at activation" — the #1 blocker in all three of the
// roadmap's ranking tables — and that shipped as #93 in S28. The
// issue was written before the detector knew, so the Warden sat in a
// blocked list for weeks with nothing actually blocking it.
//
// It is a plain CostReduction: a board modifier, keyed on the caster
// being this permanent's controller (YourSpell), on the spell being a
// creature (CreatureSpell), and on that creature spell having flying
// (SpellWithKeyword). The reduction spends against generic mana only
// and stops at zero — that rule is game.reduceGeneric's, not this
// file's, so a {U}{U} flier still costs {U}{U}.
//
// SpellWithKeyword is new and lives with the other predicates rather
// than in a batch helper: "spells with <keyword> you cast" is a whole
// template (Sephara, Urza's Incubator's cousins, the tribal lords
// that discount), not a one-card shape.
//
// The keyword is read off the SPELL, not off a battlefield
// permanent — HasKeyword falls back to the card's own Keywords and
// then to CatalogPrintedKeywords when the object is not on the
// battlefield, which is exactly the case here: the thing being priced
// is on the stack. A flier that only gains flying once it lands does
// not get the discount, which is correct (CR 601.2f prices the spell
// as it exists on the stack).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f424b5e9-8f02-4491-a7d8-c7e088611c6a",
		Name:            "Warden of Evos Isle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Warden of Evos Isle — creature spells with flying cost {1} less",
				YourSpell(), CreatureSpell(), SpellWithKeyword("flying")),
		},
	})
}

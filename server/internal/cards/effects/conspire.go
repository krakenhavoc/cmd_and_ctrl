package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// conspire.go — CR 702.78, conspire (#4690). Append-only.
//
// The cost is an optional cost keyed game.ConspireKey that taps two
// untapped creatures sharing a colour with the spell; the copy is the
// engine's trigger (game/conspire.go), read off the paid record, so a
// card declares only the cost:
//
//	OptionalCosts: []game.AdditionalCost{Conspire()},                          // a printed "Conspire"
//	GrantedOptionalCosts: []game.GrantedOptionalCost{ConspireForSpellsYouCast(  // Raiding Schemes
//	    "Each noncreature spell you cast has conspire.", Noncreature())},
//
// The two creatures ride cast_spell's conspire_ids, a list apart from the
// spell's own taps on teamwork_ids (#5688), so a spell whose own cost taps
// permanents can be conspired too.

// conspireTapLabel is the clause as printed without the verb; the
// client's picker reads "Tap two untapped creatures you control that
// share a color with it".
const conspireTapLabel = "two untapped creatures you control that share a color with it"

// Conspire is "As an additional cost to cast this spell, you may tap two
// untapped creatures you control that each share a color with it" (CR
// 702.78a). Paid once per instance: the taps ride cast_spell's
// conspire_ids.
func Conspire() game.AdditionalCost {
	tc := tapOthersPrice(2, conspireTapLabel, Creature())
	tc.SharesColorWithSource = true
	return game.AdditionalCost{
		Optional:  true,
		Key:       game.ConspireKey,
		TapOthers: tc,
		Label:     "Conspire",
	}
}

// ConspireForSpellsYouCast is "<spells> you cast have conspire" (Raiding
// Schemes), declared in Spec.GrantedOptionalCosts. EachSource, for CR
// 702.78b: "If a spell has multiple instances of conspire, each is paid
// separately and triggers based on its own payment, not any other
// instance of conspire." Two permanents granting it, or the grant beside
// a printed conspire, give the spell two costs.
func ConspireForSpellsYouCast(label string, spells CardPredicate) game.GrantedOptionalCost {
	gr := SpellsYouCastHave(label, spells, Conspire())
	gr.EachSource = true
	return gr
}

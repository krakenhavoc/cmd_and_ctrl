package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// search_rules.go — the registered set rules of a library search (#4253
// slice 5): SearchLibrary.Rule, the data form of SearchLibrary.Validate.
// A search whose set rule is one of these, and whose "then" is a
// registered tail, is a restore point while its prompt is open; a closure
// Validate keeps it off the disk.
//
// One key per clause, shared by every card that prints it. The keys are
// on-disk identities: never renamed, never reused (ledgered "searchrule
// <key>").

var (
	// searchRuleDifferentNames is "cards with different names": Gifts
	// Ungiven, Emergent Ultimatum, Tiamat, Shared Summons, Fblthp, Knows the
	// Way, and Saheeli Rai's −7.
	searchRuleDifferentNames = game.RegisterSearchRule("search/different-names",
		func(picked []game.Card, _ int) bool { return DifferentNames(picked) })

	// searchRuleBasicsShareALandType is Myriad Landscape's "two basic land
	// cards that share a land type".
	searchRuleBasicsShareALandType = game.RegisterSearchRule("search/basics-share-a-land-type",
		func(picked []game.Card, _ int) bool { return basicsShareALandType(picked) })

	// searchRuleTotalManaValueAtMost is "with total mana value N or less"
	// (Protean Hulk), N the rule's number.
	searchRuleTotalManaValueAtMost = game.RegisterSearchRule("search/total-mana-value-at-most",
		func(picked []game.Card, n int) bool { return b23TotalManaValueAtMost(n)(picked) })
)

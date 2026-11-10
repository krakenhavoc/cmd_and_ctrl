package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Conspiracy Unraveler — Creature — Sphinx Detective {5}{U}{U}, 6/6
// (Murders at Karlov Manor, #4469):
//
//	"Flying
//	 You may collect evidence 10 rather than pay the mana cost for spells
//	 you cast. (To collect evidence 10, exile cards with total mana value
//	 10 or greater from your graveyard.)"
//
// A granted alternative cost (CR 118.9) whose whole price is collect
// evidence 10 (CR 701.59a, CollectEvidenceTenForSpellsYouCast): the
// caster names graveyard cards with total mana value 10 or more, and a
// graveyard that falls short withholds the offer (CR 701.59b). It
// reaches every zone a spell is cast from; a spell cast from the
// graveyard is already on the stack (CR 601.2a), so it can't pay for
// itself. As an alternative cost it can't combine with another one (CR
// 118.9a), an {X} is 0 (CR 107.3b), and additional costs are still paid.
// Paying it collects evidence ("whenever you collect evidence" sees it),
// but it is not a spell's own collect evidence cost, so "if evidence was
// collected" stays false (CR 701.59c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:                "e4541d02-2860-46ef-b630-1d6307b60766",
		Name:                    "Conspiracy Unraveler",
		Completeness:            CompletenessFull,
		PrintedKeywords:         []string{"flying"},
		GrantedAlternativeCosts: []game.GrantedAlternativeCost{CollectEvidenceTenForSpellsYouCast()},
	})
}

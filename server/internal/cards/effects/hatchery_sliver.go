package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hatchery Sliver — Creature — Sliver {1}{G}, 2/2 (Commander Masters):
//
//	"Replicate {1}{G} (When you cast this spell, copy it for each time you
//	 paid its replicate cost.)
//	 Each Sliver spell you cast has replicate. The replicate cost is equal
//	 to its mana cost. (A copy of a permanent spell becomes a token.)"
//
// The card's own replicate is the keyword's pair (an optional cost paid any
// number of times and the cast trigger, CR 702.56a). The grant is Hugh,
// First of Many's shape: a granted optional cost on every Sliver spell
// whose price is the spell's own mana cost (AdditionalCost.
// ManaCostIsSpellCost, written in as the grant is filed), with the trigger
// that copies the spell once per payment. A copy of a permanent spell
// becomes a token as it resolves (CR 707.10, CR 111.13).
//
// A Hatchery Sliver cast with another one out has replicate twice (CR
// 702.56b), both {1}{G}: the engine files one replicate cost per spell, and
// paying it n times copies it n times, which is what two instances paid
// n/2 times each would copy. The copy trigger is the grantor's, so two
// Hatchery Slivers on the battlefield copy a Sliver spell once.
//
// No simplification.
func init() {
	spellCost := replicateMana("", 20)
	spellCost.ManaCostIsSpellCost = true
	spellCost.Label = "Replicate"
	isSliver := Subtype("Sliver")
	Register(Spec{
		OracleID:     "6227fb65-d7ef-4bc6-a076-9aa10941632c",
		Name:         "Hatchery Sliver",
		Completeness: CompletenessFull,
		OptionalCosts: []game.AdditionalCost{
			replicateMana("{1}{G}", 20),
		},
		GrantedOptionalCosts: []game.GrantedOptionalCost{SpellsYouCastHave(
			"Each Sliver spell you cast has replicate. The replicate cost is equal to its mana cost.", isSliver, spellCost)},
		Triggered: []game.TriggeredAbility{
			Replicate(),
			s82SncReplicateTrigger("Hatchery Sliver — replicate: copy the Sliver spell",
				func(g *game.Game, caster uuid.UUID, spell game.Card) bool { return isSliver(g, caster, spell) }),
		},
	})
}

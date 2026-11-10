package effects

import (
	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// granted_collect_evidence.go — collect evidence N as a granted
// alternative cost (CR 118.9, 701.59a): Conspiracy Unraveler's "You may
// collect evidence 10 rather than pay the mana cost for spells you
// cast." The engine half is game.AlternativeCost.CollectEvidence: the
// caster names graveyard cards with total mana value N or more on
// alt_cost_ids, a graveyard that falls short withholds the offer (CR
// 701.59b), and paying it emits the collect evidence event. It is not
// the spell's own optional collect evidence cost, so "if evidence was
// collected" never reads it (CR 701.59c).

// GrantedAltCostCollectEvidenceTen is "You may collect evidence 10
// rather than pay the mana cost for spells you cast." (Conspiracy
// Unraveler). An on-disk identity: never renamed, never reused.
const GrantedAltCostCollectEvidenceTen = "granted-collect-evidence-10"

// CollectEvidenceTenForSpellsYouCast is Conspiracy Unraveler's static.
// Every zone a spell is cast from, wherever its printed mana cost could
// be paid (CR 118.9a); a spell cast from the graveyard is on the stack
// by the time the cost is paid (CR 601.2a), so it can't be exiled to pay
// for itself. With no mana in the price, an {X} is 0 (CR 107.3b).
func CollectEvidenceTenForSpellsYouCast() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:             GrantedAltCostCollectEvidenceTen,
			Label:           "Collect evidence 10 rather than pay this spell's mana cost",
			PayLabel:        collectEvidenceLabel(10),
			CollectEvidence: game.CollectEvidenceCost(10, collectEvidenceLabel(10)),
		},
	}
}

// checkGrantedCollectEvidence is Register's guard on a granted collect
// evidence offer: the clause is collect evidence's own, and it is the
// whole price — no mana, life, energy or other card component beside it.
func checkGrantedCollectEvidence(name string, o game.AlternativeCost) {
	if !o.CollectEvidence.IsCollectEvidence() {
		panic(fmt.Sprintf("effects.Register: %q granted alternative cost %q collects evidence with a clause that is not collect evidence N — build it with game.CollectEvidenceCost", name, o.Key))
	}
	if o.ManaCost != "" || o.Life != 0 || o.Energy != 0 || o.ExileFromHand != nil || o.DiscardFromHand != nil ||
		o.ReturnToHand != nil || o.ExileFromGraveyard != nil || o.Sacrifice != nil || !o.TapOthers.Empty() ||
		o.Behold != nil || o.RemoveCounters != nil || o.ReducedBySacrificedManaValue {
		panic(fmt.Sprintf("effects.Register: %q granted alternative cost %q collects evidence beside another price", name, o.Key))
	}
}

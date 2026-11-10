package effects

import (
	"fmt"
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// collect_evidence_mana.go — collect evidence N as a MANA ability's cost
// (CR 701.59a, 605.1a): Cryptex's "{T}, Collect evidence 3: Add one mana
// of any color." It is the graveyard ExileCost clause a CR 602 ability's
// CollectEvidenceN is (game.CollectEvidenceCost), on
// ManaAbilityCost.ExileCards: the activator names the cards as
// activate_mana_ability's exile_ids, the validator refuses a pick whose
// total mana value falls short (ExileCost.FloorMet), and with too little
// in the graveyard the ability can't be activated (CR 701.59b, 118.3).
// Paying it emits game.EventCollectEvidence, as every other owner does.
//
// The auto-tapper never plans such an ability: which graveyard cards to
// spend is the player's decision (autoTapAcceptsBesidesManaCost refuses
// every ExileCards clause).

// ManaCollectEvidenceN is "Collect evidence N" inside a mana ability's
// cost: set it as ManaAbilityCost.ExileCards.
func ManaCollectEvidenceN(n int) *game.ExileCost {
	return game.CollectEvidenceCost(n, collectEvidenceLabel(n))
}

// checkManaCollectEvidence is Register's guard on a mana ability's
// collect evidence: it is paid out of the activator's graveyard by a
// permanent, so a hand ability (a Spirit Guide's shape) may not carry it,
// and its own source is never exiled by it. Card-file mistakes, refused
// at boot.
func checkManaCollectEvidence(name string, i int, ma ManaAbility) {
	if !ma.Cost.ExileCards.IsCollectEvidence() {
		return
	}
	if ma.Cost.ExileSelf || ma.Cost.Forage || len(ma.Zones) > 0 && !slices.Equal(ma.Zones, []game.ZoneKind{game.ZoneBattlefield}) {
		panic(fmt.Sprintf("effects.Register: %q mana ability %d collects evidence off the battlefield, beside a forage or exiling its source — a collect evidence mana ability is a permanent's", name, i))
	}
}

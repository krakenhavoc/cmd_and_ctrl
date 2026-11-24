package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap2_b03_helpers.go — card-side helpers of set GAP2, batch 3 (S82,
// #5948).

// BestowGraveyardKey is the key of a bestow offer a card lets its owner
// claim from their graveyard (Detective's Phoenix's "You may cast this card
// from your graveyard using its bestow ability."). The hand offer keeps
// game.BestowKey; an offer is keyed once per card, so the graveyard one has
// its own. An on-disk identity (StackItem.AltCost): never renamed, never
// reused.
const BestowGraveyardKey = "bestow/graveyard"

// BestowCollectEvidence is "Bestow—{mana}, Collect evidence N" (CR
// 702.103a, 701.59a): bestow's offer (Bestow), whose price is the mana and
// graveyard cards with total mana value N or greater, named on
// alt_cost_ids and exiled with the spell on the stack
// (game.AlternativeCost.CollectEvidence, the clause Conspiracy Unraveler's
// granted offer pays). `from` binds the offer to a zone: "" is the hand,
// game.ZoneGraveyard the graveyard offer keyed BestowGraveyardKey.
func BestowCollectEvidence(mana string, n int, from game.ZoneKind) game.AlternativeCost {
	ac := Bestow(mana)
	ac.Label = "Bestow—" + mana + ", Collect evidence " + strconv.Itoa(n)
	ac.PayLabel = collectEvidenceLabel(n)
	ac.CollectEvidence = game.CollectEvidenceCost(n, collectEvidenceLabel(n))
	if from != "" && from != game.ZoneHand {
		ac.Key = BestowGraveyardKey
		ac.FromZone = from
	}
	return ac
}

package effects

import (
	"fmt"
	"sync"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b43_ward_evidence.go — "Ward—Collect evidence N" (Axebane Ferox, #4452,
// CR 701.59a, CR 702.21a).
//
// The targeting player is the payer. When the ward triggers they are offered
// ONE choose_cards prompt over their own graveyard (any number, none
// included), whose registered set rule is game.CollectEvidenceMet, so a pick
// whose total mana value falls short of N is refused, as it is on every other
// collect evidence form. Choosing none lets the spell or ability be
// countered. A graveyard whose whole total is below N can't collect evidence at
// all (CR 701.59b), so nothing is asked and the spell is countered. The chosen
// cards are exiled (they are not targeted) and the collection is announced for
// "whenever you collect evidence" (game.EventCollectEvidence), by the payer. The
// pick is keyed, so a table waiting on it is a restore point, and one prompt
// shape per N is registered once.

var (
	wardEvidenceMu    sync.Mutex
	wardEvidencePicks = map[int]game.ChooseCardsThen{}
)

// WardCollectEvidence is "Ward—Collect evidence N".
func WardCollectEvidence(n int) WardCost { return WardCost{CollectEvidence: n} }

// wardEvidencePickFor registers, once per N, the pick's continuation
// ("choose-cards/ward-collect-evidence-<N>", an on-disk key) and its set rule.
func wardEvidencePickFor(n int) game.ChooseCardsThen {
	wardEvidenceMu.Lock()
	defer wardEvidenceMu.Unlock()
	if k, ok := wardEvidencePicks[n]; ok {
		return k
	}
	k := game.RegisterChooseCardsThenRule(fmt.Sprintf("choose-cards/ward-collect-evidence-%d", n),
		func(picked []game.Card) bool { return game.CollectEvidenceMet(picked, n) },
		func(g *game.Game, r game.CardsChosen) error {
			targetingItem := wardTargetingItem(r.Carry)
			if g.StackItemForEffect(targetingItem) == nil {
				return nil
			}
			if len(r.Picked) == 0 {
				return wardCounter(g, targetingItem)
			}
			for _, id := range r.Picked {
				if err := g.ExileCardForEffect(id); err != nil {
					return err
				}
			}
			g.NoteCollectEvidenceForEffect(r.Chooser, r.Source, r.Picked)
			return nil
		})
	wardEvidencePicks[n] = k
	return k
}

// wardCollectEvidenceOrCounter is "counter it unless that player collects
// evidence N".
func wardCollectEvidenceOrCounter(g *game.Game, source, payer uuid.UUID, n int, targetingItem uuid.UUID) error {
	options := g.CollectEvidenceOptionsForEffect(payer, uuid.Nil, n)
	if len(options) == 0 {
		return wardCounter(g, targetingItem)
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  payer,
		Source:   source,
		Question: fmt.Sprintf("Ward — collect evidence %d, or it is countered (choose none to let it be countered)", n),
		Cards:    options,
		Min:      0,
		Max:      len(options),
		LiveZone: game.ZoneGraveyard,
		ThenKey:  wardEvidencePickFor(n),
		Carry:    []uuid.UUID{targetingItem},
	})
	return nil
}

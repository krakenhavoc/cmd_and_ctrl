package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gapB36ExileAnyNumberWithTotalManaValue is "Exile any number of <kind> cards
// from your graveyard with total mana value N or greater" as a CR 602
// activated ability's cost (The Capitoline Triad): the activator names the
// cards at announce, and the activation is refused unless their mana values
// sum to the floor. It is collect evidence's clause (CR 701.59a) over a chosen
// kind of card, and never collect evidence itself.
func gapB36ExileAnyNumberWithTotalManaValue(label string, atLeast int, match func(game.Card) bool) game.AbilityCost {
	return game.AbilityCost{ExileCards: &game.ExileCost{
		AnyNumber:        true,
		Label:            label,
		From:             game.ZoneGraveyard,
		Match:            match,
		ManaValueAtLeast: atLeast,
	}}
}

// gapB36HistoricCardsInYourGraveyard counts the historic cards (CR 700.6) in
// the casting player's graveyard, as a cost reduction reads them.
func gapB36HistoricCardsInYourGraveyard(q game.CostQuery) int {
	if q.Game == nil || q.Controller == uuid.Nil {
		return 0
	}
	p := q.Game.PlayerByIDForEffect(q.Controller)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		if b09IsHistoric(c) {
			n++
		}
	}
	return n
}

// gapB36HasFreerunning is "card with freerunning" (CR 702.173a): a card whose
// catalog row prints the freerunning alternative cost.
func gapB36HasFreerunning(_ *game.Game, _ uuid.UUID, c game.Card) bool {
	spec, ok := Lookup(c.OracleID)
	if !ok {
		return false
	}
	for _, alt := range spec.AlternativeCosts {
		if alt.Key == "freerunning" {
			return true
		}
	}
	return false
}

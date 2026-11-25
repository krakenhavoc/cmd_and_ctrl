package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b041_helpers.go — shared pieces for the Strixhaven: School of Mages
// batch 041 (S82, #5502).

// OldB41InstantsAndSorceriesOwnedInExileAndGraveyard counts the instant and
// sorcery cards `owner` owns in exile (face up: a face-down card has no types,
// CR 708.2) and in their graveyard. Serpentine Curve reads it as the spell
// resolves (CR 608.2h).
func OldB41InstantsAndSorceriesOwnedInExileAndGraveyard(g *game.Game, owner uuid.UUID) int {
	n := FandanielInstantsAndSorceries(g, owner)
	if g == nil || g.Exile == nil {
		return n
	}
	for i := range g.Exile.Cards {
		c := &g.Exile.Cards[i]
		if c.Owner == owner && !c.FaceDown && (c.IsInstant() || c.IsSorcery()) {
			n++
		}
	}
	return n
}

// OldB41StudyCounter is the counter kind Kianne and Imbraham put on exiled cards.
const OldB41StudyCounter = "study"

// OldB41StudyCardsOwnedInExile lists the cards `owner` owns in exile with a
// study counter on them, face-up ones only (these cards are exiled face up).
func OldB41StudyCardsOwnedInExile(g *game.Game, owner uuid.UUID) []game.Card {
	var out []game.Card
	if g == nil || g.Exile == nil {
		return out
	}
	for _, c := range g.Exile.Cards {
		if c.Owner == owner && !c.FaceDown && c.Counters[OldB41StudyCounter] > 0 {
			out = append(out, c)
		}
	}
	return out
}

// OldB41DifferentManaValuesAmongStudyNonlands is Kianne's "each different mana
// value among nonland cards you own in exile with study counters on them".
func OldB41DifferentManaValuesAmongStudyNonlands(g *game.Game, owner uuid.UUID) int {
	seen := map[int]bool{}
	for _, c := range OldB41StudyCardsOwnedInExile(g, owner) {
		if !c.IsLand() {
			seen[c.ManaValue()] = true
		}
	}
	return len(seen)
}

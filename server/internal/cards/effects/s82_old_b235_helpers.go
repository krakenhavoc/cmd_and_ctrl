package effects

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b235_helpers.go — Hour of Devastation, batch 235 (#5534).

// S82OldB235CyclingCardsInGraveyard is "the number of cards with cycling in
// your graveyard" (Vile Manifestation): the cards in the player's graveyard
// that offer a cycling ability, typecycling included (CR 702.29a, 702.29e).
func S82OldB235CyclingCardsInGraveyard(g *game.Game, player uuid.UUID) int {
	p := g.PlayerByIDForEffect(player)
	if p == nil || p.Graveyard == nil {
		return 0
	}
	n := 0
	for _, c := range p.Graveyard.Cards {
		for _, ab := range game.ActivatedAbilitiesForCard(c) {
			if ab.Cycling {
				n++
				break
			}
		}
	}
	return n
}

// S82OldB235IsEternalizeOrEmbalm reports whether an activation event is of an
// eternalize or embalm ability (Vizier of the Anointed). The activation's
// label is the ability's printed label, which Embalm and Eternalize begin with
// their keyword (CR 702.128a, 702.129a); a SacrificeSelf-style cost has
// already ended the object by the time a watcher runs, so the label on the
// event is what is read.
func S82OldB235IsEternalizeOrEmbalm(ev game.Event) bool {
	return strings.HasPrefix(ev.Label, "Embalm ") || strings.HasPrefix(ev.Label, "Eternalize ")
}

// S82OldB235HasEternalizeOrEmbalm reports whether a card offers an eternalize
// or embalm ability (Vizier of the Anointed's search), read off the card's
// own activated rows: both keywords are rows whose label begins with the
// keyword (CR 702.128a, 702.129a).
func S82OldB235HasEternalizeOrEmbalm(c game.Card) bool {
	for _, ab := range game.ActivatedAbilitiesForCard(c) {
		if strings.HasPrefix(ab.Label, "Embalm ") || strings.HasPrefix(ab.Label, "Eternalize ") {
			return true
		}
	}
	return false
}

package effectstest

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// library_graveyard_search_test.go — #3217, "search your library and/or
// graveyard", through the real catalog cards (Tower Winder, Finale of
// Devastation, Ajani's Aid, and the five cards that shipped with the
// seam). The engine rule is pinned beside the code in
// game/search_graveyard_test.go; this file pins that the cards inherit
// it.
//
//   - a card in the graveyard is found, and a graveyard-only take does
//     not shuffle the library or fire EventSearchLibrary;
//   - a library take shuffles, and fires the event;
//   - "and/or" lets the searcher choose their zones, and asks them FIRST
//     (#4519): a graveyard match does not take away the right to find
//     nothing (the rulings on Ajani's Aid and Finale of Devastation), and
//     the library is shuffled if and only if the searcher chose it;
//   - the prompt offers the chosen zones, and says which is which.
//
// The helpers that answer a search (answerSearchByID and the others)
// choose every zone on offer; a test about a graveyard-only search answers
// the zone question itself (answerSearchZones) first.

const (
	lgTowerWinderOracle = "c1aecc8a-db3a-4158-b1f1-bfe06d782482"
	LgFinaleOracle      = "69872a9a-fe54-4e58-940c-89395af71acd"
	LgAjanisAidOracle   = "b3779fbe-7701-433b-afad-bebd6096a3c8"
	LgMoogleOracle      = "60043b87-6008-4cfa-9d8f-6ee6e3f70766"
)

// lgKnowLibrary marks every library card known to the seat, so a
// shuffle (which un-knows the zone) is observable as a change.
func LgKnowLibrary(p *game.Player) {
	for i := range p.Library.Cards {
		p.Library.Cards[i].AddKnower(p.ID)
	}
}

// lgCastTowerWinder casts Tower Winder and lets the enters trigger
// resolve to the point where the search either prompted or finished.
func LgCastTowerWinder(t *testing.T, g *game.Game) {
	t.Helper()
	CastCatalogSpell(t, g, "Tower Winder", "Creature — Snake", lgTowerWinderOracle, nil)
	PassPriorityAroundTable(t, g)
	PassPriorityAroundTable(t, g)
}

func LgTower() game.Card { return game.Card{Name: "Command Tower", TypeLine: "Land"} }

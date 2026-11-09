package coverage

import (
	_ "embed"
	"strings"
)

// planar.go — the census's planar line (ADR 0216 §6, slice 4, #4870):
// "planes and phenomena: N of M in the card database". A plane is a
// catalog card like any other, so it is counted in the registry keys
// above; this line says how much of the planar deck plays automatically,
// separately, because no deck list holds a plane and the Commander
// numbers must not be read as including them.
//
// M cannot be measured from the registry, and the Scryfall dump is not
// in CI (see census.go). So the planar rows of the card database are a
// checked-in list, planar_cards.txt, refreshed and checked against the
// dump by TestPlanarCardListMatchesTheDump whenever CMDCTRL_SCRYFALL_DUMP
// is set (the nightly), the same posture as the oracle fixtures.

//go:embed planar_cards.txt
var planarCardsFile string

// PlanarCard is one row of planar_cards.txt.
type PlanarCard struct {
	OracleID, Name, TypeLine string
}

// PlanarCards is every plane and phenomenon card in the card database,
// one per oracle ID, in oracle-ID order.
func PlanarCards() []PlanarCard {
	var out []PlanarCard
	for _, line := range strings.Split(planarCardsFile, "\n") {
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f := strings.Split(line, "\t")
		if len(f) != 3 {
			continue
		}
		out = append(out, PlanarCard{OracleID: f[0], Name: f[1], TypeLine: f[2]})
	}
	return out
}

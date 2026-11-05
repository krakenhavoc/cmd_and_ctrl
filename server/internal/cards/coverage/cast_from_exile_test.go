package coverage

import (
	"regexp"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// castFromExileLine is the printed clause on its own line: "You may cast
// this card from exile." Rebound's reminder text ("… you may cast this
// card from exile without paying its mana cost.") is a different sentence
// in parentheses and does not match.
var castFromExileLine = regexp.MustCompile(`(?im)^you may cast this card from exile\.$`)

// TestEveryExileCastableCardPrintsTheClause holds Spec.CastsItselfFromExile
// to the printed text (ADR 0066, amendment of 2026-11-05). The marker is
// the only thing that lets a card list ZoneExile in CastableZones, so a
// spec that sets it must print the clause, and a catalogued face that
// prints it must set it, or it is weaker than printed.
//
// The printed text is the checked-in oracle fixture. A spec whose fixture
// is not checked in yet is skipped here; the missing fixture is reported
// by TestOracleFixtureCoversRegistry.
func TestEveryExileCastableCardPrintsTheClause(t *testing.T) {
	fx := loadOracleFixture(t)
	checked := 0
	for _, s := range effects.All() {
		base, face := BaseOracleID(s.OracleID)
		oc, ok := fx[base]
		if !ok {
			continue
		}
		faces := oc.AllFaces()
		prints := face < len(faces) && castFromExileLine.MatchString(faces[face].Text)
		switch {
		case prints && !s.CastsItselfFromExile:
			t.Errorf("%s prints \"You may cast this card from exile.\" and does not declare it: list game.ZoneExile in CastableZones and set CastsItselfFromExile", s.Name)
		case s.CastsItselfFromExile && !prints:
			t.Errorf("%s sets CastsItselfFromExile and its printed text has no \"You may cast this card from exile.\" line: an exile cast another effect grants is a game.CastPermission (#659)", s.Name)
		}
		if prints {
			checked++
		}
	}
	t.Logf("%d catalogued faces print the clause", checked)
}

// The probe's line, pinned against the two texts it must tell apart: Khan,
// Vengeful Heart's clause, and rebound's reminder text.
func TestCastFromExileLineMatchesTheClauseAndNotRebound(t *testing.T) {
	khan := "When Khan enters, you may sacrifice another creature or artifact. If you do, draw two cards.\nYou may cast this card from exile."
	rebound := "Draw a card.\nRebound (If you cast this spell from your hand, exile it as it resolves. At the beginning of your next upkeep, you may cast this card from exile without paying its mana cost.)"
	if !castFromExileLine.MatchString(khan) {
		t.Error("the probe misses Khan's printed clause")
	}
	if castFromExileLine.MatchString(rebound) {
		t.Error("the probe matches rebound's reminder text")
	}
}

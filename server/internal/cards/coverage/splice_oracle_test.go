package coverage

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// TestSpliceMatchesThePrintedSplice holds every declared splice ability
// (ADR 0202, #1989) to the card's printed text: its quality and cost are
// the ones the card prints, so a typo in a splice cost cannot make a card
// cheaper — or pricier — than printed. And every catalogued card that
// prints splice declares it, so none ships as a card that can only be
// cast alone.
func TestSpliceMatchesThePrintedSplice(t *testing.T) {
	fx := loadOracleFixture(t)
	for _, spec := range effects.All() {
		base, _ := BaseOracleID(spec.OracleID)
		card, ok := fx[base]
		if !ok {
			continue
		}
		var text strings.Builder
		for _, f := range card.AllFaces() {
			text.WriteString(f.Text)
			text.WriteString("\n")
		}
		printed := strings.Contains(text.String(), "Splice onto ")
		switch {
		case spec.Splice == nil && printed:
			t.Errorf("%s prints splice and declares none", spec.Name)
		case spec.Splice != nil && !strings.Contains(text.String(), spec.Splice.Label()+" ("):
			t.Errorf("%s declares %q, which the card does not print", spec.Name, spec.Splice.Label())
		}
	}
}

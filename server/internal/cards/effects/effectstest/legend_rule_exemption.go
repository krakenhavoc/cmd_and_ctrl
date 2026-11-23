package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// legend_rule_exemption_test.go — "the legend rule doesn't apply"
// (CR 704.5j, #2177): Mirror Box, Mirror Gallery, Sakashima of a
// Thousand Faces.

const (
	OracleMirrorBox      = "3bed1944-58dc-4679-9aee-7be4d94fb55c"
	OracleMirrorGallery  = "7c9c3060-ca9d-4868-84c6-2a95b0fa8885"
	OracleSakashimaFaces = "8ecdaf4b-4442-42da-9714-4257a83faf50"
	LegendTypeLine       = "Legendary Creature — Human"
	LegendName           = "Test Legend"
	BothLegendsMustStay  = "both legends must stay"
)

func LegendPrompt(g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceLegendRule && c.Chooser == chooser {
			return c
		}
	}
	return nil
}

func OnBoard(g *game.Game, id uuid.UUID) bool {
	_, ok := FindBattlefieldByID(g, id)
	return ok
}

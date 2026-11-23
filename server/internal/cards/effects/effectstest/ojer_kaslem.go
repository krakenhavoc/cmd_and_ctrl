package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardso "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/o"
)

func OjerKaslemRow() cards.Card {
	return TransformRow(cardso.OjerKaslemOracleID,
		"Ojer Kaslem, Deepest Growth", "Legendary Creature — God", "{3}{G}{G}",
		"Temple of Cultivation", "Land", "", "", []string{"G"})
}

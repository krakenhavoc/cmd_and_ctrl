package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	cardsu "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/u"
)

func UltraMagnusRow() cards.Card {
	return TransformerRow(cardsu.UltraMagnusOracleID,
		DisturbFace{"Ultra Magnus, Tactician", "Legendary Artifact Creature — Robot", "{4}{R}{G}{W}", "7", "7"},
		DisturbFace{"Ultra Magnus, Armored Carrier", "Legendary Artifact — Vehicle", "", "4", "7"},
		"More Than Meets the Eye {2}{R}{G}{W}\nWard {2}\nWhenever Ultra Magnus attacks, you may put an artifact creature card from your hand onto the battlefield tapped and attacking. If you do, convert Ultra Magnus at end of combat.",
		"Living metal\nHaste\nFormidable — Whenever Ultra Magnus attacks, attacking creatures you control gain indestructible until end of turn. If those creatures have total power 8 or greater, convert Ultra Magnus.",
		"Living metal", "Convert", "Ward", "Haste")
}

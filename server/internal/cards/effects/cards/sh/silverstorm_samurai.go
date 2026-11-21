package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Silverstorm Samurai — Creature — Fox Samurai {4}{W}{W}, 3/3 (Betrayers
// of Kamigawa, #5602):
//
//	"Flash
//	 Bushido 1 (Whenever this creature blocks or becomes blocked, it gets
//	 +1/+1 until end of turn.)"
//
// Flash rides PrintedKeywords and bushido is the shared Bushido(1)
// trigger (CR 702.45a), through the Champions of Kamigawa builder.
//
// No simplification.
func init() {
	effects.ChkB1Bushido("3df0a392-bcc1-4456-bd6e-d044bcc83a16", "Silverstorm Samurai", 1, "flash")
}

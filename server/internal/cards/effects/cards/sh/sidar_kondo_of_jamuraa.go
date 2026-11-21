package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sidar Kondo of Jamuraa — Legendary Creature — Human Knight {2}{G}{W}, 2/5:
//
//	"Flanking
//	 Creatures your opponents control without flying or reach can't block
//	 creatures with power 2 or less.
//	 Partner"
//
// Flanking is the keyword token, which the engine triggers from
// (game/flanking.go, #5784); the row beside it is the retired pre-#5784
// trigger, kept for restore (RetiredFlankingRow). The static is a blocker-side block
// rule (CantBlockAttackers): blockers are your opponents' creatures with
// neither flying nor reach, read live through the layers; the attackers
// it protects are any creatures with power 2 or less, whoever controls
// them, so Sidar Kondo protects his own team and any other attacker
// alike. Partner is a deck-construction rule and needs nothing on the
// card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "331b6834-6bb4-40fa-b80b-22d67b34734f",
		Name:            "Sidar Kondo of Jamuraa",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordFlanking},
		Triggered:       []game.TriggeredAbility{effects.RetiredFlankingRow("Sidar Kondo of Jamuraa")},
		BlockRules: []game.BlockRule{
			effects.CantBlockAttackers(
				effects.OnMatching(effects.And(effects.Creature(), effects.OpponentControls(), effects.WithoutKeyword("flying"), effects.WithoutKeyword("reach"))),
				effects.OnMatching(effects.And(effects.Creature(), effects.PowerLE(2))),
				"creatures your opponents control without flying or reach can't block creatures with power 2 or less"),
		},
	})
}

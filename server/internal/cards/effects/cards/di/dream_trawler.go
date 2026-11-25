package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dream Trawler — Creature — Sphinx {2}{W}{W}{U}{U}, 3/5 (Theros Beyond Death,
// S82 OLD, #5514):
//
//	"Flying, lifelink
//	 Whenever you draw a card, this creature gets +1/+0 until end of turn.
//	 Whenever this creature attacks, draw a card.
//	 Discard a card: This creature gains hexproof until end of turn. Tap it."
//
// Burlfist Oak's draw watcher (one trigger per card drawn, so the attack
// draw pumps it too) and Prognostic Sphinx's discard-for-hexproof, whose
// second sentence taps the Trawler whether or not the hexproof landed. Flying
// and lifelink ride PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "661b2159-4040-4a44-b182-46085416ebd9",
		Name:            "Dream Trawler",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouDraw("Dream Trawler — this creature gets +1/+0 until end of turn",
				effects.ThisCreatureUntilEOT("Dream Trawler — +1/+0 until end of turn", 1, 0)),
			effects.WheneverThisAttacks("Dream Trawler — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Discard a card: This creature gains hexproof until end of turn. Tap it.",
			Purpose: game.Purpose{Answers: game.AnswerProtect},
			Cost:    effects.DiscardACard(),
			Effect:  effects.Cmdr2S22HexproofAndTapThis("Dream Trawler — hexproof until end of turn"),
		}},
	})
}

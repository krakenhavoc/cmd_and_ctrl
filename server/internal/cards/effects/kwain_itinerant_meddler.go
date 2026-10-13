package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kwain, Itinerant Meddler — Legendary Creature — Rabbit Wizard
// {W}{U}, 1/3 (EDHREC rank 2151):
//
//	"{T}: Each player may draw a card, then each player who drew a
//	 card this way gains 1 life."
//
// The group-hug Rabbit. A tap ability (summoning sickness applies, CR
// 302.6) whose body walks the table APNAP: each player draws, then
// each player who drew gains 1 — two passes, in the printed order, so
// every "whenever a player draws" watcher fires before any life
// changes.
//
// Each player is asked in APNAP order whether to draw (MayChoice), and
// the life is gained after the last answer by exactly the players who
// drew.
func init() {
	Register(Spec{
		OracleID:     "ce6a85b5-e249-4582-a6be-ce200da5fd53",
		Name:         "Kwain, Itinerant Meddler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Each player may draw a card, then each player who drew a card this way gains 1 life",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect:  b20EachPlayerDrawsAndGainsOne,
		}},
	})
}

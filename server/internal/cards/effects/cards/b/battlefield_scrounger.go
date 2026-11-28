package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Battlefield Scrounger — Creature — Centaur {3}{G}{G}, 3/3 (Judgment,
// #5611):
//
//	"Threshold — Put three cards from your graveyard on the bottom of your
//	 library: This creature gets +3/+3 until end of turn. Activate only once
//	 each turn and only if there are seven or more cards in your graveyard."
//
// The cost is the #6121 graveyard-to-bottom component; the once-a-turn limit is
// the announce tally and the threshold an activation condition, both checked
// before the cost is paid (CR 602.2).
//
// No simplification.
func init() {
	const label = "Threshold — Put three cards from your graveyard on the bottom of your library: This creature gets +3/+3 until end of turn. Activate only once each turn and only if there are seven or more cards in your graveyard."
	effects.Register(effects.Spec{
		OracleID:     "92ac7531-10bb-47a2-9dd0-378d547b4ed6",
		Name:         "Battlefield Scrounger",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 3, Toughness: 3}},
			Cost:      effects.PutFromGraveyardOnLibraryBottom(3, "three cards", nil),
			Condition: effects.AllConditions(effects.GraveyardAtLeast(7, nil), effects.OncePerTurnActivation(label)),
			Effect:    effects.ThisGetsUntilEndOfTurn(3, 3, "Battlefield Scrounger — +3/+3"),
		}},
	})
}

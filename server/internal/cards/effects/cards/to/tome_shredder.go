package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tome Shredder — Creature — Wolf {2}{R}, 2/2 (EDHREC rank 25532):
//
//	"Haste
//	 {T}, Exile an instant or sorcery card from your graveyard: Put a
//	 +1/+1 counter on this creature."
//
// The filtered form of #1297's graveyard exile cost: only an instant or
// sorcery card in your own graveyard pays, through the one candidate
// walk every reader shares. Haste is printed, so the {T} works the turn
// it arrives.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b145952b-52e3-4a66-b47e-f08a489f9443",
		Name:            "Tome Shredder",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Exile an instant or sorcery card from your graveyard: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.Plus(effects.TapCost(), effects.ExileFromGraveyard(1, "an instant or sorcery card", effects.MatchInstantOrSorcery)),
			Effect:  effects.B35PutCounterOnSelf,
		}},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glint-Eye Nephilim — Creature — Nephilim {U}{B}{R}{G}, 2/2
// (Commander deck reprint):
//
//	"Whenever this creature deals combat damage to a player, draw that
//	 many cards.
//	 {1}, Discard a card: This creature gets +1/+1 until end of turn."
//
// The draw reads the amount off the damage event (Body of Knowledge's
// "draw that many cards"), so it is the damage actually dealt. The pump
// is Carrion Ants' self pump with a discard added to its cost, so it can
// be activated any number of times and summoning-sick.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6b6d965-1401-4566-adb3-90b03218af8d",
		Name:         "Glint-Eye Nephilim",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Glint-Eye Nephilim — draw that many cards", cmdr2B08DrawThatMany),
		},
		Activated: []ActivatedAbility{{
			Label:   "{1}, Discard a card: This creature gets +1/+1 until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump, Pump: &game.Pump{Power: 1, Toughness: 1}},
			Cost:    Plus(ManaCost("{1}"), DiscardACard()),
			Effect:  thisGetsUntilEndOfTurn(1, 1, "Glint-Eye Nephilim — +1/+1 until end of turn"),
		}},
	})
}

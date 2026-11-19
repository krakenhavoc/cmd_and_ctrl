package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lavinia, Foil to Conspiracy — Legendary Creature — Human Detective
// {1}{W/U}{W/U}, 2/3 (Ravnica: Clue Edition, #5929):
//
//	"Vigilance
//	 Whenever you cast your second spell each turn, investigate.
//	 {T}: Add {C}{C}. Activate only during an opponent's turn."
//
// The trigger is Brimstone Roundup's second-spell read
// (YouCastYourSecondSpellEachTurn, the per-turn cast tally) creating a Clue
// (CR 701.16a). The mana ability is a tap ability (so summoning sickness
// applies) gated to turns other than the controller's (Nettling Imp's
// DuringAnOpponentsTurn).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b5b26cea-9be3-46f7-9603-a6eee0dd107e",
		Name:            "Lavinia, Foil to Conspiracy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, YouCastYourSecondSpellEachTurn,
				"Lavinia, Foil to Conspiracy — investigate (second spell this turn)", Do(Investigate{N: 1})),
		},
		ManaAbilities: []ManaAbility{{
			Cost:      ManaAbilityCost{Tap: true},
			Produced:  "{C}{C}",
			Label:     "{T}: Add {C}{C}. Activate only during an opponent's turn.",
			Condition: DuringAnOpponentsTurn(),
		}},
	})
}

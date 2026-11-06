package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foggy Swamp Vinebender — Creature — Human Plant Ally {3}{G}, 4/3
// (Avatar: The Last Airbender, #4459):
//
//	"This creature can't be blocked by creatures with power 2 or less.
//	 Waterbend {5}: Put a +1/+1 counter on this creature. Activate only
//	 during your turn."
//
// The evasion is Legolas Greenleaf's CantBeBlockedBy, judged against
// each blocker's power as blockers are declared (CR 509.1b).
// WaterbendCost("{5}") is the whole cost (CR 701.67a): each generic mana
// may be paid by tapping an untapped artifact or creature you control,
// this one included. "Activate only during your turn" is the activation
// condition; the counter goes on the creature if it is still there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "82fefb82-a436-488d-8ecd-594686117e34",
		Name:         "Foggy Swamp Vinebender",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
		Activated: []ActivatedAbility{{
			Label:     "Waterbend {5}: Put a +1/+1 counter on this creature. Activate only during your turn.",
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      WaterbendCost("{5}"),
			Condition: DuringYourTurn(),
			Effect:    Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}),
		}},
	})
}

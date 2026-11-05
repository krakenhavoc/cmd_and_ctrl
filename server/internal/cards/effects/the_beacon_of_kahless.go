package effects

import (
	"strconv"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Beacon of Kahless — Legendary Artifact {2}{R} (Star Trek Commander,
// #4974):
//
//	"{T}: Add one mana of any color.
//	 Battlefront — {T}: Choose a color. Add an amount of mana of that
//	 color equal to the number of opponents you attacked this turn."
//
// Battlefront is an ability word. The second ability is Nyx Lotus's
// shape: one colour pick whose amount is X for every colour ("{W2|U2|…}"
// is "two of one colour"), X counted as it is activated
// (OpponentsYouAttackedThisTurn): each opponent still in the game that a
// creature of yours was declared attacking this turn counts once,
// whichever combat; an attack on a planeswalker or a battle is not an
// attack on its player. Only the active player attacks (CR 508.1), so
// off its controller's turn, or before an attack, X is 0 and the ability
// adds nothing, though it can still be activated (it still taps).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "66fc3c1e-acea-46eb-8af8-b0740f006add",
		Name:         "The Beacon of Kahless",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{
			{Cost: ManaAbilityCost{Tap: true}, Produced: "{W|U|B|R|G}", Label: "Add one mana of any color"},
			{
				Cost:         ManaAbilityCost{Tap: true},
				ProducedFunc: beaconOfKahlessProduced,
				Label:        "Battlefront — choose a color: add mana of that color equal to the number of opponents you attacked this turn",
			},
		},
	})
}

func beaconOfKahlessProduced(g *game.Game, controller, _ uuid.UUID) string {
	x := OpponentsYouAttackedThisTurn(g, controller)
	if x <= 0 {
		return ""
	}
	parts := make([]string, 0, len(game.AllColors))
	for _, c := range game.AllColors {
		parts = append(parts, c+strconv.Itoa(x))
	}
	return "{" + strings.Join(parts, "|") + "}"
}

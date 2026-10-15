package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Opal Palace — Land:
//
//	"{T}: Add {C}.
//	 {1}, {T}: Add one mana of any color in your commander's color
//	 identity. If you spend this mana to cast your commander, it
//	 enters with a number of additional +1/+1 counters on it equal to
//	 the number of times it's been cast from the command zone this
//	 game."
//
// Both mana abilities are real: the colourless tap, and the {1}, {T}
// filter ability that narrows to the controller's commander's colour
// identity exactly as Arcane Signet does (NarrowToCommanderIdentity),
// with the Signet cycle's {1} mana cost added on top of the tap.
//
// The bonus is a spend rider whose number is counted as the commander
// ENTERS (ADR 0109 §11 decision 5, #1552): a counted enters-with-
// counters rider, its count registered under the card's name. "Your
// commander" is the rider's condition, checked at the spend against the
// spell the mana paid for; the count is that commander's command-zone
// cast tally (CR 903.8's commander tax counter), which includes the
// cast being paid for when it came from the command zone. Cast from
// anywhere else, the commander still gets the counters for its earlier
// command-zone casts, as printed.
//
// No simplification is left (#3844, ADR 0068 amendment 2026-10-15). A
// waived payment (strict mana off with too little in the pool, or Cast
// anyway) still spends the share of the cost the pool can pay, so this
// mana pays whenever it can and its rider fires. Mana that never
// entered the pool was not this card's mana.
func init() {
	Register(Spec{
		OracleID:       "aa6723a2-75da-49f5-a1ba-cbfa82c55301",
		Name:           "Opal Palace",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   CompletenessFull,
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:                      ManaAbilityCost{Tap: true, Mana: "{1}"},
				Produced:                  "{W|U|B|R|G}",
				Label:                     "{1}, {T}: Add one mana of any color in your commander's color identity",
				NarrowToCommanderIdentity: true,
				SpendRiders: []game.ManaSpendRider{
					SpentEntersWithCountersCounted(game.CounterPlusOne, "Opal Palace", game.ManaRiderCount{
						Condition: opalPalaceYourCommander,
						Count:     opalPalaceCommandZoneCasts,
					}, ManaRestrictCast),
				},
			},
		},
	})
}

// opalPalaceYourCommander is "to cast your commander": the spell is a
// commander its caster owns.
func opalPalaceYourCommander(_ *game.Game, controller uuid.UUID, spell game.Card) bool {
	return spell.IsCommander && spell.Owner == controller
}

// opalPalaceCommandZoneCasts is "the number of times it's been cast
// from the command zone this game" — the commander tax tally.
func opalPalaceCommandZoneCasts(g *game.Game, controller uuid.UUID, entering game.Card) int {
	p := g.PlayerByIDForEffect(controller)
	if p == nil {
		return 0
	}
	return p.CommanderCasts[entering.InstanceID]
}

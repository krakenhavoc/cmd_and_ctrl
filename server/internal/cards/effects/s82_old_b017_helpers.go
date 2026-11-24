package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// SnowBasicSpec is a Basic Snow Land's spec. It has no abilities of its own:
// the intrinsic mana ability of its basic land type is derived from the type
// line (CR 305.6), and "Snow" is the supertype the type line prints
// (CR 205.4g). The Spec exists so the card is a catalogued, Full card.
func SnowBasicSpec(oracle, name string) Spec {
	return Spec{
		OracleID:     oracle,
		Name:         name,
		Completeness: CompletenessFull,
	}
}

// HideawayLandSpec is a hideaway land of the Lorwyn shape: "Hideaway 4. This
// land enters tapped. {T}: Add <color>. <color>, {T}: You may play the exiled
// card without paying its mana cost if <clause>." It is the table row of
// hideaway_lands.go as a function, for a land that lives in its own file.
// The condition is asked as the ability resolves (see hideaway_lands.go).
func HideawayLandSpec(oracle, name, color, clause string, condition func(g *game.Game, controller uuid.UUID) bool) Spec {
	mana := "{" + color + "}"
	return Spec{
		OracleID:     oracle,
		Name:         name,
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Triggered:    []game.TriggeredAbility{Hideaway(name, 4)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: mana,
			Label:    "Add " + mana,
		}},
		Activated: []ActivatedAbility{{
			Label:   mana + ", {T}: You may play the exiled card without paying its mana cost " + clause + ".",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost(mana), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !condition(g, item.Controller) {
					return nil
				}
				return PlayHiddenCard{
					Source: HiddenRefOfActivation(item),
					Label:  name + " — play the exiled card without paying its mana cost",
				}.Apply(NewContext(g, item))
			},
		}},
	}
}

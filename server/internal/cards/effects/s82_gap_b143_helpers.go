package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b143_helpers.go — the Dominaria Remastered Lair land (#5948):
// "When this land enters, sacrifice it unless you return a non-Lair land you
// control to its owner's hand. {T}: Add <three colours>."

// gapB143NonLairLand is the land the Lairs' payment can return: any land you
// control that is not a Lair.
func GapB143NonLairLand(_ *game.Card, land game.Card) bool {
	return !land.HasSubtype("Lair")
}

// gapB143LairSpec is one Lair land: the enters trigger is the bounce lands' keyed
// return (returnLandsOrSacrifice), whose prompt and tail keys are the card's
// own on-disk identities, and the mana ability taps for any one of its
// three colours.
func GapB143LairSpec(oracleID, name, colors string, ret *returnLandsOrSacrifice) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{ReturnsLandOnEntry: true},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{" + colors + "}",
			Label:    "Add " + gapB143LairColorWords(colors),
		}},
		Triggered: []game.TriggeredAbility{
			ret.Trigger(name + " — sacrifice it unless you return a non-Lair land you control to its owner's hand"),
		},
	}
}

func gapB143LairColorWords(colors string) string {
	c := strings.Split(colors, "|")
	return "{" + c[0] + "}, {" + c[1] + "}, or {" + c[2] + "}"
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ancestor's Prophet — Creature — Human Cleric {4}{W}, 1/5 (Onslaught):
//
//	"Tap five untapped Clerics you control: You gain 10 life."
//
// A tap-five cost over Clerics (tapFiveUntapped). The Prophet is a
// Cleric, so it may be one of the five, and with no {T} symbol a Cleric
// that arrived this turn may pay it (CR 302.6). The row has no target,
// so it declares what it answers: nothing (ADR 0142).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "274ce6b0-af29-4afc-873b-b4126a67bd81",
		Name:         "Ancestor's Prophet",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Tap five untapped Clerics you control: You gain 10 life.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    tapFiveUntapped("five untapped Clerics you control", "Cleric"),
			Effect:  Do(GainLife{Amount: 10}),
		}},
	})
}

// tapFiveUntapped is "Tap five untapped <creature type>s you control" as
// a cost, the source included when it has the type: tapTwoUntapped's
// shape with five (Ancestor's Prophet's Clerics, Catapult Master's
// Soldiers).
func tapFiveUntapped(label, creatureType string) game.AbilityCost {
	return game.AbilityCost{TapOthers: &game.TapOthersCost{
		Count:  5,
		Filter: TargetPermanent(label, Creature(), OfCreatureType(creatureType)),
		Label:  label,
	}}
}

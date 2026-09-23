package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// morph.go — #1194 / ADR 0082: morph (CR 702.37), megamorph
// (CR 702.109) and disguise (CR 702.168).
//
// Each of the three is printed as TWO permissions on one line, and a
// card file should spell out only the half that varies:
//
//	"Morph {1}{U}"  =  you may cast this face down as a 2/2 for {3}
//	                +  you may turn it face up for {1}{U}
//
// The {3} is the KEYWORD's (identical on every card that prints any of
// the three) and lives here; the face-up cost is the card's and is the
// only argument. The same division effects.Foretell already draws
// between foretell's fixed {2} and the card's own foretell cost.
//
// Nothing else belongs in the card file either. The face-down object,
// its 2/2 body, its missing text, the controller-only look, the
// turn-face-up special action, megamorph's counter and CR 708.9's
// reveal on the way out are the keyword's, and the engine carries all
// seven.

// MorphCastCost is what casting a card face down costs — {3} on every
// card that prints morph (CR 702.37b), megamorph (CR 702.109a) or
// disguise (CR 702.168a). The card's own printed cost is what it
// costs to turn the permanent back FACE UP, which is the argument
// each constructor takes.
const MorphCastCost = "{3}"

// Morph is "Morph <cost>" — CR 702.37a. `cost` is the printed MORPH
// COST, what the controller pays to turn the permanent face up, not
// what the face-down cast costs.
//
// Give it to the card the way its oracle text reads:
//
//	AlternativeCosts: []game.AlternativeCost{Morph("{1}{U}")},
//
// A creature with morph and nothing else is otherwise a plain Spec:
// the face-down body is a rules projection, not a second card file.
func Morph(cost string) game.AlternativeCost {
	return faceDownCast("morph", "Morph", game.FaceDownMorphed, cost, false)
}

// Megamorph is "Megamorph <cost>" — CR 702.109a. Morph, plus a +1/+1
// counter when the permanent is turned face up (CR 702.109b).
func Megamorph(cost string) game.AlternativeCost {
	return faceDownCast("megamorph", "Megamorph", game.FaceDownMorphed, cost, true)
}

// Disguise is "Disguise <cost>" — CR 702.168a. Morph, but the
// face-down permanent has ward {2} (CR 702.168b), which rides the
// KIND rather than this declaration: FaceDownKind.HasWard is the
// predicate and the engine hangs the ability off it, so a card file
// never writes the ward down.
func Disguise(cost string) game.AlternativeCost {
	return faceDownCast("disguise", "Disguise", game.FaceDownDisguised, cost, false)
}

// faceDownCast is the shared body of the three. One function because
// the three keywords differ in exactly three values — the wire key,
// the printed word, and which of (face-down kind, counter) they carry
// — and a copy each would be three places for the {3} to drift.
func faceDownCast(key, printed string, kind game.FaceDownKind, faceUpCost string, counter bool) game.AlternativeCost {
	return game.AlternativeCost{
		Key: key,
		// The label is the picker row, so it says what the player is
		// about to DO as well as which keyword they are using: the
		// cost in it is the {3} being paid now, never the face-up
		// cost printed on the card.
		Label:    printed + " — cast face down " + MorphCastCost,
		ManaCost: MorphCastCost,
		FaceDown: &game.FaceDownCast{
			Kind:          kind,
			FaceUpCost:    faceUpCost,
			FaceUpCounter: counter,
		},
	}
}

// WhenThisIsTurnedFaceUp — "When this permanent is turned face up,
// ..." (CR 708.8). Willbender, Kadena's Silencer, the whole
// morph-trigger family.
//
// An ordinary triggered ability watching an ordinary event kind, on
// the battlefield, with the default zone list — and that is the point
// of EventTurnedFaceUp existing. The ability is on the permanent,
// which has no text at all while it is face down (CR 708.2a), so the
// trigger is only harvestable in the instant AFTER the state is
// cleared. turnFaceUpLocked clears first and emits second for exactly
// that reason (ADR 0082 decision 7).
func WhenThisIsTurnedFaceUp(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventTurnedFaceUp, Self, label, effect)
}

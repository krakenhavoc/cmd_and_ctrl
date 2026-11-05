package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Terrasymbiosis — Enchantment {2}{G} (EDHREC rank 769, #3654):
//
//	"Whenever you put one or more +1/+1 counters on a creature you
//	 control, you may draw that many cards. Do this only once each turn."
//
// The trigger is WheneverYouPutCountersOfKindOnACreature (#3929): once
// per creature per placement (CR 603.2c), so Cathars' Crusade's counter
// on each of five creatures is five triggers, and +1/+1 counters only. A
// placement whose recorded placer is an opponent is not "you" (every
// placement records its placer, #3942).
//
// "That many" is the number of +1/+1 counters that placement put on the
// creature, after replacements (a Hardened Scales'd single counter is
// two), read as the trigger resolves (CountersOfKindYouPutInTheSameBatch).
//
// "Do this only once each turn" limits the draw, not the trigger, as
// Nykthos Paragon's does: every placement triggers it and asks its "you
// may" as it resolves (CR 603.5); the first "yes" draws and uses up the
// turn, and a "no" does not. The record is per object
// (NoteDoneThisTurn), and the gate is also read as the ability triggers,
// so no trigger is made once the turn's draw is done. The body is
// onceEachTurnMayDrawThatMany (trek_b46_helpers.go), shared with Chekov
// and McCoy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4ac94586-3ac7-4ade-af31-ad0d0e3ffc4a",
		Name:         "Terrasymbiosis",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{onceEachTurnMayDrawThatMany(creatureYouControl,
			terrasymbiosisLabel, terrasymbiosisDone, "Terrasymbiosis — draw that many cards?")},
	})
}

// terrasymbiosisLabel is the trigger's stack label.
const terrasymbiosisLabel = "Terrasymbiosis — draw that many cards"

// terrasymbiosisDone is the once-each-turn record's label: the cards
// were drawn this turn (b28DoneThisTurn).
const terrasymbiosisDone = "Terrasymbiosis — cards drawn this turn"

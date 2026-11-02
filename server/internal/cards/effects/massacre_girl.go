package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Massacre Girl — Legendary Creature — Human Assassin {3}{B}{B}, 4/4:
//
//	"Menace
//	 When Massacre Girl enters, each other creature gets -1/-1 until end
//	 of turn. Whenever a creature dies this turn, each creature other than
//	 Massacre Girl gets -1/-1 until end of turn."
//
// One printed ability, one registered trigger. Its resolution does the
// first sentence and then creates a CR 603.7b delayed triggered ability
// for the second ("this turn" is its duration): a repeating one that
// fires once for each creature that dies from then until cleanup, tokens
// included, Massacre Girl herself included. Each firing is a separate
// stack item, so a chain of deaths is a chain of sweeps.
//
// Both sweeps are read when they resolve (CR 611.2c): a creature that
// enters later in the turn is not shrunk by one that already resolved,
// and "each creature other than Massacre Girl" is that one object (a
// copy of her under another controller is shrunk like any other
// creature). The Massacre Girl the delayed trigger names is the source
// the sweep was created by, so it keeps excluding her after she leaves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "63e6cb7d-cc27-4200-85fc-ff6472318c1a",
		Name:            "Massacre Girl",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Purpose:         game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 1}},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Massacre Girl — each other creature gets -1/-1 until end of turn; whenever a creature dies this turn, each creature other than Massacre Girl gets -1/-1",
				massacreGirlEnters),
		},
	})
}

// massacreGirlSweep is "-1/-1 to each creature other than Massacre Girl
// until end of turn", for the enter sentence and for each death after it.
func massacreGirlSweep(g *game.Game, item *game.StackItem) error {
	return BoostUntilEOT{
		Match:     And(Creature(), OtherThan(item.SourceCardID)),
		Power:     -1,
		Toughness: -1,
		Label:     "Massacre Girl — -1/-1",
	}.Apply(NewContext(g, item))
}

// massacreGirlEnters is the enter trigger's resolution: the first sweep,
// then the repeating delayed trigger on a creature dying.
func massacreGirlEnters(g *game.Game, item *game.StackItem) error {
	if err := massacreGirlSweep(g, item); err != nil {
		return err
	}
	return DelayedOnEvent{
		Label:     "Massacre Girl — each creature other than Massacre Girl gets -1/-1 until end of turn",
		On:        []game.EventKind{game.EventLTB},
		Condition: aCreatureDiedCondition,
		Body:      massacreGirlSweepBody,
		Repeats:   true,
	}.Apply(NewContext(g, item))
}

var (
	// "Whenever a creature dies": an EventLTB into a graveyard of something
	// that was a creature as it last existed (CR 603.10a), anyone's.
	aCreatureDiedCondition = game.DelayedCondition("dies/a-creature", aCreatureDied)

	// Massacre Girl's repeated sweep.
	massacreGirlSweepBody = game.SimpleDelayedBody("massacre-girl/sweep", massacreGirlSweep)
)

func aCreatureDied(ev game.Event, _ *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
	_, ok := diedCreature(ev, g)
	return ok
}

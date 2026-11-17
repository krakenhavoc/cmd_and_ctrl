package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Massacre — Sorcery {2}{B}{B} (Nemesis, #5618):
//
//	"If an opponent controls a Plains and you control a Swamp, you may
//	 cast this spell without paying its mana cost.
//	 All creatures get -2/-2 until end of turn."
//
// Sivvi's Ruse's free cast (the Nemesis cycle's condition: an
// OPPONENT's Plains and your Swamp) on Hideous Laughter's sweep. The
// creatures are the ones on the battlefield as it resolves (CR 611.2c),
// so one that enters later is unaffected; the shrink ends in the
// cleanup step (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ce1866e9-28de-4bcc-9999-11958e6c13a6",
		Name:         "Massacre",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepMinus, Amount: 2}},
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "free",
			Label:     "Cast without paying its mana cost (an opponent controls a Plains and you control a Swamp)",
			ManaCost:  "",
			Condition: AnOpponentControlsAAndYouControlA("Plains", "Swamp"),
		}},
		OnResolve: ResolveSteps(BoostUntilEOT{Match: Creature(), Power: -2, Toughness: -2, Label: "Massacre — -2/-2"}),
	})
}

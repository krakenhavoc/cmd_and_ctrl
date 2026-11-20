package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Catharsis — Creature — Elemental Incarnation {4}{R/W}{R/W}, 3/4
// (Lorwyn Eclipsed, #4458):
//
//	"When this creature enters, if {W}{W} was spent to cast it, create
//	 two 1/1 green and white Kithkin creature tokens.
//	 When this creature enters, if {R}{R} was spent to cast it,
//	 creatures you control get +1/+1 and gain haste until end of turn.
//	 Evoke {R/W}{R/W} (You may cast this spell for its evoke cost. If
//	 you do, it's sacrificed when it enters.)"
//
// Deceit's cycle-mate, on the same reading: two enters triggers, each
// behind an intervening if (CR 603.4) that reads the mana that paid for
// the spell this permanent was (deceitSpent: at least two of that
// colour, hybrid symbols or not). The evoke cost counts, so an evoke
// paid {W}{W} still makes the Kithkin before the CR 702.74a sacrifice,
// and one paid {R}{W} makes nothing. With {W}{W}{R}{R} spent both
// trigger and their controller orders them; if the Kithkin are made
// first, they get the +1/+1 and haste too.
//
// The pump's set of creatures is fixed as it resolves (CR 611.2c), so a
// creature that enters later in the turn gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "9e3303aa-2484-4832-86ac-f5ae290d02f1",
		Name:           "Catharsis",
		ReadsManaSpent: game.ReadsOwnPayment,
		Completeness:   effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{
			effects.Evoke("{R/W}{R/W}"),
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.DeceitSpent("W")),
				"Catharsis — create two 1/1 green and white Kithkin creature tokens",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green and white Kithkin"), N: 2})),
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.DeceitSpent("R")),
				"Catharsis — creatures you control get +1/+1 and gain haste until end of turn",
				effects.Do(
					effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Toughness: 1, Label: "Catharsis — +1/+1"},
					effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"haste"}, Label: "Catharsis — haste"},
				)),
		},
	})
}

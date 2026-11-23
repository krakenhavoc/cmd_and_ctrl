package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dark Apostle — Creature — Astartes Warlock {3}{R}, 3/3 (Warhammer 40,000
// Commander, S82 GAP batch 194, #5948):
//
//	"Gift of Chaos — {3}, {T}: The next noncreature spell you cast this turn
//	 has cascade. (When you cast that spell, exile cards from the top of your
//	 library until you exile a nonland card that costs less. You may cast it
//	 without paying its mana cost. Put the exiled cards on the bottom in a
//	 random order.)"
//
// A delayed trigger on the next cast of a noncreature spell by you this turn
// (Ride the Avalanche's shape): a creature spell cast first neither fires it
// nor uses it up, and it ends with the turn. Cascade is a cast trigger, so the
// delayed trigger that fires as the spell is cast, above it on the stack,
// cascades from the spell's mana value exactly as the keyword would. The
// ability's {T} is the creature's own, so it is summoning-sick the turn it
// arrives (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8a446307-d9ff-4854-90e6-9cdff5b77aa6",
		Name:         "Dark Apostle",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: The next noncreature spell you cast this turn has cascade.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect:  effects.GapB194GrantNextNoncreatureSpellCascade,
		}},
	})
}

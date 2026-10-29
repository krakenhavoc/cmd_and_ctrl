package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rhovanion Rampager — Creature — Wolf {2}{B}, 3/2 (The Hobbit, #4454):
//
//	"Whenever this creature attacks, you may sacrifice another creature.
//	 If you do, put a number of +1/+1 counters on this creature equal to
//	 the sacrificed creature's power.
//	 When this creature dies, amass Goblins X, where X is this
//	 creature's power."
//
// The sacrifice is asked as the trigger resolves (maySacrificeOne), over
// whatever else the controller controls then, and the counters follow only
// a creature that really went. Its power is last-known information
// (CR 608.2h, departedCreaturePower), so an anthem's bonus counts. The
// counters go on the Rampager only while it is still on the battlefield.
//
// The death trigger reads the Rampager's own power as it last stood
// (counters and anthems included), the same way, and amasses Goblins by it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "008a11c1-d283-49fe-abd7-ff4fe8b1fe79",
		Name:         "Rhovanion Rampager",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Rhovanion Rampager — sacrifice another creature for +1/+1 counters", rhovanionRampagerSacrifice),
			WhenThisDies("Rhovanion Rampager — amass Goblins X", rhovanionRampagerAmass),
		},
	})
}

// rhovanionRampagerSacrifice is the attack trigger: "you may sacrifice
// another creature".
func rhovanionRampagerSacrifice(g *game.Game, item *game.StackItem) error {
	return maySacrificeOne(g, item, "Rhovanion Rampager — sacrifice another creature to put +1/+1 counters on it equal to its power?",
		Creature(), true, rhovanionRampagerSacrificed)
}

// rhovanionRampagerSacrificed is the "if you do": counters equal to the
// sacrificed creature's power, keyed (#4290).
var rhovanionRampagerSacrificed = RegisterSacrificedOne("prompt-run/rhovanion-rampager-sacrifice", func(ctx *Context, sacrificed uuid.UUID) error {
	n := departedCreaturePower(ctx.Game, sacrificed)
	if n == 0 || !onBattlefield(ctx.Game, ctx.Source()) {
		return nil
	}
	return AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: n}.Apply(ctx)
})

// rhovanionRampagerAmass is the death trigger: amass Goblins X, X being the
// Rampager's power as it last existed.
func rhovanionRampagerAmass(g *game.Game, item *game.StackItem) error {
	return Amass{Subtype: "Goblin", N: departedCreaturePower(g, item.SourceCardID)}.Apply(NewContext(g, item))
}

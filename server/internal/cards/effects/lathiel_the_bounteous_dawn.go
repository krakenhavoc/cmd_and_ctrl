package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lathiel, the Bounteous Dawn — Legendary Creature — Unicorn
// {2}{G}{W}, 2/2 (EDHREC rank 3472):
//
//	"Lifelink
//	 At the beginning of each end step, if you gained life this turn,
//	 distribute up to that many +1/+1 counters among any number of
//	 other target creatures."
//
// The lifegain deck's counters engine: every point of life gained
// becomes a counter at the end of the turn — any player's end step,
// as printed. The intervening if is b15LifeGainedThisTurn, checked
// at announce and again at resolution (CR 603.4); the targets are
// "any number of other target creatures", Appa, Steadfast Guardian's
// Min-0 unbounded clause with Lathiel kept out by name; and the
// counters are placed on the announced creatures that are still
// legal when the trigger resolves.
//
// Sandbox simplification, declared: "distribute" is a split of the
// controller's choosing, and the engine has no prompt for one. The
// counters are dealt out one at a time around the chosen creatures
// in the order they were picked — one creature gets them all, three
// creatures share them as evenly as the count allows, with the first
// picked getting the remainder. Every split is one the printed card
// allows and the total is never more than the life gained; "up to"
// is answered by choosing fewer creatures, or none. Weaker than
// printed — an uneven split cannot be asked for — never stronger.
//
// With no other creature on the battlefield the trigger is dropped
// before the prompt (CR 603.3d); the printed card would put it on
// the stack to do nothing.
func init() {
	Register(Spec{
		OracleID:        "d3c56fc4-3611-41b1-952e-4c5311b1510b",
		Name:            "Lathiel, the Bounteous Dawn",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"The counters are dealt out one at a time around the creatures you chose, in the order you picked them — you can't ask for an uneven split."},
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginEndStep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b33EndStepAndYouGainedLifeThisTurn(ev, source, g)
			},
			Targets: TargetCreature("any number of other target creatures",
				b03NotNamed("Lathiel, the Bounteous Dawn")).WithCount(0, 0),
			Key:    "Lathiel, the Bounteous Dawn — distribute +1/+1 counters among the chosen creatures",
			Effect: b33DistributeCountersRoundRobin,
		}},
	})
}

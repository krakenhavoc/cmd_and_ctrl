package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shalai and Hallar — Legendary Creature — Angel Elf {1}{R}{G}{W},
// 3/3 (EDHREC rank 4065):
//
//	"Flying, vigilance
//	 Whenever one or more +1/+1 counters are put on a creature you
//	 control, Shalai and Hallar deals that much damage to target
//	 opponent."
//
// A Naya counters commander that converts the whole +1/+1 theme into
// reach. A Cathars' Crusade trigger, a Hardened Scales-doubled
// Duskshell Crawler, an evolve creature growing — each is a separate
// batch of counters and each is a separate ping at a player who
// thought they were safe behind blockers. In a four-player pod the
// target is chosen per trigger, so the damage can be spread or piled.
//
// "ONE OR MORE … ON A CREATURE YOU CONTROL" IS ONE TRIGGER PER
// CREATURE TOUCHED, not one per counter and not one per spell: the
// engine emits one counter-placed event per permanent, and this
// ability fires once for each with the batch's size as the damage.
// That is why it is NOT wrapped in OncePerBatch — a Cathars' Crusade
// resolving over five creatures is five triggers on the printed
// card, and collapsing them would be weaker than printed.
//
// The counters must be +1/+1 and they must land on a CREATURE the
// controller controls: a loyalty counter, a charge counter, or a
// +1/+1 counter on a noncreature artifact does nothing.
//
// SANDBOX SIMPLIFICATION, DECLARED. The printed text does not care
// WHO puts the counters — an opponent's Ivy Lane Denizen growing your
// creature would trigger it. The engine attributes counter placement
// through b12CountersPlacedByYou (All Will Be One's read), which
// fires only when the resolving spell or ability belongs to the
// source's controller, and which errs toward not firing whenever the
// event log cannot attribute the placer at all. So counters an
// opponent puts on your creature are silent here. Weaker than
// printed, never stronger.
//
// "Target opponent" is re-checked at resolution (CR 608.2b); an
// opponent who has left the game in response takes nothing.
func init() {
	Register(Spec{
		OracleID:     "e7604cd9-d00d-4957-82c9-46a7cdb88209",
		Name:         "Shalai and Hallar",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"Only +1/+1 counters you put on your own creatures trigger it — counters an opponent's effect puts on your creatures don't.",
		},
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCounterPlaced},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b39PlusOneCountersOnCreatureYouControl(ev, source, g)
				return ok
			},
			Targets: TargetPlayer("target opponent", Opponent()),
			Key:     "Shalai and Hallar — that much damage to target opponent",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				amount, _ := b39PlusOneCountersOnCreatureYouControl(ev, source, g)
				item := game.NewTriggeredItem(source, "Shalai and Hallar — that much damage to target opponent", nil)
				item.Params.Amount = amount
				return item
			},
			Effect: b39DamageToFirstTargetPlayer,
		}},
	})
}

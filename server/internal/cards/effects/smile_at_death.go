package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smile at Death — Enchantment {3}{W}{W} (EDHREC rank 3683):
//
//	"At the beginning of your upkeep, return up to two target
//	 creature cards with power 2 or less from your graveyard to the
//	 battlefield. Put a +1/+1 counter on each of those creatures."
//
// Reveillark every upkeep. The trigger is targeted at "up to two"
// creature cards in the controller's graveyard with power 2 or less
// — Reveillark's clause, read at announce and again at resolution
// (CR 608.2b) — and each card still there returns to the battlefield
// under its owner's control ("from YOUR graveyard", so the
// controller) and then gets its +1/+1 counter. With no legal card in
// the graveyard the trigger is removed (CR 603.3d) rather than
// prompting. The counter lands a beat after the entry rather than as
// part of it, Rakdos Joins Up's posture: nothing in the catalog
// reads a creature's counters between its arrival and the next
// event, and a 2-power creature that returns still passed the
// clause when it was chosen.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9228f04c-506f-4453-a335-e66876b8ce8d",
		Name:         "Smile at Death",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBeginUpkeep},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Targets: TargetCardInGraveyard("up to two target creature cards with power 2 or less from your graveyard",
				YouOwn(), Creature(), PowerLE(2)).WithCount(0, 2),
			Key:    "Smile at Death — return up to two small creature cards to the battlefield with a +1/+1 counter",
			Effect: b35ReturnChosenToBattlefieldWithCounter,
		}},
	})
}

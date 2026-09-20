package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Enduring Tenacity — Enchantment Creature — Snake Glimmer {2}{B}{B},
// 4/3:
//
//	"Whenever you gain life, target opponent loses that much life.
//	 When Enduring Tenacity dies, if it was a creature, return it to
//	 the battlefield under its owner's control. It's an enchantment.
//	 (It's not a creature.)"
//
// The drain half is Sanguine Bond on a body: every positive
// EventChangeLife on the controller — lifelink included — becomes a
// targeted loss, and the amount is captured by value when the trigger
// is built, so a life total that moves again before it resolves does
// not change the drain.
//
// # The Glimmer half is deliberately left out
//
// The second ability returns the creature as an ENCHANTMENT that is
// no longer a creature — a permanent whose printed creature type has
// been stripped for the rest of its existence, for this object only.
// There is no per-instance "came back without its creature type"
// state to hang that on: the type line is printed data and the
// Layer 4 machinery is keyed on the catalog entry, which is shared by
// every copy of the card in every zone.
//
// The failure mode of shipping it anyway is the one that is never
// allowed. Returned as an ordinary creature, Enduring Tenacity dies
// again, triggers again, and returns again — a free, unkillable,
// infinitely recursive 4/3 drain engine. That is far STRONGER than
// printed, so the clause is dropped rather than approximated; the
// card as registered is a 4/3 that dies once, which is weaker than
// printed and therefore the right direction.
//
// The clause becomes writable once the engine can carry a
// per-instance type override through a battlefield entry — the same
// state a "loses all creature types permanently" effect would need.
func init() {
	Register(Spec{
		OracleID:     "98e698ae-1a69-469c-9cfb-0e3fedeb71d4",
		Name:         "Enduring Tenacity",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"When this dies, it doesn't return to the battlefield as an enchantment — it goes to the graveyard like any other creature.",
		},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventChangeLife},
			AppliesTo: YouGainedLife,
			Targets:   TargetPlayer("target opponent", Opponent()),
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source,
					"Enduring Tenacity — target opponent loses that much life",
					drainTargetedOpponent(ev.Amount))
			},
		}},
	})
}

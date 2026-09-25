package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rancor — Enchantment — Aura for {G}:
//
//	"Enchant creature
//	 Enchanted creature gets +2/+0 and has trample.
//	 When this Aura is put into a graveyard from the battlefield,
//	 return it to its owner's hand."
//
// The first Aura in the catalog, and a deliberately good first one:
// every clause lands on machinery that already exists.
//
//   - "Enchant creature" is an ordinary Spec.Targets. The engine
//     validates it at announce, counters the spell by game rules if
//     the creature is gone at resolution (CR 608.2b), attaches the
//     Aura to what it targeted as it enters, and re-runs this same
//     spec as a state-based action for as long as it stays there
//     (CR 704.5m). One declaration, four consumers.
//   - +2/+0 is a layer 7c modify and trample is a layer 6 grant,
//     both scoped to the enchanted creature by the attachment.
//   - The recursion clause is an LTB trigger, and it is what makes
//     Rancor the Aura that is actually safe to play: the two-for-one
//     that makes most Auras bad does not happen, so it is the right
//     card to prove the "Aura falls off into the graveyard" branch of
//     the SBA with — the failure is visible as a card that never
//     comes back.
//
// The recursion fires on ANY trip to the graveyard from the
// battlefield, which is the printed wording: the creature dying, the
// creature being exiled (the Aura is put into the graveyard as a
// state-based action), Disenchant, a board wipe. All of them route
// through EventLTB.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9d2d6479-531c-4ce1-b52b-00e36fa63b64",
		Name:         "Rancor",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			GrantToAttached("trample"),
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				// EventLTB is emitted only from the battlefield, so
				// cardDied's "landed in a graveyard" is the whole
				// condition. (It carries no OldZone — the paired
				// EventZoneMove is where that lives.)
				return cardDied(ev, source)
			},
			Key: "Rancor — return it to your hand",
			Effect: func(g *game.Game, item *game.StackItem) error {
				return ReturnFromGraveyard{
					Target: item.SourceCardID,
					Dest:   game.ZoneHand,
				}.Apply(NewContext(g, item))
			},
		}},
	})
}

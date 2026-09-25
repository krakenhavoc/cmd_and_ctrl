package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Witch Enchanter // Witch-Blessed Meadow — the FRONT face, Creature
// — Human Warlock {3}{W}, 2/2:
//
//	"When this creature enters, destroy target artifact or
//	 enchantment an opponent controls."
//
// The land back (pay 3 life or enter tapped; {T}: Add {W}) is the
// mdfc_lands.go row under "<oracle>#1"; this is face 0, which keeps
// the bare oracle ID (game.CatalogKey). Acidic Slime's shape exactly,
// narrowed to "an opponent controls" and to the two card types
// printed: mandatory, so the pick_target prompt queues as soon as the
// creature enters, and the trigger is removed if the opponent has
// nothing that qualifies (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0355249a-8e4e-41db-9cea-1b901faffbe6",
		Name:         "Witch Enchanter",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			Key:     "Witch Enchanter — destroy target artifact or enchantment",
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Targets: TargetPermanent("target artifact or enchantment an opponent controls",
				Or(Artifact(), Enchantment()), OpponentControls()),
			Effect: destroyChosenPermanent,
		}},
	})
}

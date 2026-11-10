package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stormcage Containment Facility — Plane — Unknown Planet (Doctor Who,
// #4870):
//
//	"Each creature card in your graveyard has escape. The escape cost is
//	 equal to the card's mana cost plus exile three other cards from
//	 your graveyard.
//	 Whenever chaos ensues, detain target creature an opponent controls.
//	 (Until your next turn, that creature can't attack or block and its
//	 activated abilities can't be activated.)"
//
// A plane (ADR 0216 §6), controlled by the planar controller (CR 901.6).
// The static is Underworld Breach's standing cast permission over the
// graveyard cards of the planar controller, read from the face-up plane
// (cast_permission.go), so a card has escape exactly while the plane is
// face up. A card with no mana cost has an unpayable escape cost (the
// ruling). The chaos ability (CR 311.7) detains the target until the planar
// controller's next turn (CR 701.35a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "204ee01c-74ca-4c4c-92db-aebeebb89406",
		Name:         "Stormcage Containment Facility",
		Completeness: CompletenessFull,
		CastPermissions: []game.CastPermission{{
			Zone:                    game.ZoneGraveyard,
			Filter:                  game.PermissionFilter{CardInZone: true, CreatureOnly: true},
			AltCostKey:              "escape",
			ExileOtherFromGraveyard: 3,
			Label:                   "Escape — its mana cost, Exile three other cards from your graveyard",
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverChaosEnsues("Stormcage Containment Facility — detain target creature an opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return Detain{Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID == id },
						Label: "Stormcage Containment Facility — detain"}.Apply(ctx)
				}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

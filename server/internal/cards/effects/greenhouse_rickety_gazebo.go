package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Greenhouse // Rickety Gazebo — Enchantment — Room (ADR 0103):
//
//	Greenhouse {2}{G}: "Lands you control have "{T}: Add one mana of
//	  any color.""
//	Rickety Gazebo {3}{G}: "When you unlock this door, mill four cards,
//	  then return up to two permanent cards from among them to your
//	  hand."
//
// Greenhouse is Chromatic Lantern's layer-6 grant (ADR 0093) behind the
// left door's gate. Rickety Gazebo picks from the cards the mill
// actually put into the graveyard (MillToZone.Then), through
// mayTakeUpToFromAmongThem.
//
// No simplification.
func init() {
	const grant = "greenhouse/any-color"
	Register(Room(RoomSpec{
		OracleID:     "a341eb75-e6a0-468b-8c83-61102249c648",
		Name:         "Greenhouse // Rickety Gazebo",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{AnyColorManaGrant(grant)},
		Left:         Door{Static: []game.StaticAbility{GrantAbilities(landsYouControl, grant)}},
		Right: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorRight, "Rickety Gazebo — mill four, then return up to two permanent cards", ricketyGazeboMill),
		}},
	}))
}

func ricketyGazeboMill(g *game.Game, item *game.StackItem) error {
	return MillToZone{
		N: 4,
		Then: func(ctx *Context, milled []uuid.UUID) error {
			return mayTakeUpToFromAmongThem(ctx, milled, 2, Permanent(),
				"Rickety Gazebo — return up to two permanent cards from among them to your hand")
		},
	}.Apply(NewContext(g, item))
}

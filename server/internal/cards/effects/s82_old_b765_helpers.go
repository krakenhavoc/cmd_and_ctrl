package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b765_helpers.go — shared pieces of S82 OLD batch 765 (#5628):
// Tempest enchantments.

// GrantedAltCostAnyPlayerSmallCreaturesFreeFlash is "Any player may cast
// creature spells with mana value 3 or less without paying their mana costs
// and as though they had flash." (Aluren). Never renamed, never reused.
const GrantedAltCostAnyPlayerSmallCreaturesFreeFlash = "granted-any-player-small-creatures-free-flash"

// AnyPlayerMayCastSmallCreaturesFreeWithFlash is Aluren's static. The offer is
// free (no mana cost, CR 118.9), made to every player (EveryPlayer), over
// creature spells with mana value 3 or less, and carries "as though it had
// flash" with the claim (CR 601.3c).
func AnyPlayerMayCastSmallCreaturesFreeWithFlash() game.GrantedAlternativeCost {
	three := 3
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:           GrantedAltCostAnyPlayerSmallCreaturesFreeFlash,
			Label:         "Cast it without paying its mana cost, as though it had flash",
			AsThoughFlash: true,
		},
		Spells:       game.PermissionFilter{CreatureOnly: true},
		MaxManaValue: &three,
		EveryPlayer:  true,
	}
}

// DestroyTheObjectNamedByTheTrigger is the body of a trigger whose Build
// recorded the creature to destroy as Params.Object (No Quarter): destroy it
// unless it has left the battlefield or come back as a new object (CR 400.7).
func DestroyTheObjectNamedByTheTrigger(g *game.Game, item *game.StackItem) error {
	obj := item.Params.Object
	if obj.ID == uuid.Nil || g.ObjectEpochForEffect(obj.ID) != obj.Epoch || !OnBattlefield(g, obj.ID) {
		return nil
	}
	return DestroyTarget{Target: obj.ID}.Apply(NewContext(g, item))
}

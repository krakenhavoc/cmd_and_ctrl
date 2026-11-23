package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b186_helpers.go — shared bodies for S82 GAP batch 186 (#5948),
// Warhammer 40,000 Commander artifacts.

// GapB186ReturnToBattlefieldBody is a delayed trigger's body: "return that
// card to the battlefield under its owner's control" at the beginning of the
// next end step (Resurrection Orb). The card is the object the trigger named
// when it was scheduled (CR 400.7): one that has left the graveyard since
// (reanimated, exiled) is a different object and is left alone.
var GapB186ReturnToBattlefieldBody = game.DelayedBody("gap-b186/return-to-battlefield", gapB186ReturnToBattlefield)

func gapB186ReturnToBattlefield(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	id := p.Object.ID
	c, ok := g.LookupCardForEffect(id)
	if !ok || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield}.Apply(NewContext(g, item))
}

// GapB186ThreeLandsSharingAName is "activate only if you control three or
// more lands with the same name".
func GapB186ThreeLandsSharingAName(g *game.Game, controller, _ uuid.UUID) bool {
	byName := map[string]int{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != controller || !c.IsLand() {
			continue
		}
		byName[c.Name]++
		if byName[c.Name] >= 3 {
			return true
		}
	}
	return false
}

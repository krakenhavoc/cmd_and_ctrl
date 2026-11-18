package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_rpr_b03_helpers.go — shared pieces for the S82 reprint pass, batch 3
// (#5908).

// rprB03AbilitiesFromOpponentsCantTargetThis is Shanna, Sisay's Legacy's
// "can't be the target of abilities your opponents control".
//
// It refuses only ABILITIES (the targeting source is not a spell), and
// only an opponent's: "your opponents" is read off the creature's
// controller, as hexproof's is (CR 702.11b), so Shanna's own controller
// may still target her with an ability. The source is judged as it is
// now, or as it last existed when it has left the battlefield (CR 608.2h).
// It is asked at announcement (CR 601.2c, 602.2b) and again on resolution
// (CR 608.2b).
func rprB03AbilitiesFromOpponentsCantTargetThis(label string) game.TargetingRestriction {
	return game.TargetingRestriction{
		Label: label,
		Zones: []game.ZoneKind{game.ZoneBattlefield},
		Forbids: func(q game.TargetingQuery) bool {
			if q.Card.InstanceID != q.Source.InstanceID || q.Controller == q.Card.Controller {
				return false
			}
			return q.Controller != uuid.Nil && !q.By.IsSpell()
		},
	}
}

// rprB03IsEnchantedOrEquipped is "enchanted or equipped": an Aura or an
// Equipment is attached to the permanent right now (CR 303.4, 301.5).
func rprB03IsEnchantedOrEquipped(g *game.Game, id uuid.UUID) bool {
	for _, a := range g.BattlefieldCardsForEffect() {
		if a.IsAttachedTo(id) && (a.IsAura() || a.HasSubtype("Equipment")) {
			return true
		}
	}
	return false
}

// rprB03IsOfChosenType is "of the chosen type" for Umori's choice of any
// card type, over the card's effective types.
func rprB03IsOfChosenType(chosen string, c game.Card) bool {
	switch chosen {
	case "":
		return false
	case "kindred":
		return c.HasCardType("kindred") || c.HasCardType("tribal")
	default:
		return c.HasCardType(chosen)
	}
}

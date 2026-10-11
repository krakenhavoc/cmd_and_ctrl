package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oathsworn Vampire — Creature — Vampire Knight {1}{B}, 2/2:
//
//	"This creature enters tapped.
//	 You may cast this card from your graveyard if you gained life this
//	 turn."
//
// The graveyard cast is the CastCondition/CastConditionLabel pair The
// Indomitable uses: the card opens its own graveyard (CastableZones) and
// the condition, which looks the card's current zone up, gates that path
// ONLY. A hand cast is never restricted, and the graveyard cast costs the
// printed {1}{B}, so the card declares no alternative cost.
//
// "If you gained life this turn" is read when the spell is announced, off
// the controller's own turn tally of life gained (lifelink, a drain's
// gain half and a gain-life effect all count; life LOST does not), and
// never again once the spell is on the stack. The tally resets with the
// turn, so a gain on an earlier turn opens nothing and an opponent's gain
// is not yours.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:           "e1fb696f-1df8-4571-9710-0d79001326d3",
		Name:               "Oathsworn Vampire",
		Completeness:       CompletenessFull,
		Replacements:       []game.ReplacementEffect{SelfEntersTapped()},
		CastableZones:      []game.ZoneKind{game.ZoneGraveyard},
		CastCondition:      oathswornVampireGraveyardCastCondition,
		CastConditionLabel: "You can't cast this from your graveyard unless you gained life this turn.",
	})
}

// oathswornVampireGraveyardCastCondition is "if you gained life this
// turn", scoped to the graveyard path: a cast from any other zone is
// always permitted.
func oathswornVampireGraveyardCastCondition(g *game.Game, controller uuid.UUID, card game.Card) bool {
	z := g.FindCardZoneForEffect(card.InstanceID)
	if z == nil || z.Kind != game.ZoneGraveyard {
		return true
	}
	return b15LifeGainedThisTurn(g, controller) > 0
}

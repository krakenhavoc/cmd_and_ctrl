package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// soc_b18_helpers.go — shared pieces for the eighteenth Secrets of
// Strixhaven Commander batch (#4834).

// GrantedAltCostFreeInstantOrSorceryOnceYourTurn is "Once during each of
// your turns, you may cast an instant or sorcery spell from your hand
// without paying its mana cost" (Zaffai and the Tempests). An on-disk key:
// never renamed, never reused.
const GrantedAltCostFreeInstantOrSorceryOnceYourTurn = "granted-free-instant-sorcery-once-your-turn"

// CastAnInstantOrSorceryFromHandFreeOnceEachOfYourTurns is Zaffai's static:
// a granted alternative cost (ADR 0118 §3) that is free, so a spell with
// {X} is cast with X = 0 (CR 107.3b), reaches only instant and sorcery
// spells from the hand, is spent once per Zaffai object (a second one that
// arrives later opens another use) and only during its controller's own
// turn. Additional costs are still paid (CR 118.9d).
func CastAnInstantOrSorceryFromHandFreeOnceEachOfYourTurns() game.GrantedAlternativeCost {
	return game.GrantedAlternativeCost{
		Offer: game.AlternativeCost{
			Key:   GrantedAltCostFreeInstantOrSorceryOnceYourTurn,
			Label: "Cast without paying its mana cost",
		},
		Zones:        []game.ZoneKind{game.ZoneHand},
		Spells:       game.PermissionFilter{InstantOrSorceryOnly: true},
		OnceEachTurn: true,
		YourTurnOnly: true,
	}
}

// socB18EachFractalYouControl is every Fractal permanent the controller
// controls right now, read through the layers and through the text of the
// object asking (a changed creature-type word follows). "Each Fractal you
// control" is judged as the effect applies (CR 608.2h), so a Fractal the
// same spell just made is among them.
func socB18EachFractalYouControl(ctx *Context) []uuid.UUID {
	fractal := ctx.TextSubtype("Fractal")
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.HasSubtype(fractal) {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

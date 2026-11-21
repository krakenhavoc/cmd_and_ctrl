package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unburial Rites — Sorcery {4}{B} (EDHREC rank 4275):
//
//	"Return target creature card from your graveyard to the
//	 battlefield.
//	 Flashback {3}{W} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Two reanimations for one card, and the reason the reanimator deck
// runs white: the plan is to discard or mill this along with the
// creature, cast it from the graveyard for {3}{W}, and never pay the
// five in the corner at all.
//
// Flashback is a cast PATH rather than a second mana cost, which is
// exactly why the two halves can be different colours. Two fields say
// so and both are load-bearing:
//
//   - CastableZones opens the graveyard. Without it the offer would be
//     unclaimable, and Register refuses a flashback cost declared
//     without the zone for that reason.
//   - Flashback() carries ExileOnLeavingStack, a REPLACEMENT applied
//     any time the spell would leave the stack (CR 702.34a). So a
//     flashed-back Rites that fizzles is exiled, one that is countered
//     is exiled, and the card cannot be flashed back twice. Appending
//     an exile to the resolution instead would have made it an
//     every-turn engine.
//
// The creature comes back under ITS OWNER's control — which is the
// engine's reanimation default and the printed answer, since the card
// only reaches your graveyard if you own it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "d48e1545-7997-45ed-83a1-aee45b3d3d20",
		Name:             "Unburial Rites",
		Completeness:     effects.CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{effects.Flashback("{3}{W}")},
		Targets:          effects.TargetCardInGraveyard("target creature card in your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnFirstLegalGraveyardTargetToBattlefield(ctx.Game, item)
		},
	})
}

package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dread Return — Sorcery {2}{B}{B}:
//
//	"Return target creature card from your graveyard to the
//	 battlefield.
//	 Flashback—Sacrifice three creatures. (You may cast this card from
//	 your graveyard for its flashback cost. Then exile it.)"
//
// The hand-cast half is an ordinary reanimation, the same shape
// Unburial Rites and Zombify use: the creature returns under its
// OWNER's control, which is the engine's reanimation default and the
// printed answer for "target creature card in YOUR graveyard".
//
// The flashback half (#1727) is FlashbackSacrifice: no mana at all,
// three creatures the caster controls named in alt_cost_ids and
// sacrificed with the spell already on the stack. So their dies
// triggers resolve before the reanimation does, a countered Dread
// Return leaves them dead, and CR 702.34a exiles the card however it
// leaves the stack. The target is chosen before the cost is paid (CR
// 601.2c before 601.2h), so none of the three can be the creature it
// brings back — they are still on the battlefield when it targets.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "352b64d2-2ae5-44ee-a64f-94932ef545d3",
		Name:          "Dread Return",
		Completeness:  effects.CompletenessFull,
		Targets:       effects.TargetCardInGraveyard("target creature card in your graveyard", effects.YouOwn(), effects.Creature()),
		Purpose:       effects.ForTargets(effects.Reanimates(0)),
		CastableZones: []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{
			effects.FlashbackSacrifice(3, "three creatures", effects.Creature()),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnFirstLegalGraveyardTargetToBattlefield(ctx.Game, item)
		},
	})
}

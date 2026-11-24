package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Colfenor, the Last Yew — Legendary Creature — Treefolk Shaman
// {3}{W}{B}{G}, 3/7 (Commander Legends):
//
//	"Reach, vigilance
//	 Whenever Colfenor or another creature you control dies, return up to one
//	 other target creature card with lesser toughness from your graveyard to
//	 your hand."
//
// One trigger for Colfenor's own death and each other creature of the
// controller's that dies (CR 603.10a: a leaves-the-battlefield ability looks
// back). "Lesser toughness" is measured against the creature that died, as it
// last existed on the battlefield (CR 603.10, 608.2h), and the clause is built
// from the triggering event as the trigger is put on the stack
// (TargetsFrom, Riveteers Ascendancy's shape). "Other" is the creature that
// died: it is never its own target. The target is optional (up to one).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "90dd050d-f58d-4590-9fc7-af624ee4a13f",
		Name:            "Colfenor, the Last Yew",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "vigilance"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.ACreatureYouControlDied(ev, source, lki, g)
			},
			Key:         colfenorLabel,
			TargetsFrom: colfenorClause,
			Effect:      effects.ReturnTargetCardToHand,
		}},
	})
}

const colfenorLabel = "Colfenor, the Last Yew — return up to one other target creature card with lesser toughness from your graveyard to your hand"

func colfenorClause(tc game.TriggerContext, _ *game.Card, g *game.Game) *game.TargetSpec {
	ceiling := 0
	if info, ok := g.LastKnownPermanentForEffect(tc.Event.CardID); ok {
		ceiling = info.Toughness
	}
	spec := effects.TargetCardInGraveyard("up to one other target creature card with lesser toughness from your graveyard",
		effects.YouOwn(), effects.Creature(), effects.ToughnessLE(ceiling-1),
		func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID != tc.Event.CardID })
	return spec.WithCount(0, 1)
}

package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Join the Dead — Instant {1}{B}{B} (The Lost Caverns of Ixalan, #4470):
//
//	"Target creature gets -5/-5 until end of turn.
//	 Descend 4 — That creature gets -10/-10 until end of turn instead if
//	 there are four or more permanent cards in your graveyard."
//
// Dark Deed's body with the amount chosen as the spell resolves: the
// descend count (permanentCardsInGraveyard, CR 110.4) is read then, so a
// permanent card that reached your graveyard in response counts. Only
// one of the two applies ("instead"), never both.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b153fc70-0312-4856-befd-bb6b9a04a26e",
		Name:         "Join the Dead",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -5, -5)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			n := -5
			if effects.PermanentCardsInGraveyard(ctx.Game, item.Controller) >= 4 {
				n = -10
			}
			return effects.BoostUntilEOT{
				Target:    t.ID,
				Power:     n,
				Toughness: n,
				Label:     "Join the Dead — until end of turn",
			}.Apply(ctx)
		},
	})
}

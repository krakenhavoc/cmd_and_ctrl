package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Izzet Steam Maze — Plane — Ravnica (Planechase Anthology, #4870):
//
//	"Whenever a player casts an instant or sorcery spell, that player
//	 copies it. The player may choose new targets for the copy.
//	 Whenever chaos ensues, instant and sorcery spells you cast this turn
//	 cost {3} less to cast."
//
// A plane (ADR 0216 §6). The first ability is mandatory and belongs to the
// caster, who controls the copy (CR 707.10); the copy is made on the stack,
// so it is not cast and does not trigger the plane again. The spell is
// copied from last-known information when it has already left the stack
// (the ruling: a copy is made even if the original was countered), and
// the copy keeps its modes and X. The chaos ability is a stored
// "this turn" promise for the planar controller (CR 601.2f), so it holds
// after a planeswalk away.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e5f8b2bb-e9bb-4a0a-aa88-7fa0299f52a9",
		Name:         "Izzet Steam Maze",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, planesR6APlayerCastsInstantOrSorcery,
				"Izzet Steam Maze — that player copies the spell and may choose new targets",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					ev := ctx.Trigger().Event
					return CopySpell{
						StackID:          ev.CardID,
						Controller:       ev.Actor,
						ChooseNewTargets: true,
						FromLastKnown:    true,
					}.Apply(ctx)
				}),
			WheneverChaosEnsues("Izzet Steam Maze — instant and sorcery spells you cast this turn cost {3} less",
				Do(GrantNextSpellPromise{From: "Izzet Steam Maze", Promise: game.NextSpellPromise{
					Filter:             game.PermissionFilter{InstantOrSorceryOnly: true},
					Reduce:             3,
					EveryMatchingSpell: true,
					Text:               "Instant and sorcery spells you cast this turn cost {3} less to cast.",
				}})),
		},
	})
}

package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necromantic Summons — Sorcery {4}{B} (Magic Origins, #5549):
//
//	"Put target creature card from a graveyard onto the battlefield under
//	 your control.
//	 Spell mastery — If there are two or more instant and/or sorcery cards in
//	 your graveyard, that creature enters with two additional +1/+1 counters
//	 on it."
//
// Coalstoke Gearhulk's graveyard return with the counters as the creature
// enters (CR 614.1c), so they are on it before any enters trigger. Spell
// mastery is read as the spell resolves, while it is still on the stack, so
// it does not count itself. A target that has left its graveyard does
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f921740-de98-4276-8196-47a38a1e4591",
		Name:         "Necromantic Summons",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) == 0 {
				return nil
			}
			id := legal[0].ID
			if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
				return nil
			}
			var counters map[string]int
			if effects.SpellMastery(ctx.Game, ctx.Controller()) {
				counters = map[string]int{game.CounterPlusOne: 2}
			}
			return ctx.Game.ReturnFromGraveyardWithCountersKeyedForEffect(id, item.Controller, false, counters, game.KeyedTail{})
		},
	})
}

package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cut // Ribbons — Sorcery // Sorcery (Secret Lair Drop):
//
//	Cut {1}{R}: "Cut deals 4 damage to target creature."
//	Ribbons {X}{B}{B}: "Aftermath (Cast this spell only from your graveyard.
//	 Then exile it.) Each opponent loses X life."
//
// A split card with an aftermath half: two registered faces under the card's
// oracle ID and its "#1" suffix, the aftermath half castable only from a
// graveyard and exiled as it resolves (CR 702.127a, game/split.go). Ribbons'
// life loss is life loss, not damage, for each opponent at once.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98a5bf1a-1088-4339-9a9b-6ee5e4956cf1",
		Name:         "Cut",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve:    effects.DamageToFirstTargetStep(4),
	})
	effects.Register(effects.Spec{
		OracleID:     "98a5bf1a-1088-4339-9a9b-6ee5e4956cf1#1",
		Name:         "Ribbons",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.EachOpponentLosesLifeNow(ctx.Game, item, ctx.X())
		},
	})
}

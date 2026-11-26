package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Postmortem Lunge — Sorcery {X}{B/P}:
//
//	"({B/P} can be paid with either {B} or 2 life.)
//	 Return target creature card with mana value X from your graveyard to the
//	 battlefield. It gains haste. Exile it at the beginning of the next end
//	 step."
//
// The target clause is bound to the announced X (CR 202.3e: in the graveyard
// a card's X is 0, so X picks the target and a mana value 0 card is a real
// play at X = 0).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aacd1aae-1bf0-43e4-99f4-c6d738af812c",
		Name:         "Postmortem Lunge",
		Completeness: effects.CompletenessFull,
		Targets: effects.TargetCardInGraveyard("target creature card with mana value X in your graveyard",
			effects.YouOwn(), effects.Creature()).WithManaValueEqualsX(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PostmortemLungeReturn(ctx.Game, item)
		},
	})
}

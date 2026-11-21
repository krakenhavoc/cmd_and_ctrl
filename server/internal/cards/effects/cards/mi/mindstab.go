package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mindstab — Sorcery {5}{B} (Time Spiral, #5596):
//
//	"Target player discards three cards.
//	 Suspend 4—{B} (Rather than cast this card from your hand, you may
//	 pay {B} and exile it with four time counters on it. At the
//	 beginning of your upkeep, remove a time counter. When the last is
//	 removed, you may cast it without paying its mana cost.)"
//
// Fugue's three-card discard, castable for {5}{B} or suspended for {B}.
// The targeted player picks the cards (CR 701.9b); on the free cast from
// suspend the target is chosen then, as the spell is cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "338c06ec-47b5-47e7-9fe1-39ff381b36a5",
		Name:         "Mindstab",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		SpecialActions: []game.SpecialAction{
			effects.Suspend(4, "{B}"),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TargetPlayerDiscards(3)(ctx.Game, item)
		},
	})
}

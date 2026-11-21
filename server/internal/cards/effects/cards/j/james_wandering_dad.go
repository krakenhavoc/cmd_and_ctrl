package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// James, Wandering Dad // Follow Him — Legendary Creature — Human Scientist
// {2}{U}, 2/4 // Instant — Adventure {X}{U}{U} (Fallout):
//
//	James, Wandering Dad — "{T}: Add {C}{C}. Spend this mana only to
//	 activate abilities."
//	Follow Him — "Investigate X times. (Then exile this card. You may cast
//	 the creature later from exile.)"
//
// Both faces register under the Adventure keyspace (the bare oracle ID and
// "#1"). James's mana ability carries Ambulatory Reactor's activate-only
// restriction. Follow Him announces X and investigates that many times
// (CR 701.16), one Clue each; X = 0 does nothing.
//
// No simplification.
const jamesWanderingDadOracleID = "8c90b558-de07-4689-9427-da68d823c2e8"

func init() {
	effects.Register(effects.Spec{
		OracleID:     jamesWanderingDadOracleID,
		Name:         "James, Wandering Dad",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{C}{C}",
			Label:        "Add {C}{C}. Spend this mana only to activate abilities",
			Restrictions: []string{effects.ManaRestrictActivate},
		}},
	})
	effects.Register(effects.Spec{
		OracleID:     jamesWanderingDadOracleID + "#1",
		Name:         "Follow Him",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.Investigate{N: ctx.X()}.Apply(ctx)
		},
	})
}

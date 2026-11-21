package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// jace_instant_speed_2797_test.go — #2797: "Until end of turn, you may
// activate loyalty abilities of Jace planeswalkers you control on any
// player's turn any time you could cast an instant" (Jace's Machinations),
// proved on a fixture instant and fixture planeswalkers so it does not wait
// on "empower Jace" (#2796).
//
// The Teferi helpers state this for a source (Teferi, Master of Time) or an
// emblem (Teferi, Temporal Archmage); neither lasts "until end of turn". The
// resolving-spell form is a stored statement on the player with a duration
// (game/activation_timing_grant.go).

const Pw2797MachinationsOracle = "test-2797-machinations"

func init() {
	effects.Register(effects.Spec{
		OracleID: Pw2797MachinationsOracle,
		Name:     "Fixture Machinations",
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GrantLoyaltyAbilitiesAtInstantSpeed{
				Subtype: "Jace",
				Label:   "Fixture Machinations — loyalty abilities of Jace planeswalkers at instant speed",
			}.Apply(ctx)
		},
	})
}

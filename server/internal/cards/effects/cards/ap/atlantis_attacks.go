package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Atlantis Attacks — {5}{U}{U} Sorcery:
//
//	"Teamwork 4 (As an additional cost to cast this spell, you may tap
//	 any number of creatures you control with total power 4 or more.)
//	 Choose one. If this spell was cast using teamwork, choose both
//	 instead.
//	 • Target player creates a 6/5 blue Leviathan creature token with
//	   hexproof.
//	 • Return one or two target nonland permanents to their owners'
//	   hands."
//
// #1703: Teamwork(4) with InsteadIf(2, TeamworkUsed). The Leviathan is
// created under the TARGET player's control, whoever that is. The
// bounce clause is one-or-two in one bullet, and each still-legal
// target is returned (CR 608.2b). Printed order (CR 608.2c, run by the engine): the token
// is created before anything is returned, and a target is chosen at
// announce, so the new token can never be one of them. No
// simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "b5c8e24a-f4d7-42dc-8117-905ad0bad888",
		Name:          "Atlantis Attacks",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Teamwork(4)},
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target player creates a 6/5 blue Leviathan creature token with hexproof.",
				effects.TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return effects.CreateToken{Controller: t.ID, Template: effects.TokenCard("6/5 blue Leviathan with hexproof"), N: 1}.Apply(ctx)
				}),
				effects.ForTargets(game.TargetPurpose{Slot: 0, Tokens: 1})),
			effects.ModeDoing("Return one or two target nonland permanents to their owners' hands.",
				effects.TargetPermanent("one or two target nonland permanents", effects.Nonland()).WithCount(1, 2),
				effects.BounceTheModesTargets),
		).InsteadIf(2, effects.TeamworkUsed),
	})
}

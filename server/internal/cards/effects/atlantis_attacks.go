package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:      "b5c8e24a-f4d7-42dc-8117-905ad0bad888",
		Name:          "Atlantis Attacks",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Teamwork(4)},
		Modes: ChooseOne(
			ModeDoing("Target player creates a 6/5 blue Leviathan creature token with hexproof.",
				TargetPlayer("target player"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok || t.Kind != game.TargetPlayer {
						return nil
					}
					return CreateToken{Controller: t.ID, Template: TokenCard("6/5 blue Leviathan with hexproof"), N: 1}.Apply(ctx)
				}),
			ModeDoing("Return one or two target nonland permanents to their owners' hands.",
				TargetPermanent("one or two target nonland permanents", Nonland()).WithCount(1, 2),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					for _, t := range ctx.ModeTargets(occ) {
						if !ctx.IsTargetLegal(t) {
							continue
						}
						if err := (BounceToHand{Target: t.ID}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}),
		).InsteadIf(2, TeamworkUsed),
	})
}

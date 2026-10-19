package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venser, the Sojourner — Legendary Planeswalker — Venser {3}{W}{U},
// loyalty 3:
//
//	"+2: Exile target permanent you own. Return it to the battlefield
//	     under your control at the beginning of the next end step.
//	 −1: Creatures can't be blocked this turn.
//	 −8: You get an emblem with "Whenever you cast a spell, exile target
//	     permanent.""
//
// +2 is the delayed blink (exileTargetsThenScheduleReturn): exiled as
// the ability resolves, returned by a CR 603.7 delayed trigger at the
// next end step. The target must be a permanent you OWN, so the return
// "under your control" and the helper's "under its owner's control" are
// the same player. Venser may target itself: it comes back at the end
// step as a new object with its printed loyalty (CR 306.5b).
//
// −1 changes the rules, not a characteristic, so it covers every
// creature for the rest of the turn, one that enters later included
// (CR 611.2c): RestrictUntilEOT over game.ScopeCreatures, Reverse the
// Polarity's third mode. Every player's creatures, as printed.
//
// −8's emblem lives in the command zone (CR 114.3). Its trigger fires
// on every spell its owner casts, goes on the stack above the spell,
// and targets as it is put there (CR 603.3d); with no permanent to
// target it is removed. The exile is checked against the target on
// resolution (CR 608.2b).
//
// No simplification. Roon deck request (#4128).
func init() {
	Register(Spec{
		OracleID:        "a8bf8ff8-d924-4fd2-b5ed-05b38f55325a",
		Name:            "Venser, the Sojourner",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &EmblemSpec{
			Label: "Venser, the Sojourner emblem",
			Text:  "Whenever you cast a spell, exile target permanent.",
			Triggered: []game.TriggeredAbility{
				Targeting(
					WheneverYouCast(nil, "Venser, the Sojourner emblem — exile target permanent", exileFirstLegalCardTarget),
					TargetPermanent("target permanent"),
				),
			},
		},
		Activated: []ActivatedAbility{
			{
				Label:   "+2: Exile target permanent you own. Return it to the battlefield under your control at the beginning of the next end step.",
				Cost:    LoyaltyCost(2),
				Targets: TargetPermanent("target permanent you own", YouOwn()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return exileTargetsThenScheduleReturn(NewContext(g, item),
						"Venser, the Sojourner — return the exiled permanent")
				},
			},
			{
				Label: "−1: Creatures can't be blocked this turn.",
				Cost:  LoyaltyCost(-1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return RestrictUntilEOT{
						Scope:        game.ScopeCreatures,
						Restrictions: game.CantBeBlocked,
						Label:        "Venser, the Sojourner — creatures can't be blocked this turn",
					}.Apply(NewContext(g, item))
				},
			},
			{
				Label: "−8: You get an emblem with \"Whenever you cast a spell, exile target permanent.\"",
				Cost:  LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

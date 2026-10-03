package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mathas, Fiend Seeker — Legendary Creature — Vampire (3/3) for {R}{W}{B}:
//
//	"Menace
//	 At the beginning of your end step, put a bounty counter on target creature an opponent controls. For as long as that creature has a bounty counter on it, it has "When this creature dies, each opponent draws a card and gains 2 life.""
//
// ADR 0109 §2 (#1604): the bounty counter grants a dies trigger, a
// granted ability bundle (ADR 0093 PR 4), for as long as the creature
// has a bounty counter on it. The trigger is the creature's,
// controlled by its controller, so "each opponent" means that player's
// opponents, and it fires from last-known information as the creature
// dies (CR 603.10a). Mathas leaving does not end it.
//
// No simplification.
const mathasBountyGrant = "mathas-fiend-seeker/bounty"

func init() {
	Register(Spec{
		OracleID:        "1e8d9d0c-185b-4149-9ec9-74a2f2fe8764",
		Name:            "Mathas, Fiend Seeker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Grants: []AbilityGrant{{
			Key:  mathasBountyGrant,
			Text: "When this creature dies, each opponent draws a card and gains 2 life.",
			Triggered: []game.TriggeredAbility{
				WhenThisDies("Mathas, Fiend Seeker — bounty: each opponent draws a card and gains 2 life",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						var err error
						eachOpponent(g, item.Controller, func(opp uuid.UUID) bool {
							if err = (DrawCards{Player: opp, N: 1}).Apply(ctx); err != nil {
								return true
							}
							err = GainLife{Player: opp, Amount: 2}.Apply(ctx)
							return err != nil
						})
						return err
					}),
			},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourEndStep("Mathas, Fiend Seeker — put a bounty counter on target creature an opponent controls",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return GrantWhileItHasCounter(ctx, FirstLegalBattlefieldTarget(ctx), "bounty",
						"Mathas, Fiend Seeker — bounty while it has a bounty counter", mathasBountyGrant)
				}), TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}

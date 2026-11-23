package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Haunted Hellride — Artifact — Vehicle {1}{U}{B}, 3/3 (Aetherdrift,
// #4464):
//
//	"Whenever you attack, target creature you control gets +1/+0 and
//	 gains deathtouch until end of turn. Untap it.
//	 Crew 1 (Tap any number of creatures you control with total power 1
//	 or more: This Vehicle becomes an artifact creature until end of
//	 turn.)"
//
// Great Gilded Boat's "whenever you attack" (once per attack declaration
// of yours, OncePerBatch over attackDeclaredByYou), which triggers whether
// or not the Hellride is crewed or attacking, with Most Valuable Slayer's
// pump-and-grant and an untap after it. The target is any creature you
// control, attacking or not, so the creature that crewed the Hellride can
// be untapped. With no creature to target the trigger is removed
// (CR 603.3d); a target gone by resolution gets nothing (CR 608.2b).
// Crew is the engine's (CR 702.122a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "47944bf1-b954-4143-b479-45c10071e35a",
		Name:         "Haunted Hellride",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.OncePerBatch(effects.Targeting(
				effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.AttackDeclaredByYou(ev, source.Controller)
				}, "Haunted Hellride — target creature you control gets +1/+0 and gains deathtouch until end of turn. Untap it.",
					hauntedHellridePumpAndUntap),
				effects.TargetCreature("target creature you control", effects.YouControl()))),
				effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "deathtouch"))),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 1",
			Cost:    effects.CrewCost(1),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Haunted Hellride"),
		}},
	})
}

// hauntedHellridePumpAndUntap gives the still-legal target +1/+0 and
// deathtouch until end of turn, then untaps it.
func hauntedHellridePumpAndUntap(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target := effects.FirstLegalBattlefieldTarget(ctx)
	if target == uuid.Nil {
		return nil
	}
	if err := (effects.BoostUntilEOT{Target: target, Power: 1, Label: "Haunted Hellride — +1/+0"}).Apply(ctx); err != nil {
		return err
	}
	if err := (effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"deathtouch"},
		Label: "Haunted Hellride — deathtouch"}).Apply(ctx); err != nil {
		return err
	}
	return effects.UntapTarget{Target: target}.Apply(ctx)
}

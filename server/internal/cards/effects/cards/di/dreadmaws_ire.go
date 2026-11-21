package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dreadmaw's Ire — Instant {R} (The Lost Caverns of Ixalan, #4470):
//
//	"Until end of turn, target attacking creature gets +2/+2 and gains
//	 trample and 'Whenever this creature deals combat damage to a player,
//	 destroy target artifact that player controls.'"
//
// Herd Heirloom's duration grant (ADR 0093 PR 4) with a +2/+2 beside the
// trample: one ScopedEffect record, one timestamp, gone at cleanup. The
// granted trigger is the CREATURE's, so its controller controls it, and
// it is Trygon Predator's without the "you may": "that player controls"
// is ControlledByThatPlayer, judged against the player the trigger's own
// damage event named. The attacking test is a target clause, checked as
// the spell is cast and again as it resolves (CR 608.2b).
//
// No simplification.
const dreadmawsIreGrant = "dreadmaws-ire/destroy-artifact"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "781d6981-d389-4677-9eaf-9c46c9192ed6",
		Name:         "Dreadmaw's Ire",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "trample")),
		Targets:      effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
		Grants: []effects.AbilityGrant{{
			Key: dreadmawsIreGrant,
			Triggered: []game.TriggeredAbility{{
				Watches: []game.EventKind{game.EventDealDamage},
				Key:     "Dreadmaw's Ire — destroy target artifact that player controls",
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Source == source.InstanceID && effects.CombatDamageToPlayerBy(ev, source.Controller, g)
				},
				Targets: effects.ControlledByThatPlayer(effects.TargetPermanent("target artifact that player controls", effects.Artifact())),
				Effect:  effects.DestroyFirstLegalTarget,
			}},
			Text: "Whenever this creature deals combat damage to a player, destroy target artifact that player controls.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.GrantAbilitiesFor{
				Target: target,
				Keys:   []string{dreadmawsIreGrant},
				Also:   []game.Mod{game.ModifyPTMod(2, 2), game.AddKeywordsMod("trample")},
				Label:  "Dreadmaw's Ire — +2/+2, trample, and destroy an artifact on combat damage",
			}.Apply(ctx)
		},
	})
}

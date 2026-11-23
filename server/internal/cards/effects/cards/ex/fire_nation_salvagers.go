package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire Nation Salvagers — Creature — Human Soldier {3}{B}{B}, 3/3
// (Avatar: The Last Airbender Eternal, #5075):
//
//	"Menace
//	 When this creature enters, put a +1/+1 counter on target creature or
//	 Vehicle you control.
//	 Whenever one or more creatures you control with counters on them deal
//	 combat damage to a player, put target creature or Vehicle card from
//	 that player's graveyard onto the battlefield under your control."
//
// The damage trigger is one per player dealt combat damage by a creature
// you control that has a counter of any kind on it (CR 603.2c, the
// ruling), and its target clause names that player's graveyard
// (Shark Shredder's TargetsFrom). The card comes back under your control,
// not its owner's (Moseo, Vein's New Dean's body, moseoInfusion).
//
// No simplification.
func init() {
	damage := effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(effects.B64HasAnyCounter(),
		"Fire Nation Salvagers — put target creature or Vehicle card from that player's graveyard onto the battlefield under your control",
		effects.MoseoInfusion)
	damage.TargetsFrom = func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
		return effects.TargetCardInGraveyard("target creature or Vehicle card from that player's graveyard",
			effects.Or(effects.Creature(), effects.Subtype("Vehicle")), effects.OwnedBy(tc.Event.Target))
	}
	effects.Register(effects.Spec{
		OracleID:        "9e76d56a-67b6-4bd9-8f2d-3328d18a2880",
		Name:            "Fire Nation Salvagers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Fire Nation Salvagers — put a +1/+1 counter on target creature or Vehicle you control",
				effects.PlusOneCounterOnTheTarget),
				effects.TargetPermanent("target creature or Vehicle you control", effects.Or(effects.Creature(), effects.Subtype("Vehicle")), effects.YouControl())),
			damage,
		},
	})
}

package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Destined Black Mage — Legendary Creature — Human Wizard {2}{B}, 3/2
// (Final Fantasy, #4462):
//
//	"Deathtouch
//	 {B}, {T}: Another target creature you control gains deathtouch until
//	 end of turn.
//	 Whenever you cast a noncreature spell, The Destined Black Mage deals
//	 1 damage to each opponent. If you have a full party, it deals 3
//	 damage to each opponent instead."
//
// The damage is DamageEachOpponent (Harsh Sustenance's drain family) with
// the amount decided as the trigger resolves: three with a full party
// (CR 700.8c), else one. A Mage that left deals it as it last existed
// (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ef2aa67a-9531-495c-85c7-7f745bc19ce1",
		Name:            "The Destined Black Mage",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.Noncreature(),
				"The Destined Black Mage — 1 damage to each opponent, or 3 with a full party",
				effects.Do(
					effects.If{Cond: effects.FinB9FullParty, Then: []effects.Applier{effects.DamageEachOpponent{N: 3}}},
					effects.If{Cond: effects.FinB9NoFullParty, Then: []effects.Applier{effects.DamageEachOpponent{N: 1}}},
				)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, {T}: Another target creature you control gains deathtouch until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "deathtouch")),
			Cost:    effects.Plus(effects.ManaCost("{B}"), effects.TapCost()),
			Targets: effects.Another(effects.TargetCreature("another target creature you control", effects.YouControl())),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("deathtouch", "The Destined Black Mage — deathtouch until end of turn"),
		}},
	})
}

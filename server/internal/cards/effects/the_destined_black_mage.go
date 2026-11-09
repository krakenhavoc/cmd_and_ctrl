package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "ef2aa67a-9531-495c-85c7-7f745bc19ce1",
		Name:            "The Destined Black Mage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Noncreature(),
				"The Destined Black Mage — 1 damage to each opponent, or 3 with a full party",
				Do(
					If{Cond: finB9FullParty, Then: []Applier{DamageEachOpponent{N: 3}}},
					If{Cond: finB9NoFullParty, Then: []Applier{DamageEachOpponent{N: 1}}},
				)),
		},
		Activated: []ActivatedAbility{{
			Label:   "{B}, {T}: Another target creature you control gains deathtouch until end of turn.",
			Purpose: ForTargets(BenefitUntilEOT(0, 0, 0, "deathtouch")),
			Cost:    Plus(ManaCost("{B}"), TapCost()),
			Targets: Another(TargetCreature("another target creature you control", YouControl())),
			Effect:  ebTargetCreatureGainsUntilEOT("deathtouch", "The Destined Black Mage — deathtouch until end of turn"),
		}},
	})
}

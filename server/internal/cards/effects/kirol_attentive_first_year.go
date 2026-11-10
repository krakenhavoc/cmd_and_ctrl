package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kirol, Attentive First-Year — Legendary Creature — Vampire Cleric
// {1}{R/W}{R/W}, 3/3 (Lorwyn Eclipsed, #4458):
//
//	"Tap two untapped creatures you control: Copy target triggered
//	 ability you control. You may choose new targets for the copy.
//	 Activate only once each turn."
//
// Strionic Resonator's ability with another cost: the target clause
// names a triggered ability item you control on the stack
// (AbilityOnStack with AnAbilityYouControl and TriggeredAbilityOnly),
// and the copy keeps the original's triggering event, mode and source
// (CR 707.10), with the CR 707.10c re-target offer (copyTargetedAbility).
// Choices made on resolution are made separately for the copy (the
// ruling), which is the copy resolving as its own item.
//
// The cost taps two untapped creatures you control, Kirol included: it
// is not the {T} symbol, so a creature that arrived this turn may be
// tapped for it (CR 302.6). "Activate only once each turn" counts
// announcements of this ability on this object (CR 602.1b,
// OncePerTurnActivation).
//
// No simplification.
func init() {
	const (
		label  = "Tap two untapped creatures you control: Copy target triggered ability you control. You may choose new targets for the copy. Activate only once each turn."
		tapTwo = "two untapped creatures you control"
	)
	Register(Spec{
		OracleID:     "3a272fa9-5f28-43a4-95de-10cb511f1b43",
		Name:         "Kirol, Attentive First-Year",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: label,
			Cost: game.AbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: TargetPermanent(tapTwo, Creature()),
				Label:  tapTwo,
			}},
			Targets:   AbilityOnStack("target triggered ability you control", AnAbilityYouControl(), TriggeredAbilityOnly()),
			Condition: OncePerTurnActivation(label),
			Effect:    copyTargetedAbility,
		}},
	})
}

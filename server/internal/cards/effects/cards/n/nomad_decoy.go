package cardsn

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Nomad Decoy — Creature — Human Nomad {2}{W}, 1/2 (Dominaria Remastered,
// #5948):
//
//	"{W}, {T}: Tap target creature.
//	 Threshold — {W}{W}, {T}: Tap two target creatures. Activate only if
//	 there are seven or more cards in your graveyard."
//
// Two activated abilities. The threshold one is gated on the graveyard
// count at activation (a restriction, not a condition that is rechecked on
// resolution), and needs exactly two targets.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "73c40477-7061-47db-af8f-2a5cf6c22c9b",
		Name:         "Nomad Decoy",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}, {T}: Tap target creature.",
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.TapChosenPermanent,
			},
			{
				Label:     "Threshold — {W}{W}, {T}: Tap two target creatures. Activate only if there are seven or more cards in your graveyard.",
				Cost:      effects.Plus(effects.ManaCost("{W}{W}"), effects.TapCost()),
				Targets:   effects.TargetCreature("two target creatures").WithCount(2, 2),
				Condition: effects.GraveyardAtLeast(7, nil),
				Effect:    effects.Msc13TapTheLegalTargets,
			},
		},
	})
}

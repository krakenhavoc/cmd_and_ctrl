package cardsh

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"

// Hallowed Ground — Enchantment {1}{W} (Ice Age, S82 OLD batch 850, #5636):
//
//	"{W}{W}: Return target nonsnow land you control to its owner's hand."
//
// Trade Routes' bounce of a land you control, narrowed to a land without the
// snow supertype (CR 205.4a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ee8fd844-4551-4e44-9740-00d2e76eb337",
		Name:         "Hallowed Ground",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{W}: Return target nonsnow land you control to its owner's hand.",
			Cost:    effects.ManaCost("{W}{W}"),
			Targets: effects.TargetPermanent("target nonsnow land you control", effects.Land(), effects.Not(effects.Snow()), effects.YouControl()),
			Effect:  effects.BounceTheTargetToHand,
		}},
	})
}

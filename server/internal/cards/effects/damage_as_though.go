package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// damage_as_though.go — the printed "damage is dealt as though its source
// had wither / infect" statics (ADR 0108 §10, #1889). The engine reads
// them as each damage event lands (game/damage_as_though.go); a card file
// declares them on Spec.DamageAsThough with one of these.

// AllDamageAsThoughWither is "All damage is dealt as though its source had
// wither." (Everlasting Torment): damage to any creature, from any source,
// is -1/-1 counters (CR 120.3d, 702.80a).
func AllDamageAsThoughWither() []game.DamageAsThoughStatic {
	return []game.DamageAsThoughStatic{{
		Label:  "All damage is dealt as though its source had wither.",
		Wither: true,
	}}
}

// DamageToYouAsThoughInfectAtOrBelowZeroLife is "As long as you have 0 or
// less life, all damage is dealt to you as though its source had infect."
// (Phyrexian Unlife): damage to its controller is poison counters (CR
// 120.3b, 702.90b) while that player had 0 or less life as the damage
// instance began.
func DamageToYouAsThoughInfectAtOrBelowZeroLife() []game.DamageAsThoughStatic {
	return []game.DamageAsThoughStatic{{
		Label:                  "As long as you have 0 or less life, all damage is dealt to you as though its source had infect.",
		Infect:                 true,
		ToYou:                  true,
		WhileAtOrBelowZeroLife: true,
	}}
}

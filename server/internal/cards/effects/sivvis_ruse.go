package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sivvi's Ruse — Instant {2}{W}{W}:
//
//	"If an opponent controls a Mountain and you control a Plains, you
//	 may cast this spell without paying its mana cost.
//	 Prevent all damage that would be dealt this turn to creatures you
//	 control."
//
// The Nemesis free-spell cycle's alternative cost (Submerge's shape):
// an OPPONENT controls a Mountain, and you control a Plains, read off
// effective subtypes as you cast it. The shield is Divine Light's
// (#2045): creatures you control as the damage would be dealt (CR
// 611.2c), and not you.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "c5edfaae-3ecd-41ad-970f-8cecb09bf1e4",
		Name:         "Sivvi's Ruse",
		Completeness: CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{{
			Key:       "free",
			Label:     "Cast without paying its mana cost (an opponent controls a Mountain and you control a Plains)",
			ManaCost:  "",
			Condition: AnOpponentControlsAAndYouControlA("Mountain", "Plains"),
		}},
		OnResolve: sourceShieldSpell(PreventDamageFromSource{Protect: ShieldCreaturesYouControl}),
	})
}

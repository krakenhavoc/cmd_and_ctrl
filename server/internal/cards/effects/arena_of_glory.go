package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Arena of Glory — Land:
//
//	"This land enters tapped unless you control a Mountain.
//	 {T}: Add {R}.
//	 {R}, {T}, Exert this land: Add {R}{R}. If any of that mana is spent
//	 on a creature spell, it gains haste until end of turn. (An exerted
//	 permanent won't untap during your next untap step.)"
//
// The second row is a mana ability (CR 605.1a) with a mana, a tap and
// an exert component (ADR 0130 §4, CR 701.43a). Exerting a land is not
// exerting a creature, so "whenever you exert a creature" never sees it.
// The auto-tapper never pays it on its own: the player activates it.
// The haste is Generator Servant's spend rider: a grant to the SPELL at
// the spend, which the permanent keeps until end of turn (CR 400.7a).
//
// One declared simplification, the rider's (ADR 0068 §3): with strict
// mana off the engine never spends the pool, so the haste never applies.
func init() {
	Register(Spec{
		OracleID:     "63dfe794-5f56-41ec-9883-5523b41cc3e0",
		Name:         "Arena of Glory",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't see which mana you spent, so the creature never gains haste."},
		Replacements: []game.ReplacementEffect{b06EntersTappedUnlessLandType("mountain")},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{R}",
				Label:    "{T}: Add {R}.",
			},
			{
				Cost:     ManaAbilityCost{Tap: true, Exert: true, Mana: "{R}"},
				Produced: "{R}{R}",
				Label:    "{R}, {T}, Exert this land: Add {R}{R}. If any of that mana is spent on a creature spell, it gains haste until end of turn.",
				SpendRiders: []game.ManaSpendRider{
					SpentSpellGains([]string{"haste"}, true, ManaRestrictCast, ManaRestrictType("Creature")),
				},
			},
		},
	})
}

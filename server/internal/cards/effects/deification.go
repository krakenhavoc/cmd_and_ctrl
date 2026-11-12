package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deification — Enchantment {1}{W} (March of the Machine: The Aftermath,
// tracker #5476):
//
//	"As this enchantment enters, choose a planeswalker type.
//	 Planeswalkers you control of the chosen type have hexproof.
//	 As long as you control a creature, if damage dealt to a planeswalker
//	 you control of the chosen type would result in all loyalty counters
//	 on it being removed, instead all but one of those counters are
//	 removed."
//
// The type is chosen as it enters (CR 614.12) among CR 205.3j's
// planeswalker types, the ones you control offered first, and is read
// live: a planeswalker of that type that arrives later is covered, and
// one that loses the type is not. Until the choice is answered neither
// line covers anything.
//
// The last line replaces a RESULT of damage, not the damage (CR 614.1a,
// 120.4c): the damage is dealt in full, so lifelink, "is dealt damage"
// triggers and damage marked on a planeswalker that is also a creature
// are unchanged, and it works under "damage can't be prevented". Only
// the loyalty counters removed are capped (CR 120.3c), at all but one. A
// planeswalker with one loyalty counter keeps it. The creature is looked
// for as the damage lands (#5646, game/damage_last_loyalty.go). A loyalty
// cost or a counter-removing effect is not damage and is not changed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c8f93d7-d89a-401b-8764-76be71ad16e3",
		Name:         "Deification",
		Completeness: CompletenessFull,
		AsEnters:     ChoosePlaneswalkerTypeAsEnters("Deification"),
		Static: []game.StaticAbility{
			Labelled(b16GrantKeywords(PlaneswalkersYouControlOfTheChosenType, "hexproof"),
				"Planeswalkers you control of the chosen type have hexproof."),
		},
		DamageLastLoyalty: DamageLeavesChosenTypeAPlaneswalkersLastLoyaltyWhileYouControlACreature(),
	})
}

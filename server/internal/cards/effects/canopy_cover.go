package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Canopy Cover — Enchantment — Aura {1}{G} (EDHREC rank 4308):
//
//	"Enchant creature
//	 Enchanted creature can't be blocked except by creatures with
//	 flying or reach.
//	 Enchanted creature can't be the target of spells or abilities
//	 your opponents control."
//
// Two mana of Voltron insurance: the commander gets most of the way
// past the ground and, more importantly, stops being a legal target
// for the removal spell that would two-for-one you. It is the second
// clause the card is played for, and the first that makes it green.
//
// THE SECOND CLAUSE IS HEXPROOF, EXACTLY. CR 702.11b defines hexproof
// as "this permanent can't be the target of spells or abilities your
// opponents control", word for word, so GrantToAttached("hexproof") is
// not an approximation — it is the same sentence spelled with the
// keyword the engine's targeting gate reads. Your own Auras and pump
// spells still work on the creature, which is the difference between
// this and a Whispersilk Cloak's shroud.
//
// Declared simplification, weaker than printed (#259): the EVASION
// clause is not implemented. The engine's restriction vocabulary has
// "can't be blocked" and "can't block", and no way to say "can't be
// blocked EXCEPT BY creatures with <keyword>" — a conditional block
// restriction (#750). Shipping the unconditional "can't be blocked"
// would be STRONGER than printed, which the #259 rule forbids, and
// shipping nothing is weaker, which it allows. So the enchanted
// creature is blockable by anything; only the hexproof half is real.
func init() {
	Register(Spec{
		OracleID:     "5b84101e-7e23-437d-835c-409bc061ecbb",
		Name:         "Canopy Cover",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"The enchanted creature can be blocked by anything — the printed \"can't be blocked except by creatures with flying or reach\" isn't enforced, so the Cover is protection only, not evasion.",
		},
		Targets: EnchantCreature(),
		Static: []game.StaticAbility{
			GrantToAttached("hexproof"),
		},
	})
}

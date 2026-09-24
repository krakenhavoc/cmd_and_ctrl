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
// THE FIRST CLAUSE IS A BLOCK RULE on the enchanted creature (#750,
// ADR 0045 addendum Decision 11): every blocker without flying or
// reach is refused at declaration, with a sentence that names the
// clause. Both keywords are read through the effective abilities, so
// a creature granted reach this turn can block it. Until #750 the
// clause was left out, because the only thing the engine could say
// was the unconditional "can't be blocked" — stronger than printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b84101e-7e23-437d-835c-409bc061ecbb",
		Name:         "Canopy Cover",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			GrantToAttached("hexproof"),
		},
		BlockRules: []game.BlockRule{
			CantBeBlockedExceptBy(OnAttached(), Or(HasKeyword("flying"), HasKeyword("reach")),
				"creatures with flying or reach"),
		},
	})
}

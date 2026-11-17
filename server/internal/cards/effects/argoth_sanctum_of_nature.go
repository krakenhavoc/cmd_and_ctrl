package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Argoth, Sanctum of Nature — Land (The Brothers' War, #5482):
//
//	"This land enters tapped unless you control a legendary green
//	 creature.
//	 {T}: Add {G}.
//	 {2}{G}{G}, {T}: Create a 2/2 green Bear creature token, then mill
//	 three cards. Activate only as a sorcery.
//	 (Melds with Titania, Voice of Gaea.)"
//
// The entry is a checkland's shape over a legendary green creature, read
// through the layers when it enters. The mill follows the token in printed
// order and is the controller's own. The reminder line is Titania, Voice of
// Gaea's meld ability and needs nothing here.
func init() {
	Register(Spec{
		OracleID:     "62648946-1708-48f4-ba40-c057563ab11b",
		Name:         "Argoth, Sanctum of Nature",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			EntersTappedUnless(s82BroB20YouControlALegendaryGreenCreature),
		},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{2}{G}{G}, {T}: Create a 2/2 green Bear creature token, then mill three cards. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{2}{G}{G}"), TapCost()),
			SorcerySpeed: true,
			Effect: Do(
				CreateToken{Template: TokenCard("2/2 green Bear"), N: 1},
				MillCards{N: 3},
			),
		}},
	})
}

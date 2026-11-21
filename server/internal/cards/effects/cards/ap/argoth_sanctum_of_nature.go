package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "62648946-1708-48f4-ba40-c057563ab11b",
		Name:         "Argoth, Sanctum of Nature",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.EntersTappedUnless(effects.S82BroB20YouControlALegendaryGreenCreature),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{G}{G}, {T}: Create a 2/2 green Bear creature token, then mill three cards. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{2}{G}{G}"), effects.TapCost()),
			SorcerySpeed: true,
			Effect: effects.Do(
				effects.CreateToken{Template: effects.TokenCard("2/2 green Bear"), N: 1},
				effects.MillCards{N: 3},
			),
		}},
	})
}

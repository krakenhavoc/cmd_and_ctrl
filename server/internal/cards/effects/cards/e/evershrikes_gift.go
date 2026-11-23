package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Evershrike's Gift — Enchantment — Aura {W}:
//
//	"Enchant creature
//	 Enchanted creature gets +1/+0 and has flying.
//	 {1}{W}, Blight 2: Return this card from your graveyard to your hand.
//	 Activate only as a sorcery. (To blight 2, put two -1/-1 counters on a
//	 creature you control.)"
//
// Feather of Flight's two attachment statics, and Convenient Target's
// graveyard recursion with a blight in its cost (BlightN, #4684): the
// card is in the graveyard and the creature it blights is on the
// battlefield, any one the activator controls. With no creature to blight
// it can't be returned (CR 701.68b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5bc4adf-9105-4901-989c-63612314af5c",
		Name:         "Evershrike's Gift",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("flying"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{1}{W}, Blight 2: Return this card from your graveyard to your hand. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.ManaCost("{1}{W}"), effects.BlightN(2)),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Effect:       effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}

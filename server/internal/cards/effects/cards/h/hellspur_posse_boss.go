package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hellspur Posse Boss — Creature — Lizard Rogue {2}{R}{R}, 2/4 (Outlaws
// of Thunder Junction, #4468):
//
//	"Other outlaws you control have haste. (Assassins, Mercenaries,
//	 Pirates, Rogues, and Warlocks are outlaws.)
//	 When this creature enters, create two 1/1 red Mercenary creature
//	 tokens with '{T}: Target creature you control gets +1/+0 until end
//	 of turn. Activate only as a sorcery.'"
//
// The haste is a layer 6 lord grant over the outlaw types (CR 700.12):
// your creatures with any of the five types, read after the layers so a
// changeling or a creature given a type counts, never the Boss itself.
// The two Mercenaries are the set's shared token (MercenaryToken), and
// they are outlaws, so they enter able to attack and to tap for their
// own ability at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15260fa6-6062-423a-a248-81292ea46178",
		Name:         "Hellspur Posse Boss",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{
				Tribes:    []string{"Assassin", "Mercenary", "Pirate", "Rogue", "Warlock"},
				Others:    true,
				YoursOnly: true,
			}, "haste"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Hellspur Posse Boss — create two 1/1 red Mercenary creature tokens",
				effects.Do(effects.CreateToken{Template: effects.MercenaryToken(), N: 2})),
		},
	})
}

package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undercity Informer — Creature — Human Rogue {2}{B}, 2/3:
//
//	"{1}, Sacrifice a creature: Target player reveals cards from the
//	 top of their library until they reveal a land card, then puts
//	 those cards into their graveyard."
//
// Consuming Aberration's run (b14MillUntilLand) aimed at one target
// player: a real reveal, then the revealed cards, the land included,
// go to the graveyard in one mill. The Informer may sacrifice itself
// ("a creature", not "another").
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ccf5a0d4-69e8-4607-a76c-cfca336899e4",
		Name:         "Undercity Informer",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, Sacrifice a creature: Target player reveals cards from the top of their library until they reveal a land card, then puts those cards into their graveyard.",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.SacrificeACreature()),
			Targets: effects.TargetPlayer("target player"),
			Effect:  effects.TargetPlayerMillsUntilLand,
		}},
	})
}

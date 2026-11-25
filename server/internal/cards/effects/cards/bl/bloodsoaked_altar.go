package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodsoaked Altar — Artifact {4}{B}{B} (Core Set 2020):
//
//	"{T}, Pay 2 life, Discard a card, Sacrifice a creature: Create a 5/5 black
//	 Demon creature token with flying. Activate only as a sorcery."
//
// Skirsdag High Priest's Demon behind Susur Secundi's costs: tap, 2 life, a
// discard and a creature to sacrifice, all paid together (CR 601.2h), at
// sorcery speed. Nothing resolves if any part can't be paid, so a hand with no
// card to discard or a table with no creature to sacrifice can't activate it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "67631753-f37a-4f41-a6e6-1edead385acc",
		Name:         "Bloodsoaked Altar",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:        "{T}, Pay 2 life, Discard a card, Sacrifice a creature: Create a 5/5 black Demon creature token with flying. Activate only as a sorcery.",
			Cost:         effects.Plus(effects.TapCost(), effects.PayLife(2), effects.DiscardACard(), effects.SacrificeACreature()),
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("5/5 black Demon with flying"), N: 1}),
		}},
	})
}

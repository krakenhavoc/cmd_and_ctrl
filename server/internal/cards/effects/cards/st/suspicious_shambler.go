package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Suspicious Shambler — Creature — Zombie {3}{B}, 4/2 (Foundations,
// #4465):
//
//	"{4}{B}{B}, Exile this card from your graveyard: Create two 2/2 black
//	 Zombie creature tokens. Activate only as a sorcery."
//
// Eternal Student's graveyard ability (the exile is part of the cost, CR
// 602.2b, so the card is gone before the ability resolves) at Stone
// Docent's sorcery speed. It has no ability on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea6ee33f-0cba-4e96-817f-dc89c9064ead",
		Name:         "Suspicious Shambler",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label: "{4}{B}{B}, Exile this card from your graveyard: Create two 2/2 black Zombie creature tokens. Activate only as a sorcery.",
			Cost:  effects.Plus(effects.ManaCost("{4}{B}{B}"), effects.ExileThis()),
			Zones: []game.ZoneKind{game.ZoneGraveyard},
			// Sorcery speed: never offered while something waits on the
			// stack, so it declares no Answers (ADR 0142).
			SorcerySpeed: true,
			Effect:       effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 2}),
		}},
	})
}

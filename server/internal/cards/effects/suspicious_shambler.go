package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ea6ee33f-0cba-4e96-817f-dc89c9064ead",
		Name:         "Suspicious Shambler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{4}{B}{B}, Exile this card from your graveyard: Create two 2/2 black Zombie creature tokens. Activate only as a sorcery.",
			Cost:  Plus(ManaCost("{4}{B}{B}"), ExileThis()),
			Zones: []game.ZoneKind{game.ZoneGraveyard},
			// Sorcery speed: never offered while something waits on the
			// stack, so it declares no Answers (ADR 0142).
			SorcerySpeed: true,
			Effect:       Do(CreateToken{Template: TokenCard("2/2 black Zombie"), N: 2}),
		}},
	})
}

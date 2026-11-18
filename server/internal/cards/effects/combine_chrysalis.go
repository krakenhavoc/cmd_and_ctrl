package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Combine Chrysalis — Artifact {G}{U}:
//
//	"Creature tokens you control have flying.
//	 {2}{G}{U}, {T}, Sacrifice a token: Create a 4/4 green Beast creature
//	 token. Activate only as a sorcery."
//
// The flying grant is Brudiclad's layer-6 shape over creatureTokensYouControl,
// read through the layers, so a token that stops being a creature loses it.
// The activation is Arcade Cabinet's "Sacrifice a token" cost (CR 602.2b): any
// token you control, a noncreature one such as a Treasure included, paid as the
// ability is announced. The Beast comes whether or not the sacrificed token
// was a creature. The Chrysalis is an artifact, so its {T} needs no wait
// (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a1cee8d7-f619-47b4-ba8c-faf3106d4e41",
		Name:         "Combine Chrysalis",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{b16GrantKeywords(creatureTokensYouControl, "flying")},
		Activated: []ActivatedAbility{{
			Label:        "{2}{G}{U}, {T}, Sacrifice a token: Create a 4/4 green Beast creature token. Activate only as a sorcery.",
			Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
			Cost:         Plus(ManaCost("{2}{G}{U}"), TapCost(), b08SacrificeAToken()),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: TokenCard("4/4 green Beast"), N: 1}.Apply(NewContext(g, item))
			},
		}},
	})
}

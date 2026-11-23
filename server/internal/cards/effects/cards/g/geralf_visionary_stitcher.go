package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Geralf, Visionary Stitcher — Legendary Creature — Human Wizard {2}{U},
// 1/4 (Innistrad: Crimson Vow):
//
//	"Zombies you control have flying.
//	 {U}, {T}, Sacrifice another nontoken creature: Create an X/X blue
//	 Zombie creature token, where X is the sacrificed creature's toughness."
//
// The flying grant is the tribal one, so it reaches the Zombie tokens and
// any Zombie card, Geralf's own tokens included. X is the sacrificed
// creature's toughness as it last existed on the battlefield (CR 608.2h),
// floored at zero (CR 107.1b); a 0/0 Zombie then dies to CR 704.5f. The
// token is built by hand because its size is fixed at resolution (the
// printed token card is */*).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a0743b0-07c2-4e11-8b30-6b1b8156bdd0",
		Name:         "Geralf, Visionary Stitcher",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Zombie"}, YoursOnly: true}, "flying"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}, Sacrifice another nontoken creature: Create an X/X blue Zombie creature token, where X is the sacrificed creature's toughness.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost: effects.Plus(effects.ManaCost("{U}"), effects.TapCost(),
				effects.SacrificeAnotherN(1, "another nontoken creature", effects.Creature(), effects.Not(effects.IsTokenPredicate()))),
			Effect: effects.Do(effects.Instruction(func(ctx *effects.Context) error {
				x := 0
				if info, ok := ctx.SacrificedPermanent(); ok && info.Toughness > 0 {
					x = info.Toughness
				}
				return effects.CreateToken{Template: geralfVisionaryStitcherZombie(x), N: 1}.Apply(ctx)
			})),
		}},
	})
}

// geralfVisionaryStitcherZombie is the X/X blue Zombie. PrintedPTKnown,
// because 0/0 is a real size here (Corpse Cobble's Zombie).
func geralfVisionaryStitcherZombie(x int) game.Card {
	return game.Card{
		Name:           "Zombie",
		TypeLine:       "Token Creature — Zombie",
		Power:          x,
		Toughness:      x,
		PrintedPTKnown: true,
		Colors:         []string{"U"},
	}
}

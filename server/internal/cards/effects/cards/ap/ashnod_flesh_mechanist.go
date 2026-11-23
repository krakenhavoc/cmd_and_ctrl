package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ashnod, Flesh Mechanist — Legendary Creature — Human Artificer {B}, 1/1
// (The Brothers' War, S82):
//
//	"Deathtouch
//	 Whenever Ashnod attacks, you may sacrifice another creature. If you do,
//	 create a tapped Powerstone token.
//	 {5}, Exile a creature card from your graveyard: Create a tapped 3/3
//	 colorless Zombie artifact creature token."
//
// The attack trigger is maySacrificeOne (asked as it resolves, Ashnod
// excluded); the Powerstone follows only a creature that really went. The
// exile is a cost chosen on activation, so the token is made even if Ashnod
// has left by then.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7dfae95e-f67e-49e3-b68a-40ea7e5099dd",
		Name:            "Ashnod, Flesh Mechanist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Ashnod — you may sacrifice another creature to create a tapped Powerstone token",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MaySacrificeOne(g, item, "Ashnod — sacrifice another creature to create a tapped Powerstone token?",
						effects.Creature(), true, ashnodPowerstone)
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, Exile a creature card from your graveyard: Create a tapped 3/3 colorless Zombie artifact creature token.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.ExileFromGraveyard(1, "a creature card", effects.IsCreatureCard)),
			Effect: effects.Do(effects.CreateTokenAdvanced{
				Spec: effects.Token(effects.TokenCard("3/3 colorless Zombie artifact")).EntersTapped(),
				N:    1,
			}),
		}},
	})
}

// ashnodPowerstone is "if you do, create a tapped Powerstone token". An
// on-disk key.
var ashnodPowerstone = effects.RegisterSacrificedOne("prompt-run/ashnod-flesh-mechanist-powerstone", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1}.Apply(ctx)
})

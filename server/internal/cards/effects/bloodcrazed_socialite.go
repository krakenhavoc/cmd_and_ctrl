package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bloodcrazed Socialite — Creature — Vampire {3}{B}, 3/3 (Innistrad:
// Crimson Vow, #5495):
//
//	"Menace
//	 When this creature enters, create a Blood token.
//	 Whenever this creature attacks, you may sacrifice a Blood token. If
//	 you do, it gets +2/+2 until end of turn."
//
// The sacrifice is asked as the attack trigger resolves (maySacrificeOne);
// the pump is the Socialite's own, until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7e165b4e-4c8f-4a0f-bd6e-7b0a15b12d1d",
		Name:            "Bloodcrazed Socialite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Bloodcrazed Socialite — create a Blood token",
				Do(CreateToken{Template: BloodToken(), N: 1})),
			WheneverThisAttacks("Bloodcrazed Socialite — you may sacrifice a Blood token to get +2/+2 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return maySacrificeOne(g, item, "Bloodcrazed Socialite — sacrifice a Blood token? If you do, it gets +2/+2 until end of turn.",
						HasSubtype("Blood"), false, bloodcrazedSocialitePump)
				}),
		},
	})
}

var bloodcrazedSocialitePump = RegisterSacrificedOne("prompt-run/bloodcrazed-socialite-pump", func(ctx *Context, _ uuid.UUID) error {
	return BoostUntilEOT{Target: ctx.Source(), Power: 2, Toughness: 2, Label: "Bloodcrazed Socialite — +2/+2"}.Apply(ctx)
})

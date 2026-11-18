package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const catharsCallGrant = "cathars-call/end-step-human"

// Cathar's Call — Enchantment — Aura {2}{W} (Innistrad: Midnight Hunt):
//
//	"Enchant creature
//	 Enchanted creature has vigilance and "At the beginning of your end
//	 step, create a 1/1 white Human creature token.""
//
// Vigilance is GrantToAttached. The end-step trigger is a granted bundle
// (ADR 0093): the creature has it, so "your" end step is its controller's
// and the token is created under that player's control, not the Aura's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "433e92d2-899a-4eed-98b9-80de5d25494b",
		Name:         "Cathar's Call",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []AbilityGrant{{
			Key: catharsCallGrant,
			Triggered: []game.TriggeredAbility{
				AtYourEndStep("Cathar's Call — create a 1/1 white Human creature token",
					Do(CreateToken{Template: TokenCard("1/1 white Human"), N: 1})),
			},
			Text: "At the beginning of your end step, create a 1/1 white Human creature token.",
		}},
		Static: []game.StaticAbility{
			GrantToAttached("vigilance"),
			GrantAbilitiesToAttached(catharsCallGrant),
		},
	})
}

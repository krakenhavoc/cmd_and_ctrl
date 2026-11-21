package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "433e92d2-899a-4eed-98b9-80de5d25494b",
		Name:         "Cathar's Call",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: catharsCallGrant,
			Triggered: []game.TriggeredAbility{
				effects.AtYourEndStep("Cathar's Call — create a 1/1 white Human creature token",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human"), N: 1})),
			},
			Text: "At the beginning of your end step, create a 1/1 white Human creature token.",
		}},
		Static: []game.StaticAbility{
			effects.GrantToAttached("vigilance"),
			effects.GrantAbilitiesToAttached(catharsCallGrant),
		},
	})
}

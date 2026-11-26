package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gleam of Authority — Enchantment — Aura {1}{W} (Dragons of Tarkir, S82 old
// batch 330):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 for each +1/+1 counter on other creatures
//	 you control.
//	 Enchanted creature has vigilance and "{W}, {T}: Bolster 1.""
//
// Righteous Authority's per-count pump (PumpAttachedPer), counting the +1/+1
// counters on the Aura controller's other creatures (the enchanted one's own do
// not count), read live each layer pass. Vigilance is GrantToAttached and the
// bolster ability a granted activated ability (ADR 0093): the enchanted
// creature's controller activates it, so "you" is that player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68a91b87-1845-419b-aae7-ee1237554946",
		Name:         "Gleam of Authority",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: gleamOfAuthorityGrant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{W}, {T}: Bolster 1.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
				Effect:  effects.Do(effects.Bolster{N: 1}),
			}},
			Text: "{W}, {T}: Bolster 1.",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 1, effects.PlusOneCountersOnOtherCreaturesYouControl),
			effects.GrantToAttached("vigilance"),
			effects.GrantAbilitiesToAttached(gleamOfAuthorityGrant),
		},
	})
}

// gleamOfAuthorityGrant names the granted ability's bundle: an on-disk identity.
const gleamOfAuthorityGrant = "gleam-of-authority/tap-bolster"

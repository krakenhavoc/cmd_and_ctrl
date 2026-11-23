package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teferi's Talent — Enchantment — Aura {3}{U}{U}:
//
//	"Enchant planeswalker
//	 Enchanted planeswalker has "[−12]: You get an emblem with 'You
//	 may activate loyalty abilities of planeswalkers you control on
//	 any player's turn any time you could cast an instant.'"
//	 Whenever you draw a card, put a loyalty counter on enchanted
//	 planeswalker."
//
// Enchant planeswalker, through EnchantPlaneswalker, so the Aura
// falls off a host that stops being one (CR 704.5m).
//
// The −12 is a loyalty ability GRANTED to the enchanted planeswalker
// (ADR 0109 §2, owner decision 2): a layer-6 bundle, so it is the
// walker's ability. The walker pays the 12 (CR 606.6), and the row
// shares the walker's one loyalty activation a turn with its own
// (CR 606.3). The emblem is the Talent's (CR 114.2): the stack item
// names the Talent as its grantor, and CreateEmblem files the emblem
// under the grantor's key, which is the emblem declared here —
// Teferi, Temporal Archmage's, under this card's name.
//
// The draw trigger, once per card drawn. "Enchanted planeswalker"
// is read off the Aura at RESOLUTION, so an Aura destroyed in
// response still names the planeswalker it was on (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efc23669-1bbb-487f-94b3-cc07f3cd35f6",
		Name:         "Teferi's Talent",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPlaneswalker(),
		Grants: []effects.AbilityGrant{{
			Key: teferisTalentGrant,
			Activated: []effects.ActivatedAbility{{
				Label:  "−12: You get an emblem with \"You may activate loyalty abilities of planeswalkers you control on any player's turn any time you could cast an instant.\"",
				Cost:   effects.LoyaltyCost(-12),
				Effect: effects.Do(effects.CreateEmblem{}),
			}},
			Text: "[−12]: You get an emblem with \"You may activate loyalty abilities of planeswalkers you control on any player's turn any time you could cast an instant.\"",
		}},
		Emblem: effects.LoyaltyAtInstantSpeedEmblem("Teferi's Talent emblem"),
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(teferisTalentGrant)},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouDraw("Teferi's Talent — put a loyalty counter on enchanted planeswalker",
				effects.PutLoyaltyCounterOnEnchantedPlaneswalker),
		},
	})
}

const teferisTalentGrant = "teferis-talent/emblem"

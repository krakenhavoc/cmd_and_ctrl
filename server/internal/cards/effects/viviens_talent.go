package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vivien's Talent — Enchantment — Aura {1}{G}{G}:
//
//	"Enchant planeswalker
//	 Enchanted planeswalker has "[+1]: Look at the top four cards of
//	 your library. You may reveal a creature or land card from among
//	 them and put it into your hand. Put the rest on the bottom of your
//	 library in a random order."
//	 Whenever a nontoken creature you control enters, put a loyalty
//	 counter on enchanted planeswalker."
//
// The +1 is a loyalty ability granted to the enchanted planeswalker
// (ADR 0109 §2): the walker's, so it shares the walker's one
// loyalty activation a turn (CR 606.3). The dig is
// LookAtTopThenMayTakeToHand: a look, the taken card revealed, the
// rest to the bottom in a random order.
//
// The trigger is The Great Henge's condition; "enchanted
// planeswalker" is read off the Aura at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "968f0edb-b81f-465a-84b2-f077af61e51b",
		Name:         "Vivien's Talent",
		Completeness: CompletenessFull,
		Targets:      EnchantPlaneswalker(),
		Grants: []AbilityGrant{{
			Key: viviensTalentGrant,
			Activated: []ActivatedAbility{{
				Label: "+1: Look at the top four cards of your library. You may reveal a creature or land card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
				Cost:  LoyaltyCost(1),
				Effect: LookAtTopThenMayTakeToHand(4, Or(Creature(), Land()), 1,
					"Vivien's Talent — reveal a creature or land card and put it into your hand"),
			}},
			Text: "[+1]: Look at the top four cards of your library. You may reveal a creature or land card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(viviensTalentGrant)},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, nontokenCreatureEnteredUnderYourControl,
				"Vivien's Talent — put a loyalty counter on enchanted planeswalker",
				putLoyaltyCounterOnEnchantedPlaneswalker),
		},
	})
}

const viviensTalentGrant = "viviens-talent/dig"

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elspeth's Talent — Enchantment — Aura {2}{W}{W}:
//
//	"Enchant planeswalker
//	 Enchanted planeswalker has "[+1]: Create three 1/1 white Soldier
//	 creature tokens."
//	 Whenever you activate a loyalty ability of enchanted planeswalker,
//	 creatures you control get +2/+2 and gain vigilance until end of
//	 turn."
//
// The +1 is a loyalty ability granted to the enchanted planeswalker
// (ADR 0109 §2), sharing the walker's one loyalty activation a turn
// (CR 606.3).
//
// The trigger watches every loyalty ability of that walker, its own
// and the granted +1 alike. It goes on the stack above the ability
// (CR 603.3b) and resolves first, so the three Soldiers the +1 then
// makes are not pumped: "creatures you control" is fixed as the
// effect begins (CR 611.2c). The +2/+2 and the vigilance are one
// effect, one record.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "969b81d2-e95b-4a3f-9548-9394e3a6727e",
		Name:         "Elspeth's Talent",
		Completeness: CompletenessFull,
		Targets:      EnchantPlaneswalker(),
		Grants: []AbilityGrant{{
			Key: elspethsTalentGrant,
			Activated: []ActivatedAbility{{
				Label:  "+1: Create three 1/1 white Soldier creature tokens.",
				Cost:   LoyaltyCost(1),
				Effect: Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 3}),
			}},
			Text: "[+1]: Create three 1/1 white Soldier creature tokens.",
		}},
		Static: []game.StaticAbility{GrantAbilitiesToAttached(elspethsTalentGrant)},
		Triggered: []game.TriggeredAbility{
			WheneverYouActivateALoyaltyAbilityOfEnchanted(
				"Elspeth's Talent — creatures you control get +2/+2 and gain vigilance until end of turn",
				elspethsTalentPump),
		},
	})
}

const elspethsTalentGrant = "elspeths-talent/soldiers"

// elspethsTalentPump is "creatures you control get +2/+2 and gain
// vigilance until end of turn", one effect over the set fixed now.
func elspethsTalentPump(g *game.Game, item *game.StackItem) error {
	return untilEndOfTurn(NewContext(g, item), uuid.Nil, And(Creature(), YouControl()),
		"Elspeth's Talent — +2/+2 and vigilance",
		game.ModifyPTMod(2, 2), game.AddKeywordsMod("vigilance"))
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Discreet Retreat — Enchantment — Aura {3}{B} (Outlaws of Thunder Junction
// Commander, #5368):
//
//	"Enchant land
//	 Enchanted land has '{T}: Add two mana of any one color. Spend this mana
//	 only to cast outlaw spells or activate abilities of outlaw sources.'
//	 (Assassins, Mercenaries, Pirates, Rogues, and Warlocks are outlaws.)
//	 Whenever you cast your first outlaw spell each turn, you draw a card and
//	 you lose 1 life."
//
// The mana ability is granted to the enchanted land (Find the Path's shape),
// so it is the land's and its controller's. It adds two mana of ONE color
// (picked once, ProducedOneColor), each restricted by the same tag: cast an
// outlaw spell, or activate an ability of an outlaw source (Flamebraider's
// shape over the five outlaw subtypes). The trigger is the first outlaw spell
// its controller casts each turn, found from the event log; the Aura's
// controller draws and loses 1 life, in that order.
//
// No simplification.
func init() {
	const grant = "discreet-retreat/restricted-mana"
	Register(Spec{
		OracleID:     "0158634a-08a3-4eb8-8216-bf85161fd168",
		Name:         "Discreet Retreat",
		Completeness: CompletenessFull,
		Targets:      EnchantLand(),
		Grants: []AbilityGrant{{
			Key:  grant,
			Text: "{T}: Add two mana of any one color. Spend this mana only to cast outlaw spells or activate abilities of outlaw sources.",
			Mana: []ManaAbility{{
				Cost:           ManaAbilityCost{Tap: true},
				ProducedAmount: ProducedOneColor(func(*game.Game, uuid.UUID, uuid.UUID) int { return 2 }),
				Label:          "{T}: Add two mana of any one color. Spend this mana only to cast outlaw spells or activate abilities of outlaw sources",
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{game.ManaRestrictCast, game.ManaRestrictSubtype(cmdr2S33OutlawSubtypes)},
					[]string{game.ManaRestrictActivate, game.ManaRestrictSubtype(cmdr2S33OutlawSubtypes)},
				)},
			}},
		}},
		Static: []game.StaticAbility{GrantAbilities(AttachedToSource, grant)},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, cmdr2S33YouCastYourFirstOutlawSpellEachTurn,
				"Discreet Retreat — you draw a card and you lose 1 life", Do(DrawCards{N: 1}, GainLife{Amount: -1})),
		},
	})
}

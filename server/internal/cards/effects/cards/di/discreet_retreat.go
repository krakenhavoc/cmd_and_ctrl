package cardsdi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
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
	effects.Register(effects.Spec{
		OracleID:     "0158634a-08a3-4eb8-8216-bf85161fd168",
		Name:         "Discreet Retreat",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(),
		Grants: []effects.AbilityGrant{{
			Key:  grant,
			Text: "{T}: Add two mana of any one color. Spend this mana only to cast outlaw spells or activate abilities of outlaw sources.",
			Mana: []effects.ManaAbility{{
				Cost:           effects.ManaAbilityCost{Tap: true},
				ProducedAmount: effects.ProducedOneColor(func(*game.Game, uuid.UUID, uuid.UUID) int { return 2 }),
				Label:          "{T}: Add two mana of any one color. Spend this mana only to cast outlaw spells or activate abilities of outlaw sources",
				Restrictions: []string{game.ManaRestrictAnyOf(
					[]string{game.ManaRestrictCast, game.ManaRestrictSubtype(effects.Cmdr2S33OutlawSubtypes)},
					[]string{game.ManaRestrictActivate, game.ManaRestrictSubtype(effects.Cmdr2S33OutlawSubtypes)},
				)},
			}},
		}},
		Static: []game.StaticAbility{effects.GrantAbilities(effects.AttachedToSource, grant)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.Cmdr2S33YouCastYourFirstOutlawSpellEachTurn,
				"Discreet Retreat — you draw a card and you lose 1 life", effects.Do(effects.DrawCards{N: 1}, effects.GainLife{Amount: -1})),
		},
	})
}

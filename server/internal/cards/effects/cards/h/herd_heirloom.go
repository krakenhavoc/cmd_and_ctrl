package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Herd Heirloom — Artifact {1}{G}:
//
//	"{T}: Add one mana of any color. Spend this mana only to cast a
//	 creature spell.
//	 {T}: Until end of turn, target creature you control with power 4
//	 or greater gains trample and 'Whenever this creature deals combat
//	 damage to a player, draw a card.'"
//
// The mana ability is Ancient Ziggurat's restricted-spend shape
// (ManaRestrictCast + ManaRestrictType("Creature")) with a full colour
// pipe instead of a fixed colour. Both tags, because they AND: the type
// tag alone also matches an ACTIVATION whose source is a creature
// (ManaSpendForAbility reads the source's types), and "only to cast a
// creature spell" does not pay a creature's ability (#2059).
//
// The second ability is a duration grant (ADR 0093 PR 4, #1584): the
// target gains trample and a catalog bundle carrying the draw trigger,
// as one ScopedEffect record at one timestamp. The trigger is the
// CREATURE's, so its controller draws, and it is gone at the cleanup
// step. The power-4 test is a target clause, checked at activation and
// again at resolution (CR 608.2b).
//
// No simplification.
const herdHeirloomDraw = "herd-heirloom/draw"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "78b6dd40-c182-4037-a4d3-9fd012b2c584",
		Name:         "Herd Heirloom",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:         effects.ManaAbilityCost{Tap: true},
			Produced:     "{W|U|B|R|G}",
			Restrictions: []string{effects.ManaRestrictCast, effects.ManaRestrictType("Creature")},
			Label:        "Add one mana of any color. Spend this mana only to cast a creature spell",
		}},
		Grants: []effects.AbilityGrant{{
			Key: herdHeirloomDraw,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Herd Heirloom — draw a card", func(g *game.Game, item *game.StackItem) error {
					return effects.DrawCards{Player: item.Controller, N: 1}.Apply(effects.NewContext(g, item))
				}),
			},
			Text: "Whenever this creature deals combat damage to a player, draw a card.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Until end of turn, target creature you control with power 4 or greater gains trample and \"Whenever this creature deals combat damage to a player, draw a card.\"",
			Cost:    effects.TapCost(),
			Targets: effects.TargetCreature("target creature you control with power 4 or greater", effects.YouControl(), effects.PowerGE(4)),
			Effect:  herdHeirloomGrant,
		}},
	})
}

// herdHeirloomGrant is the activated ability's effect: the target, if
// still legal, gains trample and the draw trigger until end of turn.
func herdHeirloomGrant(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	target, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	return effects.GrantAbilitiesFor{
		Target: target,
		Keys:   []string{herdHeirloomDraw},
		Also:   []game.Mod{game.AddKeywordsMod("trample")},
		Label:  "Herd Heirloom — trample and draw on combat damage",
	}.Apply(ctx)
}

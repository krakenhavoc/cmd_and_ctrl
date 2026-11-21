package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon-Cursed Halls — Land (The Hobbit):
//
//	"{T}: Add {C}.
//	 {1}, {T}: Until end of turn, target creature gains \"Whenever this
//	 creature deals combat damage to a player, create a Treasure token.\""
//
// The mana ability is the plain colourless tap. The second ability is
// Blitzball Stadium's grant: a layer-6 bundle with one triggered ability
// (WheneverThisDealsCombatDamageToAPlayer), applied to the target with
// GrantAbilitiesFor, which lasts until end of turn. Because "this creature"
// is the recipient, the Treasure goes to the recipient's controller when
// that creature deals combat damage to a player, and each instance of
// combat damage triggers once (a creature that deals first-strike and
// regular combat damage makes two).
//
// No simplification.
const dragonCursedHallsGrant = "dragon-cursed-halls/treasure-on-combat-damage"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "5be7a4d5-33b7-464b-8851-d4ad35302e62",
		Name:         "Dragon-Cursed Halls",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Grants: []effects.AbilityGrant{{
			Key: dragonCursedHallsGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Dragon-Cursed Halls — create a Treasure token",
					effects.Do(effects.CreateToken{Template: effects.TreasureToken(), N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, create a Treasure token.",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Until end of turn, target creature gains \"Whenever this creature deals combat damage to a player, create a Treasure token.\"",
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantAbilitiesFor{
					Target: id,
					Keys:   []string{dragonCursedHallsGrant},
					Label:  "Dragon-Cursed Halls — makes a Treasure when it deals combat damage to a player",
				}.Apply(ctx)
			},
		}},
	})
}

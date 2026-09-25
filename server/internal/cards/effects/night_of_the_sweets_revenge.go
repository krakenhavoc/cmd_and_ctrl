package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Night of the Sweets' Revenge — Enchantment {3}{G}:
//
//	"When this enchantment enters, create a Food token. (It's an
//	 artifact with '{2}, {T}, Sacrifice this token: You gain 3 life.')
//	 Foods you control have '{T}: Add {G}.'
//	 {5}{G}{G}, Sacrifice this enchantment: Creatures you control get
//	 +X/+X until end of turn, where X is the number of Foods you
//	 control. Activate only as a sorcery."
//
// Three printed abilities, all buildable: the ETB Food, the ADR 0093
// grant scoped to a token subtype instead of a class or an
// attachment, and a sacrifice ability whose pump amount is read off
// the board at resolution — after the enchantment itself, which
// makes the Foods, has already left.
//
// No simplification.
func init() {
	const grant = "night-of-the-sweets-revenge/green-mana"
	Register(Spec{
		OracleID:     "4bcf4f12-470e-4f73-a19a-643a20084e16",
		Name:         "Night of the Sweets' Revenge",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key:  grant,
			Text: "{T}: Add {G}.",
			Mana: []ManaAbility{{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{G}",
				Label:    "Add {G}",
			}},
		}},
		Static: []game.StaticAbility{
			GrantAbilities(foodsYouControl, grant),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Night of the Sweets' Revenge — create a Food", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{5}{G}{G}, Sacrifice this enchantment: Creatures you control get +X/+X until end of turn, where X is the number of Foods you control.",
			Cost:         Plus(ManaCost("{5}{G}{G}"), SacrificeThis()),
			SorcerySpeed: true,
			Effect:       nightOfTheSweetsRevengePump,
		}},
	})
}

// foodsYouControl is "Foods you control" — Night of the Sweets'
// Revenge.
func foodsYouControl(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.Controller == source.Controller && target.HasSubtype("Food")
}

// nightOfTheSweetsRevengePump counts the controller's Foods AT
// RESOLUTION — after the enchantment that made them has already been
// sacrificed to pay the cost — and pumps every creature the
// controller controls by that much until end of turn.
func nightOfTheSweetsRevengePump(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	controller := item.Controller
	x := 0
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtype("Food") {
			x++
		}
	}
	if x == 0 {
		return nil
	}
	return BoostUntilEOT{
		Match:     And(Creature(), ControlledBy(controller)),
		Power:     x,
		Toughness: x,
		Label:     "Night of the Sweets' Revenge — pump until end of turn",
	}.Apply(ctx)
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gilded Goose — Creature — Bird {G}, 0/2 (EDHREC rank 796):
//
//	"Flying
//	 When this creature enters, create a Food token.
//	 {1}{G}, {T}: Create a Food token.
//	 {T}, Sacrifice a Food: Add one mana of any color."
//
// A one-drop that ramps once for free and fixes forever after. Four
// printed abilities, four mechanisms, all real:
//
//   - Flying rides PrintedKeywords.
//   - The ETB is a real trigger (a response window), making the Food
//     template every Food-maker shares.
//   - The Food-maker is a CR 602 ability with a mana and a tap
//     component.
//   - The mana ability's cost is tap plus "sacrifice a Food" — the
//     same SacrificeOther clause Phyrexian Tower's creature uses, with
//     the shared HasSubtype predicate (post-layer subtypes, so a Food
//     made by a copy effect counts) — so the Goose cannot eat itself,
//     and "any color" means no commander-identity narrowing.
//
// CR 302.1 summoning sickness applies to both tap abilities, as
// printed; the ETB Food is why the card is still a one-drop ramp
// spell on turn one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f2f09757-1931-47c0-a5f0-39280445489d",
		Name:            "Gilded Goose",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Gilded Goose — create a Food", Do(CreateToken{Template: FoodToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label: "{1}{G}, {T}: Create a Food token.",
			Cost:  Plus(ManaCost("{1}{G}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return CreateToken{Controller: item.Controller, Template: FoodToken(), N: 1}.Apply(NewContext(g, item))
			},
		}},
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{
				Tap:            true,
				SacrificeOther: sacrificeSpec("a Food", HasSubtype("Food")),
			},
			Produced:                "{W|U|B|R|G}",
			Label:                   "{T}, Sacrifice a Food: Add one mana of any color",
			IgnoreCommanderIdentity: true,
		}},
	})
}

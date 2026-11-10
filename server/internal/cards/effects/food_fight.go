package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Food Fight — Enchantment {1}{R} (Wilds of Eldraine, #4471):
//
//	"Artifacts you control have "{2}, Sacrifice this artifact: It deals
//	 damage to any target equal to 1 plus the number of permanents named
//	 Food Fight you control.""
//
// The grant is Galazeth Prismari's layer-6 bundle (ADR 0093) over every
// artifact you control, with a granted activated ability in place of a
// mana ability. Sacrificing the artifact is a cost, so the damage is dealt
// by the artifact as it last existed, and the number of Food Fights you
// control is counted as the ability resolves (the Food Fight's controller
// is the ability's controller, who activated it). Each Food Fight grants
// the ability, so two of them give each artifact two abilities, and each
// one counts both enchantments.
//
// No simplification.
const foodFightGrant = "food-fight/sacrifice-for-damage"

func init() {
	Register(Spec{
		OracleID:     "00571e7f-4e40-4a07-a529-34580e942338",
		Name:         "Food Fight",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: foodFightGrant,
			Activated: []ActivatedAbility{{
				Label:   "{2}, Sacrifice this artifact: It deals damage to any target equal to 1 plus the number of permanents named Food Fight you control.",
				Cost:    Plus(ManaCost("{2}"), SacrificeThis()),
				Targets: TargetAny(),
				Effect:  std54FoodFightDamage,
			}},
			Text: "{2}, Sacrifice this artifact: It deals damage to any target equal to 1 plus the number of permanents named Food Fight you control.",
		}},
		Static: []game.StaticAbility{Labelled(GrantAbilities(func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.IsArtifact() && target.Controller == source.Controller
		}, foodFightGrant), "Artifacts you control have \"{2}, Sacrifice this artifact: It deals damage to any target equal to 1 plus the number of permanents named Food Fight you control.\"")},
	})
}

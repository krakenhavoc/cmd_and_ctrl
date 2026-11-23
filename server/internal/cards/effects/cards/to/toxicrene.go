package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Toxicrene — Creature — Tyranid {3}{G}, 2/4:
//
//	"Reach, deathtouch
//	 Hypertoxic Miasma — All lands have '{T}: Add one mana of any color' and
//	 lose all other abilities."
//
// One layer-6 effect over every land, whoever controls it: it removes every
// ability and grants the any-color mana ability back (CR 613.1f, the
// Alpine Moon shape). No land type changes, so the lands keep their types.
const toxicreneAnyColor = "toxicrene/any-color"

func toxicreneApplies(target *game.Card, _ *game.Game, _ *game.Card) bool { return target.IsLand() }

func init() {
	effects.Register(effects.Spec{
		OracleID:        "fbfe98f0-110b-4d4b-9240-2767b08acf68",
		Name:            "Toxicrene",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "deathtouch"},
		Grants:          []effects.AbilityGrant{effects.AnyColorManaGrant(toxicreneAnyColor)},
		Static: []game.StaticAbility{
			effects.Labelled(effects.LosesAllAbilitiesAndHasIncludingLandMana(toxicreneApplies, toxicreneAnyColor), "All lands have \"{T}: Add one mana of any color\" and lose all other abilities."),
		},
	})
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Breath of Dreams — Enchantment {2}{U}{U}:
//
//	"Cumulative upkeep {U}
//	 Green creatures have 'Cumulative upkeep {1}.'"
//
// #3869. Its own cumulative upkeep is the ordinary constructor; the
// second line is a layer-6 grant of the CumulativeUpkeepGrant bundle to
// every green creature, whoever controls it (ADR 0093). Each creature's
// controller is asked at their own upkeep and sacrifices the creature on
// a decline. A green creature with its own cumulative upkeep triggers
// once per instance (CR 702.24b).
//
// "Green" is read through the source's word lens (TextColor), so a
// Sleight of Mind on Breath of Dreams changes which creatures it reaches.
//
// No simplification.
const breathOfDreamsGrant = "breath-of-dreams/cumulative-upkeep-1"

func init() {
	Register(Spec{
		OracleID:     "477f5011-aaeb-4a5f-8a65-5fb45cb06389",
		Name:         "Breath of Dreams",
		Completeness: CompletenessFull,
		Grants:       []AbilityGrant{CumulativeUpkeepGrant(breathOfDreamsGrant, "{1}")},
		Triggered: []game.TriggeredAbility{
			CumulativeUpkeep("Breath of Dreams — cumulative upkeep {U}", "{U}"),
		},
		Static: []game.StaticAbility{
			Labelled(GrantAbilities(breathOfDreamsGreenCreature, breathOfDreamsGrant), "Green creatures have \"Cumulative upkeep {1}.\""),
		},
	})
}

// breathOfDreamsGreenCreature is "green creatures": every creature on
// the battlefield, any controller, whose colours include the source's
// word for green.
func breathOfDreamsGreenCreature(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.HasColor(source.TextColor("G"))
}

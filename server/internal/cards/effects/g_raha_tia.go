package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// G'raha Tia — Legendary Creature — Cat Archer {4}{W}, 3/5 (Final
// Fantasy, #4462):
//
//	"Reach
//	 The Allagan Eye — Whenever one or more other creatures and/or
//	 artifacts you control die, draw a card. This ability triggers only
//	 once each turn."
//
// Homicide Investigator's batch trigger (CR 603.2c) over a creature
// and/or an artifact: a creature or artifact permanent you controlled as
// it died, read from its last-known types (CR 603.10a), the Tia herself
// excluded. A wrath that kills creatures and artifacts together is one
// trigger, and a second batch the same turn finds the ability spent
// (TriggersOnlyOnceEachTurn).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0c4603be-d71a-4d33-b62c-04ead1987dbe",
		Name:            "G'raha Tia",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			TriggersOnlyOnceEachTurn(OncePerBatch(On(game.EventLTB, grahaOtherCreatureOrArtifactYouControlDied,
				"G'raha Tia — The Allagan Eye: draw a card (other creatures and/or artifacts you control died)",
				Do(DrawCards{N: 1})))),
		},
	})
}

// grahaOtherCreatureOrArtifactYouControlDied is one death in the batch.
func grahaOtherCreatureOrArtifactYouControlDied(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard || ev.CardID == source.InstanceID {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || leftUnderControlOf(ev, c) != source.Controller {
		return false
	}
	return leftAsType(ev, c, "creature") || leftAsType(ev, c, "artifact")
}

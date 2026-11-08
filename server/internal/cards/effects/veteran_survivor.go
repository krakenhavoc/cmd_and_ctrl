package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Veteran Survivor — Creature — Human Survivor {W}, 2/1 (Duskmourn:
// House of Horror, #4466):
//
//	"Survival — At the beginning of your second main phase, if this
//	 creature is tapped, exile up to one target card from a graveyard.
//	 As long as there are three or more cards exiled with this creature,
//	 it gets +3/+3 and has hexproof."
//
// The trigger is survival.go's (Savior of the Small's shape) with
// Afterthought Sentry's "up to one target card from a graveyard": with no
// card chosen it does nothing, and a card that has left by resolution is
// skipped (CR 608.2b).
//
// The two abilities are linked (CR 607.2a): "cards exiled with this
// creature" are the ones its Survival trigger exiled and that are still
// in exile, read off the event log by the trigger's label (b27ExiledWith,
// Unlicensed Hearse's record). A Survivor that leaves and comes back is a
// new object and counts from zero (CR 400.7). The static is a layer 7c
// pump and a layer 6 hexproof grant on that live count; DependsOnExile
// re-reads it when a card leaves exile.
//
// No simplification.
func init() {
	pump := game.StaticAbility{
		Label:          "Gets +3/+3 as long as there are three or more cards exiled with it.",
		Layer:          game.Layer7PT,
		SubLayer:       game.SubLayer7C_Modify,
		DependsOnExile: true,
		AppliesTo:      veteranSurvivorThreeExiled,
		Apply:          pumpBy(3, 3),
	}
	hexproof := KeywordGrant(veteranSurvivorThreeExiled, "hexproof")
	hexproof.Label = "Has hexproof as long as there are three or more cards exiled with it."
	hexproof.DependsOnExile = true
	Register(Spec{
		OracleID:     "cc5b1276-d13e-4392-988d-1315bc90a9b1",
		Name:         "Veteran Survivor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(Survival(veteranSurvivorExileLabel, b27ExileChosenTarget),
				TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1)),
		},
		Static: []game.StaticAbility{pump, hexproof},
	})
}

// veteranSurvivorExileLabel is the Survival trigger's label, the key its
// exiled-with record is read by.
const veteranSurvivorExileLabel = "Veteran Survivor — exile up to one target card from a graveyard"

// veteranSurvivorThreeExiled is "this creature, as long as there are three
// or more cards exiled with it".
func veteranSurvivorThreeExiled(target *game.Card, g *game.Game, source *game.Card) bool {
	return selfCreature(target, g, source) &&
		len(b27ExiledWith(g, source.InstanceID, veteranSurvivorExileLabel)) >= 3
}

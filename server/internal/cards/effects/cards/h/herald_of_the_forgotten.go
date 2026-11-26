package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Herald of the Forgotten — {6}{W}{W} Creature — Cat Beast 6/6 (Commander 2020):
//
//	"Flying
//	 When this creature enters, if you cast it, return any number of target
//	 permanent cards with cycling abilities from your graveyard to the
//	 battlefield."
//
// Cresting Mosasaurus' "if you cast it" (the one catalog reading, checked as the
// trigger would fire and again as it resolves, CR 603.4) and Rooting Moloch's "a
// card with a cycling ability" (typecycling counts, CR 702.29e). The targets are
// any number (zero is allowed), re-checked as it resolves (CR 608.2b), and come
// back together (ReturnFromGraveyardTogether), so an Aura chooses its host as it
// enters.
func init() {
	t := effects.On(game.EventETB, effects.AllOf(effects.Self, effects.CrestingMosasaurusYouCastIt),
		"Herald of the Forgotten — return any number of target permanent cards with cycling abilities from your graveyard to the battlefield",
		effects.RalZarekReanimate)
	t.Targets = effects.TargetCardInGraveyard("any number of target permanent cards with cycling abilities from your graveyard",
		effects.Permanent(), effects.YouOwn(), effects.GapB158HasCyclingAbility).WithCount(0, 0)
	effects.Register(effects.Spec{
		OracleID:        "62855d00-4a6b-45a5-a1c0-2e7699e821cb",
		Name:            "Herald of the Forgotten",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{t},
	})
}

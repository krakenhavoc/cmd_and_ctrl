package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gurgling Anointer — Creature — Phyrexian Horror {2}{B}, 1/3 (The
// Brothers' War, #5482):
//
//	"Flying
//	 Whenever you draw your second card each turn, put a +1/+1 counter
//	 on this creature.
//	 When this creature dies, return another target creature card with
//	 mana value less than or equal to this creature's power from your
//	 graveyard to the battlefield."
//
// "This creature's power" is its last-known power, counters included
// (CR 608.2h, LastKnownPermanentForEffect): a three-counter Anointer
// reaches a mana value 4 card. The ceiling is read when the target is
// chosen as the trigger goes on the stack and the legal set is checked
// again as it resolves (CR 608.2b). Mana value is the card's printed one
// in the graveyard (CR 202.3b). "Another" excludes the Anointer itself by
// instance, as Gixian Puppeteer's reanimation does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "22fa5324-ed3b-4b97-89b4-31c1f86d6ead",
		Name:            "Gurgling Anointer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.S82BroB07SecondDrawCounter("Gurgling Anointer"),
			gurglingAnointerDiesTrigger(),
		},
	})
}

func gurglingAnointerDiesTrigger() game.TriggeredAbility {
	t := effects.WhenThisDies("Gurgling Anointer — return another target creature card with mana value less than or equal to its power from your graveyard to the battlefield",
		effects.HobB8ReturnTargetGraveyardCardToBattlefield)
	t.TargetsFrom = func(_ game.TriggerContext, source *game.Card, g *game.Game) *game.TargetSpec {
		self, ceiling := source.InstanceID, 0
		if info, ok := g.LastKnownPermanentForEffect(self); ok {
			ceiling = max(info.Power, 0)
		}
		return effects.TargetCardInGraveyard("another target creature card with mana value less than or equal to Gurgling Anointer's power from your graveyard",
			effects.YouOwn(), effects.Creature(), effects.ManaValueLE(ceiling), effects.NotSelf(self))
	}
	return t
}

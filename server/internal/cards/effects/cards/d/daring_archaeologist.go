package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Daring Archaeologist — Creature — Human Artificer {3}{W}, 3/3 (Dominaria):
//
//	"When this creature enters, you may return target artifact card from your
//	 graveyard to your hand.
//	 Whenever you cast a historic spell, put a +1/+1 counter on this creature.
//	 (Artifacts, legendaries, and Sagas are historic.)"
//
// The enters trigger is Carrion Imp's optional graveyard target, chosen as the
// trigger goes on the stack and the "may" asked as it resolves. The cast trigger
// is Jhoira's historic test on a spell you cast (an artifact, a legendary or a
// Saga, as it sits on the stack).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9d89f7c-a362-48dc-a335-f0e556e83f35",
		Name:         "Daring Archaeologist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.Targeting(
				effects.WhenThisEnters("Daring Archaeologist — you may return target artifact card from your graveyard to your hand",
					effects.S82Mom16ReturnTargetToHand),
				effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.Artifact(), effects.YouOwn())),
				"Daring Archaeologist — return the artifact card to your hand?"),
			effects.On(game.EventCast, effects.YouCastAHistoricSpell,
				"Daring Archaeologist — put a +1/+1 counter on it",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}

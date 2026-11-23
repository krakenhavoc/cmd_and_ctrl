package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dread Tiller — Artifact Creature — Scarecrow {1}{B}{G}, 2/4 (Lorwyn
// Eclipsed Commander, #5368):
//
//	"When this creature enters, put a -1/-1 counter on target creature.
//	 Whenever a creature with a -1/-1 counter on it dies, you may put a
//	 land card from your hand or graveyard onto the battlefield tapped."
//
// The dies trigger is any creature's, the Tiller's own included, and reads
// the -1/-1 counter the creature last had (CR 603.10a, b29CreatureWithMinusCounterDied).
// The "you may" is the pick itself, made as the ability resolves over the
// lands in the hand and the graveyard together, none being the answer. The
// land is put, not played (CR 305.4), so it does not use the land drop.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ffcffad7-6c33-4265-82c9-04045fdf371b",
		Name:         "Dread Tiller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Dread Tiller — put a -1/-1 counter on target creature", effects.Cmdr2B25CountersOnTarget(1)),
				effects.Harmful(effects.TargetCreature("target creature")),
			),
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B29CreatureWithMinusCounterDied(ev, g)
			}, "Dread Tiller — you may put a land card from your hand or graveyard onto the battlefield tapped",
				effects.Cmdr2B25PutLandFromHandOrGraveyardTapped("Dread Tiller")),
		},
	})
}

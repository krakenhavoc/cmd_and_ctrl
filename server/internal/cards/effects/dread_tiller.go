package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ffcffad7-6c33-4265-82c9-04045fdf371b",
		Name:         "Dread Tiller",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Dread Tiller — put a -1/-1 counter on target creature", cmdr2B25CountersOnTarget(1)),
				Harmful(TargetCreature("target creature")),
			),
			On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b29CreatureWithMinusCounterDied(ev, g)
			}, "Dread Tiller — you may put a land card from your hand or graveyard onto the battlefield tapped",
				cmdr2B25PutLandFromHandOrGraveyardTapped("Dread Tiller")),
		},
	})
}

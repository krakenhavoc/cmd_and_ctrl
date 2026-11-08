package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vengeful Townsfolk — Creature — Human Citizen {2}{W}, 3/3 (Outlaws of
// Thunder Junction, #4468):
//
//	"Whenever one or more other creatures you control die, put a +1/+1
//	 counter on this creature."
//
// Vraan, Executioner Thane's trigger without the once-each-turn limit:
// "one or more" is a batch (CR 603.2d), so a sweeper that kills three of
// your other creatures is one trigger and one counter (OncePerBatch),
// while three separate deaths are three. "Other" is object identity: the
// Townsfolk's own death does not trigger it, and when it dies alongside
// the others the trigger still happens but has nothing to put the
// counter on (CR 400.7). The controller is the one each creature had as
// it died (CR 603.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c649d5ae-2f38-4737-8123-8069a2ba0bde",
		Name:         "Vengeful Townsfolk",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventLTB, vengefulTownsfolkOtherCreatureDied,
				"Vengeful Townsfolk — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}

// vengefulTownsfolkOtherCreatureDied is "another creature you control
// died", one event of the batch.
func vengefulTownsfolkOtherCreatureDied(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	return ev.CardID != source.InstanceID && ACreatureYouControlDied(ev, source, lki, g)
}

package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Al Bhed Salvagers — Creature — Human Artificer Warrior {2}{B}, 2/3
// (Final Fantasy, #4462):
//
//	"Whenever this creature or another creature or artifact you control
//	 dies, target opponent loses 1 life and you gain 1 life."
//
// One trigger, two conditions: the Salvagers' own death, and another
// creature or artifact its controller controlled being put into a
// graveyard from the battlefield (CR 700.4), read from the departure
// (CR 603.10a) so a crewed Vehicle counts as a creature and a token
// counts. The drain is Blood Artist's body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "43f81514-e812-4e17-9b1a-704c136d8ffd",
		Name:         "Al Bhed Salvagers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return ThisDied(ev, source, lki, g) || AnotherCreatureOrArtifactYouControlWasPutIntoAGraveyard(ev, source, lki, g)
			},
			Targets: TargetPlayer("target opponent", Opponent()),
			Key:     "Al Bhed Salvagers — target opponent loses 1 life and you gain 1 life",
			Purpose: game.Purpose{DeathPayoff: true},
			Effect:  Do(TargetPlayerLosesLife{N: 1}, GainLife{Amount: 1}),
		}},
	})
}

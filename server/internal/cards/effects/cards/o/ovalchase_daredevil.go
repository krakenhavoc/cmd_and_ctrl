package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ovalchase Daredevil — Creature — Human Pilot {3}{B}, 4/2 (Kaladesh, #5539):
//
//	"Whenever an artifact you control enters, you may return this card from
//	 your graveyard to your hand."
//
// A graveyard-zone trigger (Zones, CR 113.6k) on Weldfast Wingsmith's
// "artifact entered under your control" watcher; the Daredevil is a creature,
// not an artifact, so it never triggers off itself. "You control" reads the
// Daredevil's owner while it is in the graveyard. The "you may" is asked as
// the ability resolves (CR 603.5), and a Daredevil that has left the graveyard
// by then is not returned (CR 400.7).
//
// No simplification.
func init() {
	t := effects.Optional(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return effects.IsArtifactEnteredUnderYourControl(ev, source, g)
	}, "Ovalchase Daredevil — you may return this card from your graveyard to your hand", effects.ReturnThisCardFromYourGraveyardToYourHand),
		"Ovalchase Daredevil — return this card from your graveyard to your hand?")
	t.Zones = []game.ZoneKind{game.ZoneGraveyard}
	effects.Register(effects.Spec{
		OracleID:     "a33961bb-fff0-4f4a-aa35-f904dbf2b075",
		Name:         "Ovalchase Daredevil",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

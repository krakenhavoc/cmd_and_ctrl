package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burning Vengeance — Enchantment {2}{R} (Innistrad Remastered):
//
//	"Whenever you cast a spell from your graveyard, this enchantment deals
//	 2 damage to any target."
//
// The cast event records the zone the spell was cast from (Event.OldZone),
// so flashback, Snapcaster-style grants and Yawgmoth's Will all count and a
// spell cast from hand does not.
func init() {
	t := effects.On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return ev.Kind == game.EventCast && ev.Actor == source.Controller && ev.OldZone == game.ZoneGraveyard
	}, "Burning Vengeance — 2 damage to any target", effects.GapTwoDamageToChosenTarget)
	t.Targets = effects.TargetAny()
	effects.Register(effects.Spec{
		OracleID:     "ef1fb174-9a34-43d0-8caf-9c1938147a12",
		Name:         "Burning Vengeance",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}

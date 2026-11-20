package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scalelord Reckoner — Creature — Dragon {3}{W}{W}, 4/4:
//
//	"Flying
//	 Whenever a Dragon you control becomes the target of a spell or ability
//	 an opponent controls, destroy target nonland permanent that player
//	 controls."
//
// Black Bolt, Inhuman King's shape (#4707) with the watched object widened
// from the source to any Dragon its controller controls: the trigger reads the
// event's Actor (the controller of the targeting spell or ability) and its
// second clause is limited to that player's nonland permanents
// (ControlledByThatEventActor). The Dragon is read on the battlefield as it
// was targeted (a Dragon spell is not a "Dragon you control" permanent).
func init() {
	Register(Spec{
		OracleID:        "9da65d66-71f0-4e17-a1f9-6da6cd1cfa8a",
		Name:            "Scalelord Reckoner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventBecomesTarget, scalelordReckonerDragonTargeted,
					"Scalelord Reckoner — destroy target nonland permanent that player controls",
					destroyFirstLegalCardTarget),
				ControlledByThatEventActor(TargetPermanent("target nonland permanent that player controls", Nonland()))),
		},
	})
}

func scalelordReckonerDragonTargeted(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventBecomesTarget || source == nil || ev.CardID == uuid.Nil ||
		ev.Actor == uuid.Nil || ev.Actor == source.Controller || g.PlayerByIDForEffect(ev.Actor) == nil {
		return false
	}
	if !onBattlefield(g, ev.CardID) {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.Controller == source.Controller && c.HasSubtype(source.TextSubtype("Dragon"))
}

package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tamanoa — Creature — Spirit {R}{G}{W}, 2/4 (Coldsnap):
//
//	"Whenever a noncreature source you control deals damage, you gain that
//	 much life."
//
// One trigger per damage event, gaining the amount that event dealt (CR 120.3,
// CR 603.2c). The source is a noncreature permanent or spell as it was when it
// dealt the damage: a permanent that has since left the battlefield (a
// sacrifice cost) is read from its last-known information (CR 608.2h). A
// creature that is not a creature any longer counts, and one that became a
// creature does not. Damage prevented to zero does not trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14440669-ddf0-46ac-b140-13948a8113b8",
		Name:         "Tamanoa",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, tamanoaNoncreatureSourceYouControlDealt,
				"Tamanoa — you gain that much life",
				effects.S82OldB562GainThatMuchLife),
		},
	})
}

func tamanoaNoncreatureSourceYouControlDealt(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventDealDamage || ev.Amount <= 0 || ev.Source == uuid.Nil {
		return false
	}
	if c, ok := g.LookupCardForEffect(ev.Source); ok {
		if z := g.FindCardZoneForEffect(ev.Source); z != nil && z.Kind == game.ZoneBattlefield {
			return c.Controller == source.Controller && !c.IsCreature()
		}
		// A spell on the stack (or any other source still findable).
		return c.Controller == source.Controller && !c.IsCreature()
	}
	if info, ok := g.LastKnownPermanentForEffect(ev.Source); ok {
		return info.Controller == source.Controller && !effects.HasFold(info.Characteristic.Types, "creature")
	}
	return false
}

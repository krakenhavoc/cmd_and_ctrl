package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// attacks_you_or_a_planeswalker_you_control.go — the delayed condition
// Jace, Reality Sculptor, Tamiyo, Inquisitive Student and Tamiyo Meets the
// Story Circle share.

// JaceRealitySculptorAttacksYouCondition is "a creature attacks you or
// a planeswalker you control": the declared defender (Event.Target) is
// the trigger's controller, or a planeswalker that player controls.
var JaceRealitySculptorAttacksYouCondition = game.DelayedCondition("attack/attacks-you-or-a-planeswalker-you-control",
	func(ev game.Event, dt *game.DelayedTrigger, g *game.Game, _ game.EffectParams) bool {
		if ev.Kind != game.EventAttack || ev.Target == uuid.Nil {
			return false
		}
		if ev.Target == dt.Controller {
			return true
		}
		pw, ok := g.LookupCardForEffect(ev.Target)
		return ok && pw.IsPlaneswalker() && pw.Controller == dt.Controller && g.Battlefield.Contains(ev.Target)
	})

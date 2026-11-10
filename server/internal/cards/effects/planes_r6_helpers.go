package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// planes_r6_helpers.go holds the shapes the redo-6 plane cards share
// (ADR 0216 slice 4, #4870). Append-only.

// planesR6APlayerCastsInstantOrSorcery is "Whenever a player casts an
// instant or sorcery spell": any player's cast, the plane's controller's
// included.
func planesR6APlayerCastsInstantOrSorcery(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast || ev.Actor == uuid.Nil {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && (c.IsInstant() || c.IsSorcery())
}

// planesR6WhenYouPlaneswalkToThisAndAtYourFirstMain is "When you planeswalk
// to ~ and at the beginning of your first main phase, <effect>": ONE
// triggered ability with two conditions (see
// WhenYouPlaneswalkToThisAndAtYourUpkeep), the second the planar
// controller's precombat main phase (CR 505.1, 901.6).
func planesR6WhenYouPlaneswalkToThisAndAtYourFirstMain(label string, effect Effect) game.TriggeredAbility {
	return OnAny([]game.EventKind{game.EventPlaneswalkTo, game.EventBeginPrecombatMain}, planesR6PlaneswalkedToThisOrYourFirstMain, label, effect)
}

func planesR6PlaneswalkedToThisOrYourFirstMain(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	switch ev.Kind {
	case game.EventPlaneswalkTo:
		return ev.CardID == source.InstanceID && ev.Actor == source.Controller
	case game.EventBeginPrecombatMain:
		return ByYou(ev, source, lki, g)
	}
	return false
}

// planesR6APlayerCastsASpell is "Whenever a player casts a spell": any
// player's cast, the plane's controller's included.
func planesR6APlayerCastsASpell(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventCast && ev.Actor != uuid.Nil
}

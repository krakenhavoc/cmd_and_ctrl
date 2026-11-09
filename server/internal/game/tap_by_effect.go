package game

import "github.com/google/uuid"

// TapTargetForEffectBy is TapTargetForEffect for a tap that a player's
// spell or ability performs ("tap target creature"): the EventTapCard it
// emits names that player as its Actor, which is what "whenever you tap
// an untapped creature an opponent controls" reads (Solitary Sanctuary,
// Icewrought Sentry; CR 701.26a, #4452). A permanent that is already
// tapped is not tapped again and nothing is emitted, so an untapped
// permanent is exactly the one that announces.
//
// The Actor rides the event only; every other EventTapCard reader looks at
// CardID, so stamping it changes nothing for them. Caller must hold g.mu.
func (g *Game) TapTargetForEffectBy(cardID, actor uuid.UUID) error {
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.InstanceID != cardID {
			continue
		}
		if c.Tapped {
			return nil
		}
		c.Tapped = true
		g.EmitEvent(Event{Kind: EventTapCard, Actor: actor, CardID: cardID})
		return nil
	}
	return ErrCardNotFound
}

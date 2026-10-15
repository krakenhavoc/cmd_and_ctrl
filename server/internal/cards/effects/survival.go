package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Survival — the ability word on "Survival — At the beginning of your
// second main phase, if this creature is tapped, <effect>." (Defiant
// Survivor).
//
// The "if" is an intervening if (CR 603.4): asked as the phase begins, so
// an untapped creature never triggers, and asked again as the ability
// resolves. The second look reads the source as it is now, or as it last
// existed on the battlefield once it has left (CR 608.2h), so a creature
// untapped in response does nothing and one sacrificed while tapped still
// pays off.
func Survival(label string, effect Effect) game.TriggeredAbility {
	t := On(game.EventStepBegan, AllOf(YourSecondMainPhaseBegan, survivalSourceTapped), label,
		func(g *game.Game, item *game.StackItem) error {
			info, ok := NewContext(g, item).SourcePermanent()
			if !ok || !info.Tapped {
				return nil
			}
			return effect(g, item)
		})
	return t
}

// survivalSourceTapped is the trigger-time half of the intervening if.
func survivalSourceTapped(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return source.Tapped
}

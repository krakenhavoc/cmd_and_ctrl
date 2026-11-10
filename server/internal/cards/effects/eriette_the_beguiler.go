package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eriette, the Beguiler — Legendary Creature — Human Warlock {1}{W}{U}{B},
// 3/3 (Outlaws of Thunder Junction, #4468):
//
//	"Lifelink
//	 Whenever an Aura you control becomes attached to a nonland permanent an
//	 opponent controls with mana value less than or equal to that Aura's
//	 mana value, gain control of that permanent for as long as that Aura is
//	 attached to it."
//
// The attach event names the Aura (CardID) and the host (Target). The
// condition is judged as the Aura attaches (both mana values as they stand
// on the battlefield: a token is 0, an {X} spell counts X as 0, CR 202.3e);
// a re-attachment to the host it was already on is not "becomes attached"
// (CR 701.3b). The control effect lasts exactly as long as that Aura stays
// on that host (the engine's while-attached duration, pinned to both
// objects, CR 400.7), so it ends the moment the Aura moves or falls off,
// and an Aura already gone as the trigger resolves takes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "710a3499-64f8-4d2c-8c9c-d58950e47ea7",
		Name:            "Eriette, the Beguiler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttach, eriettePredicate,
				"Eriette, the Beguiler — gain control of that permanent for as long as that Aura is attached to it",
				func(g *game.Game, item *game.StackItem) error {
					ev := NewContext(g, item).Trigger().Event
					duration, ok := g.WhileAttachedToDuration(ev.CardID, ev.Target)
					if !ok {
						return nil
					}
					return GainControl{
						Target:   ev.Target,
						Duration: duration,
						Label:    "Eriette, the Beguiler — gain control while the Aura is attached",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

// eriettePredicate is the trigger's condition: one of the source controller's
// Auras became attached to a nonland permanent an opponent controls whose
// mana value is at most the Aura's.
func eriettePredicate(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventAttach || ev.Label != "Aura" || ev.AlreadyAttached || source == nil {
		return false
	}
	aura, ok := g.LookupCardForEffect(ev.CardID)
	if !ok || aura.Controller != source.Controller {
		return false
	}
	host, ok := g.LookupCardForEffect(ev.Target)
	if !ok || host.IsLand() || host.Controller == source.Controller || host.Controller == uuid.Nil {
		return false
	}
	return host.ManaValue() <= aura.ManaValue()
}

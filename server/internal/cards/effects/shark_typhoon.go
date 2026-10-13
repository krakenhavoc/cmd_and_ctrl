package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shark Typhoon — Enchantment {5}{U}:
//
//	"Whenever you cast a noncreature spell, create an X/X blue Shark
//	 creature token with flying, where X is that spell's mana value.
//	 Cycling {X}{1}{U} ({X}{1}{U}, Discard this card: Draw a card.)
//	 When you cycle this card, create an X/X blue Shark creature token
//	 with flying."
//
// The cast trigger reads the spell's mana value as the trigger resolves
// (CR 202.3e: an {X} counts only while the spell is on the stack), and
// a spell that has already left the stack counts as zero.
//
// The cycling half is the printed pair: the cycling ability draws the
// card, and a separate "when you cycle this card" trigger, fired from
// the graveyard (CR 702.29c), makes the Shark. EventCycle carries the X
// the cycling cost was announced with, so the trigger reads it off its
// event (ADR 0062 Decision 7, #3172) and countering the cycling ability
// leaves the Shark standing, as printed.
func init() {
	cycling := Cycling("{1}{U}")
	cycling.Label = "Cycling {X}{1}{U} ({X}{1}{U}, Discard this card: Draw a card.)"
	cycling.Cost = Plus(ManaCost("{X}{1}{U}"), DiscardThis())
	Register(Spec{
		OracleID:     "8c0520fa-276b-4d21-b4a9-dce1fce59f6b",
		Name:         "Shark Typhoon",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{cycling},
		Triggered: []game.TriggeredAbility{
			InGraveyard(On(game.EventCycle, Self, "Shark Typhoon — cycled: create an X/X blue Shark with flying, X the cycling cost's X",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return rfReprintBSharkToken(NewContext(g, item), item.Controller, item.Trigger.Event.Amount)
				})),
			WheneverYouCast(Noncreature(), "Shark Typhoon — create an X/X blue Shark with flying, X the spell's mana value",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return rfReprintBSharkToken(NewContext(g, item), item.Controller, spellManaValueForEffect(g, item.Trigger.Event.CardID))
				}),
		},
	})
}

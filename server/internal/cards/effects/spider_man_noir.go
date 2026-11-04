package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider-Man Noir — Legendary Creature — Spider Human Hero {4}{B}, 4/4
// (Marvel's Spider-Man, #4460):
//
//	"Menace
//	 Whenever a creature you control attacks alone, put a +1/+1 counter on it.
//	 Then surveil X, where X is the number of counters on it. (Look at the top X
//	 cards of your library, then put any number of them into your graveyard and
//	 the rest on top of your library in any order.)"
//
// Black Widow, Double Agent's lone-attacker trigger (the attacker is the event's
// card). X counts every kind of counter on the creature after the +1/+1 counter
// landed. If the attacker has left by the time the trigger resolves, no counter
// is put on it and X is the number of counters it last had (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2efd6494-c5e1-41de-bc6e-8490ff046cd1",
		Name:            "Spider-Man Noir",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttack, mshB1YouAttackAlone,
				"Spider-Man Noir — put a +1/+1 counter on the lone attacker, then surveil X",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					attacker := item.Trigger.Event.CardID
					var counters map[string]int
					if onBattlefield(g, attacker) {
						if err := (AddCounter{Target: attacker, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
						if c, ok := g.LookupCardForEffect(attacker); ok {
							counters = c.Counters
						}
					} else if info, ok := g.LastKnownPermanentForEffect(attacker); ok {
						counters = info.Counters
					}
					x := 0
					for _, n := range counters {
						if n > 0 {
							x += n
						}
					}
					if x == 0 {
						return nil
					}
					return Surveil{Player: item.Controller, N: x}.Apply(ctx)
				}),
		},
	})
}

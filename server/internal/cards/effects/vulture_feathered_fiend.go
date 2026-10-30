package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vulture, Feathered Fiend — Legendary Creature — Human Artificer Villain
// {2}{U}{U}, 2/4 (Marvel Super Heroes Commander, #4455):
//
//	"Flying
//	 Whenever one or more creatures you control with flying deal combat
//	 damage to a player, put a +1/+1 counter on each of those creatures
//	 and draw a card."
//
// One printed ability, two rows. The draw is Mu Yanling's batch trigger
// (one per player hit in a combat damage step, CR 603.2c). "Each of those
// creatures" must be the creatures that dealt the damage while they had
// flying, fixed when the ability triggers: a collapsed batch trigger keeps
// only its first event, so the counters are a row per damaging creature
// (the subject is the event's source, SubjectIsEventSource), whose
// condition reads flying as the damage is dealt. A creature that loses
// flying in response still gets its counter, one without flying that also
// connected gets none, and a creature that has left the battlefield by
// resolution gets nothing (the counter cannot be put on it). The two rows' effects are independent, so their order and number
// on the stack change nothing a player could observe except a response
// aimed at one row.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c598d97f-fec6-4238-8eed-8051e9541d31",
		Name:            "Vulture, Feathered Fiend",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(HasKeyword("flying"),
				"Vulture, Feathered Fiend — draw a card", Do(DrawCards{N: 1})),
			{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if !combatDamageToPlayerBy(ev, source.Controller, g) {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.Source)
					return ok && HasKeyword("flying")(g, source.Controller, game.ThroughTextOf(source, c))
				},
				Key:                  "Vulture, Feathered Fiend — put a +1/+1 counter on that flying creature",
				SubjectIsEventSource: true,
				Effect: func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || !onBattlefield(g, item.Trigger.Event.Source) {
						return nil
					}
					return AddCounter{Target: item.Trigger.Event.Source, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Janice Rand, Essential Aide — Legendary Creature — Human Advisor {3}{G},
// 3/3 (Star Trek Commander, #4974):
//
//	"Whenever a commander you control enters or attacks, put a +1/+1
//	 counter on it, then reveal cards from the top of your library until
//	 you reveal a land card. Put that card onto the battlefield tapped
//	 and the rest on the bottom of your library in a random order."
//
// One trigger, two conditions: a permanent that is a commander and is
// controlled by Janice's controller entering the battlefield, or being
// declared as an attacker. Janice herself counts when she is your
// commander. "It" is the commander the event named: the counter goes on
// that object only while it is still on the battlefield (a commander
// that left and came back is a new object, CR 400.7); the reveal happens
// either way. The reveal waits for the counter to land, so a CR 616
// ordering prompt over the placement is a restore point. The reveal is
// The Regalia's whole sentence (RevealUntilThenPutOntoBattlefield): the
// land enters tapped and everything else revealed goes to the bottom in
// a random order; a library with no land card is revealed whole and all
// of it goes to the bottom.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2130193c-3518-4a7d-9cb3-ea65aec845b4",
		Name:         "Janice Rand, Essential Aide",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OnAny([]game.EventKind{game.EventETB, game.EventAttack}, janiceRandCommanderYouControl,
				"Janice Rand, Essential Aide — put a +1/+1 counter on that commander, then reveal until a land card and put it onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					then := ctx.ItemTail(janiceRandReveal)
					obj := ctx.Trigger().Object
					if obj == nil {
						return then.Run(g, nil)
					}
					info, ok := g.PermanentForEffect(obj.Ref())
					if !ok || info.Left {
						return then.Run(g, nil)
					}
					return g.AddCounterByKeyedForEffect(uuid.Nil, obj.ID, game.CounterPlusOne, 1, then)
				}),
		},
	})
}

// janiceRandCommanderYouControl is "whenever a commander you control
// enters or attacks": the permanent the event names is a commander and
// the source's controller controls it.
func janiceRandCommanderYouControl(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventETB && ev.Kind != game.EventAttack {
		return false
	}
	c, ok := g.LookupCardForEffect(ev.CardID)
	return ok && c.IsCommander && c.Controller == source.Controller
}

// janiceRandReveal is the clause after the counter: reveal until a land
// card, put it onto the battlefield tapped, the rest on the bottom in a
// random order. An on-disk key: never renamed, never reused.
var janiceRandReveal = CardTail("janice-rand-essential-aide/reveal-until-land", func(ctx *Context, _ game.TailOutcome) error {
	return RevealUntilThenPutOntoBattlefield{
		Match:  game.Card.IsLand,
		Tapped: true,
		Reason: "Janice Rand, Essential Aide — revealed until a land card",
	}.Apply(ctx)
})

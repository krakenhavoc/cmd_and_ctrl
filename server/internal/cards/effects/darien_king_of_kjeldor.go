package effects

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Darien, King of Kjeldor — Legendary Creature — Human Soldier
// {4}{W}{W}, 3/3 (EDHREC rank 4477):
//
//	"Whenever you're dealt damage, you may create that many 1/1 white
//	 Soldier creature tokens."
//
// The commander that turns a burn spell into an army. Every point
// that gets through — an unblocked attacker, a Blasphemous Act
// pointed at you, your own Pyrohemia — is a body, and the deck is
// built to hurt itself on purpose: Star of Extinction, Repercussion,
// a Boros Reckoner in the red zone.
//
// Three things the card says that are easy to get wrong, and all
// three are implemented as printed:
//
//   - "You're dealt DAMAGE", not "you lose life". Paying life for a
//     Mana Confluence, a Vizkopa drain and a Necropotence activation
//     are all life loss, and Darien sees none of them (CR 119.3).
//   - "THAT MANY" is the damage that was actually APPLIED. A
//     prevention shield or a damage-halving replacement changes the
//     number before the event is emitted, so a fogged hit makes no
//     Soldiers — the damage never happened.
//   - "YOU MAY" is a real question, asked of Darien's controller as
//     the trigger resolves. It matters: a player at one life who has
//     just been hit for twenty does not always want twenty blockers,
//     and more to the point Darien may have died to that same damage,
//     in which case the trigger still resolves and the tokens still
//     arrive (CR 603.10 — the ability is independent of its source
//     once it is on the stack).
//
// One trigger per damage EVENT. Two creatures connecting in the same
// combat damage step are two events, so they are two triggers and two
// separate questions, each for its own amount — which is what "that
// many" means per instance.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d05336cb-6157-47e7-942f-43becd67a5bf",
		Name:         "Darien, King of Kjeldor",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := b43YouWereDealtDamage(ev, source, g)
				return ok
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				n, _ := b43YouWereDealtDamage(ev, source, g)
				return game.NewTriggeredItem(source, "Darien, King of Kjeldor — you may create "+strconv.Itoa(n)+" Soldiers",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						return MayChoice{
							Player:   item.Controller,
							Question: "Darien, King of Kjeldor — create " + strconv.Itoa(n) + " 1/1 white Soldier creature tokens?",
							YesLabel: "Create them",
							NoLabel:  "Decline",
							OnYes: func(ctx *Context) error {
								return CreateToken{
									Controller: item.Controller,
									Template:   TokenCard("1/1 white Soldier"),
									N:          n,
								}.Apply(ctx)
							},
						}.Apply(ctx)
					})
			},
		}},
	})
}

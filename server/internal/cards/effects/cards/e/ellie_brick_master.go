package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ellie, Brick Master — Legendary Creature — Human Survivor {1}{R}, 2/1
// (Secret Lair Drop):
//
//	"Distract the Horde — Whenever a player attacks one of your opponents,
//	 that attacking player creates a tapped 1/1 black Fungus Zombie creature
//	 token named Cordyceps Infected that's attacking that opponent.
//	 Partner—Survivors (You can have two commanders if both have this
//	 ability.)"
//
// Jolene, the Plunder Queen's attack trigger, which fires for the attacking
// player (Ellie's controller included) however many creatures attack, with the
// batch keyed on the attacker AND the opponent attacked, since the clause names
// "one of your opponents" and the token joins that opponent's attackers: a
// player who attacks two of Ellie's controller's opponents gets a token at
// each. An attack at a planeswalker or a battle is not an attack on the
// opponent (CR 508.3e, as Jolene's), and an attack at Ellie's controller is not
// one on "one of your opponents". The token is the attacker's, tapped and
// attacking that opponent, and was never declared as an attacker (CR 508.4).
// The token is Abby's Cordyceps Infected. Partner—Survivors is a
// deck-construction rule read off the oracle text.
//
// No simplification.
func init() {
	t := effects.OncePerBatch(effects.On(game.EventAttack, effects.AttackAimedAtAnOpponentOfYours,
		"Ellie, Brick Master — the attacking player creates a tapped Cordyceps Infected attacking that opponent",
		func(g *game.Game, item *game.StackItem) error {
			ev := item.Trigger.Event
			tmpl := effects.TokenCard("1/1 black Fungus Zombie named Cordyceps Infected")
			tmpl.Tapped = true
			tmpl.AttackingTarget = ev.Target
			return effects.CreateToken{Controller: ev.Actor, Template: tmpl, N: 1}.Apply(effects.NewContext(g, item))
		}))
	t.BatchKey = func(ev game.Event, _ *game.Card, g *game.Game) string {
		return ev.Actor.String() + "/" + effects.AttackedPlayer(g, ev).String()
	}
	effects.Register(effects.Spec{
		OracleID:     "299ff264-c6db-4b48-88fb-ca006cb84f88",
		Name:         "Ellie, Brick Master",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered:    []game.TriggeredAbility{t},
	})
}

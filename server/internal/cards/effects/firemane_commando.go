package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	b17FiremaneYouLabel   = "Firemane Commando — you attacked with two or more creatures: draw a card"
	b17FiremaneOtherLabel = "Firemane Commando — another player attacked with two or more creatures: they draw if none attacked you"
)

// Firemane Commando — Creature — Angel Soldier {3}{W}, 4/3 (EDHREC
// rank 1852):
//
//	"Flying
//	 Whenever you attack with two or more creatures, draw a card.
//	 Whenever another player attacks with two or more creatures, they
//	 draw a card if none of those creatures attacked you."
//
// The political Angel: you draw for going wide, and everyone else
// draws for going wide somewhere other than at you. Both triggers
// are b16PlayerAttackedWithAtLeast (Aurelia's shape): the engine
// emits EventAttack per creature, so the ability fires on the
// declaration that brings the attacking player to two, and the
// dedup — every later event of the same batch (OncePerBatch, see
// AGENTS.md §7), then the rest of the turn — declines every later
// one. The second ability's "if none of those creatures attacked
// you" is not an intervening if — it is checked
// as the trigger resolves, against every creature that player has
// attacking at that moment, with an attack at your planeswalker or
// battle counting as an attack at you (S27).
//
// Declared weaker than printed, as Aurelia: with an extra combat in
// the same turn the abilities would not fire again. The engine has
// no extra combats, so nothing reaches it today.
func init() {
	Register(Spec{
		OracleID:        "ac675899-25bb-4f06-9c8c-bf188024495f",
		Name:            "Firemane Commando",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && b16PlayerAttackedWithAtLeast(ev, source, g, 2, b17FiremaneYouLabel)
			}, b17FiremaneYouLabel, Do(DrawCards{N: 1}))),
			{
				OncePerBatch: true,
				Watches:      []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Actor != source.Controller && b16PlayerAttackedWithAtLeast(ev, source, g, 2, b17FiremaneOtherLabel)
				},
				Key: b17FiremaneOtherLabel,
				Effect: func(g *game.Game, item *game.StackItem) error {
					attacker := item.Trigger.Event.Actor
					if !b17AttackersAllAvoid(g, attacker, item.Controller) {
						return nil
					}
					return DrawCards{Player: attacker, N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

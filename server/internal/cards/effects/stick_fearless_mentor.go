package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stick, Fearless Mentor — Legendary Creature — Human Monk Hero {1}{R},
// 2/1 (Marvel Super Heroes Commander, #4455):
//
//	"First strike
//	 Whenever a source you control deals damage to you, exile the top card
//	 of your library. You may play it until the end of your next turn. This
//	 ability triggers only once each turn."
//
// The condition is Daredevil, Fearless Fighter's: any damage event aimed at
// the controller whose source (read as it was, CR 608.2h, or the combat
// actor) is controlled by the same player. The exile is Bloodsoaked
// Insight's impulse window (UntilEndOfYourNextTurnDuration) over your own
// library. "Only once each turn" is the per-turn trigger record
// (b11TriggeredThisTurn): a trigger already put on the stack this turn keeps
// every later damage event from triggering.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e7bd3882-a0f2-4706-80e7-1d3ec455801f",
		Name:            "Stick, Fearless Mentor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered: []game.TriggeredAbility{
			On(game.EventDealDamage, stickDamagedByYourSource, stickLabel, stickExile),
		},
	})
}

const stickLabel = "Stick, Fearless Mentor — exile the top card of your library; you may play it until the end of your next turn"

func stickDamagedByYourSource(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Amount <= 0 || ev.Target != source.Controller {
		return false
	}
	controller, ok := b34DamageSourceController(ev, g)
	return ok && controller == source.Controller && !b11TriggeredThisTurn(g, source.InstanceID, stickLabel)
}

func stickExile(g *game.Game, item *game.StackItem) error {
	window := g.UntilEndOfYourNextTurnDuration(item.Controller)
	return ExileTopWithPermission{
		From:    item.Controller,
		GrantTo: item.Controller,
		N:       1,
		Window:  &window,
	}.Apply(NewContext(g, item))
}

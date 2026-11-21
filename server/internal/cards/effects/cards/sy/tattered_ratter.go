package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tattered Ratter — Creature — Human Peasant {1}{R}, 2/2 (Wilds of
// Eldraine, #4471):
//
//	"Whenever a Rat you control becomes blocked, it gets +2/+0 until end
//	 of turn."
//
// She-Hulk, Wallbreaker's "a Hero you control becomes blocked" with a Rat:
// "Rat" is a word of the Ratter's text (CR 612.2), and the Rat is read
// with its layers applied, so a changeling counts. A creature becomes
// blocked once however many creatures block it (CR 509.1h). The pump is
// Flanking's on the event's creature; a Rat that has left the
// battlefield by the time the trigger resolves gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ba36f27a-08fc-46a5-bc55-2a71091cebcc",
		Name:         "Tattered Ratter",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesBlocked, tatteredRatterRatBecameBlocked,
				"Tattered Ratter — that Rat gets +2/+0 until end of turn",
				tatteredRatterPump),
		},
	})
}

// tatteredRatterRatBecameBlocked is "a Rat you control becomes blocked".
func tatteredRatterRatBecameBlocked(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	rat := source.TextSubtype("Rat")
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.InstanceID == ev.CardID {
			return c.Controller == source.Controller && c.IsCreature() && c.HasSubtype(rat)
		}
	}
	return false
}

// tatteredRatterPump gives the blocked Rat +2/+0 until end of turn.
func tatteredRatterPump(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
		return nil
	}
	return effects.BoostUntilEOT{Target: ctx.Trigger().Event.CardID, Power: 2,
		Label: "Tattered Ratter — +2/+0 until end of turn"}.Apply(ctx)
}

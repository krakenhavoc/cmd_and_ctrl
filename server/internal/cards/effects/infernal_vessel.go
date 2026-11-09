package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Infernal Vessel — Creature — Human Cleric {2}{B}, 2/1 (Foundations,
// #4465):
//
//	"When this creature dies, if it wasn't a Demon, return it to the
//	 battlefield under its owner's control with two +1/+1 counters on it.
//	 It's a Demon in addition to its other types."
//
// Thunderbolts Conspiracy's return-and-retype shape on the Vessel itself.
// "If it wasn't a Demon" is an intervening "if" (CR 603.4) read off the
// creature as it last existed on the battlefield (CR 603.10a), as the card's
// own text spells the type (CR 612.2), and rechecked as the trigger
// resolves. The return puts the two +1/+1 counters on as entry counters
// (CR 614.1c), and the Demon type is a layer-4 effect pinned to the new
// object (CR 611.2a, 400.7), so a second death finds it a Demon and
// returns nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0e13c283-f462-441f-a8a5-0c75a59ff1d9",
		Name:         "Infernal Vessel",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, AllOf(ThisDied, stdB15VesselWasNoDemon),
				"Infernal Vessel — return it with two +1/+1 counters, as a Demon too",
				func(g *game.Game, item *game.StackItem) error {
					src, ok := g.LookupCardForEffect(item.SourceCardID)
					if !ok || !stdB15VesselWasNoDemon(item.Trigger.Event, &src, game.Characteristic{}, g) {
						return nil
					}
					return stdB15ReturnDiedWithCounters(g, item, map[string]int{game.CounterPlusOne: 2}, stdB15VesselDemon)
				}),
		},
	})
}

// stdB15VesselWasNoDemon is the intervening "if it wasn't a Demon".
func stdB15VesselWasNoDemon(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	dead, ok := g.LookupCardForEffect(ev.CardID)
	if !ok {
		return false
	}
	return !leftAsSubtype(ev, dead, source.TextSubtype("Demon"))
}

// stdB15VesselDemon makes the returned creature a Demon in addition to its
// other types, for as long as it stays on the battlefield. An on-disk key.
var stdB15VesselDemon = EnteredTail("infernal-vessel/demon", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	g := ctx.Game
	if entered == uuid.Nil || !onBattlefield(g, entered) {
		return nil
	}
	return ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Demon")...)},
		Duration: g.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Infernal Vessel — it is a Demon in addition to its other types",
	}.Apply(ctx)
})

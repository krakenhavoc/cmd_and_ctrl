package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thunderbolts Conspiracy — Enchantment {3}{B} (Marvel Super Heroes,
// #4455):
//
//	"Flash
//	 Whenever a Villain you control dies, return it to the battlefield
//	 under its owner's control with a finality counter on it. That
//	 creature is a Hero in addition to its other types. (If a creature
//	 with a finality counter on it would die, exile it instead.)"
//
// "A Villain you control" is read as the creature last existed on the
// battlefield (CR 603.10a: leftAsSubtype, leftUnderControlOf), the type
// word as the card's own text spells it (CR 612.2). The return puts the
// card back with the finality counter as an entry counter (CR 614.1c),
// and the Hero type is a layer-4 effect pinned to the new object (CR 611.2a,
// 400.7), as Ascent of the Worthy's is. A token Villain has no card to
// return. The finality exile is the engine's (CR 122.1h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d90a5213-3363-4cd1-97d6-4fc057f78b2a",
		Name:            "Thunderbolts Conspiracy",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, mshB11VillainYouControlDied,
				"Thunderbolts Conspiracy — return it with a finality counter, as a Hero too", mshB11ThunderboltsReturn),
		},
	})
}

func mshB11VillainYouControlDied(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
	if !ACreatureYouControlDied(ev, source, lki, g) {
		return false
	}
	dead, _ := g.LookupCardForEffect(ev.CardID)
	return leftAsSubtype(ev, dead, source.TextSubtype("Villain"))
}

func mshB11ThunderboltsReturn(g *game.Game, item *game.StackItem) error {
	id := item.Trigger.Event.CardID
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	dead, ok := g.LookupCardForEffect(id)
	if !ok {
		return nil
	}
	ctx := NewContext(g, item)
	return g.ReturnFromGraveyardWithCountersKeyedForEffect(id, dead.Owner, false,
		map[string]int{game.CounterFinality: 1}, ctx.ItemTail(mshB11ThunderboltsHero))
}

// mshB11ThunderboltsHero makes the returned creature a Hero in addition to
// its other types, for as long as it stays on the battlefield. An on-disk
// key.
var mshB11ThunderboltsHero = EnteredTail("thunderbolts-conspiracy/hero", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	g := ctx.Game
	if entered == uuid.Nil || !onBattlefield(g, entered) {
		return nil
	}
	return ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddSubtypesMod(ctx.TextSubtypes("Hero")...)},
		Duration: g.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Thunderbolts Conspiracy — it is a Hero in addition to its other types",
	}.Apply(ctx)
})

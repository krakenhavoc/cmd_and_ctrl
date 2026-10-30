package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hellcat, Undying Vigilante — {G}{G} Legendary Creature — Human Hero 2/2
// (Marvel Super Heroes, #4455):
//
//	"Haste
//	 When Hellcat dies, return her to the battlefield under her owner's
//	 control with a +1/+1 counter on her. She loses all abilities and gains
//	 haste."
//
// The +1/+1 counter is an "enters with" clause (CR 614.1c): it rides the
// entry, so she is a 3/3 the moment she arrives. What she loses is every
// ability, her dies trigger included, so she returns once; the haste is
// granted after the loss in the same effect (CR 613.1f, later in the list),
// so the loss does not take it. Both are indefinite and pinned to this
// object (CR 400.7), the way Ascent of the Worthy's subtypes are. If she is
// no longer in the graveyard (exiled in response), nothing returns.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4c27ba61-6812-4dd7-8efc-65435a8df7b5",
		Name:            "Hellcat, Undying Vigilante",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Hellcat — return her with a +1/+1 counter, no abilities and haste", hellcatReturns),
		},
	})
}

func hellcatReturns(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	if id == uuid.Nil {
		return nil
	}
	if zone := g.FindCardZoneForEffect(id); zone == nil || zone.Kind != game.ZoneGraveyard {
		return nil
	}
	return g.ReturnFromGraveyardWithCountersKeyedForEffect(id, uuid.Nil, false,
		map[string]int{game.CounterPlusOne: 1}, NewContext(g, item).ItemTail(hellcatBareTail))
}

// hellcatBareTail is the entry's tail as a keyed tail: she loses all
// abilities and gains haste. An on-disk key.
var hellcatBareTail = EnteredTail("hellcat/loses-abilities-gains-haste", func(ctx *Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil || !onBattlefield(ctx.Game, entered) {
		return nil
	}
	return ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.LoseAllAbilitiesMod(), game.AddKeywordsMod("haste")},
		Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Hellcat — loses all abilities and gains haste",
	}.Apply(ctx)
})

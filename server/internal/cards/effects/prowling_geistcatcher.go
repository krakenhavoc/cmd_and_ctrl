package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Prowling Geistcatcher — Creature — Human Rogue {3}{B}, 2/4 (Midnight
// Hunt Commander, #5497):
//
//	"Whenever you sacrifice another creature, exile it. If that creature
//	 was a token, put a +1/+1 counter on this creature.
//	 When this creature leaves the battlefield, return each card exiled
//	 with it to the battlefield under your control."
//
// The first trigger is Eloise's reading of a sacrifice (announced while
// the creature is still on the battlefield, so "creature" and "token"
// are read off it there, CR 603.10), and Pirate Peddlers' "another".
// Resolution exiles the sacrificed card from the graveyard, linked to
// the Geistcatcher object (CR 607.2a, ExileTarget.Linked); a token has
// ceased to exist by then, so there is nothing to exile and the counter
// comes from the last-known fact that it was a token, not from what is
// left. A card that was already moved elsewhere in response is not
// exiled. The leave trigger is Angel of Serenity's: every card still in
// exile linked to the object that left comes back together, under the
// Geistcatcher controller's control rather than its owner's (the
// controller as the ability resolves, CR 603.10a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "175db9f1-0188-4640-9b54-6b461c5440aa",
		Name:         "Prowling Geistcatcher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventSacrifice, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Actor != source.Controller || ev.CardID == uuid.Nil || ev.CardID == source.InstanceID {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature()
			}, "Prowling Geistcatcher — exile it; if it was a token, put a +1/+1 counter on it", prowlingGeistcatcherExile),
			On(game.EventLTB, Self,
				"Prowling Geistcatcher — return each card exiled with it to the battlefield under your control",
				prowlingGeistcatcherReturn),
		},
	})
}

func prowlingGeistcatcherExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	trig := ctx.Trigger()
	if trig.Event.CardID == uuid.Nil {
		return nil
	}
	wasToken := false
	if trig.Object != nil {
		wasToken = trig.Object.Token
	}
	if wasToken {
		return CounterOnThis{Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
	}
	return ExileTarget{Target: trig.Event.CardID, Linked: true}.Apply(ctx)
}

func prowlingGeistcatcherReturn(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		ids = append(ids, c.InstanceID)
	}
	return ReturnFromExileTogether{Targets: ids, Controller: ctx.Controller()}.Apply(ctx)
}

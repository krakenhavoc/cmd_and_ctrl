package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// monarch.go is the card half of the monarch (CR 724, #1722, ADR
// 0096). The mechanic itself — the designation, the end-step draw and
// the combat-damage steal — is engine (game/monarch.go). What a card
// needs from it is four things, and they are all here:
//
//	BecomeTheMonarch{}                      // "you become the monarch"
//	WhenThisEntersYouBecomeTheMonarch(name) // the ETB every monarch card prints
//	WheneverYouBecomeTheMonarch(label, fx)  // Custodi Lich
//	YoureTheMonarch(g, you)                 // the Courts' "if you're the monarch"
//
// plus AnOpponentBecameTheMonarch for Knights of the Black Rose and the
// "until an opponent becomes the monarch" delayed condition Palace
// Jailer's exile is keyed to (delayed_bodies.go).
//
// # Why a primitive and not a direct call
//
// Before #1722 the only way to move the crown was the public
// Game.SetMonarch, which takes g.mu itself — so from inside a resolving
// OnResolve or a trigger's Effect, which already hold it, it
// deadlocked. BecomeTheMonarch goes through SetMonarchForEffect, the
// locked-context twin, and so emits the same EventMonarchChanged a
// manual set does: "whenever you become the monarch" and every "as long
// as you're the monarch" static behave identically whichever route
// moved the crown.
//
// # "Become" is a change
//
// A player who is already the monarch and is told to become it again
// does not become it (CR 724.3: one monarch; "as a player becomes the
// monarch" is a change of holder). The engine emits nothing, so a
// second Court entering while you wear the crown does not trigger your
// Custodi Lich. Nothing here has to check for it.

// BecomeTheMonarch is "<player> becomes the monarch". Player defaults
// to the resolving item's controller — "you become the monarch", which
// is what every card in the catalog prints.
//
// A player who has left the game is not crowned (CR 800.4a): the
// effect does as much as it can, which is nothing, and the resolution
// goes on.
type BecomeTheMonarch struct {
	Player uuid.UUID
}

func (b BecomeTheMonarch) Apply(ctx *Context) error {
	player := b.Player
	if player == uuid.Nil {
		player = ctx.Controller()
	}
	if p := ctx.Game.PlayerByIDForEffect(player); p == nil || p.Eliminated {
		return nil
	}
	return ctx.Game.SetMonarchForEffect(player)
}

// WhenThisEntersYouBecomeTheMonarch is the line every monarch
// permanent prints: "When this <permanent> enters, you become the
// monarch." `name` is the card's name, for the stack label.
//
// A trigger, on the stack, like any ETB (CR 603.6a): the table gets to
// respond before the crown moves, and the ETB triggers of the same
// permanent are ordered by its controller (Palace Jailer's two).
func WhenThisEntersYouBecomeTheMonarch(name string) game.TriggeredAbility {
	return WhenThisEnters(name+" — you become the monarch", Do(BecomeTheMonarch{}))
}

// YouBecameTheMonarch — "whenever you become the monarch": the crown
// moved TO the source's controller. The engine emits the event only on
// a real change, so this is once per becoming.
func YouBecameTheMonarch(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventMonarchChanged && ev.Actor != uuid.Nil && ev.Actor == source.Controller
}

// AnOpponentBecameTheMonarch — "whenever an opponent becomes the
// monarch": the crown moved to a player other than the source's
// controller. A cleared designation (Actor uuid.Nil) is nobody
// becoming anything.
func AnOpponentBecameTheMonarch(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventMonarchChanged && ev.Actor != uuid.Nil && ev.Actor != source.Controller
}

// WheneverYouBecomeTheMonarch — "Whenever you become the monarch,
// <effect>" (Custodi Lich).
func WheneverYouBecomeTheMonarch(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventMonarchChanged, YouBecameTheMonarch, label, effect)
}

// YoureTheMonarch is the condition reader: "if you're the monarch" /
// "as long as you're the monarch" / "while you're the monarch".
//
// Read it where the card reads it. The Courts' upkeep payoff is "If
// you're the monarch, <bigger> instead" — a clause of the effect,
// checked as the trigger RESOLVES (CR 608.2), not an intervening "if"
// on the trigger: a Court whose controller lost the crown in response
// still fires and pays the small half. Entourage of Trest's extra
// block is a static, read by the layer pass, which EventMonarchChanged
// invalidates. Regal Behemoth's is the condition of a triggered mana
// ability, read as the land is tapped.
func YoureTheMonarch(g *game.Game, you uuid.UUID) bool {
	return you != uuid.Nil && g.Monarch == you
}

// AnOpponentIsTheMonarch is "if an opponent is the monarch" (Queen
// Marchesa): somebody holds the crown and it is not you.
func AnOpponentIsTheMonarch(g *game.Game, you uuid.UUID) bool {
	return g.Monarch != uuid.Nil && g.Monarch != you
}

// anOpponentBecameTheMonarchSince is the delayed half of Palace
// Jailer's "until an opponent becomes the monarch": an
// EventMonarchChanged whose new monarch is somebody other than the
// delayed trigger's controller (CR 603.7d — the controller of the
// ability that created it, which is who "an opponent" is relative to).
func anOpponentBecameTheMonarchSince(ev game.Event, dt *game.DelayedTrigger, _ *game.Game, _ game.EffectParams) bool {
	return ev.Kind == game.EventMonarchChanged && ev.Actor != uuid.Nil && ev.Actor != dt.Controller
}

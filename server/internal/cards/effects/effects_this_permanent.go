package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// effects_this_permanent.go — whole-effect bodies for an activated
// ability that acts on its own source: "Destroy this enchantment",
// "Remove a doom counter from this artifact". Written for the
// any-player cards of ADR 0106 PR 6 (#1793), where the player who
// activates the ability is often not the one who controls the
// permanent it acts on, but nothing here reads who activated it.
//
// Append-only, per the shared-vocabulary rule: add a body, never
// change what an existing one does.

// destroyThisPermanent is "Destroy this <permanent>." as an ability's
// whole effect (Volrath's Dungeon), with "It can't be regenerated"
// when cantBeRegenerated is set (Aether Storm, CR 701.19c).
//
// The destruction is part of the EFFECT, not a cost, so it happens on
// resolution and the ability can be answered: a source already gone
// is not destroyed again, and one that left and came back is a new
// object the ability knows nothing about (CR 400.7) — DestroyTarget
// asks that question itself (#1432).
func destroyThisPermanent(cantBeRegenerated bool) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !onBattlefield(g, item.SourceCardID) {
			return nil
		}
		return DestroyTarget{Target: item.SourceCardID, CantBeRegenerated: cantBeRegenerated}.
			Apply(NewContext(g, item))
	}
}

// removeACounterFromThis is "Remove a <kind> counter from this
// <permanent>." as an ability's whole EFFECT (Armageddon Clock,
// Infinite Hourglass) — the effect, not RemoveCountersFromThis's
// cost, so it can be activated with no counter there and then does
// nothing (CR 609.3: it does only as much as possible).
func removeACounterFromThis(kind string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		c, ok := g.LookupCardForEffect(item.SourceCardID)
		if !ok || !onBattlefield(g, item.SourceCardID) || c.Counters[kind] <= 0 {
			return nil
		}
		return AddCounter{Target: item.SourceCardID, Kind: kind, N: -1}.Apply(NewContext(g, item))
	}
}

// putACounterOnThis is "put a <kind> counter on this <permanent>" as a
// trigger's whole effect (Armageddon Clock's doom counter, Infinite
// Hourglass's time counter). A source that has left has nothing to put
// it on, and one that came back is a new object (#1432, AddCounter).
func putACounterOnThis(kind string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !onBattlefield(g, item.SourceCardID) {
			return nil
		}
		return AddCounter{Target: item.SourceCardID, Kind: kind, N: 1}.Apply(NewContext(g, item))
	}
}

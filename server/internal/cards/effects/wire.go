package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// wire.go populates the function-variable hooks declared in
// server/internal/game/effect_hooks.go at package init time.
// Importing this package (blank import from main.go is enough)
// switches the game's resolution path from manual-sandbox-only to
// catalog-aware.
//
// The hooks are set unconditionally — absence of a catalog entry
// still matters (Lookup returns false) and is what preserves the
// opt-in invariant for non-catalog cards.
//
// Added in S14 sub-PR 3.

func init() {
	game.EffectResolver = resolveSpell
	game.ETBEffectHook = fireOnETB
	game.IsCatalogCard = Has
}

// resolveSpell is the EffectResolver implementation. Called from
// resolveTopOfStackLocked after target re-check, before zone
// routing. Looks up the spell's Scryfall ID; runs OnResolve if
// the spec has one. A miss (non-catalog card) or a Spec with nil
// OnResolve (vanilla permanent in the catalog — Birds of Paradise,
// Sol Ring) is a no-op. The resolution path carries on with
// standard zone routing either way.
func resolveSpell(g *game.Game, item *game.StackItem, oracleID string) error {
	spec, ok := Lookup(oracleID)
	if !ok {
		return nil
	}
	if spec.OnResolve == nil {
		return nil
	}
	ctx := NewContext(g, item)
	return spec.OnResolve(item, ctx)
}

// fireOnETB is the ETBEffectHook implementation. Looks up the
// card's Scryfall ID, stamps StartingLoyalty (planeswalkers), and
// runs OnETB if the spec has one. Invoked from every battlefield-
// entry site in the game package.
//
// Looks up the live Card via FindCardZoneForEffect so the hook
// reads the current card state (post-ETB zone move) rather than a
// stale copy captured before the move.
func fireOnETB(g *game.Game, cardID uuid.UUID, oracleID string) error {
	spec, ok := Lookup(oracleID)
	if !ok {
		return nil
	}
	// Stamp starting loyalty first — the SBA check that kills
	// 0-loyalty planeswalkers runs on the next priority-grant
	// boundary, so the counters have to be in place before then.
	if spec.StartingLoyalty > 0 {
		if err := g.AddCounterForEffect(cardID, "loyalty", spec.StartingLoyalty); err != nil {
			return err
		}
	}
	if spec.OnETB == nil {
		return nil
	}
	// Find the live card in the battlefield so the hook receives
	// the current state (post-move).
	zone := g.FindCardZoneForEffect(cardID)
	if zone == nil {
		return nil
	}
	var card *game.Card
	for i := range zone.Cards {
		if zone.Cards[i].InstanceID == cardID {
			card = &zone.Cards[i]
			break
		}
	}
	if card == nil {
		return nil
	}
	ctx := NewContext(g, nil)
	return spec.OnETB(card, ctx)
}

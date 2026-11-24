package game

import "github.com/google/uuid"

// ForetellPermanentForEffect is "exile it face down. It becomes foretold":
// The Foretold Soldier's damage trigger (CR 702.143b). The permanent goes to its
// owner's exile face down with the foretell kind, which is what a foretold card
// IS, and its owner gets the foretell permission: cast it from exile for `cost`
// on a LATER turn (CR 702.143a), for as long as it remains exiled.
//
// It is not the foretell special action (CR 702.143a), so it costs nothing and
// does not count toward "the first card you foretell each turn"
// (Game.ForetoldThisTurn). The exile is one routed move, so replacement effects
// and a commander's CR 903.9 offer apply; a permanent that does not land in exile
// (replaced, or gone already) is not foretold and gets no permission.
//
// Caller must hold g.mu in write mode.
func (g *Game) ForetellPermanentForEffect(cardID uuid.UUID, cost string) error {
	c, ok := g.battlefieldCardLocked(cardID)
	if !ok {
		return nil
	}
	owner := c.Owner
	return g.routeAllKeyedLocked(zoneRoute{
		Dst:      ZoneExile,
		Actor:    owner,
		FaceDown: FaceDownForetold,
		Cause:    MoveCause{Kind: MoveCauseEffect, Controller: owner},
	}, []uuid.UUID{cardID}, KeyedTail{Ref: foretoldByEffect, Player: owner, N: g.Turn.Seq + 1, Word: cost})
}

// foretoldByEffect is the continuation of ForetellPermanentForEffect, as data:
// Player is the owner, N the first turn the card may be cast on and Word the
// foretell cost. It stamps the same per-instance cast permission foretoldExiled
// does, without the foretell tally. An on-disk key: never renamed, never reused.
var foretoldByEffect = RegisterTail("engine/foretold-by-effect", func(g *Game, o TailOutcome) error {
	if len(o.IDs) != 1 {
		return nil
	}
	exiled := exiledCardByIDLocked(g, o.IDs[0])
	if exiled == nil || o.Tail.Word == "" {
		// A card with no mana cost is foretold (face down) but has no
		// foretell cost to cast it for.
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(CastPermission{
		Player:       o.Tail.Player,
		Zone:         ZoneExile,
		AltCostKey:   AltCostKeyForetell,
		Cost:         o.Tail.Word,
		NotBeforeSeq: o.Tail.N,
		Duration:     WhileInZoneDuration(),
		CastOnly:     true,
		Label:        "Foretell",
	}, []Card{*exiled})
	return nil
})

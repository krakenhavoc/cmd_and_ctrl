package game

import "github.com/google/uuid"

// exile_play.go — S21 sub-PR 6: impulse exile. "Exile the top card
// of your library. Until end of turn, you may cast that card."
//
// Cards in exile are normally inert — the zone is where things go to
// stop mattering. Impulse exile inverts that for one card and one
// player: Ragavan exiles off the top of the player he hit and lets
// YOU cast it, this turn only, from a zone every player can see.
//
// So the permission has to live on the card rather than in a
// player-scoped map: it names a player who is not the card's owner,
// it expires, and it must survive Clone / RestoreFrom (undo) intact.
// A value struct on Card does all three for free — cloneCard's
// `out := c` copies it, and there's no pointer to alias.
//
// This is not "cast from exile" in general (Foretell, adventure
// cards, Prowl-style alternative zones all differ). It is exactly
// CR 118.7's "a player may play a card from a zone they normally
// couldn't", with a duration.

// ExilePlayPermission is a standing permission to play one exiled
// card. The zero value grants nothing.
type ExilePlayPermission struct {
	// Player is who may play the card — usually NOT its owner.
	Player uuid.UUID

	// UntilTurn is the last turn number on which the permission is
	// live. "Until end of turn" grants the turn it was created in.
	// A turn-number bound rather than a cleanup sweep because extra
	// turns, and because a card re-exiled by something else later
	// must not inherit a stale grant.
	UntilTurn int

	// CastOnly restricts the grant to casting. Ragavan says "you may
	// CAST that card"; a land exiled by Ragavan is stranded, since
	// playing a land is not casting (CR 305.1). Breeches says "you
	// may PLAY those cards", so its grant leaves this false.
	CastOnly bool

	// AnyColor lets the holder spend mana as though it were mana of
	// any color for this card (Breeches, Brazen Plunderer). Only
	// observable under the strict mana gate.
	AnyColor bool
}

// Active reports whether playerID may play the card on `turn`.
func (p ExilePlayPermission) Active(playerID uuid.UUID, turn int) bool {
	return p.Player != uuid.Nil && p.Player == playerID && turn <= p.UntilTurn
}

// Granted reports whether the permission names anyone at all.
func (p ExilePlayPermission) Granted() bool {
	return p.Player != uuid.Nil
}

// clearExpiredExilePlayLocked drops every permission whose window
// has closed. Called from the cleanup-step hook, alongside the
// damage wipe and the turn-scoped replacement clear — the same
// "this turn is over" sweep. Caller must hold g.mu.
func (g *Game) clearExpiredExilePlayLocked() {
	if g.Exile == nil {
		return
	}
	for i := range g.Exile.Cards {
		p := g.Exile.Cards[i].ExilePlay
		if p.Granted() && p.UntilTurn <= g.Turn.Number {
			g.Exile.Cards[i].ExilePlay = ExilePlayPermission{}
		}
	}
}

// ExileTopWithPermissionForEffect exiles the top n cards of
// `fromPlayer`'s library and grants `grantTo` permission to play
// them until the end of the current turn. Returns the exiled card
// IDs in the order they left the library.
//
// The exiled cards are revealed to everyone: they're face up in a
// public zone, and a permission nobody can see is unplayable in
// practice.
//
// Caller must hold g.mu.
func (g *Game) ExileTopWithPermissionForEffect(fromPlayer, grantTo uuid.UUID, n int, perm ExilePlayPermission) ([]uuid.UUID, error) {
	owner := g.playerByIDLocked(fromPlayer)
	if owner == nil {
		return nil, ErrPlayerNotFound
	}
	perm.Player = grantTo
	if perm.UntilTurn == 0 {
		perm.UntilTurn = g.Turn.Number
	}
	var out []uuid.UUID
	for i := 0; i < n; i++ {
		if owner.Library.Size() == 0 {
			return out, nil
		}
		top := owner.Library.Cards[0].InstanceID
		if _, err := MoveCard(owner.Library, g.Exile, top); err != nil {
			return out, err
		}
		for j := range g.Exile.Cards {
			if g.Exile.Cards[j].InstanceID == top {
				g.Exile.Cards[j].ExilePlay = perm
				break
			}
		}
		g.markCardKnownInZoneLocked(g.Exile, top)
		g.EmitEvent(Event{
			Kind:    EventZoneMove,
			Actor:   grantTo,
			CardID:  top,
			OldZone: ZoneLibrary,
			NewZone: ZoneExile,
		})
		out = append(out, top)
	}
	return out, nil
}

// asAnyColorCost rewrites a cost so every colored requirement is
// payable by any mana — "you may spend mana as though it were mana
// of any color" (Breeches). Folding the colored slots into the
// generic demand is exactly equivalent for the pool solver: a
// requirement any token can satisfy IS a generic requirement.
//
// Colorless {C} requirements are left alone. "Mana of any color"
// does not include colorless (CR 106.1b), so a {C} slot still needs
// real colorless mana.
func asAnyColorCost(cost ParsedCost) ParsedCost {
	out := cost
	out.Required = nil
	for _, req := range cost.Required {
		if requiresColorless(req) {
			out.Required = append(out.Required, req)
			continue
		}
		out.Generic++
	}
	return out
}

// requiresColorless reports whether a requirement can only be paid
// with colorless mana — i.e. every option it admits is {C}.
func requiresColorless(req ColorRequirement) bool {
	if len(req.Options) == 0 {
		return false
	}
	for _, c := range req.Options {
		if c != "C" {
			return false
		}
	}
	return true
}

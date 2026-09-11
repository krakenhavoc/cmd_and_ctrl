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
	// must not inherit a stale grant. Ignored when WhileExiled is
	// set.
	UntilTurn int

	// WhileExiled makes the grant unbounded — airbend's "WHILE IT'S
	// EXILED, its owner may cast it for {2}". There is no turn to
	// count to, so UntilTurn is not consulted and the cleanup sweep
	// skips the card entirely.
	//
	// A named boolean rather than a sentinel UntilTurn (-1, MaxInt):
	// the failure mode of a sentinel is that a caller who simply
	// forgets to set UntilTurn gets an unbounded grant by accident,
	// since 0 is both the zero value and a plausible turn number.
	// The boolean can only be switched on deliberately.
	//
	// "While it's exiled" is self-limiting rather than eternal: the
	// cast path zeroes ExilePlay as the card leaves exile (see the
	// two clear sites in mutations.go), so a card airbent, cast, and
	// later exiled again by something else does not inherit the
	// permission it was granted the first time. That is the same
	// staleness guard the turn-bounded grants rely on, and it is the
	// only thing standing between an unbounded grant and a
	// permanently-castable card.
	WhileExiled bool

	// CostOverride is a mana cost in Scryfall brace notation that
	// the holder pays INSTEAD of the card's printed mana cost —
	// airbend's "{2} rather than its mana cost" (CR 118.9). Empty
	// means the ordinary printed cost, which is what impulse exile
	// grants.
	//
	// This is the grant-carried sibling of AlternativeCost. It is
	// NOT an AlternativeCost, because the two are keyed differently
	// and deliberately so: an AlternativeCost is a property of the
	// CARD, looked up by oracle ID and offered to anyone casting it,
	// while this one is a property of a single exiled INSTANCE, and
	// two copies of the same card in exile can easily carry
	// different grants (one airbent, one impulse-exiled).
	//
	// SIMPLIFICATION: the override is charged rather than offered.
	// Printed airbend says the owner "MAY cast it for {2} RATHER
	// THAN its mana cost" — both prices are legal, and a {W}
	// creature is cheaper at its printed cost. The engine charges
	// {2} unconditionally, so the cheaper option is unavailable on a
	// card whose printed cost is below {2}. Strictly weaker than
	// printed (an offer of two prices, narrowed to one of them),
	// never stronger — and the alternative, threading a second
	// claimable key through the announce path, would have to
	// restructure validateAlternativeCost, which is shared cost
	// machinery.
	CostOverride string

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

// Active reports whether playerID may play the card on `turn`. An
// unbounded grant (WhileExiled) ignores `turn` — its duration is the
// card's continued presence in exile, which this type cannot see and
// does not need to: the cast path clears the permission as the card
// leaves.
func (p ExilePlayPermission) Active(playerID uuid.UUID, turn int) bool {
	if p.Player == uuid.Nil || p.Player != playerID {
		return false
	}
	return p.WhileExiled || turn <= p.UntilTurn
}

// Granted reports whether the permission names anyone at all.
func (p ExilePlayPermission) Granted() bool {
	return p.Player != uuid.Nil
}

// clearExpiredExilePlayLocked drops every permission whose window
// has closed. Called from the cleanup-step hook, alongside the
// damage wipe and the turn-scoped replacement clear — the same
// "this turn is over" sweep. Caller must hold g.mu.
//
// Unbounded grants (WhileExiled) are skipped: airbend's window
// closes when the card leaves exile, not when a turn ends, and the
// sweep has no business reaping one. Reaping it here would be the
// silent kind of wrong — the card stays visible in exile, the
// client keeps offering the cast button from a stale snapshot, and
// the cast comes back ErrNoPlayPermission with nothing on screen
// explaining why.
func (g *Game) clearExpiredExilePlayLocked() {
	if g.Exile == nil {
		return
	}
	for i := range g.Exile.Cards {
		p := g.Exile.Cards[i].ExilePlay
		if p.WhileExiled {
			continue
		}
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

// ExileCardWithPermissionForEffect exiles one specific card —
// typically a targeted battlefield permanent — and stamps `perm`
// onto it in its new home. The airbend half of the exile-play
// primitives, as against ExileTopWithPermissionForEffect's
// impulse-exile half: that one picks its cards off the top of a
// library and grants to someone who is usually NOT the owner; this
// one is handed a card someone chose and grants to the owner.
//
// A zero perm.Player means "the card's owner", which is what every
// airbend card says and what makes this primitive different from
// the impulse one in the way that matters. The owner is read off
// the card AFTER the move, not before: Card.Owner is fixed at deck
// build and rides the card through every zone, so the two are the
// same value, and reading it after means one scan rather than two.
//
// The move goes through ExileCardForEffect so the LKI snapshot,
// the ZoneMove event and the LTB event are exactly the ones every
// other exile produces — an airbent creature's leaves-the-
// battlefield triggers fire normally, and dies-triggers correctly
// do not (CR 700.4: exile is not dying).
//
// A card that is no longer in exile once the move completes is not
// an error. The LTB event dispatches triggers synchronously, and
// one of them may move the card on (a commander heading for the
// command zone is the live example); the grant simply does not land
// on a card that isn't there to receive it.
//
// Caller must hold g.mu.
func (g *Game) ExileCardWithPermissionForEffect(cardID uuid.UUID, perm ExilePlayPermission) error {
	if err := g.ExileCardForEffect(cardID); err != nil {
		return err
	}
	if g.Exile == nil {
		return nil
	}
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID != cardID {
			continue
		}
		if perm.Player == uuid.Nil {
			perm.Player = g.Exile.Cards[i].Owner
		}
		if !perm.WhileExiled && perm.UntilTurn == 0 {
			perm.UntilTurn = g.Turn.Number
		}
		// The card is already publicly known — ExileCardForEffect
		// marks every seat as a knower on the way in, which is what
		// the grant needs: a permission nobody can see is unplayable
		// in practice.
		g.Exile.Cards[i].ExilePlay = perm
		return nil
	}
	return nil
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

package game

import (
	"github.com/google/uuid"
)

// foretell.go — foretell (CR 702.143), #658.
//
// Foretell is three things, and the point of this file is how few of
// them are new:
//
//  1. A SPECIAL ACTION from hand (CR 702.143a, CR 116.2h): pay {2},
//     exile the card face down. That is the CR 116.2 verb — one
//     entry, one timing table — in special_action.go.
//  2. A FACE-DOWN OBJECT IN EXILE its owner may look at and nobody
//     else may (CR 702.143d). That is ADR 0069's model:
//     `FaceDownKind = foretold` is a row in one viewers table, and
//     the route, the wire, the redaction and the client already draw
//     it.
//  3. A PER-INSTANCE CAST PERMISSION over that one exiled card, priced
//     at the foretell cost and live from a later turn (CR 702.143a).
//     That is ADR 0066's model, with the same `NotBeforeTurn` floor
//     warp already uses for "on a later turn".
//
// So foretell adds no model. What it adds is the wiring, and one
// question the three models do not answer between them: what makes a
// spell "foretold" on the stack.
//
// # "If this spell was foretold" (CR 702.143c)
//
// Poison the Cup, Haunting Voyage and Starnheim Unleashed read it,
// and it is true of a spell cast from a foretold card HOWEVER it was
// cast — paying the foretell cost is the usual way but not the
// definition.
//
// It needs no field on Card, because the face-down state IS the
// foretold state: a foretold card is precisely a card in exile face
// down with kind `foretold`, and CR 406.3a turns it face up as it is
// cast — which is `MoveCard`'s unconditional ClearFaceDown, ADR 0069
// decision 5. So the question is asked of the card in the source
// zone, before it moves, and the answer is carried on the stack item
// (`StackItem.Foretold`). One bit on the item rather than a second
// per-instance flag to keep in step with the first.

// CardIsForetold reports whether this card object is a foretold card
// (CR 702.143b): in exile, face down, with the foretell kind.
//
// It reads the kind rather than a flag of its own because they would
// be the same fact twice — and the one that is already snapshotted,
// cloned, redacted and drawn by the client is the kind.
func CardIsForetold(c Card) bool {
	return c.FaceDown && c.FaceDownKind == FaceDownForetold
}

// ForetellExileCost is the fixed cost of the foretell special action
// (CR 702.143a). It is the same {2} on every card that prints
// foretell — the card's own foretell COST is what varies, and that is
// SpecialAction.CastCost.
const ForetellExileCost = "{2}"

// AltCostKeyForetell is the alternative-cost key the granted cast is
// claimed under. Shared with the printed-keyword keys on purpose
// (ADR 0066 Decision 3): StackItem.AltCost is what a card reads back,
// and it must say the same thing however the permission arrived.
const AltCostKeyForetell = "foretell"

// foretellLocked is the foretell special action's performer, run by
// PerformSpecialAction once the {2} is paid (CR 702.143a-b).
//
// The exile goes through the shared exit primitive with the face-down
// kind on the route, so three things come free: the CR 614
// replacement window runs over it, a commander gets CR 903.9's offer,
// and the landing writes the decision-2 viewers (the owner, and only
// the owner) instead of marking the whole table a knower.
//
// The grant is stamped from the CONTINUATION rather than on the next
// line, and #870 is why: a leg that merely paused leaves the card in
// hand, and stamping a permission over a card that has not moved
// would grant a cast out of a zone it is not in. `landed` is CR
// 400.7's reading — a commander that took the command zone left, but
// not to exile, so it was never foretold and gets no grant.
//
// Caller must hold g.mu (write).
func (g *Game) foretellLocked(p *Player, cardID uuid.UUID, sa SpecialAction) error {
	owner := p.ID
	notBefore := g.Turn.Number + 1
	return g.routeAllThenLocked(zoneRoute{
		Dst:      ZoneExile,
		Actor:    owner,
		FaceDown: FaceDownForetold,
		// #1320: a special action (CR 116.2h), not a spell or ability.
		Cause: MoveCause{Kind: MoveCauseSpecialAction, Controller: owner},
	}, []uuid.UUID{cardID}, func(g *Game, landed []uuid.UUID) error {
		if len(landed) != 1 {
			return nil
		}
		exiled := exiledCardByIDLocked(g, landed[0])
		if exiled == nil {
			return nil
		}
		// CR 702.143a: "Cast it on a LATER turn for its foretell
		// cost." NotBeforeTurn is a FLOOR and composes with the
		// Duration rather than replacing it (#945), which is exactly
		// the pair warp already needs — the window opens next turn
		// and stays open for as long as the card remains exiled, with
		// CR 400.7 enforced by the object epoch rather than by a
		// sweep.
		//
		// CastOnly, because CR 702.143a says "cast it": a land with
		// foretell would not become a land drop out of exile. No
		// printed card is one; the flag is the rule rather than a
		// guard against a card.
		g.GrantCastPermissionToCardsForEffect(CastPermission{
			Player:        owner,
			Zone:          ZoneExile,
			AltCostKey:    AltCostKeyForetell,
			Cost:          sa.CastCost,
			NotBeforeTurn: notBefore,
			Duration:      WhileInZoneDuration(),
			CastOnly:      true,
			Label:         "Foretell",
		}, []Card{*exiled})
		return nil
	})
}

// exiledCardByIDLocked returns a pointer to the live copy of a card in
// the shared exile zone, or nil.
//
// Caller must hold g.mu.
func exiledCardByIDLocked(g *Game, id uuid.UUID) *Card {
	if g.Exile == nil {
		return nil
	}
	for i := range g.Exile.Cards {
		if g.Exile.Cards[i].InstanceID == id {
			return &g.Exile.Cards[i]
		}
	}
	return nil
}

// revealForetoldAtGameEndLocked is the game-end half of CR 702.143f:
// when the game ends, all face-down foretold cards are revealed.
//
// ADR 0069 decision 5 built the owner-leaves half
// (revealFaceDownOwnedByLocked, called from leaveGameObjectsLocked)
// and left this one to #658, because there was no game-end sweep to
// hang it on and nothing was foretold.
//
// It reveals every face-down card in exile, not only the foretold
// ones, and that is the rule read literally rather than a widening:
// CR 406.3's "no player may examine it" is a rule about a game in
// progress, and the sibling hook already reveals a Necropotence exile
// when its owner leaves. A game that has ended has no hidden
// information left to protect, and a replay that showed a blank card
// back forever would be the worse answer.
//
// It runs BEFORE State becomes StateEnded, because RevealForEffect
// emits an event and the listeners it runs are entitled to a game
// that is still active.
//
// Caller must hold g.mu (write).
func (g *Game) revealForetoldAtGameEndLocked() {
	if g.Exile == nil {
		return
	}
	byOwner := map[uuid.UUID][]uuid.UUID{}
	var order []uuid.UUID
	for i := range g.Exile.Cards {
		c := &g.Exile.Cards[i]
		if !c.FaceDown || c.Owner == uuid.Nil {
			continue
		}
		if _, seen := byOwner[c.Owner]; !seen {
			order = append(order, c.Owner)
		}
		byOwner[c.Owner] = append(byOwner[c.Owner], c.InstanceID)
	}
	// One reveal per owner, in a stable order: "its owner reveals it"
	// names a player, and a deterministic order keeps the event log
	// reproducible under an undo and a snapshot restore.
	for _, owner := range order {
		g.RevealForEffect(RevealSpec{
			Player: owner,
			Reason: "foretold cards revealed as the game ends",
			Cards:  byOwner[owner],
		})
	}
}

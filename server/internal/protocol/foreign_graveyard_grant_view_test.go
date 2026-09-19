package protocol

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// foreign_graveyard_grant_view_test.go — #1022. `stampLegalTargets`
// walked each seat's graveyard and asked `grantedCast` for the seat
// that OWNS the pile, so a CastPermission whose Player is somebody else
// reached the wire as nothing at all: no offers, no legal targets, no
// modes, no `cant_cast`, no `castable_here`, for anybody. ADR 0066
// makes a ScopeCards permission a statement about an OBJECT — Wrexial's
// "you may cast target instant or sorcery card from that player's
// graveyard" is the printed shape — and the wire never said so.
//
// The fix is exile's shape (#978) on the one per-seat zone that can
// carry a foreign holder's stamps: compute for the holder, mark with
// `castOffersFor`, strip per viewer. `castable_here` travels with the
// stamps, because a public bit with a per-viewer answer is what #1015
// took off this surface.

// knownToEveryone marks one card in a zone known to every seat, so an
// assertion about the per-viewer STRIP is not silently satisfied by
// the per-viewer redaction instead.
func knownToEveryone(g *game.Game, z *game.Zone, id uuid.UUID) {
	known := map[uuid.UUID]bool{}
	for _, p := range g.Seats {
		known[p.ID] = true
	}
	for i := range z.Cards {
		if z.Cards[i].InstanceID == id {
			z.Cards[i].KnownBy = known
		}
	}
}

// grantOverCard gives `holder` a ScopeCards permission over one card
// wherever it sits, the way Snapcaster and Wrexial both grant.
func grantOverCard(t *testing.T, g *game.Game, holder, cardID uuid.UUID, perm game.CastPermission) {
	t.Helper()
	perm.Player = holder
	g.WithWriteLock(func() {
		if !g.GrantCastPermissionOverCardForEffect(cardID, perm) {
			t.Fatalf("GrantCastPermissionOverCardForEffect(%s): card not found", cardID)
		}
	})
}

// seatViewOf picks one seat's projection out of a filtered view.
func seatViewOf(t *testing.T, v GameView, seat uuid.UUID) *PlayerView {
	t.Helper()
	for i := range v.Seats {
		if v.Seats[i].ID == seat.String() {
			return &v.Seats[i]
		}
	}
	t.Fatalf("seat %s is missing from the view", seat)
	return nil
}

// foreignFlashbackGrant seeds a card in `owner`'s graveyard that
// `holder` has a flashback-shaped permission over, and returns its ID.
func foreignFlashbackGrant(t *testing.T, g *game.Game, owner, holder *game.Player) uuid.UUID {
	t.Helper()
	const oracle = "test-view-foreign-yard"
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{oracle: nil})
	owner.Graveyard.Cards = nil
	id := graveyardCard(owner, "Their Dead Spell", oracle)
	// An INSTANT, so the enumerator's CR 307.1 speed gate is not what
	// the coherence test below is measuring: the holder is not the
	// active seat, and a sorcery in somebody else's graveyard would
	// produce two empty lists that agree about nothing.
	for i := range owner.Graveyard.Cards {
		if owner.Graveyard.Cards[i].InstanceID == id {
			owner.Graveyard.Cards[i].TypeLine = "Instant"
		}
	}
	knownToEveryone(g, owner.Graveyard, id)
	grantOverCard(t, g, holder.ID, id, game.CastPermission{
		Zone: game.ZoneGraveyard, Scope: game.ScopeCards, AltCostKey: "flashback",
	})
	return id
}

// The holder gets the whole announce surface, computed for THEM.
func TestAForeignGraveyardGrantReachesItsHolder(t *testing.T) {
	g := buildActiveGame(t)
	owner, holder := g.Seats[0], g.Seats[1]
	id := foreignFlashbackGrant(t, g, owner, holder)

	card := cardInSeatZone(t, ViewOfGameFor(g, holder.ID.String()).Seats[0].Graveyard, id)
	if !card.CastableHere {
		t.Errorf("castable_here is clear for the permission's holder")
	}
	if got := keysOf(card.AlternativeCosts); !sameStrings(got, []string{"flashback"}) {
		t.Errorf("holder offers = %v, want [flashback]", got)
	}
	if !card.AlternativeCostRequired {
		t.Errorf("alternative_cost_required is clear, but the grant prices the cast")
	}
}

// The zone's OWNER holds no permission over their own card here, and
// the honest answer for them is the one that shipped before #1022:
// nothing. The card itself is still in their public graveyard.
func TestTheZoneOwnerWithoutAPermissionSeesNoOffers(t *testing.T) {
	g := buildActiveGame(t)
	owner, holder := g.Seats[0], g.Seats[1]
	id := foreignFlashbackGrant(t, g, owner, holder)

	card := cardInSeatZone(t, ViewOfGameFor(g, owner.ID.String()).Seats[0].Graveyard, id)
	if card.CastableHere {
		t.Errorf("castable_here is set for the zone's owner, who may not cast it")
	}
	if len(card.AlternativeCosts) != 0 {
		t.Errorf("owner offers = %v, want none", keysOf(card.AlternativeCosts))
	}
	if card.AlternativeCostRequired {
		t.Errorf("the owner carries alternative_cost_required for an offer list they do not get")
	}
	if card.Name == "" {
		t.Errorf("the card itself vanished from a public zone")
	}
}

// And a third seat — and a spectator, who is the same case with an
// empty viewer ID — gets the card and none of one seat's answers.
func TestABystanderSeesNoForeignGraveyardOffers(t *testing.T) {
	g := busyTable(t, 0)
	const oracle = "test-view-foreign-yard-bystander"
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{oracle: nil})
	owner, holder, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	owner.Graveyard.Cards = nil
	id := graveyardCard(owner, "Their Dead Spell", oracle)
	knownToEveryone(g, owner.Graveyard, id)
	grantOverCard(t, g, holder.ID, id, game.CastPermission{
		Zone: game.ZoneGraveyard, Scope: game.ScopeCards, AltCostKey: "flashback",
	})

	theirs := cardInSeatZone(t, ViewOfGameFor(g, bystander.ID.String()).Seats[0].Graveyard, id)
	if theirs.CastableHere {
		t.Errorf("a bystander is told the card is a cast surface")
	}
	if len(theirs.AlternativeCosts) != 0 {
		t.Errorf("bystander offers = %v, want none", keysOf(theirs.AlternativeCosts))
	}
	if theirs.Name == "" {
		t.Errorf("the bystander lost the card itself, not just the stamps")
	}

	spectator := cardInSeatZone(t, FilterViewFor(ViewOfGame(g), "").Seats[0].Graveyard, id)
	if spectator.CastableHere || len(spectator.AlternativeCosts) != 0 {
		t.Errorf("a spectator got one seat's cast surface: castable=%v offers=%v",
			spectator.CastableHere, keysOf(spectator.AlternativeCosts))
	}
}

// The coherence property #1012 / #1015 built, asked of the new path:
// what the view stamps for the holder is what the enumerator offers
// them, out of the same zone, at the same prices.
func TestViewAndEnumeratorAgreeOnAForeignGraveyardCast(t *testing.T) {
	// busyTable rather than buildActiveGame, and the ACTIVE seat as the
	// holder: the enumerator answers only mulligan moves while that
	// window is open, and only the seat holding priority gets a move
	// list at all.
	g := busyTable(t, 0)
	holder := g.Seats[g.Turn.ActiveSeat]
	owner := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	id := foreignFlashbackGrant(t, g, owner, holder)

	moves := enumeratedPrices(t, g, holder.ID)
	got := moves[id.String()+"@graveyard"]
	if got.keys == nil {
		got.keys = map[string]bool{}
	}
	// Non-vacuity: two empty lists agree about nothing.
	if !got.keys["flashback"] {
		t.Fatalf("the enumerator offers %v for the foreign graveyard card, want [flashback]", got.sorted())
	}

	yard := seatViewOf(t, ViewOfGameFor(g, holder.ID.String()), owner.ID).Graveyard
	card := cardInSeatZone(t, yard, id)
	view := viewPrices(card, "graveyard")
	if view.printed != got.printed {
		t.Errorf("view says printed-cost-claimable=%v, the enumerator says %v", view.printed, got.printed)
	}
	if !sameStrings(view.sorted(), got.sorted()) {
		t.Errorf("view offers %v, the enumerator offers %v", view.sorted(), got.sorted())
	}
}

// A STANDING permission is a permanent's printed text, and every one
// of them says "your graveyard". Underworld Breach must not give escape
// to the cards in an opponent's graveyard — the rule that went
// unwritten while every caller scoped the question to its own pile.
func TestAStandingGraveyardGrantStopsAtItsHoldersOwnYard(t *testing.T) {
	g := buildActiveGame(t)
	const oracle = "test-view-standing-breach"
	withStandingPermission(t, oracle, game.CastPermission{
		Zone:       game.ZoneGraveyard,
		Scope:      game.ScopeStanding,
		Duration:   game.WhileInZoneDuration(),
		Filter:     game.PermissionFilter{NonLandOnly: true},
		AltCostKey: "escape",
		Label:      "Escape — its mana cost",
	})
	me, them := g.Seats[0], g.Seats[1]
	breach := game.NewCard("Test Breach", me.ID)
	breach.TypeLine = "Enchantment"
	breach.OracleID = oracle
	breach.Controller = me.ID
	breach.KnownBy = map[uuid.UUID]bool{me.ID: true, them.ID: true}
	g.Battlefield.PushTop(breach)

	me.Graveyard.Cards, them.Graveyard.Cards = nil, nil
	mine := graveyardCard(me, "My Dead Spell", "")
	theirs := graveyardCard(them, "Their Dead Spell", "")
	knownToEveryone(g, me.Graveyard, mine)
	knownToEveryone(g, them.Graveyard, theirs)

	v := ViewOfGameFor(g, me.ID.String())
	if own := cardInSeatZone(t, v.Seats[0].Graveyard, mine); !own.CastableHere {
		t.Errorf("the Breach controller's OWN graveyard card is not a cast surface")
	}
	if other := cardInSeatZone(t, v.Seats[1].Graveyard, theirs); other.CastableHere {
		t.Errorf("a standing \"your graveyard\" permission reached an opponent's graveyard")
	}
	// And the engine agrees, which is where it matters: the enumerator
	// must not offer the cast either.
	for _, m := range legal.EnumerateLocked(g, me.ID, legal.Options{}) {
		if m.Type == legal.TypeCastSpell && strings.Contains(string(m.Params), theirs.String()) {
			t.Errorf("the enumerator offered a cast of an opponent's graveyard card under a standing grant: %s", m.Params)
		}
	}
}

// --- #1037: a card two seats may both cast ---------------------------
//
// A CardView is one struct, and until #1037 the announce-time stamps
// were one seat's answer written into it: `foreignCastHolder` took the
// first seat in seat order for a graveyard and
// CastPermissionOnCardForEffect the first live permission for exile.
// The second holder got the public zone, the public `exile_play` and
// no stamps — no picker and no button, for a cast the engine would
// accept. The stamps are filed per holder now and FilterViewFor hands
// out exactly one; the wire shape is unchanged, because the frame was
// always built per viewer.

// A card its OWNER may cast out of their own graveyard (printed
// flashback) while an opponent holds a Wrexial-style grant over the
// same card. The owner's answer is the public one; the other holder's
// is theirs alone.
func TestTwoHoldersEachSeeTheirOwnGraveyardCastStamps(t *testing.T) {
	g := busyTable(t, 0)
	owner, holder, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	const oracle = "test-view-two-holders-yard"
	withCastableZones(t, map[string][]game.ZoneKind{oracle: {game.ZoneGraveyard}})
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{
		oracle: {{
			Key: "flashback", Label: "Flashback {2}{R}", ManaCost: "{2}{R}",
			FromZone: game.ZoneGraveyard,
		}},
	})
	owner.Graveyard.Cards = nil
	id := graveyardCard(owner, "Twice-Claimed Spell", oracle)
	knownToEveryone(g, owner.Graveyard, id)
	// The SECOND permission, held by another seat, with its own key so
	// the two answers are told apart by what they offer.
	grantOverCard(t, g, holder.ID, id, game.CastPermission{
		Zone: game.ZoneGraveyard, Scope: game.ScopeCards, AltCostKey: "escape",
	})

	// The owner: their own printed flashback, and no sight of the
	// other seat's escape.
	own := cardInSeatZone(t, ViewOfGameFor(g, owner.ID.String()).Seats[0].Graveyard, id)
	if !own.CastableHere {
		t.Errorf("the graveyard's owner lost their own printed cast surface")
	}
	if got := keysOf(own.AlternativeCosts); !sameStrings(got, []string{"flashback"}) {
		t.Errorf("owner offers = %v, want [flashback]", got)
	}

	// The second holder: their own grant's price, and the printed
	// flashback beside it — CastOffersForLocked answers for THEM out
	// of the same zone, and a card's printed offers are not the
	// owner's private property.
	theirs := cardInSeatZone(t, ViewOfGameFor(g, holder.ID.String()).Seats[0].Graveyard, id)
	if !theirs.CastableHere {
		t.Errorf("the second holder is not told the card is a cast surface")
	}
	if got := keysOf(theirs.AlternativeCosts); !sameStrings(got, []string{"escape", "flashback"}) {
		t.Errorf("second holder offers = %v, want [escape flashback]", got)
	}

	// And a bystander gets the public answer — the owner's — and no
	// trace of the private one.
	other := cardInSeatZone(t, ViewOfGameFor(g, bystander.ID.String()).Seats[0].Graveyard, id)
	if got := keysOf(other.AlternativeCosts); !sameStrings(got, []string{"flashback"}) {
		t.Errorf("bystander offers = %v, want the public [flashback]", got)
	}
}

// Two impulse grants over one exiled card, from two different effects.
// Exile has no owner, so NOTHING about the announce surface is public:
// each holder gets their own, and each gets their own `exile_play`,
// which is the field their client's impulse button reads.
func TestTwoHoldersEachSeeTheirOwnExileCastStamps(t *testing.T) {
	g := busyTable(t, 0)
	first, second, bystander := g.Seats[0], g.Seats[1], g.Seats[2]
	const oracle = "test-view-two-holders-exile"
	withAltCostsByOracle(t, map[string][]game.AlternativeCost{oracle: nil})

	spell := game.NewCard("Twice-Granted Spell", first.ID)
	spell.TypeLine = "Instant"
	spell.ManaCost = "{1}{R}"
	spell.OracleID = oracle
	id := exileWithGrant(t, g, spell, game.CastPermission{
		Player: first.ID, Zone: game.ZoneExile, Scope: game.ScopeCards,
	})
	grantOverCard(t, g, second.ID, id, game.CastPermission{
		Zone: game.ZoneExile, Scope: game.ScopeCards, AltCostKey: "escape",
	})

	one := cardInSeatZone(t, ViewOfGameFor(g, first.ID.String()).Exile, id)
	if one.ExilePlay == nil || one.ExilePlay.Player != first.ID.String() {
		t.Errorf("first holder's exile_play = %+v, want their own grant", one.ExilePlay)
	}
	if got := keysOf(one.AlternativeCosts); len(got) != 0 {
		t.Errorf("first holder offers = %v, want none — their grant prices nothing", got)
	}

	two := cardInSeatZone(t, ViewOfGameFor(g, second.ID.String()).Exile, id)
	if two.ExilePlay == nil || two.ExilePlay.Player != second.ID.String() {
		t.Errorf("second holder's exile_play = %+v, want their OWN grant, not the first one found",
			two.ExilePlay)
	}
	if got := keysOf(two.AlternativeCosts); !sameStrings(got, []string{"escape"}) {
		t.Errorf("second holder offers = %v, want [escape]", got)
	}

	// The bystander keeps the public grant — it is public information,
	// the trigger that made it resolved in the open — and gets no
	// seat's announce surface.
	none := cardInSeatZone(t, ViewOfGameFor(g, bystander.ID.String()).Exile, id)
	if none.ExilePlay == nil {
		t.Errorf("a bystander lost the public exile_play")
	}
	if len(none.AlternativeCosts) != 0 || none.LegalTargets != nil {
		t.Errorf("a bystander got a holder's announce surface: offers=%v targets=%+v",
			keysOf(none.AlternativeCosts), none.LegalTargets)
	}
}

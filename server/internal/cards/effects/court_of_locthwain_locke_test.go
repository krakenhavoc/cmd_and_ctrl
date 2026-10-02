package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #1729: the one-spell cast permission (CastPermission.CastsLeft) and
// the two cards it unblocks.

const (
	courtOfLocthwainOracle = "6106ba65-f07b-41ec-8b59-e052bbde5529"
	lockeOracle            = "a800bec4-bacc-43f9-a773-dd795959935b"
)

// seedLibraryTop puts a named card on top of p's library.
func topOfLibrary(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = typeLine
	c.ManaCost = manaCost
	p.Library.PushTop(c)
	return c.InstanceID
}

// courtUpkeep runs one of seat 1's upkeeps with Court of Locthwain on
// the battlefield, targeting seat 2.
func courtUpkeep(t *testing.T, g *game.Game) {
	t.Helper()
	owner, target := g.Seats[1], g.Seats[2]
	advanceToUpkeepOf(t, g, 1)
	monSettle(t, g)
	pickPlayer(t, g, owner.ID, target.ID)
	monSettle(t, g)
}

func courtFreeGrants(p *game.Player) int {
	n := 0
	for _, perm := range p.CastPermissions {
		if perm.AltCostKey == courtOfLocthwainFreeCastKey {
			n++
		}
	}
	return n
}

// Not the monarch: the card is exiled and playable for as long as it
// stays exiled, with mana of any type, and nothing is free.
func TestCourtOfLocthwainExilesAPlayableCard(t *testing.T) {
	g := newCatalogGame(t)
	owner, target := g.Seats[1], g.Seats[2]
	pushPermanentForTest(g, owner.ID, "Court of Locthwain", courtOfLocthwainOracle, "Enchantment")
	monCrown(t, g, g.Seats[3].ID)
	top := topOfLibrary(target, "Their Land", "Land", "")
	courtUpkeep(t, g)
	if z := e2Zone(g, top); z != game.ZoneExile {
		t.Fatalf("their top card is in %q, want exile", z)
	}
	perm := e2Permission(g, top)
	if perm == nil || perm.Player != owner.ID || perm.CastOnly || !perm.AnyType {
		t.Fatalf("permission %+v, want a play (not cast-only) any-type grant to the Court's controller", perm)
	}
	if perm.Duration.Kind != game.WhileInZone {
		t.Errorf("duration %+v, want for as long as it remains exiled", perm.Duration)
	}
	if courtFreeGrants(owner) != 0 {
		t.Error("a free cast was granted to a controller who is not the monarch")
	}
}

// The monarch casts ONE spell from among every card exiled with the
// Court so far, free, and the rest stay playable for their cost.
func TestCourtOfLocthwainMonarchCastsOneSpellFree(t *testing.T) {
	g := newCatalogGame(t)
	owner, target := g.Seats[1], g.Seats[2]
	pushPermanentForTest(g, owner.ID, "Court of Locthwain", courtOfLocthwainOracle, "Enchantment")
	monCrown(t, g, g.Seats[3].ID)
	first := topOfLibrary(target, "Their Bolt", "Instant", "{2}{R}")
	courtUpkeep(t, g)
	advanceTo(t, g, game.StepPrecombatMain)

	monCrown(t, g, owner.ID)
	// Seeded once the target's own draw step has gone by.
	advanceToUpkeepOf(t, g, 1)
	second := topOfLibrary(target, "Their Ponder", "Instant", "{3}{U}")
	courtUpkeep(t, g)
	if z := e2Zone(g, second); z != game.ZoneExile {
		t.Fatalf("the second card is in %q, want exile", z)
	}
	if courtFreeGrants(owner) != 1 {
		t.Fatalf("%d free grants, want one over the whole set", courtFreeGrants(owner))
	}

	// The earlier card, cast free with an empty pool.
	if err := g.CastSpell(owner.ID, first, game.CastSpellParams{
		FromZone: "exile", AlternativeCost: courtOfLocthwainFreeCastKey, Strict: true,
	}); err != nil {
		t.Fatalf("free cast of a card exiled with the Court on an earlier turn: %v", err)
	}
	if courtFreeGrants(owner) != 0 {
		t.Fatal("the free cast was not spent by the first spell")
	}
	// The second is not free any more…
	if err := g.CastSpell(owner.ID, second, game.CastSpellParams{
		FromZone: "exile", AlternativeCost: courtOfLocthwainFreeCastKey, Strict: true,
	}); err == nil {
		t.Fatal("a second spell was cast free")
	}
	// …but it is still playable under the first permission, for its cost.
	if perm := e2Permission(g, second); perm == nil || perm.Player != owner.ID {
		t.Fatalf("the play permission over the second card is gone: %+v", perm)
	}
}

// Paying for one card under the play permission leaves the free cast
// for another: the caster chooses which permission a cast uses.
func TestCourtOfLocthwainPaidCastKeepsTheFreeOne(t *testing.T) {
	g := newCatalogGame(t)
	owner, target := g.Seats[1], g.Seats[2]
	pushPermanentForTest(g, owner.ID, "Court of Locthwain", courtOfLocthwainOracle, "Enchantment")
	monCrown(t, g, owner.ID)
	card := topOfLibrary(target, "Their Bolt", "Instant", "{R}")
	courtUpkeep(t, g)
	if err := g.CastSpell(owner.ID, card, game.CastSpellParams{FromZone: "exile"}); err != nil {
		t.Fatalf("paid cast under the play permission: %v", err)
	}
	if courtFreeGrants(owner) != 1 {
		t.Error("a cast that did not claim the free cast spent it")
	}
}

// --- Locke, Treasure Hunter -------------------------------------------

// Mug: each player mills, a Treasure for a land among them, and ONE
// spell from among the milled cards — including one in an opponent's
// graveyard.
func TestLockeMugsOneSpellFromTheMilledCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third, fourth := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	locke := pushDiesCreatureForTest(g, me.ID, "Locke, Treasure Hunter", lockeOracle,
		"Legendary Creature — Human Rogue", 2, 3)
	topOfLibrary(me, "My Land", "Land", "")
	a := topOfLibrary(opp, "Their Shock", "Instant", "{R}")
	b := topOfLibrary(third, "Their Opt", "Instant", "{U}")
	topOfLibrary(fourth, "Their Bear", "Creature — Bear", "{1}{G}")
	treasures := monTokens(g, me.ID, "Treasure")

	declareAttack(t, g, opp.ID, locke)
	passPriorityAroundTable(t, g)
	if z := e2Zone(g, a); z != game.ZoneGraveyard {
		t.Fatalf("the opponent's top card is in %q, want their graveyard", z)
	}
	if got := monTokens(g, me.ID, "Treasure") - treasures; got != 1 {
		t.Errorf("%d Treasures, want 1 for the milled land", got)
	}
	if err := g.CastSpell(me.ID, a, game.CastSpellParams{FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast from an opponent's graveyard under Mug: %v", err)
	}
	if err := g.CastSpell(me.ID, b, game.CastSpellParams{FromZone: "graveyard"}); err == nil {
		t.Error("a second spell was cast from among the milled cards")
	}
}

// No land milled, no Treasure.
func TestLockeMugMakesNoTreasureWithoutALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	locke := pushDiesCreatureForTest(g, me.ID, "Locke, Treasure Hunter", lockeOracle,
		"Legendary Creature — Human Rogue", 2, 3)
	for _, p := range g.Seats {
		topOfLibrary(p, "Spell", "Instant", "{U}")
	}
	treasures := monTokens(g, me.ID, "Treasure")
	declareAttack(t, g, opp.ID, locke)
	passPriorityAroundTable(t, g)
	if got := monTokens(g, me.ID, "Treasure") - treasures; got != 0 {
		t.Errorf("%d Treasures with no land milled", got)
	}
}

// "Locke can't be blocked by creatures with greater power."
func TestLockeCantBeBlockedByGreaterPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	locke := pushDiesCreatureForTest(g, me.ID, "Locke, Treasure Hunter", lockeOracle,
		"Legendary Creature — Human Rogue", 2, 3)
	big := pushVanillaCreature(g, opp.ID, "Their Giant", 3, 3)
	same := pushVanillaCreature(g, opp.ID, "Their Peer", 2, 2)
	declareAttack(t, g, opp.ID, locke)
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(big, locke); err == nil {
		t.Error("a creature with greater power blocked Locke")
	}
	if err := g.DeclareBlocker(same, locke); err != nil {
		t.Errorf("a creature with equal power could not block Locke: %v", err)
	}
}

package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spellweaver_volute_test.go — #3157: Spellweaver Volute end to end, over
// card_copy_cast_test.go's and animate_dead_test.go's helpers.

const spellweaverVoluteOracle = "3d92e5ce-9d30-4203-9f84-6c36117440cf"

// svCastVolute casts Spellweaver Volute from the active seat's hand at
// the instant card `target` and lets it resolve, attached to that card.
func svCastVolute(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Spellweaver Volute", TypeLine: "Enchantment — Aura",
		OracleID: spellweaverVoluteOracle, ManaCost: "{3}{U}{U}", Colors: []string{"U"},
		Owner: me.ID, Controller: me.ID,
	})
	toMainForCost(t, g)
	fillPoolColored(me, "U", 5)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Strict:  true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: target}},
	}); err != nil {
		t.Fatalf("CastSpell Spellweaver Volute: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := adCard(t, g, id); adZone(g, id) != game.ZoneBattlefield || c.AttachedTo.ID != target {
		t.Fatalf("Spellweaver Volute is in %v attached to %v, want the battlefield attached to %v", adZone(g, id), c.AttachedTo.ID, target)
	}
	return id
}

// svSetup seats the Volute on a Lightning Bolt in the active player's
// graveyard.
func svSetup(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	volute := svCastVolute(t, g, bolt)
	return g, me, opp, bolt, volute
}

// svCastSorcery casts Divination, which triggers the Volute.
func svCastSorcery(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return castCatalogSpell(t, g, "Divination", "Sorcery", divinationOracle, nil)
}

func svCastCopy(g *game.Game, who, copyID uuid.UUID, targets ...game.TargetRef) error {
	return g.CastSpell(who, copyID, game.CastSpellParams{FromZone: "graveyard", Strict: true, Targets: targets})
}

func svHost(t *testing.T, g *game.Game, volute uuid.UUID) uuid.UUID {
	t.Helper()
	return adCard(t, g, volute).AttachedTo.ID
}

// The whole card: a sorcery triggers it, the copy is made in the
// graveyard and cast free from there, and casting it exiles the
// enchanted card and moves the Volute onto the only other instant card
// in a graveyard (an opponent's), with no question asked.
func TestSpellweaverVoluteCopiesCastsExilesAndMovesToTheOnlyOtherInstant(t *testing.T) {
	g, me, opp, bolt, volute := svSetup(t)
	shock := gyCard(opp, "Shock", "Instant", "{R}", shockOracleForCopies)
	gyCard(opp, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")

	svCastSorcery(t, g)
	passUntilCopies(t, g)

	cp := copyOf(t, g, "Lightning Bolt")
	if !me.Graveyard.Contains(cp.InstanceID) || cp.InstanceID == bolt {
		t.Fatal("the copy is a new object in the graveyard")
	}
	if svHost(t, g, volute) != bolt || !me.Graveyard.Contains(bolt) {
		t.Fatal("nothing moves until the copy is cast")
	}
	before := opp.Life
	if err := svCastCopy(g, me.ID, cp.InstanceID, playerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the copy without paying its mana cost: %v", err)
	}
	if !inExile(g, bolt) {
		t.Error("casting the copy exiles the enchanted card")
	}
	if svHost(t, g, volute) != shock || adZone(g, volute) != game.ZoneBattlefield {
		t.Fatalf("the Volute is in %v attached to %v, want the battlefield on the Shock", adZone(g, volute), svHost(t, g, volute))
	}
	if n := len(g.PendingChoices); n != 0 {
		t.Errorf("%d choices pending: one legal new host is a forced choice", n)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("life = %d, want %d: the copy resolves as Lightning Bolt", opp.Life, before-3)
	}
	if svHost(t, g, volute) != shock || adZone(g, volute) != game.ZoneBattlefield {
		t.Error("the Volute stays on its new instant card")
	}
}

// With several instant cards, the caster chooses among them; the card
// the copy was made from, a creature card and the copy are not offered.
func TestSpellweaverVoluteAsksWhichInstantCardToMoveTo(t *testing.T) {
	g, me, opp, bolt, volute := svSetup(t)
	shock := gyCard(opp, "Shock", "Instant", "{R}", shockOracleForCopies)
	mine := gyCard(me, "Shock", "Instant", "{R}", shockOracleForCopies)
	bears := gyCard(me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")

	svCastSorcery(t, g)
	passUntilCopies(t, g)
	cp := copyOf(t, g, "Lightning Bolt")
	if err := svCastCopy(g, me.ID, cp.InstanceID, playerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the copy: %v", err)
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoiceChooseCards, me.ID)
	if pick == nil {
		t.Fatal("no choice of a new instant card")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.ChooseCards {
		offered[id] = true
	}
	if len(offered) != 2 || !offered[shock] || !offered[mine] || offered[bolt] || offered[bears] || offered[cp.InstanceID] {
		t.Fatalf("offered %v, want exactly the two Shocks", pick.ChooseCards)
	}
	if adZone(g, volute) != game.ZoneBattlefield || svHost(t, g, volute) != bolt || !me.Graveyard.Contains(bolt) {
		t.Fatal("while the question is open the Volute stays on the Bolt, which is still in the graveyard")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{shock}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if !inExile(g, bolt) || svHost(t, g, volute) != shock || adZone(g, volute) != game.ZoneBattlefield {
		t.Fatalf("after the answer: bolt exiled %v, Volute in %v on %v", inExile(g, bolt), adZone(g, volute), svHost(t, g, volute))
	}
	passPriorityAroundTable(t, g)
	if svHost(t, g, volute) != shock || adZone(g, volute) != game.ZoneBattlefield {
		t.Error("the Volute stays on the chosen card")
	}
}

// The ruling: with no other instant card in a graveyard the enchanted
// card is still exiled, and the Volute, enchanting nothing, goes to its
// owner's graveyard.
func TestSpellweaverVoluteWithNoOtherInstantFallsOff(t *testing.T) {
	g, me, opp, bolt, volute := svSetup(t)
	gyCard(opp, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")

	svCastSorcery(t, g)
	passUntilCopies(t, g)
	cp := copyOf(t, g, "Lightning Bolt")
	if err := svCastCopy(g, me.ID, cp.InstanceID, playerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the copy: %v", err)
	}
	if !inExile(g, bolt) {
		t.Error("the enchanted card is exiled")
	}
	if !me.Graveyard.Contains(volute) {
		t.Errorf("the Volute is in %v, want its owner's graveyard", adZone(g, volute))
	}
}

// Declining: the copy ceases to exist when the caster passes, and the
// Volute stays on its card, which stays in the graveyard.
func TestSpellweaverVoluteDeclinedCopyLeavesEverythingWhereItWas(t *testing.T) {
	g, me, opp, bolt, volute := svSetup(t)
	gyCard(opp, "Shock", "Instant", "{R}", shockOracleForCopies)

	svCastSorcery(t, g)
	passUntilCopies(t, g)
	passPriorityAroundTable(t, g)
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left after the caster passed", n)
	}
	if !me.Graveyard.Contains(bolt) || svHost(t, g, volute) != bolt || adZone(g, volute) != game.ZoneBattlefield {
		t.Error("nothing is exiled or moved when the copy is not cast")
	}
}

// "Whenever you cast a SORCERY spell": an instant does not trigger it.
func TestSpellweaverVoluteDoesNotTriggerOnAnInstant(t *testing.T) {
	g, me, opp, bolt, volute := svSetup(t)

	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if n := len(cardCopies(g)); n != 0 {
		t.Fatalf("an instant made %d copies", n)
	}
	if !me.Graveyard.Contains(bolt) || svHost(t, g, volute) != bolt {
		t.Error("nothing moved")
	}
}

// The first response ruling: the enchanted card exiled in response
// makes the Volute fall off, and as the trigger resolves the Volute
// enchanted nothing, so nothing is copied.
func TestSpellweaverVoluteCardExiledInResponseCopiesNothing(t *testing.T) {
	g, me, _, bolt, volute := svSetup(t)

	svCastSorcery(t, g)
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(bolt); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(volute) {
		t.Errorf("the Volute is in %v, want its owner's graveyard", adZone(g, volute))
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies of a card the Volute no longer enchanted", n)
	}
}

// The second response ruling: the Volute leaves first, then its card is
// exiled. The trigger reads the card it last enchanted and copies it;
// casting the copy exiles nothing more and attaches nothing.
func TestSpellweaverVoluteRemovedThenCardExiledStillCopies(t *testing.T) {
	g, me, opp, bolt, volute := svSetup(t)
	shock := gyCard(opp, "Shock", "Instant", "{R}", shockOracleForCopies)

	svCastSorcery(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(volute); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(bolt); err != nil {
			t.Fatalf("exile: %v", err)
		}
	})
	passUntilCopies(t, g)
	cp := copyOf(t, g, "Lightning Bolt")
	if !me.Graveyard.Contains(cp.InstanceID) {
		t.Error("the copy exists in the graveyard")
	}
	before := opp.Life
	if err := svCastCopy(g, me.ID, cp.InstanceID, playerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the copy: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("life = %d, want %d", opp.Life, before-3)
	}
	if !me.Graveyard.Contains(volute) || !opp.Graveyard.Contains(shock) {
		t.Error("a Volute that has left attaches to nothing")
	}
}

// The Volute removed in response while its card stays: the copy is
// still made, and casting it still exiles that card.
func TestSpellweaverVoluteRemovedInResponseStillExilesItsCard(t *testing.T) {
	g, me, opp, bolt, volute := svSetup(t)

	svCastSorcery(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(volute); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
	passUntilCopies(t, g)
	cp := copyOf(t, g, "Lightning Bolt")
	if err := svCastCopy(g, me.ID, cp.InstanceID, playerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the copy: %v", err)
	}
	if !inExile(g, bolt) {
		t.Error("the card it last enchanted is exiled")
	}
	if !me.Graveyard.Contains(volute) {
		t.Errorf("the Volute is in %v", adZone(g, volute))
	}
}

// The offer survives a restore with the card it names: the restored
// cast still exiles the original and moves the Volute.
func TestSpellweaverVoluteOfferSurvivesARestore(t *testing.T) {
	g, me, opp, bolt, volute := svSetup(t)
	shock := gyCard(opp, "Shock", "Instant", "{R}", shockOracleForCopies)

	svCastSorcery(t, g)
	passUntilCopies(t, g)
	cp := copyOf(t, g, "Lightning Bolt")

	r := restoreRoundTrip(t, g, false)
	rme, ropp := r.PlayerByIDForEffect(me.ID), r.PlayerByIDForEffect(opp.ID)
	if err := svCastCopy(r, rme.ID, cp.InstanceID, playerTarget(ropp.ID)); err != nil {
		t.Fatalf("cast the restored copy: %v", err)
	}
	if !inExile(r, bolt) || svHost(t, r, volute) != shock {
		t.Errorf("restored: bolt exiled %v, Volute on %v (want the Shock)", inExile(r, bolt), svHost(t, r, volute))
	}
}

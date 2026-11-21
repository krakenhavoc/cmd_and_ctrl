package testsbr

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mizzix's Mastery, targeted: the card is exiled, a copy of it is made
// and offered for free, the copy's cast does what the card does, the
// original stays exiled and Mastery itself is exiled.
func TestMizzixsMasteryCopiesAndCastsTheExiledCardForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)

	mastery := effectstest.CastMizzix(t, g, false, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)

	if !effectstest.InExile(g, bolt) {
		t.Fatal("the targeted card is exiled")
	}
	if !effectstest.InExile(g, mastery) || me.Graveyard.Contains(mastery) {
		t.Error("Exile Mizzix's Mastery: the spell is exiled, not put into the graveyard")
	}
	cp := effectstest.CopyOf(t, g, "Lightning Bolt")
	if cp.InstanceID == bolt || cp.OracleID != effectstest.LightningBoltOracle || cp.Owner != me.ID {
		t.Errorf("copy = %+v: a NEW object with the card's oracle ID, owned by the caster", cp)
	}
	g.ReadSnapshot(func() {
		perm := g.CastPermissionOnCardForEffect(cp, game.ZoneExile)
		if perm == nil || perm.Player != me.ID {
			t.Fatalf("the caster has no permission over the copy: %+v", perm)
		}
		if perm.Cost != "{0}" || perm.Timing != game.TimingFlash || !perm.CastOnly {
			t.Errorf("permission = cost %q timing %q castOnly %v, want {0} / flash / cast-only", perm.Cost, perm.Timing, perm.CastOnly)
		}
	})
	if err := effectstest.CastCopy(g, opp.ID, cp.InstanceID, false, effectstest.PlayerTarget(me.ID)); err == nil {
		t.Error("an opponent may cast the copy")
	}

	before := opp.Life
	// Strict mana with an empty pool: only a free cast gets through.
	if err := effectstest.CastCopy(g, me.ID, cp.InstanceID, true, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("casting the copy without paying {R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("life = %d, want %d: the copy resolves as Lightning Bolt", opp.Life, before-3)
	}
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies left: a resolved copy ceases to exist", n)
	}
	if g.FindCardZoneForEffect(cp.InstanceID) != nil {
		t.Error("the resolved copy is in a zone")
	}
	if !effectstest.InExile(g, bolt) {
		t.Error("the original stays exiled")
	}
	if me.Graveyard.Contains(cp.InstanceID) || opp.Graveyard.Contains(cp.InstanceID) {
		t.Error("a copy of a card is not put into a graveyard")
	}
}

// Declining the offer ends the copy. Passing priority is not a decline:
// nobody has priority while Mastery resolves (CR 608.2g, #2947).
func TestMizzixsMasteryUncastCopyCeasesWhenThePlayerDeclines(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)

	effectstest.CastMizzix(t, g, false, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.CardCopies(g)); n != 1 {
		t.Fatalf("%d copies after resolution, want 1 waiting to be cast", n)
	}
	if err := g.PassPriority(); err == nil {
		t.Fatal("a pass went through while Mastery was resolving")
	}
	effectstest.DeclineResolutionCasts(t, g, me.ID)
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies after the caster declined, want none (CR 704.5e)", n)
	}
	if !effectstest.InExile(g, bolt) {
		t.Error("the original stays exiled")
	}
}

// Overload: every instant and sorcery card in the caster's graveyard is
// exiled and copied — and nothing else, and no one else's.
func TestMizzixsMasteryOverloadOffersACopyOfEverythingAtOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	shock := effectstest.GyCard(me, "Shock", "Instant", "{R}", effectstest.ShockOracleForCopies)
	div := effectstest.GyCard(me, "Divination", "Sorcery", "{2}{U}", effectstest.DivinationOracle)
	bear := effectstest.GyCard(me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")
	theirs := effectstest.GyCard(opp, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)

	mastery := effectstest.CastMizzix(t, g, true, nil)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{bolt, shock, div} {
		if !effectstest.InExile(g, id) {
			t.Errorf("%v: every instant and sorcery card in your graveyard is exiled", id)
		}
	}
	if !me.Graveyard.Contains(bear) || !opp.Graveyard.Contains(theirs) {
		t.Error("a creature card and a card in another graveyard are left alone")
	}
	if !effectstest.InExile(g, mastery) {
		t.Error("Mastery exiles itself")
	}
	if n := len(effectstest.CardCopies(g)); n != 3 {
		t.Fatalf("%d copies, want 3", n)
	}

	// The order is the caster's: Divination first, so it is under the
	// Bolt copy and resolves last.
	divCopy := effectstest.CopyOf(t, g, "Divination")
	boltCopy := effectstest.CopyOf(t, g, "Lightning Bolt")
	handBefore := me.Hand.Size()
	if err := effectstest.CastCopy(g, me.ID, divCopy.InstanceID, true); err != nil {
		t.Fatalf("cast the Divination copy: %v", err)
	}
	if err := effectstest.CastCopy(g, me.ID, boltCopy.InstanceID, true, effectstest.PlayerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the Bolt copy after the first: %v", err)
	}
	if top := g.Stack.Cards[len(g.Stack.Cards)-1]; top.InstanceID != boltCopy.InstanceID {
		t.Errorf("the last copy cast is on top of the stack, got %q", top.Name)
	}
	// The third copy (Shock) is never cast: declining the rest ends it.
	effectstest.DeclineResolutionCasts(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies left, want none: cast ones resolved, the uncast one ceased", n)
	}
	if me.Hand.Size() != handBefore+2 {
		t.Errorf("hand grew by %d, want 2 (the Divination copy resolved)", me.Hand.Size()-handBefore)
	}
}

// A target that left the graveyard in response fizzles Mastery: it goes
// to the graveyard like any fizzled spell and makes no copy.
func TestMizzixsMasteryFizzlesWhenItsTargetLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)

	mastery := effectstest.CastMizzix(t, g, false, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(bolt); err != nil {
			t.Fatalf("bounce the target: %v", err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.CardCopies(g)); n != 0 {
		t.Errorf("%d copies, want none: the spell fizzled", n)
	}
	if !me.Graveyard.Contains(mastery) {
		t.Error("a fizzled Mastery goes to the graveyard")
	}
}

// With nothing to exile an overloaded Mastery still exiles itself.
func TestMizzixsMasteryOverloadWithAnEmptyGraveyardStillExilesItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mastery := effectstest.CastMizzix(t, g, true, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, mastery) || me.Graveyard.Contains(mastery) {
		t.Error("Exile Mizzix's Mastery holds with nothing exiled before it")
	}
}

// The window survives a snapshot: the copy, its marker and its offer
// come back, and the restored copy is cast and ceases like the original.
func TestMizzixsMasteryCopyAndOfferSurviveARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	effectstest.CastMizzix(t, g, false, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)
	cp := effectstest.CopyOf(t, g, "Lightning Bolt")

	r := effectstest.RestoreRoundTrip(t, g, false)
	rc := effectstest.CopyOf(t, r, "Lightning Bolt")
	if rc.InstanceID != cp.InstanceID || !rc.CardCopy {
		t.Fatalf("restored copy = %+v, want the same object, still a copy", rc)
	}
	rme, ropp := r.PlayerByIDForEffect(me.ID), r.PlayerByIDForEffect(opp.ID)
	before := ropp.Life
	if err := effectstest.CastCopy(r, rme.ID, rc.InstanceID, true, effectstest.PlayerTarget(ropp.ID)); err != nil {
		t.Fatalf("cast the restored copy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, r)
	if ropp.Life != before-3 || len(effectstest.CardCopies(r)) != 0 {
		t.Errorf("restored copy: life %d (want %d), %d copies left", ropp.Life, before-3, len(effectstest.CardCopies(r)))
	}
}

// A copy that is NOT offered at the printed cost is charged for it: the
// helper's Cost "" is "cast the copy", not "cast it free".
func TestAPricedCardCopyChargesItsManaCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := effectstest.GyCard(me, "Lightning Bolt", "Instant", "{R}", effectstest.LightningBoltOracle)
	effectstest.ToMainForCost(t, g)
	var copies []uuid.UUID
	g.WithWriteLock(func() {
		copies = g.CopyCardsAndOfferCastsForEffect(game.CardCopyOffer{
			Player: me.ID, SourceName: "Test", Timing: game.TimingFlash, Lapse: game.LapseStaysInExile,
		}, []uuid.UUID{bolt})
	})
	if len(copies) != 1 {
		t.Fatalf("copies = %v, want 1 (made in the graveyard the card is in)", copies)
	}
	if !me.Graveyard.Contains(copies[0]) {
		t.Fatal("a copy of a graveyard card is made in a graveyard")
	}
	err := g.CastSpell(me.ID, copies[0], game.CastSpellParams{
		FromZone: "graveyard", Strict: true, Targets: []game.TargetRef{effectstest.PlayerTarget(opp.ID)},
	})
	if err == nil {
		t.Error("cast the copy with an empty pool under strict mana: the printed {R} must be charged")
	}
}

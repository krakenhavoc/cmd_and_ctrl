package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// card_copy_cast_test.go — #3258, ADR 0066's 2026-10-13 card-copy
// amendment: Mizzix's Mastery and Arcane Bombardment over CopyAndCast.

const (
	mizzixsMasteryOracle   = "40362fe0-a1a9-4d76-8c35-eac474b91af5"
	arcaneBombardmentOracl = "1e36a42f-58e2-49bc-8c14-7f75e1d8ef98"
	shockOracleForCopies   = "a9d288b8-cdc1-4e55-a0c9-d6edfc95e65d"
	divinationOracle       = "273b339c-964b-4a18-8eb5-ceb8abcdfd9e"
)

// gyCard puts a catalog card with a real mana cost into p's graveyard.
func gyCard(p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
		OracleID: oracle, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// cardCopies lists every card copy (Card.CardCopy) in every zone but
// the stack, the battlefield and the libraries.
func cardCopies(g *game.Game) []game.Card {
	var out []game.Card
	scan := func(z *game.Zone) {
		for _, c := range z.Cards {
			if c.CardCopy {
				out = append(out, c)
			}
		}
	}
	scan(g.Exile)
	for _, p := range g.Seats {
		scan(p.Hand)
		scan(p.Graveyard)
	}
	return out
}

// copyOf finds the copy whose printed name is name, or fails.
func copyOf(t *testing.T, g *game.Game, name string) game.Card {
	t.Helper()
	for _, c := range cardCopies(g) {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("no copy of %q in exile or a graveyard", name)
	return game.Card{}
}

// castMizzix casts Mizzix's Mastery from the active player's hand.
func castMizzix(t *testing.T, g *game.Game, overload bool, targets []game.TargetRef) uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Mizzix's Mastery", TypeLine: "Sorcery", ManaCost: "{3}{R}",
		OracleID: mizzixsMasteryOracle, Owner: active.ID, Controller: active.ID,
	})
	toMainForCost(t, g)
	params := game.CastSpellParams{Targets: targets}
	if overload {
		params.AlternativeCost = "overload"
		params.Targets = nil
	}
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("CastSpell Mizzix's Mastery: %v", err)
	}
	return id
}

func castCopy(g *game.Game, who, copyID uuid.UUID, strict bool, targets ...game.TargetRef) error {
	return g.CastSpell(who, copyID, game.CastSpellParams{FromZone: "exile", Strict: strict, Targets: targets})
}

func playerTarget(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id}
}

// Mizzix's Mastery, targeted: the card is exiled, a copy of it is made
// and offered for free, the copy's cast does what the card does, the
// original stays exiled and Mastery itself is exiled.
func TestMizzixsMasteryCopiesAndCastsTheExiledCardForFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)

	mastery := castMizzix(t, g, false, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)

	if !inExile(g, bolt) {
		t.Fatal("the targeted card is exiled")
	}
	if !inExile(g, mastery) || me.Graveyard.Contains(mastery) {
		t.Error("Exile Mizzix's Mastery: the spell is exiled, not put into the graveyard")
	}
	cp := copyOf(t, g, "Lightning Bolt")
	if cp.InstanceID == bolt || cp.OracleID != lightningBoltOracle || cp.Owner != me.ID {
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
	if err := castCopy(g, opp.ID, cp.InstanceID, false, playerTarget(me.ID)); err == nil {
		t.Error("an opponent may cast the copy")
	}

	before := opp.Life
	// Strict mana with an empty pool: only a free cast gets through.
	if err := castCopy(g, me.ID, cp.InstanceID, true, playerTarget(opp.ID)); err != nil {
		t.Fatalf("casting the copy without paying {R}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != before-3 {
		t.Errorf("life = %d, want %d: the copy resolves as Lightning Bolt", opp.Life, before-3)
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left: a resolved copy ceases to exist", n)
	}
	if g.FindCardZoneForEffect(cp.InstanceID) != nil {
		t.Error("the resolved copy is in a zone")
	}
	if !inExile(g, bolt) {
		t.Error("the original stays exiled")
	}
	if me.Graveyard.Contains(cp.InstanceID) || opp.Graveyard.Contains(cp.InstanceID) {
		t.Error("a copy of a card is not put into a graveyard")
	}
}

// Declining the offer ends the copy. Passing priority is not a decline:
// nobody has priority while Mastery resolves (CR 608.2g, #2947).
func TestMizzixsMasteryUncastCopyCeasesWhenThePlayerDeclines(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)

	castMizzix(t, g, false, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)
	if n := len(cardCopies(g)); n != 1 {
		t.Fatalf("%d copies after resolution, want 1 waiting to be cast", n)
	}
	if err := g.PassPriority(); err == nil {
		t.Fatal("a pass went through while Mastery was resolving")
	}
	declineResolutionCasts(t, g, me.ID)
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies after the caster declined, want none (CR 704.5e)", n)
	}
	if !inExile(g, bolt) {
		t.Error("the original stays exiled")
	}
}

// Overload: every instant and sorcery card in the caster's graveyard is
// exiled and copied — and nothing else, and no one else's.
func TestMizzixsMasteryOverloadOffersACopyOfEverythingAtOnce(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	shock := gyCard(me, "Shock", "Instant", "{R}", shockOracleForCopies)
	div := gyCard(me, "Divination", "Sorcery", "{2}{U}", divinationOracle)
	bear := gyCard(me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")
	theirs := gyCard(opp, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)

	mastery := castMizzix(t, g, true, nil)
	passPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{bolt, shock, div} {
		if !inExile(g, id) {
			t.Errorf("%v: every instant and sorcery card in your graveyard is exiled", id)
		}
	}
	if !me.Graveyard.Contains(bear) || !opp.Graveyard.Contains(theirs) {
		t.Error("a creature card and a card in another graveyard are left alone")
	}
	if !inExile(g, mastery) {
		t.Error("Mastery exiles itself")
	}
	if n := len(cardCopies(g)); n != 3 {
		t.Fatalf("%d copies, want 3", n)
	}

	// The order is the caster's: Divination first, so it is under the
	// Bolt copy and resolves last.
	divCopy := copyOf(t, g, "Divination")
	boltCopy := copyOf(t, g, "Lightning Bolt")
	handBefore := me.Hand.Size()
	if err := castCopy(g, me.ID, divCopy.InstanceID, true); err != nil {
		t.Fatalf("cast the Divination copy: %v", err)
	}
	if err := castCopy(g, me.ID, boltCopy.InstanceID, true, playerTarget(opp.ID)); err != nil {
		t.Fatalf("cast the Bolt copy after the first: %v", err)
	}
	if top := g.Stack.Cards[len(g.Stack.Cards)-1]; top.InstanceID != boltCopy.InstanceID {
		t.Errorf("the last copy cast is on top of the stack, got %q", top.Name)
	}
	// The third copy (Shock) is never cast: declining the rest ends it.
	declineResolutionCasts(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies left, want none: cast ones resolved, the uncast one ceased", n)
	}
	if me.Hand.Size() != handBefore+2 {
		t.Errorf("hand grew by %d, want 2 (the Divination copy resolved)", me.Hand.Size()-handBefore)
	}
}

// A target that left the graveyard in response fizzles Mastery: it goes
// to the graveyard like any fizzled spell and makes no copy.
func TestMizzixsMasteryFizzlesWhenItsTargetLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)

	mastery := castMizzix(t, g, false, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	g.WithWriteLock(func() {
		if err := g.BounceToHandForEffect(bolt); err != nil {
			t.Fatalf("bounce the target: %v", err)
		}
	})
	passPriorityAroundTable(t, g)
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies, want none: the spell fizzled", n)
	}
	if !me.Graveyard.Contains(mastery) {
		t.Error("a fizzled Mastery goes to the graveyard")
	}
}

// With nothing to exile an overloaded Mastery still exiles itself.
func TestMizzixsMasteryOverloadWithAnEmptyGraveyardStillExilesItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	mastery := castMizzix(t, g, true, nil)
	passPriorityAroundTable(t, g)
	if !inExile(g, mastery) || me.Graveyard.Contains(mastery) {
		t.Error("Exile Mizzix's Mastery holds with nothing exiled before it")
	}
}

// The window survives a snapshot: the copy, its marker and its offer
// come back, and the restored copy is cast and ceases like the original.
func TestMizzixsMasteryCopyAndOfferSurviveARestore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	castMizzix(t, g, false, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	passPriorityAroundTable(t, g)
	cp := copyOf(t, g, "Lightning Bolt")

	r := restoreRoundTrip(t, g, false)
	rc := copyOf(t, r, "Lightning Bolt")
	if rc.InstanceID != cp.InstanceID || !rc.CardCopy {
		t.Fatalf("restored copy = %+v, want the same object, still a copy", rc)
	}
	rme, ropp := r.PlayerByIDForEffect(me.ID), r.PlayerByIDForEffect(opp.ID)
	before := ropp.Life
	if err := castCopy(r, rme.ID, rc.InstanceID, true, playerTarget(ropp.ID)); err != nil {
		t.Fatalf("cast the restored copy: %v", err)
	}
	passPriorityAroundTable(t, r)
	if ropp.Life != before-3 || len(cardCopies(r)) != 0 {
		t.Errorf("restored copy: life %d (want %d), %d copies left", ropp.Life, before-3, len(cardCopies(r)))
	}
}

// A copy that is NOT offered at the printed cost is charged for it: the
// helper's Cost "" is "cast the copy", not "cast it free".
func TestAPricedCardCopyChargesItsManaCost(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bolt := gyCard(me, "Lightning Bolt", "Instant", "{R}", lightningBoltOracle)
	toMainForCost(t, g)
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
		FromZone: "graveyard", Strict: true, Targets: []game.TargetRef{playerTarget(opp.ID)},
	})
	if err == nil {
		t.Error("cast the copy with an empty pool under strict mana: the printed {R} must be charged")
	}
}

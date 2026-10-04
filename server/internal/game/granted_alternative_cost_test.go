package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// granted_alternative_cost_test.go — ADR 0118 §3, #2163. The seam
// against stubbed catalog hooks: the four cards land in PR 6, so the
// grantors here are test declarations spelled exactly as the effects
// package's constructors build them.
//
// What is worth pinning is what the ADR argued about: where a granted
// offer may be claimed (CR 118.9a — only where the printed mana cost
// could be paid), that it is priced like any other alternative cost
// (CR 118.9d's commander tax, CR 107.3b's X), that it keeps the
// spell's own timing, that duplicates are dropped (call 3), and that a
// claim survives a restore point.

const (
	testJodah       = "test-2163-jodah"
	testJodahToo    = "test-2163-fist"
	testOmniscience = "test-2163-omniscience"
)

// testPayWUBRG is effects.PayWUBRGForSpellsYouCast, spelled out so the
// game package can test the seam without importing the catalog.
func testPayWUBRG() GrantedAlternativeCost {
	return GrantedAlternativeCost{Offer: AlternativeCost{
		Key:      "granted-wubrg",
		Label:    "Pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost",
		ManaCost: "{W}{U}{B}{R}{G}",
	}}
}

// testFreeFromHand is effects.CastFromHandWithoutPayingManaCost.
func testFreeFromHand() GrantedAlternativeCost {
	return GrantedAlternativeCost{
		Offer: AlternativeCost{Key: "granted-free", Label: "Cast it without paying its mana cost"},
		Zones: []ZoneKind{ZoneHand},
	}
}

// withGrantedAltCosts stubs the granted-alternative-cost hook with the
// three test grantors.
func withGrantedAltCosts(t *testing.T) {
	t.Helper()
	prev := CatalogGrantedAlternativeCosts
	CatalogGrantedAlternativeCosts = func(key string) []GrantedAlternativeCost {
		switch key {
		case testJodah, testJodahToo:
			return []GrantedAlternativeCost{testPayWUBRG()}
		case testOmniscience:
			return []GrantedAlternativeCost{testFreeFromHand()}
		}
		return nil
	}
	t.Cleanup(func() { CatalogGrantedAlternativeCosts = prev })
}

// grantor puts a test grantor on the battlefield under p's control.
func grantor(t *testing.T, g *Game, p *Player, name, oracle string) uuid.UUID {
	t.Helper()
	return modifierSource(t, g, p, name, oracle)
}

// offerKeys lists what CastOffersForLocked offers a cast of `id` out of
// `zone`, the printed cost spelled "printed", in list order.
func offerKeys(t *testing.T, g *Game, p *Player, id uuid.UUID, zone ZoneKind) []string {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	var out []string
	for _, ac := range g.CastOffersForLocked(p.ID, c, zone, g.CastPermissionForLocked(p.ID, c, zone)) {
		if ac == nil {
			out = append(out, "printed")
			continue
		}
		out = append(out, ac.Key)
	}
	return out
}

// grantedOffer returns the listed offer named `key`, or nil.
func grantedOffer(t *testing.T, g *Game, p *Player, id uuid.UUID, zone ZoneKind, key string) *AlternativeCost {
	t.Helper()
	c, ok := g.LookupCardForEffect(id)
	if !ok {
		t.Fatalf("card %s not found", id)
	}
	for _, ac := range g.CastOffersForLocked(p.ID, c, zone, g.CastPermissionForLocked(p.ID, c, zone)) {
		if ac != nil && ac.Key == key {
			return ac
		}
	}
	return nil
}

func hasKey(keys []string, key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}

// fundWUBRG puts one mana of each colour in p's pool.
func fundWUBRG(p *Player) {
	p.ManaPool.AddMana(ManaToken{Color: "W"}, ManaToken{Color: "U"}, ManaToken{Color: "B"},
		ManaToken{Color: "R"}, ManaToken{Color: "G"})
}

// --- where it is offered --------------------------------------------

// The headline: a hand card is offered {W}{U}{B}{R}{G} beside its
// printed cost, labelled after its source, and the cast pays exactly
// five mana — instead of the printed cost, not on top of it.
func TestGrantedWUBRGIsOfferedAndPaidFromHand(t *testing.T) {
	withGrantedAltCosts(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	grantor(t, g, me, "Test Jodah", testJodah)
	id := spellInHand(t, g, me, "Test Bringer", "Creature — Bringer", "{7}{B}{B}")

	if got := offerKeys(t, g, me, id, ZoneHand); strings.Join(got, ",") != "printed,granted-wubrg" {
		t.Fatalf("hand offers = %v, want [printed granted-wubrg]", got)
	}
	offer := grantedOffer(t, g, me, id, ZoneHand, "granted-wubrg")
	if !offer.Granted {
		t.Errorf("the listed offer is not marked Granted")
	}
	if want := "Pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost (Test Jodah)"; offer.Label != want {
		t.Errorf("label = %q, want %q", offer.Label, want)
	}

	fundWUBRG(me)
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AlternativeCost: "granted-wubrg"}); err != nil {
		t.Fatalf("cast for {W}{U}{B}{R}{G}: %v", err)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("pool after the cast = %+v, want empty", me.ManaPool)
	}
	if item := g.StackMeta[id]; item == nil || item.AltCost != "granted-wubrg" {
		t.Errorf("stack item did not record the granted claim: %+v", item)
	}
}

// The offer is a static of the PERMANENT: it reaches only its
// controller's spells, and ends the moment the source leaves.
func TestGrantedOfferFollowsItsSourceAndController(t *testing.T) {
	withGrantedAltCosts(t)
	g := newActiveGame(t)
	me, them := g.Seats[0], g.Seats[1]
	src := grantor(t, g, me, "Test Jodah", testJodah)
	mine := spellInHand(t, g, me, "Mine", "Sorcery", "{3}{U}")
	theirs := NewCard("Theirs", them.ID)
	theirs.TypeLine, theirs.ManaCost = "Sorcery", "{3}{U}"
	them.Hand.PushTop(theirs)

	if got := offerKeys(t, g, them, theirs.InstanceID, ZoneHand); hasKey(got, "granted-wubrg") {
		t.Errorf("an opponent's spell was offered the controller's grant: %v", got)
	}
	if got := offerKeys(t, g, me, mine, ZoneHand); !hasKey(got, "granted-wubrg") {
		t.Fatalf("the controller's spell has no granted offer: %v", got)
	}
	g.WithWriteLock(func() {
		if _, err := MoveCard(g.Battlefield, me.Graveyard, src); err != nil {
			t.Fatalf("remove the source: %v", err)
		}
	})
	if got := offerKeys(t, g, me, mine, ZoneHand); hasKey(got, "granted-wubrg") {
		t.Errorf("the granted offer outlived its source: %v", got)
	}
	if err := g.CastSpell(me.ID, mine, CastSpellParams{AlternativeCost: "granted-wubrg"}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("claiming a grant whose source is gone: got %v, want ErrInvalidParam", err)
	}
}

// CR 118.9d, 903.8: the tax is an additional cost, so a commander cast
// for {W}{U}{B}{R}{G} from the command zone still pays it.
func TestGrantedWUBRGFromTheCommandZoneKeepsTheTax(t *testing.T) {
	withGrantedAltCosts(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantor(t, g, me, "Test Jodah", testJodah)
	c := NewCard("Test Commander", me.ID)
	c.TypeLine = "Legendary Creature — Bear"
	c.ManaCost = "{4}{G}{G}"
	me.Command.PushTop(c)
	me.CommanderCasts[c.InstanceID] = 1

	if got := offerKeys(t, g, me, c.InstanceID, ZoneCommand); !hasKey(got, "granted-wubrg") {
		t.Fatalf("command-zone offers = %v, want granted-wubrg among them", got)
	}
	card, _ := g.LookupCardForEffect(c.InstanceID)
	price, err := g.PriceCast(me.ID, card, CastSpellParams{FromZone: "command", AlternativeCost: "granted-wubrg"})
	if err != nil {
		t.Fatalf("PriceCast: %v", err)
	}
	if got, want := costKey(price.Total), "generic=2 x=0 colors=[B G R U W]"; got != want {
		t.Errorf("price = %s, want %s", got, want)
	}
}

// An impulse-exiled card pays its printed cost (the permission charges
// no price of its own), so it may pay Jodah's instead.
func TestGrantedOfferReachesAnImpulseExiledCard(t *testing.T) {
	withGrantedAltCosts(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantor(t, g, me, "Test Jodah", testJodah)
	c := NewCard("Exiled Sorcery", me.ID)
	c.TypeLine, c.ManaCost = "Sorcery", "{6}{R}"
	g.Exile.PushTop(c)
	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(c.InstanceID, CastPermission{Player: me.ID, Zone: ZoneExile})
	})

	if got := offerKeys(t, g, me, c.InstanceID, ZoneExile); strings.Join(got, ",") != "printed,granted-wubrg" {
		t.Fatalf("impulse offers = %v, want [printed granted-wubrg]", got)
	}
	fundWUBRG(me)
	if err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{
		FromZone: "exile", Strict: true, AlternativeCost: "granted-wubrg",
	}); err != nil {
		t.Fatalf("impulse cast for {W}{U}{B}{R}{G}: %v", err)
	}
}

// CR 118.9a and 601.2b: one alternative cost per spell. Every path
// that is already alternatively costed — a printed flashback, a
// Snapcaster grant, cascade and discover — gets no granted offer, and
// a claim of one is refused.
func TestGrantedOfferIsRefusedWhereAnotherAlternativeCostIsOwed(t *testing.T) {
	const flashbackOracle = "test-2163-flashback"

	t.Run("a printed flashback from the graveyard", func(t *testing.T) {
		withGrantedAltCosts(t)
		withCatalogCastableZones(t, castableZonesFor(flashbackOracle, ZoneGraveyard))
		withCatalogAlternativeCosts(t, altCostFor(flashbackOracle, flashbackOffer("{2}{R}")))
		g := newActiveGame(t)
		me := g.Seats[0]
		toMainPhase(t, g)
		grantor(t, g, me, "Test Jodah", testJodah)
		id := seedGraveyard(me, "Test Looting", "Sorcery", "{R}")
		setOracle(me.Graveyard, id, flashbackOracle)

		if got := offerKeys(t, g, me, id, ZoneGraveyard); strings.Join(got, ",") != "flashback" {
			t.Errorf("graveyard offers = %v, want [flashback]", got)
		}
		err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", AlternativeCost: "granted-wubrg"})
		if !errors.Is(err, ErrCastCostRequired) {
			t.Errorf("granted claim on a flashback card: got %v, want ErrCastCostRequired", err)
		}
	})

	t.Run("a Snapcaster grant", func(t *testing.T) {
		withGrantedAltCosts(t)
		g := newActiveGame(t)
		me := g.Seats[0]
		toMainPhase(t, g)
		grantor(t, g, me, "Test Jodah", testJodah)
		id := seedGraveyard(me, "Granted Sorcery", "Sorcery", "{3}{R}")
		g.WithWriteLock(func() {
			g.GrantCastPermissionOverCardForEffect(id, CastPermission{
				Player: me.ID, Zone: ZoneGraveyard, AltCostKey: "flashback", ExileOnResolution: true,
			})
		})
		if got := offerKeys(t, g, me, id, ZoneGraveyard); strings.Join(got, ",") != "flashback" {
			t.Errorf("Snapcaster offers = %v, want [flashback]", got)
		}
		err := g.CastSpell(me.ID, id, CastSpellParams{FromZone: "graveyard", AlternativeCost: "granted-wubrg"})
		if !errors.Is(err, ErrCastCostRequired) {
			t.Errorf("granted claim under Snapcaster: got %v, want ErrCastCostRequired", err)
		}
	})

	for _, tc := range []struct {
		name  string
		grant func(g *Game, me *Player, id uuid.UUID)
	}{
		{"cascade", func(g *Game, me *Player, id uuid.UUID) { g.grantFreeCastLocked(me.ID, id, 10) }},
		{"discover", func(g *Game, me *Player, id uuid.UUID) { g.grantDiscoverCastLocked(me.ID, uuid.New(), id, 9) }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			withGrantedAltCosts(t)
			g := newActiveGame(t)
			me := g.Seats[0]
			toMainPhase(t, g)
			grantor(t, g, me, "Test Jodah", testJodah)
			c := NewCard("Hit", me.ID)
			c.TypeLine, c.ManaCost = "Sorcery", "{3}{R}"
			g.Exile.PushTop(c)
			g.WithWriteLock(func() { tc.grant(g, me, c.InstanceID) })

			if got := offerKeys(t, g, me, c.InstanceID, ZoneExile); hasKey(got, "granted-wubrg") {
				t.Errorf("%s hit offered the granted cost: %v", tc.name, got)
			}
			err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{FromZone: "exile", AlternativeCost: "granted-wubrg"})
			if !errors.Is(err, ErrCastCostRequired) {
				t.Errorf("granted claim on a %s hit: got %v, want ErrCastCostRequired", tc.name, err)
			}
		})
	}
}

// A hand permission (miracle's, #1665) opens one claim and nothing
// else, so it does not stand between a hand card and Jodah's price: the
// card may still pay its printed cost from hand.
func TestGrantedOfferSurvivesAMiracleGrantInHand(t *testing.T) {
	withGrantedAltCosts(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	grantor(t, g, me, "Test Jodah", testJodah)
	id := spellInHand(t, g, me, "Test Terminus", "Sorcery", "{4}{W}{W}")
	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(id, CastPermission{
			Player: me.ID, Zone: ZoneHand, AltCostKey: AltCostKeyMiracle, Timing: TimingFlash,
		})
	})
	if perm := grantOn(g, me.ID, id, ZoneHand); !perm.Granted() {
		t.Fatalf("the miracle grant did not land")
	}
	if got := offerKeys(t, g, me, id, ZoneHand); !hasKey(got, "granted-wubrg") {
		t.Fatalf("offers under a miracle grant = %v, want granted-wubrg", got)
	}
	fundWUBRG(me)
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AlternativeCost: "granted-wubrg"}); err != nil {
		t.Fatalf("Jodah's price under a miracle grant: %v", err)
	}
}

// setOracle stamps an oracle ID on a card already in a zone.
func setOracle(z *Zone, id uuid.UUID, oracle string) {
	for i := range z.Cards {
		if z.Cards[i].InstanceID == id {
			z.Cards[i].OracleID = oracle
		}
	}
}

// --- duplicates and lost abilities ---------------------------------

// ADR 0118 call 3: two sources granting the same key give one offer,
// labelled after the first in battlefield order.
func TestTwoGrantorsGiveOneOffer(t *testing.T) {
	withGrantedAltCosts(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	grantor(t, g, me, "Test Jodah", testJodah)
	grantor(t, g, me, "Test Fist", testJodahToo)
	id := spellInHand(t, g, me, "Spell", "Sorcery", "{8}")

	got := offerKeys(t, g, me, id, ZoneHand)
	if strings.Join(got, ",") != "printed,granted-wubrg" {
		t.Fatalf("offers = %v, want one granted-wubrg", got)
	}
	first := ""
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == testJodah || c.OracleID == testJodahToo {
			first = c.Name
			break
		}
	}
	if offer := grantedOffer(t, g, me, id, ZoneHand, "granted-wubrg"); !strings.HasSuffix(offer.Label, "("+first+")") {
		t.Errorf("label = %q, want it named after %q, the first source", offer.Label, first)
	}
}

// ADR 0118 call 3: a granted offer that only repeats a listed price is
// dropped — a Bringer's own {W}{U}{B}{R}{G}, and a printed cost of
// {W}{U}{B}{R}{G} (Sliver Queen).
func TestGrantedOfferRepeatingAListedPriceIsDropped(t *testing.T) {
	const bringer = "test-2163-bringer"

	t.Run("a Bringer keeps its own key", func(t *testing.T) {
		withGrantedAltCosts(t)
		withCatalogAlternativeCosts(t, altCostFor(bringer, AlternativeCost{
			Key: "bringer-wubrg", Label: "Pay {W}{U}{B}{R}{G} rather than pay this spell's mana cost",
			ManaCost: "{W}{U}{B}{R}{G}",
		}))
		g := newActiveGame(t)
		me := g.Seats[0]
		grantor(t, g, me, "Test Fist", testJodahToo)
		id := spellInHand(t, g, me, "Test Bringer", "Creature — Bringer", "{7}{B}{B}")
		setOracle(me.Hand, id, bringer)

		if got := offerKeys(t, g, me, id, ZoneHand); strings.Join(got, ",") != "printed,bringer-wubrg" {
			t.Errorf("Bringer offers = %v, want [printed bringer-wubrg]", got)
		}
	})

	t.Run("a printed cost of {W}{U}{B}{R}{G}", func(t *testing.T) {
		withGrantedAltCosts(t)
		g := newActiveGame(t)
		me := g.Seats[0]
		grantor(t, g, me, "Test Fist", testJodahToo)
		// Symbol order is not part of the price.
		id := spellInHand(t, g, me, "Test Sliver Queen", "Legendary Creature — Sliver", "{G}{W}{U}{B}{R}")

		if got := offerKeys(t, g, me, id, ZoneHand); strings.Join(got, ",") != "printed" {
			t.Errorf("Sliver Queen offers = %v, want [printed]", got)
		}
	})

	t.Run("a price with another component is not a repeat", func(t *testing.T) {
		withGrantedAltCosts(t)
		withCatalogAlternativeCosts(t, altCostFor(bringer, AlternativeCost{
			Key: "life-and-wubrg", Label: "Pay {W}{U}{B}{R}{G} and 2 life", ManaCost: "{W}{U}{B}{R}{G}", Life: 2,
		}))
		g := newActiveGame(t)
		me := g.Seats[0]
		grantor(t, g, me, "Test Fist", testJodahToo)
		id := spellInHand(t, g, me, "Test Card", "Sorcery", "{9}")
		setOracle(me.Hand, id, bringer)

		if got := offerKeys(t, g, me, id, ZoneHand); strings.Join(got, ",") != "printed,life-and-wubrg,granted-wubrg" {
			t.Errorf("offers = %v, want the granted offer kept beside a life-and-mana one", got)
		}
	})
}

// A permanent that has lost its abilities (Humility on Jodah, CR
// 613.1f) grants nothing.
func TestGrantorWithoutAbilitiesGrantsNothing(t *testing.T) {
	withGrantedAltCosts(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	src := grantor(t, g, me, "Test Jodah", testJodah)
	id := spellInHand(t, g, me, "Spell", "Sorcery", "{8}")
	if got := offerKeys(t, g, me, id, ZoneHand); !hasKey(got, "granted-wubrg") {
		t.Fatalf("the grantor grants nothing before losing its abilities: %v", got)
	}
	registerScopedEffectForTest(t, g, src, []Mod{LoseAllAbilitiesMod()}, IndefiniteDuration())
	readyLayers(g)

	if got := offerKeys(t, g, me, id, ZoneHand); hasKey(got, "granted-wubrg") {
		t.Errorf("a grantor with no abilities still grants: %v", got)
	}
}

// --- Omniscience ----------------------------------------------------

// The free offer reaches the hand and nothing else: not a commander in
// the command zone, not an impulse-exiled card.
func TestGrantedFreeCastReachesTheHandOnly(t *testing.T) {
	withGrantedAltCosts(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	toMainPhase(t, g)
	grantor(t, g, me, "Test Omniscience", testOmniscience)
	inHand := spellInHand(t, g, me, "Hand Sorcery", "Sorcery", "{9}{U}")
	cmd := NewCard("Test Commander", me.ID)
	cmd.TypeLine, cmd.ManaCost = "Legendary Creature — Bear", "{1}{G}"
	me.Command.PushTop(cmd)
	ex := NewCard("Exiled Sorcery", me.ID)
	ex.TypeLine, ex.ManaCost = "Sorcery", "{3}{R}"
	g.Exile.PushTop(ex)
	g.WithWriteLock(func() {
		g.GrantCastPermissionOverCardForEffect(ex.InstanceID, CastPermission{Player: me.ID, Zone: ZoneExile})
	})

	if got := offerKeys(t, g, me, inHand, ZoneHand); strings.Join(got, ",") != "printed,granted-free" {
		t.Errorf("hand offers = %v, want [printed granted-free]", got)
	}
	if got := offerKeys(t, g, me, cmd.InstanceID, ZoneCommand); hasKey(got, "granted-free") {
		t.Errorf("a commander was offered the hand-only free cast: %v", got)
	}
	if got := offerKeys(t, g, me, ex.InstanceID, ZoneExile); hasKey(got, "granted-free") {
		t.Errorf("an exiled card was offered the hand-only free cast: %v", got)
	}
	if err := g.CastSpell(me.ID, cmd.InstanceID, CastSpellParams{FromZone: "command", AlternativeCost: "granted-free"}); !errors.Is(err, ErrInvalidParam) {
		t.Errorf("free claim on a commander: got %v, want ErrInvalidParam", err)
	}

	if err := g.CastSpell(me.ID, inHand, CastSpellParams{Strict: true, AlternativeCost: "granted-free"}); err != nil {
		t.Fatalf("free cast from hand with an empty pool: %v", err)
	}
}

// CR 117.1a, 307.1: the free cast keeps the spell's own timing — a
// sorcery waits for a main phase (owner decision 3).
func TestGrantedFreeCastKeepsASorcerysTiming(t *testing.T) {
	withGrantedAltCosts(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	advanceTo(t, g, StepUpkeep)
	grantor(t, g, me, "Test Omniscience", testOmniscience)
	c := NewCard("Hand Sorcery", me.ID)
	c.TypeLine, c.ManaCost = "Sorcery", "{9}{U}"
	me.Hand.PushTop(c)

	err := g.CastSpell(me.ID, c.InstanceID, CastSpellParams{AlternativeCost: "granted-free"})
	if !errors.Is(err, ErrSorcerySpeedRequired) {
		t.Errorf("free sorcery in the upkeep: got %v, want ErrSorcerySpeedRequired", err)
	}
}

// CR 107.3b: neither the mana cost nor an alternative cost with X is
// paid, so X is 0 — and an announced X is refused, not clamped.
func TestGrantedOffersFixXAtZero(t *testing.T) {
	for _, tc := range []struct {
		oracle, key string
	}{
		{testOmniscience, "granted-free"},
		{testJodah, "granted-wubrg"},
	} {
		t.Run(tc.key, func(t *testing.T) {
			withGrantedAltCosts(t)
			g := newActiveGame(t)
			me := g.Seats[0]
			grantor(t, g, me, "Test Grantor", tc.oracle)
			id := spellInHand(t, g, me, "Test Fireball", "Sorcery", "{X}{R}")
			fundWUBRG(me)

			if err := g.CastSpell(me.ID, id, CastSpellParams{AlternativeCost: tc.key, XValue: 3}); !errors.Is(err, ErrInvalidParam) {
				t.Fatalf("X=3 under %s: got %v, want ErrInvalidParam", tc.key, err)
			}
			if err := g.CastSpell(me.ID, id, CastSpellParams{AlternativeCost: tc.key}); err != nil {
				t.Fatalf("X=0 under %s: %v", tc.key, err)
			}
		})
	}
}

// CR 118.6a: an alternative cost may be paid for a spell with no mana
// cost, so Omniscience casts Ancestral Vision from hand — and its free
// offer is not dropped as a repeat of the (unpayable) printed cost.
func TestGrantedFreeCastCastsACardWithNoManaCost(t *testing.T) {
	withGrantedAltCosts(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	grantor(t, g, me, "Test Omniscience", testOmniscience)
	id := spellInHand(t, g, me, "Test Ancestral Vision", "Sorcery", "")
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if me.Hand.Cards[i].InstanceID == id {
				me.Hand.Cards[i].Layout = "normal"
			}
		}
	})

	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true}); !errors.Is(err, ErrNoManaCost) {
		t.Fatalf("printed cast of a card with no mana cost: got %v, want ErrNoManaCost", err)
	}
	if got := offerKeys(t, g, me, id, ZoneHand); !hasKey(got, "granted-free") {
		t.Fatalf("offers = %v, want granted-free", got)
	}
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AlternativeCost: "granted-free"}); err != nil {
		t.Fatalf("free cast of a card with no mana cost: %v", err)
	}
}

// --- restore --------------------------------------------------------

// The claim is captured with the stack as its key alone, and a restored
// item still resolves: no new snapshot field (ADR 0118 §3).
func TestGrantedClaimSurvivesCaptureAndRestore(t *testing.T) {
	withGrantedAltCosts(t)
	g := newActiveGame(t)
	me := g.Seats[0]
	grantor(t, g, me, "Test Jodah", testJodah)
	id := spellInHand(t, g, me, "Test Bear", "Creature — Bear", "{7}{G}")
	fundWUBRG(me)
	if err := g.CastSpell(me.ID, id, CastSpellParams{Strict: true, AlternativeCost: "granted-wubrg"}); err != nil {
		t.Fatalf("cast: %v", err)
	}

	restored, err := g.CaptureSnapshot().Restore()
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	item := restored.StackMeta[id]
	if item == nil || item.AltCost != "granted-wubrg" {
		t.Fatalf("restored stack item lost the claim: %+v", item)
	}
	resolveTop(t, restored)
	if !restored.Battlefield.Contains(id) {
		t.Errorf("the restored spell did not resolve onto the battlefield")
	}
}

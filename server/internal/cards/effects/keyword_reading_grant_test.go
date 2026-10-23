package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// keyword_reading_grant_test.go — #2663, ADR 0067 amendment 2026-10-23:
// a layer-6 grant that reads a keyword ("Other creatures you control with
// flying have indestructible") waits for every effect that adds or removes
// that keyword (CR 613.8a, 613.8b), against the real catalog.

const (
	sepharaOracle     = "ac8952c9-bd30-48cd-b964-7e6271f9cbc8"
	kwendeOracle      = "3adebbbf-39fe-4b42-b719-2f8edb009693"
	aetherChaserOracl = "8aab0b55-2779-43c7-a3ee-151b8e5c72f3"
)

// krgHas forces a recompute and asks whether the battlefield card has the
// keyword.
func krgHas(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	var has, found bool
	g.ReadSnapshot(func() {
		c, ok := battlefieldCard(g, id)
		found = ok
		has = ok && game.HasKeyword(&c, kw)
	})
	if !found {
		t.Fatalf("card %s is not on the battlefield", id)
	}
	return has
}

// krgFlier is a non-catalog creature with printed flying.
func krgFlier(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
}

// krgGrantUntilEOT gives the creature keywords until end of turn, as a
// resolving spell would (Jump), with a timestamp later than everything
// already on the battlefield.
func krgGrantUntilEOT(t *testing.T, g *game.Game, controller, target uuid.UUID, kws ...string) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := (GrantKeywordUntilEOT{Target: target, Keywords: kws}).Apply(NewContext(g, &game.StackItem{Controller: controller})); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSepharaAndKwendeAreFull(t *testing.T) {
	for _, oracle := range []string{sepharaOracle, kwendeOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v, want Full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
		if len(spec.Static) != 1 || !spec.Static[0].ReadsAbilities || spec.Static[0].Layer != game.Layer6Ability {
			t.Errorf("%s: the grant does not declare ReadsAbilities in layer 6", spec.Name)
		}
	}
	offers := game.AlternativeCostsFor(sepharaOracle)
	if len(offers) != 1 || offers[0].Key != tapAltCostKey || offers[0].ManaCost != "{W}" || offers[0].CardPaymentCount() != 4 {
		t.Errorf("Sephara offers %+v, want one tap cost of {W} and four creatures", offers)
	}
}

// The grant sees flying however and whenever it arrived: printed on a
// non-catalog creature, from a catalog creature's printed keyword row that
// entered after Sephara (a layer-6 effect with a later timestamp), and from
// an until-end-of-turn grant that resolved after both. It skips Sephara
// herself, a creature without flying and an opponent's flier.
func TestSepharaGivesIndestructibleToFliersThatArriveOrGainFlyingLater(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sephara := pushCatalogPermanent(g, me.ID, "Sephara, Sky's Blade", "Legendary Creature — Angel", sepharaOracle, false)
	bird := krgFlier(g, me.ID, "Printed Bird")
	zahid := pushCatalogPermanent(g, me.ID, "Zahid, Djinn of the Lamp", "Legendary Creature — Djinn", zahidOracle, false)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	theirs := krgFlier(g, opp.ID, "Their Bird")

	if !krgHas(t, g, bird, "indestructible") {
		t.Error("a creature with printed flying has no indestructible")
	}
	if !krgHas(t, g, zahid, "flying") || !krgHas(t, g, zahid, "indestructible") {
		t.Error("a catalog flier that entered after Sephara has no indestructible (CR 613.8b)")
	}
	if krgHas(t, g, sephara, "indestructible") {
		t.Error("Sephara gave herself indestructible (\"other\")")
	}
	if krgHas(t, g, bear, "indestructible") {
		t.Error("a creature without flying has indestructible")
	}
	if krgHas(t, g, theirs, "indestructible") {
		t.Error("an opponent's flier has indestructible")
	}

	krgGrantUntilEOT(t, g, me.ID, bear, "flying")
	if !krgHas(t, g, bear, "flying") || !krgHas(t, g, bear, "indestructible") {
		t.Error("a creature that gained flying after Sephara has no indestructible (CR 613.8b)")
	}
}

// Gravity Sphere after Sephara: in timestamp order the grant would have
// applied while the fliers still had flying and kept indestructible on
// them, stronger than printed. A flying grant newer than the Sphere gives
// flying back (it is "lose", not "can't have"), and indestructible with it.
func TestSepharaGivesNothingToACreatureThatLosesFlyingLater(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sephara := pushCatalogPermanent(g, me.ID, "Sephara, Sky's Blade", "Legendary Creature — Angel", sepharaOracle, false)
	bird := krgFlier(g, me.ID, "Printed Bird")
	other := krgFlier(g, me.ID, "Second Bird")
	if !krgHas(t, g, bird, "indestructible") {
		t.Fatal("no indestructible before the Sphere")
	}
	pushCatalogPermanent(g, me.ID, "Gravity Sphere", "World Enchantment", gravitySphereOracle, false)
	for _, id := range []uuid.UUID{sephara, bird, other} {
		if krgHas(t, g, id, "flying") {
			t.Fatal("Gravity Sphere left a creature flying")
		}
	}
	if krgHas(t, g, bird, "indestructible") || krgHas(t, g, other, "indestructible") {
		t.Error("a creature that lost flying to a later Gravity Sphere kept indestructible (CR 613.8b)")
	}
	krgGrantUntilEOT(t, g, me.ID, bird, "flying")
	if !krgHas(t, g, bird, "flying") || !krgHas(t, g, bird, "indestructible") {
		t.Error("a flying grant newer than the Sphere did not bring indestructible back")
	}
	if krgHas(t, g, other, "indestructible") {
		t.Error("the grant to one creature gave indestructible to another")
	}
}

// "Can't have flying" is stripped once the layer-6 bucket is over, after
// the grant has applied; the grant reads it anyway (HasKeywordItCanHave),
// so a creature that can't have flying gets nothing, even from a flying
// grant newer than the lockout.
func TestSepharaGivesNothingToACreatureThatCantHaveFlying(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Sephara, Sky's Blade", "Legendary Creature — Angel", sepharaOracle, false)
	bird := krgFlier(g, me.ID, "Printed Bird")
	g.WithWriteLock(func() {
		ctx := NewContext(g, &game.StackItem{Controller: me.ID})
		if err := (ScopedEffectFor{Target: bird, Mods: []game.Mod{game.CantHaveKeywordsMod("flying")},
			Duration: DurationUntilEndOfTurn(ctx), Label: "test — can't have flying"}).Apply(ctx); err != nil {
			t.Fatal(err)
		}
	})
	krgGrantUntilEOT(t, g, me.ID, bird, "flying")
	if krgHas(t, g, bird, "flying") {
		t.Fatal("a creature that can't have flying has it")
	}
	if krgHas(t, g, bird, "indestructible") {
		t.Error("a creature that can't have flying has indestructible")
	}
}

// The grant is layer-pass state, recomputed from the board after a
// restore: no new snapshot field, and the same answer.
func TestSepharasGrantSurvivesARestore(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Sephara, Sky's Blade", "Legendary Creature — Angel", sepharaOracle, false)
	zahid := pushCatalogPermanent(g, me.ID, "Zahid, Djinn of the Lamp", "Legendary Creature — Djinn", zahidOracle, false)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	krgGrantUntilEOT(t, g, me.ID, bear, "flying")
	if !krgHas(t, g, zahid, "indestructible") || !krgHas(t, g, bear, "indestructible") {
		t.Fatal("no indestructible before the restore")
	}
	r := restoreRoundTrip(t, g, true)
	if !krgHas(t, r, zahid, "indestructible") {
		t.Error("the later catalog flier lost indestructible across the restore")
	}
	if !krgHas(t, r, bear, "indestructible") {
		t.Error("the creature given flying until end of turn lost indestructible across the restore")
	}
}

// The alternative cost: {W} and four untapped fliers you control, tapped
// with Sephara on the stack. Three fliers and a bear do not pay it.
func TestSepharaIsCastForWhiteAndFourTappedFliers(t *testing.T) {
	g, me, _ := tapTable(t)
	sephara := handCardOf(g, me, "Sephara, Sky's Blade", "Legendary Creature — Angel", "{4}{W}{W}{W}", sepharaOracle)
	apaPush(g, me.ID, me.ID, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	var fliers []uuid.UUID
	for i := 0; i < 4; i++ {
		fliers = append(fliers, krgFlier(g, me.ID, "Bird"))
	}
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	if o := offerWithKey(grantedCardOffers(g, me.ID, sephara, game.ZoneHand), tapAltCostKey); o == nil {
		t.Fatal("Sephara's tap offer is not listed with four untapped fliers")
	}
	bad := append(append([]uuid.UUID(nil), fliers[:3]...), bear)
	if err := g.CastSpell(me.ID, sephara, game.CastSpellParams{AlternativeCost: tapAltCostKey, AltCostIDs: bad, Strict: true, AutoTap: true}); err == nil {
		t.Fatal("a creature without flying paid Sephara's tap cost")
	}
	if err := g.CastSpell(me.ID, sephara, game.CastSpellParams{AlternativeCost: tapAltCostKey, AltCostIDs: fliers, Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("Sephara for {W} and four fliers: %v", err)
	}
	for _, id := range fliers {
		if !tappedForTest(t, g, id) {
			t.Error("a flier named to the cost was not tapped")
		}
	}
	if tappedForTest(t, g, bear) {
		t.Error("the bear was tapped")
	}
	passPriorityAroundTable(t, g)
	if _, ok := battlefieldCard(g, sephara); !ok {
		t.Fatal("Sephara did not resolve onto the battlefield")
	}
	if !krgHas(t, g, fliers[0], "indestructible") {
		t.Error("a tapped flier has no indestructible once Sephara is out")
	}
}

// Kwende's 2018-04-27 ruling: "If a creature you control gains first
// strike after Kwende has entered the battlefield, that creature also
// gains double strike." A catalog first striker that entered later, and a
// creature given first strike later, both have double strike; a creature
// with neither does not.
func TestKwendeGivesDoubleStrikeToCreaturesThatGainFirstStrikeLater(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Kwende, Pride of Femeref", "Legendary Creature — Human Knight", kwendeOracle, false)
	chaser := pushCatalogPermanent(g, me.ID, "Aether Chaser", "Creature — Human Artificer", aetherChaserOracl, false)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	theirs := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Striker", TypeLine: "Creature — Test", Power: 1, Toughness: 1, Keywords: []string{"first strike"}})
	if !krgHas(t, g, chaser, "double strike") {
		t.Error("a first striker that entered after Kwende has no double strike (CR 613.8b)")
	}
	if krgHas(t, g, bear, "double strike") {
		t.Error("a creature without first strike has double strike")
	}
	if krgHas(t, g, theirs, "double strike") {
		t.Error("an opponent's first striker has double strike")
	}
	krgGrantUntilEOT(t, g, me.ID, bear, "first strike")
	if !krgHas(t, g, bear, "double strike") {
		t.Error("a creature given first strike after Kwende has no double strike (the 2018-04-27 ruling)")
	}
}

func TestRegisterRefusesMisplacedAbilityReaders(t *testing.T) {
	reader := WithKeywordHave(CreaturesYouControl, "flying", "test", "indestructible")
	cases := map[string]func(s *game.StaticAbility){
		"outside layer 6": func(s *game.StaticAbility) { s.Layer = game.Layer7PT; s.SubLayer = game.SubLayer7C_Modify },
		"no AppliesTo":    func(s *game.StaticAbility) { s.AppliesTo = nil },
		"a removal":       func(s *game.StaticAbility) { s.RemovesAbilities = true },
		"over spells":     func(s *game.StaticAbility) { s.AffectsSpells = true },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			s := reader
			mutate(&s)
			defer func() {
				r := recover()
				if r == nil || !strings.Contains(r.(string), "reads abilities") {
					t.Fatalf("checkAbilityReaders did not refuse it: %v", r)
				}
			}()
			checkAbilityReaders("Test Card", "static", []game.StaticAbility{s})
		})
	}
	if why := game.GrantedStaticProblem(reader); why == "" {
		t.Error("a granted bundle may carry a static that reads abilities")
	}
}

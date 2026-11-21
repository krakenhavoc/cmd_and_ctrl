package testsi

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// krgFlier is a non-catalog creature with printed flying.
func krgFlier(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.ApaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Bird", Power: 1, Toughness: 1, Keywords: []string{"flying"}})
}

func TestSepharaAndKwendeAreFull(t *testing.T) {
	for _, oracle := range []string{effectstest.SepharaOracle, effectstest.KwendeOracle} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Fatalf("%s is not registered", oracle)
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v, want Full with none", spec.Name, spec.Completeness, spec.Caveats)
		}
		if len(spec.Static) != 1 || !spec.Static[0].ReadsAbilities || spec.Static[0].Layer != game.Layer6Ability {
			t.Errorf("%s: the grant does not declare ReadsAbilities in layer 6", spec.Name)
		}
	}
	offers := game.AlternativeCostsFor(effectstest.SepharaOracle)
	if len(offers) != 1 || offers[0].Key != effectstest.TapAltCostKey || offers[0].ManaCost != "{W}" || offers[0].CardPaymentCount() != 4 {
		t.Errorf("Sephara offers %+v, want one tap cost of {W} and four creatures", offers)
	}
}

// The grant sees flying however and whenever it arrived: printed on a
// non-catalog creature, printed on a catalog creature that entered after
// Sephara, and from an until-end-of-turn grant that resolved after both. It skips Sephara
// herself, a creature without flying and an opponent's flier.
func TestSepharaGivesIndestructibleToFliersThatArriveOrGainFlyingLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sephara := effectstest.PushCatalogPermanent(g, me.ID, "Sephara, Sky's Blade", "Legendary Creature — Angel", effectstest.SepharaOracle, false)
	bird := krgFlier(g, me.ID, "Printed Bird")
	zahid := effectstest.PushCatalogPermanent(g, me.ID, "Zahid, Djinn of the Lamp", "Legendary Creature — Djinn", effectstest.ZahidOracle, false)
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	theirs := krgFlier(g, opp.ID, "Their Bird")

	if !effectstest.KrgHas(t, g, bird, "indestructible") {
		t.Error("a creature with printed flying has no indestructible")
	}
	if !effectstest.KrgHas(t, g, zahid, "flying") || !effectstest.KrgHas(t, g, zahid, "indestructible") {
		t.Error("a catalog flier that entered after Sephara has no indestructible (CR 613.8b)")
	}
	if effectstest.KrgHas(t, g, sephara, "indestructible") {
		t.Error("Sephara gave herself indestructible (\"other\")")
	}
	if effectstest.KrgHas(t, g, bear, "indestructible") {
		t.Error("a creature without flying has indestructible")
	}
	if effectstest.KrgHas(t, g, theirs, "indestructible") {
		t.Error("an opponent's flier has indestructible")
	}

	effectstest.KrgGrantUntilEOT(t, g, me.ID, bear, "flying")
	if !effectstest.KrgHas(t, g, bear, "flying") || !effectstest.KrgHas(t, g, bear, "indestructible") {
		t.Error("a creature that gained flying after Sephara has no indestructible (CR 613.8b)")
	}
}

// Gravity Sphere after Sephara: in timestamp order the grant would have
// applied while the fliers still had flying and kept indestructible on
// them, stronger than printed. A flying grant newer than the Sphere gives
// flying back (it is "lose", not "can't have"), and indestructible with it.
func TestSepharaGivesNothingToACreatureThatLosesFlyingLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sephara := effectstest.PushCatalogPermanent(g, me.ID, "Sephara, Sky's Blade", "Legendary Creature — Angel", effectstest.SepharaOracle, false)
	bird := krgFlier(g, me.ID, "Printed Bird")
	other := krgFlier(g, me.ID, "Second Bird")
	if !effectstest.KrgHas(t, g, bird, "indestructible") {
		t.Fatal("no indestructible before the Sphere")
	}
	effectstest.PushCatalogPermanent(g, me.ID, "Gravity Sphere", "World Enchantment", effectstest.GravitySphereOracle, false)
	for _, id := range []uuid.UUID{sephara, bird, other} {
		if effectstest.KrgHas(t, g, id, "flying") {
			t.Fatal("Gravity Sphere left a creature flying")
		}
	}
	if effectstest.KrgHas(t, g, bird, "indestructible") || effectstest.KrgHas(t, g, other, "indestructible") {
		t.Error("a creature that lost flying to a later Gravity Sphere kept indestructible (CR 613.8b)")
	}
	effectstest.KrgGrantUntilEOT(t, g, me.ID, bird, "flying")
	if !effectstest.KrgHas(t, g, bird, "flying") || !effectstest.KrgHas(t, g, bird, "indestructible") {
		t.Error("a flying grant newer than the Sphere did not bring indestructible back")
	}
	if effectstest.KrgHas(t, g, other, "indestructible") {
		t.Error("the grant to one creature gave indestructible to another")
	}
}

// "Can't have flying" is stripped once the layer-6 bucket is over, after
// the grant has applied; the grant reads it anyway (HasKeywordItCanHave),
// so a creature that can't have flying gets nothing, even from a flying
// grant newer than the lockout.
func TestSepharaGivesNothingToACreatureThatCantHaveFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Sephara, Sky's Blade", "Legendary Creature — Angel", effectstest.SepharaOracle, false)
	bird := krgFlier(g, me.ID, "Printed Bird")
	g.WithWriteLock(func() {
		ctx := effects.NewContext(g, &game.StackItem{Controller: me.ID})
		if err := (effects.ScopedEffectFor{Target: bird, Mods: []game.Mod{game.CantHaveKeywordsMod("flying")},
			Duration: effects.DurationUntilEndOfTurn(ctx), Label: "test — can't have flying"}).Apply(ctx); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.KrgGrantUntilEOT(t, g, me.ID, bird, "flying")
	if effectstest.KrgHas(t, g, bird, "flying") {
		t.Fatal("a creature that can't have flying has it")
	}
	if effectstest.KrgHas(t, g, bird, "indestructible") {
		t.Error("a creature that can't have flying has indestructible")
	}
}

// The grant is layer-pass state, recomputed from the board after a
// restore: no new snapshot field, and the same answer.
func TestSepharasGrantSurvivesARestore(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Sephara, Sky's Blade", "Legendary Creature — Angel", effectstest.SepharaOracle, false)
	zahid := effectstest.PushCatalogPermanent(g, me.ID, "Zahid, Djinn of the Lamp", "Legendary Creature — Djinn", effectstest.ZahidOracle, false)
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	effectstest.KrgGrantUntilEOT(t, g, me.ID, bear, "flying")
	if !effectstest.KrgHas(t, g, zahid, "indestructible") || !effectstest.KrgHas(t, g, bear, "indestructible") {
		t.Fatal("no indestructible before the restore")
	}
	r := effectstest.RestoreRoundTrip(t, g, true)
	if !effectstest.KrgHas(t, r, zahid, "indestructible") {
		t.Error("the later catalog flier lost indestructible across the restore")
	}
	if !effectstest.KrgHas(t, r, bear, "indestructible") {
		t.Error("the creature given flying until end of turn lost indestructible across the restore")
	}
}

// The alternative cost: {W} and four untapped fliers you control, tapped
// with Sephara on the stack. Three fliers and a bear do not pay it.
func TestSepharaIsCastForWhiteAndFourTappedFliers(t *testing.T) {
	g, me, _ := effectstest.TapTable(t)
	sephara := effectstest.HandCardOf(g, me, "Sephara, Sky's Blade", "Legendary Creature — Angel", "{4}{W}{W}{W}", effectstest.SepharaOracle)
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"})
	var fliers []uuid.UUID
	for i := 0; i < 4; i++ {
		fliers = append(fliers, krgFlier(g, me.ID, "Bird"))
	}
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	if o := effectstest.OfferWithKey(effectstest.GrantedCardOffers(g, me.ID, sephara, game.ZoneHand), effectstest.TapAltCostKey); o == nil {
		t.Fatal("Sephara's tap offer is not listed with four untapped fliers")
	}
	bad := append(append([]uuid.UUID(nil), fliers[:3]...), bear)
	if err := g.CastSpell(me.ID, sephara, game.CastSpellParams{AlternativeCost: effectstest.TapAltCostKey, AltCostIDs: bad, Strict: true, AutoTap: true}); err == nil {
		t.Fatal("a creature without flying paid Sephara's tap cost")
	}
	if err := g.CastSpell(me.ID, sephara, game.CastSpellParams{AlternativeCost: effectstest.TapAltCostKey, AltCostIDs: fliers, Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("Sephara for {W} and four fliers: %v", err)
	}
	for _, id := range fliers {
		if !effectstest.TappedForTest(t, g, id) {
			t.Error("a flier named to the cost was not tapped")
		}
	}
	if effectstest.TappedForTest(t, g, bear) {
		t.Error("the bear was tapped")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, sephara); !ok {
		t.Fatal("Sephara did not resolve onto the battlefield")
	}
	if !effectstest.KrgHas(t, g, fliers[0], "indestructible") {
		t.Error("a tapped flier has no indestructible once Sephara is out")
	}
}

// Kwende's 2018-04-27 ruling: "If a creature you control gains first
// strike after Kwende has entered the battlefield, that creature also
// gains double strike." A catalog first striker that entered later, and a
// creature given first strike later, both have double strike; a creature
// with neither does not.
func TestKwendeGivesDoubleStrikeToCreaturesThatGainFirstStrikeLater(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Kwende, Pride of Femeref", "Legendary Creature — Human Knight", effectstest.KwendeOracle, false)
	chaser := effectstest.PushCatalogPermanent(g, me.ID, "Aether Chaser", "Creature — Human Artificer", effectstest.AetherChaserOracl, false)
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "G")
	theirs := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Striker", TypeLine: "Creature — Test", Power: 1, Toughness: 1, Keywords: []string{"first strike"}})
	if !effectstest.KrgHas(t, g, chaser, "double strike") {
		t.Error("a first striker that entered after Kwende has no double strike (CR 613.8b)")
	}
	if effectstest.KrgHas(t, g, bear, "double strike") {
		t.Error("a creature without first strike has double strike")
	}
	if effectstest.KrgHas(t, g, theirs, "double strike") {
		t.Error("an opponent's first striker has double strike")
	}
	effectstest.KrgGrantUntilEOT(t, g, me.ID, bear, "first strike")
	if !effectstest.KrgHas(t, g, bear, "double strike") {
		t.Error("a creature given first strike after Kwende has no double strike (the 2018-04-27 ruling)")
	}
}

func TestRegisterRefusesMisplacedAbilityReaders(t *testing.T) {
	reader := effects.WithKeywordHave(effects.CreaturesYouControl, "flying", "test", "indestructible")
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
			effects.CheckAbilityReaders("Test Card", "static", []game.StaticAbility{s})
		})
	}
	if why := game.GrantedStaticProblem(reader); why == "" {
		t.Error("a granted bundle may carry a static that reads abilities")
	}
}

package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b128_cards_test.go — Theros Beyond Death cards, S82 OLD batch 128
// (#5514), played through the real catalog entries.

const (
	oldB128ForerunnerOracle = "06d2a96c-4f4d-4a4a-ac48-6f0fbc9d0afd"
	oldB128BoarOracle       = "139a5f0b-6f76-4918-96e6-cdcf27253230"
	oldB128BlessingOracle   = "17f573c4-3ed4-453e-bfbe-4e42d54633e5"
	oldB128GalliaOracle     = "43f135b5-d33e-4873-bc5a-8c939582d624"
	oldB128TrawlerOracle    = "661b2159-4040-4a44-b182-46085416ebd9"
	oldB128IncarnationOracl = "915091d5-3f58-4429-8f3a-0fd7aef87afe"
	oldB128WarCryOracle     = "947835a2-a72f-44e2-a183-99dcb019d6ff"
	oldB128InterventionOrac = "acf388b2-c4e3-4f1b-a16c-88f991d5c17b"
	oldB128DevoteeOracle    = "cacbdbcf-b5d2-40b3-bd8a-93636a248d3c"
)

func TestOldB128CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB128ForerunnerOracle: "Ashiok's Forerunner",
		oldB128BoarOracle:       "Nessian Boar",
		oldB128BlessingOracle:   "Treacherous Blessing",
		oldB128GalliaOracle:     "Gallia of the Endless Dance",
		oldB128TrawlerOracle:    "Dream Trawler",
		oldB128IncarnationOracl: "Enigmatic Incarnation",
		oldB128WarCryOracle:     "Deathbellow War Cry",
		oldB128InterventionOrac: "Nylea's Intervention",
		oldB128DevoteeOracle:    "Elspeth's Devotee",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func oldB128Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// --- Ashiok's Forerunner and Elspeth's Devotee -------------------------------

func TestOldB128NamedPlaneswalkerSearchesFetchFromLibraryAndGraveyard(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, fetches string
	}{
		{"Ashiok's Forerunner", oldB128ForerunnerOracle, "Creature — Human Wizard", "Ashiok, Sculptor of Fears"},
		{"Elspeth's Devotee", oldB128DevoteeOracle, "Creature — Human Soldier", "Elspeth, Undaunted Hero"},
	} {
		t.Run(tc.name+" from the library", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _ := oldB128Seats(g)
			ids := effectstest.SeedSearchLibrary(me,
				game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
				game.Card{Name: tc.fetches, TypeLine: "Legendary Planeswalker"})
			effectstest.LgKnowLibrary(me)
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerMayChoice(t, g, me.ID, true)
			effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneLibrary)
			effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(ids[1]) {
				t.Fatalf("%s was not fetched from the library", tc.fetches)
			}
		})
		t.Run(tc.name+" from the graveyard", func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _ := oldB128Seats(g)
			effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
			effectstest.LgKnowLibrary(me)
			pw := effectstest.PushGraveyardCardWithTypeLineID(me, tc.fetches, "Legendary Planeswalker")
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			effectstest.AnswerMayChoice(t, g, me.ID, true)
			effectstest.AnswerSearchZones(t, g, me.ID, game.SearchZoneGraveyard)
			effectstest.AnswerSearchByID(t, g, me.ID, pw)
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Hand.Contains(pw) {
				t.Fatalf("%s was not fetched from the graveyard", tc.fetches)
			}
		})
	}
}

func TestOldB128AshioksForerunnerHasFlash(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB128Seats(g)
	id := effectstest.B12Push(g, me.ID, "Ashiok's Forerunner", "Creature — Human Wizard", oldB128ForerunnerOracle, 3, 3)
	effectstest.AssertKeywords(t, g, id, "flash")
}

// --- Nessian Boar ----------------------------------------------------------------

func TestOldB128NessianBoarForcesEveryBlockerAndDrawsThemCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB128Seats(g)
	boar := effectstest.B12Push(g, me.ID, "Nessian Boar", "Creature — Boar", oldB128BoarOracle, 10, 6)
	b1 := effectstest.B12Creature(g, opp.ID, "Blocker One", "Creature — Bear", 1, 1)
	b2 := effectstest.B12Creature(g, opp.ID, "Blocker Two", "Creature — Bear", 1, 1)
	effectstest.ReqToBlockers(t, g, boar)
	effectstest.ReqRefusal(t, g.PassPriority())
	before := opp.Hand.Size()
	if err := effectstest.ReqBlock(t, g, boar, b1, b2); err != nil {
		t.Fatalf("both blockers: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("both blockers declared, the requirement is still unmet: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Hand.Size() - before; got != 2 {
		t.Errorf("the blockers' controller drew %d cards, want one per blocking creature (2)", got)
	}
}

// --- Treacherous Blessing -----------------------------------------------------------

func TestOldB128TreacherousBlessingDrawsThreeAndCostsLifePerSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB128Seats(g)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Treacherous Blessing", "Enchantment", oldB128BlessingOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	// The seeded Blessing entered the hand and left it again as it was cast.
	if got := me.Hand.Size() - hand; got != 3 {
		t.Errorf("hand changed by %d, want 3 cards drawn", got)
	}
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Test Instant", "Instant", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-1 {
		t.Errorf("life %d → %d after casting a spell, want a loss of 1", life, me.Life)
	}
}

func TestOldB128TreacherousBlessingIsSacrificedWhenTargeted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB128Seats(g)
	blessing := effectstest.PushCatalogPermanent(g, opp.ID, "Treacherous Blessing", "Enchantment", oldB128BlessingOracle, false)
	_ = me
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Test Aura Target", "Instant", "", nil)
	// A spell or ability choosing it as a target announces EventBecomesTarget.
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{Kind: game.EventBecomesTarget, Actor: me.ID, Source: uuid.New(), CardID: blessing, Target: blessing})
	})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, blessing); ok {
		t.Error("the enchantment survived being targeted")
	}
}

// --- Gallia of the Endless Dance ----------------------------------------------------

func TestOldB128GalliaPumpsOtherSatyrsAndGivesThemHaste(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB128Seats(g)
	gallia := effectstest.B12Push(g, me.ID, "Gallia of the Endless Dance", "Legendary Creature — Satyr", oldB128GalliaOracle, 2, 2)
	satyr := effectstest.B12Creature(g, me.ID, "Satyr", "Creature — Satyr", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Satyr", "Creature — Satyr", 2, 2)
	if p, _ := effectstest.PtOf(t, g, satyr); p != 3 {
		t.Errorf("my other Satyr has power %d, want 3", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, satyr, "haste") {
		t.Error("my other Satyr has no haste")
	}
	if p, _ := effectstest.PtOf(t, g, gallia); p != 2 {
		t.Errorf("Gallia pumped herself: %d", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 || effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's Satyr got the bonus")
	}
}

func TestOldB128GalliaDiscardsAtRandomForTwoCardsOnlyWhenAttackingWithThree(t *testing.T) {
	for _, attackers := range []int{2, 3} {
		g := effectstest.NewCatalogGame(t)
		me, opp := oldB128Seats(g)
		ids := []uuid.UUID{effectstest.B12Push(g, me.ID, "Gallia of the Endless Dance", "Legendary Creature — Satyr", oldB128GalliaOracle, 2, 2)}
		for i := 1; i < attackers; i++ {
			ids = append(ids, effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2))
		}
		effectstest.D2dSetHand(me, 3)
		effectstest.DeclareAttack(t, g, opp.ID, ids...)
		effectstest.PassPriorityAroundTable(t, g)
		if attackers < 3 {
			if effectstest.LatestTriggerPrompt(g, me.ID) != nil {
				t.Errorf("%d attackers triggered Gallia", attackers)
			}
			continue
		}
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if n := me.Hand.Size(); n != 3-1+2 {
			t.Errorf("hand is %d after discarding one at random and drawing two, want 4", n)
		}
	}
}

// --- Dream Trawler ----------------------------------------------------------------------

func TestOldB128DreamTrawlerDrawsOnAttackAndGrowsPerCardDrawn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB128Seats(g)
	trawler := effectstest.B12Push(g, me.ID, "Dream Trawler", "Creature — Sphinx", oldB128TrawlerOracle, 3, 5)
	effectstest.AssertKeywords(t, g, trawler, "flying", "lifelink")
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, trawler)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn for attacking", hand, me.Hand.Size())
	}
	if p, _ := effectstest.PtOf(t, g, trawler); p != 4 {
		t.Errorf("power %d after the draw, want 4", p)
	}
}

func TestOldB128DreamTrawlerDiscardsForHexproofAndTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB128Seats(g)
	trawler := effectstest.B12Push(g, me.ID, "Dream Trawler", "Creature — Sphinx", oldB128TrawlerOracle, 3, 5)
	effectstest.D2dSetHand(me, 1)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, trawler, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{me.Hand.Cards[0].InstanceID}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, trawler, "hexproof") {
		t.Error("no hexproof")
	}
	if !effectstest.TappedOnBattlefield(t, g, trawler) {
		t.Error("the Trawler is untapped")
	}
}

// --- Enigmatic Incarnation ----------------------------------------------------------------

func TestOldB128EnigmaticIncarnationFindsACreatureOneManaValueHigher(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB128Seats(g)
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Enigmatic Incarnation", TypeLine: "Enchantment",
		OracleID: oldB128IncarnationOracl, Owner: me.ID, Controller: me.ID})
	shrine := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Two Shrine", TypeLine: "Enchantment",
		ManaCost: "{1}{W}", Owner: me.ID, Controller: me.ID})
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Two Drop", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2},
		game.Card{Name: "Three Drop", TypeLine: "Creature — Bear", ManaCost: "{2}{G}", Power: 3, Toughness: 3},
		game.Card{Name: "Three Sorcery", TypeLine: "Sorcery", ManaCost: "{2}{G}"})
	effectstest.LgKnowLibrary(me)
	effectstest.AdvanceToEndStepOf(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {shrine}}) == 0 {
		t.Fatal("Enigmatic Incarnation asked for no enchantment to sacrifice")
	}
	if _, ok := effectstest.BattlefieldCard(g, shrine); ok {
		t.Fatal("the enchantment was not sacrificed")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, ids[1])
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, ids[1]); !ok {
		t.Error("the mana value 3 creature did not enter the battlefield")
	}
}

// --- Deathbellow War Cry ---------------------------------------------------------------------

func TestOldB128DeathbellowWarCryPutsOutMinotaursWithDifferentNames(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB128Seats(g)
	mino := func(name string) game.Card {
		return game.Card{Name: name, TypeLine: "Creature — Minotaur Warrior", ManaCost: "{2}{R}", Power: 3, Toughness: 3}
	}
	ids := effectstest.SeedSearchLibrary(me, mino("Alpha"), mino("Alpha"), mino("Beta"), mino("Gamma"), mino("Delta"), mino("Epsilon"),
		game.Card{Name: "Not a Minotaur", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.LgKnowLibrary(me)
	effectstest.CastCatalogSpell(t, g, "Deathbellow War Cry", "Sorcery", oldB128WarCryOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0], ids[1]}); err == nil {
		t.Fatal("two Minotaurs with the same name were accepted")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0], ids[2], ids[3], ids[4], ids[5]}); err == nil {
		t.Fatal("five cards were accepted for an up-to-four search")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0], ids[2], ids[3], ids[4]}); err != nil {
		t.Fatalf("four Minotaurs with different names: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, i := range []int{0, 2, 3, 4} {
		if _, ok := effectstest.BattlefieldCard(g, ids[i]); !ok {
			t.Errorf("card %d did not enter the battlefield", i)
		}
	}
	if _, ok := effectstest.BattlefieldCard(g, ids[1]); ok {
		t.Error("the duplicate-name Minotaur entered the battlefield")
	}
}

// --- Nylea's Intervention ---------------------------------------------------------------------

func TestOldB128NyleasInterventionFetchesUpToXLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := oldB128Seats(g)
	land := func() game.Card { return game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"} }
	ids := effectstest.SeedSearchLibrary(me, land(), land(), land(), game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	effectstest.LgKnowLibrary(me)
	effectstest.AdvanceToMain(t, g)
	effectstest.B19CastXModal(t, g, "Nylea's Intervention", "Sorcery", oldB128InterventionOrac, "{X}{G}{G}", 2, []int{0}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	effectstest.BeginSearchIfAsked(t, g, c)
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0], ids[1], ids[2]}); err == nil {
		t.Fatal("three lands were accepted for X = 2")
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[0], ids[1]}); err != nil {
		t.Fatalf("two lands: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[0]) || !me.Hand.Contains(ids[1]) {
		t.Error("the lands did not reach the hand")
	}
}

func TestOldB128NyleasInterventionDealsTwiceXToEachFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB128Seats(g)
	mineFlier := effectstest.B12Creature(g, me.ID, "My Bird", "Creature — Bird", 1, 4)
	theirFlier := effectstest.B12Creature(g, opp.ID, "Their Bird", "Creature — Bird", 1, 5)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if id := g.Battlefield.Cards[i].InstanceID; id == mineFlier || id == theirFlier {
				g.Battlefield.Cards[i].Keywords = append(g.Battlefield.Cards[i].Keywords, "flying")
			}
		}
	})
	ground := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.B19CastXModal(t, g, "Nylea's Intervention", "Sorcery", oldB128InterventionOrac, "{X}{G}{G}", 2, []int{1}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, mineFlier); ok {
		t.Error("my 1/4 flier survived 4 damage")
	}
	c, ok := effectstest.BattlefieldCard(g, theirFlier)
	if !ok || c.DamageMarked != 4 {
		t.Errorf("their 1/5 flier: present %v, damage %d, want 4", ok, c.DamageMarked)
	}
	if c, ok := effectstest.BattlefieldCard(g, ground); !ok || c.DamageMarked != 0 {
		t.Error("a creature without flying was damaged")
	}
}

package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b108_cards_test.go — Core Set 2021, batch 108 (#5511). Every test
// plays the real catalog entry through the engine.

const (
	m21B108InfernalScarring = "3fc640dd-6292-4d36-82fb-bd366bc00bd3"
	m21B108Setessan         = "3ff3b9c4-80ec-44d9-99ec-6fea3a86468b"
	m21B108DemonicEmbrace   = "d185c829-ef12-403a-a071-7520a0293d4a"
	m21B108Dub              = "e3ca6b70-01d2-44e9-937a-a2d50bbf9d10"
	m21B108Scrounger        = "595f1dee-8c3b-44a6-9e71-a4bb4334b309"
	m21B108BasrisAegis      = "9825dab8-ca1a-44ea-b1d7-77c66f00351f"
	m21B108Adherent         = "b5f5400e-f341-40e3-b935-36d9cf983a20"
	m21B108QuirionDryad     = "f42b6510-7322-4733-85ea-7ae6bd6ca8fc"
	m21B108RunedHalo        = "14cefb15-d833-4c70-80c9-f2d34b6c846b"
	m21B108Freebooter       = "8aa1bbc9-b047-4d68-93f4-728e40dafc5e"
	m21B108Stormwing        = "ea43fd04-e9fa-42fb-ba20-22dcc21c6c43"

	m21B108Bear = "Creature — Bear"
	m21B108Aura = "Enchantment — Aura"
)

func m21B108Seats(g *game.Game) (me, opp *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func TestM21B108EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		m21B108InfernalScarring: "Infernal Scarring",
		m21B108Setessan:         "Setessan Training",
		m21B108DemonicEmbrace:   "Demonic Embrace",
		m21B108Dub:              "Dub",
		m21B108Scrounger:        "Liliana's Scrounger",
		m21B108BasrisAegis:      "Basri's Aegis",
		m21B108Adherent:         "Adherent of Hope",
		m21B108QuirionDryad:     "Quirion Dryad",
		m21B108RunedHalo:        "Runed Halo",
		m21B108Freebooter:       "Kitesail Freebooter",
		m21B108Stormwing:        "Stormwing Entity",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestM21B108PrintedKeywords(t *testing.T) {
	for _, c := range []struct {
		name, typeLine, oracle string
		power, toughness       int
		keywords               []string
	}{
		{"Kitesail Freebooter", "Creature — Human Pirate", m21B108Freebooter, 1, 2, []string{"flying"}},
		{"Stormwing Entity", "Creature — Elemental", m21B108Stormwing, 3, 3, []string{"flying", "prowess"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id := effectstest.B12Push(g, g.Seats[0].ID, c.name, c.typeLine, c.oracle, c.power, c.toughness)
			effectstest.AssertKeywords(t, g, id, c.keywords...)
		})
	}
}

// --- the Auras ------------------------------------------------------------------

func TestM21B108DubAndDemonicEmbraceLayerTheirThreeEffects(t *testing.T) {
	for _, c := range []struct {
		name, oracle string
		p, tough     int
		keyword      string
		subtype      string
	}{
		{"Dub", m21B108Dub, 4, 4, "first strike", "Knight"},
		{"Demonic Embrace", m21B108DemonicEmbrace, 5, 3, "flying", "Demon"},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _ := m21B108Seats(g)
			bear := effectstest.B12Creature(g, me.ID, "Bear", m21B108Bear, 2, 2)
			effectstest.CastCatalogSpell(t, g, c.name, m21B108Aura, c.oracle, effectstest.B16TargetCard(bear))
			effectstest.PassPriorityAroundTable(t, g)
			if p, tough := effectstest.PtOf(t, g, bear); p != c.p || tough != c.tough {
				t.Errorf("the Bear is %d/%d, want %d/%d", p, tough, c.p, c.tough)
			}
			if !effectstest.HasEffectiveKeyword(t, g, bear, c.keyword) {
				t.Errorf("no %s", c.keyword)
			}
			card := effectstest.LayeredCard(t, g, bear)
			if !card.HasSubtype(c.subtype) || !card.HasSubtype("Bear") {
				t.Errorf("subtypes %v: want Bear and %s", card.Effective().Subtypes, c.subtype)
			}
		})
	}
}

func TestM21B108DemonicEmbraceCastsFromTheGraveyardForLifeAndADiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _ := m21B108Seats(g)
	for me.Hand.Size() > 0 {
		_, _ = me.Hand.PopTop()
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", m21B108Bear, 2, 2)
	embrace := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: embrace, Name: "Demonic Embrace", TypeLine: m21B108Aura, OracleID: m21B108DemonicEmbrace,
		ManaCost: "{1}{B}{B}", Owner: me.ID, Controller: me.ID})
	params := game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "s82_old_b108_life_and_discard_from_graveyard",
		Targets: effectstest.B16TargetCard(bear)}
	effectstest.ApaMana(me, "B", "B", "B")
	if err := g.CastSpell(me.ID, embrace, params); err == nil {
		t.Fatal("cast with no card to discard")
	}
	filler := effectstest.EoHand(me, "Filler", "Instant")
	life := me.Life
	params.AltCostIDs = []uuid.UUID{filler}
	if err := g.CastSpell(me.ID, embrace, params); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d, want %d", me.Life, life-3)
	}
	if me.Hand.Contains(filler) {
		t.Error("the card was not discarded")
	}
	if p, _ := effectstest.PtOf(t, g, bear); p != 5 {
		t.Errorf("power %d, want 5", p)
	}
}

func TestM21B108SetessanTrainingDrawsAndOnlyEnchantsYourCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m21B108Seats(g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", m21B108Bear, 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", m21B108Bear, 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Setessan Training", m21B108Aura, m21B108Setessan, effectstest.B16TargetCard(theirs)); err == nil {
		t.Fatal("enchanted an opponent's creature")
	}
	effectstest.CastCatalogSpell(t, g, "Setessan Training", m21B108Aura, m21B108Setessan, effectstest.B16TargetCard(mine))
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, mine); p != 3 || !effectstest.HasEffectiveKeyword(t, g, mine, "trample") {
		t.Errorf("power %d or no trample", p)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

func TestM21B108InfernalScarringDrawsWhenTheCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := m21B108Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", m21B108Bear, 2, 2)
	effectstest.CastCatalogSpell(t, g, "Infernal Scarring", m21B108Aura, m21B108InfernalScarring, effectstest.B16TargetCard(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, bear); p != 4 {
		t.Errorf("power %d, want 4", p)
	}
	hand := me.Hand.Size()
	effectstest.B27Kill(g, bear)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
}

// --- creatures ------------------------------------------------------------------

func TestM21B108QuirionDryadGrowsForWhiteBlueBlackAndRedSpells(t *testing.T) {
	for _, c := range []struct {
		color string
		grows bool
	}{{"W", true}, {"U", true}, {"B", true}, {"R", true}, {"G", false}, {"", false}} {
		g := effectstest.NewCatalogGame(t)
		me, _ := m21B108Seats(g)
		dryad := effectstest.B12Push(g, me.ID, "Quirion Dryad", "Creature — Dryad", m21B108QuirionDryad, 1, 1)
		card := game.Card{Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
		if c.color != "" {
			card.Colors = []string{c.color}
		}
		card.InstanceID = uuid.New()
		me.Hand.PushTop(card)
		effectstest.AdvanceToMain(t, g)
		if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{}); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if c.grows {
			want = 1
		}
		if got := effectstest.PlusOnes(g, dryad); got != want {
			t.Errorf("a %q spell: %d counters, want %d", c.color, got, want)
		}
	}
}

func TestM21B108AdherentOfHopeGrowsOnlyWithABasri(t *testing.T) {
	for _, basri := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, _ := m21B108Seats(g)
		adherent := effectstest.B12Push(g, me.ID, "Adherent of Hope", "Creature — Human Soldier", m21B108Adherent, 2, 1)
		name, typeLine := "Basri Ket", "Legendary Planeswalker — Basri"
		if !basri {
			name, typeLine = "Not Basri", "Legendary Planeswalker — Jace"
		}
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Owner: me.ID, Controller: me.ID,
			Counters: map[string]int{game.CounterLoyalty: 3}})
		effectstest.AdvanceTo(t, g, game.StepBeginCombat)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if basri {
			want = 1
		}
		if got := effectstest.PlusOnes(g, adherent); got != want {
			t.Errorf("a Basri %v: %d counters, want %d", basri, got, want)
		}
	}
}

func TestM21B108LilianasScroungerNeedsADeathAndALiliana(t *testing.T) {
	for _, died := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := m21B108Seats(g)
		effectstest.B12Push(g, me.ID, "Liliana's Scrounger", "Creature — Human Wizard", m21B108Scrounger, 3, 2)
		lili := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Liliana Test", TypeLine: "Legendary Planeswalker — Liliana",
			Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 3}})
		other := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Jace Test", TypeLine: "Legendary Planeswalker — Jace",
			Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 3}})
		victim := effectstest.B12Creature(g, opp.ID, "Victim", m21B108Bear, 2, 2)
		effectstest.AdvanceToMain(t, g)
		if died {
			effectstest.KillCreature(t, g, opp.ID, victim)
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		if died {
			effectstest.AnswerPermanentPicks(t, g, me.ID, map[uuid.UUID][]uuid.UUID{me.ID: {lili}})
			effectstest.PassPriorityAroundTable(t, g)
		}
		wantLili := 3
		if died {
			wantLili = 4
		}
		if got := effectstest.LoyaltyOf(g, lili); got != wantLili {
			t.Errorf("a creature died %v: Liliana has %d loyalty, want %d", died, got, wantLili)
		}
		if got := effectstest.LoyaltyOf(g, other); got != 3 {
			t.Errorf("a non-Liliana planeswalker has %d loyalty, want 3", got)
		}
	}
}

func TestM21B108StormwingEntityCostsLessAfterAnInstantOrSorcery(t *testing.T) {
	for _, cast := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me, opp := m21B108Seats(g)
		if cast {
			effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
			effectstest.PassPriorityAroundTable(t, g)
		}
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Stormwing Entity", TypeLine: "Creature — Elemental", OracleID: m21B108Stormwing,
			ManaCost: "{3}{U}{U}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
		effectstest.ApaMana(me, "U", "G")
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true})
		if cast && err != nil {
			t.Errorf("after an instant: {1}{U} should cast it: %v", err)
		}
		if !cast && err == nil {
			t.Error("cast for {1}{U} with no instant or sorcery this turn")
		}
		if cast {
			effectstest.PassPriorityAroundTable(t, g)
			if c := effectstest.ScryChoiceFor(g, me.ID); c == nil || len(c.ScryCards) != 2 {
				t.Errorf("scry 2 prompt: %+v", c)
			}
		}
	}
}

func TestM21B108BasrisAegisCountersTwoAndMayFetchBasri(t *testing.T) {
	for _, mode := range []string{"library", "decline"} {
		g := effectstest.NewCatalogGame(t)
		me, opp := m21B108Seats(g)
		a := effectstest.B12Creature(g, me.ID, "A", m21B108Bear, 2, 2)
		b := effectstest.B12Creature(g, opp.ID, "B", m21B108Bear, 2, 2)
		c := effectstest.B12Creature(g, me.ID, "C", m21B108Bear, 2, 2)
		effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
		basri := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Basri, Devoted Paladin", TypeLine: "Legendary Planeswalker — Basri"})
		effectstest.CastCatalogSpell(t, g, "Basri's Aegis", "Sorcery", m21B108BasrisAegis,
			[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
		effectstest.PassPriorityAroundTable(t, g)
		if effectstest.PlusOnes(g, a) != 1 || effectstest.PlusOnes(g, b) != 1 || effectstest.PlusOnes(g, c) != 0 {
			t.Errorf("counters %d, %d, %d; want 1, 1, 0", effectstest.PlusOnes(g, a), effectstest.PlusOnes(g, b), effectstest.PlusOnes(g, c))
		}
		if mode == "decline" {
			effectstest.AnswerMayChoice(t, g, me.ID, false)
			effectstest.PassPriorityAroundTable(t, g)
			if me.Hand.Contains(basri) {
				t.Error("declining still took Basri")
			}
			continue
		}
		effectstest.AnswerMayChoice(t, g, me.ID, true)
		effectstest.AnswerSearchByID(t, g, me.ID, basri)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Hand.Contains(basri) {
			t.Error("Basri did not reach the hand")
		}
	}
}

func TestM21B108KitesailFreebooterTakesANoncreatureNonlandCardUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := m21B108Seats(g)
	for opp.Hand.Size() > 0 {
		_, _ = opp.Hand.PopTop()
	}
	land := effectstest.EoHand(opp, "Forest", "Basic Land — Forest")
	creature := effectstest.EoHand(opp, "Bear", m21B108Bear)
	spell := effectstest.EoHand(opp, "Opt", "Instant")
	effectstest.CastCatalogSpell(t, g, "Kitesail Freebooter", "Creature — Human Pirate", m21B108Freebooter, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickPlayer(t, g, me.ID, opp.ID)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.OpenVariant(t, g)
	for _, bad := range []uuid.UUID{land, creature} {
		if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{bad}); err == nil {
			t.Fatal("a land or a creature card was accepted")
		}
	}
	if err := g.ResolvePendingChoice(pick.ID, me.ID, []uuid.UUID{spell}); err != nil {
		t.Fatalf("resolve the pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ZoneOfCard(g, spell) != game.ZoneExile {
		t.Fatalf("the instant is in %v, want exile", effectstest.ZoneOfCard(g, spell))
	}
	var freebooter uuid.UUID
	for _, card := range g.BattlefieldCardsForEffect() {
		if card.Name == "Kitesail Freebooter" {
			freebooter = card.InstanceID
		}
	}
	effectstest.B27Kill(g, freebooter)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(spell) {
		t.Errorf("the card did not return to its owner's hand (it is in %v)", effectstest.ZoneOfCard(g, spell))
	}
}

func TestM21B108RunedHaloProtectsYouFromTheChosenName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := m21B108Seats(g)
	halo := effectstest.PushCatalogPermanent(g, opp.ID, "Runed Halo", "Enchantment", m21B108RunedHalo, false)
	if abilities := effectstest.PlayerAbilities(g, opp); len(abilities) != 0 {
		t.Fatalf("unanswered, the player has %v", abilities)
	}
	g.WithWriteLock(func() { g.QueueCardNameChoiceForEffect(opp.ID, halo, "Runed Halo — choose a card name") })
	effectstest.ActNameFor(t, g, opp.ID, "Lightning Bolt")
	if !effectstest.HasPlayerAbility(effectstest.PlayerAbilities(g, opp), "protection from the card named Lightning Bolt") {
		t.Fatalf("abilities %v", effectstest.PlayerAbilities(g, opp))
	}
	if err := effectstest.CastCatalogSpellErr(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err == nil {
		t.Error("the named spell targeted the protected player")
	}
	life := opp.Life
	if err := effectstest.CastCatalogSpellErr(t, g, "Blaze", "Sorcery", effectstest.BlazeOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)}); err != nil {
		t.Fatalf("a spell with another name: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life > life {
		t.Error("life went up")
	}
	_ = me
}

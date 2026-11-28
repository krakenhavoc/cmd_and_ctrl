package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b839_cards_test.go — Ice Age's cumulative-upkeep enchantments, played from
// the real catalog entries (S82 OLD batch 839, #5636).

const (
	b839PresenceOracle = "8f319b31-efb0-4f48-a2a3-1f308e7b1dda"
	b839BrandOracle    = "947bd180-ee95-4db4-b5d9-432428018eb8"
	b839MightOracle    = "967eca77-d273-4fc6-97d1-df7d8cc6ff59"
	b839WindOracle     = "c9798915-3f9e-466b-bbd6-f4e126bdb67d"
	b839TranceOracle   = "d1cc6979-d786-4035-963b-768ac8cef116"
	b839IllusionOracle = "d7c81b3e-8082-4aaf-8979-2717d118234c"
	b839DarknessOracle = "dd821fca-79a0-48a9-b1cf-5a496bc04f65"
	b839RitualOracle   = "e0587218-206e-41ed-af3c-f06b7a668e90"
	b839AscentOracle   = "f2cedb38-12b0-4489-a62b-9b6a3776d695"
	b839SingularOracle = "fb5ec55d-4a35-432a-be6e-c295f1b2e603"
)

func TestS82OldB839CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b839PresenceOracle: "Illusionary Presence",
		b839BrandOracle:    "Brand of Ill Omen",
		b839MightOracle:    "Mystic Might",
		b839WindOracle:     "Maddening Wind",
		b839TranceOracle:   "Mesmeric Trance",
		b839IllusionOracle: "Illusions of Grandeur",
		b839DarknessOracle: "Infernal Darkness",
		b839RitualOracle:   "Ritual of Subdual",
		b839AscentOracle:   "Arnjlot's Ascent",
		b839SingularOracle: "Naked Singularity",
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

// b839Cast casts `name` from the active player's hand with the mana floated.
func b839Cast(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, targets []game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "{2}{G}{G}{B}{U}{R}{W}")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost,
			OracleID: oracle, Owner: me.ID, Controller: me.ID})
	})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets})
}

func TestInfernalDarknessMakesEveryLandTapForBlackAndChargesManaAndLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedEnchantment(g, b839DarknessOracle, "Infernal Darkness", "Enchantment", me.ID)
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.TapForMana(t, g, opp.ID, forest)
	if got := effectstest.SortedColors(opp); !slices.Equal(got, []string{"B"}) {
		t.Errorf("an opponent's Forest made %v, want [B]", got)
	}
	tomb := effectstest.SeedPermanentWithOracle(g, me.ID, "Ancient Tomb", "Land", effectstest.AncientTombOracle)
	effectstest.TapForMana(t, g, me.ID, tomb)
	if got := effectstest.SortedColors(me); !slices.Equal(got, []string{"B", "B"}) {
		t.Errorf("Ancient Tomb made %v, want [B B] (the amount is unchanged)", got)
	}
}

func TestInfernalDarknessUpkeepPaysBlackAndOneLifePerAgeCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	darkness := effectstest.PushCatalogPermanent(g, me.ID, "Infernal Darkness", "Enchantment", b839DarknessOracle, false)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.RemoraUpkeepPrompt(t, g, me.ID)
	if prompt.PayCost != "{B}" || prompt.LifeCost != 1 {
		t.Fatalf("payment = %q and %d life, want {B} and 1 life", prompt.PayCost, prompt.LifeCost)
	}
	life := me.Life
	me.ManaPool.AddMana(game.ManaToken{Color: "B"})
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if _, ok := effectstest.BattlefieldCard(g, darkness); !ok {
		t.Error("a paid upkeep sacrificed Infernal Darkness")
	}
	if me.Life != life-1 {
		t.Errorf("life %d -> %d, want 1 paid", life, me.Life)
	}
}

func TestRitualOfSubdualMakesEveryLandTapForColorless(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedEnchantment(g, b839RitualOracle, "Ritual of Subdual", "Enchantment", me.ID)
	forest := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	effectstest.TapForMana(t, g, opp.ID, forest)
	if got := effectstest.SortedColors(opp); !slices.Equal(got, []string{"C"}) {
		t.Errorf("an opponent's Forest made %v, want [C]", got)
	}
}

func TestNakedSingularitySwapsTheFiveBasicLandTypes(t *testing.T) {
	for _, tc := range []struct{ land, typeLine, want string }{
		{"Plains", "Basic Land — Plains", "R"},
		{"Island", "Basic Land — Island", "G"},
		{"Swamp", "Basic Land — Swamp", "W"},
		{"Mountain", "Basic Land — Mountain", "U"},
		{"Forest", "Basic Land — Forest", "B"},
	} {
		t.Run(tc.land, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			effectstest.SeedEnchantment(g, b839SingularOracle, "Naked Singularity", "Artifact", me.ID)
			land := effectstest.PushLandFor(g, opp.ID, tc.land, tc.typeLine)
			effectstest.TapForMana(t, g, opp.ID, land)
			if got := effectstest.SortedColors(opp); !slices.Equal(got, []string{tc.want}) {
				t.Errorf("%s made %v, want [%s]", tc.land, got, tc.want)
			}
		})
	}
}

func TestIllusionsOfGrandeurGainsTwentyOnEntryAndLosesTwentyOnLeaving(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	if err := b839Cast(t, g, "Illusions of Grandeur", "Enchantment", "{3}{U}", b839IllusionOracle, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+20 {
		t.Fatalf("life %d, want %d after entering", me.Life, life+20)
	}
	var id uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.OracleID == b839IllusionOracle {
			id = c.InstanceID
		}
	}
	if id == uuid.Nil {
		t.Fatal("Illusions of Grandeur is not on the battlefield")
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, id); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d after leaving", me.Life, life)
	}
}

func TestArnjlotsAscentGrantsFlyingUntilEndOfTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ascent := effectstest.PushCatalogPermanent(g, me.ID, "Arnjlot's Ascent", "Enchantment", b839AscentOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "G")
	effectstest.B16Activate(t, g, me.ID, ascent, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "flying") {
		t.Error("the creature did not gain flying")
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "flying") {
		t.Error("flying outlived the turn")
	}
}

func TestMesmericTranceRummagesForAUBlueMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	trance := effectstest.PushCatalogPermanent(g, me.ID, "Mesmeric Trance", "Enchantment", b839TranceOracle, false)
	fodder := game.Card{InstanceID: uuid.New(), Name: "Fodder", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Hand.PushTop(fodder) })
	hand := me.Hand.Size()
	effectstest.ApaMana(me, "U")
	effectstest.Pr7Activate(t, g, me.ID, trance, 0, game.ActivateAbilityParams{DiscardIDs: []uuid.UUID{fodder.InstanceID}})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand = %d, want %d (discard one, draw one)", me.Hand.Size(), hand)
	}
	if !me.Graveyard.Contains(fodder.InstanceID) {
		t.Error("the discarded card is not in the graveyard")
	}
}

func TestMaddeningWindDamagesTheEnchantedCreaturesControllerAtTheirUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := b839Cast(t, g, "Maddening Wind", "Enchantment — Aura", "{2}{G}", b839WindOracle, effectstest.CardTarget(host)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	oppLife, myLife := opp.Life, me.Life
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("the creature's controller is at %d, want %d", opp.Life, oppLife-2)
	}
	if me.Life != myLife {
		t.Errorf("the Aura's controller lost %d life", myLife-me.Life)
	}
}

func TestBrandOfIllOmenStopsTheEnchantedCreaturesControllerCastingCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	if err := b839Cast(t, g, "Brand of Ill Omen", "Enchantment — Aura", "{3}{R}", b839BrandOracle, effectstest.CardTarget(host)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.FloatForTest(g, opp, "{2}{G}{U}")
	creature := uuid.New()
	instant := uuid.New()
	g.WithWriteLock(func() {
		opp.Hand.PushTop(game.Card{InstanceID: creature, Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
		opp.Hand.PushTop(game.Card{InstanceID: instant, Name: "Test Cantrip", TypeLine: "Instant", ManaCost: "{U}", Owner: opp.ID, Controller: opp.ID})
	})
	if err := g.CastSpell(opp.ID, creature, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("the enchanted creature's controller cast a creature spell")
	}
	if err := g.CastSpell(opp.ID, instant, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("a noncreature spell was refused: %v", err)
	}
	_ = me
}

func TestMysticMightGivesTheLandAPumpAbility(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	land := effectstest.PushLandFor(g, me.ID, "Island", "Basic Land — Island")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	if err := b839Cast(t, g, "Mystic Might", "Enchantment — Aura", "{U}", b839MightOracle, effectstest.CardTarget(bear)); err == nil {
		t.Fatal("enchanted a creature")
	}
	if err := b839Cast(t, g, "Mystic Might", "Enchantment — Aura", "{U}", b839MightOracle, effectstest.CardTarget(land)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	idx, ref := effectstest.GrantedActivatedIndex(t, g, land)
	if idx < 0 {
		t.Fatal("the land has no granted ability")
	}
	effectstest.B16Activate(t, g, me.ID, land, idx, game.ActivateAbilityParams{Ref: ref, Targets: effectstest.CardTarget(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, bear); p != 4 || tough != 4 {
		t.Errorf("the creature is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.B16Tapped(t, g, land) {
		t.Error("the ability did not tap the land")
	}
}

func TestIllusionaryPresenceGainsLandwalkOfTheChosenType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	presence := effectstest.PushCatalogPermanent(g, me.ID, "Illusionary Presence", "Creature — Illusion", b839PresenceOracle, false)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		me.ManaPool.AddMana(game.ManaToken{Color: "U"})
		effectstest.AnswerPayUnless(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
	}
	words := effectstest.PickOptionWords(t, g, me.ID)
	idx := slices.Index(words, "Forest")
	if idx < 0 {
		t.Fatalf("the offered land types %v do not include Forest", words)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, idx)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, presence), game.LandwalkToken("Forest")) {
		t.Errorf("the creature did not gain forestwalk: %v", effectstest.EffectiveAbilities(t, g, presence))
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, presence), game.LandwalkToken("Forest")) {
		t.Error("forestwalk outlived the turn")
	}
}

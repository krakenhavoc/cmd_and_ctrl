package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	s82g58OrimsChant     = "26d3a818-7faf-443a-a6bd-9773edaedf17"
	s82g58UginsLabyrinth = "d565cd3d-68d4-4039-9e45-7e69e31d0ffb"
	s82g58SpymastersVlt  = "69ddca4b-5cc0-45f3-b2e6-a047c8d601be"
	s82g58TwistedLand    = "db659cae-2078-423e-a6ed-63898dbab87f"
	s82g58SeethingLand   = "2d8635bd-ed96-4bb1-8718-6962a0eee3d5"
	s82g58ShelteringLand = "5b932be0-4dac-41b9-9c59-f79e4cecc31a"
	s82g58Grist          = "196f8f13-d4df-4740-8746-8fed3bc71840"
	s82g58Tamiyo         = "1ed9e351-0000-4b75-8ecb-d15bd88df68a"
	s82g58Ral            = "16d40f56-691b-42a7-a5ab-157568921e03"
	s82g58Ajani          = "2588f348-d7a3-46c8-9ace-dca53ed5ef99"
	s82g58FlareCultivate = "1766f1f4-0924-4436-8379-94030adcccc5"
)

func TestS82GapB58EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82g58OrimsChant:     "Orim's Chant",
		s82g58UginsLabyrinth: "Ugin's Labyrinth",
		s82g58SpymastersVlt:  "Spymaster's Vault",
		s82g58TwistedLand:    "Twisted Landscape",
		s82g58SeethingLand:   "Seething Landscape",
		s82g58ShelteringLand: "Sheltering Landscape",
		s82g58Grist:          "Grist, Voracious Larva",
		s82g58Grist + "#1":   "Grist, the Plague Swarm",
		s82g58Tamiyo:         "Tamiyo, Inquisitive Student",
		s82g58Tamiyo + "#1":  "Tamiyo, Seasoned Scholar",
		s82g58Ral:            "Ral, Monsoon Mage",
		s82g58Ral + "#1":     "Ral, Leyline Prodigy",
		s82g58Ajani:          "Ajani, Nacatl Pariah",
		s82g58Ajani + "#1":   "Ajani, Nacatl Avenger",
		s82g58FlareCultivate: "Flare of Cultivation",
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

// --- Orim's Chant --------------------------------------------------

func TestS82GapB58OrimsChantStopsTheTargetCastingAndKickedStopsAttacks(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.P1g2Seats(g)
		mine := effectstest.P1Creature(g, me.ID, 2, 2)
		theirs := effectstest.P1Creature(g, opp.ID, 2, 2)
		params := game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
		effectstest.B06AddMana(me, "W")
		if kicked {
			params.OptionalCosts = []int{0}
			effectstest.B06AddMana(me, "W")
		}
		effectstest.P1g2Cast(t, g, "Orim's Chant", "Instant", s82g58OrimsChant, "{W}", params)
		effectstest.PassPriorityAroundTable(t, g)

		fodder := effectstest.PushCatalogHandCard(opp, "Fodder", "Instant", "")
		if err := g.CastSpell(opp.ID, fodder, game.CastSpellParams{Strict: true}); err == nil {
			t.Errorf("kicked=%v: the target player cast a spell", kicked)
		}
		late := effectstest.P1Creature(g, me.ID, 1, 1) // enters after the spell resolved
		for name, id := range map[string]uuid.UUID{"mine": mine, "theirs": theirs, "late": late} {
			banned := effectstest.RestrictionsOf(t, g, id)&game.CantAttack != 0
			if banned != kicked {
				t.Errorf("kicked=%v: %s creature can't attack = %v", kicked, name, banned)
			}
		}
		if me.ManaPool.Total() != 0 {
			t.Errorf("kicked=%v: mana left in the pool: %+v", kicked, me.ManaPool)
		}
	}
}

// --- the Landscape cycle -------------------------------------------

func TestS82GapB58LandscapesTapFetchAndCycle(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, cycle string
		fetch, refuse       []string
	}{
		{"Twisted Landscape", s82g58TwistedLand, "{B}{R}{G}", []string{"Swamp", "Mountain", "Forest"}, []string{"Island", "Plains"}},
		{"Seething Landscape", s82g58SeethingLand, "{U}{B}{R}", []string{"Island", "Swamp", "Mountain"}, []string{"Forest", "Plains"}},
		{"Sheltering Landscape", s82g58ShelteringLand, "{R}{G}{W}", []string{"Mountain", "Forest", "Plains"}, []string{"Island", "Swamp"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			var lib []game.Card
			for _, n := range append(append([]string{}, tc.fetch...), tc.refuse...) {
				lib = append(lib, game.Card{Name: n, TypeLine: "Basic Land — " + n})
			}
			effectstest.SeedSearchLibrary(me, lib...)
			land := effectstest.B43Catalog(g, me.ID, tc.name, "Land", tc.oracle, 0, 0)
			effectstest.B28TapForMana(t, g, me.ID, land, "C")
			if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "C" {
				t.Fatalf("tapped for %v, want {C}", got)
			}
			g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
			effectstest.AdvanceToMain(t, g)
			effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
			effectstest.PassPriorityAroundTable(t, g)
			c := effectstest.SearchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("the fetch opens no search")
			}
			for _, n := range tc.refuse {
				if effectstest.SearchOptionNamed(g, c, n) != uuid.Nil {
					t.Errorf("a basic %s is not among %v", n, tc.fetch)
				}
			}
			var pick uuid.UUID
			for _, n := range tc.fetch {
				id := effectstest.SearchOptionNamed(g, c, n)
				if id == uuid.Nil {
					t.Fatalf("basic %s is not offered", n)
				}
				pick = id
			}
			effectstest.AnswerSearchByID(t, g, me.ID, pick)
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(land) {
				t.Error("the Landscape was not sacrificed")
			}
			if !g.Battlefield.Contains(pick) || !effectstest.B16Tapped(t, g, pick) {
				t.Error("the basic should arrive tapped")
			}
		})
	}
}

func TestS82GapB58LandscapeCyclesForItsThreeColors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	id := effectstest.PushCatalogHandCard(me, "Twisted Landscape", "Land", s82g58TwistedLand)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{R}{G}"); err != nil {
		t.Fatal(err)
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 0 {
		t.Errorf("cycling should cost {B}{R}{G}: pool %+v", me.ManaPool)
	}
	if !me.Graveyard.Contains(id) || me.Hand.Size() != hand {
		t.Errorf("cycling should discard the card and draw one: graveyard=%v hand %d -> %d", me.Graveyard.Contains(id), hand, me.Hand.Size())
	}
}

// --- Flare of Cultivation ------------------------------------------

func TestS82GapB58FlareOfCultivationSacrificesAGreenCreatureAndFetchesTwoBasics(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var forests []uuid.UUID
	effectstest.SeedSearchLibrary(me)
	for i := 0; i < 3; i++ {
		forests = append(forests, effectstest.PushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}))
	}
	flare := effectstest.HandCardAtMain(t, g, "Flare of Cultivation", "Sorcery", s82g58FlareCultivate)
	green := effectstest.SeedColoredCreature(g, me.ID, "Elf", []string{"G"}, 0)
	token := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Saproling",
		TypeLine: "Token Creature — Saproling", Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	white := effectstest.SeedColoredCreature(g, me.ID, "Cleric", []string{"W"}, 0)
	for name, id := range map[string]uuid.UUID{"a green token": token, "a white creature": white} {
		if err := effectstest.CastForAltCost(g, flare, "", "sacrifice", []uuid.UUID{id}, nil); err == nil {
			t.Fatalf("%s paid \"sacrifice a nontoken green creature\"", name)
		}
	}
	handBefore := me.Hand.Size()
	if err := effectstest.CastForAltCost(g, flare, "", "sacrifice", []uuid.UUID{green}, nil); err != nil {
		t.Fatalf("Flare for a nontoken green creature: %v", err)
	}
	if g.Battlefield.Contains(green) {
		t.Error("the green creature was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOrDecline(t, g, me.ID, 0, forests)
	effectstest.AnswerOrDecline(t, g, me.ID, 1, forests)
	if !g.Battlefield.Contains(forests[0]) || !effectstest.IsTapped(g, forests[0]) {
		t.Error("the first basic should arrive tapped")
	}
	if !me.Hand.Contains(forests[1]) {
		t.Error("the second basic should go to the hand")
	}
	if got := me.Hand.Size() - handBefore; got != 0 {
		// the Flare left the hand (-1) and a basic joined it (+1)
		t.Errorf("hand changed by %d, want 0", got)
	}
}

// --- Spymaster's Vault ---------------------------------------------

func TestS82GapB58SpymastersVaultEntersTappedUnlessASwampAndConnivesForTheDead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	vault := importLandForTest(t, g, me, "Spymaster's Vault", s82g58SpymastersVlt)
	if !effectstest.IsTapped(g, vault) {
		t.Error("with no Swamp the Vault enters tapped")
	}
	effectstest.SeedLand(g, me.ID, "Swamp", "Basic Land — Swamp", "")
	vault2 := importLandForTest(t, g, me, "Spymaster's Vault", s82g58SpymastersVlt)
	if effectstest.IsTapped(g, vault2) {
		t.Error("with a Swamp the Vault enters untapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vault) })

	effectstest.B28TapForMana(t, g, me.ID, vault, "B")
	if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != "B" {
		t.Fatalf("tapped for %v, want {B}", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vault) })

	// Two creatures die this turn, so the connive is X = 2.
	bear := effectstest.P1Creature(g, me.ID, 2, 2)
	d1 := effectstest.P1Creature(g, me.ID, 1, 1)
	d2 := effectstest.P1Creature(g, g.Seats[1].ID, 1, 1)
	for _, id := range []uuid.UUID{d1, d2} {
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	}
	g.RunStateChecksForTest()
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Draw A", TypeLine: "Creature — A", Power: 1, Toughness: 1})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Draw B", TypeLine: "Sorcery"})
	hand := me.Hand.Size()
	effectstest.B06AddMana(me, "B")
	effectstest.B16Activate(t, g, me.ID, vault, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Fatalf("connive 2 drew %d, want 2 (before discarding)", got)
	}
	if !effectstest.B16Tapped(t, g, vault) {
		t.Error("the Vault should tap to connive")
	}
}

// --- Ugin's Labyrinth ----------------------------------------------

func TestS82GapB58UginsLabyrinthImprintsDoublesAndReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	eldrazi := me.Hand
	_ = eldrazi
	big := pushHandCardMV(me, "Kozilek", "Creature — Eldrazi", "{8}", nil)
	colored := pushHandCardMV(me, "Big Red", "Creature — Dragon", "{5}{R}{R}", []string{"R"})
	small := pushHandCardMV(me, "Wurmcoil Lite", "Artifact Creature — Wurm", "{6}", nil)
	land := importLandForTest(t, g, me, "Ugin's Labyrinth", s82g58UginsLabyrinth)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the imprint opens no pick")
	}
	if pick.ChooseMin != 0 {
		t.Errorf("imprint is \"you may\": floor %d", pick.ChooseMin)
	}
	if !effectstest.HasID(pick.ChooseCards, big) || effectstest.HasID(pick.ChooseCards, colored) || effectstest.HasID(pick.ChooseCards, small) {
		t.Errorf("only the colorless card of mana value 7+ is offered: %+v", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{big}); err != nil {
		t.Fatal(err)
	}
	if !effectstest.InExile(g, big) {
		t.Fatal("Kozilek was not exiled")
	}
	if effectstest.IsTapped(g, land) {
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	}
	effectstest.B28TapForMana(t, g, me.ID, land, "")
	if me.ManaPool.Total() != 2 {
		t.Errorf("with an imprinted card the land adds {C}{C}: pool %+v", me.ManaPool)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	effectstest.B16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(big) {
		t.Error("the exiled card should return to its owner's hand")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	me.ManaPool = game.ManaPool{}
	effectstest.B28TapForMana(t, g, me.ID, land, "")
	if me.ManaPool.Total() != 1 {
		t.Errorf("with nothing exiled the land adds {C}: pool %+v", me.ManaPool)
	}
}

func TestS82GapB58UginsLabyrinthImprintCanBeDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	big := pushHandCardMV(me, "Kozilek", "Creature — Eldrazi", "{8}", nil)
	land := importLandForTest(t, g, me, "Ugin's Labyrinth", s82g58UginsLabyrinth)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the imprint opens no pick")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(big) {
		t.Error("declining keeps the card")
	}
	effectstest.B28TapForMana(t, g, me.ID, land, "")
	if me.ManaPool.Total() != 1 {
		t.Errorf("pool %+v, want {C}", me.ManaPool)
	}
}

// pushHandCardMV puts a card with a printed mana cost into a hand.
func pushHandCardMV(p *game.Player, name, typeLine, cost string, colors []string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, Colors: colors,
		Owner: p.ID, Controller: p.ID})
	return id
}

// importLandForTest puts a catalog land onto the battlefield through the entry
// pipeline, so its enters-tapped replacement and its ETB trigger run.
func importLandForTest(t *testing.T, g *game.Game, p *game.Player, name, oracle string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Land", OracleID: oracle, Owner: p.ID, Controller: p.ID})
	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() { entered, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{}) })
	if err != nil || entered == uuid.Nil {
		t.Fatalf("put %s onto the battlefield: %v", name, err)
	}
	return entered
}

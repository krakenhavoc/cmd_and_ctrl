package effects

import (
	"testing"

	"github.com/google/uuid"

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
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- Orim's Chant --------------------------------------------------

func TestS82GapB58OrimsChantStopsTheTargetCastingAndKickedStopsAttacks(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := p1g2Seats(g)
		mine := p1Creature(g, me.ID, 2, 2)
		theirs := p1Creature(g, opp.ID, 2, 2)
		params := game.CastSpellParams{Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}
		b06AddMana(me, "W")
		if kicked {
			params.OptionalCosts = []int{0}
			b06AddMana(me, "W")
		}
		p1g2Cast(t, g, "Orim's Chant", "Instant", s82g58OrimsChant, "{W}", params)
		passPriorityAroundTable(t, g)

		fodder := pushCatalogHandCard(opp, "Fodder", "Instant", "")
		if err := g.CastSpell(opp.ID, fodder, game.CastSpellParams{Strict: true}); err == nil {
			t.Errorf("kicked=%v: the target player cast a spell", kicked)
		}
		late := p1Creature(g, me.ID, 1, 1) // enters after the spell resolved
		for name, id := range map[string]uuid.UUID{"mine": mine, "theirs": theirs, "late": late} {
			banned := restrictionsOf(t, g, id)&game.CantAttack != 0
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
			g := newCatalogGame(t)
			me := g.Seats[0]
			var lib []game.Card
			for _, n := range append(append([]string{}, tc.fetch...), tc.refuse...) {
				lib = append(lib, game.Card{Name: n, TypeLine: "Basic Land — " + n})
			}
			seedSearchLibrary(me, lib...)
			land := b43Catalog(g, me.ID, tc.name, "Land", tc.oracle, 0, 0)
			b28TapForMana(t, g, me.ID, land, "C")
			if got := poolColors(me); len(got) != 1 || got[0] != "C" {
				t.Fatalf("tapped for %v, want {C}", got)
			}
			g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
			advanceToMain(t, g)
			b16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
			passPriorityAroundTable(t, g)
			c := searchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("the fetch opens no search")
			}
			for _, n := range tc.refuse {
				if searchOptionNamed(g, c, n) != uuid.Nil {
					t.Errorf("a basic %s is not among %v", n, tc.fetch)
				}
			}
			var pick uuid.UUID
			for _, n := range tc.fetch {
				id := searchOptionNamed(g, c, n)
				if id == uuid.Nil {
					t.Fatalf("basic %s is not offered", n)
				}
				pick = id
			}
			answerSearchByID(t, g, me.ID, pick)
			passPriorityAroundTable(t, g)
			if g.Battlefield.Contains(land) {
				t.Error("the Landscape was not sacrificed")
			}
			if !g.Battlefield.Contains(pick) || !b16Tapped(t, g, pick) {
				t.Error("the basic should arrive tapped")
			}
		})
	}
}

func TestS82GapB58LandscapeCyclesForItsThreeColors(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Sorcery"})
	id := pushCatalogHandCard(me, "Twisted Landscape", "Land", s82g58TwistedLand)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{R}{G}"); err != nil {
		t.Fatal(err)
	}
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.ManaPool.Total() != 0 {
		t.Errorf("cycling should cost {B}{R}{G}: pool %+v", me.ManaPool)
	}
	if !me.Graveyard.Contains(id) || me.Hand.Size() != hand {
		t.Errorf("cycling should discard the card and draw one: graveyard=%v hand %d -> %d", me.Graveyard.Contains(id), hand, me.Hand.Size())
	}
}

// --- Flare of Cultivation ------------------------------------------

func TestS82GapB58FlareOfCultivationSacrificesAGreenCreatureAndFetchesTwoBasics(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	var forests []uuid.UUID
	seedSearchLibrary(me)
	for i := 0; i < 3; i++ {
		forests = append(forests, pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"}))
	}
	flare := handCardAtMain(t, g, "Flare of Cultivation", "Sorcery", s82g58FlareCultivate)
	green := seedColoredCreature(g, me.ID, "Elf", []string{"G"}, 0)
	token := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Saproling",
		TypeLine: "Token Creature — Saproling", Colors: []string{"G"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	white := seedColoredCreature(g, me.ID, "Cleric", []string{"W"}, 0)
	for name, id := range map[string]uuid.UUID{"a green token": token, "a white creature": white} {
		if err := castForAltCost(g, flare, "", "sacrifice", []uuid.UUID{id}, nil); err == nil {
			t.Fatalf("%s paid \"sacrifice a nontoken green creature\"", name)
		}
	}
	handBefore := me.Hand.Size()
	if err := castForAltCost(g, flare, "", "sacrifice", []uuid.UUID{green}, nil); err != nil {
		t.Fatalf("Flare for a nontoken green creature: %v", err)
	}
	if g.Battlefield.Contains(green) {
		t.Error("the green creature was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	answerOrDecline(t, g, me.ID, 0, forests)
	answerOrDecline(t, g, me.ID, 1, forests)
	if !g.Battlefield.Contains(forests[0]) || !isTapped(g, forests[0]) {
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
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	vault := importLandForTest(t, g, me, "Spymaster's Vault", s82g58SpymastersVlt)
	if !isTapped(g, vault) {
		t.Error("with no Swamp the Vault enters tapped")
	}
	seedLand(g, me.ID, "Swamp", "Basic Land — Swamp", "")
	vault2 := importLandForTest(t, g, me, "Spymaster's Vault", s82g58SpymastersVlt)
	if isTapped(g, vault2) {
		t.Error("with a Swamp the Vault enters untapped")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vault) })

	b28TapForMana(t, g, me.ID, vault, "B")
	if got := poolColors(me); len(got) != 1 || got[0] != "B" {
		t.Fatalf("tapped for %v, want {B}", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(vault) })

	// Two creatures die this turn, so the connive is X = 2.
	bear := p1Creature(g, me.ID, 2, 2)
	d1 := p1Creature(g, me.ID, 1, 1)
	d2 := p1Creature(g, g.Seats[1].ID, 1, 1)
	for _, id := range []uuid.UUID{d1, d2} {
		g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
	}
	g.RunStateChecksForTest()
	pushLibraryCardForTest(me, game.Card{Name: "Draw A", TypeLine: "Creature — A", Power: 1, Toughness: 1})
	pushLibraryCardForTest(me, game.Card{Name: "Draw B", TypeLine: "Sorcery"})
	hand := me.Hand.Size()
	b06AddMana(me, "B")
	b16Activate(t, g, me.ID, vault, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		t.Fatalf("connive 2 drew %d, want 2 (before discarding)", got)
	}
	if !b16Tapped(t, g, vault) {
		t.Error("the Vault should tap to connive")
	}
}

// --- Ugin's Labyrinth ----------------------------------------------

func TestS82GapB58UginsLabyrinthImprintsDoublesAndReturns(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	eldrazi := me.Hand
	_ = eldrazi
	big := pushHandCardMV(me, "Kozilek", "Creature — Eldrazi", "{8}", nil)
	colored := pushHandCardMV(me, "Big Red", "Creature — Dragon", "{5}{R}{R}", []string{"R"})
	small := pushHandCardMV(me, "Wurmcoil Lite", "Artifact Creature — Wurm", "{6}", nil)
	land := importLandForTest(t, g, me, "Ugin's Labyrinth", s82g58UginsLabyrinth)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the imprint opens no pick")
	}
	if pick.ChooseMin != 0 {
		t.Errorf("imprint is \"you may\": floor %d", pick.ChooseMin)
	}
	if !hasID(pick.ChooseCards, big) || hasID(pick.ChooseCards, colored) || hasID(pick.ChooseCards, small) {
		t.Errorf("only the colorless card of mana value 7+ is offered: %+v", pick.ChooseCards)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{big}); err != nil {
		t.Fatal(err)
	}
	if !inExile(g, big) {
		t.Fatal("Kozilek was not exiled")
	}
	if isTapped(g, land) {
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	}
	b28TapForMana(t, g, me.ID, land, "")
	if me.ManaPool.Total() != 2 {
		t.Errorf("with an imprinted card the land adds {C}{C}: pool %+v", me.ManaPool)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	b16Activate(t, g, me.ID, land, 0, game.ActivateAbilityParams{})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(big) {
		t.Error("the exiled card should return to its owner's hand")
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
	me.ManaPool = game.ManaPool{}
	b28TapForMana(t, g, me.ID, land, "")
	if me.ManaPool.Total() != 1 {
		t.Errorf("with nothing exiled the land adds {C}: pool %+v", me.ManaPool)
	}
}

func TestS82GapB58UginsLabyrinthImprintCanBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	big := pushHandCardMV(me, "Kozilek", "Creature — Eldrazi", "{8}", nil)
	land := importLandForTest(t, g, me, "Ugin's Labyrinth", s82g58UginsLabyrinth)
	passPriorityAroundTable(t, g)
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("the imprint opens no pick")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, nil); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(big) {
		t.Error("declining keeps the card")
	}
	b28TapForMana(t, g, me.ID, land, "")
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

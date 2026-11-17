package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 Dominaria United batch 6: the kicker creatures (#5487).

const (
	dmuBattlewingMysticOracle  = "49f0b317-6b06-482e-ab64-e02713e6c7e8"
	dmuBenalishSleeperOracle   = "14c3692d-16a9-4ef0-98cd-60d15e4aa31f"
	dmuBogBadgerOracle         = "6d8e464c-9df1-46b1-88e6-dcc7fe7b7caa"
	dmuRootweaverOracle        = "5914d4d5-51b8-49c2-9459-bf5579df4990"
	dmuMissionaryOracle        = "be04f54f-59fa-4f0a-b8b5-790d568932b6"
	dmuWarhorseOracle          = "05075332-6856-4147-a22d-3a5bd1e96b7b"
	dmuVineshaperProdigyOracle = "c41172c6-5133-48ed-bde0-3168b5371c18"
	dmuIconoclastOracle        = "9900c771-ab81-415e-90a7-123c1bacfa83"
	dmuShalaisAcolyteOracle    = "53ec394a-5aaa-4d83-82ec-f00e99902c0d"
	dmuBranchriderOracle       = "50fcb431-626c-4db5-aa2c-445f3b81e7fa"
	dmuVolsheTideturnerOracle  = "eab16fd1-735c-411f-a23a-0b93131c2eb2"
)

func TestDMUB06AllElevenAreFullAndKicked(t *testing.T) {
	for _, oracle := range []string{
		dmuBattlewingMysticOracle, dmuBenalishSleeperOracle, dmuBogBadgerOracle, dmuRootweaverOracle,
		dmuMissionaryOracle, dmuWarhorseOracle, dmuVineshaperProdigyOracle, dmuIconoclastOracle,
		dmuShalaisAcolyteOracle, dmuBranchriderOracle, dmuVolsheTideturnerOracle,
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// dmuCast casts a kicker creature from the active seat's hand with or without
// its kicker, settles the stack and returns the permanent's id.
func dmuCast(t *testing.T, g *game.Game, name, typeLine, oracle string, kicked bool) uuid.UUID {
	t.Helper()
	var opt []int
	if kicked {
		opt = []int{0}
	}
	id, err := castWithOptionalCosts(t, g, name, typeLine, oracle, nil, opt, nil)
	_ = id
	if err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

func TestBattlewingMysticKickedDiscardsHandThenDrawsTwo(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		base := me.Hand.Size()
		var old []uuid.UUID
		for i := 0; i < 3; i++ {
			old = append(old, handSpell(me, "Old", "Sorcery", "{9}"))
		}
		id := dmuCast(t, g, "Battlewing Mystic", "Creature — Bird Wizard", dmuBattlewingMysticOracle, kicked)
		passPriorityAroundTable(t, g)
		if findBattlefieldCardByID(g, id) == nil {
			t.Fatalf("kicked=%v: the Mystic is not on the battlefield", kicked)
		}
		if !kicked {
			if me.Hand.Size() != base+3 {
				t.Errorf("unkicked: hand = %d, want %d untouched", me.Hand.Size(), base+3)
			}
			continue
		}
		if me.Hand.Size() != 2 {
			t.Errorf("kicked: hand = %d, want the two drawn", me.Hand.Size())
		}
		for _, o := range old {
			if me.Hand.Contains(o) || !me.Graveyard.Contains(o) {
				t.Errorf("kicked: old card %s should be in the graveyard", o)
			}
		}
	}
}

func TestBenalishSleeperKickedEachPlayerSacrifices(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		var theirs []uuid.UUID
		for _, s := range g.Seats[1:] {
			theirs = append(theirs, b12Creature(g, s.ID, "Theirs", "Creature — Bear", 2, 2))
		}
		sleeper := dmuCast(t, g, "Benalish Sleeper", "Creature — Phyrexian Human Soldier", dmuBenalishSleeperOracle, kicked)
		passPriorityAroundTable(t, g)
		if kicked {
			answerSacrifice(t, g, me.ID, mine)
			passPriorityAroundTable(t, g)
		}
		for i, id := range theirs {
			if gone := findBattlefieldCardByID(g, id) == nil; gone != kicked {
				t.Errorf("kicked=%v: opponent %d creature gone = %v", kicked, i, gone)
			}
		}
		if gone := findBattlefieldCardByID(g, mine) == nil; gone != kicked {
			t.Errorf("kicked=%v: my creature gone = %v", kicked, gone)
		}
		if findBattlefieldCardByID(g, sleeper) == nil {
			t.Errorf("kicked=%v: the Sleeper should have survived (I sacrificed the other creature)", kicked)
		}
	}
}

func TestBogBadgerKickedGivesMenaceToCreaturesYouControlNow(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		advanceToMain(t, g)
		mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
		theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
		badger := dmuCast(t, g, "Bog Badger", "Creature — Badger", dmuBogBadgerOracle, kicked)
		passPriorityAroundTable(t, g)
		for _, tc := range []struct {
			name string
			id   uuid.UUID
			want bool
		}{{"mine", mine, kicked}, {"badger", badger, kicked}, {"theirs", theirs, false}} {
			c := findBattlefieldCardByID(g, tc.id)
			if c == nil {
				t.Fatalf("%s missing", tc.name)
			}
			if got := effectiveHasKeyword(c, "menace"); got != tc.want {
				t.Errorf("kicked=%v: %s menace = %v, want %v", kicked, tc.name, got, tc.want)
			}
		}
	}
}

func TestJuniperOrderRootweaverKickedPutsACounterOnTarget(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		ally := b12Creature(g, me.ID, "Ally", "Creature — Bear", 2, 2)
		dmuCast(t, g, "Juniper Order Rootweaver", "Creature — Human Druid", dmuRootweaverOracle, kicked)
		asked := settleAnsweringPicks(t, g, me.ID, me.ID)
		if want := map[bool]int{false: 0, true: 1}[kicked]; asked != want {
			t.Errorf("kicked=%v: %d target prompts, want %d", kicked, asked, want)
		}
		total := 0
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.Controller == me.ID {
					total += c.Counters[game.CounterPlusOne]
				}
			}
		})
		if want := map[bool]int{false: 0, true: 1}[kicked]; total != want {
			t.Errorf("kicked=%v: +1/+1 counters = %d, want %d", kicked, total, want)
		}
		_ = ally
	}
}

func TestPhyrexianMissionaryKickedReturnsACreatureCard(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		dead := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Dead Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		spell := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Spent Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
		dmuCast(t, g, "Phyrexian Missionary", "Creature — Phyrexian Human Cleric", dmuMissionaryOracle, kicked)
		settleAnsweringPicks(t, g, me.ID, me.ID)
		if got := me.Hand.Contains(dead); got != kicked {
			t.Errorf("kicked=%v: creature card in hand = %v", kicked, got)
		}
		if !me.Graveyard.Contains(spell) {
			t.Errorf("kicked=%v: the sorcery left the graveyard", kicked)
		}
	}
}

func TestPhyrexianWarhorseKickedMakesASoldierAndSacrificesForPump(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		horse := dmuCast(t, g, "Phyrexian Warhorse", "Creature — Phyrexian Horse", dmuWarhorseOracle, kicked)
		passPriorityAroundTable(t, g)
		soldiers := 0
		g.ReadSnapshot(func() {
			for _, c := range g.Battlefield.Cards {
				if c.Controller == me.ID && c.IsToken() && c.Name == "Soldier" {
					soldiers++
				}
			}
		})
		if want := map[bool]int{false: 0, true: 1}[kicked]; soldiers != want {
			t.Errorf("kicked=%v: %d Soldier tokens, want %d", kicked, soldiers, want)
		}
		if !kicked {
			continue
		}
		// {1}, sacrifice another creature: +2/+1 until end of turn.
		floatMana(t, g, me, "{R}")
		fodder := b12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
		if err := g.ActivateCatalogAbility(me.ID, horse, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		passPriorityAroundTable(t, g)
		if p, tough := effectivePTOf(t, g, horse); p != 2 || tough != 1 {
			t.Errorf("Warhorse is %d/%d, want +2/+1 on a 0/0 test body", p, tough)
		}
		if findBattlefieldCardByID(g, fodder) != nil {
			t.Error("the sacrificed creature is still on the battlefield")
		}
	}
}

func TestVineshaperProdigyKickedTakesOneOfTopThree(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		advanceToMain(t, g)
		size := me.Library.Size()
		top := make([]uuid.UUID, 3)
		for i := range top {
			top[i] = me.Library.Cards[i].InstanceID
		}
		dmuCast(t, g, "Vineshaper Prodigy", "Creature — Elf Druid", dmuVineshaperProdigyOracle, kicked)
		passPriorityAroundTable(t, g)
		if !kicked {
			if me.Library.Size() != size {
				t.Errorf("unkicked: library %d -> %d", size, me.Library.Size())
			}
			continue
		}
		pick := latestChooseCards(g, me.ID)
		if pick == nil {
			t.Fatal("no choose_cards prompt for the look")
		}
		if len(pick.ChooseCards) != 3 {
			t.Fatalf("prompt offers %d cards, want the top three", len(pick.ChooseCards))
		}
		top = append([]uuid.UUID(nil), pick.ChooseCards...)
		if err := g.ResolveChooseCards(pick.ID, me.ID, top[1:2]); err != nil {
			t.Fatalf("ResolveChooseCards: %v", err)
		}
		passPriorityAroundTable(t, g)
		if !me.Hand.Contains(top[1]) {
			t.Error("the chosen card is not in hand")
		}
		if me.Library.Size() != size-1 {
			t.Errorf("library %d -> %d, want one fewer", size, me.Library.Size())
		}
		n := me.Library.Size()
		bottom := map[uuid.UUID]bool{me.Library.Cards[n-1].InstanceID: true, me.Library.Cards[n-2].InstanceID: true}
		if !bottom[top[0]] || !bottom[top[2]] {
			t.Error("the other two cards are not on the bottom of the library")
		}
	}
}

func TestYavimayaIconoclastKickedGetsPumpAndHasteThisTurn(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := newCatalogGame(t)
		advanceToMain(t, g)
		id := dmuCast(t, g, "Yavimaya Iconoclast", "Creature — Elf", dmuIconoclastOracle, kicked)
		passPriorityAroundTable(t, g)
		c := findBattlefieldCardByID(g, id)
		if c == nil {
			t.Fatal("the Iconoclast is not on the battlefield")
		}
		if got := effectiveHasKeyword(c, "haste"); got != kicked {
			t.Errorf("kicked=%v: haste = %v", kicked, got)
		}
		// The hand card carries no printed P/T here, so the pump is the whole of it.
		want := map[bool]int{false: 0, true: 1}[kicked]
		if p, tough := effectivePTOf(t, g, id); p != want || tough != want {
			t.Errorf("kicked=%v: %d/%d, want %d/%d", kicked, p, tough, want, want)
		}
	}
}

func TestShalaisAcolyteAndViashinoBranchriderEnterWithCountersOnlyWhenKicked(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle string
		p, tough               int
	}{
		{"Shalai's Acolyte", "Creature — Angel", dmuShalaisAcolyteOracle, 0, 0},
		{"Viashino Branchrider", "Creature — Lizard Warrior", dmuBranchriderOracle, 0, 0},
	} {
		for _, kicked := range []bool{false, true} {
			g := newCatalogGame(t)
			advanceToMain(t, g)
			id := dmuCast(t, g, tc.name, tc.typeLine, tc.oracle, kicked)
			passPriorityAroundTable(t, g)
			want := 0
			if kicked {
				want = 2
			}
			if got := counterCount(g, id, game.CounterPlusOne); got != want {
				t.Errorf("%s kicked=%v: %d counters, want %d", tc.name, kicked, got, want)
			}
			if p, tough := ptOf(t, g, id); p != want || tough != want {
				t.Errorf("%s kicked=%v: %d/%d", tc.name, kicked, p, tough)
			}
		}
	}
}

func TestViashinoBranchriderPumpsItself(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := pushCatalogPermanent(g, me.ID, "Viashino Branchrider", "Creature — Lizard Warrior", dmuBranchriderOracle, false)
	floatMana(t, g, me, "{R}{R}{R}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, id); p != 3 {
		t.Errorf("power = %d, want 3", p)
	}
}

// Volshe Tideturner: the mana pays an instant, a sorcery or a kicked spell,
// and nothing else.
func TestVolsheTideturnerManaPaysInstantSorceryAndKickedSpells(t *testing.T) {
	const serpentOracle = "c39c1604-3bae-454d-9985-85101e51ec6e" // Kicker {2}
	for _, tc := range []struct {
		name, typeLine, oracle string
		cost                   string
		volshes                int
		optional               []int
		ok                     bool
	}{
		{"instant", "Instant", "", "{U}", 1, nil, true},
		{"sorcery", "Sorcery", "", "{U}", 1, nil, true},
		{"unkicked creature", "Creature — Serpent", serpentOracle, "{U}", 1, nil, false},
		{"kicked creature", "Creature — Serpent", serpentOracle, "{U}", 3, []int{0}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g, me, _ := spendTable(t)
			for i := 0; i < tc.volshes; i++ {
				v := pushCatalogPermanent(g, me.ID, "Volshe Tideturner", "Creature — Merfolk Wizard", dmuVolsheTideturnerOracle, false)
				if err := g.ActivateManaAbility(me.ID, v, 0, game.ManaAbilityParams{}); err != nil {
					t.Fatalf("tap for {U}: %v", err)
				}
			}
			id := uuid.New()
			me.Hand.PushTop(game.Card{
				InstanceID: id, Name: "Test " + tc.name, TypeLine: tc.typeLine, ManaCost: tc.cost,
				OracleID: tc.oracle, Owner: me.ID, Controller: me.ID,
			})
			err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: tc.optional})
			if tc.ok && err != nil {
				t.Errorf("the mana could not pay for a %s: %v", tc.name, err)
			}
			if !tc.ok {
				refusedForMana(t, err, "the mana for a "+tc.name)
			}
		})
	}
}

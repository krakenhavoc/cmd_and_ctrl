package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dark Ascension batch 441: fateful hour and the flashback sorceries.

const (
	oldB441GavonyIronwright    = "99a4f104-c358-442a-8273-6fe94286edb5"
	oldB441VillageSurvivors    = "c26dd36f-a81a-4d97-9be2-d350e93da79e"
	oldB441ShatteredPerception = "9853d789-b048-4cba-a64d-db2694ca9173"
	oldB441IncreasingConfusion = "11293f96-3263-45e5-89fc-850df6b25a63"
	oldB441IncreasingSavagery  = "616c7f45-5295-4b8c-a828-ff3a5ba4d917"
	oldB441IncreasingAmbition  = "09027657-bc9a-4702-88c1-e5b3534ad188"
	oldB441TrackersInstincts   = "1bc7c38b-9571-4395-be90-922f92816987"
	oldB441BreakOfDay          = "59674589-b174-4150-b8d6-30d25c5ead4c"
	oldB441FaithsShield        = "dd3cc9e5-b99a-429c-baeb-663a7af70081"
	oldB441ClingingMists       = "9be48462-2ec0-4d28-9977-f0232e8a1333"
)

func oldB441SetLife(g *game.Game, seat *game.Player, life int) {
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, seat.ID, life-seat.Life) })
}

func oldB441Push(g *game.Game, owner uuid.UUID, name, oracle string, p, t int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Human", OracleID: oracle,
		Power: p, Toughness: t, Owner: owner, Controller: owner,
	})
}

func TestOldB441GavonyIronwrightIsOnlyFatefulHour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	smith := oldB441Push(g, me.ID, "Gavony Ironwright", oldB441GavonyIronwright, 1, 4)
	bear := effectstest.B04Creature(g, me.ID, "Bear", 2, 2)
	foe := effectstest.B04Creature(g, g.Seats[1].ID, "Foe", 2, 2)

	for _, tc := range []struct {
		life, bearP, bearT int
	}{{40, 2, 2}, {6, 2, 2}, {5, 3, 6}, {1, 3, 6}, {7, 2, 2}} {
		oldB441SetLife(g, me, tc.life)
		if p, tt := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != tc.bearP || tt != tc.bearT {
			t.Errorf("life %d: other creature is %d/%d, want %d/%d", tc.life, p, tt, tc.bearP, tc.bearT)
		}
		if p := effectstest.EffectivePower(t, g, smith); p != 1 {
			t.Errorf("life %d: the Ironwright itself is %d power, want 1 (\"other\")", tc.life, p)
		}
		if p := effectstest.EffectivePower(t, g, foe); p != 2 {
			t.Errorf("life %d: an opponent's creature is %d power, want 2", tc.life, p)
		}
	}
}

func TestOldB441VillageSurvivorsGrantsVigilanceOnlyAtFiveOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	survivors := oldB441Push(g, me.ID, "Village Survivors", oldB441VillageSurvivors, 4, 5)
	bear := effectstest.B04Creature(g, me.ID, "Bear", 2, 2)

	oldB441SetLife(g, me, 6)
	if effectstest.EffectiveAbilitiesContain(t, g, bear, "vigilance") {
		t.Error("vigilance at 6 life")
	}
	oldB441SetLife(g, me, 5)
	if !effectstest.EffectiveAbilitiesContain(t, g, bear, "vigilance") {
		t.Error("no vigilance at 5 life")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, survivors, "vigilance") != true {
		t.Error("the printed vigilance is unconditional")
	}
}

func TestOldB441ShatteredPerceptionWheelsAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Shattered Perception", "Sorcery", oldB441ShatteredPerception, nil)
	handBefore, libBefore := me.Hand.Size(), me.Library.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != handBefore || me.Library.Size() != libBefore-handBefore {
		t.Fatalf("hand %d→%d library %d→%d: want the same hand size and %d drawn", handBefore, me.Hand.Size(), libBefore, me.Library.Size(), handBefore)
	}

	for i := 0; i < 10; i++ { // the second wheel must not draw from an empty library
		effectstest.StapleLibraryCard(me, "Filler", "Sorcery")
	}
	id := effectstest.SeedGraveyardCard(t, g, "Shattered Perception", "Sorcery", oldB441ShatteredPerception)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Errorf("flashed-back card should be exiled ")
	}
}

func TestOldB441IncreasingConfusionMillsXThenTwiceXFromGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, foe := g.Seats[0], g.Seats[1]
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: foe.ID}}
	before := foe.Library.Size()
	effectstest.Pr8CastX(t, g, "Increasing Confusion", "Sorcery", oldB441IncreasingConfusion, 3, target)
	if got := before - foe.Library.Size(); got != 3 {
		t.Fatalf("hand cast milled %d, want 3", got)
	}

	id := effectstest.SeedGraveyardCard(t, g, "Increasing Confusion", "Sorcery", oldB441IncreasingConfusion)
	before = foe.Library.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", XValue: 4, Targets: target}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := before - foe.Library.Size(); got != 8 {
		t.Fatalf("flashback milled %d, want 8 (twice X=4)", got)
	}
	if !g.Exile.Contains(id) {
		t.Error("flashed-back card should be exiled")
	}
}

func TestOldB441IncreasingSavageryFiveThenTenCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B04Creature(g, me.ID, "Bear", 2, 2)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: bear}}
	effectstest.CastCatalogSpell(t, g, "Increasing Savagery", "Sorcery", oldB441IncreasingSavagery, target)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindBattlefieldCardForTest(g, bear).Counters[game.CounterPlusOne]; got != 5 {
		t.Fatalf("counters = %d, want 5", got)
	}
	id := effectstest.SeedGraveyardCard(t, g, "Increasing Savagery", "Sorcery", oldB441IncreasingSavagery)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Targets: target}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.FindBattlefieldCardForTest(g, bear).Counters[game.CounterPlusOne]; got != 15 {
		t.Fatalf("counters = %d, want 15 (5 + 10)", got)
	}
}

func TestOldB441IncreasingAmbitionSearchesOneThenTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.StapleLibraryCard(me, "Needle A", "Sorcery")
	b := effectstest.StapleLibraryCard(me, "Needle B", "Sorcery")
	effectstest.CastCatalogSpell(t, g, "Increasing Ambition", "Sorcery", oldB441IncreasingAmbition, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, a)
	if !me.Hand.Contains(a) || me.Hand.Contains(b) {
		t.Fatal("a hand cast takes exactly one card")
	}

	id := effectstest.SeedGraveyardCard(t, g, "Increasing Ambition", "Sorcery", oldB441IncreasingAmbition)
	c := effectstest.StapleLibraryCard(me, "Needle C", "Sorcery")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, b, c)
	if !me.Hand.Contains(b) || !me.Hand.Contains(c) {
		t.Fatal("a graveyard cast takes two cards")
	}
}

func TestOldB441TrackersInstinctsTakesACreatureAndBinsTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	lib := effectstest.TrekB48Library(me,
		game.Card{Name: "Spell A", TypeLine: "Sorcery"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Fifth", TypeLine: "Instant"},
	)
	effectstest.CastCatalogSpell(t, g, "Tracker's Instincts", "Sorcery", oldB441TrackersInstincts, nil)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.LatestChooseCards(g, me.ID)
	if ch == nil || ch.ChooseMin != 1 || ch.ChooseMax != 1 {
		t.Fatalf("want a required pick of one creature card: %+v", ch)
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[0]}); err == nil {
		t.Error("a sorcery was accepted")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[2]}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(lib[2]) {
		t.Error("the chosen creature is not in hand")
	}
	for _, i := range []int{0, 1, 3} {
		if !me.Graveyard.Contains(lib[i]) {
			t.Errorf("revealed card %d is not in the graveyard", i)
		}
	}
	if !me.Library.Contains(lib[4]) {
		t.Error("the fifth card left the library")
	}
}

func TestOldB441BreakOfDayPumpsAndGivesIndestructibleAtFiveOrLess(t *testing.T) {
	for _, life := range []int{6, 5} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		bear := effectstest.B04Creature(g, me.ID, "Bear", 2, 2)
		foe := effectstest.B04Creature(g, g.Seats[1].ID, "Foe", 2, 2)
		oldB441SetLife(g, me, life)
		effectstest.CastCatalogSpell(t, g, "Break of Day", "Instant", oldB441BreakOfDay, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if p := effectstest.EffectivePower(t, g, bear); p != 3 {
			t.Errorf("life %d: bear power %d, want 3", life, p)
		}
		if p := effectstest.EffectivePower(t, g, foe); p != 2 {
			t.Errorf("life %d: an opponent's creature got the pump", life)
		}
		if got := effectstest.EffectiveAbilitiesContain(t, g, bear, "indestructible"); got != (life <= 5) {
			t.Errorf("life %d: indestructible = %v", life, got)
		}
	}
}

func TestOldB441FaithsShieldTargetOnlyThenWholeBoardAtFiveOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.B04Creature(g, me.ID, "A", 2, 2)
	b := effectstest.B04Creature(g, me.ID, "B", 2, 2)
	target := []game.TargetRef{{Kind: game.TargetCard, ID: a}}
	effectstest.CastCatalogSpell(t, g, "Faith's Shield", "Instant", oldB441FaithsShield, target)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "R")
	if !effectstest.EffectiveAbilitiesContain(t, g, a, "protection from red") {
		t.Error("the target should have protection from red")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, b, "protection from red") {
		t.Error("only the target gains it above 5 life")
	}

	oldB441SetLife(g, me, 5)
	effectstest.CastCatalogSpell(t, g, "Faith's Shield", "Instant", oldB441FaithsShield, target)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerColor(t, g, me.ID, "U")
	if !effectstest.EffectiveAbilitiesContain(t, g, b, "protection from blue") || !effectstest.EffectiveAbilitiesContain(t, g, a, "protection from blue") {
		t.Error("fateful hour: every permanent you control gains protection from blue")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, b, "protection from red") {
		t.Error("the first Shield only protected its target")
	}
}

func TestOldB441ClingingMistsTapsAndFreezesAttackersAtFiveOrLess(t *testing.T) {
	for _, life := range []int{6, 5} {
		g := effectstest.NewCatalogGame(t)
		me, foe := g.Seats[0], g.Seats[1]
		atk := effectstest.B04Creature(g, foe.ID, "Attacker", 2, 2)
		g.WithWriteLock(func() {
			for i := range g.Battlefield.Cards {
				if g.Battlefield.Cards[i].InstanceID == atk {
					g.Battlefield.Cards[i].AttackingTarget = me.ID
				}
			}
		})
		oldB441SetLife(g, me, life)
		effectstest.CastCatalogSpell(t, g, "Clinging Mists", "Instant", oldB441ClingingMists, nil)
		effectstest.PassPriorityAroundTable(t, g)
		tapped := effectstest.FindBattlefieldCardForTest(g, atk).Tapped
		skips := len(effectstest.HmlUntapSkips(t, g, atk))
		if tapped != (life <= 5) || (skips == 1) != (life <= 5) {
			t.Errorf("life %d: tapped=%v untap skips=%d", life, tapped, skips)
		}
	}
}

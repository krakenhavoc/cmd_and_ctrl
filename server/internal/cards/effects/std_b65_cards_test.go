package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// std_b65_cards_test.go — Lorwyn Eclipsed leftovers, batch 65 (#4458).

const (
	stdB65SowerOracle        = "4121b7d9-5fbe-40e2-a2a6-d7633f423b5f"
	stdB65IntimidationOracle = "da099c34-74b2-4db8-8583-09bf2c56606f"
	stdB65WretchOracle       = "0d3b6e27-2f8b-42e0-9490-7b134e84bb6b"
	stdB65ScoutOracle        = "f6758e8f-5462-48a7-96d3-60d07b1563fc"
	stdB65MaliceOracle       = "66d57851-2844-4b9d-be78-e90fc620b750"
	stdB65ShimmercreepOracle = "52244c23-7680-4bed-8ef8-8d10e289d3f7"
	stdB65UnburyOracle       = "e2928d3c-b64a-443e-a1c1-fd204bb098ba"
)

func TestStdB65EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		stdB65SowerOracle:        "Nightmare Sower",
		stdB65IntimidationOracle: "Perfect Intimidation",
		stdB65WretchOracle:       "Retched Wretch",
		stdB65ScoutOracle:        "Scarblade Scout",
		stdB65MaliceOracle:       "Scarblade's Malice",
		stdB65ShimmercreepOracle: "Shimmercreep",
		stdB65UnburyOracle:       "Unbury",
	}
	for oracle, name := range want {
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

func TestStdB65NightmareSowerPutsACounterOnlyOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Nightmare Sower", TypeLine: "Creature — Faerie Assassin",
		OracleID: stdB65SowerOracle, Power: 2, Toughness: 3, Owner: opp.ID, Controller: opp.ID,
	})
	victim := b12Creature(g, me.ID, "Victim", "Creature — Bear", 3, 3)
	batch01OpponentCasts(t, g, opp, "Bolt", lightningBoltOracle, "{R}",
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}})
	pick := latestPickTarget(g, opp.ID)
	if pick == nil {
		t.Fatal("no target prompt for the Sower")
	}
	pickCard(t, g, opp.ID, victim)
	passPriorityAroundTable(t, g)
	if got := minusCounters(g, victim); got != 1 {
		t.Errorf("%d -1/-1 counters, want 1", got)
	}
}

func TestStdB65NightmareSowerIgnoresYourOwnTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Nightmare Sower", TypeLine: "Creature — Faerie Assassin",
		OracleID: stdB65SowerOracle, Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID,
	})
	b12Creature(g, g.Seats[1].ID, "Victim", "Creature — Bear", 3, 3)
	castCatalogSpell(t, g, "Bolt", "Instant", lightningBoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}})
	if latestPickTarget(g, me.ID) != nil || triggerOnStackAny(g) {
		t.Error("the Sower triggered on a spell cast during its controller's own turn")
	}
}

func triggerOnStackAny(g *game.Game) bool {
	for _, it := range g.StackMeta {
		if it != nil && it.Kind == game.StackItemTriggered {
			return true
		}
	}
	return false
}

func TestStdB65PerfectIntimidationBothModes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := b12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(big, game.CounterPlusOne, 2)
		_ = g.AddCounterForEffect(big, "shield", 1)
	})
	keep := handCardFull(opp, "Keep", "Instant", "", "", nil)
	a := handCardFull(opp, "A", "Instant", "", "", nil)
	b := handCardFull(opp, "B", "Instant", "", "", nil)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Perfect Intimidation", TypeLine: "Sorcery", OracleID: stdB65IntimidationOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Modes: []int{0, 1},
		Targets: []game.TargetRef{
			{Kind: game.TargetPlayer, ID: opp.ID, Mode: 0},
			{Kind: game.TargetCard, ID: big, Mode: 1},
		},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, opp.ID)
	if c == nil || c.ChooseMin != 2 || c.ChooseMax != 2 {
		t.Fatalf("the opponent's prompt is %+v, want exactly two cards", c)
	}
	if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{a, b}); err != nil {
		t.Fatalf("answer: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) || !opp.Hand.Contains(keep) {
		t.Error("the two chosen cards were not the ones exiled")
	}
	cc := findBattlefieldCardByID(g, big)
	if cc == nil || len(cc.Counters) != 0 && (cc.Counters[game.CounterPlusOne] != 0 || cc.Counters["shield"] != 0) {
		t.Errorf("counters left on the creature: %v", cc)
	}
}

func TestStdB65PerfectIntimidationExilesAWholeSmallHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	only := handCardFull(opp, "Only", "Instant", "", "", nil)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Perfect Intimidation", TypeLine: "Sorcery", OracleID: stdB65IntimidationOracle, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		Modes:   []int{0},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID, Mode: 0}},
	}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	passPriorityAroundTable(t, g)
	c := latestChooseCardsFor(g, opp.ID)
	if c != nil && c.ChooseMax == 1 {
		if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{only}); err != nil {
			t.Fatalf("answer: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if !g.Exile.Contains(only) {
		t.Error("a one-card hand was not exiled")
	}
}

func TestStdB65RetchedWretchReturnsWithNoAbilitiesOnlyIfItHadACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	wretch := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Retched Wretch", TypeLine: "Creature — Goblin",
		OracleID: stdB65WretchOracle, Power: 4, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(wretch, game.CounterMinusOne, 1) })
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wretch) })
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wretch) {
		t.Fatal("the Wretch did not return")
	}
	// It does not loop: it dies again, with no counter and no abilities.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wretch) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wretch) {
		t.Error("the returned Wretch came back a second time")
	}
}

func TestStdB65RetchedWretchStaysDeadWithoutACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	wretch := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Retched Wretch", TypeLine: "Creature — Goblin",
		OracleID: stdB65WretchOracle, Power: 4, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(wretch) })
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(wretch) || !me.Graveyard.Contains(wretch) {
		t.Error("the Wretch returned without a -1/-1 counter")
	}
}

func TestStdB65ScarbladeScoutMillsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	seedLibrary(me, "A", "B", "C")
	gy := me.Graveyard.Size()
	stdB63CastCreature(t, g, "Scarblade Scout", "Creature — Elf Scout", "{1}{B}", stdB65ScoutOracle, []string{"B"}, 2, 2)
	if got := me.Graveyard.Size() - gy; got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
}

func TestStdB65ScarbladesMaliceMakesAnElfWhenThatCreatureDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := b12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Scarblade's Malice", "Instant", stdB65MaliceOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "deathtouch") || !hasEffectiveKeyword(t, g, bear, "lifelink") {
		t.Error("the target did not gain deathtouch and lifelink")
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(other) })
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Elf"); n != 0 {
		t.Fatalf("an Elf for a different creature's death: %d", n)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(bear) })
	passPriorityAroundTable(t, g)
	if n, _ := countTokensNamed(g, me.ID, "Elf"); n != 1 {
		t.Errorf("%d Elf tokens, want 1", n)
	}
}

func TestStdB65ShimmercreepDrainsByTheColorsAmongYourPermanents(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	third := g.Seats[(g.Turn.ActiveSeat+2)%4]
	b16Creature(g, me.ID, "Red Guy", "Creature — Goblin", 1, 1, "R")
	b16Creature(g, me.ID, "Green Guy", "Creature — Elf", 1, 1, "G")
	life, oppLife, thirdLife := me.Life, opp.Life, third.Life
	stdB63CastCreature(t, g, "Shimmercreep", "Creature — Elemental", "{4}{B}", stdB65ShimmercreepOracle, []string{"B"}, 3, 5)
	// Black (itself), red and green.
	if me.Life != life+3 {
		t.Errorf("gained to %d, want %d (once, not once per opponent)", me.Life, life+3)
	}
	if opp.Life != oppLife-3 || third.Life != thirdLife-3 {
		t.Errorf("opponents at %d and %d, want %d each", opp.Life, third.Life, oppLife-3)
	}
}

func TestStdB65UnburyReturnsOneCreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elf := batch01GraveyardCard(me, "Elf", "Creature — Elf")
	spell := batch01GraveyardCard(me, "Spell", "Instant")
	castModal(t, g, "Unbury", "Instant", stdB65UnburyOracle, []int{0},
		[]game.TargetRef{{Kind: game.TargetCard, ID: elf}})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(elf) {
		t.Error("the creature card did not return to hand")
	}
	if err := stdB65CastUnburyErr(t, g, []int{0}, []game.TargetRef{{Kind: game.TargetCard, ID: spell}}); err == nil {
		t.Error("a noncreature card was accepted")
	}
}

func TestStdB65UnburyTwoCardsMustShareACreatureType(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	elfA := batch01GraveyardCard(me, "Elf Warrior", "Creature — Elf Warrior")
	elfB := batch01GraveyardCard(me, "Elf Druid", "Creature — Elf Druid")
	goblin := batch01GraveyardCard(me, "Goblin", "Creature — Goblin")
	if err := stdB65CastUnburyErr(t, g, []int{1}, []game.TargetRef{
		{Kind: game.TargetCard, ID: elfA, Mode: 0, Slot: 0},
		{Kind: game.TargetCard, ID: goblin, Mode: 0, Slot: 1},
	}); err == nil {
		t.Fatal("two cards sharing no creature type were accepted")
	}
	castModal(t, g, "Unbury", "Instant", stdB65UnburyOracle, []int{1}, []game.TargetRef{
		{Kind: game.TargetCard, ID: elfA, Mode: 0, Slot: 0},
		{Kind: game.TargetCard, ID: elfB, Mode: 0, Slot: 1},
	})
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(elfA) || !me.Hand.Contains(elfB) || me.Hand.Contains(goblin) {
		t.Error("the two Elf cards were not both returned")
	}
}

// stdB65CastUnburyErr casts Unbury and returns the announce error.
func stdB65CastUnburyErr(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) error {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{InstanceID: id, Name: "Unbury", TypeLine: "Instant", OracleID: stdB65UnburyOracle,
		Owner: active.ID, Controller: active.ID})
	return g.CastSpell(active.ID, id, game.CastSpellParams{Modes: modes, Targets: targets, Strict: true})
}

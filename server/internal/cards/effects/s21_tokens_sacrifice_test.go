package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s21_tokens_sacrifice_test.go — the rest of S21's catalog tail:
// token producers, sacrifice outlets, and the two aristocrats
// commanders. Plus the CreateTokenAdvanced entry options, which are
// engine-level and tested directly.

const (
	dragonFodderOracle      = "d0d2c45b-b6e3-4999-bdab-976e8f0d6617"
	hordelingOutburstOracle = "a6450b8e-eb18-431c-9eb7-7daf107978b2"
	graveTitanOracle        = "f3abd4d1-a975-4e85-8684-aa0fce029670"
	pawnOfUlamogOracle      = "9bcaf141-1f1f-491f-aced-13dc093b9e2c"
	sternLessonOracle       = "8315aa34-08b8-403e-a2e6-796a2d6978ad"
	bloodflowOracle         = "ddb2fb87-235a-4365-aa25-40c197425e43"
	korvoldOracle           = "9ae669dd-7e60-4649-b96e-35da28be641a"
	mazirekOracle           = "e0420f2c-d578-421e-ae75-e7dc5f70661a"
)

// --- token producers ---------------------------------------------

func TestGoblinSorceriesMakeTheirTokens(t *testing.T) {
	for _, tc := range []struct {
		name   string
		oracle string
		want   int
	}{
		{"Dragon Fodder", dragonFodderOracle, 2},
		{"Hordeling Outburst", hordelingOutburstOracle, 3},
	} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		castCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, nil)
		passPriorityAroundTable(t, g)
		if got := countBattlefieldNamed(g, me.ID, "Goblin"); got != tc.want {
			t.Errorf("%s made %d Goblins, want %d", tc.name, got, tc.want)
		}
	}
}

// Grave Titan is the catalog's proof that "enters OR attacks" is one
// ability with two conditions: the same declaration fires on both,
// and the attack half runs on EventAttack — which issue #73 recorded
// as missing and which has existed since S22.
func TestGraveTitanMakesZombiesOnEntryAndOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	castCatalogSpell(t, g, "Grave Titan", "Creature — Giant", graveTitanOracle, nil)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Zombie"); got != 2 {
		t.Fatalf("%d Zombies after the ETB, want 2", got)
	}

	// The cast Titan is summoning sick, so attack with one that is
	// already settled.
	titan := pushCatalogPermanent(g, me.ID, "Grave Titan", "Creature — Giant", graveTitanOracle, false)
	declareAttack(t, g, opp.ID, titan)
	if n := triggersOnStackFrom(g, titan); n != 1 {
		t.Fatalf("%d attack triggers, want 1", n)
	}
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Zombie"); got != 4 {
		t.Errorf("%d Zombies after the attack, want 4", got)
	}
}

func TestPawnOfUlamogMakesSpawnForNontokenDeathsOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Pawn of Ulamog", "Creature — Vampire Shaman", pawnOfUlamogOracle, false)

	// A token dying gives nothing — that clause is what stops the
	// Spawns looping into more Spawns.
	token := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: token, Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Eldrazi Spawn"); got != 0 {
		t.Errorf("a token death made %d Spawn, want 0", got)
	}

	// A nontoken one does, once the "you may" is answered.
	fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Human", "", false)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(fodder) })
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if got := countBattlefieldNamed(g, me.ID, "Eldrazi Spawn"); got != 1 {
		t.Errorf("a nontoken death made %d Spawn, want 1", got)
	}
}

func TestSternLessonCreatesATappedPowerstone(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Stern Lesson", "Instant", sternLessonOracle, nil)
	passPriorityAroundTable(t, g)

	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Powerstone" || c.Controller != me.ID {
			continue
		}
		found = true
		if !c.Tapped {
			t.Errorf("the Powerstone should enter TAPPED")
		}
	}
	if !found {
		t.Fatalf("no Powerstone token created")
	}
}

// The entry options are engine-level, so they are tested against the
// engine helper rather than through a card: a card that wanted all
// three at once would be a card nobody has printed.
func TestCreateTokensForEffectAppliesEntryOptions(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	var ids []uuid.UUID
	g.WithWriteLock(func() {
		var err error
		ids, err = g.CreateTokensForEffect(me.ID, RedGoblinToken(), 2, game.TokenEntryOptions{
			Tapped:   true,
			Counters: map[string]int{game.CounterPlusOne: 1},
			Keywords: []string{"haste"},
		})
		if err != nil {
			t.Errorf("CreateTokensForEffect: %v", err)
		}
	})
	if len(ids) != 2 {
		t.Fatalf("created %d tokens, want 2", len(ids))
	}
	for _, id := range ids {
		c, ok := battlefieldCard(g, id)
		if !ok {
			t.Fatalf("token %s not on the battlefield", id)
		}
		if !c.Tapped {
			t.Errorf("token should be tapped")
		}
		if c.Counters[game.CounterPlusOne] != 1 {
			t.Errorf("token counters = %v, want one +1/+1", c.Counters)
		}
		if len(c.Keywords) != 1 || c.Keywords[0] != "haste" {
			t.Errorf("token keywords = %v, want the granted haste on top of the template's none", c.Keywords)
		}
	}
	// The grant must not leak back into the shared template.
	if kw := RedGoblinToken().Keywords; len(kw) != 0 {
		t.Errorf("the template grew keywords: %v", kw)
	}
}

// --- sacrifice outlets -------------------------------------------

func TestBloodflowConnoisseurGrowsOnEachSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vamp := pushCatalogPermanent(g, me.ID, "Bloodflow Connoisseur", "Creature — Vampire", bloodflowOracle, false)
	fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)

	if err := g.ActivateCatalogAbility(me.ID, vamp, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Errorf("the sacrifice is a cost, paid at announce")
	}
	passPriorityAroundTable(t, g)
	if got := counterCount(g, vamp, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

// --- the commanders ----------------------------------------------

// Korvold's two abilities feed each other: the first one's mandatory
// sacrifice is what triggers the second.
func TestKorvoldSacrificesOnEntryThenGrowsAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	fodder := pushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	handBefore := len(me.Hand.Cards)

	korvold := castCatalogSpell(t, g, "Korvold, Fae-Cursed King",
		"Legendary Creature — Dragon Noble", korvoldOracle, nil)
	passPriorityAroundTable(t, g)

	c := sacrificeChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("Korvold's entry should demand a sacrifice")
	}
	for _, opt := range c.SacrificeOptions {
		if opt == korvold {
			t.Errorf("'another permanent' must not offer Korvold himself")
		}
	}
	answerSacrifice(t, g, me.ID, fodder)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(fodder) {
		t.Errorf("the sacrifice should have happened")
	}
	if got := counterCount(g, korvold, game.CounterPlusOne); got != 1 {
		t.Errorf("Korvold has %d +1/+1 counters, want 1", got)
	}
	// Cast Korvold (net zero out of hand) and drew one off his own
	// sacrifice.
	if got := len(me.Hand.Cards); got != handBefore+1 {
		t.Errorf("hand = %d, want %d", got, handBefore+1)
	}
}

// With nothing else on the board the mandatory sacrifice finds
// nothing, which is a no-op rather than an error or a self-sacrifice.
func TestKorvoldWithNoOtherPermanentSacrificesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	korvold := castCatalogSpell(t, g, "Korvold, Fae-Cursed King",
		"Legendary Creature — Dragon Noble", korvoldOracle, nil)
	passPriorityAroundTable(t, g)

	if c := sacrificeChoiceFor(g, me.ID); c != nil {
		t.Errorf("no legal permanent should mean no prompt, got %+v", c.SacrificeOptions)
	}
	if !g.Battlefield.Contains(korvold) {
		t.Errorf("Korvold must not eat himself")
	}
}

func TestMazirekGrowsEveryCreatureOnAnyPlayersSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mazirek := pushCatalogPermanent(g, me.ID, "Mazirek, Kraul Death Priest",
		"Legendary Creature — Insect Shaman", mazirekOracle, false)
	mine := pushCatalogPermanent(g, me.ID, "Mine", "Creature — Bear", "", false)
	theirs := pushCatalogPermanent(g, opp.ID, "Theirs", "Creature — Bear", "", false)
	theirFodder := pushCatalogPermanent(g, opp.ID, "Their Fodder", "Creature — Goblin", "", false)

	// An OPPONENT's sacrifice grows MY board.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirFodder) })
	passPriorityAroundTable(t, g)

	if got := counterCount(g, mazirek, game.CounterPlusOne); got != 1 {
		t.Errorf("Mazirek counts himself: %d counters, want 1", got)
	}
	if got := counterCount(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("my other creature has %d counters, want 1", got)
	}
	if got := counterCount(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("'creatures YOU control' must not grow theirs: %d", got)
	}
}

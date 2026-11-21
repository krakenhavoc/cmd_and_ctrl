package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- token producers ---------------------------------------------

func TestGoblinSorceriesMakeTheirTokens(t *testing.T) {
	for _, tc := range []struct {
		name   string
		oracle string
		want   int
	}{
		{"Dragon Fodder", effectstest.DragonFodderOracle, 2},
		{"Hordeling Outburst", effectstest.HordelingOutburstOracle, 3},
	} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.CastCatalogSpell(t, g, tc.name, "Sorcery", tc.oracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CountBattlefieldNamed(g, me.ID, "Goblin"); got != tc.want {
			t.Errorf("%s made %d Goblins, want %d", tc.name, got, tc.want)
		}
	}
}

// Grave Titan is the catalog's proof that "enters OR attacks" is one
// ability with two conditions: the same declaration fires on both,
// and the attack half runs on EventAttack — which issue #73 recorded
// as missing and which has existed since S22.
func TestGraveTitanMakesZombiesOnEntryAndOnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]

	effectstest.CastCatalogSpell(t, g, "Grave Titan", "Creature — Giant", effectstest.GraveTitanOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); got != 2 {
		t.Fatalf("%d Zombies after the ETB, want 2", got)
	}

	// The cast Titan is summoning sick, so attack with one that is
	// already settled.
	titan := effectstest.PushCatalogPermanent(g, me.ID, "Grave Titan", "Creature — Giant", effectstest.GraveTitanOracle, false)
	effectstest.DeclareAttack(t, g, opp.ID, titan)
	if n := effectstest.TriggersOnStackFrom(g, titan); n != 1 {
		t.Fatalf("%d attack triggers, want 1", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Zombie"); got != 4 {
		t.Errorf("%d Zombies after the attack, want 4", got)
	}
}

func TestPawnOfUlamogMakesSpawnForNontokenDeathsOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Pawn of Ulamog", "Creature — Vampire Shaman", effectstest.PawnOfUlamogOracle, false)

	// A token dying gives nothing — that clause is what stops the
	// Spawns looping into more Spawns.
	token := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: token, Name: "Goblin", TypeLine: "Token Creature — Goblin",
		Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID,
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(token) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Eldrazi Spawn"); got != 0 {
		t.Errorf("a token death made %d Spawn, want 0", got)
	}

	// A nontoken one does, once the "you may" is answered.
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Human", "", false)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(fodder) })
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountBattlefieldNamed(g, me.ID, "Eldrazi Spawn"); got != 1 {
		t.Errorf("a nontoken death made %d Spawn, want 1", got)
	}
}

func TestSternLessonCreatesATappedPowerstone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Stern Lesson", "Instant", effectstest.SternLessonOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	// The Powerstone follows the discard (CR 608.2c, #4155): answer it first.
	effectstest.AnswerChooseCards(t, g, me.ID, me.Hand.Cards[0].InstanceID)
	g.SettleResolution()

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]

	var ids []uuid.UUID
	g.WithWriteLock(func() {
		var err error
		ids, err = g.CreateTokensForEffect(me.ID, effects.RedGoblinToken(), 2, game.TokenEntryOptions{
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
		c, ok := effectstest.BattlefieldCard(g, id)
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
	if kw := effects.RedGoblinToken().Keywords; len(kw) != 0 {
		t.Errorf("the template grew keywords: %v", kw)
	}
}

// --- sacrifice outlets -------------------------------------------

func TestBloodflowConnoisseurGrowsOnEachSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vamp := effectstest.PushCatalogPermanent(g, me.ID, "Bloodflow Connoisseur", "Creature — Vampire", effectstest.BloodflowOracle, false)
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)

	if err := g.ActivateCatalogAbility(me.ID, vamp, 0, game.ActivateAbilityParams{
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(fodder) {
		t.Errorf("the sacrifice is a cost, paid at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, vamp, game.CounterPlusOne); got != 1 {
		t.Errorf("+1/+1 counters = %d, want 1", got)
	}
}

// --- the commanders ----------------------------------------------

// Korvold's two abilities feed each other: the first one's mandatory
// sacrifice is what triggers the second.
func TestKorvoldSacrificesOnEntryThenGrowsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	fodder := effectstest.PushCatalogPermanent(g, me.ID, "Fodder", "Creature — Goblin", "", false)
	handBefore := len(me.Hand.Cards)

	korvold := effectstest.CastCatalogSpell(t, g, "Korvold, Fae-Cursed King",
		"Legendary Creature — Dragon Noble", effectstest.KorvoldOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SacrificeChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("Korvold's entry should demand a sacrifice")
	}
	for _, opt := range c.SacrificeOptions {
		if opt == korvold {
			t.Errorf("'another permanent' must not offer Korvold himself")
		}
	}
	effectstest.AnswerSacrifice(t, g, me.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)

	if g.Battlefield.Contains(fodder) {
		t.Errorf("the sacrifice should have happened")
	}
	if got := effectstest.CounterCount(g, korvold, game.CounterPlusOne); got != 1 {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	korvold := effectstest.CastCatalogSpell(t, g, "Korvold, Fae-Cursed King",
		"Legendary Creature — Dragon Noble", effectstest.KorvoldOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if c := effectstest.SacrificeChoiceFor(g, me.ID); c != nil {
		t.Errorf("no legal permanent should mean no prompt, got %+v", c.SacrificeOptions)
	}
	if !g.Battlefield.Contains(korvold) {
		t.Errorf("Korvold must not eat himself")
	}
}

func TestMazirekGrowsEveryCreatureOnAnyPlayersSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mazirek := effectstest.PushCatalogPermanent(g, me.ID, "Mazirek, Kraul Death Priest",
		"Legendary Creature — Insect Shaman", effectstest.MazirekOracle, false)
	mine := effectstest.PushCatalogPermanent(g, me.ID, "Mine", "Creature — Bear", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Theirs", "Creature — Bear", "", false)
	theirFodder := effectstest.PushCatalogPermanent(g, opp.ID, "Their Fodder", "Creature — Goblin", "", false)

	// An OPPONENT's sacrifice grows MY board.
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirFodder) })
	effectstest.PassPriorityAroundTable(t, g)

	if got := effectstest.CounterCount(g, mazirek, game.CounterPlusOne); got != 1 {
		t.Errorf("Mazirek counts himself: %d counters, want 1", got)
	}
	if got := effectstest.CounterCount(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("my other creature has %d counters, want 1", got)
	}
	if got := effectstest.CounterCount(g, theirs, game.CounterPlusOne); got != 0 {
		t.Errorf("'creatures YOU control' must not grow theirs: %d", got)
	}
}

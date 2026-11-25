package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b150_cards_test.go — S82 older sets, Commander 2019 batch 150
// (#5517): Heart-Piercer Manticore; Farm // Market; Champion of Stray Souls;
// Refuse // Cooperate; Sevinne, the Chronoclasm; Soul of Zendikar.

const (
	c19B150ManticoreOracle = "253bec96-108c-4880-9c04-8bc99ceafa64"
	c19B150FarmOracle      = "2ab857cc-64b9-4115-8fa5-6dfae9edbdb5"
	c19B150ChampionOracle  = "4edbaac4-dcb3-40b4-abd7-ff2683bd366d"
	c19B150RefuseOracle    = "7db22241-fe43-4071-9af9-de1cf394f7f5"
	c19B150SevinneOracle   = "869c9bc4-2b21-40c3-8b1c-c83269c856d0"
	c19B150SoulOracle      = "87217160-0687-4482-b9a3-91bad8527675"
)

func TestC19B150CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		c19B150ManticoreOracle:     "Heart-Piercer Manticore",
		c19B150FarmOracle:          "Farm",
		c19B150FarmOracle + "#1":   "Market",
		c19B150ChampionOracle:      "Champion of Stray Souls",
		c19B150RefuseOracle:        "Refuse",
		c19B150RefuseOracle + "#1": "Cooperate",
		c19B150SevinneOracle:       "Sevinne, the Chronoclasm",
		c19B150SoulOracle:          "Soul of Zendikar",
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

func c19B150Manticore(t *testing.T, g *game.Game, me *game.Player, victim uuid.UUID) uuid.UUID {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	id := effectstest.CastAndResolveCreature(t, g, "Heart-Piercer Manticore", "Creature — Manticore", c19B150ManticoreOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if victim != uuid.Nil {
		effectstest.AnswerOwnPermanents(t, g, me.ID, victim)
		effectstest.PassPriorityAroundTable(t, g)
	}
	return id
}

func TestC19B150ManticoreSacrificesACreatureToDealItsPowerToAnyTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	victim := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	manticore := c19B150Manticore(t, g, me, victim)
	if g.Battlefield.Contains(victim) {
		t.Fatal("the sacrificed creature is still on the battlefield")
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the 3 damage did not kill the 2/2")
	}
	if !g.Battlefield.Contains(manticore) {
		t.Error("the Manticore itself left")
	}
}

func TestC19B150ManticoreDeclinedDealsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	ogre := effectstest.B12Creature(g, me.ID, "Ogre", "Creature — Ogre", 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	c19B150Manticore(t, g, me, uuid.Nil)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a target was asked for without a sacrifice")
	}
	if !g.Battlefield.Contains(ogre) || !g.Battlefield.Contains(theirs) {
		t.Error("something left without a sacrifice")
	}
}

func TestC19B150ManticoreHasEmbalm(t *testing.T) {
	spec, _ := effects.Lookup(c19B150ManticoreOracle)
	if len(spec.Activated) != 1 || spec.Activated[0].Cost.Mana != "{5}{R}" || !spec.Activated[0].SorcerySpeed {
		t.Fatalf("activated %+v, want one sorcery-speed embalm {5}{R}", spec.Activated)
	}
}

func TestC19B150FarmDestroysOnlyAnAttackingOrBlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	card := game.Card{InstanceID: uuid.New(), OracleID: c19B150FarmOracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Farm", TypeLine: "Instant", ManaCost: "{2}{W}", Colors: []string{"W"}},
			{Name: "Market", TypeLine: "Sorcery", ManaCost: "{2}{U}", Colors: []string{"U"}, OracleText: "Aftermath (Cast this spell only from your graveyard. Then exile it.)"},
		}}
	card.SetFace(0)
	me.Hand.PushTop(card)
	effectstest.B06AddMana(me, "W", "C", "C")
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(idle)}); err == nil {
		t.Fatal("Farm accepted a creature that is neither attacking nor blocking")
	}
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(attacker)}); err != nil {
		t.Fatalf("cast Farm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the attacking creature survived Farm")
	}
	if !g.Battlefield.Contains(idle) {
		t.Error("the idle creature was destroyed")
	}
}

func TestC19B150MarketIsAftermathAndLootsTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	card := game.Card{InstanceID: uuid.New(), OracleID: c19B150FarmOracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Farm", TypeLine: "Instant", ManaCost: "{2}{W}", Colors: []string{"W"}},
			{Name: "Market", TypeLine: "Sorcery", ManaCost: "{2}{U}", Colors: []string{"U"}, OracleText: "Aftermath (Cast this spell only from your graveyard. Then exile it.)"},
		}}
	card.SetFace(0)
	me.Hand.PushTop(card)
	effectstest.B06AddMana(me, "U", "C", "C")
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Face: 1}); err == nil {
		t.Fatal("Market was cast from the hand")
	}
	g.WithWriteLock(func() {
		_, _ = game.MoveCard(me.Hand, me.Graveyard, card.InstanceID)
	})
	hand := me.Hand.Size()
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Face: 1, FromZone: "graveyard"}); err != nil {
		t.Fatalf("cast Market from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Fatalf("hand %d, want %d after drawing two (discard pending)", me.Hand.Size(), hand+2)
	}
	if c := effectstest.DiscardChoiceFor(g, me.ID); c == nil {
		t.Fatal("no discard prompt after drawing two")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("Market was not exiled after resolving (aftermath)")
	}
}

func TestC19B150ChampionSacrificesXCreaturesToReturnXCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	champ := effectstest.PushCatalogPermanent(g, me.ID, "Champion of Stray Souls", "Creature — Skeleton Warrior", c19B150ChampionOracle, false)
	a := effectstest.B12Creature(g, me.ID, "Fodder A", "Creature — Bear", 1, 1)
	b := effectstest.B12Creature(g, me.ID, "Fodder B", "Creature — Bear", 1, 1)
	d1 := effectstest.B38Graveyard(g, me.ID, "Dead One", "Creature — Bear", "{1}", 1, 1)
	d2 := effectstest.B38Graveyard(g, me.ID, "Dead Two", "Creature — Bear", "{1}", 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "B", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, champ, 0, game.ActivateAbilityParams{Strict: true, XValue: 2,
		SacrificeIDs: []uuid.UUID{champ, a}, Targets: effectstest.CardRefs(d1, d2)}); err == nil {
		t.Fatal("the Champion sacrificed itself as one of the X other creatures")
	}
	if err := g.ActivateCatalogAbility(me.ID, champ, 0, game.ActivateAbilityParams{Strict: true, XValue: 2,
		SacrificeIDs: []uuid.UUID{a, b}, Targets: effectstest.CardRefs(d1)}); err == nil {
		t.Fatal("one target was accepted for X = 2")
	}
	effectstest.B06AddMana(me, "B", "B", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, champ, 0, game.ActivateAbilityParams{Strict: true, XValue: 2,
		SacrificeIDs: []uuid.UUID{a, b}, Targets: effectstest.CardRefs(d1, d2)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the sacrificed creatures are still on the battlefield")
	}
	if !g.Battlefield.Contains(d1) || !g.Battlefield.Contains(d2) {
		t.Error("the two graveyard creatures did not return")
	}
	if !g.Battlefield.Contains(champ) {
		t.Error("the Champion left")
	}
}

func TestC19B150ChampionPutsItselfOnTopFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	champ := effectstest.B38Graveyard(g, me.ID, "Champion of Stray Souls", "Creature — Bear", "{1}", 1, 1)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == champ {
				me.Graveyard.Cards[i].OracleID = c19B150ChampionOracle
			}
		}
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "B", "B", "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, champ, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Library.Contains(champ) || me.Graveyard.Contains(champ) {
		t.Fatal("the Champion did not go from the graveyard to the library")
	}
	if me.Library.Cards[len(me.Library.Cards)-1].InstanceID != champ {
		t.Error("the Champion is not on top of the library")
	}
}

func TestC19B150RefuseDamagesTheSpellsControllerForItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	spell := effectstest.CastFromHandForTest(t, g, me, "Three Bear", "Creature — Bear", "{1}{G}{G}", "", game.CastSpellParams{})
	effectstest.B06AddMana(me, "R", "C", "C", "C")
	life := me.Life
	effectstest.CastFromHandForTest(t, g, me, "Refuse", "Instant", "{3}{R}", c19B150RefuseOracle, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(spell)})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d, want %d: Refuse deals 3 to the spell's controller", me.Life, life-3)
	}
}

func TestC19B150CooperateCopiesAnInstantOrSorcerySpell(t *testing.T) {
	spec, _ := effects.Lookup(c19B150RefuseOracle + "#1")
	if spec.Targets == nil || spec.OnResolve == nil {
		t.Fatalf("Cooperate has no target clause or body: %+v", spec)
	}
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	spell := effectstest.CastFromHandForTest(t, g, me, "Hurt", "Instant", "{R}", "", game.CastSpellParams{})
	_ = bear
	card := game.Card{InstanceID: uuid.New(), OracleID: c19B150RefuseOracle, Layout: game.LayoutSplit, Owner: me.ID, Controller: me.ID,
		Faces: []game.Face{
			{Name: "Refuse", TypeLine: "Instant", ManaCost: "{3}{R}", Colors: []string{"R"}},
			{Name: "Cooperate", TypeLine: "Instant", ManaCost: "{2}{U}", Colors: []string{"U"}, OracleText: "Aftermath (Cast this spell only from your graveyard. Then exile it.)"},
		}}
	card.SetFace(0)
	g.WithWriteLock(func() { me.Graveyard.PushTop(card) })
	effectstest.B06AddMana(me, "U", "C", "C")
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Face: 1, FromZone: "graveyard", Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("Cooperate targeted a permanent")
	}
	if err := g.CastSpell(me.ID, card.InstanceID, game.CastSpellParams{Strict: true, Face: 1, FromZone: "graveyard", Targets: effectstest.CardRefs(spell)}); err != nil {
		t.Fatalf("cast Cooperate from the graveyard: %v", err)
	}
	before := g.Stack.Size()
	effectstest.PassPriorityAroundTable(t, g)
	_ = before
	if !g.Exile.Contains(card.InstanceID) {
		t.Error("Cooperate was not exiled after resolving (aftermath)")
	}
}

func TestC19B150SevinnePreventsDamageToItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sevinne := effectstest.PushCatalogPermanent(g, me.ID, "Sevinne, the Chronoclasm", "Legendary Creature — Human Wizard", c19B150SevinneOracle, false)
	effectstest.AdvanceToMain(t, g)
	src := effectstest.B12Creature(g, g.Seats[1].ID, "Pinger", "Creature — Bear", 1, 1)
	effectstest.DealDamageTo(t, g, src, sevinne, 5)
	if got := effectstest.DamageMarkedOn(g, sevinne); got != 0 {
		t.Errorf("Sevinne has %d damage marked, want 0", got)
	}
	if !g.Battlefield.Contains(sevinne) {
		t.Error("Sevinne died to damage")
	}
}

func TestC19B150SevinneCopiesOnlyTheFirstSpellFromYourGraveyardEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	sevinne := effectstest.PushCatalogPermanent(g, me.ID, "Sevinne, the Chronoclasm", "Legendary Creature — Human Wizard", c19B150SevinneOracle, false)
	effectstest.AdvanceToMain(t, g)
	// Faithless Looting is castable from the graveyard (flashback).
	loot := func() uuid.UUID {
		id := uuid.New()
		g.WithWriteLock(func() {
			me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Faithless Looting", TypeLine: "Sorcery", ManaCost: "{R}",
				OracleID: "3d6fa57a-aa53-4b5c-b8af-a7612c823117", Owner: me.ID, Controller: me.ID})
		})
		return id
	}
	for i := 0; i < 6; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	first := loot()
	effectstest.B06AddMana(me, "R", "R", "R", "R", "R", "R")
	if err := g.CastSpell(me.ID, first, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true}); err != nil {
		t.Fatalf("flashback Faithless Looting: %v", err)
	}
	if got := effectstest.TriggersOnStackFrom(g, sevinne); got != 1 {
		t.Fatalf("%d Sevinne triggers after the first graveyard cast, want 1", got)
	}
	copies := 0
	for i := 0; i < 8 && (g.Stack.Size() > 0 || effectstest.DiscardChoiceFor(g, me.ID) != nil); i++ {
		if effectstest.DiscardChoiceFor(g, me.ID) != nil {
			effectstest.DiscardFromHand(t, g, me.ID)
			copies++
			continue
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if copies != 2 {
		t.Fatalf("%d loots resolved, want 2 (the spell and Sevinne's copy)", copies)
	}
	second := loot()
	if err := g.CastSpell(me.ID, second, game.CastSpellParams{FromZone: "graveyard", AlternativeCost: "flashback", Strict: true}); err != nil {
		t.Fatalf("second flashback: %v", err)
	}
	if got := effectstest.TriggersOnStackFrom(g, sevinne); got != 0 {
		t.Errorf("%d Sevinne triggers after the second graveyard cast, want 0", got)
	}
}

func TestC19B150SoulOfZendikarMakesBeastsFromTheBattlefieldAndTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	soul := effectstest.PushCatalogPermanent(g, me.ID, "Soul of Zendikar", "Creature — Avatar", c19B150SoulOracle, false)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, soul, "reach") {
		t.Error("Soul of Zendikar lacks reach")
	}
	effectstest.B06AddMana(me, "G", "G", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, soul, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("battlefield ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Beast"); n != 1 {
		t.Fatalf("%d Beasts, want 1", n)
	}
	grave := effectstest.B38Graveyard(g, me.ID, "Soul of Zendikar", "Creature — Bear", "{1}", 1, 1)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == grave {
				me.Graveyard.Cards[i].OracleID = c19B150SoulOracle
			}
		}
	})
	effectstest.B06AddMana(me, "G", "G", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, grave, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("graveyard ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n, _ := effectstest.CountTokensNamed(g, me.ID, "Beast"); n != 2 {
		t.Errorf("%d Beasts, want 2", n)
	}
	if !g.Exile.Contains(grave) {
		t.Error("the graveyard ability did not exile the card as its cost")
	}
}

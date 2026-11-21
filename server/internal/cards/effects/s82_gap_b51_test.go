package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapTuskerOracle     = "0a994401-d37c-4e30-9410-422e7ae5a715"
	gapRepurposerOracle = "ae3e5c80-6a6a-4781-a7b9-4d63357ffa3b"
	gapBroodscaleOracle = "a7c69802-99c4-4778-934e-cc09ba58c267"
	gapOozewaggOracle   = "07b0b2ad-4364-4a8d-94d5-8bf984b23d05"
	gapNyxbornOracle    = "4d1b655f-05b8-4b82-9cb9-67660f570a9b"
	gapElkOracle        = "5bca05c1-1c0a-45f4-aa20-69f68b99cd7e"
	gapMycospawnOracle  = "63a19c7c-49d7-4b75-aac8-d0ff36c45fb8"
	gapCullerOracle     = "a075a37b-beb6-409d-b5b6-d291921cdef0"
	gapWumpusOracle     = "c116801e-9893-46e8-a967-352b5fa27e86"
)

func gapB51Spawns(g *game.Game) int { return randomTokenCount(g, "Eldrazi Spawn") }

func gapB51Float(t *testing.T, g *game.Game, p *game.Player, mana ...string) {
	t.Helper()
	for _, m := range mana {
		floatMana(t, g, p, m)
	}
}

func gapB51Stock(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			p.Library.PushTop(game.NewCard("Filler", p.ID))
		}
	})
}

func gapB51Counters(g *game.Game, id uuid.UUID) int {
	c, _ := battlefieldCard(g, id)
	return c.Counters[game.CounterPlusOne]
}

func TestGapB51TuskerMakesASpawnWhenCast(t *testing.T) {
	g := newCatalogGame(t)
	castCatalogSpell(t, g, "Warped Tusker", "Creature — Eldrazi Boar Beast", gapTuskerOracle, nil)
	passPriorityAroundTable(t, g)
	if got := gapB51Spawns(g); got != 1 {
		t.Errorf("Spawn tokens = %d, want 1", got)
	}
}

func TestGapB51TuskerMakesASpawnWhenCycled(t *testing.T) {
	g := newCatalogGame(t)
	lciSTActivateFromHand(t, g, "Warped Tusker", "Creature — Eldrazi Boar Beast", gapTuskerOracle, "{G}{G}{G}", 0)
	passPriorityAroundTable(t, g)
	if got := gapB51Spawns(g); got != 1 {
		t.Errorf("Spawn tokens = %d, want 1", got)
	}
}

func TestGapB51RepurposerMakesASpawnWhenCastAndWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := castCatalogSpell(t, g, "Eldrazi Repurposer", "Creature — Eldrazi Drone", gapRepurposerOracle, nil)
	passPriorityAroundTable(t, g)
	if got := gapB51Spawns(g); got != 1 {
		t.Fatalf("Spawn tokens after the cast = %d, want 1", got)
	}
	killCreature(t, g, me.ID, id)
	passPriorityAroundTable(t, g)
	if got := gapB51Spawns(g); got != 2 {
		t.Errorf("Spawn tokens after it died = %d, want 2", got)
	}
}

func TestGapB51BroodscaleMakesASpawnOnlyIfYouSayYes(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		id := pushCatalogPermanent(g, me.ID, "Basking Broodscale", "Creature — Eldrazi Lizard", gapBroodscaleOracle, false)
		gapB51Float(t, g, me, "{G}", "{G}")
		if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		answerLatestTriggerPrompt(t, g, me.ID, yes)
		passPriorityAroundTable(t, g)
		if got := gapB51Counters(g, id); got != 1 {
			t.Errorf("counters = %d, want 1", got)
		}
		want := 0
		if yes {
			want = 1
		}
		if got := gapB51Spawns(g); got != want {
			t.Errorf("yes=%v: Spawn tokens = %d, want %d", yes, got, want)
		}
	}
}

func TestGapB51OozewaggGivesTrampleToModifiedCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Temperamental Oozewagg", "Creature — Ooze Brushwagg", gapOozewaggOracle, false)
	modified := pushVanillaCreature(g, me.ID, "Modified", 2, 2)
	plain := pushVanillaCreature(g, me.ID, "Plain", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	for _, id := range []uuid.UUID{modified, theirs} {
		g.WithWriteLock(func() {
			if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(NewContext(g, nil)); err != nil {
				t.Fatal(err)
			}
		})
	}
	g.WithWriteLock(func() { g.BumpLayerVersionForTest() })
	if !hasEffectiveKeyword(t, g, modified, "trample") {
		t.Error("a modified creature you control has no trample")
	}
	if hasEffectiveKeyword(t, g, plain, "trample") {
		t.Error("an unmodified creature has trample")
	}
	if hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's modified creature has trample")
	}
}

func TestGapB51ElkTurnsTheEnchantedCreatureIntoAGreenElk(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	host := pushDiesCreatureForTest(g, opp.ID, "Dragon", "", "Creature — Dragon", 5, 5)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == host {
				g.Battlefield.Cards[i].Keywords = []string{"flying"}
			}
		}
	})
	id := bestowCard(t, g, "Trickster's Elk", "Enchantment Creature — Elk", gapElkOracle, 3, 3)
	gapB51Float(t, g, g.Seats[0], "{G}", "{G}")
	castBestowed(t, g, id, host)
	passPriorityAroundTable(t, g)
	if p, tough := effectivePower(t, g, host), effectiveToughness(t, g, host); p != 3 || tough != 3 {
		t.Errorf("enchanted creature is %d/%d, want 3/3", p, tough)
	}
	if hasEffectiveKeyword(t, g, host, "flying") {
		t.Error("the enchanted creature kept its flying")
	}
	c := battlefieldCardCopy(t, g, host)
	if !c.HasSubtype("Elk") || c.HasSubtype("Dragon") {
		t.Errorf("Elk=%v Dragon=%v, want an Elk and no Dragon", c.HasSubtype("Elk"), c.HasSubtype("Dragon"))
	}
}

func TestGapB51NyxbornHydraEntersWithXCountersAndBestowsThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	host := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	id := bestowCard(t, g, "Nyxborn Hydra", "Enchantment Creature — Hydra", gapNyxbornOracle, 0, 1)
	gapB51Float(t, g, me, "{G}", "{G}", "{G}", "{G}", "{G}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		AlternativeCost: game.BestowKey, XValue: 3, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: host}},
	}); err != nil {
		t.Fatalf("bestow with X=3: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := gapB51Counters(g, id); got != 3 {
		t.Fatalf("the Aura has %d counters, want 3", got)
	}
	if p, tough := effectivePower(t, g, host), effectiveToughness(t, g, host); p != 5 || tough != 5 {
		t.Errorf("enchanted creature is %d/%d, want 5/5", p, tough)
	}
	if !hasEffectiveKeyword(t, g, host, "reach") || !hasEffectiveKeyword(t, g, host, "trample") {
		t.Error("the enchanted creature lacks reach or trample")
	}
}

func TestGapB51NyxbornHydraCastNormallyIsAnXCounterCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	id := bestowCard(t, g, "Nyxborn Hydra", "Enchantment Creature — Hydra", gapNyxbornOracle, 0, 1)
	gapB51Float(t, g, me, "{G}", "{G}", "{G}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Strict: true}); err != nil {
		t.Fatalf("cast with X=2: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := gapB51Counters(g, id); got != 2 {
		t.Fatalf("hydra has %d counters, want 2", got)
	}
	if p := effectivePowerWithCounters(t, g, id); p != 2 {
		t.Errorf("hydra power with its counters is %d, want 2", p)
	}
}

func TestGapB51MycospawnFetchesALandAndExilesOneWhenKicked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	forest := game.NewCard("Forest", me.ID)
	forest.TypeLine = "Basic Land — Forest"
	g.WithWriteLock(func() { me.Library.PushTop(forest) })
	theirLand := pushCatalogPermanent(g, opp.ID, "Their Island", "Basic Land — Island", "", false)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Sowing Mycospawn", TypeLine: "Creature — Eldrazi Fungus",
		OracleID: gapMycospawnOracle, ManaCost: "{3}{G}", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	gapB51Float(t, g, me, "{G}", "{G}", "{G}", "{G}", "{G}", "{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("kicked cast: %v", err)
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID)
	if pick != nil {
		if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: theirLand}); err != nil {
			t.Fatal(err)
		}
	}
	if latestChoiceOfKindFor(g, game.PendingChoiceTriggerOrder, me.ID) != nil {
		answerTriggerOrderInOfferedOrder(t, g)
	}
	passPriorityAroundTable(t, g)
	if search := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID); search != nil {
		if err := g.ResolveSearchLibrary(search.ID, me.ID, []uuid.UUID{forest.InstanceID}); err != nil {
			t.Fatal(err)
		}
	}
	passPriorityAroundTable(t, g)
	c, ok := battlefieldCard(g, forest.InstanceID)
	if !ok || c.Tapped {
		t.Errorf("the fetched land: on battlefield=%v tapped=%v, want untapped on the battlefield", ok, c.Tapped)
	}
	if g.Battlefield.Contains(theirLand) {
		t.Error("the kicked trigger did not exile target land")
	}
}

func TestGapB51CullerTakesACreatureCardOffTheTopOnLandfall(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Territory Culler", "Creature — Eldrazi", gapCullerOracle, false)
	bear := game.NewCard("Top Bear", me.ID)
	bear.TypeLine = "Creature — Bear"
	g.WithWriteLock(func() { me.Library.PushTop(bear) })
	playLandFromHand(t, g, "Forest", "")
	passPriorityAroundTable(t, g)
	ask := latestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
	if ask == nil {
		t.Fatal("no prompt to take the creature card")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatal(err)
	}
	if !me.Hand.Contains(bear.InstanceID) {
		t.Error("the creature card did not reach the hand")
	}
}

func TestGapB51WumpusLetsAnOpponentPutACreatureInOnlyWithoutColorlessMana(t *testing.T) {
	for _, colorless := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		held := game.NewCard("Held Bear", opp.ID)
		held.TypeLine = "Creature — Bear"
		opp.Hand.PushTop(held)
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Wumpus Aberration", TypeLine: "Creature — Eldrazi Beast",
			OracleID: gapWumpusOracle, ManaCost: "{3}{G}", Power: 6, Toughness: 6, Owner: me.ID, Controller: me.ID})
		advanceToMain(t, g)
		if colorless {
			gapB51Float(t, g, me, "{G}", "{G}", "{G}", "{C}")
		} else {
			gapB51Float(t, g, me, "{G}", "{G}", "{G}", "{G}")
		}
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
			t.Fatal(err)
		}
		if pick := latestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID); pick != nil {
			if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID}); err != nil {
				t.Fatal(err)
			}
		}
		passPriorityAroundTable(t, g)
		if colorless {
			if len(g.PendingChoices) != 0 && chooseCardsChoiceFor(g, opp.ID) != nil {
				t.Error("the trigger fired although {C} was spent")
			}
			continue
		}
		answerChooseCards(t, g, opp.ID, held.InstanceID)
		c, ok := battlefieldCard(g, held.InstanceID)
		if !ok || c.Controller != opp.ID {
			t.Errorf("the creature card: on battlefield=%v, controller %v, want under the opponent's control", ok, c.Controller)
		}
	}
}

const gapEladamriOracle = "9aff3d55-e25d-4f06-a783-c3aa9eb9fd50"

func gapB51EladamriTable(t *testing.T) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	el := pushCatalogPermanent(g, me.ID, "Eladamri, Korvecdal", "Legendary Creature — Elf Warrior", gapEladamriOracle, false)
	a := pushVanillaCreature(g, me.ID, "Tapper A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "Tapper B", 1, 1)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	advanceToMain(t, g)
	return g, me, el, []uuid.UUID{a, b}
}

func gapB51EladamriActivate(t *testing.T, g *game.Game, me *game.Player, el uuid.UUID, tapped []uuid.UUID) error {
	t.Helper()
	floatMana(t, g, me, "{G}")
	return g.ActivateCatalogAbility(me.ID, el, 0, game.ActivateAbilityParams{Strict: true, TapIDs: tapped})
}

func TestGapB51EladamriPutsARevealedTopCreatureOntoTheBattlefield(t *testing.T) {
	g, me, el, tapped := gapB51EladamriTable(t)
	bear := game.NewCard("Top Bear", me.ID)
	bear.TypeLine = "Creature — Bear"
	g.WithWriteLock(func() { me.Library.PushTop(bear) })
	if err := gapB51EladamriActivate(t, g, me, el, tapped); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceOptionPick {
			t.Fatal("asked where to reveal from with an empty hand")
		}
	}
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Top Bear" {
			found = true
		}
	}
	if !found {
		t.Error("the revealed creature card did not enter")
	}
}

func TestGapB51EladamriHandRevealPutsOnlyCreaturesIn(t *testing.T) {
	for _, creature := range []bool{true, false} {
		g, me, el, tapped := gapB51EladamriTable(t)
		gapB51Stock(g, me, 2)
		card := game.NewCard("Held Card", me.ID)
		card.TypeLine = "Instant"
		if creature {
			card.TypeLine = "Creature — Bear"
		}
		me.Hand.PushTop(card)
		if err := gapB51EladamriActivate(t, g, me, el, tapped); err != nil {
			t.Fatal(err)
		}
		passPriorityAroundTable(t, g)
		answerOptionPick(t, g, me.ID, 0)
		answerChooseCards(t, g, me.ID, card.InstanceID)
		if got := g.Battlefield.Contains(card.InstanceID); got != creature {
			t.Errorf("creature=%v: on the battlefield=%v", creature, got)
		}
	}
}

func TestGapB51EladamriRefusesWithoutTwoOtherUntappedCreaturesAndOnAnOpponentsTurn(t *testing.T) {
	g, me, el, tapped := gapB51EladamriTable(t)
	if err := gapB51EladamriActivate(t, g, me, el, tapped[:1]); err == nil {
		t.Error("activated with only one other creature tapped")
	}
	g.WithWriteLock(func() { g.Turn.ActiveSeat = 1 })
	if err := gapB51EladamriActivate(t, g, me, el, tapped); err == nil {
		t.Error("activated during an opponent's turn")
	}
}

func TestGapB51EladamriLetsYouCastCreaturesFromTheTopOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Eladamri, Korvecdal", "Legendary Creature — Elf Warrior", gapEladamriOracle, false)
	bear := game.Card{InstanceID: uuid.New(), Name: "Top Bear", TypeLine: "Creature — Bear", ManaCost: "{G}", Power: 2, Toughness: 2,
		Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Library.PushTop(bear) })
	advanceToMain(t, g)
	floatMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, bear.InstanceID, game.CastSpellParams{Strict: true, FromZone: string(game.ZoneLibrary)}); err != nil {
		t.Fatalf("cast a creature from the top of the library: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear.InstanceID) {
		t.Error("the creature did not resolve")
	}
	sorcery := game.Card{InstanceID: uuid.New(), Name: "Top Spell", TypeLine: "Sorcery", ManaCost: "{G}", Owner: me.ID, Controller: me.ID}
	g.WithWriteLock(func() { me.Library.PushTop(sorcery) })
	floatMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, sorcery.InstanceID, game.CastSpellParams{Strict: true, FromZone: string(game.ZoneLibrary)}); err == nil {
		t.Error("cast a sorcery from the top of the library")
	}
}

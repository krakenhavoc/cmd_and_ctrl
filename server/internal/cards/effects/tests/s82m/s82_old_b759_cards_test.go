package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 OLD batch 759 (Tempest, #5628): Legacy's Allure, Torture Chamber,
// Bounty Hunter, Krakilin, Flowstone Sculpture, Dirtcowl Wurm, Ertai's
// Meddling and Souldrinker.

const (
	oldB759Allure    = "106ca7bd-6dad-43fe-af58-8667ebddb2eb"
	oldB759Chamber   = "12436f2f-eaed-4b46-8570-81316ce15962"
	oldB759Hunter    = "2fd1cfc5-fbcd-4d31-ba8b-759d8c41b435"
	oldB759Krakilin  = "3eb49097-42f8-4388-907d-32a738749d8a"
	oldB759Sculpture = "42b01697-8663-4b53-86d5-699063bd0778"
	oldB759Wurm      = "608293f1-34aa-4e55-8389-73f975858de6"
	oldB759Meddling  = "609c529a-b3b3-46d5-b9d0-ae0117507d6e"
	oldB759Drinker   = "624d4bda-448f-43d1-92b6-53a77afe2374"
)

func TestOldB759EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB759Allure: "Legacy's Allure", oldB759Chamber: "Torture Chamber", oldB759Hunter: "Bounty Hunter",
		oldB759Krakilin: "Krakilin", oldB759Sculpture: "Flowstone Sculpture", oldB759Wurm: "Dirtcowl Wurm",
		oldB759Meddling: "Ertai's Meddling", oldB759Drinker: "Souldrinker",
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

func TestOldB759SouldrinkerPaysThreeLifeForACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Souldrinker", "Creature — Spirit", oldB759Drinker, false)
	life := me.Life
	if err := effectstest.InvB1Activate(g, me.ID, id, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 || effectstest.PlusOnes(g, id) != 1 {
		t.Errorf("life %d → %d, counters %d; want 3 life paid and one counter", life, me.Life, effectstest.PlusOnes(g, id))
	}
	me.Life = 2
	if err := effectstest.InvB1Activate(g, me.ID, id, 0); err == nil {
		t.Error("paid 3 life at 2 life")
	}
}

func TestOldB759KrakilinEntersWithXCountersAndRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.CastXSpell(t, g, "Krakilin", "Creature — Beast", oldB759Krakilin, "{X}{G}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, id); got != 3 {
		t.Fatalf("Krakilin entered with %d counters, want 3", got)
	}
	effectstest.FillPoolColored(me, "G", 2)
	if err := effectstest.InvB1Activate(g, me.ID, id, 0); err != nil {
		t.Fatalf("regenerate: %v", err)
	}
}

func TestOldB759DirtcowlWurmGrowsOnlyWhenAnOpponentPlaysALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	wurm := effectstest.PushCatalogPermanent(g, me.ID, "Dirtcowl Wurm", "Creature — Wurm", oldB759Wurm, false)
	effectstest.B34PlayLandAs(t, g, 0, "My Forest", "Basic Land — Forest")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, wurm); got != 0 {
		t.Errorf("my own land play gave %d counters, want 0", got)
	}
	effectstest.B34PlayLandAs(t, g, 1, "Their Island", "Basic Land — Island")
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.PlusOnes(g, wurm); got != 1 {
		t.Errorf("an opponent's land play gave %d counters, want 1", got)
	}
}

func TestOldB759BountyHunterMarksNonblackAndDestroysMarked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hunter := effectstest.PushCatalogPermanent(g, me.ID, "Bounty Hunter", "Creature — Human Archer Minion", oldB759Hunter, false)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	black := effectstest.TmpB1Push(g, opp.ID, "Black Bear", "Creature — Bear", "", 2, 2, "B")
	if err := effectstest.InvB1Activate(g, me.ID, hunter, 1, effectstest.CardTarget(bear)...); err == nil {
		t.Error("destroyed a creature with no bounty counter")
	}
	if err := effectstest.InvB1Activate(g, me.ID, hunter, 0, effectstest.CardTarget(black)...); err == nil {
		t.Error("put a bounty counter on a black creature")
	}
	if err := effectstest.InvB1Activate(g, me.ID, hunter, 0, effectstest.CardTarget(bear)...); err != nil {
		t.Fatalf("mark: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, "bounty"); got != 1 {
		t.Fatalf("bounty counters %d, want 1", got)
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == hunter {
				g.Battlefield.Cards[i].Tapped = false
			}
		}
	})
	if err := effectstest.InvB1Activate(g, me.ID, hunter, 1, effectstest.CardTarget(bear)...); err != nil {
		t.Fatalf("collect: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.CardOnBattlefield(g, bear); ok {
		t.Error("the marked creature survived")
	}
}

func TestOldB759TortureChamberCountsPainAndBurnsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	chamber := effectstest.PushCatalogPermanent(g, me.ID, "Torture Chamber", "Artifact", oldB759Chamber, false)
	effectstest.BeginUpkeep(t, g, me.ID)
	effectstest.BeginUpkeep(t, g, me.ID)
	if got := effectstest.CounterCount(g, chamber, "pain"); got != 2 {
		t.Fatalf("pain counters %d, want 2", got)
	}
	life := me.Life
	g.WithWriteLock(func() { g.EmitEvent(game.Event{Kind: game.EventBeginEndStep, Actor: me.ID}) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-2 {
		t.Errorf("end step: life %d → %d, want 2 damage", life, me.Life)
	}
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 3, 3)
	effectstest.FillPool(me, 1)
	if err := effectstest.InvB1Activate(g, me.ID, chamber, 0, effectstest.CardTarget(bear)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, bear); got != 2 {
		t.Errorf("damage on the bear %d, want 2", got)
	}
	if got := effectstest.CounterCount(g, chamber, "pain"); got != 0 {
		t.Errorf("pain counters left %d, want 0", got)
	}
}

func TestOldB759LegacysAllureStealsACreatureNoStrongerThanItsCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	allure := effectstest.PushCatalogPermanent(g, me.ID, "Legacy's Allure", "Enchantment", oldB759Allure, false)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 3, 3)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 1, 1)
	for i := 0; i < 2; i++ {
		if i == 0 {
			effectstest.WalkToUpkeepOfSeat0(t, g)
		} else {
			effectstest.B12ToMyNextUpkeep(t, g)
		}
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CounterCount(g, allure, "treasure"); got != 2 {
		t.Fatalf("treasure counters %d, want 2", got)
	}
	if err := effectstest.InvB1Activate(g, me.ID, allure, 0, effectstest.CardTarget(big)...); err == nil {
		t.Error("targeted a creature with power 3 at 2 counters")
	}
	if err := effectstest.InvB1Activate(g, me.ID, allure, 0, effectstest.CardTarget(small)...); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.CardOnBattlefield(g, allure); ok {
		t.Error("the Allure was not sacrificed")
	}
	c, ok := effectstest.CardOnBattlefield(g, small)
	if !ok || c.Controller != me.ID {
		t.Errorf("the small creature is not under my control (found %v, controller %v)", ok, c.Controller)
	}
}

func TestOldB759LegacysAllureNeedsCountersToTakeAnything(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	allure := effectstest.PushCatalogPermanent(g, me.ID, "Legacy's Allure", "Enchantment", oldB759Allure, false)
	zero := effectstest.B12Creature(g, opp.ID, "Zero", "Creature — Wall", 0, 4)
	if err := effectstest.InvB1Activate(g, me.ID, allure, 0, effectstest.CardTarget(zero)...); err != nil {
		t.Fatalf("a power-0 creature at 0 counters: %v", err)
	}
}

func TestOldB759FlowstoneSculptureOffersTheCounterAndEachKeyword(t *testing.T) {
	for i, want := range []string{"", "flying", "first strike", "trample"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		sc := effectstest.PushCatalogPermanent(g, me.ID, "Flowstone Sculpture", "Artifact Creature — Shapeshifter", oldB759Sculpture, false)
		effectstest.PushCatalogHandCard(me, "Filler", "Sorcery", "")
		effectstest.FillPool(me, 2)
		if err := g.ActivateCatalogAbility(me.ID, sc, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{me.Hand.Cards[len(me.Hand.Cards)-1].InstanceID}}); err != nil {
			t.Fatalf("activate: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		var pick *game.PendingChoice
		for _, c := range g.PendingChoices {
			if c != nil && c.Chooser == me.ID && c.Kind == game.PendingChoiceOptionPick {
				pick = c
			}
		}
		if pick == nil {
			t.Fatal("no option prompt")
		}
		if err := g.ResolveOptionPick(pick.ID, me.ID, i); err != nil {
			t.Fatalf("pick %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if want == "" {
			if effectstest.PlusOnes(g, sc) != 1 {
				t.Errorf("no +1/+1 counter")
			}
			continue
		}
		if !effectstest.HasEffectiveKeyword(t, g, sc, want) {
			t.Errorf("did not gain %s", want)
		}
		if effectstest.PlusOnes(g, sc) != 0 {
			t.Errorf("%s also took a counter", want)
		}
	}
}

func TestOldB759ErtaisMeddlingExilesTheSpellAndItReturnsAtUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	effectstest.FillPoolColored(me, "U", 2)
	effectstest.CastXSpell(t, g, "Ertai's Meddling", "Instant", oldB759Meddling, "{X}{U}", 1, effectstest.CardRefs(bolt))
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, bolt) {
		t.Fatal("the Bolt was not exiled")
	}
	life := opp.Life
	if life != 40 {
		t.Fatalf("the exiled Bolt still resolved: life %d", life)
	}
	effectstest.B12ToMyNextUpkeep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("the recast Bolt: life %d → %d, want 3 damage", life, opp.Life)
	}
}

func TestOldB759ErtaisMeddlingRefusesXOfZero(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bolt := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", effectstest.LightningBoltOracle, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Ertai's Meddling", TypeLine: "Instant", OracleID: oldB759Meddling, ManaCost: "{X}{U}", Owner: me.ID, Controller: me.ID})
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 0, Strict: true, Targets: effectstest.CardRefs(bolt)}); err == nil {
		t.Error("cast Ertai's Meddling with X=0")
	}
}

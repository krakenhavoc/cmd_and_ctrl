package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapDreamdrinkerOracle = "5b4e1163-09aa-4597-9a41-e1e384975056"
	gapEmperorOracle      = "cb70a360-898e-443a-b053-38e35ea2baa2"
	gapPteramanderOracle  = "3e254cae-6e14-4a40-bc02-7ca23750bb26"
	gapFetidOracle        = "b8c35a80-2c72-49ff-af5c-e8885c97f6fc"
	gapKamiOracle         = "5973d8f6-242c-4014-b66e-aa5c78440be7"
	gapFamiliarOracle     = "009e1638-9055-4872-93e1-85841bce4648"
	gapTransmograntOracle = "058dae67-807e-4542-ae51-f2b1216ffaed"
	gapSlingGangOracle    = "4e93b23c-a7f7-4bdd-bbca-0dc48bb5c223"
	gapLarvaOracle        = "72a39730-acc3-470c-960e-3a91c5eae4be"
)

func gapB49Counters(g *game.Game, id uuid.UUID) int {
	c, _ := effectstest.BattlefieldCard(g, id)
	return c.Counters[game.CounterPlusOne]
}

func gapB49Activate(t *testing.T, g *game.Game, who, id uuid.UUID, idx int) error {
	t.Helper()
	return g.ActivateCatalogAbility(who, id, idx, game.ActivateAbilityParams{Strict: true})
}

// gapFloatB puts n black mana in the pool: generic costs take any colour.
func gapFloatB(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		effectstest.FloatMana(t, g, p, "{B}")
	}
}

func gapB49Stock(g *game.Game, p *game.Player, n int) {
	g.WithWriteLock(func() {
		for i := 0; i < n; i++ {
			p.Library.PushTop(game.NewCard("Filler", p.ID))
		}
	})
}

func TestGapB49DreamdrinkerAdaptsOnceAndGainsMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Dreamdrinker Vampire", "Creature — Vampire", gapDreamdrinkerOracle, false)
	gapFloatB(t, g, me, 2)
	if err := gapB49Activate(t, g, me.ID, id, 0); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB49Counters(g, id); got != 1 {
		t.Fatalf("counters = %d, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, id, "menace") {
		t.Error("no menace after the counter was put on it")
	}
	// Adapt does nothing with counters already on it (CR 701.46a).
	gapFloatB(t, g, me, 2)
	if err := gapB49Activate(t, g, me.ID, id, 0); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB49Counters(g, id); got != 1 {
		t.Errorf("second adapt changed the counters to %d", got)
	}
}

func TestGapB49FetidGargantuaDrawsTwoAndLosesTwoWhenYouSayYes(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		gapB49Stock(g, me, 5)
		id := effectstest.PushCatalogPermanent(g, me.ID, "Fetid Gargantua", "Creature — Horror", gapFetidOracle, false)
		hand, life := me.Hand.Size(), me.Life
		gapFloatB(t, g, me, 3)
		if err := gapB49Activate(t, g, me.ID, id, 0); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := gapB49Counters(g, id); got != 2 {
			t.Fatalf("counters = %d, want 2", got)
		}
		ask := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceConfirm, me.ID)
		if ask == nil {
			t.Fatal("no may-draw prompt")
		}
		if err := g.ResolveConfirm(ask.ID, me.ID, yes); err != nil {
			t.Fatal(err)
		}
		wantHand, wantLife := hand, life
		if yes {
			wantHand, wantLife = hand+2, life-2
		}
		if me.Hand.Size() != wantHand || me.Life != wantLife {
			t.Errorf("yes=%v: hand %d life %d, want %d / %d", yes, me.Hand.Size(), me.Life, wantHand, wantLife)
		}
	}
}

func TestGapB49PteramanderCostsLessForEachOtherArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Etherium Pteramander", "Artifact Creature — Salamander Drake", gapPteramanderOracle, false)
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogPermanent(g, me.ID, "Trinket", "Artifact", "", false)
	}
	gapFloatB(t, g, me, 4) // {6}{B} less three other artifacts
	if err := gapB49Activate(t, g, me.ID, id, 0); err != nil {
		t.Fatalf("discounted adapt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB49Counters(g, id); got != 4 {
		t.Errorf("counters = %d, want 4", got)
	}
}

func TestGapB49PteramanderRefusesWithoutTheMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Etherium Pteramander", "Artifact Creature — Salamander Drake", gapPteramanderOracle, false)
	gapFloatB(t, g, me, 4)
	if err := gapB49Activate(t, g, me.ID, id, 0); err == nil {
		t.Error("adapt was paid for with four mana and no other artifacts")
	}
}

func TestGapB49KamiDrainsOnceEachTurnAndIsFreeAfterThreeDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapB49Stock(g, me, 6)
	id := effectstest.PushCatalogPermanent(g, me.ID, "Kami of Jealous Thirst", "Creature — Spirit", gapKamiOracle, false)
	if err := gapB49Activate(t, g, me.ID, id, 0); err == nil {
		t.Fatal("activated for free with no cards drawn")
	}
	g.WithWriteLock(func() { _ = g.DrawNForEffect(me.ID, 3) })
	life, oppLife := me.Life, opp.Life
	if err := gapB49Activate(t, g, me.ID, id, 0); err != nil {
		t.Fatalf("free activation after three draws: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 || opp.Life != oppLife-2 {
		t.Errorf("life %d/%d, want %d/%d", me.Life, opp.Life, life+2, oppLife-2)
	}
	gapFloatB(t, g, me, 5)
	if err := gapB49Activate(t, g, me.ID, id, 0); err == nil {
		t.Error("a second activation in one turn was allowed")
	}
}

func TestGapB49FamiliarMakesEachOpponentDiscardAndDrawsForThoseWhoCant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapB49Stock(g, me, 5)
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			p.Hand.Cards = nil
		}
	})
	held := pushHandCardForGap(opp, "Held")
	gapFloatB(t, g, me, 4)
	effectstest.CastCatalogSpell(t, g, "Refurbished Familiar", "Artifact Creature — Zombie Rat", gapFamiliarOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Chooser == opp.ID && len(c.ChooseCards) > 0 {
			if err := g.ResolveChooseCards(c.ID, opp.ID, []uuid.UUID{held}); err != nil {
				t.Fatal(err)
			}
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(held) {
		t.Error("the opponent did not discard")
	}
	opponents := len(g.Seats) - 1
	if got := me.Hand.Size(); got != opponents-1 {
		t.Errorf("drew %d cards, want one for each of the %d opponents who couldn't discard", got, opponents-1)
	}
}

func pushHandCardForGap(p *game.Player, name string) uuid.UUID {
	c := game.NewCard(name, p.ID)
	c.TypeLine = "Instant"
	p.Hand.PushTop(c)
	return c.InstanceID
}

func TestGapB49FamiliarCostsLessForEachArtifact(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Trinket", "Artifact", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Trinket", "Artifact", "", false)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Refurbished Familiar", TypeLine: "Artifact Creature — Zombie Rat",
		OracleID: gapFamiliarOracle, ManaCost: "{3}{B}", Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	gapFloatB(t, g, me, 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast with {B} against {3}{B} less two artifacts")
	}
	gapFloatB(t, g, me, 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("{1}{B} should pay {3}{B} with affinity for two artifacts: %v", err)
	}
}

func TestGapB49TransmograntReturnsTappedWithTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushGraveyardPermanent(me, "Retrofitted Transmogrant", "Artifact Creature — Zombie", "{B}")
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == id {
				me.Graveyard.Cards[i].OracleID = gapTransmograntOracle
			}
		}
	})
	gapFloatB(t, g, me, 4)
	if err := gapB49Activate(t, g, me.ID, id, 0); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, id)
	if !ok {
		t.Fatal("it did not return")
	}
	if !c.Tapped || c.Counters[game.CounterPlusOne] != 2 {
		t.Errorf("tapped=%v counters=%d, want tapped with two +1/+1 counters", c.Tapped, c.Counters[game.CounterPlusOne])
	}
}

func TestGapB49SlingGangMakesTwoGoblinsAndDrainsBySacrificingOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gapFloatB(t, g, me, 4)
	id := effectstest.CastCatalogSpell(t, g, "Sling-Gang Lieutenant", "Creature — Goblin", gapSlingGangOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.RandomTokenCount(g, "Goblin"); got != 2 {
		t.Fatalf("Goblin tokens = %d, want 2", got)
	}
	life, oppLife := me.Life, opp.Life
	var goblin uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Goblin" {
			goblin = c.InstanceID
			break
		}
	}
	err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
		SacrificeIDs: []uuid.UUID{goblin},
	})
	if err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(goblin) || me.Life != life+1 || opp.Life != oppLife-1 {
		t.Errorf("goblin gone=%v life %d/%d, want %d/%d", !g.Battlefield.Contains(goblin), me.Life, opp.Life, life+1, oppLife-1)
	}
}

func TestGapB49WurmcoilLarvaLeavesTwoWurms(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushDiesCreatureForTest(g, me.ID, "Wurmcoil Larva", gapLarvaOracle, "Artifact Creature — Phyrexian Wurm", 3, 3)
	effectstest.KillCreature(t, g, me.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	var deathtouch, lifelink bool
	for _, c := range g.Battlefield.Cards {
		if c.Name != "Phyrexian Wurm" || !c.IsToken() {
			continue
		}
		p, tough := effectstest.EffectivePower(t, g, c.InstanceID), effectstest.EffectiveToughness(t, g, c.InstanceID)
		switch {
		case p == 1 && tough == 2 && effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "deathtouch"):
			deathtouch = true
		case p == 2 && tough == 1 && effectstest.HasEffectiveKeyword(t, g, c.InstanceID, "lifelink"):
			lifelink = true
		}
	}
	if !deathtouch || !lifelink {
		t.Errorf("deathtouch 1/2 = %v, lifelink 2/1 = %v", deathtouch, lifelink)
	}
}

func TestGapB49EmperorExilesAtCombatThenReturnsOnAdaptAndSacrifices(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.PushCatalogPermanent(g, me.ID, "Emperor of Bones", "Creature — Skeleton Noble", gapEmperorOracle, false)
	corpse := effectstest.PushGraveyardPermanent(me, "Corpse", "Creature — Zombie", "")
	for g.Turn.Step != game.StepBeginCombat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoicePickTarget, me.ID)
	if pick == nil {
		t.Fatal("no target prompt for the begin-combat trigger")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: corpse}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(corpse) {
		t.Fatal("the targeted card was not exiled")
	}
	gapFloatB(t, g, me, 2)
	if err := gapB49Activate(t, g, me.ID, id, 0); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := gapB49Counters(g, id); got != 2 {
		t.Fatalf("counters = %d, want 2", got)
	}
	var back uuid.UUID
	var c game.Card
	for _, bc := range g.Battlefield.Cards {
		if bc.Name == "Corpse" {
			back, c = bc.InstanceID, bc
		}
	}
	if back == uuid.Nil {
		t.Fatal("the exiled creature did not come back")
	}
	if c.Controller != me.ID || c.Counters[game.CounterFinality] != 1 {
		t.Errorf("controller %v finality %d", c.Controller, c.Counters[game.CounterFinality])
	}
	if !effectstest.HasEffectiveKeyword(t, g, back, "haste") {
		t.Error("it has no haste")
	}
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(back) {
		t.Error("it was not sacrificed at the beginning of the end step")
	}
}

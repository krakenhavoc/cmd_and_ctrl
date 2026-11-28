package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b742_test.go — Exodus, S82 OLD batch 742 (#5626).

const (
	oldB742Survival  = "119d719d-e965-45b4-9bc9-ac03211b10c2"
	oldB742Druid     = "1f3fb367-60b5-4b36-9604-3064afdac0d9"
	oldB742KeeperMnd = "45393fba-b1cc-4db7-b7e0-3851bf421a65"
	oldB742Necro     = "925f11c3-fd32-4aa7-a130-6554b81ad3eb"
	oldB742Scholars  = "da4a5129-b005-4fc6-b175-0416627881db"
	oldB742Pegasus   = "16b4a48a-fec2-433c-948d-e0b16c961172"
	oldB742Tide      = "2fbf95b4-bcf4-4b5e-b5dc-0294f2b48d3e"
	oldB742Resus     = "4390e251-e154-4ef8-9929-0c78aa2c8f10"
	oldB742Cataclysm = "b0479ca4-e925-4aa7-a4fa-f7e01f899ac5"
	oldB742FadeAway  = "f92c2d47-f2f7-43d9-acf3-0145bad1977d"
)

func TestOldB742CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB742Survival: "Survival of the Fittest", oldB742Druid: "Avenging Druid", oldB742KeeperMnd: "Keeper of the Mind",
		oldB742Necro: "Necrologia", oldB742Scholars: "Oath of Scholars", oldB742Pegasus: "Pegasus Stampede",
		oldB742Tide: "Aether Tide", oldB742Resus: "Resuscitate", oldB742Cataclysm: "Cataclysm", oldB742FadeAway: "Fade Away",
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

func oldB742Main(t *testing.T, g *game.Game) *game.Player {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	return g.Seats[g.Turn.ActiveSeat]
}

func TestOldB742SurvivalOfTheFittestDiscardsACreatureToTutor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB742Main(t, g)
	surv := effectstest.PushCatalogPermanent(g, me.ID, "Survival of the Fittest", "Enchantment", oldB742Survival, false)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	bear := effectstest.PushCatalogHandCard(me, "Bear", "Creature — Bear", "")
	bolt := effectstest.PushCatalogHandCard(me, "Bolt", "Instant", "")
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Wolf", TypeLine: "Creature — Wolf"},
		game.Card{Name: "Elk", TypeLine: "Creature — Elk"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
	)
	wolf := ids[0]
	if err := g.ActivateCatalogAbility(me.ID, surv, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("activated {G} with an empty mana pool")
	}
	effectstest.DcMana(t, g, me.ID, "{G}")
	if err := g.ActivateCatalogAbility(me.ID, surv, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{bolt}}); err == nil {
		t.Fatal("an instant paid a \"discard a creature card\" cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, surv, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("the discard cost was not paid at announce")
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil || len(c.SearchCards) != 2 {
		t.Fatalf("search prompt %+v, want the two creature cards", c)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{wolf}); err != nil {
		t.Fatalf("ResolveSearchLibrary: %v", err)
	}
	if !me.Hand.Contains(wolf) {
		t.Error("the tutored creature did not reach the hand")
	}
	// It is usable again the same turn (no {T}).
	effectstest.DcMana(t, g, me.ID, "{G}")
	again := effectstest.PushCatalogHandCard(me, "Bear 2", "Creature — Bear", "")
	if err := g.ActivateCatalogAbility(me.ID, surv, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{again}}); err != nil {
		t.Fatalf("second activation: %v", err)
	}
}

func oldB742Druid742(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := oldB742Main(t, g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	druid := effectstest.PushCatalogPermanent(g, me.ID, "Avenging Druid", "Creature — Human Druid", oldB742Druid, false)
	return g, me, opp, druid
}

func TestOldB742AvengingDruidRevealsToALandAndBinsTheRest(t *testing.T) {
	g, me, opp, druid := oldB742Druid742(t)
	// Top to bottom: Spell, Spell 2, Forest, Beyond.
	beyond := effectstest.PutOnTopOfLibrary(g, me, "Beyond", "Creature — Bear", "")
	forest := effectstest.PutOnTopOfLibrary(g, me, "Forest", "Basic Land — Forest", "")
	spell2 := effectstest.PutOnTopOfLibrary(g, me, "Spell 2", "Sorcery", "")
	spell := effectstest.PutOnTopOfLibrary(g, me, "Spell", "Instant", "")
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(druid, opp.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	if !g.Battlefield.Contains(forest) {
		t.Error("the land card did not enter the battlefield")
	}
	for _, id := range []uuid.UUID{spell, spell2} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("a nonland revealed card %v is not in the graveyard", id)
		}
	}
	if z := g.FindCardZoneForEffect(beyond); z == nil || z.Kind != game.ZoneLibrary {
		t.Errorf("the card past the land is in %+v, want the library", z)
	}
}

func TestOldB742AvengingDruidMayDeclineAndIgnoresDamageToYou(t *testing.T) {
	g, me, opp, druid := oldB742Druid742(t)
	forest := effectstest.PutOnTopOfLibrary(g, me, "Forest", "Basic Land — Forest", "")
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(druid, me.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceConfirm {
			t.Fatal("damage to its own controller asked about the reveal")
		}
	}
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(druid, opp.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	if z := g.FindCardZoneForEffect(forest); z == nil || z.Kind != game.ZoneLibrary {
		t.Errorf("the top card is in %+v after a decline, want the library", z)
	}
}

func TestOldB742AvengingDruidWithNoLandBinsTheWholeLibrary(t *testing.T) {
	g, me, opp, druid := oldB742Druid742(t)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	a := effectstest.PutOnTopOfLibrary(g, me, "A", "Instant", "")
	b := effectstest.PutOnTopOfLibrary(g, me, "B", "Sorcery", "")
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(druid, opp.ID, 1); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	if !me.Graveyard.Contains(a) || !me.Graveyard.Contains(b) {
		t.Error("the revealed library did not all reach the graveyard")
	}
}

func TestOldB742KeeperOfTheMindNeedsAnOpponentTwoCardsAhead(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB742Main(t, g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	keeper := effectstest.PushCatalogPermanent(g, me.ID, "Keeper of the Mind", "Creature — Human Wizard", oldB742KeeperMnd, false)
	g.WithWriteLock(func() { me.Hand.Cards = nil; opp.Hand.Cards = nil })
	effectstest.PushCatalogHandCard(me, "Mine", "Instant", "")
	for i := 0; i < 2; i++ {
		effectstest.PushCatalogHandCard(opp, "Theirs", "Instant", "")
	}
	target := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	effectstest.DcMana(t, g, me.ID, "{U}")
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err == nil {
		t.Fatal("an opponent with only one more card was accepted")
	}
	effectstest.PushCatalogHandCard(opp, "Theirs 3", "Instant", "")
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}}); err == nil {
		t.Fatal("targeting yourself was accepted")
	}
	before := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, keeper, 0, game.ActivateAbilityParams{Strict: true, Targets: target}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	// The hand sizes are read as the ability is activated: the draw happens
	// though the opponent's hand shrinks in response.
	g.WithWriteLock(func() { opp.Hand.Cards = nil })
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size() - before; got != 1 {
		t.Errorf("drew %d cards, want 1", got)
	}
}

func oldB742AtEndStep(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	for i := 0; !(g.Turn.Step == game.StepEnd && g.Turn.ActiveSeat == seat); i++ {
		if i > 300 {
			t.Fatalf("never reached seat %d's end step", seat)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func TestOldB742NecrologiaCastsOnlyInYourEndStepAndDrawsX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB742Main(t, g)
	id := effectstest.HandCardOf(g, me, "Necrologia", "Instant", "{3}{B}{B}", oldB742Necro)
	effectstest.DcMana(t, g, me.ID, "{B}{B}{B}{B}{B}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2}); err == nil {
		t.Fatal("cast in the main phase")
	}
	oldB742AtEndStep(t, g, g.Turn.ActiveSeat)
	effectstest.DcMana(t, g, me.ID, "{B}{B}{B}{B}{B}")
	life, hand := me.Life, me.Hand.Size()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 3}); err != nil {
		t.Fatalf("cast in the end step: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := life - me.Life; got != 3 {
		t.Errorf("paid %d life, want 3", got)
	}
	if got := me.Hand.Size() - (hand - 1); got != 3 {
		t.Errorf("drew %d cards, want 3", got)
	}
}

func TestOldB742NecrologiaCantBeCastInAnOpponentsEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	oldB742Main(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := (g.Turn.ActiveSeat + 1) % len(g.Seats)
	id := effectstest.HandCardOf(g, me, "Necrologia", "Instant", "{3}{B}{B}", oldB742Necro)
	oldB742AtEndStep(t, g, opp)
	effectstest.DcMana(t, g, me.ID, "{B}{B}{B}{B}{B}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 1}); err == nil {
		t.Fatal("cast in an opponent's end step")
	}
}

func TestOldB742OathOfScholarsUpkeepPlayerMayDiscardAndDrawThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, upkeep, poor, rich := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	effectstest.PushPermanentForTest(g, owner.ID, "Oath of Scholars", oldB742Scholars, "Enchantment")
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			p.Hand.Cards = nil
		}
	})
	for _, n := range []struct {
		p *game.Player
		n int
	}{{owner, 5}, {upkeep, 2}, {poor, 1}, {rich, 4}} {
		for i := 0; i < n.n; i++ {
			effectstest.PushCatalogHandCard(n.p, "Card", "Instant", "")
		}
	}
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	ch := effectstest.PickTargetPrompt(t, g)
	if ch.Chooser != upkeep.ID {
		t.Fatalf("the prompt is for %s, want the upkeep player", ch.Chooser)
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range ch.PickTargetPlayers {
		offered[id] = true
	}
	if !offered[owner.ID] || !offered[rich.ID] || offered[upkeep.ID] || offered[poor.ID] || len(offered) != 2 {
		t.Fatalf("offered %v, want only the opponents with more cards in hand than the upkeep player", ch.PickTargetPlayers)
	}
	if err := g.ResolvePickTarget(ch.ID, upkeep.ID, game.TargetRef{Kind: game.TargetPlayer, ID: poor.ID}); err == nil {
		t.Fatal("a player with fewer cards was accepted")
	}
	if err := g.ResolvePickTarget(ch.ID, upkeep.ID, game.TargetRef{Kind: game.TargetPlayer, ID: owner.ID}); err != nil {
		t.Fatalf("picking: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, upkeep, true)
	if got := upkeep.Hand.Size(); got != 3 {
		t.Errorf("the upkeep player has %d cards, want 3 (two discarded, three drawn)", got)
	}
	if got := upkeep.Graveyard.Size(); got < 2 {
		t.Errorf("graveyard holds %d cards, want the two discarded", got)
	}
}

func TestOldB742OathOfScholarsDeclinesAndHasNoTargetWithoutABiggerHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner, upkeep := g.Seats[0], g.Seats[1]
	effectstest.PushPermanentForTest(g, owner.ID, "Oath of Scholars", oldB742Scholars, "Enchantment")
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			p.Hand.Cards = nil
		}
	})
	effectstest.PushCatalogHandCard(upkeep, "Card", "Instant", "")
	effectstest.PushCatalogHandCard(owner, "Card", "Instant", "")
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	if p := effectstest.LatestPickTarget(g, upkeep.ID); p != nil {
		t.Fatalf("asked for a target with nobody holding more cards: %+v", p)
	}

	effectstest.PushCatalogHandCard(owner, "Card 2", "Instant", "")
	g2 := effectstest.NewCatalogGame(t)
	owner2, upkeep2 := g2.Seats[0], g2.Seats[1]
	effectstest.PushPermanentForTest(g2, owner2.ID, "Oath of Scholars", oldB742Scholars, "Enchantment")
	g2.WithWriteLock(func() {
		for _, p := range g2.Seats {
			p.Hand.Cards = nil
		}
	})
	effectstest.PushCatalogHandCard(upkeep2, "Card", "Instant", "")
	effectstest.PushCatalogHandCard(owner2, "Card", "Instant", "")
	effectstest.PushCatalogHandCard(owner2, "Card 2", "Instant", "")
	effectstest.AdvanceToUpkeepOf(t, g2, 1)
	ch := effectstest.PickTargetPrompt(t, g2)
	if err := g2.ResolvePickTarget(ch.ID, upkeep2.ID, game.TargetRef{Kind: game.TargetPlayer, ID: owner2.ID}); err != nil {
		t.Fatalf("picking: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerConfirmFor(t, g2, upkeep2, false)
	if got := upkeep2.Hand.Size(); got != 1 {
		t.Errorf("the upkeep player has %d cards after declining, want 1", got)
	}
}

func TestOldB742PegasusStampedeMakesAPegasusAndBuysBackForALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	land := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Plains", TypeLine: "Basic Land — Plains", Owner: me.ID, Controller: me.ID,
	})
	id, err := effectstest.CastWithOptionalCosts(t, g, "Pegasus Stampede", "Sorcery", oldB742Pegasus, nil, []int{0}, []uuid.UUID{land})
	if err != nil {
		t.Fatalf("CastSpell with buyback: %v", err)
	}
	if g.Battlefield.Contains(land) {
		t.Error("the land was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(id) {
		t.Error("the bought-back Stampede is not in hand")
	}
	var pegasus *game.Card
	for i := range g.Battlefield.Cards {
		if c := g.Battlefield.Cards[i]; c.Name == "Pegasus" && c.Controller == me.ID {
			pegasus = &g.Battlefield.Cards[i]
		}
	}
	if pegasus == nil || pegasus.Power != 1 || pegasus.Toughness != 1 || !effectstest.HasKeyword(pegasus.Keywords, "flying") {
		t.Fatalf("no 1/1 flying Pegasus token: %+v", pegasus)
	}
	// Without the buyback it goes to the graveyard.
	id2, err := effectstest.CastWithOptionalCosts(t, g, "Pegasus Stampede", "Sorcery", oldB742Pegasus, nil, nil, nil)
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id2) {
		t.Error("an unbought Stampede did not reach the graveyard")
	}
	// An unpayable buyback is refused.
	if _, err := effectstest.CastWithOptionalCosts(t, g, "Pegasus Stampede", "Sorcery", oldB742Pegasus, nil, []int{0}, nil); err == nil {
		t.Error("a buyback with no land named was accepted")
	}
}

func TestOldB742AetherTideDiscardsXCreaturesAndBouncesXCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB742Main(t, g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	third := effectstest.B12Creature(g, opp.ID, "Third", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { me.Hand.Cards = nil })
	c1 := effectstest.HandCard(me, "Hand Bear", "Creature — Bear")
	c2 := effectstest.HandCard(me, "Hand Elk", "Creature — Elk")
	spell := effectstest.HandCardOf(g, me, "Aether Tide", "Sorcery", "{X}{U}", oldB742Tide)
	nonc := effectstest.HandCard(me, "Bolt", "Instant")
	two := effectstest.CardRefs(mine, theirs)
	for name, params := range map[string]game.CastSpellParams{
		"a noncreature discard":      {XValue: 2, DiscardIDs: []uuid.UUID{c1, nonc}, Targets: two},
		"one target at X=2":          {XValue: 2, DiscardIDs: []uuid.UUID{c1, c2}, Targets: two[:1]},
		"one discard at X=2":         {XValue: 2, DiscardIDs: []uuid.UUID{c1}, Targets: two},
		"three targets at X=2":       {XValue: 2, DiscardIDs: []uuid.UUID{c1, c2}, Targets: effectstest.CardRefs(mine, theirs, third)},
		"a non-creature as a target": {XValue: 2, DiscardIDs: []uuid.UUID{c1, c2}, Targets: effectstest.CardRefs(mine, spell)},
	} {
		params.Strict = true
		if err := g.CastSpell(me.ID, spell, params); err == nil {
			t.Fatalf("%s: the cast was accepted", name)
		}
	}
	effectstest.DcMana(t, g, me.ID, "{C}{C}{U}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, XValue: 2, DiscardIDs: []uuid.UUID{c1, c2}, Targets: two}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	if !me.Graveyard.Contains(c1) || !me.Graveyard.Contains(c2) {
		t.Error("the creature cards were not discarded as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(mine) || g.Battlefield.Contains(theirs) {
		t.Error("a target stayed on the battlefield")
	}
	if !me.Hand.Contains(mine) || !opp.Hand.Contains(theirs) {
		t.Error("a target did not return to its owner's hand")
	}
	if !g.Battlefield.Contains(third) {
		t.Error("an untargeted creature left the battlefield")
	}
}

func TestOldB742ResuscitateGivesEachCreatureYouControlARegenerateAbilityForTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB742Main(t, g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	id := effectstest.HandCardOf(g, me, "Resuscitate", "Instant", "{1}{G}", oldB742Resus)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DcMana(t, g, me.ID, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, theirs, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("an opponent's creature was given the ability")
	}
	if err := g.ActivateCatalogAbility(me.ID, bear, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the granted ability: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(bear); err != nil {
			t.Fatal(err)
		}
	})
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the regenerating creature was destroyed")
	}
	if !effectstest.AlliTapped(t, g, bear) {
		t.Error("regeneration did not tap it")
	}
}

func TestOldB742CataclysmKeepsOneOfEachTypeAndSacrificesTheRest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB742Main(t, g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	bear1 := effectstest.B12Creature(g, me.ID, "Bear 1", "Creature — Bear", 2, 2)
	bear2 := effectstest.B12Creature(g, me.ID, "Bear 2", "Creature — Bear", 2, 2)
	forest := effectstest.AcPermanent(g, me.ID, "Forest", "Basic Land — Forest", 0, 0)
	island := effectstest.AcPermanent(g, me.ID, "Island", "Basic Land — Island", 0, 0)
	rock := effectstest.AcPermanent(g, me.ID, "Rock", "Artifact", 0, 0)
	aura := effectstest.AcPermanent(g, me.ID, "Aura", "Enchantment", 0, 0)
	walker := effectstest.AcPermanent(g, me.ID, "Jace", "Legendary Planeswalker — Jace", 0, 0)
	// An opponent whose board is already one of each of the types it has needs no choice.
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	theirLand := effectstest.AcPermanent(g, opp.ID, "Swamp", "Basic Land — Swamp", 0, 0)
	id := effectstest.HandCardOf(g, me, "Cataclysm", "Sorcery", "{2}{W}{W}", oldB742Cataclysm)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("the caster was not asked what to keep")
	}
	for _, c := range prompt.ChooseCards {
		if c == walker {
			t.Error("a planeswalker was offered")
		}
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{bear1, bear2, forest, island}); err == nil {
		t.Fatal("a keep set with no artifact or enchantment was accepted")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, bear1, forest, rock, aura)
	effectstest.PassPriorityAroundTable(t, g)
	for name, id := range map[string]uuid.UUID{"Bear 1": bear1, "Forest": forest, "Rock": rock, "Aura": aura, "Their Bear": theirBear, "Swamp": theirLand} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s was sacrificed", name)
		}
	}
	for name, id := range map[string]uuid.UUID{"Bear 2": bear2, "Island": island, "Jace": walker} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s was kept", name)
		}
	}
}

func TestOldB742FadeAwayAsksOnceForEachCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := oldB742Main(t, g)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.B12Creature(g, me.ID, "Bear 1", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Bear 2", "Creature — Bear", 2, 2)
	rock := effectstest.AcPermanent(g, me.ID, "Rock", "Artifact", 0, 0)
	theirBear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	id := effectstest.HandCardOf(g, me, "Fade Away", "Sorcery", "{2}{U}", oldB742FadeAway)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The caster pays for the first creature and declines for the second.
	effectstest.DcMana(t, g, me.ID, "{C}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.AnswerSacrifice(t, g, me.ID, rock)
	// The opponent declines for their one creature and sacrifices.
	effectstest.AnswerPayUnless(t, g, opp.ID, false)
	effectstest.AnswerSacrifice(t, g, opp.ID, theirBear)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("the caster's chosen permanent was not sacrificed")
	}
	if g.Battlefield.Contains(theirBear) {
		t.Error("the opponent's only permanent was not sacrificed")
	}
	if got := effectstest.CountOnBattlefield(g, "Bear 1", me.ID) + effectstest.CountOnBattlefield(g, "Bear 2", me.ID); got != 2 {
		t.Errorf("the caster has %d of their bears, want both (one decline, one sacrifice of the Rock)", got)
	}
	if effectstest.HasPayUnlessFor(g, me.ID) || effectstest.HasPayUnlessFor(g, opp.ID) {
		t.Error("a pay-unless prompt is still open")
	}
}

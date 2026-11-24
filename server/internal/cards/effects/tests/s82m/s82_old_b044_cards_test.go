package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b044_cards_test.go — Strixhaven: School of Mages, batch 044 (S82,
// #5502). Each test plays the real catalog entry through the engine.

const (
	oldB44ExcavationOracle  = "0ed2a07c-a7a5-4910-a63c-e72803947afb"
	oldB44DustspeakerOracle = "370974e7-d242-4f24-a031-cba71c935306"
	oldB44PastcallerOracle  = "58b9f933-7b13-4335-996c-4722ab9e08b2"
	oldB44HistorianOracle   = "5a131407-c217-48fc-aa15-9e3392445768"
	oldB44SilverquillOracle = "6075b999-75c5-41c8-898d-ee612ba79a92"
	oldB44WardenOracle      = "6ee2a7ba-a676-4d05-b214-7db52e41afe7"
	oldB44ScrollOracle      = "8f5ef838-839a-4eb0-8d65-c8c8def0a233"
	oldB44HelixOracle       = "9c83979a-b354-4e20-9638-85e8f354da81"
	oldB44SpiritOracle      = "9fbab5bc-c651-46dc-89b6-0bcfeadb3891"
	oldB44ConfrontOracle    = "a92c1ffe-88b0-467c-883c-6ca3564a39e3"
	oldB44TrudgeOracle      = "c775d3ef-2184-4425-8c8e-2bbc49687f76"
	oldB44BookwurmOracle    = "ce77fff1-32f8-4a55-9124-b1dafe95840d"
	oldB44EfreetOracle      = "fb5e8b8c-bad2-45bb-abb4-2ed454525749"
	oldB44BoltOracle        = "4457ed35-7c10-48c8-9776-456485fdf070"
	oldB44OptOracle         = "713332c1-5bd8-400f-bfff-c1ca0697a043"
)

func TestOldB44EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		oldB44ExcavationOracle:  "Lorehold Excavation",
		oldB44DustspeakerOracle: "Ardent Dustspeaker",
		oldB44PastcallerOracle:  "Returned Pastcaller",
		oldB44HistorianOracle:   "Illustrious Historian",
		oldB44SilverquillOracle: "Silverquill Command",
		oldB44WardenOracle:      "Pillardrop Warden",
		oldB44ScrollOracle:      "Radiant Scrollwielder",
		oldB44HelixOracle:       "Aether Helix",
		oldB44SpiritOracle:      "Stonerise Spirit",
		oldB44ConfrontOracle:    "Confront the Past",
		oldB44TrudgeOracle:      "Brackish Trudge",
		oldB44BookwurmOracle:    "Bookwurm",
		oldB44EfreetOracle:      "Efreet Flamepainter",
	}
	for oracle, name := range want {
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

func oldB44GraveCard(p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle, Owner: p.ID, Controller: p.ID})
	return id
}

func oldB44Spirits(g *game.Game, owner uuid.UUID) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == "Spirit" && c.Controller == owner {
			out = append(out, c)
		}
	}
	return out
}

func TestOldB44LoreholdExcavationMillsAtTheEndStep(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine    string
		wantLife, wantOpp int
	}{
		{"land gains life", "Basic Land — Forest", 1, 0},
		{"nonland hurts every opponent", "Sorcery", 0, -1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.PushCatalogPermanent(g, me.ID, "Lorehold Excavation", "Enchantment", oldB44ExcavationOracle, false)
			top := uuid.New()
			me.Library.PushTop(game.Card{InstanceID: top, Name: "Milled", TypeLine: tc.typeLine, Owner: me.ID, Controller: me.ID})
			life := me.Life
			opps := map[uuid.UUID]int{}
			for _, s := range g.Seats {
				opps[s.ID] = s.Life
			}
			effectstest.AdvanceTo(t, g, game.StepEnd)
			effectstest.PassPriorityAroundTable(t, g)
			if !me.Graveyard.Contains(top) {
				t.Fatal("the top card was not milled")
			}
			if me.Life != life+tc.wantLife {
				t.Errorf("life %d, want %d", me.Life, life+tc.wantLife)
			}
			for _, s := range g.Seats {
				if s.ID == me.ID {
					continue
				}
				if s.Life != opps[s.ID]+tc.wantOpp {
					t.Errorf("%s at %d, want %d", s.Name, s.Life, opps[s.ID]+tc.wantOpp)
				}
			}
		})
	}
}

func TestOldB44LoreholdExcavationMakesATappedSpiritForAnExiledCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	ench := effectstest.PushCatalogPermanent(g, me.ID, "Lorehold Excavation", "Enchantment", oldB44ExcavationOracle, false)
	creature := oldB44GraveCard(me, "Bear", "Creature — Bear", "{1}{G}", "")
	spell := oldB44GraveCard(me, "Shock", "Instant", "{R}", "")
	// A noncreature card is not a legal payment.
	effectstest.FloatForTest(g, me, "CCCCC")
	err := g.ActivateCatalogAbility(me.ID, ench, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{spell}})
	if err == nil {
		t.Fatal("exiled a noncreature card to a cost that wants a creature card")
	}
	effectstest.ApaActivate(t, g, me, ench, 0, game.ActivateAbilityParams{ExileIDs: []uuid.UUID{creature}})
	if !g.Exile.Contains(creature) {
		t.Error("the creature card was not exiled")
	}
	sp := oldB44Spirits(g, me.ID)
	if len(sp) != 1 {
		t.Fatalf("%d Spirits, want 1", len(sp))
	}
	if !sp[0].Tapped || sp[0].CurrentPower() != 3 || sp[0].CurrentToughness() != 2 {
		t.Errorf("Spirit tapped %v %d/%d, want a tapped 3/2", sp[0].Tapped, sp[0].CurrentPower(), sp[0].CurrentToughness())
	}
}

func TestOldB44IllustriousHistorianWorksOnlyFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	card := oldB44GraveCard(me, "Illustrious Historian", "Creature — Human Shaman", "{1}{R}", oldB44HistorianOracle)
	effectstest.FloatForTest(g, me, "CCCCC")
	effectstest.ApaActivate(t, g, me, card, 0, game.ActivateAbilityParams{})
	if !g.Exile.Contains(card) {
		t.Error("the Historian was not exiled as the cost")
	}
	sp := oldB44Spirits(g, me.ID)
	if len(sp) != 1 || !sp[0].Tapped {
		t.Fatalf("Spirits %d (tapped %v), want one tapped Spirit", len(sp), len(sp) == 1 && sp[0].Tapped)
	}
}

func TestOldB44StoneriseSpiritGrantsFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	spirit := effectstest.PushCatalogPermanent(g, me.ID, "Stonerise Spirit", "Creature — Spirit Bird", oldB44SpiritOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	pitch := oldB44GraveCard(me, "Pitch", "Sorcery", "{1}", "")
	effectstest.FloatForTest(g, me, "CCCC")
	effectstest.ApaActivate(t, g, me, spirit, 0, game.ActivateAbilityParams{
		ExileIDs: []uuid.UUID{pitch},
		Targets:  []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	if !g.Exile.Contains(pitch) {
		t.Error("the graveyard card was not exiled as the cost")
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the Bear did not gain flying")
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("flying lasts past the turn")
	}
}

func TestOldB44PillardropWardenSacrificesItselfToReturnASpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	warden := effectstest.PushCatalogPermanent(g, me.ID, "Pillardrop Warden", "Creature — Spirit Dwarf", oldB44WardenOracle, false)
	bolt := oldB44GraveCard(me, "Shock", "Instant", "{R}", "")
	bear := oldB44GraveCard(me, "Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.FloatForTest(g, me, "CC")
	err := g.ActivateCatalogAbility(me.ID, warden, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}}})
	if err == nil {
		t.Fatal("a creature card was a legal target")
	}
	effectstest.ApaActivate(t, g, me, warden, 0, game.ActivateAbilityParams{Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bolt}}})
	if _, ok := effectstest.BattlefieldCard(g, warden); ok {
		t.Error("the Warden is still on the battlefield")
	}
	if !effectstest.InHandOf(me, bolt) {
		t.Error("the instant did not return to the hand")
	}
}

func TestOldB44AetherHelixReturnsBothAndSurvivesOneIllegalTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	card := oldB44GraveCard(me, "Old Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.CastCatalogSpell(t, g, "Aether Helix", "Sorcery", oldB44HelixOracle, []game.TargetRef{
		{Kind: game.TargetCard, ID: theirs}, {Kind: game.TargetCard, ID: card}})
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, theirs); ok {
		t.Error("the permanent is still on the battlefield")
	}
	if !effectstest.InHandOf(opp, theirs) {
		t.Error("the permanent did not go to its owner's hand")
	}
	if !effectstest.InHandOf(me, card) {
		t.Error("the graveyard card did not return")
	}
}

func TestOldB44ReturnedPastcallerTargetsASpiritInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spirit := oldB44GraveCard(me, "Ghost", "Creature — Spirit", "{1}{W}", "")
	spell := oldB44GraveCard(me, "Shock", "Instant", "{R}", "")
	bear := oldB44GraveCard(me, "Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.CastHoldCreature(t, g, "Returned Pastcaller", "Creature — Spirit Cleric", oldB44PastcallerOracle, 4, 2)
	c := effectstest.PickTargetPrompt(t, g)
	legal := map[uuid.UUID]bool{}
	for _, id := range c.PickTargetCards {
		legal[id] = true
	}
	if !legal[spirit] || !legal[spell] || legal[bear] {
		t.Fatalf("legal targets spirit %v spell %v bear %v, want true true false", legal[spirit], legal[spell], legal[bear])
	}
	effectstest.AnswerPickTarget(t, g, spirit)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InHandOf(me, spirit) {
		t.Error("the Spirit card did not return to the hand")
	}
}

func TestOldB44BrackishTrudgeEntersTappedAndReturnsOnlyAfterGainingLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	trudge := effectstest.CastHoldCreature(t, g, "Brackish Trudge", "Creature — Fungus Beast", oldB44TrudgeOracle, 4, 2)
	if c, ok := effectstest.BattlefieldCard(g, trudge); !ok || !c.Tapped {
		t.Fatalf("the Trudge on battlefield %v tapped %v, want a tapped one", ok, c.Tapped)
	}
	card := oldB44GraveCard(me, "Brackish Trudge", "Creature — Fungus Beast", "{2}{B}", oldB44TrudgeOracle)
	effectstest.FloatForTest(g, me, "CB")
	if err := g.ActivateCatalogAbility(me.ID, card, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without having gained life this turn")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, me.ID, 1) })
	effectstest.ApaActivate(t, g, me, card, 0, game.ActivateAbilityParams{})
	if !effectstest.InHandOf(me, card) {
		t.Error("the Trudge did not return to the hand")
	}
}

func TestOldB44BookwurmGainsDrawsAndTucksItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E")
	effectstest.AdvanceToMain(t, g)
	life, hand := me.Life, me.Hand.Size()
	effectstest.CastHoldCreature(t, g, "Bookwurm", "Creature — Wurm", oldB44BookwurmOracle, 7, 7)
	if me.Life != life+3 || me.Hand.Size() != hand+1 {
		t.Errorf("life %d (want %d), hand %d (want %d)", me.Life, life+3, me.Hand.Size(), hand+1)
	}
	card := oldB44GraveCard(me, "Bookwurm", "Creature — Wurm", "{7}{G}", oldB44BookwurmOracle)
	effectstest.FloatForTest(g, me, "CCG")
	effectstest.ApaActivate(t, g, me, card, 0, game.ActivateAbilityParams{})
	n := len(me.Library.Cards)
	if n < 3 {
		t.Fatalf("library %d cards", n)
	}
	// Index 0 is the bottom, so the third card from the top is at n-3.
	if me.Library.Cards[n-3].InstanceID != card {
		t.Errorf("the Bookwurm is not third from the top of the library")
	}
}

func TestOldB44SilverquillCommandBulletPairs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	fodder := effectstest.B12Creature(g, opp.ID, "Their Only Creature", "Creature — Bear", 2, 2)
	back := oldB44GraveCard(me, "Cheap", "Creature — Bear", "{1}", "")
	costly := oldB44GraveCard(me, "Costly", "Creature — Bear", "{3}", "")
	hand, life := me.Hand.Size(), me.Life
	effectstest.P1g3CastModal(t, g, "Silverquill Command", "Sorcery", oldB44SilverquillOracle, []int{0, 1},
		game.TargetRef{Kind: game.TargetCard, ID: bear, Mode: 0},
		game.TargetRef{Kind: game.TargetCard, ID: back, Mode: 1})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.EffectivePower(t, g, bear); got != 5 {
		t.Errorf("the Bear's power %d, want 5", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "flying") {
		t.Error("the Bear did not gain flying")
	}
	if _, ok := effectstest.BattlefieldCard(g, back); !ok {
		t.Error("the cheap creature card did not return to the battlefield")
	}
	effectstest.P1g3CastModal(t, g, "Silverquill Command", "Sorcery", oldB44SilverquillOracle, []int{2, 3},
		game.TargetRef{Kind: game.TargetPlayer, ID: me.ID, Mode: 0},
		game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID, Mode: 1})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, opp.ID, fodder)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Errorf("hand %d (want %d), life %d (want %d)", me.Hand.Size(), hand+1, me.Life, life-1)
	}
	if _, ok := effectstest.BattlefieldCard(g, fodder); ok {
		t.Error("the opponent did not sacrifice their creature")
	}
	if _, ok := effectstest.BattlefieldCard(g, costly); ok {
		t.Error("a creature card with mana value 3 returned")
	}
}

func TestOldB44SilverquillCommandRefusesMana3Reanimation(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	costly := oldB44GraveCard(me, "Costly", "Creature — Bear", "{3}", "")
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Silverquill Command", TypeLine: "Sorcery", OracleID: oldB44SilverquillOracle, Owner: me.ID, Controller: me.ID})
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{0, 1}, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Mode: 0}, {Kind: game.TargetCard, ID: costly, Mode: 1}}})
	if err == nil {
		t.Error("a creature card with mana value 3 was a legal target")
	}
}

func oldB44Walker(g *game.Game, owner uuid.UUID, loyalty int) uuid.UUID {
	return effectstest.P1g2Planeswalker(g, owner, loyalty)
}

func TestOldB44ConfrontThePastBothModes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	theirs := oldB44Walker(g, opp.ID, 5)
	effectstest.FloatForTest(g, me, "CCB")
	effectstest.CastFromHandForTest(t, g, me, "Confront the Past", "Sorcery — Lesson", "{X}{B}", oldB44ConfrontOracle,
		game.CastSpellParams{Strict: true, XValue: 2, Modes: []int{1}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: theirs, Mode: 0}}})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, theirs, game.CounterLoyalty); got != 1 {
		t.Errorf("loyalty %d, want 5 - 2*2 = 1", got)
	}
}

func TestOldB44ConfrontThePastReturnsAWalkerUpToX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	walker := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: walker, Name: "Walker", TypeLine: "Legendary Planeswalker — Test", ManaCost: "{2}{B}",
		Owner: me.ID, Controller: me.ID, StartingLoyalty: 3})
	effectstest.FloatForTest(g, me, "CCCCB")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Confront the Past", TypeLine: "Sorcery — Lesson", ManaCost: "{X}{B}", OracleID: oldB44ConfrontOracle, Owner: me.ID, Controller: me.ID})
	// X = 2 is below the walker's mana value 3: not a legal target.
	err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 2, Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: walker, Mode: 0}}})
	if err == nil {
		t.Fatal("a mana value 3 planeswalker was a legal target at X = 2")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 3, Modes: []int{0}, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: walker, Mode: 0}}}); err != nil {
		t.Fatalf("X = 3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, walker); !ok {
		t.Error("the planeswalker did not return to the battlefield")
	}
}

func TestOldB44ArdentDustspeakerBottomsASpellAndExilesTwoToPlay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	dust := effectstest.PushCatalogPermanent(g, me.ID, "Ardent Dustspeaker", "Creature — Minotaur Shaman", oldB44DustspeakerOracle, false)
	spell := oldB44GraveCard(me, "Shock", "Instant", "{R}", "")
	oldB44GraveCard(me, "Bear", "Creature — Bear", "{1}{G}", "")
	a, b := uuid.New(), uuid.New()
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Third", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: b, Name: "B", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: a, Name: "A", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.Pr7bAttack(t, g, g.Seats[1].ID, dust)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no pick offered")
	}
	effectstest.AnswerChooseCards(t, g, me.ID, spell)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Library.Cards[0].InstanceID != spell {
		t.Error("the instant is not on the bottom of the library")
	}
	if !g.Exile.Contains(a) || !g.Exile.Contains(b) {
		t.Error("the top two cards were not exiled")
	}
}

func TestOldB44ArdentDustspeakerDeclinedDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	dust := effectstest.PushCatalogPermanent(g, me.ID, "Ardent Dustspeaker", "Creature — Minotaur Shaman", oldB44DustspeakerOracle, false)
	spell := oldB44GraveCard(me, "Shock", "Instant", "{R}", "")
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Top", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	effectstest.Pr7bAttack(t, g, g.Seats[1].ID, dust)
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(spell) || g.Exile.Contains(top) {
		t.Error("declining the pick still did something")
	}
}

func TestOldB44ScrollwielderGivesSpellsLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Radiant Scrollwielder", "Creature — Dwarf Cleric", oldB44ScrollOracle, false)
	life, oppLife := me.Life, opp.Life
	effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", oldB44BoltOracle, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-3 {
		t.Fatalf("the opponent is at %d, want %d", opp.Life, oppLife-3)
	}
	if me.Life != life+3 {
		t.Errorf("you are at %d, want %d (lifelink on your instant)", me.Life, life+3)
	}
}

func TestOldB44ScrollwielderUpkeepExilesARandomSpellYouMayCast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Radiant Scrollwielder", "Creature — Dwarf Cleric", oldB44ScrollOracle, false)
	opt := oldB44GraveCard(me, "Opt", "Instant", "{U}", oldB44OptOracle)
	bear := oldB44GraveCard(me, "Bear", "Creature — Bear", "{1}{G}", "")
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.AdvanceToUpkeepOfSeat(t, g, g.Turn.ActiveSeat)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(opt) {
		t.Fatal("the only instant or sorcery card was not exiled")
	}
	if !me.Graveyard.Contains(bear) {
		t.Error("a creature card was exiled")
	}
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "U")
	if err := g.CastSpell(me.ID, opt, game.CastSpellParams{Strict: true, FromZone: string(game.ZoneExile), AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast from exile: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryIfAsked(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(opt) || me.Graveyard.Contains(opt) {
		t.Error("the spell cast this way did not return to exile instead of the graveyard")
	}
}

func TestOldB44EfreetFlamepainterCastsAGraveyardSpellFreeAndExilesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D", "E", "F")
	efreet := effectstest.PushCatalogPermanent(g, me.ID, "Efreet Flamepainter", "Creature — Efreet Shaman", oldB44EfreetOracle, false)
	opt := oldB44GraveCard(me, "Opt", "Instant", "{U}", oldB44OptOracle)
	effectstest.Pr7bAttack(t, g, opp.ID, efreet)
	for i := 0; i < 8 && effectstest.LatestPickTarget(g, me.ID) == nil; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if effectstest.LatestPickTarget(g, me.ID) == nil {
		t.Fatal("no target prompt for the combat damage trigger")
	}
	effectstest.AnswerPickTarget(t, g, opt)
	effectstest.PassPriorityAroundTable(t, g)
	var offer *game.PendingChoice
	for _, c := range g.PendingChoices {
		if effectstest.IsKeywordCastPrompt(c) && c.Chooser == me.ID {
			offer = c
		}
	}
	if offer == nil {
		t.Fatal("no free-cast offer")
	}
	if err := g.CastForResolutionChoice(me.ID, offer.ID, opt, game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast the offered card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerScryIfAsked(t, g, me)
	if !g.Exile.Contains(opt) {
		t.Errorf("the spell cast without paying was not exiled (graveyard has it: %v)", me.Graveyard.Contains(opt))
	}
}

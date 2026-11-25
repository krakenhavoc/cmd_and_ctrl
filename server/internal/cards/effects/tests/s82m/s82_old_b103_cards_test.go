package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b103_cards_test.go — S82 OLD batch 103 (#5510, Jumpstart): card draw
// and library, pump and combat tricks, counters. Each test plays the real
// catalog entry.

const (
	oldB103Urge      = "43f5a93b-0f8d-48d2-ab9d-275d44cf88b5"
	oldB103Blind     = "9abeb38c-2243-424c-bfbc-87aa45b6ea4f"
	oldB103Terrarion = "9e7ade56-ab68-42ab-8a6f-bc38957f1563"
	oldB103Wayfinder = "dc858fc1-d97a-4cc5-af31-2c33d9639e1b"
	oldB103Scroll    = "f728c4f1-178f-421b-a713-21f85892f051"
	oldB103Shaman    = "2608f905-a678-4e40-9e73-cf1c42ec880d"
	oldB103Visionary = "3ad5af12-1741-427d-92c4-6a730cd4faa2"
	oldB103Keep      = "6a88784c-d3be-4076-83eb-895cc1b50947"
	oldB103Idol      = "736ce43b-28cd-4b8b-9936-3e3975a56b6a"
	oldB103Settle    = "2f18bb2d-249f-4d7a-a0f7-8a285d365777"
	oldB103Sentry    = "65f8e726-be50-49ee-96fc-d3a2080d7b8c"
	oldB103Broker    = "d17d085f-aeda-4b23-ab80-729d070ae813"
)

func TestOldB103EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB103Urge:      "Aggressive Urge",
		oldB103Blind:     "Blindblast",
		oldB103Terrarion: "Terrarion",
		oldB103Wayfinder: "Silhana Wayfinder",
		oldB103Scroll:    "Scroll of Avacyn",
		oldB103Shaman:    "Rageblood Shaman",
		oldB103Visionary: "Lightning Visionary",
		oldB103Keep:      "Roving Keep",
		oldB103Idol:      "Dragonloft Idol",
		oldB103Settle:    "Settle the Score",
		oldB103Sentry:    "Chamber Sentry",
		oldB103Broker:    "Lawless Broker",
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

func TestOldB103AggressiveUrgePumpsAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Aggressive Urge", "Instant", oldB103Urge, effectstest.CardRefs(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("target is %d/%d, want 3/3", p, tough)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (cast one, drew one)", me.Hand.Size(), hand+1)
	}
}

func TestOldB103BlindblastPingsStopsABlockAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foe := effectstest.B12Push(g, opp.ID, "Foe", "Creature — Bear", "", 3, 3)
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Blindblast", "Instant", oldB103Blind, effectstest.CardRefs(foe))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, foe); got != 1 {
		t.Errorf("the target took %d damage, want 1", got)
	}
	effectstest.AssertRestrictions(t, g, foe, game.CantBlock)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (cast one, drew one)", me.Hand.Size(), hand+1)
	}
}

func TestOldB103TerrarionEntersTappedAndSacrificesForTwoColorsAndADraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.DftABCast(t, g, me, game.Card{Name: "Terrarion", TypeLine: "Artifact", OracleID: oldB103Terrarion})
	effectstest.Top100AssertEnteredTapped(t, g, id, "Terrarion")
	effectstest.UntapForTest(g, id)
	me.ManaPool = game.ManaPool{}
	effectstest.FillPool(me, 2)
	hand := me.Hand.Size()
	effectstest.ActivateManaFor(t, g, me.ID, id, 0, game.ManaAbilityParams{Colors: []string{"W", "B"}})
	if g.Battlefield.Contains(id) {
		t.Error("the Terrarion was not sacrificed")
	}
	if effectstest.PoolCount(me, "W") != 1 || effectstest.PoolCount(me, "B") != 1 {
		t.Errorf("pool W=%d B=%d, want one of each", effectstest.PoolCount(me, "W"), effectstest.PoolCount(me, "B"))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (the Terrarion drew a card)", me.Hand.Size(), hand+1)
	}
}

func TestOldB103SilhanaWayfinderPutsAChosenCreatureOrLandOnTopAndTheRestUnder(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	effectstest.SeedLibraryTop(me, "Deep", "Sorcery", "{1}")
	spell := effectstest.SeedLibraryTop(me, "Spell", "Instant", "{1}")
	bear := effectstest.SeedLibraryTop(me, "Bear", "Creature — Bear", "{1}{G}")
	land := effectstest.SeedLibraryTop(me, "Forest", "Basic Land — Forest", "")
	art := effectstest.SeedLibraryTop(me, "Rock", "Artifact", "{1}")
	effectstest.DftABCast(t, g, me, game.Card{Name: "Silhana Wayfinder", TypeLine: "Creature — Elf Scout", OracleID: oldB103Wayfinder, Power: 2, Toughness: 1})
	p := effectstest.LatestChoiceOfKind(g, game.PendingChoiceChooseCards)
	if p == nil {
		t.Fatal("no choose_cards prompt")
	}
	if effectstest.HasID(p.ChooseCards, art) || effectstest.HasID(p.ChooseCards, spell) {
		t.Errorf("a card that is neither a creature nor a land was offered: %v", p.ChooseCards)
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{bear}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Library.Cards[len(me.Library.Cards)-1].InstanceID; got != bear {
		t.Errorf("top of library is %s, want the chosen creature", got)
	}
	n := len(me.Library.Cards)
	for _, id := range []uuid.UUID{land, art, spell} {
		idx := -1
		for i, c := range me.Library.Cards {
			if c.InstanceID == id {
				idx = i
			}
		}
		if idx < 0 || idx > 2 {
			t.Errorf("a card that was not chosen sits at index %d of %d, want it under the rest", idx, n)
		}
	}
}

func TestOldB103ScrollOfAvacynDrawsAndGainsFiveOnlyWithAnAngel(t *testing.T) {
	for _, angel := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		id := effectstest.PushCatalogPermanent(g, me.ID, "Scroll of Avacyn", "Artifact", oldB103Scroll, false)
		if angel {
			effectstest.B12Push(g, me.ID, "Serra Angel", "Creature — Angel", "", 4, 4)
		}
		life, hand := me.Life, me.Hand.Size()
		effectstest.FillPool(me, 1)
		effectstest.Pr7Activate(t, g, me.ID, id, 0, game.ActivateAbilityParams{})
		if g.Battlefield.Contains(id) {
			t.Error("the Scroll was not sacrificed")
		}
		if me.Hand.Size() != hand+1 {
			t.Errorf("angel=%v: hand %d, want %d", angel, me.Hand.Size(), hand+1)
		}
		want := life
		if angel {
			want += 5
		}
		if me.Life != want {
			t.Errorf("angel=%v: life %d, want %d", angel, me.Life, want)
		}
	}
}

func TestOldB103RagebloodShamanLendsOtherYourMinotaursPlusOneAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	shaman := effectstest.B12Push(g, me.ID, "Rageblood Shaman", "Creature — Minotaur Shaman", oldB103Shaman, 2, 3)
	mine := effectstest.B12Push(g, me.ID, "Minotaur", "Creature — Minotaur", "", 2, 2)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Minotaur", "Creature — Minotaur", "", 2, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, mine); p != 3 || tough != 3 || !effectstest.HasEffectiveKeyword(t, g, mine, "trample") {
		t.Errorf("your other Minotaur is %d/%d trample=%v, want 3/3 trample", p, tough, effectstest.HasEffectiveKeyword(t, g, mine, "trample"))
	}
	if p, _ := effectstest.EffectivePTOf(t, g, shaman); p != 2 {
		t.Errorf("the Shaman pumped itself: power %d", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, shaman, "trample") {
		t.Error("the Shaman lost its own trample")
	}
	for name, id := range map[string]uuid.UUID{"your Bear": bear, "their Minotaur": theirs} {
		if p, _ := effectstest.EffectivePTOf(t, g, id); p != 2 || effectstest.HasEffectiveKeyword(t, g, id, "trample") {
			t.Errorf("%s was affected: power %d", name, p)
		}
	}
}

func TestOldB103LightningVisionaryHasProwess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	v := effectstest.B12Push(g, me.ID, "Lightning Visionary", "Creature — Minotaur Shaman", oldB103Visionary, 2, 1)
	if !effectstest.HasEffectiveKeyword(t, g, v, "prowess") {
		t.Fatal("no prowess")
	}
	effectstest.CastCatalogSpell(t, g, "Shock", "Instant", "", effectstest.AcPlayerRef(g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID))
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, v); p != 3 {
		t.Errorf("power %d after a noncreature spell, want 3", p)
	}
}

func TestOldB103RovingKeepPumpsAndMayAttackThroughDefender(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	keep := effectstest.B12Push(g, me.ID, "Roving Keep", "Artifact Creature — Wall", oldB103Keep, 5, 7)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPool(me, 7)
	if err := g.ActivateCatalogAbility(me.ID, keep, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, keep); p != 7 {
		t.Errorf("power %d, want 7", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, keep, "trample") || !effectstest.HasEffectiveKeyword(t, g, keep, "defender") {
		t.Error("it should gain trample and keep defender")
	}
	effectstest.DeclareAttack(t, g, opp.ID, keep)
	if c, ok := effectstest.BattlefieldCard(g, keep); !ok || c.AttackingTarget != opp.ID {
		t.Errorf("the Keep is not attacking the opponent")
	}
}

func TestOldB103DragonloftIdolBecomesFourByFourFlyingTrampleWithADragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	idol := effectstest.B12Push(g, me.ID, "Dragonloft Idol", "Artifact Creature — Gargoyle", oldB103Idol, 3, 3)
	if p, tough := effectstest.EffectivePTOf(t, g, idol); p != 3 || tough != 3 || effectstest.HasEffectiveKeyword(t, g, idol, "flying") {
		t.Fatalf("no Dragon: %d/%d", p, tough)
	}
	effectstest.B12Push(g, me.ID, "Whelp", "Creature — Dragon", "", 2, 2)
	if p, tough := effectstest.EffectivePTOf(t, g, idol); p != 4 || tough != 4 {
		t.Errorf("with a Dragon the Idol is %d/%d, want 4/4", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, idol, "flying") || !effectstest.HasEffectiveKeyword(t, g, idol, "trample") {
		t.Error("with a Dragon the Idol should have flying and trample")
	}
}

func TestOldB103SettleTheScoreExilesAndCountersAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foe := effectstest.B12Push(g, opp.ID, "Foe", "Creature — Bear", "", 3, 3)
	pw := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Walker", TypeLine: "Legendary Planeswalker — Test", Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterLoyalty: 3}})
	effectstest.CastCatalogSpell(t, g, "Settle the Score", "Sorcery", oldB103Settle, effectstest.CardRefs(foe))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(foe) || !g.Exile.Contains(foe) {
		t.Error("the creature was not exiled")
	}
	if got := effectstest.CounterCount(g, pw, game.CounterLoyalty); got != 5 {
		t.Errorf("loyalty %d, want 5", got)
	}
}

func TestOldB103SettleTheScoreAsksWhichPlaneswalkerWhenThereAreTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foe := effectstest.B12Push(g, opp.ID, "Foe", "Creature — Bear", "", 3, 3)
	a := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Walker A", TypeLine: "Legendary Planeswalker — Test", Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 3}})
	b := effectstest.PushPermanent(g, me.ID, game.Card{Name: "Walker B", TypeLine: "Legendary Planeswalker — Test", Owner: me.ID, Controller: me.ID, Counters: map[string]int{game.CounterLoyalty: 3}})
	theirs := effectstest.PushPermanent(g, opp.ID, game.Card{Name: "Their Walker", TypeLine: "Legendary Planeswalker — Test", Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLoyalty: 3}})
	effectstest.CastCatalogSpell(t, g, "Settle the Score", "Sorcery", oldB103Settle, effectstest.CardRefs(foe))
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestChoiceOfKind(g, game.PendingChoiceChooseCards)
	if p == nil {
		t.Fatal("no planeswalker choice")
	}
	if effectstest.HasID(p.ChooseCards, theirs) {
		t.Error("an opponent's planeswalker was offered")
	}
	if err := g.ResolveChooseCards(p.ID, me.ID, []uuid.UUID{b}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	if effectstest.CounterCount(g, b, game.CounterLoyalty) != 5 || effectstest.CounterCount(g, a, game.CounterLoyalty) != 3 {
		t.Errorf("loyalty A=%d B=%d, want 3 and 5", effectstest.CounterCount(g, a, game.CounterLoyalty), effectstest.CounterCount(g, b, game.CounterLoyalty))
	}
}

func TestOldB103SettleTheScoreWithNoPlaneswalkerJustExiles(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	foe := effectstest.B12Push(g, opp.ID, "Foe", "Creature — Bear", "", 3, 3)
	effectstest.CastCatalogSpell(t, g, "Settle the Score", "Sorcery", oldB103Settle, effectstest.CardRefs(foe))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(foe) {
		t.Error("the creature was not exiled")
	}
}

func TestOldB103ChamberSentryCountsColorsAndShootsAndComesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.FillPoolColored(me, "W", 2)
	effectstest.FillPoolColored(me, "U", 1)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Chamber Sentry", TypeLine: "Artifact Creature — Construct", OracleID: oldB103Sentry,
		ManaCost: "{X}", Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 3}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 2 {
		t.Fatalf("%d counters, want 2 (white and blue were spent)", got)
	}
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, id).SummonedThisTurn = false })
	me.ManaPool = game.ManaPool{}
	effectstest.FillPool(me, 2)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true, XValue: 2,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("shoot: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent at %d, want %d", opp.Life, life-2)
	}
	if g.Battlefield.Contains(id) || !me.Graveyard.Contains(id) {
		t.Error("with its last counters removed the 0/0 Sentry should have died")
	}
}

func TestOldB103ChamberSentryReturnsFromTheGraveyardForFiveColors(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.PushCatalogGraveyardCard(me, "Chamber Sentry", "Artifact Creature — Construct", oldB103Sentry, 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, id, 1, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without the five colors")
	}
	for _, c := range []string{"W", "U", "B", "R", "G"} {
		effectstest.FillPoolColored(me, c, 1)
	}
	effectstest.Pr7Activate(t, g, me.ID, id, 1, game.ActivateAbilityParams{Strict: true})
	if !effectstest.HandHasID(me, id) {
		t.Error("the Sentry did not return to hand")
	}
}

func TestOldB103LawlessBrokerPutsACounterOnAnotherCreatureWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	broker := effectstest.B12Push(g, me.ID, "Lawless Broker", "Creature — Aetherborn Rogue", oldB103Broker, 3, 2)
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.KillCreature(t, g, me.ID, broker)
	g.RunStateChecksForTest()
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters on the Bear, want 1", got)
	}
}

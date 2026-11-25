package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b115_test.go — Ikoria: Lair of Behemoths, older-sets batch 115
// (#5512). Every test plays the real catalog entry through the engine.

const (
	oldB115PollywogSymbiote  = "0a766368-c982-4cba-9638-c7a56978c240"
	oldB115DrannithHealer    = "13e2a2c3-5ab5-4de4-96ca-5f9c72137000"
	oldB115SleeperDart       = "1f1b6130-3f18-4989-922b-438e41d683da"
	oldB115MemoryLeak        = "256d884f-5dea-4906-a53d-cf14d43e6bc8"
	oldB115KeepSafe          = "5602c14d-cc22-47d7-a1b0-6491749f00dd"
	oldB115OfOneMind         = "5a23c734-cbfb-419a-9bf1-126cb07fe427"
	oldB115RielleTheEverwise = "6f981e70-5828-49ac-8a72-cf1cc09be991"
	oldB115MonstrousStep     = "af045109-ed00-469f-814b-c759c0c5357c"
	oldB115Slitherwisp       = "ccc4f21f-cf8c-40bb-a6d6-1f624320c5b2"
	oldB115WhisperSquad      = "e6b27729-fea5-46c9-8624-07ba291791df"
)

func TestOldB115EveryCardIsRegisteredFull(t *testing.T) {
	effectstest.NewCatalogGame(t)
	for _, id := range []string{oldB115PollywogSymbiote, oldB115DrannithHealer, oldB115SleeperDart, oldB115MemoryLeak, oldB115KeepSafe, oldB115OfOneMind, oldB115RielleTheEverwise, oldB115MonstrousStep, oldB115Slitherwisp, oldB115WhisperSquad} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not in the catalog", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB115CyclingCardsCycleForTheirCost(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine, oracle, mana string
	}{
		{"Drannith Healer", "Creature — Human Cleric", oldB115DrannithHealer, "{C}"},
		{"Memory Leak", "Sorcery", oldB115MemoryLeak, "{C}"},
		{"Monstrous Step", "Sorcery", oldB115MonstrousStep, "{C}{C}"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			id, me := effectstest.CycleFromHand(t, g, tc.name, tc.typeLine, tc.oracle, tc.mana)
			if !me.Graveyard.Contains(id) {
				t.Fatal("the cycled card is not in the graveyard")
			}
			hand := me.Hand.Size()
			effectstest.PassPriorityAroundTable(t, g)
			if me.Hand.Size() != hand+1 {
				t.Errorf("hand %d -> %d, want the draw", hand, me.Hand.Size())
			}
		})
	}
}

func TestOldB115DrannithHealerGainsWhenYouCycleAnotherCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Drannith Healer", "Creature — Human Cleric", oldB115DrannithHealer, false)
	life := me.Life
	effectstest.CycleFromHand(t, g, "Imposing Vantasaur", "Creature — Dinosaur", "a0d5fce0-a439-496a-9fcf-9271503f3ce8", "{C}")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d", me.Life, life+1)
	}
}

func TestOldB115SleeperDartDrawsAndFreezesACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	hand := me.Hand.Size()
	dart := effectstest.CastHoldCreature(t, g, "Sleeper Dart", "Artifact", oldB115SleeperDart, 0, 0)
	if me.Hand.Size() != hand+1 {
		// the test seeds the card (+1), the cast takes it (-1) and the enters trigger draws one (+1)
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+1)
	}
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Bear", "", 2, 2)
	g.WithWriteLock(func() { _ = g.TapTargetForEffect(victim) })
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == dart {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, dart, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dart) {
		t.Error("the Dart was not sacrificed")
	}
	effectstest.AdvanceToMainOf(t, g, 1)
	if !effectstest.HmlTapped(t, g, victim) {
		t.Error("the creature untapped in its controller's untap step")
	}
}

func TestOldB115MemoryLeakExilesANonlandCardFromHandOrGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	spell := uuid.New()
	grave := uuid.New()
	land := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: spell, Name: "Hand Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	opp.Hand.PushTop(game.Card{InstanceID: land, Name: "Hand Forest", TypeLine: "Basic Land — Forest", Owner: opp.ID, Controller: opp.ID})
	opp.Graveyard.PushTop(game.Card{InstanceID: grave, Name: "Grave Bear", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	effectstest.CastCatalogSpell(t, g, "Memory Leak", "Sorcery", oldB115MemoryLeak, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	var c *game.PendingChoice
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Kind == game.PendingChoiceRevealedHandPick {
			c = pc
		}
	}
	if c == nil {
		t.Fatalf("no revealed-hand pick is open: %+v", g.PendingChoices)
	}
	for _, id := range c.DiscardOptions {
		if id == land {
			t.Error("a land card was offered")
		}
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.DiscardOptions {
		offered[id] = true
	}
	if !offered[spell] || !offered[grave] {
		t.Errorf("offered %v, want the hand's nonland card and the graveyard's", c.DiscardOptions)
	}
	if err := g.ResolvePendingChoice(c.ID, me.ID, []uuid.UUID{grave}); err != nil {
		t.Fatalf("resolve: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Contains(grave) {
		t.Error("the graveyard card was not exiled")
	}
	if !opp.Hand.Contains(spell) || !opp.Hand.Contains(land) {
		t.Error("the hand was touched")
	}
}

func TestOldB115KeepSafeCountersASpellAimedAtYourPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	// An opponent's Fully Grown, aimed at my creature, is on the stack.
	fully := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: fully, Name: "Fully Grown", TypeLine: "Instant", OracleID: "c98163f4-4a78-43f8-9855-4a23ebb9ceeb", Owner: opp.ID, Controller: opp.ID})
	effectstest.ApaMana(opp, "C", "C", "G")
	if err := g.CastSpell(opp.ID, fully, game.CastSpellParams{Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	keep := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: keep, Name: "Keep Safe", TypeLine: "Instant", OracleID: oldB115KeepSafe, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "U")
	if err := g.CastSpell(me.ID, keep, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(fully)}); err != nil {
		t.Fatalf("cast Keep Safe: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.CounterOn(g, bear, game.CounterTrample) != 0 {
		t.Error("the countered spell resolved")
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (Keep Safe drew a card)", me.Hand.Size(), hand+1)
	}
}

func TestOldB115KeepSafeRefusesASpellThatTargetsSomethingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	theirs := effectstest.B12Push(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	fully := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: fully, Name: "Fully Grown", TypeLine: "Instant", OracleID: "c98163f4-4a78-43f8-9855-4a23ebb9ceeb", Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "C", "G")
	if err := g.CastSpell(me.ID, fully, game.CastSpellParams{Targets: effectstest.CardTarget(theirs)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	keep := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: keep, Name: "Keep Safe", TypeLine: "Instant", OracleID: oldB115KeepSafe, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "U")
	if err := g.CastSpell(me.ID, keep, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(fully)}); err == nil {
		t.Error("Keep Safe targeted a spell aimed at an opponent's permanent")
	}
}

func TestOldB115OfOneMindDrawsTwoAndCostsLessWithBothKinds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.B12Push(g, me.ID, "Soldier", "Creature — Human Soldier", "", 2, 2)
	effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	hand := me.Hand.Size()
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Of One Mind", TypeLine: "Sorcery", ManaCost: "{2}{U}", OracleID: oldB115OfOneMind, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "U")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {U}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand+2)
	}
}

func TestOldB115OfOneMindIsFullPriceWithOnlyHumans(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.B12Push(g, me.ID, "Soldier", "Creature — Human Soldier", "", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Of One Mind", TypeLine: "Sorcery", ManaCost: "{2}{U}", OracleID: oldB115OfOneMind, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "U")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("Of One Mind cost {U} with no non-Human creature")
	}
}

func TestOldB115WhisperSquadFetchesAnotherTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	squad := effectstest.PushCatalogPermanent(g, me.ID, "Whisper Squad", "Creature — Human Soldier", oldB115WhisperSquad, false)
	other := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: other, Name: "Whisper Squad", TypeLine: "Creature — Human Soldier", OracleID: oldB115WhisperSquad, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "B")
	effectstest.ApaActivate(t, g, me, squad, 0, game.ActivateAbilityParams{})
	effectstest.AnswerSearchByID(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(other) {
		t.Fatalf("the second Whisper Squad is not on the battlefield: %+v", g.PendingChoices)
	}
	if !effectstest.HmlTapped(t, g, other) {
		t.Error("the fetched Whisper Squad entered untapped")
	}
}

func TestOldB115SlitherwispDrawsAndDrainsForAnotherFlashSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Slitherwisp", "Creature — Elemental Nightmare", oldB115Slitherwisp, false)
	flash := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: flash, Name: "Flash Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Keywords: []string{"flash"}, Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C")
	hand, life := me.Hand.Size(), opp.Life
	if err := g.CastSpell(me.ID, flash, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand-1+1 {
		t.Errorf("hand %d, want %d", me.Hand.Size(), hand)
	}
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
	// A spell without flash does nothing.
	plain := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: plain, Name: "Plain Bear", TypeLine: "Creature — Bear", ManaCost: "{1}", Power: 1, Toughness: 1, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C")
	life = opp.Life
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Error("a spell without flash triggered Slitherwisp")
	}
}

func TestOldB115PollywogSymbioteDiscountsAndLootsForMutateCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.PushCatalogPermanent(g, me.ID, "Pollywog Symbiote", "Creature — Frog", oldB115PollywogSymbiote, false)
	mut := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: mut, Name: "Mutater", TypeLine: "Creature — Beast", ManaCost: "{1}{G}", Keywords: []string{"mutate"}, Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "G")
	if err := g.CastSpell(me.ID, mut, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {G}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The loot draws a card and asks for a discard.
	if effectstest.LatestChooseCardsFor(g, me.ID) == nil && effectstest.LatestOptionPickFor(g, me.ID) == nil && len(g.PendingChoices) == 0 {
		t.Error("the loot asked for no discard")
	}
	// A creature without mutate pays full price.
	plain := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: plain, Name: "Plain", TypeLine: "Creature — Beast", ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	me.ManaPool.EmptyPool()
	effectstest.ApaMana(me, "G")
	if err := g.CastSpell(me.ID, plain, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a creature spell without mutate was discounted")
	}
}

func TestOldB115MonstrousStepPumpsAndForcesABlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	mine := effectstest.B12Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Monstrous Step", TypeLine: "Sorcery", OracleID: oldB115MonstrousStep, Owner: me.ID, Controller: me.ID})
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: mine}, {Kind: game.TargetCard, ID: theirs}}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, mine); p != 9 || tt != 9 {
		t.Errorf("target is %d/%d, want 9/9", p, tt)
	}
	if p, tt := effectstest.PtOf(t, g, theirs); p != 2 || tt != 2 {
		t.Errorf("the second target is %d/%d, want unchanged 2/2", p, tt)
	}
}

func TestOldB115MonstrousStepWithOneTargetJustPumps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	mine := effectstest.B12Push(g, me.ID, "Mine", "Creature — Bear", "", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Monstrous Step", "Sorcery", oldB115MonstrousStep, effectstest.CardTarget(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, mine); p != 9 || tt != 9 {
		t.Errorf("target is %d/%d, want 9/9", p, tt)
	}
}

func TestOldB115RielleGrowsWithInstantsAndSorceriesInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	rielle := effectstest.B12Push(g, me.ID, "Rielle, the Everwise", "Legendary Creature — Human Wizard", oldB115RielleTheEverwise, 0, 3)
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "I", TypeLine: "Instant", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "S", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "C", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	if p, _ := effectstest.PtOf(t, g, rielle); p != 2 {
		t.Errorf("Rielle's power %d, want 2", p)
	}
}

func TestOldB115RielleDrawsForTheFirstDiscardOfTheTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.PushCatalogPermanent(g, me.ID, "Rielle, the Everwise", "Legendary Creature — Human Wizard", oldB115RielleTheEverwise, false)
	effectstest.CycleFromHand(t, g, "Imposing Vantasaur", "Creature — Dinosaur", "a0d5fce0-a439-496a-9fcf-9271503f3ce8", "{C}")
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	// The cycling draw plus Rielle's draw for the discard.
	if me.Hand.Size() != hand+2 {
		t.Fatalf("hand %d -> %d, want +2 (cycling draw and Rielle)", hand, me.Hand.Size())
	}
	effectstest.CycleFromHand(t, g, "Imposing Vantasaur", "Creature — Dinosaur", "a0d5fce0-a439-496a-9fcf-9271503f3ce8", "{C}")
	hand = me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want +1: the second discard of the turn draws only the cycling card", hand, me.Hand.Size())
	}
}

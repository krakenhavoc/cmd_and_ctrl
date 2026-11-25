package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b040_cards_test.go — S82 older sets, Strixhaven batch 40 (#5502).
// Uvilda, Dean of Perfection // Nassari, Dean of Expression is skipped.

const (
	oldB40Squad     = "41ee78ce-a5ef-4f88-b479-2eb6651123f7"
	oldB40Duskmage  = "44958014-7e30-4e70-9aef-b148843f9c45"
	oldB40Laureate  = "53a17d4f-cdbc-4669-a30e-5d5a334cc3c0"
	oldB40Elocution = "58acd10d-d81e-49ae-9e08-cff4c6f0c6a1"
	oldB40Show      = "6f7c29ff-955c-496b-9300-2efb9e99ba2d"
	oldB40Closing   = "79ac7b6a-1c25-45f6-84b4-5c4a78d6fd6e"
	oldB40Loremage  = "7d1dbe80-9ad0-41ad-8ba6-e1094f596eaf"
	oldB40Familiar  = "87266dec-ac03-4f21-b235-67c14c3024db"
	oldB40Sculptor  = "8a9fe2dd-50a7-4010-96ba-df4dec265e16"
)

func TestOldB40CardsAreFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB40Squad: "Spiteful Squad", oldB40Duskmage: "Promising Duskmage", oldB40Laureate: "Shadewing Laureate",
		oldB40Elocution: "Exhilarating Elocution", oldB40Show: "Show of Confidence", oldB40Closing: "Closing Statement",
		oldB40Loremage: "Oriq Loremage", oldB40Familiar: "Stonebinder's Familiar",
		oldB40Sculptor: "Torrent Sculptor", oldB40Sculptor + "#1": "Flamethrower Sonata",
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

func TestOldB40SpitefulSquadMovesItsCountersWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	squad := effectstest.MrdEnter(t, g, me, "Spiteful Squad", "Creature — Human Warlock", oldB40Squad)
	if got := effectstest.CountersOn(g, squad, game.CounterPlusOne); got != 2 {
		t.Fatalf("entered with %d counters, want 2", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, squad, "deathtouch") {
		t.Error("no deathtouch")
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B25Destroy(g, squad)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the Bear has %d counters, want 2", got)
	}
}

func TestOldB40PromisingDuskmageDrawsOnlyIfItHadACounter(t *testing.T) {
	for _, counter := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		mage := effectstest.B12Push(g, me.ID, "Promising Duskmage", "Creature — Human Warlock", oldB40Duskmage, 2, 3)
		if counter {
			effectstest.AddCounters(t, g, mage, game.CounterPlusOne, 1)
		}
		before := me.Hand.Size()
		effectstest.B25Destroy(g, mage)
		effectstest.PassPriorityAroundTable(t, g)
		want := 0
		if counter {
			want = 1
		}
		if got := me.Hand.Size() - before; got != want {
			t.Errorf("counter=%v: drew %d, want %d", counter, got, want)
		}
	}
}

func TestOldB40ShadewingLaureateCountsOnlyOtherFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	laureate := effectstest.B12Push(g, me.ID, "Shadewing Laureate", "Creature — Human Warlock", oldB40Laureate, 2, 2)
	flyer := effectstest.B12Creature(g, me.ID, "Flyer", "Creature — Bird", 1, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, flyer).Keywords = []string{"flying"} })
	walker := effectstest.B12Creature(g, me.ID, "Walker", "Creature — Bear", 1, 1)
	// A creature without flying dying: nothing.
	effectstest.B25Destroy(g, walker)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a non-flyer's death triggered the Laureate")
	}
	// A flyer dying: a counter on the chosen creature.
	effectstest.B25Destroy(g, flyer)
	effectstest.AnswerPickTarget(t, g, laureate)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, laureate, game.CounterPlusOne); got != 1 {
		t.Errorf("the Laureate has %d counters, want 1", got)
	}
	// An opponent's flyer: nothing.
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Their Flyer", "Creature — Bird", 1, 1)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, theirs).Keywords = []string{"flying"} })
	effectstest.B25Destroy(g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's flyer triggered the Laureate")
	}
}

func TestOldB40ExhilaratingElocution(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Exhilarating Elocution", "Sorcery", oldB40Elocution, effectstest.CardTarget(a))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, a, game.CounterPlusOne); got != 2 {
		t.Errorf("the target has %d counters, want 2", got)
	}
	for id, want := range map[uuid.UUID]int{a: 4, b: 3, theirs: 2} {
		if p := effectstest.EffectivePowerWithCounters(t, g, id); p != want {
			t.Errorf("%v has power %d, want %d", id, p, want)
		}
	}
	effectstest.AdvancePastCleanupOf(t, g, 0)
	if p := effectstest.EffectivePowerWithCounters(t, g, b); p != 2 {
		t.Errorf("the pump outlived the turn: %d", p)
	}
}

func TestOldB40ShowOfConfidenceCopiesPerOtherInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	// A creature spell does not count; two instants/sorceries do.
	effectstest.P1g2Cast(t, g, "Filler Creature", "Creature — Bear", "", "{0}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1g2Cast(t, g, "Filler Instant", "Instant", "", "{0}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1g2Cast(t, g, "Filler Sorcery", "Sorcery", "", "{0}", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.CastCatalogSpell(t, g, "Show of Confidence", "Instant", oldB40Show, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	// The copies may choose new targets: answer any such prompt with the Bear.
	for i := 0; i < 4; i++ {
		if effectstest.LatestPickTarget(g, me.ID) == nil {
			break
		}
		effectstest.PickCard(t, g, me.ID, bear)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 3 {
		t.Errorf("the Bear has %d counters, want 3 (the spell and two copies)", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("no vigilance")
	}
}

func TestOldB40ShowOfConfidenceAloneMakesNoCopies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Show of Confidence", "Instant", oldB40Show, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d counters, want 1", got)
	}
}

func TestOldB40ClosingStatementDestroysAndCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Closing Statement", "Instant", oldB40Closing,
		[]game.TargetRef{{Kind: game.TargetCard, ID: ogre}, {Kind: game.TargetCard, ID: mine}})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the Ogre survived")
	}
	if got := effectstest.CountersOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("my creature has %d counters, want 1", got)
	}
}

func TestOldB40ClosingStatementRefusesYourOwnCreatureAndCostsLessAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	id := effectstest.HandCardFull(me, "Closing Statement", "Instant", "{3}{W}{B}", oldB40Closing, []string{"W", "B"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(mine)}); err == nil {
		t.Error("destroyed a creature I control")
	}
	// Main phase: three mana is not enough (it costs five).
	effectstest.FloatForTest(g, me, "WBCC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(ogre)}); err == nil {
		t.Error("cast for four mana in the main phase")
	}
	// My end step: {3}{W}{B} costs {2} less, so {1}{W}{B} is enough.
	me.ManaPool.EmptyPool()
	effectstest.AdvanceToEndStep(t, g)
	effectstest.FloatForTest(g, me, "WBC")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(ogre)}); err != nil {
		t.Fatalf("cast at my end step for three mana: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("the Ogre survived")
	}
}

func TestOldB40OriqLoremageSearchesIntoTheGraveyard(t *testing.T) {
	for _, instant := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		mage := effectstest.B12Push(g, me.ID, "Oriq Loremage", "Creature — Human Warlock", oldB40Loremage, 3, 3)
		typeLine := "Basic Land — Forest"
		if instant {
			typeLine = "Instant"
		}
		card := effectstest.SeedTopOfLibrary(me, game.Card{Name: "Found", TypeLine: typeLine})
		effectstest.B16Activate(t, g, me.ID, mage, 0, game.ActivateAbilityParams{})
		effectstest.TakeSearchPrompts(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if !me.Graveyard.Contains(card) {
			t.Fatalf("instant=%v: the card did not reach the graveyard", instant)
		}
		want := 0
		if instant {
			want = 1
		}
		if got := effectstest.CountersOn(g, mage, game.CounterPlusOne); got != want {
			t.Errorf("instant=%v: %d counters, want %d", instant, got, want)
		}
		if !effectstest.IsTapped(g, mage) {
			t.Error("the Loremage did not tap")
		}
	}
}

func TestOldB40StonebindersFamiliarGrowsOncePerTurnOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fam := effectstest.B12Push(g, me.ID, "Stonebinder's Familiar", "Creature — Spirit Dog", oldB40Familiar, 1, 1)
	exile := func(owner *game.Player, name string) {
		id := effectstest.B13HandCard(owner, name, "Instant", "")
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: owner.ID}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
			t.Fatal(err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	exile(me, "First")
	if got := effectstest.CountersOn(g, fam, game.CounterPlusOne); got != 1 {
		t.Fatalf("after one exile the Familiar has %d counters, want 1", got)
	}
	exile(opp, "Second")
	if got := effectstest.CountersOn(g, fam, game.CounterPlusOne); got != 1 {
		t.Errorf("a second exile the same turn made %d counters, want 1", got)
	}
}

func TestOldB40StonebindersFamiliarIgnoresOpponentsTurns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	fam := effectstest.B12Push(g, me.ID, "Stonebinder's Familiar", "Creature — Spirit Dog", oldB40Familiar, 1, 1)
	effectstest.AdvanceToMainOf(t, g, 1)
	id := effectstest.B13HandCard(opp, "Theirs", "Instant", "")
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: opp.ID}, game.ZoneRef{Kind: game.ZoneExile}, id); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, fam, game.CounterPlusOne); got != 0 {
		t.Errorf("an exile on an opponent's turn made %d counters", got)
	}
}

func TestOldB40TorrentSculptorExilesForHalfTheManaValueRoundedUp(t *testing.T) {
	for _, tc := range []struct {
		mana string
		want int
	}{{"{2}{R}", 2}, {"{3}{R}", 2}, {"{1}", 1}} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.AdvanceToMain(t, g)
		spell := effectstest.PushGraveyardPermanent(me, "Spell", "Instant", tc.mana)
		other := effectstest.PushGraveyardPermanent(me, "Creature Card", "Creature — Bear", "{5}")
		id := effectstest.CastCatalogSpell(t, g, "Torrent Sculptor", "Creature — Merfolk Wizard", oldB40Sculptor, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if c := effectstest.LatestChoiceOfKind(g, game.PendingChoiceChooseCards); c != nil {
			if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{spell}); err != nil {
				t.Fatalf("pick: %v", err)
			}
		}
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Exile.Contains(spell) {
			t.Errorf("%s: the instant was not exiled", tc.mana)
		}
		if !me.Graveyard.Contains(other) {
			t.Errorf("%s: the creature card left the graveyard", tc.mana)
		}
		if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != tc.want {
			t.Errorf("%s: %d counters, want %d", tc.mana, got, tc.want)
		}
		if !effectstest.HasEffectiveKeyword(t, g, id, "ward 2") && !effectstest.HasEffectiveKeyword(t, g, id, "ward") {
			t.Error("no ward")
		}
	}
}

func TestOldB40FlamethrowerSonataLoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	burn := effectstest.HandCardFull(me, "Burn", "Instant", "{2}{R}", "", []string{"R"})
	effectstest.CastCatalogSpell(t, g, "Flamethrower Sonata", "Sorcery", oldB40Sculptor+"#1", nil)
	effectstest.PassPriorityAroundTable(t, g)
	hand := me.Hand.Size()
	effectstest.AnswerDiscard(t, g, me.ID, burn)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(burn) {
		t.Fatal("the instant was not discarded")
	}
	if got := me.Hand.Size(); got != hand {
		// discarded one, drew one
		t.Errorf("hand %d, want %d after discard and draw", got, hand)
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickCard(t, g, me.ID, ogre)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.DamageOn(t, g, ogre); got != 3 {
		t.Errorf("the Ogre has %d damage, want 3 (the discarded card's mana value)", got)
	}
}

func TestOldB40FlamethrowerSonataNeedsAnInstantOrSorceryForDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 4, 4)
	bear := effectstest.HandCardFull(me, "Bear Card", "Creature — Bear", "{2}", "", nil)
	effectstest.CastCatalogSpell(t, g, "Flamethrower Sonata", "Sorcery", oldB40Sculptor+"#1", nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a creature card discarded asked for a damage target")
	}
	if got := effectstest.DamageOn(t, g, ogre); got != 0 {
		t.Errorf("the Ogre took %d", got)
	}
}

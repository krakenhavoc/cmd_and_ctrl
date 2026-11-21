package testss82v

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// probeClone is a green one-drop creature card that is a Clone: it enters
// with a "copy a creature?" question.
func probeClone() game.Card {
	return game.Card{
		Name: "Clone", TypeLine: "Creature — Shapeshifter", OracleID: effectstest.OracleClone,
		ManaCost: "{G}", Colors: []string{"G"},
	}
}

// --- library searches to the battlefield ------------------------------------

// Cultivate's first half, a basic land onto the battlefield tapped, stops
// on the land's entry question. The restored answer lands it and the
// hand half (a registered tail, cultivate/basic-land-to-hand) asks for the
// second land.
func TestCultivateIsARestorePointAtTheFetchedLandsQuestion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	ids := effectstest.SeedSearchLibrary(me, effectstest.BasicShockland(), game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	crypt, forest := ids[0], ids[1]

	effectstest.CastCatalogSpell(t, g, "Cultivate", "Sorcery", effectstest.CultivateOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, crypt)
	if effectstest.EntryPayLifeChoiceFor(g, me.ID) == nil {
		t.Fatal("the fetched land did not ask")
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	me = g.Seats[g.Turn.ActiveSeat]
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Fatal("the restored table asked for the hand half before the land was answered")
	}
	effectstest.AnswerEntryPayLife(t, g, me.ID, false)
	card, ok := effectstest.BattlefieldCard(g, crypt)
	if !ok || !card.Tapped {
		t.Fatalf("the fetched land is on the battlefield %v, tapped %v, want tapped", ok, card.Tapped)
	}
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("the restored answer did not go on to the hand half")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, forest)
	if !me.Hand.Contains(forest) {
		t.Error("the second land did not reach the hand")
	}
	if n, shuffled := effectstest.SearchEventsSince(g, 0); n != 2 || !shuffled {
		t.Errorf("%d search events (shuffled %v), want two and the last shuffled", n, shuffled)
	}
}

// Fabled Passage: the fetched land's question is a restore point, and the
// restored answer untaps it (four lands), the registered tail
// fabled-passage/untap-that-land told the land that entered.
func TestFabledPassageIsARestorePointAtTheFetchedLandsQuestion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := effectstest.Top100ActiveSeat(g)
	effectstest.SeedPermanentFor(g, me.ID, "Forest A", "Basic Land — Forest")
	effectstest.SeedPermanentFor(g, me.ID, "Forest B", "Basic Land — Forest")
	effectstest.SeedPermanentFor(g, me.ID, "Forest C", "Basic Land — Forest")
	passage := effectstest.PushCatalogPermanent(g, me.ID, "Fabled Passage", "Land", effectstest.FabledPassageOracle, false)
	crypt := effectstest.SeedSearchLibrary(me, effectstest.BasicShockland())[0]

	if err := g.ActivateCatalogAbility(me.ID, passage, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, crypt)
	if effectstest.EntryPayLifeChoiceFor(g, me.ID) == nil {
		t.Fatal("the fetched land did not ask")
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AnswerEntryPayLife(t, g, me.ID, false)
	card, ok := effectstest.BattlefieldCard(g, crypt)
	if !ok {
		t.Fatal("the fetched land did not enter")
	}
	if card.Tapped {
		t.Error("the fetched land is still tapped: the restored tail did not untap it")
	}
	if n, shuffled := effectstest.SearchEventsSince(g, 0); n != 1 || !shuffled {
		t.Errorf("%d search events (shuffled %v), want one that shuffled", n, shuffled)
	}
}

// Green Sun's Zenith: the creature's copy question is a restore point, and
// the restored answer shuffles the Zenith into the library (the tail
// green-suns-zenith/shuffle-into-library, its owner riding Tail.Player).
func TestGreenSunsZenithIsARestorePointAtTheFoundCreaturesQuestion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	clone := effectstest.PushLibraryCardForTest(me, probeClone())

	zenith := effectstest.CastXSpell(t, g, "Green Sun's Zenith", "Sorcery", effectstest.SweepCGreenSunsZenithOracle, "{X}{G}", 1, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, clone)
	if effectstest.CopyPrompt(g) == nil {
		t.Fatal("the found Clone did not ask")
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	me = g.Seats[0]
	if me.Library.Contains(zenith) {
		t.Fatal("the restored table shuffled the Zenith in before the Clone's answer")
	}
	effectstest.AnswerCloneCopy(t, g, bear)
	if !g.Battlefield.Contains(clone) {
		t.Error("the Clone did not enter")
	}
	if !me.Library.Contains(zenith) || me.Graveyard.Contains(zenith) {
		t.Errorf("the Zenith is in the library %v, in the graveyard %v, want it shuffled into the library",
			me.Library.Contains(zenith), me.Graveyard.Contains(zenith))
	}
}

// Nature's Rhythm: "then shuffle" is the shared tail then/shuffle, whose
// player is the spell's owner.
func TestNaturesRhythmShuffleIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.PushCatalogPermanent(g, me.ID, "Bear", "Creature — Bear", "", false)
	clone := effectstest.PushLibraryCardForTest(me, probeClone())

	effectstest.CastXSpell(t, g, "Nature's Rhythm", "Sorcery", effectstest.NaturesRhythmOracle, "{X}{G}{G}", 1, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, clone)
	if effectstest.CopyPrompt(g) == nil {
		t.Fatal("the found Clone did not ask")
	}
	before := effectstest.LibraryShuffles(g)

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	if n := effectstest.LibraryShuffles(g) - before; n != 0 {
		t.Fatalf("the restored table shuffled %d times before the answer", n)
	}
	effectstest.AnswerCloneCopy(t, g, bear)
	if !g.Battlefield.Contains(clone) {
		t.Error("the Clone did not enter")
	}
	if n := effectstest.LibraryShuffles(g) - before; n != 1 {
		t.Errorf("%d shuffles after the restored answer, want one", n)
	}
}

// Anchor to Reality: the sacrificed permanent's mana value rides the tail
// (Tail.N), so the scry 2 for a cheaper find still happens after a CR 616
// ordering prompt over the Equipment's entry and a restore.
func TestAnchorToRealityIsARestorePointAtTheEquipmentsOrderingPrompt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.Pe2EntersTapped(g, opp.ID)
	for i := 0; i < 3; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Land"})
	}
	// An artifact creature, so that Kismet and Thalia both apply.
	golem := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Golem Equipment", TypeLine: "Artifact Creature — Equipment Golem", ManaCost: "{3}"})
	effectstest.PaCast(t, g, "Anchor to Reality", "Sorcery", "{2}{U}{U}", effectstest.AnchorToRealityOracle, game.CastSpellParams{
		SacrificeIDs: []uuid.UUID{effectstest.PaPermanent(g, me.ID, "Relic", "Artifact", "{5}", 0, 0)},
	})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, golem)
	if effectstest.B06ReplacementOrderFor(g, me.ID) == nil {
		t.Fatal("the Equipment's entry did not ask for an order")
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	me = g.Seats[0]
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry) != nil {
		t.Fatal("the restored table scried before the entry was settled")
	}
	effectstest.PeAnswerOrder(t, g, me.ID)
	if !g.Battlefield.Contains(golem) {
		t.Fatal("the Equipment did not enter")
	}
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceScry) == nil {
		t.Error("the Equipment costs less than the sacrificed {5} permanent: the restored tail did not scry 2")
	}
}

// The Overlook lands: "you gain 1 life" is the shared tail then/change-life
// (Source the land, N the life), run after the fetched land's question.
func TestAnOverlookLandsLifeGainIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	grave := game.Card{Name: "Watery Grave", TypeLine: "Basic Land — Island Swamp", OracleID: effectstest.ShocklandOracleIDs["Watery Grave"]}
	id := effectstest.SeedSearchLibrary(me, grave, game.Card{Name: "Island A", TypeLine: "Basic Land — Island"})[0]
	effectstest.PlayLandFromHand(t, g, "Maestros Theater", effectstest.B11MaestrosTheaterOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, id)
	if effectstest.EntryPayLifeChoiceFor(g, me.ID) == nil {
		t.Fatal("the fetched land did not ask")
	}
	life := me.Life

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	me = g.Seats[g.Turn.ActiveSeat]
	effectstest.AnswerEntryPayLife(t, g, me.ID, false)
	if _, ok := effectstest.BattlefieldCard(g, id); !ok {
		t.Fatal("the fetched land did not enter")
	}
	if me.Life != life+1 {
		t.Errorf("life %d, want %d after the restored tail", me.Life, life+1)
	}
}

// Hour of Promise: the Zombies are words of the spell's text, read through
// the item, which rides the tail as data (Context.ItemTail, KeyedTail.Asked).
func TestHourOfPromiseIsARestorePointAtTheFetchedLandsQuestion(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLand(g, me.ID, "Desert of the Glorified", "Land — Desert", "")
	effectstest.SeedLand(g, me.ID, "Desert of the True", "Land — Desert", "")
	ids := effectstest.SeedSearchLibrary(me,
		game.Card{Name: "Blood Crypt", TypeLine: "Land — Swamp Mountain", OracleID: effectstest.BloodCryptOracle},
		game.Card{Name: "Desert A", TypeLine: "Land — Desert"},
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
	)
	effectstest.CastCatalogSpell(t, g, "Hour of Promise", "Sorcery", effectstest.B30HourOfPromiseOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0], ids[1])
	if effectstest.EntryPayLifeChoiceFor(g, me.ID) == nil {
		t.Fatal("the fetched Blood Crypt did not ask")
	}

	g = effectstest.StrictRestoreWithFrame(t, g, game.ReplacementResumeMoveTail)
	me = g.Seats[0]
	if got := effectstest.B30TokensNamed(g, me.ID, "Zombie"); got != 0 {
		t.Fatalf("%d Zombies before the restored answer", got)
	}
	effectstest.AnswerEntryPayLife(t, g, me.ID, false)
	if !g.Battlefield.Contains(ids[0]) || !g.Battlefield.Contains(ids[1]) {
		t.Fatal("both fetched lands did not enter")
	}
	if got := effectstest.B30TokensNamed(g, me.ID, "Zombie"); got != 2 {
		t.Errorf("three Deserts after the restored search is two Zombies, got %d", got)
	}
}

// --- scry and surveil -------------------------------------------------------

// Behold the Multiverse: "scry 2, then draw two cards" is the shared tail
// then/draw-n-cards.
func TestBeholdTheMultiverseScryIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Behold the Multiverse", "Instant", effectstest.BeholdTheMultiverseOracl, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("Behold the Multiverse did not scry")
	}

	g = effectstest.StrictRestoreWithScryThen(t, g)
	me = g.Seats[0]
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the restored table has no scry")
	}
	hand := me.Hand.Size()
	if err := g.ResolveScry(c.ID, me.ID, nil, c.ScryCards); err != nil {
		t.Fatalf("ResolveScry: %v", err)
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d -> %d, want two cards drawn", hand, got)
	}
}

// Glassblower's Puzzleknot: "scry 2, then you get {E}{E}" is the shared tail
// then/get-energy.
func TestGlassblowersPuzzleknotScryIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B", "C")
	effectstest.CastAndResolveCreature(t, g, "Glassblower's Puzzleknot", "Artifact", effectstest.EaGlassblower)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ScryChoiceFor(g, me.ID) == nil {
		t.Fatal("the Puzzleknot did not scry")
	}

	g = effectstest.StrictRestoreWithScryThen(t, g)
	me = g.Seats[0]
	if effectstest.EnergyOf(me) != 0 {
		t.Fatalf("energy %d before the restored scry is answered", effectstest.EnergyOf(me))
	}
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	if effectstest.EnergyOf(me) != 2 {
		t.Errorf("energy %d after the restored scry, want 2", effectstest.EnergyOf(me))
	}
}

// Eye of Jace's surveil: the upkeep trigger rides the tail as data
// (KeyedTail.Asked), so the seven-card check and the sacrifice, which read
// the item, still happen after a restore.
func TestEyeOfJaceSurveilIsARestorePoint(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	eye := effectstest.RfmaPush(g, me.ID, "Eye of Jace", "Artifact", effectstest.RfmaEyeOfJace, 0, 0)
	for i := 0; i < 6; i++ {
		effectstest.Batch01GraveyardCard(me, "Fodder", "Sorcery")
	}
	ids := effectstest.SeedLibrary(me, "Top", "Next")
	lifeMe, lifeOpp := me.Life, opp.Life
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SurveilChoiceFor(g, me.ID) == nil {
		t.Fatal("the Eye did not surveil")
	}

	snap := g.CaptureSnapshot()
	carried := false
	for _, c := range snap.PendingChoices {
		if c.ScryThen != nil && c.ScryThen.Asked != nil && c.ScryThen.Asked.SourceCardID == eye {
			carried = true
		}
	}
	if !carried {
		t.Fatal("the surveil's keyed then does not carry the asking item")
	}
	g = effectstest.StrictRestoreWithScryThen(t, g)
	me, opp = g.Seats[0], g.Seats[1]
	carried = false
	for _, c := range g.CaptureSnapshot().PendingChoices {
		if c.ScryThen != nil && c.ScryThen.Asked != nil && c.ScryThen.Asked.SourceCardID == eye {
			carried = true
		}
	}
	if !carried {
		t.Fatal("the restored surveil lost the asking item its keyed then carried")
	}
	effectstest.RfmaAnswerSurveil(t, g, me, ids[0]) // the seventh card
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(eye) {
		t.Error("the Eye should have been sacrificed")
	}
	if me.Life != lifeMe+2 || opp.Life != lifeOpp-2 {
		t.Errorf("life %d / %d, want %d / %d", me.Life, opp.Life, lifeMe+2, lifeOpp-2)
	}
}

// Every tail the cards share, run off a scry that was restored: shuffle,
// life gained and lost, discard at random.
func TestSharedTailsRunAfterARestoredScry(t *testing.T) {
	for _, tc := range []struct {
		name string
		tail func(me uuid.UUID) game.KeyedTail
		want func(t *testing.T, g *game.Game, me *game.Player, before probeState)
	}{
		{"then/shuffle", func(me uuid.UUID) game.KeyedTail { return game.KeyedTail{Ref: effects.ThenShuffle, Player: me} },
			func(t *testing.T, g *game.Game, _ *game.Player, before probeState) {
				if n := effectstest.LibraryShuffles(g) - before.shuffles; n != 1 {
					t.Errorf("%d shuffles, want one", n)
				}
			}},
		{"then/change-life gain", func(me uuid.UUID) game.KeyedTail {
			return game.KeyedTail{Ref: effects.ThenChangeLife, Player: me, Source: uuid.New(), N: 3}
		}, func(t *testing.T, _ *game.Game, me *game.Player, before probeState) {
			if me.Life != before.life+3 {
				t.Errorf("life %d, want %d", me.Life, before.life+3)
			}
		}},
		{"then/change-life loss", func(me uuid.UUID) game.KeyedTail {
			return game.KeyedTail{Ref: effects.ThenChangeLife, Player: me, Source: uuid.New(), N: -3}
		}, func(t *testing.T, _ *game.Game, me *game.Player, before probeState) {
			if me.Life != before.life-3 {
				t.Errorf("life %d, want %d", me.Life, before.life-3)
			}
		}},
		{"then/discard-at-random", func(me uuid.UUID) game.KeyedTail {
			return game.KeyedTail{Ref: effects.ThenDiscardAtRandom, Player: me, N: 1}
		}, func(t *testing.T, _ *game.Game, me *game.Player, before probeState) {
			if me.Hand.Size() != before.hand-1 || me.Graveyard.Size() != before.graveyard+1 {
				t.Errorf("hand %d graveyard %d, want %d and %d", me.Hand.Size(), me.Graveyard.Size(), before.hand-1, before.graveyard+1)
			}
		}},
		{"then/draw-n-cards", func(me uuid.UUID) game.KeyedTail {
			return game.KeyedTail{Ref: effects.ThenDrawCards, Player: me, N: 3}
		}, func(t *testing.T, _ *game.Game, me *game.Player, before probeState) {
			if me.Hand.Size() != before.hand+3 {
				t.Errorf("hand %d, want %d", me.Hand.Size(), before.hand+3)
			}
		}},
		{"then/get-energy", func(me uuid.UUID) game.KeyedTail {
			return game.KeyedTail{Ref: effects.ThenGetEnergy, Player: me, N: 4}
		}, func(t *testing.T, _ *game.Game, me *game.Player, before probeState) {
			if effectstest.EnergyOf(me) != before.energy+4 {
				t.Errorf("energy %d, want %d", effectstest.EnergyOf(me), before.energy+4)
			}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			for i := 0; i < 3; i++ {
				me.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Card", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
			}
			g.WithWriteLock(func() {
				if err := (effects.Scry{N: 1, ThenTail: tc.tail(me.ID)}).Apply(effects.NewContext(g, &game.StackItem{Controller: me.ID})); err != nil {
					t.Fatalf("Scry: %v", err)
				}
			})
			if effectstest.ScryChoiceFor(g, me.ID) == nil {
				t.Fatal("no scry is open")
			}
			g = effectstest.StrictRestoreWithScryThen(t, g)
			me = g.Seats[0]
			before := probeState{
				shuffles: effectstest.LibraryShuffles(g), life: me.Life, hand: me.Hand.Size(),
				graveyard: me.Graveyard.Size(), energy: effectstest.EnergyOf(me),
			}
			effectstest.AnswerScryKeepAll(t, g, me.ID)
			tc.want(t, g, me, before)
		})
	}
}

type probeState struct {
	shuffles, life, hand, graveyard, energy int
}

// A tail the card left without a Player or a Source is given the scrying
// player and the effect's source card (Scry.ThenTail), and a search's tail
// the searcher and the source (SearchLibrary.ThenTail).
func TestATailWithoutAPlayerRunsForTheEffectsPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ctx := effects.NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()})
	g.WithWriteLock(func() {
		if err := (effects.Scry{N: 1, ThenTail: game.KeyedTail{Ref: effects.ThenDrawCards, N: 1}}).Apply(ctx); err != nil {
			t.Fatalf("Scry: %v", err)
		}
	})
	g = effectstest.StrictRestoreWithScryThen(t, g)
	me = g.Seats[0]
	hand := me.Hand.Size()
	effectstest.AnswerScryKeepAll(t, g, me.ID)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want the scrying player to draw", hand, me.Hand.Size())
	}

	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	ctx = effects.NewContext(g, &game.StackItem{Controller: me.ID, SourceCardID: uuid.New()})
	ids := effectstest.SeedSearchLibrary(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	life := me.Life
	g.WithWriteLock(func() {
		if err := (effects.SearchLibrary{
			Player: me.ID, Dest: game.ZoneHand, Limit: 1,
			ThenTail: game.KeyedTail{Ref: effects.ThenChangeLife, N: 2},
		}).Apply(ctx); err != nil {
			t.Fatalf("SearchLibrary: %v", err)
		}
	})
	effectstest.TakeSearchPrompts(t, g)
	if !me.Hand.Contains(ids[0]) || me.Life != life+2 {
		t.Errorf("hand has the card %v, life %d, want it in hand and %d", me.Hand.Contains(ids[0]), me.Life, life+2)
	}
}

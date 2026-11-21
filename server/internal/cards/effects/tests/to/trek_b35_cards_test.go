package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b35_cards_test.go — Star Trek Commander, batch 35 (#4974). Each
// test plays the real catalog entry and asserts what its text says.

const (
	trekB35KesOracle       = "714dcaf0-442b-4849-917a-7e6af2ea0210"
	trekB35UhuraOracle     = "d55f0074-2772-48b1-8d60-ca2da52d4c9d"
	trekB35GiantsOracle    = "8abb1be2-d02b-4a78-9fcb-d2c9102a840b"
	trekB35AnnexOracle     = "9f38914f-7870-44bf-887e-ee893e98989c"
	trekB35SamKirkOracle   = "eec55307-8525-4dae-a030-d210f3895138"
	trekB35SevenOracle     = "22c3b484-0f96-4051-807c-fc0fcb7cd1a8"
	trekB35TribulationOrcl = "912d6e44-b479-4e07-88a8-9ab5132163f4"
	trekB35HubOracle       = "4325d4d2-5892-4c20-a994-7396b715e94f"
	trekB35TuvokOracle     = "05c5f17e-40f9-4259-9b03-03683e1c5ba6"
	trekB35VoxOracle       = "7e58ea13-4041-4ef4-bc09-1e4bf79a842a"
)

func TestTrekB35EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB35KesOracle:       "Kes, Evolving Telepath",
		trekB35UhuraOracle:     "Lieutenant Nyota Uhura",
		trekB35GiantsOracle:    "On the Shoulders of Giants",
		trekB35AnnexOracle:     "Optimization Annex",
		trekB35SamKirkOracle:   "Sam Kirk, Xenoanthropologist",
		trekB35SevenOracle:     "Seven of Nine, Tertiary Adjunct",
		trekB35TribulationOrcl: "Temporal Tribulation",
		trekB35HubOracle:       "Transwarp Hub",
		trekB35TuvokOracle:     "Tuvok, Chief Tactical Officer",
		trekB35VoxOracle:       "Võx, Child of the Collective",
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

// --- shared helpers ------------------------------------------------------------

func trekB35Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

// trekB35CountNamed counts the battlefield permanents with the name.
func trekB35CountNamed(g *game.Game, name string) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				n++
			}
		}
	})
	return n
}

// --- Kes, Evolving Telepath ----------------------------------------------------

func TestTrekB35KesDrawsOnePerTimeCounterAndSacrificesAtNine(t *testing.T) {
	for _, c := range []struct {
		name      string
		before    int
		wantDraws int
		wantSac   bool
	}{
		{"the first end step", 0, 1, false},
		{"eight counters", 7, 8, false},
		{"nine counters", 8, 9, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _ := trekB35Seats(g)
			effectstest.TrekB35Library(g, me, 12)
			kes := effectstest.B27Push(g, me.ID, "Kes, Evolving Telepath", "Legendary Creature — Ocampa Advisor", trekB35KesOracle, "{4}{U}", 2, 6, "U")
			effectstest.TrekB35SetCounters(g, kes, game.CounterTime, c.before)
			hand := me.Hand.Size()
			trekB15ToEndStep(t, g)
			if got := me.Hand.Size() - hand; got != c.wantDraws {
				t.Errorf("drew %d, want %d", got, c.wantDraws)
			}
			onField := g.Battlefield.Contains(kes)
			if onField == c.wantSac {
				t.Errorf("Kes on the battlefield = %v, want sacrificed = %v", onField, c.wantSac)
			}
			if !c.wantSac {
				if got := effectstest.CounterOn(g, kes, game.CounterTime); got != c.before+1 {
					t.Errorf("time counters %d, want %d", got, c.before+1)
				}
			}
		})
	}
}

func TestTrekB35KesDoesNotTriggerOnAnOpponentsEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB35Seats(g)
	kes := effectstest.B27Push(g, opp.ID, "Kes, Evolving Telepath", "Legendary Creature — Ocampa Advisor", trekB35KesOracle, "{4}{U}", 2, 6, "U")
	hand := opp.Hand.Size()
	trekB15ToEndStep(t, g)
	if effectstest.CounterOn(g, kes, game.CounterTime) > 0 || opp.Hand.Size() != hand {
		t.Errorf("Kes triggered on %s's end step", me.ID)
	}
}

// --- Lieutenant Nyota Uhura ----------------------------------------------------

// trekB35UhuraBoard is Uhura with three lands of two names, and an
// opponent whose top two cards are an instant above a creature.
func trekB35UhuraBoard(t *testing.T) (g *game.Game, me, opp *game.Player, uhura, instant, creature uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me, opp = trekB35Seats(g)
	effectstest.AdvanceToMain(t, g)
	uhura = effectstest.B27Push(g, me.ID, "Lieutenant Nyota Uhura", "Legendary Creature — Human Officer", trekB35UhuraOracle, "{2}{U}", 2, 3, "U")
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, me.ID, "Island", "Basic Land — Island")
	effectstest.B12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	g.WithWriteLock(func() {
		effectstest.LibraryCardFor(opp, "Deep Filler", "Sorcery")
		creature = effectstest.LibraryCardFor(opp, "Their Bear", "Creature — Bear")
		instant = effectstest.LibraryCardFor(opp, "Their Instant", "Instant")
	})
	return g, me, opp, uhura, instant, creature
}

func TestTrekB35UhuraCastsAnExiledInstantFreeAndBottomsTheRest(t *testing.T) {
	g, me, opp, uhura, instant, creature := trekB35UhuraBoard(t)
	if err := effectstest.HobB4Activate(t, g, me.ID, uhura, 0, "{C}{C}{C}{C}", []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}}); err == nil {
		t.Fatal("Uhura targeted her own controller")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, uhura, 0, "", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, instant)
	if effectstest.HasID(offer.DiscardOptions, creature) {
		t.Error("a creature card was offered as an instant or sorcery spell")
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, instant, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast the instant: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, instant)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(instant) {
		t.Error("the free instant did not resolve into its owner's graveyard")
	}
	if !opp.Library.Contains(creature) || g.Exile.Contains(creature) {
		t.Fatal("the creature card was not put back into its owner's library")
	}
	if bottom := opp.Library.Cards[0].InstanceID; bottom != creature {
		t.Error("the rest did not go to the bottom of the library")
	}
	if g.Exile.Contains(instant) || g.Exile.Contains(creature) {
		t.Error("only two cards were exiled for two differently named lands, and none stays there")
	}
}

func TestTrekB35UhuraDeclinedBottomsEveryExiledCard(t *testing.T) {
	g, me, opp, uhura, instant, creature := trekB35UhuraBoard(t)
	if err := effectstest.HobB4Activate(t, g, me.ID, uhura, 0, "{C}{C}{C}{C}", []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordFreeCast, instant)
	effectstest.DeclineResolutionCast(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{instant, creature} {
		if !opp.Library.Contains(id) {
			t.Error("an exiled card was left in exile")
		}
	}
	top := opp.Library.Cards[len(opp.Library.Cards)-1]
	if top.Name != "Deep Filler" {
		t.Errorf("the top card is %q: the exiled cards should be under the rest", top.Name)
	}
}

// --- On the Shoulders of Giants ------------------------------------------------

func TestTrekB35OnTheShouldersOfGiantsGivesHistoricCreaturesYouControlFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB35Seats(g)
	effectstest.B27Push(g, me.ID, "On the Shoulders of Giants", "Enchantment", trekB35GiantsOracle, "{2}{U}", 0, 0, "U")
	robot := effectstest.B12Creature(g, me.ID, "My Robot", "Artifact Creature — Robot", 1, 1)
	legend := effectstest.B12Creature(g, me.ID, "My Legend", "Legendary Creature — Human", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Robot", "Artifact Creature — Robot", 1, 1)
	for id, want := range map[uuid.UUID]bool{robot: true, legend: true, bear: false, theirs: false} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "flying"); got != want {
			t.Errorf("%s flying = %v, want %v", effectstest.LayeredCard(t, g, id).Name, got, want)
		}
	}
}

// --- Optimization Annex --------------------------------------------------------

func TestTrekB35OptimizationAnnexAddsOneChargeCounter(t *testing.T) {
	g, me, ship, crew := effectstest.StationBoard(t, "The Seriema", "Legendary Artifact — Spacecraft", effectstest.TheSeriemaOracle, 2)
	effectstest.B27Push(g, me.ID, "Optimization Annex", "Artifact Creature — Borg", trekB35AnnexOracle, "{1}{U}", 1, 3, "U")
	effectstest.Station(t, g, me, ship, crew)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, ship, game.CounterCharge); got != 3 {
		t.Errorf("a 2-power station under the Annex: %d charge counters, want 3", got)
	}
}

func TestTrekB35OptimizationAnnexIgnoresOtherCountersAndOpponentsArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB35Seats(g)
	effectstest.B27Push(g, me.ID, "Optimization Annex", "Artifact Creature — Borg", trekB35AnnexOracle, "{1}{U}", 1, 3, "U")
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(mine, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(theirs, game.CounterCharge, 1)
		_ = g.AddCounterForEffect(mine, game.CounterCharge, 2)
	})
	if got := effectstest.CounterOn(g, mine, game.CounterPlusOne); got != 1 {
		t.Errorf("a +1/+1 counter became %d", got)
	}
	if got := effectstest.CounterOn(g, theirs, game.CounterCharge); got != 1 {
		t.Errorf("an opponent's artifact got %d charge counters, want 1", got)
	}
	if got := effectstest.CounterOn(g, mine, game.CounterCharge); got != 3 {
		t.Errorf("two charge counters on your artifact became %d, want 3", got)
	}
}

// --- Sam Kirk, Xenoanthropologist ----------------------------------------------

func TestTrekB35SamKirkInvestigatesWhenAnOpponentCastsACreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Sam Kirk, Xenoanthropologist", "Legendary Creature — Human Scientist", trekB35SamKirkOracle, "{3}{U}", 3, 3, "U")
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.CastWithCost(t, g, "Their Instant", "Instant", "", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(trekB15Tokens(g, me.ID, "Clue")); n != 0 {
		t.Fatalf("an instant made %d Clues", n)
	}
	effectstest.CastWithCost(t, g, "Their Bear", "Creature — Bear", "", "")
	effectstest.PassPriorityAroundTable(t, g)
	clues := trekB15Tokens(g, me.ID, "Clue")
	if len(clues) != 1 || !clues[0].IsArtifact() {
		t.Fatalf("an opponent's creature spell made %d Clues, want 1", len(clues))
	}
}

func TestTrekB35SamKirkIgnoresYourOwnCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B27Push(g, me.ID, "Sam Kirk, Xenoanthropologist", "Legendary Creature — Human Scientist", trekB35SamKirkOracle, "{3}{U}", 3, 3, "U")
	effectstest.AdvanceToMainOf(t, g, 0)
	effectstest.CastWithCost(t, g, "My Bear", "Creature — Bear", "", "")
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(trekB15Tokens(g, me.ID, "Clue")); n != 0 {
		t.Errorf("your own creature spell made %d Clues", n)
	}
}

// --- Seven of Nine, Tertiary Adjunct -------------------------------------------

func TestTrekB35SevenOfNineChargesAnArtifactOnANonlandDiscard(t *testing.T) {
	for _, c := range []struct {
		name    string
		discard string
		charges bool
	}{
		{"a nonland card", "Instant", true},
		{"a land", "Basic Land — Island", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, _ := trekB35Seats(g)
			effectstest.AdvanceToMain(t, g)
			effectstest.TrekB35Library(g, me, 3)
			seven := effectstest.B27Push(g, me.ID, "Seven of Nine, Tertiary Adjunct", "Legendary Artifact Creature — Borg", trekB35SevenOracle, "{2}{U}", 1, 4, "U")
			rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
			pitch := effectstest.HandCardOf(g, me, "Pitch", c.discard, "", "")
			hand := me.Hand.Size()
			if err := effectstest.HobB4Activate(t, g, me.ID, seven, 0, "", nil); err != nil {
				t.Fatalf("activate: %v", err)
			}
			if !effectstest.LayeredCard(t, g, seven).Tapped {
				t.Error("the {T} cost did not tap Seven")
			}
			effectstest.PassPriorityAroundTable(t, g)
			if me.Hand.Size() != hand+1 {
				t.Fatalf("hand %d after the draw, want %d", me.Hand.Size(), hand+1)
			}
			effectstest.AnswerDiscard(t, g, me.ID, pitch)
			effectstest.PassPriorityAroundTable(t, g)
			if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
				if !c.charges {
					t.Fatal("a land discard made the reflexive trigger")
				}
				effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(rock)...)
				effectstest.PassPriorityAroundTable(t, g)
			}
			want := 0
			if c.charges {
				want = 2
			}
			if got := effectstest.CounterOn(g, rock, game.CounterCharge); got != want {
				t.Errorf("%d charge counters on the artifact, want %d", got, want)
			}
			if me.Hand.Size() != hand {
				t.Errorf("hand %d after the loot, want %d", me.Hand.Size(), hand)
			}
		})
	}
}

func TestTrekB35SevenOfNineMayChargeNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB35Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.TrekB35Library(g, me, 3)
	seven := effectstest.B27Push(g, me.ID, "Seven of Nine, Tertiary Adjunct", "Legendary Artifact Creature — Borg", trekB35SevenOracle, "{2}{U}", 1, 4, "U")
	pitch := effectstest.HandCardOf(g, me, "Pitch", "Instant", "", "")
	if err := effectstest.HobB4Activate(t, g, me.ID, seven, 0, "", nil); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, me.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.PickTargetRefs(t, g, me.ID, nil)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CounterOn(g, seven, game.CounterCharge); got != 0 {
		t.Errorf("no target chosen, but Seven has %d charge counters", got)
	}
	if !effectstest.StackFullyEmpty(g) {
		t.Error("the stack did not settle")
	}
}

// --- Temporal Tribulation ------------------------------------------------------

func TestTrekB35TemporalTribulationBlinksTwoPermanentsAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB35Seats(g)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	spell := effectstest.HandCardFull(me, "Temporal Tribulation", "Instant", "{2}{U}", trekB35TribulationOrcl, []string{"U"})
	effectstest.B06AddMana(me, "C", "C", "U")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(bear, theirs)}); err == nil {
		t.Fatal("targeted a permanent an opponent controls")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("cast with one target")
	}
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(bear, rock)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if trekB35CountNamed(g, "My Bear") != 0 || trekB35CountNamed(g, "My Rock") != 0 {
		t.Fatal("the targets were not exiled")
	}
	trekB15ToEndStep(t, g)
	if trekB35CountNamed(g, "My Bear") != 1 || trekB35CountNamed(g, "My Rock") != 1 {
		t.Error("the exiled cards did not return at the next end step")
	}
}

func TestTrekB35TemporalTribulationFlashesBackForFiveAndBlue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB35Seats(g)
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	rock := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Temporal Tribulation", TypeLine: "Instant", ManaCost: "{2}{U}",
			OracleID: trekB35TribulationOrcl, Owner: me.ID, Controller: me.ID})
		me.ManaPool.EmptyPool()
	})
	effectstest.FillPoolColored(me, "U", 5)
	params := game.CastSpellParams{FromZone: string(game.ZoneGraveyard), AlternativeCost: "flashback", Strict: true, Targets: effectstest.CardRefs(bear, rock)}
	if err := g.CastSpell(me.ID, id, params); err == nil {
		t.Fatal("flashback {5}{U} was paid with five mana")
	}
	effectstest.FillPoolColored(me, "U", 1)
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("flashback cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) || me.Graveyard.Contains(id) {
		t.Error("the flashed-back spell was not exiled")
	}
	if trekB35CountNamed(g, "My Bear") != 0 {
		t.Error("the flashback did not exile the targets")
	}
}

// --- Transwarp Hub -------------------------------------------------------------

func TestTrekB35TranswarpHubMillsFiveAndTakesUpToTwoArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB35Seats(g)
	effectstest.AdvanceToMain(t, g)
	var gears [3]uuid.UUID
	var filler uuid.UUID
	g.WithWriteLock(func() {
		effectstest.LibraryCardFor(me, "Deep Gear", "Artifact") // sixth from the top: not milled
		gears[0] = effectstest.LibraryCardFor(me, "Gear A", "Artifact")
		filler = effectstest.LibraryCardFor(me, "Milled Bear", "Creature — Bear")
		gears[1] = effectstest.LibraryCardFor(me, "Gear B", "Artifact")
		effectstest.LibraryCardFor(me, "Milled Land", "Basic Land — Island")
		gears[2] = effectstest.LibraryCardFor(me, "Gear C", "Artifact Creature — Robot")
	})
	effectstest.CastWithCost(t, g, "Transwarp Hub", "Artifact — Spacecraft", "{2}{U}", trekB35HubOracle)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.LatestChooseCardsFor(g, me.ID)
	if c == nil {
		t.Fatal("no pick from among the milled cards")
	}
	for _, opt := range c.ChooseCards {
		if opt == filler {
			t.Error("a nonartifact card was offered")
		}
	}
	if len(c.ChooseCards) != 3 {
		t.Errorf("%d cards offered, want the three milled artifacts", len(c.ChooseCards))
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, gears[:]); err == nil {
		t.Fatal("took three artifact cards")
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, gears[:2]); err != nil {
		t.Fatalf("take two: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(gears[0]) || !me.Hand.Contains(gears[1]) || !me.Graveyard.Contains(gears[2]) {
		t.Error("the two picked artifacts are not in hand, or the third left the graveyard")
	}
	if !me.Graveyard.Contains(filler) {
		t.Error("the milled creature card is not in the graveyard")
	}
}

func TestTrekB35TranswarpHubMakesACreatureUnblockableAtThreeCharge(t *testing.T) {
	g, me, hub, crew := effectstest.StationBoard(t, "Transwarp Hub", "Artifact — Spacecraft", trekB35HubOracle, 2)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	if err := effectstest.HobB4Activate(t, g, me.ID, hub, 1, "", effectstest.CardRefs(crew)); err == nil {
		t.Fatal("the 3+ ability was activated with no charge counters")
	}
	effectstest.TrekB35SetCounters(g, hub, game.CounterCharge, 3)
	if effectstest.IsCreatureNow(g, hub) {
		t.Error("the Hub has no P/T box and should never be a creature")
	}
	if err := effectstest.HobB4Activate(t, g, me.ID, hub, 1, "", effectstest.CardRefs(crew)); err != nil {
		t.Fatalf("activate at three counters: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	blocker := effectstest.PushVanillaCreature(g, opp.ID, "Blocker", 2, 2)
	other := effectstest.PushVanillaCreature(g, me.ID, "Other", 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{crew, other} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if g.DeclareBlocker(blocker, crew) == nil {
		t.Error("the Hub's target was blocked")
	}
	if g.DeclareBlocker(blocker, other) != nil {
		t.Error("a creature the Hub did not choose could not be blocked")
	}
}

// --- Tuvok, Chief Tactical Officer ---------------------------------------------

func TestTrekB35TuvokReselectsUpToTwoAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	attackerSeat, me, second, third := g.Seats[0], g.Seats[1], g.Seats[2], g.Seats[3]
	a := effectstest.PushVanillaCreature(g, attackerSeat.ID, "Charger A", 2, 2)
	b := effectstest.PushVanillaCreature(g, attackerSeat.ID, "Charger B", 2, 2)
	effectstest.AttackInDeclareAttackers(t, g, map[uuid.UUID]uuid.UUID{a: me.ID, b: me.ID})
	effectstest.CastFlashPermanentInPlace(t, g, me.ID, "Tuvok, Chief Tactical Officer", "Legendary Creature — Vulcan Officer", trekB35TuvokOracle)
	for range 4 {
		if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.LatestOptionPickFor(g, me.ID) != nil {
			break
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	effectstest.PickTargetRefs(t, g, me.ID, nil, effectstest.CardRefs(a, b)...)
	first := effectstest.SettleUntilOptionPick(t, g, me.ID, uuid.Nil)
	if first == nil {
		t.Fatal("no reselect prompt")
	}
	if err := g.ResolveOptionPick(first.ID, me.ID, effectstest.OptionFor(t, first, second.ID)); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	next := effectstest.LatestOptionPickFor(g, me.ID)
	if next == nil || next.ID == first.ID {
		t.Fatal("no second prompt for the second target")
	}
	if err := g.ResolveOptionPick(next.ID, me.ID, effectstest.OptionFor(t, next, third.ID)); err != nil {
		t.Fatalf("ResolveOptionPick: %v", err)
	}
	ca, _ := effectstest.AangCardOnBF(g, a)
	cb, _ := effectstest.AangCardOnBF(g, b)
	got := map[uuid.UUID]bool{ca.AttackingTarget: true, cb.AttackingTarget: true}
	if !got[second.ID] || !got[third.ID] {
		t.Errorf("attacks ended on %v and %v, want seats 2 and 3 one each", ca.AttackingTarget, cb.AttackingTarget)
	}
	if !effectstest.HasEffectiveKeyword(t, g, trekB35Named(t, g, "Tuvok, Chief Tactical Officer"), "flash") {
		t.Error("Tuvok lacks flash")
	}
}

func TestTrekB35TuvokOutsideDeclareAttackersDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFlashPermanentInPlace(t, g, me.ID, "Tuvok, Chief Tactical Officer", "Legendary Creature — Vulcan Officer", trekB35TuvokOracle)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.LatestOptionPickFor(g, me.ID) != nil {
		t.Error("Tuvok triggered outside the declare attackers step")
	}
	if trekB35CountNamed(g, "Tuvok, Chief Tactical Officer") != 1 {
		t.Error("Tuvok did not resolve")
	}
}

// trekB35Named is the battlefield permanent with the name.
func trekB35Named(t *testing.T, g *game.Game, name string) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name {
				id = c.InstanceID
			}
		}
	})
	if id == uuid.Nil {
		t.Fatalf("no %s on the battlefield", name)
	}
	return id
}

// --- Võx, Child of the Collective ----------------------------------------------

func TestTrekB35VoxGivesTheNextArtifactCreatureAndSpacecraftSpellsUndaunted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB35Seats(g)
	artifactCreature := game.Card{Name: "Test Drone", TypeLine: "Artifact Creature — Borg", ManaCost: "{4}{U}", Owner: me.ID}
	ship := game.Card{Name: "Test Ship", TypeLine: "Artifact — Spacecraft", ManaCost: "{5}", Owner: me.ID}
	bear := game.Card{Name: "Test Bear", TypeLine: "Creature — Bear", ManaCost: "{4}{G}", Owner: me.ID}
	if p := effectstest.NsPrice(t, g, me, artifactCreature); p.Generic != 4 {
		t.Fatalf("before Võx: %+v", p)
	}
	effectstest.CastHoldCreature(t, g, "Võx, Child of the Collective", "Legendary Artifact Creature — Borg", trekB35VoxOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	opps := len(g.Seats) - 1
	if p := effectstest.NsPrice(t, g, me, artifactCreature); p.Generic != 4-opps {
		t.Errorf("an artifact creature spell: %+v, want {%d}{U}", p, 4-opps)
	}
	if p := effectstest.NsPrice(t, g, me, ship); p.Generic != 5-opps {
		t.Errorf("a Spacecraft spell: %+v, want {%d}", p, 5-opps)
	}
	if p := effectstest.NsPrice(t, g, me, bear); p.Generic != 4 {
		t.Errorf("a nonartifact creature spell got undaunted: %+v", p)
	}
	if n := len(effectstest.NsPromises(me)); n != 2 {
		t.Fatalf("%d promises, want 2", n)
	}
	effectstest.CastWithCost(t, g, "Test Drone", "Artifact Creature — Borg", "", "")
	got := effectstest.NsPromises(me)
	if len(got) != 1 || got[0].Filter.Subtype != "Spacecraft" {
		t.Errorf("the artifact creature spell should spend only its own promise: %+v", got)
	}
}

func TestTrekB35VoxUndauntedSkipsAPlayerWhoLeftTheGame(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := trekB35Seats(g)
	effectstest.CastHoldCreature(t, g, "Võx, Child of the Collective", "Legendary Artifact Creature — Borg", trekB35VoxOracle, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { opp.Eliminated = true })
	ship := game.Card{Name: "Test Ship", TypeLine: "Artifact — Spacecraft", ManaCost: "{5}", Owner: me.ID}
	if p := effectstest.NsPrice(t, g, me, ship); p.Generic != 5-(len(g.Seats)-2) {
		t.Errorf("a Spacecraft spell with one opponent gone: %+v", p)
	}
}

func TestTrekB35VoxTriggersAgainAtYourFirstMainPhase(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _ := trekB35Seats(g)
	seat := g.Turn.ActiveSeat
	effectstest.B27Push(g, me.ID, "Võx, Child of the Collective", "Legendary Artifact Creature — Borg", trekB35VoxOracle, "{4}{U}", 3, 3, "U")
	effectstest.AdvanceToMainOf(t, g, (seat+1)%len(g.Seats))
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.NsPromises(me)); n != 0 {
		t.Fatalf("an opponent's main phase gave %d promises", n)
	}
	effectstest.AdvanceToMainOf(t, g, seat)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.NsPromises(me)); n != 2 {
		t.Errorf("your first main phase gave %d promises, want 2", n)
	}
}

package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_mom_b14_test.go — March of the Machine Commander, S82 batch 14
// (#5477). Each test plays the real catalog entry and asserts what its
// text says.

const (
	s82B14GuardianOracle  = "2553ab57-ceb6-4654-a59d-bc61735da908"
	s82B14VanisherOracle  = "5d11ada1-6459-4a91-bd10-66673dc1ad3c"
	s82B14SealOracle      = "72c4c682-df97-4408-997b-e849f54aba39"
	s82B14SovereignOracle = "b08cb220-34fe-4c54-b654-64a3c24dba44"
	s82B14NegotiatorOracl = "af14101a-9cc0-4bd3-a9f9-2726353e40ff"
	s82B14RidersOracle    = "461c0c37-7213-4ece-91f7-934e6a1a8d64"
	s82B14GimbalOracle    = "e7441404-71ff-449e-812b-18cd52eabb98"
	s82B14MantleOracle    = "21b520c6-c94a-475a-a824-dd96394ea55b"
	s82B14DistributorOrac = "435df16b-3dd4-446e-a643-f9478aaa47c0"
)

func TestS82B14EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82B14GuardianOracle:  "Guardian of Ghirapur",
		s82B14VanisherOracle:  "Mistmeadow Vanisher",
		s82B14SealOracle:      "Seal from Existence",
		s82B14SovereignOracle: "Infernal Sovereign",
		s82B14NegotiatorOracl: "Cutthroat Negotiator",
		s82B14RidersOracle:    "Sandsteppe War Riders",
		s82B14GimbalOracle:    "Gimbal, Gremlin Prodigy",
		s82B14MantleOracle:    "Conjurer's Mantle",
		s82B14DistributorOrac: "Pain Distributor",
	} {
		effectstest.CmdrB21AssertFull(t, oracle, name)
	}
}

// --- Guardian of Ghirapur ---------------------------------------------------

func TestS82B14GuardianBlinksAnotherCreatureOrArtifactYouControl(t *testing.T) {
	for _, tc := range []struct{ Name, TypeLine string }{
		{"Bear", "Creature — Bear"},
		{"Trinket", "Artifact"},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			mine := effectstest.B12Permanent(g, me.ID, tc.Name, tc.TypeLine)
			theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
			guardian := effectstest.CastHoldCreature(t, g, "Guardian of Ghirapur", "Creature — Angel", s82B14GuardianOracle, 3, 3)
			ch := effectstest.PickTargetPrompt(t, g)
			if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: guardian}); err == nil {
				t.Fatal("the Guardian itself was accepted; the target is another permanent")
			}
			if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
				t.Fatal("an opponent's creature was accepted; the target is one you control")
			}
			effectstest.AnswerPickTarget(t, g, mine)
			effectstest.PassPriorityAroundTable(t, g)
			if g.Battlefield.Contains(mine) || effectstest.B12ZoneOf(g, mine) != game.ZoneExile {
				t.Fatal("the target was not exiled")
			}
			effectstest.AdvanceToEndStep(t, g)
			effectstest.PassPriorityAroundTable(t, g)
			if !effectstest.OnBattlefieldByName(g, tc.Name) {
				t.Error("the target did not return at the beginning of the next end step")
			}
		})
	}
}

func TestS82B14GuardianWithNoTargetExilesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B12Push(g, me.ID, "Bear", "Creature — Bear", "", 2, 2)
	effectstest.CastHoldCreature(t, g, "Guardian of Ghirapur", "Creature — Angel", s82B14GuardianOracle, 3, 3)
	ch := effectstest.PickTargetPrompt(t, g)
	if err := g.ResolvePickTargets(ch.ID, ch.Chooser, nil); err != nil {
		t.Fatalf("declining the up-to-one target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Error("the bear left the battlefield with no target chosen")
	}
}

// --- Mistmeadow Vanisher ----------------------------------------------------

func TestS82B14VanisherExilesANontokenPermanentWhenItAttacksAndItReturns(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vanisher := effectstest.B12Push(g, me.ID, "Mistmeadow Vanisher", "Creature — Kithkin Wizard", s82B14VanisherOracle, 3, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	var token uuid.UUID
	g.WithWriteLock(func() {
		ids, err := g.CreateTokensForEffect(opp.ID, effects.TokenCard("1/1 white Soldier"), 1, game.TokenEntryOptions{})
		if err != nil || len(ids) != 1 {
			t.Fatalf("token: %v %v", ids, err)
		}
		token = ids[0]
	})
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	effectstest.DeclareAttack(t, g, opp.ID, vanisher)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	for _, bad := range []uuid.UUID{token, land} {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatal("a token or a land was accepted; the target is a nonland, nontoken permanent")
		}
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, theirs) != game.ZoneExile {
		t.Fatal("the target was not exiled")
	}
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByName(g, "Their Bear") {
		t.Error("the bear did not return at the beginning of the next end step")
	}
}

func TestS82B14VanisherIsQuietBeforeItTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Mistmeadow Vanisher", "Creature — Kithkin Wizard", s82B14VanisherOracle, 3, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePickTarget {
			t.Fatal("a target prompt is open before the Vanisher has tapped")
		}
	}
	if !g.Battlefield.Contains(theirs) {
		t.Error("the bear left the battlefield")
	}
}

// --- Seal from Existence ----------------------------------------------------

func TestS82B14SealExilesUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Their Forest", "Basic Land — Forest")
	theirs := effectstest.B12Push(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2)
	seal := effectstest.CastCatalogSpell(t, g, "Seal from Existence", "Enchantment", s82B14SealOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.PickTargetPrompt(t, g)
	for _, bad := range []uuid.UUID{mine, land} {
		if err := g.ResolvePickTarget(ch.ID, ch.Chooser, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatal("your own permanent or a land was accepted")
		}
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, theirs) != game.ZoneExile {
		t.Fatal("the bear was not exiled")
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, seal); err != nil {
		t.Fatalf("remove the Seal: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByName(g, "Their Bear") {
		t.Error("the bear did not return when the Seal left")
	}
}

func TestS82B14SealCarriesWardThree(t *testing.T) {
	spec, _ := effects.Lookup(s82B14SealOracle)
	found := false
	for _, tr := range spec.Triggered {
		if tr.Key == "Seal from Existence — ward {3}" {
			found = true
		}
	}
	if !found {
		t.Error("the Seal has no ward {3} trigger")
	}
}

// --- Infernal Sovereign -----------------------------------------------------

func TestS82B14SovereignDrawsAndLosesLifeOnALandPlayAndOnACast(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Infernal Sovereign", "Creature — Demon", s82B14SovereignOracle, 6, 6)
	hand, life := me.Hand.Size(), me.Life
	effectstest.B12PlayFromHand(t, g, "Forest", "Basic Land — Forest", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	// The land left the hand, a card was drawn.
	if me.Hand.Size() != hand+1 || me.Life != life-1 {
		t.Fatalf("after a land play: hand %d→%d life %d→%d, want one card drawn and -1 life", hand, me.Hand.Size(), life, me.Life)
	}
	effectstest.B12PlayFromHand(t, g, "Idea", "Sorcery", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || me.Life != life-2 {
		t.Errorf("after a cast: hand %d→%d life %d→%d, want two cards drawn and -2 life in all", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestS82B14SovereignIgnoresALandThatIsNotPlayedAndAnOpponentsSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Infernal Sovereign", "Creature — Demon", s82B14SovereignOracle, 6, 6)
	hand, life := me.Hand.Size(), me.Life
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	effectstest.B13OpponentCasts(t, g, opp, "Their Spell", "Instant", "", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand || me.Life != life {
		t.Errorf("hand %d→%d life %d→%d, want no change", hand, me.Hand.Size(), life, me.Life)
	}
}

func TestS82B14SovereignSkipsOnlyItsControllersDrawStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[1]
	effectstest.B12Push(g, owner.ID, "Infernal Sovereign", "Creature — Demon", s82B14SovereignOracle, 6, 6)
	handBefore := owner.Hand.Size()
	trace := effectstest.WalkSteps(t, g, 200, func() bool {
		return g.Turn.ActiveSeat == 2 && g.Turn.Step == game.StepPrecombatMain
	})
	if effectstest.SawStep(trace, 1, game.StepDraw) {
		t.Error("the Sovereign's controller still had a draw step")
	}
	if owner.Hand.Size() != handBefore {
		t.Errorf("the Sovereign's controller drew %d cards", owner.Hand.Size()-handBefore)
	}
	if !effectstest.SawStep(trace, 2, game.StepDraw) {
		t.Error("another player's draw step was skipped")
	}
}

// --- Cutthroat Negotiator ---------------------------------------------------

func TestS82B14NegotiatorMakesATappedTreasurePerNonlandAndEveryoneDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.S49Library(g, 4)
	effectstest.S49TopCards(g, 0, 1, 3)
	me := g.Seats[0]
	neg := effectstest.PushCatalogPermanent(g, me.ID, "Cutthroat Negotiator", "Creature — Orc Pirate", s82B14NegotiatorOracl, false)
	hands := effectstest.S49Hands(g)
	effectstest.S49Attack(t, g, neg)
	if n := effectstest.CountOnBattlefield(g, "Treasure", me.ID); n != 3 {
		t.Fatalf("Treasures = %d, want 3", n)
	}
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Treasure" && c.Controller == me.ID && !c.Tapped {
				t.Error("a Treasure entered untapped")
			}
		}
	})
	for i, h := range effectstest.S49Hands(g) {
		if h != hands[i]+1 {
			t.Errorf("seat %d hand %d, want %d (each player draws)", i, h, hands[i]+1)
		}
	}
}

func TestS82B14NegotiatorMakesNothingWhenOnlyLandsAreRevealed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.S49Library(g, 4)
	effectstest.S49TopCards(g)
	me := g.Seats[0]
	neg := effectstest.PushCatalogPermanent(g, me.ID, "Cutthroat Negotiator", "Creature — Orc Pirate", s82B14NegotiatorOracl, false)
	effectstest.S49Attack(t, g, neg)
	if n := effectstest.CountOnBattlefield(g, "Treasure", me.ID); n != 0 {
		t.Errorf("Treasures = %d, want 0", n)
	}
}

// --- Sandsteppe War Riders --------------------------------------------------

func s82B14Tokens(g *game.Game, owner uuid.UUID, tmpl game.Card, n int) {
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(owner, tmpl, n) })
}

func TestS82B14RidersBolsterByDifferentlyNamedArtifactTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	riders := effectstest.B12Push(g, me.ID, "Sandsteppe War Riders", "Creature — Human Warrior", s82B14RidersOracle, 4, 4)
	small := effectstest.B12Push(g, me.ID, "Small", "Creature — Bear", "", 2, 2)
	// Two Treasures count once, a Food once: X = 2. An opponent's token
	// is not counted.
	s82B14Tokens(g, me.ID, effects.TreasureToken(), 2)
	s82B14Tokens(g, me.ID, effects.FoodToken(), 1)
	s82B14Tokens(g, opp.ID, effects.ClueToken(), 1)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, small, game.CounterPlusOne); got != 2 {
		t.Errorf("the least-toughness creature has %d counters, want 2", got)
	}
	if got := effectstest.CountersOn(g, riders, game.CounterPlusOne); got != 0 {
		t.Errorf("the Riders have %d counters, want 0", got)
	}
}

func TestS82B14RidersBolsterNothingWithoutArtifactTokens(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Sandsteppe War Riders", "Creature — Human Warrior", s82B14RidersOracle, 4, 4)
	small := effectstest.B12Push(g, me.ID, "Small", "Creature — Bear", "", 2, 2)
	effectstest.B12Permanent(g, me.ID, "Card Rock", "Artifact")
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, small, game.CounterPlusOne); got != 0 {
		t.Errorf("%d counters, want 0: a nontoken artifact does not count", got)
	}
}

// --- Gimbal, Gremlin Prodigy ------------------------------------------------

func s82B14Gremlins(g *game.Game, owner uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Gremlin" && c.Controller == owner && c.IsToken() {
				out = append(out, c.InstanceID)
			}
		}
	})
	return out
}

func TestS82B14GimbalMakesAGremlinWithCountersAtYourEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Gimbal, Gremlin Prodigy", "Legendary Creature — Gremlin Artificer", s82B14GimbalOracle, 4, 4)
	// Treasure and Food, plus the new Gremlin: three names.
	s82B14Tokens(g, me.ID, effects.TreasureToken(), 2)
	s82B14Tokens(g, me.ID, effects.FoodToken(), 1)
	effectstest.AdvanceToEndStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	gs := s82B14Gremlins(g, me.ID)
	if len(gs) != 1 {
		t.Fatalf("Gremlins = %d, want 1", len(gs))
	}
	if got := effectstest.CountersOn(g, gs[0], game.CounterPlusOne); got != 3 {
		t.Errorf("the Gremlin has %d counters, want 3 (it counts itself)", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, gs[0], "trample") {
		t.Error("the Gremlin is an artifact creature and should have trample")
	}
}

func TestS82B14GimbalGivesTrampleToArtifactCreaturesYouControlOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	gimbal := effectstest.B12Push(g, me.ID, "Gimbal, Gremlin Prodigy", "Legendary Creature — Gremlin Artificer", s82B14GimbalOracle, 4, 4)
	mine := effectstest.B12Push(g, me.ID, "My Golem", "Artifact Creature — Golem", "", 2, 2)
	plain := effectstest.B12Push(g, me.ID, "My Bear", "Creature — Bear", "", 2, 2)
	theirs := effectstest.B12Push(g, opp.ID, "Their Golem", "Artifact Creature — Golem", "", 2, 2)
	for id, want := range map[uuid.UUID]bool{mine: true, plain: false, theirs: false, gimbal: false} {
		if got := effectstest.HasEffectiveKeyword(t, g, id, "trample"); got != want {
			t.Errorf("trample on %s = %v, want %v", id, got, want)
		}
	}
}

func TestS82B14GimbalIsQuietOnAnOpponentsEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Gimbal, Gremlin Prodigy", "Legendary Creature — Gremlin Artificer", s82B14GimbalOracle, 4, 4)
	effectstest.WalkSteps(t, g, 200, func() bool { return g.Turn.ActiveSeat == 1 && g.Turn.Step == game.StepPrecombatMain })
	effectstest.PassPriorityAroundTable(t, g)
	before := len(s82B14Gremlins(g, me.ID))
	effectstest.WalkSteps(t, g, 200, func() bool { return g.Turn.ActiveSeat == 1 && g.Turn.Step == game.StepEnd })
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(s82B14Gremlins(g, me.ID)); n != before {
		t.Errorf("Gremlins %d→%d across an opponent's end step, want no change", before, n)
	}
}

// --- Conjurer's Mantle ------------------------------------------------------

func TestS82B14MantlePumpsGivesVigilanceAndDigsForASharedCreatureType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mantle := effectstest.PushCatalogPermanent(g, me.ID, "Conjurer's Mantle", "Artifact — Equipment", s82B14MantleOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly", "Creature — Bear Warrior", 2, 2)
	effectstest.BrAttach(t, g, mantle, bear)
	if p, tough := effectstest.EffectivePower(t, g, bear), effectstest.EffectiveToughness(t, g, bear); p != 3 || tough != 3 {
		t.Errorf("equipped creature %d/%d, want 3/3", p, tough)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "vigilance") {
		t.Error("equipped creature lacks vigilance")
	}
	g.WithWriteLock(func() { me.Library.Cards = nil })
	lib := effectstest.TrekB48Library(me,
		game.Card{Name: "Elf", TypeLine: "Creature — Elf"},
		game.Card{Name: "Kin Spell", TypeLine: "Kindred Instant — Warrior"},
		game.Card{Name: "Cub", TypeLine: "Creature — Bear"},
		game.Card{Name: "Four", TypeLine: "Instant"},
		game.Card{Name: "Five", TypeLine: "Instant"},
		game.Card{Name: "Six", TypeLine: "Instant"},
		game.Card{Name: "Seventh", TypeLine: "Creature — Bear"},
	)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	ch := effectstest.LatestChooseCards(g, me.ID)
	if ch == nil {
		t.Fatal("no pick offered")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[0]}); err == nil {
		t.Error("an Elf was accepted; it shares no creature type with a Bear Warrior")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[6]}); err == nil {
		t.Error("the seventh card (not looked at) was accepted")
	}
	if err := g.ResolveChooseCards(ch.ID, me.ID, []uuid.UUID{lib[1]}); err != nil {
		t.Fatalf("the Kindred Warrior spell should be a legal pick: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(lib[1]) {
		t.Error("the chosen card is not in hand")
	}
	if n := len(me.Library.Cards); n != 6 || me.Library.Cards[n-1].InstanceID != lib[6] {
		t.Errorf("library has %d cards; want the seventh card on top and the five unchosen looked-at cards beneath it", n)
	}
}

func TestS82B14MantleWithNoMatchPutsEverythingOnTheBottom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mantle := effectstest.PushCatalogPermanent(g, me.ID, "Conjurer's Mantle", "Artifact — Equipment", s82B14MantleOracle, false)
	bear := effectstest.B12Creature(g, me.ID, "Grizzly", "Creature — Bear", 2, 2)
	effectstest.BrAttach(t, g, mantle, bear)
	g.WithWriteLock(func() { me.Library.Cards = nil })
	lib := effectstest.TrekB48Library(me,
		game.Card{Name: "Elf", TypeLine: "Creature — Elf"},
		game.Card{Name: "Spell", TypeLine: "Instant"},
		game.Card{Name: "Three", TypeLine: "Instant"},
		game.Card{Name: "Four", TypeLine: "Instant"},
		game.Card{Name: "Five", TypeLine: "Instant"},
		game.Card{Name: "Six", TypeLine: "Instant"},
		game.Card{Name: "Seventh", TypeLine: "Creature — Bear"},
	)
	hand := me.Hand.Size()
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d→%d, want no card taken", hand, me.Hand.Size())
	}
	if n := len(me.Library.Cards); n != 7 || me.Library.Cards[n-1].InstanceID != lib[6] {
		t.Errorf("library has %d cards; want the seventh card on top and the six looked at beneath it", n)
	}
}

// --- Pain Distributor -------------------------------------------------------

func TestS82B14DistributorGivesEachPlayersFirstSpellATreasureForThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Pain Distributor", "Creature — Devil Citizen", s82B14DistributorOrac, 2, 3)
	effectstest.B13OpponentCasts(t, g, opp, "Their First", "Instant", "", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Treasure", opp.ID); n != 1 {
		t.Fatalf("the opponent has %d Treasures after their first spell, want 1", n)
	}
	if n := effectstest.CountOnBattlefield(g, "Treasure", me.ID); n != 0 {
		t.Fatalf("you have %d Treasures from their spell, want 0", n)
	}
	effectstest.B13OpponentCasts(t, g, opp, "Their Second", "Instant", "", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Treasure", opp.ID); n != 1 {
		t.Errorf("the opponent has %d Treasures after a second spell, want still 1", n)
	}
	effectstest.B12PlayFromHand(t, g, "My First", "Sorcery", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CountOnBattlefield(g, "Treasure", me.ID); n != 1 {
		t.Errorf("you have %d Treasures after your first spell, want 1", n)
	}
}

func TestS82B14DistributorPingsTheControllerOfAnOpponentsArtifactThatGoesToAGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Pain Distributor", "Creature — Devil Citizen", s82B14DistributorOrac, 2, 3)
	theirs := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	mine := effectstest.B12Permanent(g, me.ID, "My Rock", "Artifact")
	theirBear := effectstest.B12Permanent(g, opp.ID, "Their Statue", "Enchantment")
	life, myLife := opp.Life, me.Life
	for _, id := range []uuid.UUID{mine, theirBear} {
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: g.Seats[0].ID}, id); err != nil {
			t.Fatalf("move: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life || me.Life != myLife {
		t.Fatalf("life moved (opp %d→%d, me %d→%d) for your own artifact or a nonartifact", life, opp.Life, myLife, me.Life)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: opp.ID}, theirs); err != nil {
		t.Fatalf("move: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d→%d, want 1 damage", life, opp.Life)
	}
}

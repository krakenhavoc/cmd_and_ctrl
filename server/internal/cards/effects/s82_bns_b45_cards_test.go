package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b45_cards_test.go — S82, Foundations Jumpstart batch 45 (#5929).

const (
	s82bns45GloryOracle    = "8f446660-f72a-454f-bc96-3b6dd8633228"
	s82bns45NaiadOracle    = "38b4798d-47cc-4ffd-8b0b-e90b5da2ee2d"
	s82bns45LionfishOracle = "ffd6c4b8-928d-40f1-9549-d13fa0b3af82"
	s82bns45AbandonOracle  = "fa43793d-a215-4a64-a09c-ed9799f745a4"
	s82bns45GiftOracle     = "4e698713-6203-44f2-b438-1b26b951536d"
	s82bns45BigPlayOracle  = "607425ac-9e90-4c9e-88bf-72f8a76cce97"
	s82bns45BlinkOracle    = "51acd176-e84d-42c2-acb2-cceeb28a9fec"
	s82bns45BruteOracle    = "b7a16caa-96e0-40c5-8fcb-81e06d26f92a"
	s82bns45BuryOracle     = "d5e67016-1740-42f8-8be2-9eb9d4adb142"
	s82bns45ChillingOracle = "7ac49129-7da2-4729-b48c-b6da279c212e"
)

func s82bns45Hand(p *game.Player, name, typeLine, cost, oracle string) uuid.UUID {
	id := uuid.New()
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

func s82bns45Ref(id uuid.UUID) []game.TargetRef {
	return []game.TargetRef{{Kind: game.TargetCard, ID: id}}
}

func s82bns45Cast(t *testing.T, g *game.Game, p *game.Player, name, cost, oracle string, targets []game.TargetRef, pool ...string) {
	t.Helper()
	id := s82bns45Hand(p, name, "Instant", cost, oracle)
	for _, c := range pool {
		fillPoolColored(p, c, 1)
	}
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{Strict: true, Targets: targets}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
}

func TestS82Bns45CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		s82bns45GloryOracle: "Glory Bearers", s82bns45NaiadOracle: "Naiad of Hidden Coves",
		s82bns45LionfishOracle: "Stinging Lionfish", s82bns45AbandonOracle: "Abandon Reason",
		s82bns45GiftOracle: "Alchemist's Gift", s82bns45BigPlayOracle: "Big Play",
		s82bns45BlinkOracle: "Blink of an Eye", s82bns45BruteOracle: "Brute Strength",
		s82bns45BuryOracle: "Bury in Books", s82bns45ChillingOracle: "Chilling Trap",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestS82Bns45GloryBearersPumpsOtherAttackersOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bearers := b12Push(g, me.ID, "Glory Bearers", "Enchantment Creature — Human Cleric", s82bns45GloryOracle, 3, 4)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, bearers, bear)
	passPriorityAroundTable(t, g)
	if got := effectiveToughness(t, g, bear); got != 3 {
		t.Errorf("the other attacker's toughness is %d, want 3", got)
	}
	if got := effectiveToughness(t, g, bearers); got != 4 {
		t.Errorf("Glory Bearers itself is %d toughness, want 4 (another creature only)", got)
	}
	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("power changed to %d, want 2", got)
	}
}

func TestS82Bns45NaiadDiscountsOnlyOnOtherTurns(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Naiad of Hidden Coves", "Enchantment Creature — Nymph", s82bns45NaiadOracle, 2, 3)
	advanceToMain(t, g)
	id := s82bns45Hand(me, "Test Spell", "Instant", "{2}{U}", "")
	fillPoolColored(me, "U", 1)
	fillPoolColored(me, "C", 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("on your own turn the spell cost {1} less")
	}
	me.ManaPool = game.ManaPool{}
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepPrecombatMain)
	fillPoolColored(me, "U", 1)
	fillPoolColored(me, "C", 1)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("on another player's turn the spell should cost {1}{U}: %v", err)
	}
}

func TestS82Bns45StingingLionfishTapsOnFirstSpellOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Stinging Lionfish", "Enchantment Creature — Fish", s82bns45LionfishOracle, 2, 1)
	rock := b12Permanent(g, me.ID, "Rock", "Artifact")

	// Own turn: nothing triggers.
	castCatalogSpell(t, g, "Opt", "Instant", "", nil)
	passPriorityAroundTable(t, g)
	if len(g.PendingChoices) != 0 {
		t.Fatalf("a spell on your own turn triggered: %+v", g.PendingChoices)
	}

	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepPrecombatMain)
	b22CastInstantAs(t, g, me, "Opt")
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, rock)
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	answerOptionPick(t, g, me.ID, 0)
	if !b16Tapped(t, g, rock) {
		t.Error("the target was not tapped")
	}
	// The second spell is silent.
	b22CastInstantAs(t, g, me, "Brainstorm")
	passPriorityAroundTable(t, g)
	if latestOptionPickFor(g, me.ID) != nil {
		t.Error("the second spell that turn asked again")
	}
}

func TestS82Bns45AbandonReasonAndItsMadnessCost(t *testing.T) {
	spec, _ := Lookup(s82bns45AbandonOracle)
	if spec.Madness != "{1}{R}" {
		t.Errorf("madness %q, want {1}{R}", spec.Madness)
	}
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	s82bns45Cast(t, g, me, "Abandon Reason", "{2}{R}", s82bns45AbandonOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}, "R", "C", "C")
	for _, id := range []uuid.UUID{a, b} {
		if got := effectivePower(t, g, id); got != 3 {
			t.Errorf("power %d, want 3", got)
		}
		assertKeywords(t, g, id, "first strike")
	}
}

func TestS82Bns45AlchemistsGiftPumpsThenGivesTheChosenKeyword(t *testing.T) {
	for idx, kw := range []string{"deathtouch", "lifelink"} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
		advanceToMain(t, g)
		s82bns45Cast(t, g, me, "Alchemist's Gift", "{B}", s82bns45GiftOracle, s82bns45Ref(bear), "B")
		answerOptionPick(t, g, me.ID, idx)
		if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 3 || tg != 3 {
			t.Errorf("%s: bear is %d/%d, want 3/3", kw, p, tg)
		}
		if !hasKeywordOnBattlefield(t, g, bear, kw) {
			t.Errorf("the bear did not gain %s", kw)
		}
		other := "lifelink"
		if idx == 1 {
			other = "deathtouch"
		}
		if hasKeywordOnBattlefield(t, g, bear, other) {
			t.Errorf("the bear gained %s too", other)
		}
	}
}

func TestS82Bns45BigPlayPumpsGivesReachAndLeavesACounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	s82bns45Cast(t, g, me, "Big Play", "{1}{G}", s82bns45BigPlayOracle, s82bns45Ref(bear), "G", "C")
	if p := effectivePowerWithCounters(t, g, bear); p != 5 {
		t.Errorf("bear power is %d, want 5 (+2/+2 and a counter)", p)
	}
	assertKeywords(t, g, bear, "reach")
	if got := counterCount(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	advanceToNextSeatsTurn(t, g)
	if p := effectivePowerWithCounters(t, g, bear); p != 3 {
		t.Errorf("after the turn the bear is %d power, want 3 (counter stays)", p)
	}
}

func TestS82Bns45BlinkOfAnEyeBouncesAndOnlyKickedDraws(t *testing.T) {
	for _, kicked := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		advanceToMain(t, g)
		id := s82bns45Hand(me, "Blink of an Eye", "Instant", "{1}{U}", s82bns45BlinkOracle)
		fillPoolColored(me, "U", 1)
		fillPoolColored(me, "C", 1)
		params := game.CastSpellParams{Strict: true, Targets: s82bns45Ref(bear)}
		if kicked {
			if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, Targets: s82bns45Ref(bear)}); err == nil {
				t.Fatal("a kicked cast was paid with two mana")
			}
			fillPoolColored(me, "U", 1)
			fillPoolColored(me, "C", 1)
			params.OptionalCosts = []int{0}
		}
		hand := me.Hand.Size()
		if err := g.CastSpell(me.ID, id, params); err != nil {
			t.Fatalf("cast (kicked=%v): %v", kicked, err)
		}
		passPriorityAroundTable(t, g)
		if g.Battlefield.Contains(bear) {
			t.Errorf("kicked=%v: the permanent was not returned", kicked)
		}
		if !opp.Hand.Contains(bear) {
			t.Errorf("kicked=%v: the permanent is not in its owner's hand", kicked)
		}
		want := hand - 1
		if kicked {
			want++
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("kicked=%v: hand %d → %d, want %d", kicked, hand, got, want)
		}
	}
}

func TestS82Bns45BruteStrengthPumpsAndTramples(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	s82bns45Cast(t, g, me, "Brute Strength", "{1}{R}", s82bns45BruteOracle, s82bns45Ref(bear), "R", "C")
	if p, tg := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tg != 3 {
		t.Errorf("bear is %d/%d, want 5/3", p, tg)
	}
	assertKeywords(t, g, bear, "trample")
	advanceToNextSeatsTurn(t, g)
	if p := effectivePower(t, g, bear); p != 2 {
		t.Errorf("the pump outlasted the turn: power %d", p)
	}
}

func TestS82Bns45BuryInBooksCostsLessTargetingAnAttackerAndTucksSecond(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	idle := b12Creature(g, opp.ID, "Idle", "Creature — Bear", 2, 2)
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	_ = mine
	advanceToNextSeatsTurn(t, g)
	advanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, me.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)

	id := s82bns45Hand(me, "Bury in Books", "Instant", "{4}{U}", s82bns45BuryOracle)
	fillPoolColored(me, "U", 1)
	fillPoolColored(me, "C", 2)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: s82bns45Ref(idle)}); err == nil {
		t.Fatal("a non-attacking target got the discount")
	}
	top, _ := opp.Library.Top()
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: s82bns45Ref(attacker)}); err != nil {
		t.Fatalf("an attacking target should cost {2}{U}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Fatal("the attacker stayed on the battlefield")
	}
	n := opp.Library.Size()
	if opp.Library.Cards[n-2].InstanceID != attacker || opp.Library.Cards[n-1].InstanceID != top.InstanceID {
		t.Error("the creature is not second from the top of its owner's library")
	}
}

func TestS82Bns45ChillingTrapShrinksAndDrawsOnlyWithAWizard(t *testing.T) {
	for _, wizard := range []bool{false, true} {
		g := newCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		ogre := b12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 5, 5)
		if wizard {
			b12Creature(g, me.ID, "Mage", "Creature — Human Wizard", 1, 1)
		}
		advanceToMain(t, g)
		hand := me.Hand.Size()
		s82bns45Cast(t, g, me, "Chilling Trap", "{U}", s82bns45ChillingOracle, s82bns45Ref(ogre), "U")
		if got := effectivePower(t, g, ogre); got != 1 {
			t.Errorf("wizard=%v: power %d, want 1", wizard, got)
		}
		want := hand // the cast card entered and left the hand
		if wizard {
			want++
		}
		if got := me.Hand.Size(); got != want {
			t.Errorf("wizard=%v: hand %d → %d, want %d", wizard, hand, got, want)
		}
	}
}

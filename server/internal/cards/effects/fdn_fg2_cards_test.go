package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fdn_fg2_cards_test.go — Foundations, F–G, second batch (#4465): the
// cards from that stretch of the alphabet the first batch left
// un-attempted. Each test plays the real catalog entry through the
// engine and asserts what its text says.

const (
	fdnFG2FrenziedGoblinOracle    = "6ac470a1-c2be-4971-a5ca-10bb189ebe4d"
	fdnFG2GatewaySneakOracle      = "66c227f2-0e74-43e2-ab24-3866d15c5eef"
	fdnFG2GhituLavarunnerOracle   = "8e3f4987-6f31-4bc5-8bcd-6adb0440ee7c"
	fdnFG2GnarlidColonyOracle     = "e5481431-c952-4ad9-94fe-355076ef632b"
	fdnFG2GoblinNegotiationOracle = "79824266-bffa-476c-bdd3-976e9d789a42"
	fdnFG2GoblinSurpriseOracle    = "79cd09e8-d738-47c0-9473-b6281bec6935"
	fdnFG2GrowFromTheAshesOracle  = "645ab3c7-ee1d-4dd0-811f-2dc7f7c7e792"
)

func TestFdnFG2EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		fdnFG2FrenziedGoblinOracle:    "Frenzied Goblin",
		fdnFG2GatewaySneakOracle:      "Gateway Sneak",
		fdnFG2GhituLavarunnerOracle:   "Ghitu Lavarunner",
		fdnFG2GnarlidColonyOracle:     "Gnarlid Colony",
		fdnFG2GoblinNegotiationOracle: "Goblin Negotiation",
		fdnFG2GoblinSurpriseOracle:    "Goblin Surprise",
		fdnFG2GrowFromTheAshesOracle:  "Grow from the Ashes",
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

// --- Frenzied Goblin ------------------------------------------------------------

// fdnFG2FrenziedGoblinAttack attacks with a Frenzied Goblin, picks target
// as the trigger goes on the stack, and returns once it asks for {R}.
func fdnFG2FrenziedGoblinAttack(t *testing.T, g *game.Game, target uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	goblin := b12Push(g, me.ID, "Frenzied Goblin", "Creature — Goblin Berserker", fdnFG2FrenziedGoblinOracle, 1, 1)
	declareAttack(t, g, opp.ID, goblin)
	mscM2PickTargets(t, g, target)
	if !hasPayUnlessFor(g, me.ID) {
		t.Fatal("the Goblin did not ask whether to pay {R}")
	}
}

func TestFdnFG2FrenziedGoblinPayingStopsTheTargetBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	other := b12Creature(g, opp.ID, "Other", "Creature — Bear", 2, 2)
	fdnFG2FrenziedGoblinAttack(t, g, wall)
	fillPoolColored(me, "R", 1)
	answerPayUnless(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if restrictionsOf(t, g, wall)&game.CantBlock == 0 {
		t.Error("the target can block")
	}
	if restrictionsOf(t, g, other)&game.CantBlock != 0 {
		t.Error("an untargeted creature can't block")
	}
	if got := me.ManaPool.Total(); got != 0 {
		t.Errorf("pool %d, want the {R} spent", got)
	}
}

func TestFdnFG2FrenziedGoblinDecliningDoesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	fdnFG2FrenziedGoblinAttack(t, g, wall)
	answerPayUnless(t, g, me.ID, false)
	passPriorityAroundTable(t, g)
	if restrictionsOf(t, g, wall)&game.CantBlock != 0 {
		t.Error("the target can't block although nothing was paid")
	}
}

// --- Gateway Sneak --------------------------------------------------------------

// fdnFG2PlayLand plays a land with the given type line from the active
// player's hand, with a land drop of its own.
func fdnFG2PlayLand(t *testing.T, g *game.Game, name, typeLine string) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := handCardFull(active, name, typeLine, "", "", nil)
	toMainForCost(t, g)
	active.LandDropsPerTurn++
	if err := g.CastSpell(active.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("play %s: %v", name, err)
	}
	passPriorityAroundTable(t, g)
}

func TestFdnFG2GatewaySneakIsUnblockableAfterAGateEnters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	sneak := b12Push(g, me.ID, "Gateway Sneak", "Creature — Vedalken Rogue", fdnFG2GatewaySneakOracle, 1, 3)
	fdnFG2PlayLand(t, g, "Forest", "Basic Land — Forest")
	if restrictionsOf(t, g, sneak)&game.CantBeBlocked != 0 {
		t.Fatal("a land that is not a Gate made the Sneak unblockable")
	}
	fdnFG2PlayLand(t, g, "Gruul Guildgate", "Land — Gate")
	if restrictionsOf(t, g, sneak)&game.CantBeBlocked == 0 {
		t.Error("the Sneak can be blocked after a Gate entered")
	}
}

func TestFdnFG2GatewaySneakDrawsOnCombatDamageToAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	sneak := b12Push(g, me.ID, "Gateway Sneak", "Creature — Vedalken Rogue", fdnFG2GatewaySneakOracle, 1, 3)
	hand := me.Hand.Size()
	dealNoncombatDamage(g, me.ID, sneak, opp.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Fatalf("hand %d after noncombat damage, want %d", got, hand)
	}
	dealCombatDamage(g, me.ID, sneak, opp.ID)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d after combat damage, want %d", got, hand+1)
	}
}

// --- Ghitu Lavarunner -----------------------------------------------------------

// fdnFG2CastBlank casts and resolves an uncatalogued spell with the given
// type line, which leaves it in its owner's graveyard.
func fdnFG2CastBlank(t *testing.T, g *game.Game, p *game.Player, typeLine string) {
	t.Helper()
	toMainForCost(t, g)
	id := handCardFull(p, "Blank "+typeLine, typeLine, "", "", nil)
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", typeLine, err)
	}
	passPriorityAroundTable(t, g)
	if !p.Graveyard.Contains(id) {
		t.Fatalf("the %s is not in the graveyard", typeLine)
	}
}

func TestFdnFG2GhituLavarunnerGrowsAndHastesWithTwoSpellsInYourGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ghitu := b12Push(g, me.ID, "Ghitu Lavarunner", "Creature — Human Wizard", fdnFG2GhituLavarunnerOracle, 1, 2)
	fdnFG2CastBlank(t, g, me, "Instant")
	killCreature(t, g, me.ID, b12Creature(g, me.ID, "Dead Bear", "Creature — Bear", 2, 2))
	if p, tough := effectivePTOf(t, g, ghitu); p != 1 || tough != 2 {
		t.Fatalf("Ghitu is %d/%d with one instant card in the graveyard, want 1/2", p, tough)
	}
	if hasEffectiveKeyword(t, g, ghitu, "haste") {
		t.Fatal("Ghitu has haste with one instant card in the graveyard")
	}
	fdnFG2CastBlank(t, g, me, "Sorcery")
	if p, tough := effectivePTOf(t, g, ghitu); p != 2 || tough != 2 {
		t.Errorf("Ghitu is %d/%d with an instant and a sorcery in the graveyard, want 2/2", p, tough)
	}
	if !hasEffectiveKeyword(t, g, ghitu, "haste") {
		t.Error("Ghitu has no haste with an instant and a sorcery in the graveyard")
	}
}

// --- Gnarlid Colony -------------------------------------------------------------

func TestFdnFG2GnarlidColonyEntersWithTwoCountersOnlyWhenKicked(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		want     int
	}{
		{"unkicked", nil, 0},
		{"kicked", []int{0}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			id, err := castWithOptionalCosts(t, g, "Gnarlid Colony", "Creature — Beast", fdnFG2GnarlidColonyOracle, nil, tc.optional, nil)
			if err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			if n := plusOneCounters(g, id); n != tc.want {
				t.Errorf("+1/+1 counters %d, want %d", n, tc.want)
			}
			if got := hasEffectiveKeyword(t, g, id, "trample"); got != (tc.want > 0) {
				t.Errorf("trample %v with %d counters", got, tc.want)
			}
		})
	}
}

func TestFdnFG2GnarlidColonyGivesTrampleToYourCreaturesWithACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Gnarlid Colony", "Creature — Beast", fdnFG2GnarlidColonyOracle, 2, 2)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	putCounters(t, g, theirs, game.CounterPlusOne, 1)
	if hasEffectiveKeyword(t, g, bear, "trample") {
		t.Fatal("a creature with no counter has trample")
	}
	putCounters(t, g, bear, game.CounterPlusOne, 1)
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("your creature with a +1/+1 counter has no trample")
	}
	if hasEffectiveKeyword(t, g, theirs, "trample") {
		t.Error("an opponent's creature with a counter has trample")
	}
}

// --- Goblin Negotiation ---------------------------------------------------------

func TestFdnFG2GoblinNegotiationMakesAGoblinPerExcessDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castXSpell(t, g, "Goblin Negotiation", "Sorcery", fdnFG2GoblinNegotiationOracle, "{X}{R}{R}", 5, cardRefs(bear))
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Fatal("the 2/2 survived 5 damage")
	}
	if got := b27CountNamed(g, me.ID, "Goblin"); got != 3 {
		t.Errorf("5 damage to a 2/2 made %d Goblins, want 3", got)
	}
	ox := b12Creature(g, opp.ID, "Ox", "Creature — Ox", 4, 4)
	castXSpell(t, g, "Goblin Negotiation", "Sorcery", fdnFG2GoblinNegotiationOracle, "{X}{R}{R}", 2, cardRefs(ox))
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ox) {
		t.Fatal("the 4/4 died to 2 damage")
	}
	if got := b27CountNamed(g, me.ID, "Goblin"); got != 3 {
		t.Errorf("no excess damage made Goblins: %d, want 3", got)
	}
}

// --- Goblin Surprise ------------------------------------------------------------

func TestFdnFG2GoblinSurprisePumpsYourTeamOrMakesTwoGoblins(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if err := rfCastModal(t, g, "Goblin Surprise", fdnFG2GoblinSurpriseOracle, []int{0}, nil); err != nil {
		t.Fatalf("mode 0: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 4 || tough != 2 {
		t.Errorf("your Bear is %d/%d, want 4/2", p, tough)
	}
	if p, _ := effectivePTOf(t, g, theirs); p != 2 {
		t.Errorf("an opponent's creature has power %d, want 2", p)
	}
	if got := b27CountNamed(g, me.ID, "Goblin"); got != 0 {
		t.Errorf("mode 0 made %d Goblins", got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	if err := rfCastModal(t, g2, "Goblin Surprise", fdnFG2GoblinSurpriseOracle, []int{1}, nil); err != nil {
		t.Fatalf("mode 1: %v", err)
	}
	passPriorityAroundTable(t, g2)
	if got := b27CountNamed(g2, me2.ID, "Goblin"); got != 2 {
		t.Errorf("mode 1 made %d Goblins, want 2", got)
	}
}

// --- Grow from the Ashes --------------------------------------------------------

func TestFdnFG2GrowFromTheAshesPutsOneBasicOrTwoWhenKicked(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		want     int
	}{
		{"unkicked", nil, 1},
		{"kicked", []int{0}, 2},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			ids := seedSearchLibrary(me,
				game.Card{Name: "Forest A", TypeLine: "Basic Land — Forest"},
				game.Card{Name: "Forest B", TypeLine: "Basic Land — Forest"},
				game.Card{Name: "Forest C", TypeLine: "Basic Land — Forest"},
				game.Card{Name: "Nonbasic", TypeLine: "Land — Forest Island"},
			)
			if _, err := castWithOptionalCosts(t, g, "Grow from the Ashes", "Sorcery", fdnFG2GrowFromTheAshesOracle, nil, tc.optional, nil); err != nil {
				t.Fatalf("CastSpell: %v", err)
			}
			passPriorityAroundTable(t, g)
			c := searchChoiceFor(g, me.ID)
			if c == nil {
				t.Fatal("no search prompt")
			}
			beginSearchIfAsked(t, g, c)
			if err := g.ResolveSearchLibrary(c.ID, me.ID, []uuid.UUID{ids[3]}); err == nil {
				t.Fatal("the search accepted a nonbasic land")
			}
			if err := g.ResolveSearchLibrary(c.ID, me.ID, ids[:tc.want+1]); err == nil {
				t.Fatalf("the search accepted %d cards, want at most %d", tc.want+1, tc.want)
			}
			answerSearchByID(t, g, me.ID, ids[:tc.want]...)
			passPriorityAroundTable(t, g)
			for _, id := range ids[:tc.want] {
				if findBattlefieldCardForTest(g, id) == nil {
					t.Errorf("%s is not on the battlefield", id)
					continue
				}
				if battlefieldTapped(g, id) {
					t.Errorf("%s entered tapped", id)
				}
			}
			if findBattlefieldCardForTest(g, ids[2]) != nil {
				t.Error("an unpicked land entered")
			}
		})
	}
}

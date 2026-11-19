package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82 BNS batch 48 — Foundations Jumpstart (#5929).

const (
	s82B48PhantasmalForm = "d21facb6-27c2-49cb-9f13-7d6079096eda"
	s82B48PredatorStrike = "f62e9a9f-9989-4a62-b063-cca32e21fe8b"
	s82B48RallyOfWings   = "984336a7-85bd-4549-9771-03a0c2367db0"
	s82B48Remand         = "d71cd08e-3e84-41ff-b9db-9e343c0af6b4"
	s82B48RepelDarkness  = "47f5ea79-2bf2-4e17-9c6d-e1cbcd9d30d1"
	s82B48RunAshore      = "9b3bcd57-bf3f-4d72-810a-fdb5c8ea1551"
	s82B48SarkhansRage   = "3313aa8f-6441-4daa-b8fa-e6144f60bea8"
	s82B48SoulRead       = "7cf93170-29ca-4f44-a8ce-325d034d4ce8"
	s82B48SoulSear       = "1ef73b3b-ab4a-494b-b969-dcde995cee34"
	s82B48SporeSwarm     = "8fb599a7-2c61-446b-a7a7-cd5367f9b796"
	s82B48StonyStrength  = "da5a7cb8-1fa7-4b91-abd4-9e766291be63"
)

func TestS82BnsB48CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		s82B48PhantasmalForm: "Phantasmal Form", s82B48PredatorStrike: "Predator's Strike",
		s82B48RallyOfWings: "Rally of Wings", s82B48Remand: "Remand",
		s82B48RepelDarkness: "Repel the Darkness", s82B48RunAshore: "Run Ashore",
		s82B48SarkhansRage: "Sarkhan's Rage", s82B48SoulRead: "Soul Read",
		s82B48SoulSear: "Soul Sear", s82B48SporeSwarm: "Spore Swarm",
		s82B48StonyStrength: "Stony Strength",
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

func s82B48Tap(t *testing.T, g *game.Game, ids ...uuid.UUID) {
	t.Helper()
	for _, id := range ids {
		g.WithWriteLock(func() {
			if err := g.TapTargetForEffectBy(id, uuid.Nil); err != nil {
				t.Fatalf("tap: %v", err)
			}
		})
	}
}

func s82B48Tapped(g *game.Game, id uuid.UUID) bool {
	var tapped bool
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(id); ok {
			tapped = c.Tapped
		}
	})
	return tapped
}

// --- Spore Swarm --------------------------------------------------------

func TestS82BnsB48SporeSwarmMakesThreeSaprolings(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := p1g2Seats(g)
	castCatalogSpell(t, g, "Spore Swarm", "Instant", s82B48SporeSwarm, nil)
	passPriorityAroundTable(t, g)
	n := 0
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Saproling" && c.Controller == me.ID && c.IsToken() {
				if p, tough := effectivePTOf(t, g, c.InstanceID); p != 1 || tough != 1 {
					t.Errorf("a Saproling is %d/%d, want 1/1", p, tough)
				}
				n++
			}
		}
	})
	if n != 3 {
		t.Errorf("%d Saprolings, want 3", n)
	}
}

// --- Predator's Strike --------------------------------------------------

func TestS82BnsB48PredatorsStrikePumpsAndGrantsTrample(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := p1g2Seats(g)
	bear := p1Creature(g, me.ID, 2, 2)
	castCatalogSpell(t, g, "Predator's Strike", "Instant", s82B48PredatorStrike, pr6Card(bear))
	passPriorityAroundTable(t, g)
	if p, tough := effectivePTOf(t, g, bear); p != 5 || tough != 5 {
		t.Errorf("the bear is %d/%d, want 5/5", p, tough)
	}
	if !p1g2HasKeyword(g, bear, "trample") {
		t.Error("the bear has no trample")
	}
}

func TestS82BnsB48PredatorsStrikeRefusesANonCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := p1g2Seats(g)
	rock := pushCatalogPermanent(g, me.ID, "Rock", "Artifact", "", false)
	toMainForCost(t, g)
	if err := castCatalogSpellErrStrict(t, g, "Predator's Strike", "Instant", s82B48PredatorStrike, pr6Card(rock)); err == nil {
		t.Error("an artifact is not a legal target")
	}
}

// --- Stony Strength -----------------------------------------------------

func TestS82BnsB48StonyStrengthCountersAndUntaps(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := p1g2Seats(g)
	bear := p1Creature(g, me.ID, 2, 2)
	s82B48Tap(t, g, bear)
	castCatalogSpell(t, g, "Stony Strength", "Instant", s82B48StonyStrength, pr6Card(bear))
	passPriorityAroundTable(t, g)
	if n := findBattlefieldCardForTest(g, bear).Counters[game.CounterPlusOne]; n != 1 {
		t.Errorf("the bear has %d +1/+1 counters, want 1", n)
	}
	if s82B48Tapped(g, bear) {
		t.Error("the bear stayed tapped")
	}
}

func TestS82BnsB48StonyStrengthRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := p1g2Seats(g)
	theirs := p1Creature(g, opp.ID, 2, 2)
	toMainForCost(t, g)
	if err := castCatalogSpellErrStrict(t, g, "Stony Strength", "Instant", s82B48StonyStrength, pr6Card(theirs)); err == nil {
		t.Error("an opponent's creature is not a legal target")
	}
}

// --- Rally of Wings -----------------------------------------------------

func TestS82BnsB48RallyOfWingsUntapsAllAndPumpsFliersOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	flier := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: me.ID, Controller: me.ID})
	bear := p1Creature(g, me.ID, 2, 2)
	theirs := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Their Bird", TypeLine: "Creature — Bird",
		Power: 1, Toughness: 1, Keywords: []string{"flying"}, Owner: opp.ID, Controller: opp.ID})
	s82B48Tap(t, g, flier, bear, theirs)
	castCatalogSpell(t, g, "Rally of Wings", "Instant", s82B48RallyOfWings, nil)
	passPriorityAroundTable(t, g)
	if s82B48Tapped(g, flier) || s82B48Tapped(g, bear) {
		t.Error("a creature of mine stayed tapped")
	}
	if !s82B48Tapped(g, theirs) {
		t.Error("an opponent's creature was untapped")
	}
	if p, tough := effectivePTOf(t, g, flier); p != 3 || tough != 3 {
		t.Errorf("the flier is %d/%d, want 3/3", p, tough)
	}
	if p, tough := effectivePTOf(t, g, bear); p != 2 || tough != 2 {
		t.Errorf("the ground creature is %d/%d, want 2/2", p, tough)
	}
	if p, tough := effectivePTOf(t, g, theirs); p != 1 || tough != 1 {
		t.Errorf("their flier is %d/%d, want 1/1", p, tough)
	}
}

// --- Repel the Darkness -------------------------------------------------

func TestS82BnsB48RepelTheDarknessTapsTwoAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	a, b := p1Creature(g, opp.ID, 2, 2), p1Creature(g, opp.ID, 2, 2)
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Repel the Darkness", "Instant", s82B48RepelDarkness, cardRefs(a, b))
	passPriorityAroundTable(t, g)
	if !s82B48Tapped(g, a) || !s82B48Tapped(g, b) {
		t.Error("both creatures should be tapped")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (one card drawn)", got, hand+1)
	}
}

func TestS82BnsB48RepelTheDarknessDrawsWithNoTargetAndRefusesThree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	hand := me.Hand.Size()
	castCatalogSpell(t, g, "Repel the Darkness", "Instant", s82B48RepelDarkness, nil)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d (one card drawn)", got, hand+1)
	}
	three := cardRefs(p1Creature(g, opp.ID, 1, 1), p1Creature(g, opp.ID, 1, 1), p1Creature(g, opp.ID, 1, 1))
	if err := castCatalogSpellErrStrict(t, g, "Repel the Darkness", "Instant", s82B48RepelDarkness, three); err == nil {
		t.Error("three targets are one too many")
	}
}

// --- Phantasmal Form ----------------------------------------------------

func TestS82BnsB48PhantasmalFormMakesBlueIllusionFliers(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	big := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big Red", TypeLine: "Creature — Giant",
		Power: 6, Toughness: 6, Colors: []string{"R"}, Owner: opp.ID, Controller: opp.ID})
	small := p1Creature(g, me.ID, 1, 1)
	castCatalogSpell(t, g, "Phantasmal Form", "Instant", s82B48PhantasmalForm, cardRefs(big, small))
	passPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{big, small} {
		if p, tough := effectivePTOf(t, g, id); p != 3 || tough != 3 {
			t.Errorf("a target is %d/%d, want 3/3", p, tough)
		}
		if !p1g2HasKeyword(g, id, "flying") {
			t.Error("a target has no flying")
		}
	}
	g.ReadSnapshot(func() {
		c, _ := g.LookupCardForEffect(big)
		if !c.HasSubtype("Illusion") || !c.HasSubtype("Giant") {
			t.Errorf("%q: want Giant kept and Illusion added", c.TypeLine)
		}
		colors := map[string]bool{}
		for _, col := range effectiveColors(t, g, big) {
			colors[col] = true
		}
		if !colors["U"] || !colors["R"] {
			t.Errorf("colors %v: want red kept and blue added", effectiveColors(t, g, big))
		}
	})
}

func TestS82BnsB48PhantasmalFormEndsAtCleanup(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := p1g2Seats(g)
	pushLibraryCardForTest(me, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	bear := p1Creature(g, me.ID, 2, 2)
	castCatalogSpell(t, g, "Phantasmal Form", "Instant", s82B48PhantasmalForm, cardRefs(bear))
	passPriorityAroundTable(t, g)
	for i := 0; i < 20 && g.Turn.Step != game.StepCleanup; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep past cleanup: %v", err)
	}
	if p1g2HasKeyword(g, bear, "flying") {
		t.Error("the bear kept flying after the turn ended")
	}
}

// --- Remand -------------------------------------------------------------

func s82B48RemandBoard(t *testing.T, targetOracle string) (g *game.Game, victim uuid.UUID, caster, them *game.Player) {
	t.Helper()
	g = newCatalogGame(t)
	caster, them = g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	victim = castCatalogSpell(t, g, "Their Sorcery", "Sorcery", targetOracle, nil)
	pushLibraryCardForTest(them, game.Card{Name: "Filler", TypeLine: "Creature — Bear"})
	remand := handCardFull(them, "Remand", "Instant", "", s82B48Remand, []string{"U"})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	if err := g.CastSpell(them.ID, remand, game.CastSpellParams{Targets: cardRefs(victim)}); err != nil {
		t.Fatalf("casting Remand: %v", err)
	}
	return g, victim, caster, them
}

func TestS82BnsB48RemandReturnsTheSpellToItsOwnersHandAndDraws(t *testing.T) {
	g, victim, caster, them := s82B48RemandBoard(t, "")
	hand := them.Hand.Size()
	passPriorityAroundTable(t, g)
	if g.Stack.Contains(victim) {
		t.Fatal("the spell is still on the stack")
	}
	if !caster.Hand.Contains(victim) {
		t.Error("the countered spell is not in its owner's hand")
	}
	if caster.Graveyard.Contains(victim) {
		t.Error("the countered spell went to the graveyard")
	}
	if got := them.Hand.Size(); got != hand+1 {
		t.Errorf("Remand's controller hand %d, want %d", got, hand+1)
	}
}

// --- Sarkhan's Rage -----------------------------------------------------

func TestS82BnsB48SarkhansRageHurtsYouWithoutADragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	big := p1Creature(g, opp.ID, 6, 6)
	life := me.Life
	castCatalogSpell(t, g, "Sarkhan's Rage", "Instant", s82B48SarkhansRage, pr6Card(big))
	passPriorityAroundTable(t, g)
	p1g2WantDamage(t, g, big, 5, "the 6/6")
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
}

func TestS82BnsB48SarkhansRageSparesYouWithADragon(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Whelp", TypeLine: "Creature — Dragon",
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	life, theirs := me.Life, opp.Life
	castCatalogSpell(t, g, "Sarkhan's Rage", "Instant", s82B48SarkhansRage, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)
	if opp.Life != theirs-5 {
		t.Errorf("their life %d, want %d", opp.Life, theirs-5)
	}
	if me.Life != life {
		t.Errorf("my life %d, want %d", me.Life, life)
	}
}

// --- Soul Sear ----------------------------------------------------------

func TestS82BnsB48SoulSearKillsAnIndestructibleCreature(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := p1g2Seats(g)
	big := p1g2Indestructible(g, opp.ID, 5, 5)
	castCatalogSpell(t, g, "Soul Sear", "Instant", s82B48SoulSear, pr6Card(big))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, big, game.ZoneGraveyard, "the indestructible 5/5")
}

func TestS82BnsB48SoulSearHitsAPlaneswalkerAndRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	walker := p1g2Planeswalker(g, opp.ID, 5)
	castCatalogSpell(t, g, "Soul Sear", "Instant", s82B48SoulSear, pr6Card(walker))
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, walker, game.ZoneGraveyard, "the 5-loyalty planeswalker")
	land := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	if err := castCatalogSpellErrStrict(t, g, "Soul Sear", "Instant", s82B48SoulSear, pr6Card(land)); err == nil {
		t.Error("a land is not a legal target")
	}
}

// --- Soul Read ----------------------------------------------------------

func TestS82BnsB48SoulReadDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := p1g2Seats(g)
	pushLibraryCardForTest(me, game.Card{Name: "A", TypeLine: "Creature — Bear"})
	pushLibraryCardForTest(me, game.Card{Name: "B", TypeLine: "Creature — Bear"})
	hand := me.Hand.Size()
	castCatalogSpellWithModes(t, g, "Soul Read", "Instant", s82B48SoulRead, []int{1})
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d (the spell left, two drawn)", got, hand+2)
	}
}

func s82B48SoulReadBoard(t *testing.T) (g *game.Game, victim uuid.UUID, caster *game.Player) {
	t.Helper()
	g = newCatalogGame(t)
	caster, them := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	victim = castCatalogSpell(t, g, "Their Sorcery", "Sorcery", "", nil)
	read := handCardFull(them, "Soul Read", "Instant", "", s82B48SoulRead, []string{"U"})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority to the responder: %v", err)
	}
	if err := g.CastSpell(them.ID, read, game.CastSpellParams{Modes: []int{0}, Targets: cardRefs(victim)}); err != nil {
		t.Fatalf("casting Soul Read: %v", err)
	}
	passUntilTaxed(t, g, caster.ID)
	return g, victim, caster
}

func TestS82BnsB48SoulReadCountersUnlessFourIsPaid(t *testing.T) {
	g, victim, caster := s82B48SoulReadBoard(t)
	if prompt := answerPayUnless(t, g, caster.ID, false); prompt.PayCost != "{4}" {
		t.Errorf("the tax is %q, want {4}", prompt.PayCost)
	}
	if g.Stack.Contains(victim) {
		t.Error("a declined tax must counter the spell")
	}

	g, victim, caster = s82B48SoulReadBoard(t)
	for i := 0; i < 4; i++ {
		caster.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	answerPayUnless(t, g, caster.ID, true)
	if !g.Stack.Contains(victim) {
		t.Error("a paid tax countered the spell anyway")
	}
}

// --- Run Ashore ---------------------------------------------------------

func s82B48CastRunAshore(t *testing.T, g *game.Game, modes []int, targets []game.TargetRef) {
	t.Helper()
	me, _ := p1g2Seats(g)
	toMainForCost(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Run Ashore", TypeLine: "Instant", OracleID: s82B48RunAshore,
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets}); err != nil {
		t.Fatalf("cast Run Ashore: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestS82BnsB48RunAshoreOwnerChoosesTopOrBottomAndOtherBulletBounces(t *testing.T) {
	g := newCatalogGame(t)
	_, opp := p1g2Seats(g)
	tucked := p1Creature(g, opp.ID, 2, 2)
	bounced := p1Creature(g, opp.ID, 3, 3)
	s82B48CastRunAshore(t, g, []int{0, 1}, []game.TargetRef{
		{Kind: game.TargetCard, ID: tucked, Mode: 0, Slot: 0},
		{Kind: game.TargetCard, ID: bounced, Mode: 1, Slot: 0},
	})
	ask := putInLibraryChoiceFor(g, opp.ID)
	if ask == nil {
		t.Fatal("the owner was not asked for the top or bottom")
	}
	if err := g.ResolvePutInLibrary(ask.ID, opp.ID, []uuid.UUID{tucked}, nil); err != nil {
		t.Fatalf("ResolvePutInLibrary: %v", err)
	}
	if g.Battlefield.Contains(tucked) || !opp.Library.Contains(tucked) {
		t.Error("the first target is not in its owner's library")
	}
	if got := opp.Library.Cards[0].InstanceID; got != tucked {
		t.Error("the owner chose the bottom")
	}
	if g.Battlefield.Contains(bounced) || !opp.Hand.Contains(bounced) {
		t.Error("the second target is not in its owner's hand")
	}
}

func TestS82BnsB48RunAshoreModeTwoAloneBouncesAndRefusesALand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := p1g2Seats(g)
	bear := p1Creature(g, opp.ID, 2, 2)
	s82B48CastRunAshore(t, g, []int{1}, []game.TargetRef{{Kind: game.TargetCard, ID: bear, Mode: 0, Slot: 0}})
	if !opp.Hand.Contains(bear) {
		t.Error("the creature did not return to its owner's hand")
	}
	land := pushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Run Ashore", TypeLine: "Instant", OracleID: s82B48RunAshore,
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Modes: []int{1}, Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land, Mode: 0, Slot: 0}}}); err == nil {
		t.Error("a land is not a legal target")
	}
}

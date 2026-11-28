package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b804_cards_test.go — Mirage's counter cards, played from the real
// catalog entries (S82 OLD batch 804).

const (
	b804SporesOracle   = "0e3b7ec1-3760-4ac4-81ef-1ab389392f1b"
	b804VampireOracle  = "2c029419-3601-4158-b10f-45a19de7faa3"
	b804ShaukuOracle   = "49478593-29ff-4e5a-b1e8-6462c2d6176d"
	b804CycleOracle    = "5941374b-eeb8-4d0e-9367-f7f0f8b79739"
	b804InfluenceOrcl  = "8930b1ce-80ac-48b6-a368-ccff9fd01d38"
	b804BottleOracle   = "8f167e2f-7b89-4ae2-8332-7c978f71522e"
	b804HarbingerOracl = "9827d01d-c822-4f1b-9042-a2092c72c914"
	b804Bear           = "Creature — Bear"
)

func TestS82OldB804CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b804SporesOracle:   "Mindbender Spores",
		b804VampireOracle:  "Ravenous Vampire",
		b804ShaukuOracle:   "Shauku, Endbringer",
		b804CycleOracle:    "Cycle of Life",
		b804InfluenceOrcl:  "Jabari's Influence",
		b804BottleOracle:   "Ventifact Bottle",
		b804HarbingerOracl: "Harbinger of Night",
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

// b804NextUpkeepOfSeat0 passes through the other seat's turn to seat 0's upkeep
// and lets the upkeep triggers go on the stack.
func b804NextUpkeepOfSeat0(t *testing.T, g *game.Game) {
	t.Helper()
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
}

func TestHarbingerOfNightPutsAMinusOneCounterOnEachCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	harbinger := effectstest.B12Push(g, me.ID, "Harbinger of Night", "Creature — Spirit", b804HarbingerOracl, 2, 3)
	mine := effectstest.B12Creature(g, me.ID, "Mine", b804Bear, 3, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", b804Bear, 3, 3)
	rock := effectstest.B12Permanent(g, opp.ID, "Rock", "Artifact")
	b804NextUpkeepOfSeat0(t, g)
	for _, id := range []uuid.UUID{harbinger, mine, theirs} {
		if got := effectstest.CountersOn(g, id, game.CounterMinusOne); got != 1 {
			t.Errorf("a creature has %d -1/-1 counters, want 1", got)
		}
	}
	if got := effectstest.CountersOn(g, rock, game.CounterMinusOne); got != 0 {
		t.Errorf("a noncreature has %d -1/-1 counters", got)
	}
}

// b804Sacrifice answers the open sacrifice offer with `card`.
func b804Sacrifice(t *testing.T, g *game.Game, chooser, card uuid.UUID) error {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoicePayUnless && c.Chooser == chooser {
			return g.ResolvePayUnlessWithCards(c.ID, chooser, true, []uuid.UUID{card})
		}
	}
	t.Fatal("no sacrifice offer was open")
	return nil
}

func TestRavenousVampireSacrificesToGrowOrTaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	vampire := effectstest.B12Push(g, me.ID, "Ravenous Vampire", "Creature — Vampire", b804VampireOracle, 3, 3)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", b804Bear, 1, 1)
	golem := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Golem", TypeLine: "Artifact Creature — Golem", Power: 1, Toughness: 1})

	b804NextUpkeepOfSeat0(t, g)
	if err := b804Sacrifice(t, g, me.ID, golem); err == nil {
		t.Fatal("sacrificed an artifact creature to a nonartifact cost: accepted")
	}
	if err := b804Sacrifice(t, g, me.ID, fodder); err != nil {
		t.Fatalf("sacrifice: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
	if got := effectstest.PlusOnes(g, vampire); got != 1 {
		t.Errorf("%d +1/+1 counters after sacrificing, want 1", got)
	}
	if effectstest.TappedOnBattlefield(t, g, vampire) {
		t.Error("the Vampire tapped although a creature was sacrificed")
	}

	// Next upkeep: decline, the Vampire taps and gets no counter.
	b804NextUpkeepOfSeat0(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.TappedOnBattlefield(t, g, vampire) {
		t.Error("the Vampire did not tap after declining")
	}
	if got := effectstest.PlusOnes(g, vampire); got != 1 {
		t.Errorf("%d counters after declining, want 1", got)
	}
}

func TestShaukuCantAttackWithAnotherCreatureAndLosesLifeEachUpkeep(t *testing.T) {
	// Another creature on the battlefield, anyone's, stops the attack.
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shauku := effectstest.B12Push(g, me.ID, "Shauku, Endbringer", "Legendary Creature — Vampire", b804ShaukuOracle, 5, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b804Bear, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(shauku, opp.ID); err == nil {
		t.Fatal("attacked with another creature on the battlefield: accepted")
	}
	_ = bear

	// Alone, it attacks.
	g = effectstest.NewCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	shauku = effectstest.B12Push(g, me.ID, "Shauku, Endbringer", "Legendary Creature — Vampire", b804ShaukuOracle, 5, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(shauku, opp.ID); err != nil {
		t.Fatalf("a lone Shauku could not attack: %v", err)
	}

	// The upkeep costs 3 life.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.B12Push(g, me.ID, "Shauku, Endbringer", "Legendary Creature — Vampire", b804ShaukuOracle, 5, 5)
	life := me.Life
	b804NextUpkeepOfSeat0(t, g)
	if me.Life != life-3 {
		t.Errorf("life %d -> %d, want -3", life, me.Life)
	}
}

func TestShaukuExilesACreatureAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shauku := effectstest.B12Push(g, me.ID, "Shauku, Endbringer", "Legendary Creature — Vampire", b804ShaukuOracle, 5, 5)
	bear := effectstest.B12Creature(g, opp.ID, "Bear", b804Bear, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, shauku, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the creature was not exiled")
	}
	if got := effectstest.PlusOnes(g, shauku); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
}

func TestVentifactBottleChargesThenReleasesColorlessMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bottle := effectstest.PushCatalogPermanent(g, me.ID, "Ventifact Bottle", "Artifact", b804BottleOracle, false)
	effectstest.AdvanceToMain(t, g)
	if n := g.Stack.Size(); n != 0 {
		t.Fatalf("%d items on the stack on reaching the main phase with no charge counters: the intervening if did not hold the trigger back", n)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Turn.Step != game.StepPrecombatMain {
		t.Fatalf("passing priority left the precombat main phase: %s", g.Turn.Step)
	}
	for i := 0; i < 3; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "C"})
	}
	if err := g.ActivateCatalogAbility(me.ID, bottle, 0, game.ActivateAbilityParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("activate: %v (step %s, active seat %d)", err, g.Turn.Step, g.Turn.ActiveSeat)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bottle, game.CounterCharge); got != 2 {
		t.Fatalf("%d charge counters, want 2", got)
	}
	// Next own first main phase: the Bottle is tapped, empties and adds {C}{C}.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	effectstest.AdvanceToMain(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bottle, game.CounterCharge); got != 0 {
		t.Errorf("%d charge counters after the release, want 0", got)
	}
	if !effectstest.TappedOnBattlefield(t, g, bottle) {
		t.Error("the Bottle was not tapped")
	}
	if me.ManaPool.Total() != 2 {
		t.Errorf("%d mana in the pool, want 2", me.ManaPool.Total())
	}
}

func TestJabarisInfluenceStealsAnAttackerOnlyAfterCombat(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// The opponent's turn: their creatures attack me.
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	attacker := effectstest.B12Creature(g, opp.ID, "Raider", b804Bear, 2, 2)
	black := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Ghoul", TypeLine: b804Bear, Colors: []string{"B"}, Power: 2, Toughness: 2})
	idle := effectstest.B12Creature(g, opp.ID, "Idle", b804Bear, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{attacker, black} {
		if err := g.DeclareAttacker(id, me.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	cast := func(target uuid.UUID) error {
		id := uuid.New()
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Jabari's Influence", TypeLine: "Instant", OracleID: b804InfluenceOrcl, Owner: me.ID, Controller: me.ID})
		return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(target)})
	}
	if err := cast(attacker); err == nil {
		t.Fatal("cast during combat: accepted")
	}
	effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
	if err := cast(idle); err == nil {
		t.Fatal("targeted a creature that did not attack: accepted")
	}
	if err := cast(black); err == nil {
		t.Fatal("targeted a black creature: accepted")
	}
	if err := cast(attacker); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(attacker)
	if !ok || c.Controller != me.ID {
		t.Error("the attacker is not under the caster's control")
	}
	if got := effectstest.CountersOn(g, attacker, "-1/-0"); got != 1 {
		t.Errorf("%d -1/-0 counters, want 1", got)
	}
	if p, _ := effectstest.PtOf(t, g, attacker); p != 1 {
		t.Errorf("power %d, want 1", p)
	}
}

func TestCycleOfLifeShrinksACreatureYouCastThenGrowsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	cycle := effectstest.PushCatalogPermanent(g, me.ID, "Cycle of Life", "Enchantment", b804CycleOracle, false)
	old := effectstest.B12Creature(g, me.ID, "Old", b804Bear, 3, 3)
	bear := effectstest.B12PlayFromHand(t, g, "Bear Cub", b804Bear, "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if err := g.ActivateCatalogAbility(me.ID, cycle, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(old)}); err == nil {
		t.Fatal("targeted a creature that was not cast this turn: accepted")
	}
	if err := g.ActivateCatalogAbility(me.ID, cycle, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(cycle) {
		t.Error("the enchantment was not returned to hand as a cost")
	}
	if p, tough := effectstest.PtOf(t, g, bear); p != 0 || tough != 1 {
		t.Errorf("the creature is %d/%d, want 0/1", p, tough)
	}
	b804NextUpkeepOfSeat0(t, g)
	if got := effectstest.PlusOnes(g, bear); got != 1 {
		t.Errorf("%d +1/+1 counters at the next upkeep, want 1", got)
	}
}

func TestMindbenderSporesFreezeWhatTheyBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spores := effectstest.B12Push(g, opp.ID, "Mindbender Spores", "Creature — Fungus Wall", b804SporesOracle, 0, 1)
	attacker := effectstest.B12Creature(g, me.ID, "Ogre", b804Bear, 3, 3)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(spores, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, attacker, effects.OldB804FungusCounter); got != 4 {
		t.Fatalf("%d fungus counters on the blocked creature, want 4", got)
	}
	// It stays tapped through its controller's untap step while it has a
	// fungus counter, and loses one each upkeep.
	effectstest.TapCard(g, attacker)
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, 0)
	if !effectstest.TappedOnBattlefield(t, g, attacker) {
		t.Fatal("the creature untapped with a fungus counter on it")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, attacker, effects.OldB804FungusCounter); got != 3 {
		t.Errorf("%d fungus counters after one upkeep, want 3", got)
	}
}

package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_snc_b14_cards_test.go — Streets of New Capenna batch 14 (#5491):
// removal and conditional-cost spells. Each test plays the real catalog
// entry through the engine and asserts what the printed text says.

const (
	sncB14UnleashOracle   = "d305d2dd-578c-4964-a7c2-26490bb2ef2e"
	sncB14CorpseOracle    = "ab715312-0103-4681-b559-520026abd350"
	sncB14TakeoverOracle  = "5c7f91af-8bac-4da1-b9e5-f15989c9d53c"
	sncB14AssaultOracle   = "23cf22fe-483f-4841-882e-e1513bb1d495"
	sncB14OutOfWayOracle  = "9300874a-32a9-4869-b3e1-1b996403df4e"
	sncB14KnockoutOracle  = "7068637c-b025-41a5-be31-df1850f33b68"
	sncB14WhackOracle     = "19becb14-3774-4732-b792-d2d0e5a94b35"
	sncB14TorchOracle     = "b6d61d96-a56a-4f4c-a196-05e03fe16ed2"
	sncB14EvenScoreOracle = "25f5cca6-a414-46d9-82e3-ce2fa21b14ca"
	sncB14BodyCountOracle = "9041338d-b660-4fc2-922f-cdf9d1a25993"
	sncB14MetamorOracle   = "5779d9b2-4de1-4aa4-a56d-0da05146743a"
)

func TestSncB14EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		sncB14UnleashOracle:   "Unleash the Inferno",
		sncB14CorpseOracle:    "Corpse Explosion",
		sncB14TakeoverOracle:  "Hostile Takeover",
		sncB14AssaultOracle:   "Structural Assault",
		sncB14OutOfWayOracle:  "Out of the Way",
		sncB14KnockoutOracle:  "Knockout Blow",
		sncB14WhackOracle:     "Whack",
		sncB14TorchOracle:     "Torch Breath",
		sncB14EvenScoreOracle: "Even the Score",
		sncB14BodyCountOracle: "Body Count",
		sncB14MetamorOracle:   "Majestic Metamorphosis",
	}
	for oracle, name := range want {
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

// sncB14Cast puts the card in `me`'s hand, floats `pool` (one letter per mana,
// "C" for colourless) and casts it under the strict mana gate.
func sncB14Cast(t *testing.T, g *game.Game, me *game.Player, card game.Card, pool string, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	// An instant cast during combat stays in combat: only a turn still
	// before its first main phase is walked forward.
	if g.Turn.Step != game.StepDeclareAttackers && g.Turn.Step != game.StepDeclareBlockers {
		toMainForCost(t, g)
	}
	card.InstanceID = uuid.New()
	card.Owner, card.Controller = me.ID, me.ID
	me.Hand.PushTop(card)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	floatForTest(g, me, pool)
	params.Strict = true
	return card.InstanceID, g.CastSpell(me.ID, card.InstanceID, params)
}

func sncB14Creature(g *game.Game, owner uuid.UUID, power, toughness int, colors ...string) uuid.UUID {
	return b16Creature(g, owner, "Test Creature", "Creature — Bear", power, toughness, colors...)
}

// --- Whack ----------------------------------------------------------------------

func TestSncB14WhackCostsThreeLessOnlyForAWhiteCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	white := sncB14Creature(g, opp.ID, 4, 4, "W")
	green := sncB14Creature(g, opp.ID, 4, 4, "G")
	card := game.Card{Name: "Whack", TypeLine: "Sorcery", ManaCost: "{3}{B}", OracleID: sncB14WhackOracle}
	if _, err := sncB14Cast(t, g, me, card, "B", game.CastSpellParams{Targets: cardRefs(green)}); err == nil {
		t.Fatal("cast for {B} at a creature that is not white")
	}
	if _, err := sncB14Cast(t, g, me, card, "B", game.CastSpellParams{Targets: cardRefs(white)}); err != nil {
		t.Fatalf("cast for {B} at a white creature: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left over", me.ManaPool.Total())
	}
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, white, game.ZoneGraveyard, "the white 4/4 that got -4/-4")
}

func TestSncB14WhackShrinksForTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	big := sncB14Creature(g, opp.ID, 6, 6, "G")
	card := game.Card{Name: "Whack", TypeLine: "Sorcery", ManaCost: "{3}{B}", OracleID: sncB14WhackOracle}
	if _, err := sncB14Cast(t, g, me, card, "CCCB", game.CastSpellParams{Targets: cardRefs(big)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if p, tg := effectivePower(t, g, big), effectiveToughness(t, g, big); p != 2 || tg != 2 {
		t.Errorf("target is %d/%d, want 2/2", p, tg)
	}
	endTurn(t, g)
	if p := effectivePower(t, g, big); p != 6 {
		t.Errorf("the shrink outlasted the turn: power %d", p)
	}
}

// --- Out of the Way ---------------------------------------------------------------

func TestSncB14OutOfTheWayCostsTwoLessForAGreenPermanentBouncesAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	green := sncB14Creature(g, opp.ID, 2, 2, "G")
	red := sncB14Creature(g, opp.ID, 2, 2, "R")
	mine := sncB14Creature(g, me.ID, 2, 2, "G")
	seedLibrary(me, "Drawn")
	card := game.Card{Name: "Out of the Way", TypeLine: "Instant", ManaCost: "{3}{U}", OracleID: sncB14OutOfWayOracle}
	if _, err := sncB14Cast(t, g, me, card, "CU", game.CastSpellParams{Targets: cardRefs(red)}); err == nil {
		t.Fatal("cast for {1}{U} at a permanent that is not green")
	}
	if _, err := sncB14Cast(t, g, me, card, "CCCU", game.CastSpellParams{Targets: cardRefs(mine)}); err == nil {
		t.Fatal("targeted a permanent its controller owns; it must be an opponent's")
	}
	if _, err := sncB14Cast(t, g, me, card, "CU", game.CastSpellParams{Targets: cardRefs(green)}); err != nil {
		t.Fatalf("cast for {1}{U} at a green permanent: %v", err)
	}
	hand, oppHand := me.Hand.Size(), opp.Hand.Size()
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, green, game.ZoneHand, "the bounced green creature")
	if got := opp.Hand.Size() - oppHand; got != 1 {
		t.Errorf("owner's hand grew by %d, want 1", got)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("caster's hand grew by %d, want 1 (the draw)", got)
	}
}

// --- Knockout Blow ------------------------------------------------------------------

func TestSncB14KnockoutBlowCostsTwoLessForARedCreatureAndHitsForFourAndGainsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	red := sncB14Creature(g, me.ID, 3, 4, "R")
	green := sncB14Creature(g, me.ID, 3, 4, "G")
	idle := sncB14Creature(g, opp.ID, 3, 4, "R")
	advanceToDeclareAttackersOf(t, g, 0)
	for _, id := range []uuid.UUID{red, green} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttacks(t, g)
	passPriorityAroundTable(t, g)
	card := game.Card{Name: "Knockout Blow", TypeLine: "Instant", ManaCost: "{2}{W}", OracleID: sncB14KnockoutOracle}
	if _, err := sncB14Cast(t, g, me, card, "W", game.CastSpellParams{Targets: cardRefs(green)}); err == nil {
		t.Fatal("cast for {W} at a creature that is not red")
	}
	if _, err := sncB14Cast(t, g, me, card, "CCW", game.CastSpellParams{Targets: cardRefs(idle)}); err == nil {
		t.Fatal("targeted a creature that is neither attacking nor blocking")
	}
	life := me.Life
	if _, err := sncB14Cast(t, g, me, card, "W", game.CastSpellParams{Targets: cardRefs(red)}); err != nil {
		t.Fatalf("cast for {W} at a red attacker: %v", err)
	}
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, red, game.ZoneGraveyard, "the red attacker that took 4")
	if me.Life != life+2 {
		t.Errorf("life %d -> %d, want +2", life, me.Life)
	}
}

// --- Torch Breath -----------------------------------------------------------------------

func TestSncB14TorchBreathCostsTwoLessForABluePermanentAndDealsX(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	blue := sncB14Creature(g, opp.ID, 3, 3, "U")
	other := sncB14Creature(g, opp.ID, 3, 3, "G")
	card := game.Card{Name: "Torch Breath", TypeLine: "Instant", ManaCost: "{X}{R}", OracleID: sncB14TorchOracle}
	if _, err := sncB14Cast(t, g, me, card, "CR", game.CastSpellParams{XValue: 3, Targets: cardRefs(other)}); err == nil {
		t.Fatal("X=3 cast for {1}{R} at a permanent that is not blue")
	}
	if _, err := sncB14Cast(t, g, me, card, "CR", game.CastSpellParams{XValue: 3, Targets: cardRefs(blue)}); err != nil {
		t.Fatalf("X=3 cast for {1}{R} at a blue creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, blue, game.ZoneGraveyard, "the blue 3/3 that took X=3")
	// Full price for a non-blue target: {X}{R} with X=2 is {2}{R}.
	if _, err := sncB14Cast(t, g, me, card, "CCR", game.CastSpellParams{XValue: 2, Targets: cardRefs(other)}); err != nil {
		t.Fatalf("X=2 cast for {2}{R}: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, other); got != 2 {
		t.Errorf("damage on the target %d, want 2", got)
	}
}

func TestSncB14TorchBreathCantBeCounteredAndAimsAtCreaturesAndPlaneswalkers(t *testing.T) {
	spec, ok := Lookup(sncB14TorchOracle)
	if !ok || !spec.CantBeCountered {
		t.Fatal("Torch Breath can't be countered")
	}
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	land := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Island", TypeLine: "Basic Land — Island",
		Colors: []string{"U"}, Owner: opp.ID, Controller: opp.ID})
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Planeswalker — Test",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLoyalty: 3}})
	card := game.Card{Name: "Torch Breath", TypeLine: "Instant", ManaCost: "{X}{R}", OracleID: sncB14TorchOracle}
	if _, err := sncB14Cast(t, g, me, card, "CCCCR", game.CastSpellParams{XValue: 3, Targets: cardRefs(land)}); err == nil {
		t.Fatal("targeted a land")
	}
	if _, err := sncB14Cast(t, g, me, card, "CCR", game.CastSpellParams{XValue: 2, Targets: cardRefs(walker)}); err != nil {
		t.Fatalf("cast at a planeswalker: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, walker, game.CounterLoyalty); got != 1 {
		t.Errorf("loyalty %d, want 1", got)
	}
}

// --- Even the Score ---------------------------------------------------------------------

func sncB14OpponentDraws(t *testing.T, g *game.Game, p *game.Player, n int) {
	t.Helper()
	for i := 0; i < n; i++ {
		seedLibrary(p, "Card")
		if err := g.DrawCard(p.ID); err != nil {
			t.Fatalf("DrawCard: %v", err)
		}
	}
}

func TestSncB14EvenTheScoreCostsUUULessOnlyAfterOneOpponentDrewFour(t *testing.T) {
	g := newCatalogGame(t)
	me, opp, third := g.Seats[0], g.Seats[1], g.Seats[2]
	seedLibrary(me, "A", "B", "C")
	card := game.Card{Name: "Even the Score", TypeLine: "Instant", ManaCost: "{X}{U}{U}{U}", OracleID: sncB14EvenScoreOracle}
	// Nobody has drawn: the full {2}{U}{U}{U}.
	if _, err := sncB14Cast(t, g, me, card, "CC", game.CastSpellParams{XValue: 2}); err == nil {
		t.Fatal("cast for {2} with no opponent having drawn")
	}
	// Two opponents drew two each: four in all, but no one opponent drew four.
	sncB14OpponentDraws(t, g, opp, 2)
	sncB14OpponentDraws(t, g, third, 2)
	if _, err := sncB14Cast(t, g, me, card, "CC", game.CastSpellParams{XValue: 2}); err == nil {
		t.Fatal("cast for {2} when the draws were split across two opponents")
	}
	sncB14OpponentDraws(t, g, opp, 2)
	hand := me.Hand.Size()
	if _, err := sncB14Cast(t, g, me, card, "CC", game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast for {2} after an opponent drew four: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("%d mana left over", me.ManaPool.Total())
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		// +1 for the spell pushed into hand, -1 cast, +2 drawn
		t.Errorf("hand grew by %d, want 2", got)
	}
}

func TestSncB14EvenTheScoreDrawsXAtFullPrice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B", "C")
	card := game.Card{Name: "Even the Score", TypeLine: "Instant", ManaCost: "{X}{U}{U}{U}", OracleID: sncB14EvenScoreOracle}
	hand := me.Hand.Size()
	if _, err := sncB14Cast(t, g, me, card, "CCUUU", game.CastSpellParams{XValue: 2}); err != nil {
		t.Fatalf("cast at full price: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size() - hand; got != 2 {
		// +1 for the spell pushed into hand, -1 cast, +2 drawn
		t.Errorf("hand grew by %d, want 2", got)
	}
}

// --- Unleash the Inferno ------------------------------------------------------------------

func TestSncB14UnleashTheInfernoDestroysAnArtifactUpToTheExcess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := sncB14Creature(g, opp.ID, 4, 4, "G")
	cheap := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Cheap", TypeLine: "Artifact", ManaCost: "{3}", Owner: opp.ID, Controller: opp.ID})
	dear := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Dear", TypeLine: "Artifact", ManaCost: "{4}", Owner: opp.ID, Controller: opp.ID})
	mine := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Mine", TypeLine: "Enchantment", ManaCost: "{1}", Owner: me.ID, Controller: me.ID})
	card := game.Card{Name: "Unleash the Inferno", TypeLine: "Instant", ManaCost: "{1}{B}{R}{G}", OracleID: sncB14UnleashOracle}
	if _, err := sncB14Cast(t, g, me, card, "CBRG", game.CastSpellParams{Targets: cardRefs(victim)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, victim, game.ZoneGraveyard, "the 4/4 that took 7")
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("no prompt for the reflexive trigger's target")
	}
	// 7 damage on a 4/4 is 3 excess: the 4-cost artifact and my own enchantment are not offered.
	for _, bad := range []uuid.UUID{dear, mine} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was accepted as the trigger's target", bad)
		}
	}
	pickCard(t, g, me.ID, cheap)
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, cheap, game.ZoneGraveyard, "the artifact with mana value 3")
	p1WantZone(t, g, dear, game.ZoneBattlefield, "the artifact with mana value 4")
}

func TestSncB14UnleashTheInfernoWithNoExcessRaisesNoTrigger(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	victim := sncB14Creature(g, opp.ID, 7, 7, "G")
	art := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Trinket", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	card := game.Card{Name: "Unleash the Inferno", TypeLine: "Instant", ManaCost: "{1}{B}{R}{G}", OracleID: sncB14UnleashOracle}
	if _, err := sncB14Cast(t, g, me, card, "CBRG", game.CastSpellParams{Targets: cardRefs(victim)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, victim, game.ZoneGraveyard, "the 7/7 that took exactly lethal")
	if latestPickTarget(g, me.ID) != nil {
		t.Error("exactly lethal is no excess, but a target was asked for")
	}
	p1WantZone(t, g, art, game.ZoneBattlefield, "the artifact")
}

func TestSncB14UnleashTheInfernoCountsLoyaltyAsTheLimitOnAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Planeswalker — Test",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLoyalty: 5}})
	art := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Trinket", TypeLine: "Artifact", ManaCost: "{2}", Owner: opp.ID, Controller: opp.ID})
	card := game.Card{Name: "Unleash the Inferno", TypeLine: "Instant", ManaCost: "{1}{B}{R}{G}", OracleID: sncB14UnleashOracle}
	if _, err := sncB14Cast(t, g, me, card, "CBRG", game.CastSpellParams{Targets: cardRefs(walker)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	b04WaitForPick(t, g, me.ID)
	pickCard(t, g, me.ID, art)
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, art, game.ZoneGraveyard, "the mana value 2 artifact (7 damage on loyalty 5 is 2 excess)")
}

// --- Corpse Explosion ------------------------------------------------------------------

func TestSncB14CorpseExplosionDealsTheExiledCardsPowerToEveryCreatureAndPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	fuel := me.Graveyard
	id := uuid.New()
	fuel.PushTop(game.Card{InstanceID: id, Name: "Corpse", TypeLine: "Creature — Zombie", Power: 3, Toughness: 1, Owner: me.ID, Controller: me.ID})
	notCreature := uuid.New()
	fuel.PushTop(game.Card{InstanceID: notCreature, Name: "Spell", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	small := sncB14Creature(g, opp.ID, 2, 3)
	big := sncB14Creature(g, me.ID, 5, 5)
	walker := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Walker", TypeLine: "Planeswalker — Test",
		Owner: opp.ID, Controller: opp.ID, Counters: map[string]int{game.CounterLoyalty: 4}})
	card := game.Card{Name: "Corpse Explosion", TypeLine: "Sorcery", ManaCost: "{1}{B}{R}", OracleID: sncB14CorpseOracle}
	if _, err := sncB14Cast(t, g, me, card, "CBR", game.CastSpellParams{ExileIDs: []uuid.UUID{notCreature}}); err == nil {
		t.Fatal("exiled a card that is not a creature card")
	}
	if _, err := sncB14Cast(t, g, me, card, "CBR", game.CastSpellParams{}); err == nil {
		t.Fatal("cast without exiling a creature card")
	}
	if _, err := sncB14Cast(t, g, me, card, "CBR", game.CastSpellParams{ExileIDs: []uuid.UUID{id}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if !g.Exile.Contains(id) {
		t.Error("the creature card was not exiled by the cost")
	}
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, small, game.ZoneGraveyard, "the 2/3 that took 3")
	p1WantZone(t, g, big, game.ZoneBattlefield, "the 5/5 that took 3")
	if got := damageOn(t, g, big); got != 3 {
		t.Errorf("damage on the 5/5 = %d, want 3", got)
	}
	if got := countersOn(g, walker, game.CounterLoyalty); got != 1 {
		t.Errorf("planeswalker loyalty %d, want 1", got)
	}
}

// --- Hostile Takeover ---------------------------------------------------------------------

func TestSncB14HostileTakeoverResizesTwoCreaturesThenDamagesEveryCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	shrunk := sncB14Creature(g, opp.ID, 6, 6)
	grown := sncB14Creature(g, me.ID, 1, 1)
	bystander := sncB14Creature(g, opp.ID, 2, 2)
	hardy := sncB14Creature(g, me.ID, 2, 5)
	card := game.Card{Name: "Hostile Takeover", TypeLine: "Sorcery", ManaCost: "{2}{U}{B}{R}", OracleID: sncB14TakeoverOracle}
	targets := []game.TargetRef{
		{Kind: game.TargetCard, ID: shrunk, Slot: 0},
		{Kind: game.TargetCard, ID: grown, Slot: 1},
	}
	if _, err := sncB14Cast(t, g, me, card, "CCUBR", game.CastSpellParams{Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: shrunk, Slot: 0}, {Kind: game.TargetCard, ID: shrunk, Slot: 1},
	}}); err == nil {
		t.Fatal("one creature was accepted for both clauses; the second must be another creature")
	}
	if _, err := sncB14Cast(t, g, me, card, "CCUBR", game.CastSpellParams{Targets: targets}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	// 6/6 -> 1/1 dies to 3 damage; 1/1 -> 4/4 survives with 3 damage.
	p1WantZone(t, g, shrunk, game.ZoneGraveyard, "the creature made 1/1")
	p1WantZone(t, g, grown, game.ZoneBattlefield, "the creature made 4/4")
	p1WantZone(t, g, bystander, game.ZoneGraveyard, "the untargeted 2/2")
	p1WantZone(t, g, hardy, game.ZoneBattlefield, "the untargeted 2/5")
	if p, tg := effectivePower(t, g, grown), effectiveToughness(t, g, grown); p != 4 || tg != 4 {
		t.Errorf("target is %d/%d, want 4/4", p, tg)
	}
	if got := damageOn(t, g, grown); got != 3 {
		t.Errorf("damage on the 4/4 = %d, want 3", got)
	}
}

func TestSncB14HostileTakeoverResolvesWithNoTargetsAndStillDamages(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	two := sncB14Creature(g, opp.ID, 2, 2)
	card := game.Card{Name: "Hostile Takeover", TypeLine: "Sorcery", ManaCost: "{2}{U}{B}{R}", OracleID: sncB14TakeoverOracle}
	if _, err := sncB14Cast(t, g, me, card, "CCUBR", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast with no targets: %v", err)
	}
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, two, game.ZoneGraveyard, "the 2/2 that took 3")
}

// --- Structural Assault -----------------------------------------------------------------------

func TestSncB14StructuralAssaultCountsEveryArtifactPutIntoAGraveyardThisTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	art1 := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "One", TypeLine: "Artifact", Owner: me.ID, Controller: me.ID})
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Two", TypeLine: "Artifact Creature — Golem", Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	early := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Early", TypeLine: "Artifact", Owner: opp.ID, Controller: opp.ID})
	ench := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Ench", TypeLine: "Enchantment", Owner: opp.ID, Controller: opp.ID})
	bear := sncB14Creature(g, opp.ID, 2, 4)
	tough := sncB14Creature(g, opp.ID, 1, 1)
	// An artifact destroyed earlier this turn counts too.
	p1Destroy(t, g, early)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(ench) })
	card := game.Card{Name: "Structural Assault", TypeLine: "Sorcery", ManaCost: "{3}{R}{R}", OracleID: sncB14AssaultOracle}
	if _, err := sncB14Cast(t, g, me, card, "CCCRR", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, art1, game.ZoneGraveyard, "the artifact")
	// early + art1 + the artifact creature = 3 artifacts: the 2/4 takes 3 and lives; the 1/1 dies.
	if got := damageOn(t, g, bear); got != 3 {
		t.Errorf("damage on the 2/4 = %d, want 3 (three artifacts reached a graveyard)", got)
	}
	p1WantZone(t, g, tough, game.ZoneGraveyard, "the 1/1")
}

func TestSncB14StructuralAssaultWithNoArtifactsDamagesNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := sncB14Creature(g, opp.ID, 1, 1)
	card := game.Card{Name: "Structural Assault", TypeLine: "Sorcery", ManaCost: "{3}{R}{R}", OracleID: sncB14AssaultOracle}
	if _, err := sncB14Cast(t, g, me, card, "CCCRR", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	p1WantZone(t, g, bear, game.ZoneBattlefield, "the creature")
}

// --- Body Count ----------------------------------------------------------------------------------

func TestSncB14BodyCountDrawsForEachCreatureThatDiedUnderYourControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A", "B", "C", "D")
	mine1 := sncB14Creature(g, me.ID, 1, 1)
	mine2 := sncB14Creature(g, me.ID, 1, 1)
	theirs := sncB14Creature(g, opp.ID, 1, 1)
	p1Destroy(t, g, mine1)
	p1Destroy(t, g, mine2)
	p1Destroy(t, g, theirs)
	card := game.Card{Name: "Body Count", TypeLine: "Instant", ManaCost: "{2}{B}", OracleID: sncB14BodyCountOracle}
	hand := me.Hand.Size()
	if _, err := sncB14Cast(t, g, me, card, "CCB", game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	// +1 pushed, -1 cast, +2 drawn: only my two creatures count.
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("hand grew by %d, want 2", got)
	}
}

func TestSncB14BodyCountSpectacleNeedsAnOpponentToHaveLostLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "A")
	card := game.Card{Name: "Body Count", TypeLine: "Instant", ManaCost: "{2}{B}", OracleID: sncB14BodyCountOracle}
	if _, err := sncB14Cast(t, g, me, card, "B", game.CastSpellParams{AlternativeCost: "spectacle"}); err == nil {
		t.Fatal("spectacle was accepted with no life lost this turn")
	}
	g.WithWriteLock(func() { _ = g.ChangePlayerLifeForEffect(uuid.Nil, opp.ID, -1) })
	if _, err := sncB14Cast(t, g, me, card, "B", game.CastSpellParams{AlternativeCost: "spectacle"}); err != nil {
		t.Fatalf("spectacle after an opponent lost life: %v", err)
	}
	passPriorityAroundTable(t, g)
}

// --- Majestic Metamorphosis ----------------------------------------------------------------------

func TestSncB14MajesticMetamorphosisMakesA4x4AngelArtifactCreatureWithFlyingAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ench := pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Relic", TypeLine: "Artifact — Equipment", Owner: me.ID, Controller: me.ID})
	seedLibrary(me, "Drawn")
	card := game.Card{Name: "Majestic Metamorphosis", TypeLine: "Instant", ManaCost: "{2}{U}", OracleID: sncB14MetamorOracle}
	hand := me.Hand.Size()
	if _, err := sncB14Cast(t, g, me, card, "CCU", game.CastSpellParams{Targets: cardRefs(ench)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !hasEffectiveType(t, g, ench, "Creature") || !hasEffectiveType(t, g, ench, "Artifact") {
		t.Error("the target is not an artifact creature")
	}
	if p, tg := effectivePower(t, g, ench), effectiveToughness(t, g, ench); p != 4 || tg != 4 {
		t.Errorf("target is %d/%d, want 4/4", p, tg)
	}
	if !hasEffectiveKeyword(t, g, ench, "flying") {
		t.Error("the target did not gain flying")
	}
	var angel bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == ench {
				e := c.Effective()
				for _, s := range e.Subtypes {
					angel = angel || s == "Angel"
				}
			}
		}
	})
	if !angel {
		t.Error("the target is not an Angel")
	}
	if got := me.Hand.Size() - hand; got != 1 {
		// +1 for the spell pushed into hand, -1 cast, +1 drawn
		t.Errorf("hand changed by %d, want 1", got)
	}
	endTurn(t, g)
	if hasEffectiveType(t, g, ench, "Creature") {
		t.Error("the effect outlasted the turn")
	}
}

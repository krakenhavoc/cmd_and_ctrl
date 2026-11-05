package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Star Trek batch 14 (#4973): each card's printed text, driven through the
// engine.

const (
	trekB14TalShiarOracle    = "53641cda-7093-4fed-97f1-380f771e00b0"
	trekB14DroneOracle       = "af6351a9-ac01-4fed-9ea9-a352d21a8312"
	trekB14LoyaltyOracle     = "6cb1c3fb-033a-4de5-ad07-69d9ad0ca5ae"
	trekB14ReliantOracle     = "edaa05ba-8b15-4156-b725-1e0c418a9ae5"
	trekB14VengeanceOracle   = "eae07535-487b-48ac-b99e-b4b62286254c"
	trekB14InstructorOracle  = "8802b7f0-1ed1-4872-9939-182afb8441ca"
	trekB14MarinerOracle     = "1033a0cb-1a00-46c4-83d5-68a6eac85feb"
	trekB14BoardingOracle    = "f9f70fb6-953e-40e8-a15d-ea3ee62d09e3"
	trekB14BluffOracle       = "9ac20f38-adc4-40a9-95c9-bbdc689b928b"
	trekB14Bear              = "Creature — Bear"
	trekB14ReliantTypeLine   = "Legendary Artifact — Spacecraft"
	trekB14InstructorTypes   = "Creature — Bolian Officer"
	trekB14MarinerTypeLine   = "Legendary Creature — Human Officer"
	trekB14VengeanceTypeLine = "Enchantment"
)

func TestTrekB14EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		trekB14TalShiarOracle:   "Tal Shiar Operative",
		trekB14DroneOracle:      "Tenacious Drone",
		trekB14LoyaltyOracle:    "Test of Loyalty",
		trekB14ReliantOracle:    "U.S.S. Reliant, From Hell's Heart",
		trekB14VengeanceOracle:  "Vengeance Mission",
		trekB14InstructorOracle: "Academy Instructor",
		trekB14MarinerOracle:    "Beckett Mariner, Impetuous Ensign",
		trekB14BoardingOracle:   "Boarding Action",
		trekB14BluffOracle:      "Brazen Bluff",
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

// trekB14PickPlayer answers chooser's open pick_target prompt with a
// player.
func trekB14PickPlayer(t *testing.T, g *game.Game, chooser, player uuid.UUID) {
	t.Helper()
	p := latestPickTarget(g, chooser)
	if p == nil {
		t.Fatalf("no pick_target prompt for %s: %+v", chooser, g.PendingChoices)
	}
	if err := g.ResolvePickTarget(p.ID, chooser, game.TargetRef{Kind: game.TargetPlayer, ID: player}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
}

// --- Tal Shiar Operative ---

func trekB14OperativeAttacks(t *testing.T, g *game.Game, mode int) uuid.UUID {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	op := b12Push(g, me.ID, "Tal Shiar Operative", "Creature — Romulan Officer Spy", trekB14TalShiarOracle, 2, 2)
	declareAttack(t, g, opp.ID, op)
	mp := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mp == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(mp.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick(%d): %v", mode, err)
	}
	return op
}

func TestTrekB14OperativeGrantsDeathtouchToTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	op := trekB14OperativeAttacks(t, g, 0)
	pickCard(t, g, g.Seats[0].ID, op)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, op, "deathtouch") {
		t.Error("the targeted creature did not gain deathtouch")
	}
}

func TestTrekB14OperativeDeathtouchRefusesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	theirs := b12Creature(g, g.Seats[1].ID, "Theirs", trekB14Bear, 2, 2)
	trekB14OperativeAttacks(t, g, 0)
	p := latestPickTarget(g, g.Seats[0].ID)
	if p == nil {
		t.Fatal("no pick_target prompt for the deathtouch bullet")
	}
	if err := g.ResolvePickTarget(p.ID, g.Seats[0].ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Error("an opponent's creature was accepted as \"target creature you control\"")
	}
}

func TestTrekB14OperativeSurveilsOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "A", "B")
	trekB14OperativeAttacks(t, g, 1)
	passPriorityAroundTable(t, g)
	c := surveilChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no surveil prompt: %+v", g.PendingChoices)
	}
	if len(c.ScryCards) != 1 {
		t.Fatalf("surveil looked at %d cards, want 1", len(c.ScryCards))
	}
	top := c.ScryCards[0]
	if err := g.ResolveSurveil(c.ID, me.ID, c.ScryCards, nil); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	if !me.Graveyard.Contains(top) {
		t.Error("the surveilled card is not in the graveyard")
	}
}

// --- Tenacious Drone ---

func TestTrekB14DroneMakesABorgWhenItDies(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	drone := b12Push(g, me.ID, "Tenacious Drone", "Artifact Creature — Borg", trekB14DroneOracle, 2, 1)
	killCreature(t, g, me.ID, drone)
	passPriorityAroundTable(t, g)
	borgs := tokensNamed(g, "Borg")
	if len(borgs) != 1 {
		t.Fatalf("Borg tokens = %d, want 1", len(borgs))
	}
	b := layeredCard(t, g, borgs[0].InstanceID)
	if b.Controller != me.ID || !b.IsArtifact() || !b.IsCreature() || len(b.Colors) != 0 ||
		b.CurrentPower() != 1 || b.CurrentToughness() != 1 {
		t.Errorf("Borg token = %+v, want a 1/1 colorless artifact creature I control", b)
	}
}

func TestTrekB14DroneMakesNothingWhenExiled(t *testing.T) {
	g := newCatalogGame(t)
	drone := b12Push(g, g.Seats[0].ID, "Tenacious Drone", "Artifact Creature — Borg", trekB14DroneOracle, 2, 1)
	exileCreature(t, g, drone)
	passPriorityAroundTable(t, g)
	if n := len(tokensNamed(g, "Borg")); n != 0 {
		t.Errorf("an exiled Drone made %d Borg tokens", n)
	}
}

// --- Test of Loyalty ---

func trekB14CastLoyalty(g *game.Game, mode int, target uuid.UUID) error {
	me := g.Seats[0]
	id := handCardFull(me, "Test of Loyalty", "Sorcery", "{B}", trekB14LoyaltyOracle, []string{"B"})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{mode}, Targets: cardRefs(target)})
}

func TestTrekB14LoyaltyGarakExilesAGraveyardCardAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	seedLibrary(me, "Drawn")
	theirs := graveyardCardOf(g, opp, "Their Card", "Instant", "{U}", "")
	floatMana(t, g, me, "{B}")
	if err := trekB14CastLoyalty(g, 0, theirs); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(theirs); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the target is in %v, want exile", z)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d after the draw", me.Hand.Size(), hand+1)
	}
}

func TestTrekB14LoyaltyObsidianOrderReturnsACreatureCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	bear := graveyardCardOf(g, me, "Bear", trekB14Bear, "{1}{G}", "")
	floatMana(t, g, me, "{B}")
	if err := trekB14CastLoyalty(g, 1, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature card did not return to hand")
	}
}

func TestTrekB14LoyaltyObsidianOrderRefusesOtherCards(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	instant := graveyardCardOf(g, me, "My Instant", "Instant", "{U}", "")
	theirBear := graveyardCardOf(g, opp, "Their Bear", trekB14Bear, "{1}{G}", "")
	floatMana(t, g, me, "{B}")
	if err := trekB14CastLoyalty(g, 1, instant); err == nil {
		t.Error("a noncreature card was accepted as the target")
	}
	if err := trekB14CastLoyalty(g, 1, theirBear); err == nil {
		t.Error("an opponent's creature card was accepted as one from your graveyard")
	}
}

// --- U.S.S. Reliant, From Hell's Heart ---

// trekB14GraveyardCreature puts a 2/2 creature card with the given mana
// cost into p's graveyard.
func trekB14GraveyardCreature(p *game.Player, name, cost string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: trekB14Bear, ManaCost: cost,
		Power: 2, Toughness: 2, Owner: p.ID, Controller: p.ID})
	return id
}

func TestTrekB14ReliantEntersAndReturnsACreatureWithAFinalityCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	big := trekB14GraveyardCreature(me, "Big", "{3}{G}{G}")
	four := trekB14GraveyardCreature(me, "Four", "{2}{G}{G}")
	ship := handCardFull(me, "U.S.S. Reliant, From Hell's Heart", trekB14ReliantTypeLine, "{2}{B}{B}", trekB14ReliantOracle, []string{"B"})
	floatMana(t, g, me, "{B}{B}{C}{C}")
	if err := g.CastSpell(me.ID, ship, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no target prompt for the enters trigger: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
		t.Fatal("a mana value 5 creature card was accepted")
	}
	pickCard(t, g, me.ID, four)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(four) {
		t.Fatal("the creature card did not return to the battlefield")
	}
	if got := counterOn(g, four, game.CounterFinality); got != 1 {
		t.Errorf("finality counters = %d, want 1", got)
	}
	if isCreatureNow(g, ship) {
		t.Error("the Reliant is a creature with no charge counters")
	}
}

func TestTrekB14ReliantStationsToAFlyingCreatureAndTriggersOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	ship := b12Push(g, me.ID, "U.S.S. Reliant, From Hell's Heart", trekB14ReliantTypeLine, trekB14ReliantOracle, 4, 3)
	crew := b12Creature(g, me.ID, "Crew", "Creature — Human", 10, 10)
	station(t, g, me, ship, crew)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, ship, "charge"); got != 10 {
		t.Fatalf("charge counters = %d, want 10", got)
	}
	if !isCreatureNow(g, ship) || !hasEffectiveKeyword(t, g, ship, "flying") {
		t.Fatal("at ten charge counters the Reliant is not a flying creature")
	}
	if pw := effectivePower(t, g, ship); pw != 4 {
		t.Errorf("power %d, want 4", pw)
	}
	bear := trekB14GraveyardCreature(me, "Bear", "{1}{G}")
	battlefieldCardPtr(t, g, ship).SummonedThisTurn = false
	life := opp.Life
	declareAttack(t, g, opp.ID, ship)
	advanceTo(t, g, game.StepCombatDamage)
	pickCard(t, g, me.ID, bear)
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-4)
	}
	if !g.Battlefield.Contains(bear) || counterOn(g, bear, game.CounterFinality) != 1 {
		t.Error("combat damage did not return the creature card with a finality counter")
	}
}

// --- Vengeance Mission ---

func TestTrekB14VengeanceMissionCountsAndDrains(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mission := b12Push(g, me.ID, "Vengeance Mission", trekB14VengeanceTypeLine, trekB14VengeanceOracle, 0, 0)
	theirs := b12Creature(g, opp.ID, "Theirs", trekB14Bear, 2, 2)
	life := opp.Life
	killCreature(t, g, opp.ID, theirs)
	trekB14PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, mission, "quest"); got != 1 {
		t.Errorf("quest counters = %d, want 1", got)
	}
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
}

func TestTrekB14VengeanceMissionLosesTwoFromTheSeventhCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mission := b12Push(g, me.ID, "Vengeance Mission", trekB14VengeanceTypeLine, trekB14VengeanceOracle, 0, 0)
	battlefieldCardPtr(t, g, mission).Counters = map[string]int{"quest": 6}
	mine := b12Creature(g, me.ID, "Mine", trekB14Bear, 2, 2)
	life := opp.Life
	killCreature(t, g, me.ID, mine)
	trekB14PickPlayer(t, g, me.ID, opp.ID)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, mission, "quest"); got != 7 {
		t.Errorf("quest counters = %d, want 7", got)
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d: the seventh counter makes it two", opp.Life, life-2)
	}
}

func TestTrekB14VengeanceMissionRefusesYouAsTheTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	b12Push(g, me.ID, "Vengeance Mission", trekB14VengeanceTypeLine, trekB14VengeanceOracle, 0, 0)
	mine := b12Creature(g, me.ID, "Mine", trekB14Bear, 2, 2)
	killCreature(t, g, me.ID, mine)
	p := latestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no pick_target prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Error("the controller was accepted as \"target opponent\"")
	}
}

// --- Academy Instructor ---

func TestTrekB14InstructorMakesOfficersOnNoncreatureCastsAndCountsThem(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	inst := b12Push(g, me.ID, "Academy Instructor", trekB14InstructorTypes, trekB14InstructorOracle, 0, 3)
	if pw := effectivePower(t, g, inst); pw != 0 {
		t.Fatalf("power %d with no Officer tokens, want 0", pw)
	}
	creature := handCardFull(me, "Bear", trekB14Bear, "", "", nil)
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast creature: %v", err)
	}
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Officer", me.ID); n != 0 {
		t.Fatalf("a creature spell made %d Officers", n)
	}
	for i := 0; i < 2; i++ {
		sorcery := handCardFull(me, "Some Sorcery", "Sorcery", "", "", nil)
		if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast sorcery: %v", err)
		}
		passPriorityAroundTable(t, g)
	}
	officers := tokensNamed(g, "Officer")
	if len(officers) != 2 {
		t.Fatalf("Officer tokens = %d, want 2", len(officers))
	}
	o := layeredCard(t, g, officers[0].InstanceID)
	if !o.IsCreature() || o.CurrentPower() != 1 || o.CurrentToughness() != 1 || len(o.Colors) != 1 || o.Colors[0] != "R" {
		t.Errorf("Officer = %+v, want a 1/1 red creature", o)
	}
	if pw := effectivePower(t, g, inst); pw != 2 {
		t.Errorf("power %d with two Officer tokens, want 2", pw)
	}
	// An opponent's Officer token does not count.
	opp := g.Seats[1]
	var err error
	g.WithWriteLock(func() { err = g.CreateTokenForEffect(opp.ID, TokenCard("1/1 red Officer"), 1) })
	if err != nil {
		t.Fatalf("CreateTokenForEffect: %v", err)
	}
	if pw := effectivePower(t, g, inst); pw != 2 {
		t.Errorf("power %d after an opponent's Officer, want 2", pw)
	}
}

// --- Beckett Mariner, Impetuous Ensign ---

func TestTrekB14MarinerPromotesOnEachTapAndPaysOffAtThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	life := me.Life
	beckett := b12Push(g, me.ID, "Beckett Mariner, Impetuous Ensign", trekB14MarinerTypeLine, trekB14MarinerOracle, 2, 1)
	if !hasEffectiveKeyword(t, g, beckett, "first strike") {
		t.Error("no first strike")
	}
	lives := lifeOfOpponents(g)
	for i := 1; i <= 2; i++ {
		tapForTest(t, g, beckett)
		passPriorityAroundTable(t, g)
		if got := counterOn(g, beckett, beckettMarinerCounter); got != i {
			t.Fatalf("after tap %d promotion counters = %d", i, got)
		}
		if err := g.TapCard(beckett, false); err != nil {
			t.Fatalf("untap: %v", err)
		}
	}
	for i, life := range lifeOfOpponents(g) {
		if life != lives[i] {
			t.Fatalf("opponent %d lost life before the third counter", i)
		}
	}
	tapForTest(t, g, beckett)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, beckett, beckettMarinerCounter); got != 0 {
		t.Errorf("promotion counters = %d after the third, want them removed", got)
	}
	for i, life := range lifeOfOpponents(g) {
		if life != lives[i]-2 {
			t.Errorf("opponent %d life %d, want %d", i, life, lives[i]-2)
		}
	}
	if me.Life != life {
		t.Errorf("controller life %d, want %d: only opponents are dealt damage", me.Life, life)
	}
}

func TestTrekB14MarinerAttackingTapsAndPromotes(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	beckett := b12Push(g, me.ID, "Beckett Mariner, Impetuous Ensign", trekB14MarinerTypeLine, trekB14MarinerOracle, 2, 1)
	declareAttack(t, g, opp.ID, beckett)
	passPriorityAroundTable(t, g)
	if got := counterOn(g, beckett, beckettMarinerCounter); got != 1 {
		t.Errorf("promotion counters after attacking = %d, want 1", got)
	}
}

// --- Boarding Action ---

func TestTrekB14BoardingActionMakesAHastyKlingon(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	toMainForCost(t, g)
	id := handCardFull(me, "Boarding Action", "Sorcery", "{1}{R}", trekB14BoardingOracle, []string{"R"})
	floatMana(t, g, me, "{R}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	ks := tokensNamed(g, "Klingon Warrior")
	if len(ks) != 1 {
		t.Fatalf("Klingon Warrior tokens = %d, want 1", len(ks))
	}
	k := layeredCard(t, g, ks[0].InstanceID)
	if k.CurrentPower() != 2 || k.CurrentToughness() != 1 || len(k.Colors) != 1 || k.Colors[0] != "R" ||
		!hasEffectiveKeyword(t, g, k.InstanceID, "haste") {
		t.Errorf("Klingon Warrior = %+v, want a 2/1 red creature with haste", k)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("cast from hand, Boarding Action should go to the graveyard")
	}
}

func TestTrekB14BoardingActionBeamsUpFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	card := graveyardCardOf(g, me, "Boarding Action", "Sorcery", "{1}{R}", trekB14BoardingOracle)
	floatMana(t, g, me, "{R}{C}{C}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: trekB14AltCostKeyBeamMeUp,
	}); err == nil {
		t.Fatal("Beam me up was cast without returning a creature")
	}
	theirs := b12Creature(g, opp.ID, "Theirs", trekB14Bear, 2, 2)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: trekB14AltCostKeyBeamMeUp,
		AltCostIDs: []uuid.UUID{theirs},
	}); err == nil {
		t.Fatal("Beam me up returned an opponent's creature as its cost")
	}
	mine := b12Creature(g, me.ID, "Mine", trekB14Bear, 2, 2)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: trekB14AltCostKeyBeamMeUp,
		AltCostIDs: []uuid.UUID{mine},
	}); err != nil {
		t.Fatalf("Beam me up: %v", err)
	}
	if !me.Hand.Contains(mine) {
		t.Fatal("the returned creature is not in its owner's hand")
	}
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Klingon Warrior", me.ID); n != 1 {
		t.Errorf("Klingon Warrior tokens = %d, want 1", n)
	}
	if z := g.FindCardZoneForEffect(card); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("Boarding Action after a Beam me up cast is in %v, want exile", z)
	}
}

// --- Brazen Bluff ---

// trekB14BluffBoard puts an opponent's noncreature spell on the stack and
// Brazen Bluff on top of it, targeting it.
func trekB14BluffBoard(t *testing.T, g *game.Game) (theirSpell uuid.UUID) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	theirSpell = handCardFull(opp, "Their Instant", "Instant", "{U}", "", []string{"U"})
	if err := g.CastSpell(opp.ID, theirSpell, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent cast: %v", err)
	}
	bluff := handCardFull(me, "Brazen Bluff", "Instant", "{R}{R}", trekB14BluffOracle, []string{"R"})
	floatMana(t, g, me, "{R}{R}")
	if err := g.CastSpell(me.ID, bluff, game.CastSpellParams{Strict: true, Targets: cardRefs(theirSpell)}); err != nil {
		t.Fatalf("cast Brazen Bluff: %v", err)
	}
	passPriorityAroundTable(t, g)
	return theirSpell
}

func TestTrekB14BluffTheyExileTheirSpell(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	free := handCardFull(me, "Free Sorcery", "Sorcery", "{5}", "", nil)
	theirs := trekB14BluffBoard(t, g)
	answerMayChoice(t, g, opp.ID, true)
	passPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(theirs); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("their spell is in %v, want exile", z)
	}
	if resolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a free cast was offered although they exiled the spell")
	}
	if !me.Hand.Contains(free) {
		t.Error("the hand card moved")
	}
}

func TestTrekB14BluffTheyRefuseAndYouCastANoncreatureSpellFree(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	free := handCardFull(me, "Free Sorcery", "Sorcery", "{5}", "", nil)
	creature := handCardFull(me, "Big Bear", trekB14Bear, "{4}{G}", "", []string{"G"})
	theirs := trekB14BluffBoard(t, g)
	answerMayChoice(t, g, opp.ID, false)
	offer := mustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, free)
	if hasID(offer.DiscardOptions, creature) {
		t.Error("a creature card was offered as a noncreature spell")
	}
	if err := castThroughResolutionCast(t, g, me.ID, free, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	castInsideTheResolution(t, g, free)
	if z := g.FindCardZoneForEffect(theirs); z == nil || z.Kind == game.ZoneExile {
		t.Errorf("their spell is in %v; refusing must leave it alone", z)
	}
}

func TestTrekB14BluffRefusesACreatureSpellOrYourOwn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	toMainForCost(t, g)
	theirCreature := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: theirCreature, Name: "Their Bear", TypeLine: trekB14Bear, ManaCost: "{1}{G}",
		Keywords: []string{"flash"}, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, theirCreature, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent cast: %v", err)
	}
	mine := handCardFull(me, "My Instant", "Instant", "{U}", "", []string{"U"})
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{}); err != nil {
		t.Fatalf("my cast: %v", err)
	}
	for _, target := range []uuid.UUID{theirCreature, mine} {
		bluff := handCardFull(me, "Brazen Bluff", "Instant", "{R}{R}", trekB14BluffOracle, []string{"R"})
		floatMana(t, g, me, "{R}{R}")
		if err := g.CastSpell(me.ID, bluff, game.CastSpellParams{Strict: true, Targets: cardRefs(target)}); err == nil {
			t.Errorf("Brazen Bluff accepted %s as its target", target)
		}
	}
}

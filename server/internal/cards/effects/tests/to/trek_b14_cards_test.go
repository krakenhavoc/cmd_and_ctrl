package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsb "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/b"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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

// trekB14PickPlayer answers chooser's open pick_target prompt with a
// player.
func trekB14PickPlayer(t *testing.T, g *game.Game, chooser, player uuid.UUID) {
	t.Helper()
	p := effectstest.LatestPickTarget(g, chooser)
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
	effectstest.AdvanceToMain(t, g)
	op := effectstest.B12Push(g, me.ID, "Tal Shiar Operative", "Creature — Romulan Officer Spy", trekB14TalShiarOracle, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, op)
	mp := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if mp == nil {
		t.Fatalf("no mode_pick prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(mp.ID, me.ID, []int{mode}); err != nil {
		t.Fatalf("ResolveModePick(%d): %v", mode, err)
	}
	return op
}

func TestTrekB14OperativeGrantsDeathtouchToTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	op := trekB14OperativeAttacks(t, g, 0)
	effectstest.PickCard(t, g, g.Seats[0].ID, op)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, op, "deathtouch") {
		t.Error("the targeted creature did not gain deathtouch")
	}
}

func TestTrekB14OperativeDeathtouchRefusesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	theirs := effectstest.B12Creature(g, g.Seats[1].ID, "Theirs", trekB14Bear, 2, 2)
	trekB14OperativeAttacks(t, g, 0)
	p := effectstest.LatestPickTarget(g, g.Seats[0].ID)
	if p == nil {
		t.Fatal("no pick_target prompt for the deathtouch bullet")
	}
	if err := g.ResolvePickTarget(p.ID, g.Seats[0].ID, game.TargetRef{Kind: game.TargetCard, ID: theirs}); err == nil {
		t.Error("an opponent's creature was accepted as \"target creature you control\"")
	}
}

func TestTrekB14OperativeSurveilsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B")
	trekB14OperativeAttacks(t, g, 1)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SurveilChoiceFor(g, me.ID)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	drone := effectstest.B12Push(g, me.ID, "Tenacious Drone", "Artifact Creature — Borg", trekB14DroneOracle, 2, 1)
	effectstest.KillCreature(t, g, me.ID, drone)
	effectstest.PassPriorityAroundTable(t, g)
	borgs := effectstest.TokensNamed(g, "Borg")
	if len(borgs) != 1 {
		t.Fatalf("Borg tokens = %d, want 1", len(borgs))
	}
	b := effectstest.LayeredCard(t, g, borgs[0].InstanceID)
	if b.Controller != me.ID || !b.IsArtifact() || !b.IsCreature() || len(b.Colors) != 0 ||
		b.CurrentPower() != 1 || b.CurrentToughness() != 1 {
		t.Errorf("Borg token = %+v, want a 1/1 colorless artifact creature I control", b)
	}
}

func TestTrekB14DroneMakesNothingWhenExiled(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	drone := effectstest.B12Push(g, g.Seats[0].ID, "Tenacious Drone", "Artifact Creature — Borg", trekB14DroneOracle, 2, 1)
	effectstest.ExileCreature(t, g, drone)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.TokensNamed(g, "Borg")); n != 0 {
		t.Errorf("an exiled Drone made %d Borg tokens", n)
	}
}

// --- Test of Loyalty ---

func trekB14CastLoyalty(g *game.Game, mode int, target uuid.UUID) error {
	me := g.Seats[0]
	id := effectstest.HandCardFull(me, "Test of Loyalty", "Sorcery", "{B}", trekB14LoyaltyOracle, []string{"B"})
	return g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Modes: []int{mode}, Targets: effectstest.CardRefs(target)})
}

func TestTrekB14LoyaltyGarakExilesAGraveyardCardAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	effectstest.SeedLibrary(me, "Drawn")
	theirs := effectstest.GraveyardCardOf(g, opp, "Their Card", "Instant", "{U}", "")
	effectstest.FloatMana(t, g, me, "{B}")
	if err := trekB14CastLoyalty(g, 0, theirs); err != nil {
		t.Fatalf("cast: %v", err)
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(theirs); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("the target is in %v, want exile", z)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d after the draw", me.Hand.Size(), hand+1)
	}
}

func TestTrekB14LoyaltyObsidianOrderReturnsACreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	bear := effectstest.GraveyardCardOf(g, me, "Bear", trekB14Bear, "{1}{G}", "")
	effectstest.FloatMana(t, g, me, "{B}")
	if err := trekB14CastLoyalty(g, 1, bear); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(bear) {
		t.Error("the creature card did not return to hand")
	}
}

func TestTrekB14LoyaltyObsidianOrderRefusesOtherCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	instant := effectstest.GraveyardCardOf(g, me, "My Instant", "Instant", "{U}", "")
	theirBear := effectstest.GraveyardCardOf(g, opp, "Their Bear", trekB14Bear, "{1}{G}", "")
	effectstest.FloatMana(t, g, me, "{B}")
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	big := trekB14GraveyardCreature(me, "Big", "{3}{G}{G}")
	four := trekB14GraveyardCreature(me, "Four", "{2}{G}{G}")
	ship := effectstest.HandCardFull(me, "U.S.S. Reliant, From Hell's Heart", trekB14ReliantTypeLine, "{2}{B}{B}", trekB14ReliantOracle, []string{"B"})
	effectstest.FloatMana(t, g, me, "{B}{B}{C}{C}")
	if err := g.CastSpell(me.ID, ship, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no target prompt for the enters trigger: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: big}); err == nil {
		t.Fatal("a mana value 5 creature card was accepted")
	}
	effectstest.PickCard(t, g, me.ID, four)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(four) {
		t.Fatal("the creature card did not return to the battlefield")
	}
	if got := effectstest.CounterOn(g, four, game.CounterFinality); got != 1 {
		t.Errorf("finality counters = %d, want 1", got)
	}
	if effectstest.IsCreatureNow(g, ship) {
		t.Error("the Reliant is a creature with no charge counters")
	}
}

func TestTrekB14ReliantStationsToAFlyingCreatureAndTriggersOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	ship := effectstest.B12Push(g, me.ID, "U.S.S. Reliant, From Hell's Heart", trekB14ReliantTypeLine, trekB14ReliantOracle, 4, 3)
	crew := effectstest.B12Creature(g, me.ID, "Crew", "Creature — Human", 10, 10)
	effectstest.Station(t, g, me, ship, crew)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, ship, "charge"); got != 10 {
		t.Fatalf("charge counters = %d, want 10", got)
	}
	if !effectstest.IsCreatureNow(g, ship) || !effectstest.HasEffectiveKeyword(t, g, ship, "flying") {
		t.Fatal("at ten charge counters the Reliant is not a flying creature")
	}
	if pw := effectstest.EffectivePower(t, g, ship); pw != 4 {
		t.Errorf("power %d, want 4", pw)
	}
	bear := trekB14GraveyardCreature(me, "Bear", "{1}{G}")
	effectstest.BattlefieldCardPtr(t, g, ship).SummonedThisTurn = false
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, ship)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PickCard(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-4)
	}
	if !g.Battlefield.Contains(bear) || effectstest.CounterOn(g, bear, game.CounterFinality) != 1 {
		t.Error("combat damage did not return the creature card with a finality counter")
	}
}

// --- Vengeance Mission ---

func TestTrekB14VengeanceMissionCountsAndDrains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mission := effectstest.B12Push(g, me.ID, "Vengeance Mission", trekB14VengeanceTypeLine, trekB14VengeanceOracle, 0, 0)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", trekB14Bear, 2, 2)
	life := opp.Life
	effectstest.KillCreature(t, g, opp.ID, theirs)
	trekB14PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, mission, "quest"); got != 1 {
		t.Errorf("quest counters = %d, want 1", got)
	}
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
}

func TestTrekB14VengeanceMissionLosesTwoFromTheSeventhCounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	mission := effectstest.B12Push(g, me.ID, "Vengeance Mission", trekB14VengeanceTypeLine, trekB14VengeanceOracle, 0, 0)
	effectstest.BattlefieldCardPtr(t, g, mission).Counters = map[string]int{"quest": 6}
	mine := effectstest.B12Creature(g, me.ID, "Mine", trekB14Bear, 2, 2)
	life := opp.Life
	effectstest.KillCreature(t, g, me.ID, mine)
	trekB14PickPlayer(t, g, me.ID, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, mission, "quest"); got != 7 {
		t.Errorf("quest counters = %d, want 7", got)
	}
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d: the seventh counter makes it two", opp.Life, life-2)
	}
}

func TestTrekB14VengeanceMissionRefusesYouAsTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.B12Push(g, me.ID, "Vengeance Mission", trekB14VengeanceTypeLine, trekB14VengeanceOracle, 0, 0)
	mine := effectstest.B12Creature(g, me.ID, "Mine", trekB14Bear, 2, 2)
	effectstest.KillCreature(t, g, me.ID, mine)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatalf("no pick_target prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetPlayer, ID: me.ID}); err == nil {
		t.Error("the controller was accepted as \"target opponent\"")
	}
}

// --- Academy Instructor ---

func TestTrekB14InstructorMakesOfficersOnNoncreatureCastsAndCountsThem(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	inst := effectstest.B12Push(g, me.ID, "Academy Instructor", trekB14InstructorTypes, trekB14InstructorOracle, 0, 3)
	if pw := effectstest.EffectivePower(t, g, inst); pw != 0 {
		t.Fatalf("power %d with no Officer tokens, want 0", pw)
	}
	creature := effectstest.HandCardFull(me, "Bear", trekB14Bear, "", "", nil)
	if err := g.CastSpell(me.ID, creature, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Officer", me.ID); n != 0 {
		t.Fatalf("a creature spell made %d Officers", n)
	}
	for i := 0; i < 2; i++ {
		sorcery := effectstest.HandCardFull(me, "Some Sorcery", "Sorcery", "", "", nil)
		if err := g.CastSpell(me.ID, sorcery, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast sorcery: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	officers := effectstest.TokensNamed(g, "Officer")
	if len(officers) != 2 {
		t.Fatalf("Officer tokens = %d, want 2", len(officers))
	}
	o := effectstest.LayeredCard(t, g, officers[0].InstanceID)
	if !o.IsCreature() || o.CurrentPower() != 1 || o.CurrentToughness() != 1 || len(o.Colors) != 1 || o.Colors[0] != "R" {
		t.Errorf("Officer = %+v, want a 1/1 red creature", o)
	}
	if pw := effectstest.EffectivePower(t, g, inst); pw != 2 {
		t.Errorf("power %d with two Officer tokens, want 2", pw)
	}
	// An opponent's Officer token does not count.
	opp := g.Seats[1]
	var err error
	g.WithWriteLock(func() { err = g.CreateTokenForEffect(opp.ID, effects.TokenCard("1/1 red Officer"), 1) })
	if err != nil {
		t.Fatalf("CreateTokenForEffect: %v", err)
	}
	if pw := effectstest.EffectivePower(t, g, inst); pw != 2 {
		t.Errorf("power %d after an opponent's Officer, want 2", pw)
	}
}

// --- Beckett Mariner, Impetuous Ensign ---

func TestTrekB14MarinerPromotesOnEachTapAndPaysOffAtThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	life := me.Life
	beckett := effectstest.B12Push(g, me.ID, "Beckett Mariner, Impetuous Ensign", trekB14MarinerTypeLine, trekB14MarinerOracle, 2, 1)
	if !effectstest.HasEffectiveKeyword(t, g, beckett, "first strike") {
		t.Error("no first strike")
	}
	lives := effectstest.LifeOfOpponents(g)
	for i := 1; i <= 2; i++ {
		effectstest.TapForTest(t, g, beckett)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.CounterOn(g, beckett, cardsb.BeckettMarinerCounter); got != i {
			t.Fatalf("after tap %d promotion counters = %d", i, got)
		}
		if err := g.TapCard(beckett, false); err != nil {
			t.Fatalf("untap: %v", err)
		}
	}
	for i, life := range effectstest.LifeOfOpponents(g) {
		if life != lives[i] {
			t.Fatalf("opponent %d lost life before the third counter", i)
		}
	}
	effectstest.TapForTest(t, g, beckett)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, beckett, cardsb.BeckettMarinerCounter); got != 0 {
		t.Errorf("promotion counters = %d after the third, want them removed", got)
	}
	for i, life := range effectstest.LifeOfOpponents(g) {
		if life != lives[i]-2 {
			t.Errorf("opponent %d life %d, want %d", i, life, lives[i]-2)
		}
	}
	if me.Life != life {
		t.Errorf("controller life %d, want %d: only opponents are dealt damage", me.Life, life)
	}
}

func TestTrekB14MarinerAttackingTapsAndPromotes(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	beckett := effectstest.B12Push(g, me.ID, "Beckett Mariner, Impetuous Ensign", trekB14MarinerTypeLine, trekB14MarinerOracle, 2, 1)
	effectstest.DeclareAttack(t, g, opp.ID, beckett)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterOn(g, beckett, cardsb.BeckettMarinerCounter); got != 1 {
		t.Errorf("promotion counters after attacking = %d, want 1", got)
	}
}

// --- Boarding Action ---

func TestTrekB14BoardingActionMakesAHastyKlingon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	id := effectstest.HandCardFull(me, "Boarding Action", "Sorcery", "{1}{R}", trekB14BoardingOracle, []string{"R"})
	effectstest.FloatMana(t, g, me, "{R}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	ks := effectstest.TokensNamed(g, "Klingon Warrior")
	if len(ks) != 1 {
		t.Fatalf("Klingon Warrior tokens = %d, want 1", len(ks))
	}
	k := effectstest.LayeredCard(t, g, ks[0].InstanceID)
	if k.CurrentPower() != 2 || k.CurrentToughness() != 1 || len(k.Colors) != 1 || k.Colors[0] != "R" ||
		!effectstest.HasEffectiveKeyword(t, g, k.InstanceID, "haste") {
		t.Errorf("Klingon Warrior = %+v, want a 2/1 red creature with haste", k)
	}
	if !me.Graveyard.Contains(id) {
		t.Error("cast from hand, Boarding Action should go to the graveyard")
	}
}

func TestTrekB14BoardingActionBeamsUpFromTheGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	card := effectstest.GraveyardCardOf(g, me, "Boarding Action", "Sorcery", "{1}{R}", trekB14BoardingOracle)
	effectstest.FloatMana(t, g, me, "{R}{C}{C}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: effects.TrekB14AltCostKeyBeamMeUp,
	}); err == nil {
		t.Fatal("Beam me up was cast without returning a creature")
	}
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", trekB14Bear, 2, 2)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: effects.TrekB14AltCostKeyBeamMeUp,
		AltCostIDs: []uuid.UUID{theirs},
	}); err == nil {
		t.Fatal("Beam me up returned an opponent's creature as its cost")
	}
	mine := effectstest.B12Creature(g, me.ID, "Mine", trekB14Bear, 2, 2)
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		Strict: true, FromZone: "graveyard", AlternativeCost: effects.TrekB14AltCostKeyBeamMeUp,
		AltCostIDs: []uuid.UUID{mine},
	}); err != nil {
		t.Fatalf("Beam me up: %v", err)
	}
	if !me.Hand.Contains(mine) {
		t.Fatal("the returned creature is not in its owner's hand")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Klingon Warrior", me.ID); n != 1 {
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
	effectstest.ToMainForCost(t, g)
	theirSpell = effectstest.HandCardFull(opp, "Their Instant", "Instant", "{U}", "", []string{"U"})
	if err := g.CastSpell(opp.ID, theirSpell, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent cast: %v", err)
	}
	bluff := effectstest.HandCardFull(me, "Brazen Bluff", "Instant", "{R}{R}", trekB14BluffOracle, []string{"R"})
	effectstest.FloatMana(t, g, me, "{R}{R}")
	if err := g.CastSpell(me.ID, bluff, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(theirSpell)}); err != nil {
		t.Fatalf("cast Brazen Bluff: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	return theirSpell
}

func TestTrekB14BluffTheyExileTheirSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	free := effectstest.HandCardFull(me, "Free Sorcery", "Sorcery", "{5}", "", nil)
	theirs := trekB14BluffBoard(t, g)
	effectstest.AnswerMayChoice(t, g, opp.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if z := g.FindCardZoneForEffect(theirs); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("their spell is in %v, want exile", z)
	}
	if effectstest.ResolutionCastOfferFor(g, me.ID) != nil {
		t.Error("a free cast was offered although they exiled the spell")
	}
	if !me.Hand.Contains(free) {
		t.Error("the hand card moved")
	}
}

func TestTrekB14BluffTheyRefuseAndYouCastANoncreatureSpellFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	free := effectstest.HandCardFull(me, "Free Sorcery", "Sorcery", "{5}", "", nil)
	creature := effectstest.HandCardFull(me, "Big Bear", trekB14Bear, "{4}{G}", "", []string{"G"})
	theirs := trekB14BluffBoard(t, g)
	effectstest.AnswerMayChoice(t, g, opp.ID, false)
	offer := effectstest.MustResolutionCastOffer(t, g, me.ID, game.MayCastKeywordHandCast, free)
	if effectstest.HasID(offer.DiscardOptions, creature) {
		t.Error("a creature card was offered as a noncreature spell")
	}
	if err := effectstest.CastThroughResolutionCast(t, g, me.ID, free, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("the free cast: %v", err)
	}
	effectstest.CastInsideTheResolution(t, g, free)
	if z := g.FindCardZoneForEffect(theirs); z == nil || z.Kind == game.ZoneExile {
		t.Errorf("their spell is in %v; refusing must leave it alone", z)
	}
}

func TestTrekB14BluffRefusesACreatureSpellOrYourOwn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	theirCreature := uuid.New()
	opp.Hand.PushTop(game.Card{InstanceID: theirCreature, Name: "Their Bear", TypeLine: trekB14Bear, ManaCost: "{1}{G}",
		Keywords: []string{"flash"}, Owner: opp.ID, Controller: opp.ID})
	if err := g.CastSpell(opp.ID, theirCreature, game.CastSpellParams{}); err != nil {
		t.Fatalf("opponent cast: %v", err)
	}
	mine := effectstest.HandCardFull(me, "My Instant", "Instant", "{U}", "", []string{"U"})
	if err := g.CastSpell(me.ID, mine, game.CastSpellParams{}); err != nil {
		t.Fatalf("my cast: %v", err)
	}
	for _, target := range []uuid.UUID{theirCreature, mine} {
		bluff := effectstest.HandCardFull(me, "Brazen Bluff", "Instant", "{R}{R}", trekB14BluffOracle, []string{"R"})
		effectstest.FloatMana(t, g, me, "{R}{R}")
		if err := g.CastSpell(me.ID, bluff, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(target)}); err == nil {
			t.Errorf("Brazen Bluff accepted %s as its target", target)
		}
	}
}

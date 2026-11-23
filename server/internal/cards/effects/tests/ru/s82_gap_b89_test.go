package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB89Twelfth = "9ecf19ba-754e-4bb3-b526-fe9df87a0f08"
	gapB89Kate    = "c1b9475e-a272-4820-b5f0-25f7a94ff834"
	gapB89Boe     = "acef7666-53b1-43c5-8107-236bea22a8b3"
	gapB89Jenny   = "fbddb6a4-f817-4f3d-bf75-63a98e71954b"
	gapB89Adric   = "af566158-23b2-4f91-8d59-a58342a9f576"
	gapB89Auton   = "94357494-4ccc-43d4-b2d4-351581cbf9b6"
	gapB89Danny   = "92cda790-7907-47d1-8e5a-d3364c1e881f"
	gapB89K9      = "413bc18b-1579-41b5-a778-3a29def52bb0"
	gapB89Martha  = "5b7ccdd2-40aa-4619-a147-0792dc02b2e5"
	gapB89Nyssa   = "07802bd6-29ca-40c5-beba-c2baf7a92555"
	gapB89Osgood  = "292aafef-8945-4827-90b2-1dd9764e517b"
)

// gapB89Names is the batch's built cards; every one must be Full with no
// caveats.
var gapB89Names = map[string]string{
	gapB89Jenny:   "Jenny, Generated Anomaly",
	gapB89K9:      "K-9, Mark I",
	gapB89Martha:  "Martha Jones",
	gapB89Danny:   "Danny Pink",
	gapB89Nyssa:   "Nyssa of Traken",
	gapB89Adric:   "Adric, Mathematical Genius",
	gapB89Auton:   "Auton Soldier",
	gapB89Kate:    "Kate Stewart",
	gapB89Twelfth: "The Twelfth Doctor",
	gapB89Osgood:  "Osgood, Operation Double",
}

func TestGapB89CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range gapB89Names {
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

func gapB89Clues(g *game.Game, owner uuid.UUID) (out []uuid.UUID) {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.HasSubtype("Clue") {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func TestGapB89JennyExploresOnCombatDamageToAPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	jenny := effectstest.B12Push(g, me.ID, "Jenny, Generated Anomaly", "Legendary Creature — Time Lord Soldier", gapB89Jenny, 2, 3)
	top := effectstest.PutOnTopForTest(me, "Forest", "Basic Land — Forest")
	effectstest.DealCombatDamageToPlayer(g, jenny, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(top) {
		t.Error("the revealed land did not go to the hand")
	}
	// A nonland gives Jenny a counter.
	effectstest.PutOnTopForTest(me, "Some Sorcery", "Sorcery")
	effectstest.DealCombatDamageToPlayer(g, jenny, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, jenny, game.CounterPlusOne); got != 1 {
		t.Errorf("Jenny has %d +1/+1 counters, want 1", got)
	}
}

func TestGapB89MarthaInvestigatesAndMakesTwoCreaturesUnblockableOnAClueSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	martha := effectstest.Msc17Play(t, g, me, game.Card{Name: "Martha Jones", TypeLine: "Legendary Creature — Human Cleric", OracleID: gapB89Martha, Power: 3, Toughness: 2})
	clues := gapB89Clues(g, me.ID)
	if len(clues) != 1 {
		t.Fatalf("clues after entering %d, want 1", len(clues))
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(clues[0]) })
	g.RunStateChecksForTest()
	effectstest.PickTriggerTarget(t, g, me.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, bear)&game.CantBeBlocked == 0 {
		t.Error("the other target can be blocked")
	}
	if effectstest.RestrictionsOf(t, g, martha)&game.CantBeBlocked == 0 {
		t.Error("Martha can be blocked")
	}
	// A nonClue sacrifice does nothing.
	other := effectstest.B12Creature(g, me.ID, "Goat", "Creature — Goat", 1, 1)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(other) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("a nonClue sacrifice triggered Martha")
	}
}

func TestGapB89K9GrantsWardWhileUntappedAndMakesALegendUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	k9 := effectstest.B12Push(g, me.ID, "K-9, Mark I", "Legendary Artifact Creature — Robot Dog", gapB89K9, 1, 1)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	effectstest.ApaMana(me, "C", "U")
	effectstest.ApaActivate(t, g, me, k9, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(legend)})
	if effectstest.RestrictionsOf(t, g, legend)&game.CantBeBlocked == 0 {
		t.Error("the legendary target can be blocked")
	}
	// Strict refusal: a nonlegendary target is illegal.
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "C", "U")
	g.WithWriteLock(func() { k9c, _ := g.LookupCardForEffect(k9); _ = k9c })
	if err := g.ActivateCatalogAbility(me.ID, k9, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err == nil {
		t.Error("K-9 targeted a nonlegendary creature")
	}
	_ = opp
}

func TestGapB89DannyPinkDrawsOnTheFirstCountersEachTurnOnAnyCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	effectstest.B12Push(g, me.ID, "Danny Pink", "Legendary Creature — Human Soldier Advisor", gapB89Danny, 4, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Goat", "Creature — Goat", 1, 1)
	foe := effectstest.B12Creature(g, g.Seats[1].ID, "Foe", "Creature — Bear", 2, 2)
	// Read each creature through the layers so the granted ability is in
	// place before the counters go on, as the layer pass does in play.
	for _, id := range []uuid.UUID{bear, other, foe} {
		effectstest.LayeredCard(t, g, id)
	}
	hand, foeHand := me.Hand.Size(), g.Seats[1].Hand.Size()
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d after the first counters, want %d", me.Hand.Size(), hand+1)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(bear, game.CounterPlusOne, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d after the second counters this turn, want %d", me.Hand.Size(), hand+1)
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(other, game.CounterPlusOne, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d after a different creature's first counters, want %d", me.Hand.Size(), hand+2)
	}
	// An opponent's creature is not granted the ability.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(foe, game.CounterPlusOne, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+2 || g.Seats[1].Hand.Size() != foeHand {
		t.Errorf("an opponent's creature drew: mine %d theirs %d", me.Hand.Size(), g.Seats[1].Hand.Size())
	}
}

func TestGapB89NyssaSacrificesArtifactsToTapCreaturesAndDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B", "C", "D")
	nyssa := effectstest.B12Push(g, me.ID, "Nyssa of Traken", "Legendary Creature — Human Scientist", gapB89Nyssa, 3, 4)
	a1 := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	a2 := effectstest.B12Permanent(g, me.ID, "Pebble", "Artifact")
	foe := effectstest.B12Creature(g, opp.ID, "Foe", "Creature — Bear", 2, 2)
	foe2 := effectstest.B12Creature(g, opp.ID, "Foe Two", "Creature — Bear", 2, 2)
	hand := me.Hand.Size()
	effectstest.StdB42Attack(t, g, opp.ID, nyssa)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID, a1, a2)
	if g.Battlefield.Contains(a1) || g.Battlefield.Contains(a2) {
		t.Fatal("the chosen artifacts were not sacrificed")
	}
	// Up to two targets: pick one only.
	effectstest.AnswerTriggerTargets(t, g, me.ID, foe)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AvB1Card(t, g, foe).Tapped {
		t.Error("the chosen creature is untapped")
	}
	if effectstest.AvB1Card(t, g, foe2).Tapped {
		t.Error("an unchosen creature was tapped")
	}
	if got := me.Hand.Size(); got != hand+2 {
		t.Errorf("hand %d, want %d (drew two for two artifacts)", got, hand+2)
	}
}

func TestGapB89NyssaSacrificingNothingDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "A", "B")
	nyssa := effectstest.B12Push(g, me.ID, "Nyssa of Traken", "Legendary Creature — Human Scientist", gapB89Nyssa, 3, 4)
	effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	hand := me.Hand.Size()
	effectstest.StdB42Attack(t, g, opp.ID, nyssa)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerOwnPermanents(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand {
		t.Errorf("drew with nothing sacrificed")
	}
}

func TestGapB89AdricCountersAnAbilityBySacrificingHimself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	k9 := effectstest.B12Push(g, me.ID, "K-9, Mark I", "Legendary Artifact Creature — Robot Dog", gapB89K9, 1, 1)
	adric := effectstest.B12Push(g, me.ID, "Adric, Mathematical Genius", "Legendary Creature — Human Artificer", gapB89Adric, 1, 1)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	effectstest.ApaMana(me, "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, k9, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(legend)}); err != nil {
		t.Fatal(err)
	}
	ability, ok := dmuB17AbilityOnStack(g)
	if !ok {
		t.Fatal("no ability on the stack")
	}
	effectstest.ApaMana(me, "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, adric, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ability), Strict: true}); err != nil {
		t.Fatalf("Ultimate Sacrifice: %v", err)
	}
	if g.Battlefield.Contains(adric) {
		t.Error("Adric was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, legend)&game.CantBeBlocked != 0 {
		t.Error("the countered ability still resolved")
	}
}

func TestGapB89AdricCopiesAnAbilityYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	k9 := effectstest.B12Push(g, me.ID, "K-9, Mark I", "Legendary Artifact Creature — Robot Dog", gapB89K9, 1, 1)
	adric := effectstest.B12Push(g, me.ID, "Adric, Mathematical Genius", "Legendary Creature — Human Artificer", gapB89Adric, 1, 1)
	legend := effectstest.B12Creature(g, me.ID, "Legend", "Legendary Creature — Human", 2, 2)
	effectstest.ApaMana(me, "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, k9, 0, game.ActivateAbilityParams{Targets: effectstest.CardTarget(legend)}); err != nil {
		t.Fatal(err)
	}
	ability, _ := dmuB17AbilityOnStack(g)
	effectstest.ApaMana(me, "C", "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, adric, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(ability), Strict: true}); err != nil {
		t.Fatalf("copy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// The copy's new targets are chosen as it is put on the stack: keep the
	// same legend.
	effectstest.PickTriggerTarget(t, g, me.ID, legend)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, legend)&game.CantBeBlocked == 0 {
		t.Error("the legendary target can be blocked")
	}
	if g.Stack.Size() != 0 {
		t.Errorf("stack still has %d items", g.Stack.Size())
	}
}

func TestGapB89KateStewartMakesSoldiersForTimeCountersAndPumpsAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	kate := effectstest.B12Push(g, me.ID, "Kate Stewart", "Legendary Creature — Human Scientist", gapB89Kate, 3, 3)
	rock := effectstest.B12Permanent(g, me.ID, "Rock", "Artifact")
	soldiers := func() int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Controller == me.ID && c.HasSubtype("Soldier") {
				n++
			}
		}
		return n
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(rock, game.CounterTime, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if soldiers() != 1 {
		t.Fatalf("soldiers %d after one batch of time counters, want 1", soldiers())
	}
	// Other counters do not trigger.
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(rock, game.CounterPlusOne, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if soldiers() != 1 {
		t.Errorf("a +1/+1 counter made a Soldier")
	}
	// An opponent's permanent does not.
	foeRock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(foeRock, game.CounterTime, 1) })
	effectstest.PassPriorityAroundTable(t, g)
	if soldiers() != 1 {
		t.Errorf("a time counter on an opponent's permanent made a Soldier")
	}
	// Attack: pay {8} for +2/+2 on attackers (two time counters among mine).
	effectstest.StdB42Attack(t, g, opp.ID, kate)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.ApaMana(me, "C", "C", "C", "C", "C", "C", "C", "C")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, kate); p != 5 {
		t.Errorf("Kate has power %d, want 5 (3 plus two time counters)", p)
	}
}

func TestGapB89OsgoodMakesANonlegendaryTokenCopyOnCastAndInvestigatesOffHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.AvB1Hand(me, game.Card{Name: "Osgood, Operation Double", TypeLine: "Legendary Creature — Human Alien Shapeshifter", OracleID: gapB89Osgood, Power: 2, Toughness: 2})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var token, real *game.Card
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Name != "Osgood, Operation Double" {
			continue
		}
		if c.IsToken() {
			token = c
		} else {
			real = c
		}
	}
	if real == nil || token == nil {
		t.Fatalf("battlefield has real=%v token=%v", real != nil, token != nil)
	}
	if token.IsLegendary() || !real.IsLegendary() {
		t.Errorf("token legendary=%v, real legendary=%v; want false and true", token.IsLegendary(), real.IsLegendary())
	}
	if len(gapB89Clues(g, me.ID)) != 0 {
		t.Error("a hand cast investigated")
	}
}

func TestGapB89AutonSoldierEntersAsANonlegendaryArtifactCopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	legend := effectstest.SeedCopyableCreature(g, me.ID, "Big Legend", "Legendary Creature — Human Soldier", 4, 4)
	id := effectstest.CastCatalogSpell(t, g, "Auton Soldier", "Artifact Creature — Alien Soldier", gapB89Auton, nil)
	effectstest.ResolveWithCopyChoice(t, g, legend)
	g.RunStateChecksForTest()
	c := effectstest.LayeredCard(t, g, id)
	if c.Name != "Big Legend" || c.HasSupertype("Legendary") || !c.IsArtifact() {
		t.Errorf("copy = %q legendary=%v artifact=%v, want a nonlegendary artifact Big Legend", c.Name, c.HasSupertype("Legendary"), c.IsArtifact())
	}
	if !g.Battlefield.Contains(legend) {
		t.Error("the original is gone")
	}
}

func TestGapB89TwelfthDoctorGrowsOnYourCopiesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	doc := effectstest.B12Push(g, me.ID, "The Twelfth Doctor", "Legendary Creature — Time Lord Doctor", gapB89Twelfth, 4, 4)
	spell := effectstest.CastCatalogSpell(t, g, "Lightning Bolt", "Instant", "", nil)
	g.WithWriteLock(func() { _ = g.CopySpellForEffect(spell, opp.ID, false, nil) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, doc, game.CounterPlusOne); got != 0 {
		t.Errorf("an opponent's copy gave %d counters", got)
	}
	g.WithWriteLock(func() { _ = g.CopySpellForEffect(spell, me.ID, false, nil) })
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, doc, game.CounterPlusOne); got != 1 {
		t.Errorf("your copy gave %d counters, want 1", got)
	}
}

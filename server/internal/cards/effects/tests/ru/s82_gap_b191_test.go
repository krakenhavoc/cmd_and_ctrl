package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b191_test.go — S82 GAP batch 191 (#5948). Each test plays the real
// catalog entry.

const (
	gapB191IntercessorOracle = "3ba19659-461f-482c-b882-b4fca7d5d1f8"
	gapB191SeverinaOracle    = "c00074ec-5b31-44d3-b69b-e2bf6f6a2477"
	gapB191CommanderOracle   = "53643dd7-183b-4062-b7a3-e7723d23bdb0"
	gapB191NeyamOracle       = "9466cfc1-a8fe-44a0-abd2-5d2dfe7110e0"
	gapB191ChaplainOracle    = "bd024849-3005-4509-ad83-65a9bcf956ac"
	gapB191HospitallerOracle = "06c39c1c-0e34-4ef0-bd06-23245cb9cc59"
	gapB191RepentiaOracle    = "24e9a1cb-b58b-491f-9d08-3f039a86851e"
	gapB191AberrantOracle    = "46b5f12e-9fe5-44f6-b895-ee8c33205521"
	gapB191BroodlordOracle   = "e3a73311-25f7-4bd7-af1d-ab7b83ad01f2"
	gapB191ClamavusOracle    = "d772b16b-31db-48aa-8bb6-da32f45eebe4"
	gapB191HaruspexOracle    = "2309436e-3442-46d7-8ffc-9dc4a2314ae9"
)

func TestGapB191CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB191IntercessorOracle: "Assault Intercessor",
		gapB191SeverinaOracle:    "Commissar Severina Raine",
		gapB191CommanderOracle:   "Company Commander",
		gapB191NeyamOracle:       "Neyam Shai Murad",
		gapB191ChaplainOracle:    "Primaris Chaplain",
		gapB191HospitallerOracle: "Sister Hospitaller",
		gapB191RepentiaOracle:    "Sister Repentia",
		gapB191AberrantOracle:    "Aberrant",
		gapB191BroodlordOracle:   "Broodlord",
		gapB191ClamavusOracle:    "Clamavus",
		gapB191HaruspexOracle:    "Haruspex",
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

// --- Assault Intercessor -------------------------------------------------------

func TestGapB191IntercessorHasFirstStrikeAndMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	id := effectstest.B12Push(g, g.Seats[0].ID, "Assault Intercessor", "Creature — Astartes Warrior", gapB191IntercessorOracle, 3, 2)
	for _, kw := range []string{"first strike", "menace"} {
		if !effectstest.HasKeywordOnBattlefield(t, g, id, kw) {
			t.Errorf("Assault Intercessor lacks %s", kw)
		}
	}
}

func TestGapB191IntercessorDrainsTheControllerOfADyingOpposingCreatureOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Assault Intercessor", "Creature — Astartes Warrior", gapB191IntercessorOracle, 3, 2)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	myLife, oppLife := me.Life, opp.Life
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife || opp.Life != oppLife {
		t.Fatalf("my own creature dying changed life: me %d→%d opp %d→%d", myLife, me.Life, oppLife, opp.Life)
	}
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 || me.Life != myLife {
		t.Errorf("after the opposing creature died: me %d→%d, opponent %d→%d, want only the opponent to lose 2", myLife, me.Life, oppLife, opp.Life)
	}
}

// --- Commissar Severina Raine --------------------------------------------------

func TestGapB191SeverinaDrainsPerOtherAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sev := effectstest.B12Push(g, me.ID, "Commissar Severina Raine", "Legendary Creature — Human Soldier", gapB191SeverinaOracle, 2, 2)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.B12Creature(g, me.ID, "Idle Bear", "Creature — Bear", 2, 2)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, sev, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	// Two other attackers: two life lost to the trigger (combat damage has not happened yet).
	if opp.Life != life-2 {
		t.Errorf("opponent life %d → %d, want 2 lost to the two other attackers", life, opp.Life)
	}
}

func TestGapB191SeverinaAloneDrainsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	sev := effectstest.B12Push(g, me.ID, "Commissar Severina Raine", "Legendary Creature — Human Soldier", gapB191SeverinaOracle, 2, 2)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, sev)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("opponent life %d → %d with no other attacker", life, opp.Life)
	}
}

func TestGapB191SeverinaSacrificesAnotherCreatureToGainAndDraw(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sev := effectstest.B12Push(g, me.ID, "Commissar Severina Raine", "Legendary Creature — Human Soldier", gapB191SeverinaOracle, 2, 2)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	effectstest.AdvanceToMain(t, g)
	// Severina herself is never a legal payment.
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sev, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{sev}}); err == nil {
		t.Fatal("Severina sacrificed herself to her own ability")
	}
	life, hand := me.Life, me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, sev, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card drawn", hand, me.Hand.Size())
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
}

func TestGapB191SeverinaCannotActivateWithoutAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sev := effectstest.B12Push(g, me.ID, "Commissar Severina Raine", "Legendary Creature — Human Soldier", gapB191SeverinaOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, sev, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no other creature to sacrifice")
	}
}

// --- Company Commander ---------------------------------------------------------

func TestGapB191CompanyCommanderMakesASoldierPerOpponent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	want := len(g.Seats) - 1
	effectstest.CastCatalogSpell(t, g, "Company Commander", "Creature — Human Soldier", gapB191CommanderOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	ids := effectstest.BattlefieldIDsNamed(g, "Soldier")
	if len(ids) != want {
		t.Fatalf("%d Soldier tokens, want %d (one per opponent)", len(ids), want)
	}
	tok, _ := g.LookupCardForEffect(ids[0])
	if tok.Power != 1 || tok.Toughness != 1 || tok.Controller != me.ID {
		t.Errorf("token %d/%d under %v, want a 1/1 under me", tok.Power, tok.Toughness, tok.Controller)
	}
}

func TestGapB191CompanyCommanderAttackGivesDeathtouchToYourCreaturesOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cc := effectstest.B12Push(g, me.ID, "Company Commander", "Creature — Human Soldier", gapB191CommanderOracle, 2, 4)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, cc)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{cc, mine} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "deathtouch") {
			t.Errorf("creature %v lacks deathtouch after the attack trigger", id)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "deathtouch") {
		t.Error("an opponent's creature gained deathtouch")
	}
}

// --- Primaris Chaplain ---------------------------------------------------------

func TestGapB191ChaplainBattleCriesAndGainsIndestructible(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ch := effectstest.B12Push(g, me.ID, "Primaris Chaplain", "Creature — Astartes Cleric", gapB191ChaplainOracle, 3, 3)
	pal := effectstest.B12Creature(g, me.ID, "Pal", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, ch, pal)
	effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, ch, "indestructible") {
		t.Error("the Chaplain did not gain indestructible when it attacked")
	}
	if p, _ := effectstest.EffectivePTOf(t, g, pal); p != 3 {
		t.Errorf("the other attacker has %d power, want 3 from battle cry", p)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, ch); p != 3 {
		t.Errorf("the Chaplain itself has %d power, want 3 (battle cry is for each OTHER attacker)", p)
	}
}

func TestGapB191ChaplainIsNotIndestructibleBeforeItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	ch := effectstest.B12Push(g, g.Seats[0].ID, "Primaris Chaplain", "Creature — Astartes Cleric", gapB191ChaplainOracle, 3, 3)
	if effectstest.HasEffectiveKeyword(t, g, ch, "indestructible") {
		t.Error("indestructible without having attacked")
	}
}

// --- Sister Hospitaller --------------------------------------------------------

func TestGapB191HospitallerReturnsACreatureAndGainsItsManaValue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dead := effectstest.B38Graveyard(g, me.ID, "Fallen Giant", "Creature — Giant", "{3}{R}", 4, 4)
	land := effectstest.B38Graveyard(g, me.ID, "Fallen Land", "Land", "", 0, 0)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Sister Hospitaller", "Creature — Human Cleric", gapB191HospitallerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(land)); err == nil {
		t.Fatal("a land card was accepted as the creature-card target")
	}
	effectstest.PickCard(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(dead) {
		t.Fatal("the creature card did not return to the battlefield")
	}
	if me.Life != life+4 {
		t.Errorf("life %d → %d, want +4 (its mana value)", life, me.Life)
	}
}

func TestGapB191HospitallerWithNoCreatureCardDoesNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Sister Hospitaller", "Creature — Human Cleric", gapB191HospitallerOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life moved to %d with nothing to return", me.Life)
	}
}

// --- Sister Repentia -----------------------------------------------------------

func TestGapB191RepentiaGainsAndDrawsWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rep := effectstest.B12Push(g, me.ID, "Sister Repentia", "Creature — Human Warrior", gapB191RepentiaOracle, 5, 1)
	life, hand := me.Life, me.Hand.Size()
	effectstest.KillCreature(t, g, me.ID, rep)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d → %d, want +2", life, me.Life)
	}
	if me.Hand.Size() != hand+2 {
		t.Errorf("hand %d → %d, want two cards", hand, me.Hand.Size())
	}
}

func TestGapB191RepentiaIsCastForItsMiracleCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[1]
	rep := effectstest.DrawMiracleInOwnDrawStep(t, g, 1, "Sister Repentia", "Creature — Human Warrior", "{3}{W}{B}", gapB191RepentiaOracle)
	effectstest.RevealAndResolve(t, g, me)
	if !effectstest.HasMiracleGrant(me, rep) {
		t.Fatal("the resolved miracle trigger granted no cast")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{W}{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, rep, game.CastSpellParams{Strict: true, AlternativeCost: game.AltCostKeyMiracle}); err != nil {
		t.Fatalf("the miracle cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(rep) {
		t.Error("Sister Repentia did not resolve onto the battlefield")
	}
}

// --- Aberrant ------------------------------------------------------------------

func TestGapB191AberrantEntersWithXCountersAndTramples(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	ab := effectstest.CastXSpell(t, g, "Aberrant", "Creature — Tyranid Mutant", gapB191AberrantOracle, "{X}{1}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, ab, game.CounterPlusOne); got != 3 {
		t.Fatalf("%d +1/+1 counters, want 3", got)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, ab, "trample") {
		t.Error("Aberrant lacks trample")
	}
	if me.Hand.Size() != hand {
		t.Errorf("X = 3 drew a card (hand %d → %d)", hand, me.Hand.Size())
	}
}

func TestGapB191AberrantDrawsAtXFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	hand := me.Hand.Size()
	ab := effectstest.CastXSpell(t, g, "Aberrant", "Creature — Tyranid Mutant", gapB191AberrantOracle, "{X}{1}{G}", 5, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, ab, game.CounterPlusOne); got != 5 {
		t.Fatalf("%d +1/+1 counters, want 5", got)
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d, want one card from ravenous at X = 5", hand, me.Hand.Size())
	}
}

func TestGapB191AberrantDestroysAnArtifactOrEnchantmentOfTheDamagedPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	third := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	ab := effectstest.B12Push(g, me.ID, "Aberrant", "Creature — Tyranid Mutant", gapB191AberrantOracle, 3, 3)
	rock := effectstest.B12Permanent(g, opp.ID, "Their Rock", "Artifact")
	other := effectstest.B12Permanent(g, third.ID, "Third Rock", "Artifact")
	land := effectstest.B12Permanent(g, opp.ID, "Their Land", "Land")
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, ab)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(other)); err == nil {
		t.Fatal("another player's artifact was accepted")
	}
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(land)); err == nil {
		t.Fatal("a land was accepted")
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("defender life %d, want %d", opp.Life, life-3)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the damaged player's artifact was not destroyed")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("a third player's artifact was destroyed")
	}
}

// --- Broodlord -----------------------------------------------------------------

func TestGapB191BroodlordDistributesXCountersAmongOtherCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	elk := effectstest.B12Creature(g, me.ID, "Elk", "Creature — Elk", 2, 2)
	bl := effectstest.CastXSpell(t, g, "Broodlord", "Creature — Tyranid", gapB191BroodlordOracle, "{X}{3}{G}", 3, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bl, game.CounterPlusOne); got != 3 {
		t.Fatalf("Broodlord has %d +1/+1 counters, want 3 (ravenous)", got)
	}
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no targets")
	}
	// Broodlord itself is not a legal target ("other"), and the split must total X.
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, effectstest.CardRefs(bl, bear), map[uuid.UUID]int{bl: 1, bear: 2}); err == nil {
		t.Fatal("Broodlord was accepted as one of its own 'other' targets")
	}
	if err := g.ResolvePickTargetsDivided(p.ID, me.ID, effectstest.CardRefs(bear, elk), map[uuid.UUID]int{bear: 1, elk: 1}); err == nil {
		t.Fatal("a split of 2 was accepted for X = 3")
	}
	effectstest.B17PickCardsDivided(t, g, me.ID, map[uuid.UUID]int{bear: 2, elk: 1}, bear, elk)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("bear has %d counters, want 2", got)
	}
	if got := effectstest.CounterCount(g, elk, game.CounterPlusOne); got != 1 {
		t.Errorf("elk has %d counters, want 1", got)
	}
}

func TestGapB191BroodlordMayChooseNoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bl := effectstest.CastXSpell(t, g, "Broodlord", "Creature — Tyranid", gapB191BroodlordOracle, "{X}{3}{G}", 2, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, bl, game.CounterPlusOne); got != 2 {
		t.Fatalf("Broodlord has %d counters, want its own 2 and nothing more", got)
	}
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("declining every target: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if got := effectstest.CounterCount(g, bl, game.CounterPlusOne); got != 2 {
		t.Errorf("Broodlord has %d counters after distributing among nothing, want 2", got)
	}
}

// --- Clamavus ------------------------------------------------------------------

func TestGapB191ClamavusAddsOneForEachOfTheCreaturesOwnCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cl := effectstest.B12Push(g, me.ID, "Clamavus", "Creature — Human Tyranid Artificer", gapB191ClamavusOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	bare := effectstest.B12Creature(g, me.ID, "Bare", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.AddCountersForTest(g, bear, game.CounterPlusOne, 2)
	effectstest.AddCountersForTest(g, theirs, game.CounterPlusOne, 2)
	effectstest.AddCountersForTest(g, cl, game.CounterPlusOne, 1)
	for _, c := range []struct {
		name  string
		id    uuid.UUID
		power int
	}{
		{"a creature with two counters", bear, 6}, // 2 + two counters + two from Clamavus
		{"a creature with no counters", bare, 2},  // untouched
		{"an opponent's creature", theirs, 4},     // 2 + two counters, no pump
		{"Clamavus with one counter", cl, 5},      // 3 + one counter + one from itself
	} {
		if p := effectstest.EffectivePowerWithCounters(t, g, c.id); p != c.power {
			t.Errorf("%s has %d power, want %d", c.name, p, c.power)
		}
	}
}

// --- Haruspex ------------------------------------------------------------------

func TestGapB191HaruspexGrowsWhenAnotherCreatureDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	h := effectstest.B12Push(g, me.ID, "Haruspex", "Creature — Tyranid", gapB191HaruspexOracle, 2, 2)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, h, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters after two other creatures died, want 2", got)
	}
}

func TestGapB191HaruspexTurnsCountersIntoManaOfOneColor(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	h := effectstest.B12Push(g, me.ID, "Haruspex", "Creature — Tyranid", gapB191HaruspexOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.AddCountersForTest(g, h, game.CounterPlusOne, 3)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, h).SummonedThisTurn = false })
	effectstest.ActivateManaFor(t, g, me.ID, h, 0, game.ManaAbilityParams{CounterCounts: []int{2}})
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceMana && c.Chooser == me.ID {
			if err := g.ResolveManaChoice(c.ID, me.ID, "G"); err != nil {
				t.Fatalf("ResolveManaChoice: %v", err)
			}
			break
		}
	}
	if me.ManaPool.Total() != 2 {
		t.Fatalf("pool %d, want 2 mana for 2 counters removed", me.ManaPool.Total())
	}
	for _, u := range me.ManaPool.Units() {
		if u.Color != "G" {
			t.Errorf("mana of colour %q, want both of the one colour chosen (G)", u.Color)
		}
	}
	if got := effectstest.CounterCount(g, h, game.CounterPlusOne); got != 1 {
		t.Errorf("%d counters left, want 1", got)
	}
	// It tapped: a second activation is refused.
	if err := g.ActivateManaAbility(me.ID, h, 0, game.ManaAbilityParams{CounterCounts: []int{1}}); err == nil {
		t.Error("the tapped Haruspex made mana again")
	}
}

// --- Neyam Shai Murad ----------------------------------------------------------

func gapB191NeyamAttack(t *testing.T) (g *game.Game, me, opp *game.Player, neyam uuid.UUID) {
	t.Helper()
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[g.Turn.ActiveSeat]
	opp = g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	neyam = effectstest.B12Push(g, me.ID, "Neyam Shai Murad", "Legendary Creature — Human Rogue", gapB191NeyamOracle, 3, 3)
	return
}

func TestGapB191NeyamSwapsAPermanentCardForOneOfYoursChosenByThem(t *testing.T) {
	g, me, opp, neyam := gapB191NeyamAttack(t)
	theirs := effectstest.B38Graveyard(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)
	theirSpell := effectstest.B38Graveyard(g, opp.ID, "Their Bolt", "Instant", "{R}", 0, 0)
	cheap := effectstest.B38Graveyard(g, me.ID, "My Cheap Elf", "Creature — Elf", "{G}", 1, 1)
	pricey := effectstest.B38Graveyard(g, me.ID, "My Fine Dragon", "Creature — Dragon", "{4}{R}{R}", 5, 5)
	mySpell := effectstest.B38Graveyard(g, me.ID, "My Shock", "Instant", "{R}", 0, 0)
	effectstest.DeclareAttack(t, g, opp.ID, neyam)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.B04WaitForPick(t, g, me.ID)
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(theirSpell)); err == nil {
		t.Fatal("an instant card was accepted as the permanent-card target")
	}
	if err := g.ResolvePickTargets(effectstest.LatestPickTarget(g, me.ID).ID, me.ID, effectstest.CardRefs(cheap)); err == nil {
		t.Fatal("a card in MY graveyard was accepted; the target is in that player's graveyard")
	}
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Hand.Contains(theirs) {
		t.Fatal("the targeted card did not return to its owner's hand")
	}
	// The damaged player chooses among MY permanent cards (never my instant).
	prompt := effectstest.LatestChooseCards(g, opp.ID)
	if prompt == nil {
		t.Fatal("the damaged player was not asked to choose a permanent card in my graveyard")
	}
	for _, c := range prompt.ChooseCards {
		if c == mySpell {
			t.Error("a non-permanent card was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, cheap)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(cheap) {
		t.Fatal("the card they chose did not enter the battlefield")
	}
	if c, _ := g.LookupCardForEffect(cheap); c.Controller != me.ID {
		t.Errorf("the chosen card entered under %v, want my control", c.Controller)
	}
	if g.Battlefield.Contains(pricey) {
		t.Error("the card they did not choose entered too")
	}
}

func TestGapB191NeyamDecliningDoesNothing(t *testing.T) {
	g, me, opp, neyam := gapB191NeyamAttack(t)
	theirs := effectstest.B38Graveyard(g, opp.ID, "Their Bear", "Creature — Bear", "{1}{G}", 2, 2)
	mine := effectstest.B38Graveyard(g, me.ID, "My Elf", "Creature — Elf", "{G}", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, neyam)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(theirs) || g.Battlefield.Contains(mine) {
		t.Error("declining 'you may' still moved cards")
	}
}

func TestGapB191NeyamNeedsATargetInTheirGraveyard(t *testing.T) {
	g, me, opp, neyam := gapB191NeyamAttack(t)
	mine := effectstest.B38Graveyard(g, me.ID, "My Elf", "Creature — Elf", "{G}", 1, 1)
	effectstest.B38Graveyard(g, opp.ID, "Their Bolt", "Instant", "{R}", 0, 0)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, neyam)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("defender life %d, want %d", opp.Life, life-3)
	}
	if g.Battlefield.Contains(mine) || effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("the ability did something with no permanent card in their graveyard")
	}
}

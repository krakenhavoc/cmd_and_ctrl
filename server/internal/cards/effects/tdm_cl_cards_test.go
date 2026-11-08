package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tdm_cl_cards_test.go — Tarkir: Dragonstorm, cards from C to L (#4463).
// Each test plays the real catalog entry and asserts what its text says.

const (
	tdmChampionOfDusanOracle    = "0a427652-9745-4802-8c8f-5ae29a29b496"
	tdmCoriMountainStalwartOrcl = "322a2d97-cf71-4bfc-b52d-f642f2d0dcbd"
	tdmDalkovanPackbeastsOracle = "cdd1a043-bd47-4197-be26-e2f7d3f1d209"
	tdmDevotedDuelistOracle     = "f7620c46-b3d8-4694-ac7c-a43933c86c22"
	tdmDragonbackLancerOracle   = "7875c063-19c3-4e6e-81d8-1a378a4e16e5"
	tdmDusyutEarthcarverOracle  = "387b12b1-d8f5-487b-972f-5303b23eed5a"
	tdmFortressKinGuardOracle   = "fa3ed0f6-e44e-433d-80d3-824de5e6135a"
	tdmIceridgeSerpentOracle    = "53f07529-c212-43e4-b5b6-086efeadda39"
	tdmInspiritedVanguardOracle = "db756a16-3db6-4bde-bbc2-ecee3ee9d9d3"
	tdmJeskaiDevoteeOracle      = "6dcdd313-6df7-4d44-a4cb-0346a4fdb4e1"
	tdmJeskaiMonumentOracle     = "0b2b7d04-73f3-43f6-86e0-fa00cbbc6469"
	tdmKinTreeNurturerOracle    = "ad28a1f9-8893-4693-9464-c32ed760f6b1"
)

func TestTdmCLEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tdmChampionOfDusanOracle:    "Champion of Dusan",
		tdmCoriMountainStalwartOrcl: "Cori Mountain Stalwart",
		tdmDalkovanPackbeastsOracle: "Dalkovan Packbeasts",
		tdmDevotedDuelistOracle:     "Devoted Duelist",
		tdmDragonbackLancerOracle:   "Dragonback Lancer",
		tdmDusyutEarthcarverOracle:  "Dusyut Earthcarver",
		tdmFortressKinGuardOracle:   "Fortress Kin-Guard",
		tdmIceridgeSerpentOracle:    "Iceridge Serpent",
		tdmInspiritedVanguardOracle: "Inspirited Vanguard",
		tdmJeskaiDevoteeOracle:      "Jeskai Devotee",
		tdmJeskaiMonumentOracle:     "Jeskai Monument",
		tdmKinTreeNurturerOracle:    "Kin-Tree Nurturer",
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

// --- fixtures ------------------------------------------------------------------

// tdmCLLives is every seat's life total, by player.
func tdmCLLives(g *game.Game) map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		out[p.ID] = p.Life
	}
	return out
}

// tdmCLMobilizes attacks with `attacker` and checks it made `n` tapped and
// attacking 1/1 Warriors that are sacrificed at the next end step.
func tdmCLMobilizes(t *testing.T, g *game.Game, attacker uuid.UUID, n int) {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	declareAttack(t, g, opp.ID, attacker)
	passPriorityAroundTable(t, g)
	tdmPlaceWarriors(t, g, me.ID)
	warriors := socB8Named(g, me.ID, "Warrior")
	if len(warriors) != n {
		t.Fatalf("mobilize %d made %d Warriors", n, len(warriors))
	}
	for _, w := range warriors {
		if !w.Tapped || w.AttackingTarget == uuid.Nil || w.Power != 1 || w.Toughness != 1 ||
			len(w.Colors) != 1 || w.Colors[0] != "R" {
			t.Errorf("Warrior %d/%d %v tapped=%v attacking=%v, want a tapped and attacking 1/1 red",
				w.Power, w.Toughness, w.Colors, w.Tapped, w.AttackingTarget != uuid.Nil)
		}
	}
	advanceToEndStepOf(t, g, 0)
	passPriorityAroundTable(t, g)
	if left := len(socB8Named(g, me.ID, "Warrior")); left != 0 {
		t.Errorf("%d Warriors survived the end step", left)
	}
}

// tdmCLEnduresOnEntry casts a creature whose entry trigger endures n, and
// answers with `pick` (0 the Spirit, 1 the counters).
func tdmCLEnduresOnEntry(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness, n, pick int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := tdmSTCast(t, g, name, typeLine, oracle, power, toughness)
	passPriorityAroundTable(t, g)
	if latestOptionPickFor(g, me.ID) == nil {
		t.Fatalf("%s: no endure choice on entry", name)
	}
	answerOptionPick(t, g, me.ID, pick)
	spirits := socB8Named(g, me.ID, "Spirit")
	if pick == 1 {
		if got := countersOn(g, id, game.CounterPlusOne); got != n {
			t.Errorf("%s: %d +1/+1 counters, want %d", name, got, n)
		}
		if len(spirits) != 0 {
			t.Errorf("%s: %d Spirits after choosing the counters", name, len(spirits))
		}
		return id
	}
	if len(spirits) != 1 || spirits[0].Power != n || spirits[0].Toughness != n ||
		len(spirits[0].Colors) != 1 || spirits[0].Colors[0] != "W" {
		t.Fatalf("%s: Spirits %+v, want one %d/%d white Spirit", name, spirits, n, n)
	}
	if got := countersOn(g, id, game.CounterPlusOne); got != 0 {
		t.Errorf("%s: %d +1/+1 counters after choosing the Spirit", name, got)
	}
	return id
}

// --- Champion of Dusan ---------------------------------------------------------

func TestTdmChampionOfDusanHasTrampleAndRenewsACounterAndTrample(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	champion := b12Push(g, me.ID, "Champion of Dusan", "Creature — Human Warrior", tdmChampionOfDusanOracle, 4, 2)
	if !hasEffectiveKeyword(t, g, champion, "trample") {
		t.Error("Champion of Dusan has no trample")
	}
	bear := seedBear(g, me.ID)
	if hasEffectiveKeyword(t, g, bear, "trample") {
		t.Fatal("the Bear starts with trample")
	}
	id, _ := activateFromGraveyard(t, g, "Champion of Dusan", "Creature — Human Warrior",
		tdmChampionOfDusanOracle, 4, 2, "{G}{G}", game.ActivateAbilityParams{Targets: cardRefs(bear)})
	if !g.Exile.Contains(id) {
		t.Fatal("renew's cost did not exile the Champion")
	}
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if got := countersOn(g, bear, game.CounterTrample); got != 1 {
		t.Errorf("%d trample counters, want 1", got)
	}
	if !hasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the trample counter did not give trample")
	}
}

func TestTdmChampionOfDusanRenewCostsOneGFromTheGraveyardOnly(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: tdmChampionOfDusanOracle})
	if len(abilities) != 1 || abilities[0].Cost.Mana != "{1}{G}" ||
		!game.AbilityFunctionsFromZone(abilities[0], game.ZoneGraveyard) ||
		game.AbilityFunctionsFromZone(abilities[0], game.ZoneBattlefield) {
		t.Fatalf("abilities %+v, want one {1}{G} ability that works only from the graveyard", abilities)
	}
}

// --- Cori Mountain Stalwart ----------------------------------------------------

func TestTdmCoriMountainStalwartDrainsTwoOnTheSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b12Push(g, me.ID, "Cori Mountain Stalwart", "Creature — Human Monk", tdmCoriMountainStalwartOrcl, 3, 3)
	toMainForCost(t, g)
	before := tdmCLLives(g)
	tdmMPCastFreeSpell(t, g, me, "First")
	for _, p := range g.Seats {
		if p.Life != before[p.ID] {
			t.Fatalf("life moved on the first spell: %d -> %d", before[p.ID], p.Life)
		}
	}
	tdmMPCastFreeSpell(t, g, me, "Second")
	for _, p := range g.Seats {
		want := before[p.ID] - 2
		if p.ID == me.ID {
			want = before[p.ID] + 2
		}
		if p.Life != want {
			t.Errorf("seat %s life %d -> %d, want %d", p.ID, before[p.ID], p.Life, want)
		}
	}
	after := tdmCLLives(g)
	tdmMPCastFreeSpell(t, g, me, "Third")
	for _, p := range g.Seats {
		if p.Life != after[p.ID] {
			t.Errorf("life moved on the third spell: %d -> %d", after[p.ID], p.Life)
		}
	}
}

// --- Dalkovan Packbeasts -------------------------------------------------------

func TestTdmDalkovanPackbeastsHasVigilanceAndMobilizesThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ox := b12Push(g, me.ID, "Dalkovan Packbeasts", "Creature — Ox", tdmDalkovanPackbeastsOracle, 0, 4)
	if !hasEffectiveKeyword(t, g, ox, "vigilance") {
		t.Error("Dalkovan Packbeasts has no vigilance")
	}
	tdmCLMobilizes(t, g, ox, 3)
}

// --- Devoted Duelist -----------------------------------------------------------

func TestTdmDevotedDuelistPingsEachOpponentOnTheSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	duelist := b12Push(g, me.ID, "Devoted Duelist", "Creature — Goblin Monk", tdmDevotedDuelistOracle, 2, 1)
	if !hasEffectiveKeyword(t, g, duelist, "haste") {
		t.Error("Devoted Duelist has no haste")
	}
	toMainForCost(t, g)
	before := tdmCLLives(g)
	tdmMPCastFreeSpell(t, g, me, "First")
	tdmMPCastFreeSpell(t, g, me, "Second")
	for _, p := range g.Seats {
		want := before[p.ID] - 1
		if p.ID == me.ID {
			want = before[p.ID]
		}
		if p.Life != want {
			t.Errorf("seat %s life %d -> %d, want %d", p.ID, before[p.ID], p.Life, want)
		}
	}
}

// --- Dragonback Lancer ---------------------------------------------------------

func TestTdmDragonbackLancerFliesAndMobilizesOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	lancer := b12Push(g, me.ID, "Dragonback Lancer", "Creature — Human Soldier", tdmDragonbackLancerOracle, 3, 3)
	if !hasEffectiveKeyword(t, g, lancer, "flying") {
		t.Error("Dragonback Lancer has no flying")
	}
	tdmCLMobilizes(t, g, lancer, 1)
}

// --- Dusyut Earthcarver --------------------------------------------------------

func TestTdmDusyutEarthcarverEnduresThree(t *testing.T) {
	g := newCatalogGame(t)
	carver := tdmCLEnduresOnEntry(t, g, "Dusyut Earthcarver", "Creature — Elephant Druid", tdmDusyutEarthcarverOracle, 4, 4, 3, 1)
	if !hasEffectiveKeyword(t, g, carver, "reach") {
		t.Error("Dusyut Earthcarver has no reach")
	}
}

func TestTdmDusyutEarthcarverCanMakeAThreeThreeSpirit(t *testing.T) {
	g := newCatalogGame(t)
	tdmCLEnduresOnEntry(t, g, "Dusyut Earthcarver", "Creature — Elephant Druid", tdmDusyutEarthcarverOracle, 4, 4, 3, 0)
}

// --- Fortress Kin-Guard --------------------------------------------------------

func TestTdmFortressKinGuardEnduresOne(t *testing.T) {
	g := newCatalogGame(t)
	tdmCLEnduresOnEntry(t, g, "Fortress Kin-Guard", "Creature — Dog Soldier", tdmFortressKinGuardOracle, 1, 2, 1, 1)
	g2 := newCatalogGame(t)
	tdmCLEnduresOnEntry(t, g2, "Fortress Kin-Guard", "Creature — Dog Soldier", tdmFortressKinGuardOracle, 1, 2, 1, 0)
}

// --- Iceridge Serpent ----------------------------------------------------------

func TestTdmIceridgeSerpentBouncesAnOpponentsCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	serpent := castHoldCreature(t, g, "Iceridge Serpent", "Creature — Serpent", tdmIceridgeSerpentOracle, 3, 3)
	b04WaitForPick(t, g, me.ID)
	p := latestPickTarget(g, me.ID)
	for _, bad := range []uuid.UUID{mine, serpent} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was accepted as a target", bad)
		}
	}
	answerPickTarget(t, g, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the opponent's creature was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(serpent) {
		t.Error("your creature or the Serpent left the battlefield")
	}
}

// --- Inspirited Vanguard -------------------------------------------------------

func TestTdmInspiritedVanguardEnduresTwoOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	tdmCLEnduresOnEntry(t, g, "Inspirited Vanguard", "Creature — Human Soldier", tdmInspiritedVanguardOracle, 3, 2, 2, 1)
	g2 := newCatalogGame(t)
	tdmCLEnduresOnEntry(t, g2, "Inspirited Vanguard", "Creature — Human Soldier", tdmInspiritedVanguardOracle, 3, 2, 2, 0)
}

func TestTdmInspiritedVanguardEnduresTwoWhenItAttacks(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vanguard := b12Push(g, me.ID, "Inspirited Vanguard", "Creature — Human Soldier", tdmInspiritedVanguardOracle, 3, 2)
	declareAttack(t, g, opp.ID, vanguard)
	passPriorityAroundTable(t, g)
	if latestOptionPickFor(g, me.ID) == nil {
		t.Fatal("no endure choice when the Vanguard attacked")
	}
	answerOptionPick(t, g, me.ID, 1)
	if got := countersOn(g, vanguard, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	if n := len(socB8Named(g, me.ID, "Spirit")); n != 0 {
		t.Errorf("%d Spirits after choosing the counters", n)
	}
}

// --- Jeskai Devotee ------------------------------------------------------------

func TestTdmJeskaiDevoteeGrowsOnTheSecondSpell(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	devotee := b12Push(g, me.ID, "Jeskai Devotee", "Creature — Orc Monk", tdmJeskaiDevoteeOracle, 2, 2)
	toMainForCost(t, g)
	tdmMPCastFreeSpell(t, g, me, "First")
	if p := effectivePower(t, g, devotee); p != 2 {
		t.Fatalf("power %d after the first spell, want 2", p)
	}
	tdmMPCastFreeSpell(t, g, me, "Second")
	if p, tough := effectivePower(t, g, devotee), effectiveToughness(t, g, devotee); p != 3 || tough != 3 {
		t.Errorf("the Devotee is %d/%d after the second spell, want 3/3", p, tough)
	}
}

func TestTdmJeskaiDevoteeFiltersOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	devotee := b12Push(g, me.ID, "Jeskai Devotee", "Creature — Orc Monk", tdmJeskaiDevoteeOracle, 2, 2)
	advanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	fillPoolColored(me, "G", 2)
	if err := g.ActivateManaAbility(me.ID, devotee, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err == nil {
		t.Fatal("the Devotee made {B}; it adds {U}, {R}, or {W}")
	}
	if err := g.ActivateManaAbility(me.ID, devotee, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("{1}: add {U}: %v", err)
	}
	if got := mcPoolCount(me, "U"); got != 1 {
		t.Errorf("blue in the pool %d, want 1", got)
	}
	if err := g.ActivateManaAbility(me.ID, devotee, 0, game.ManaAbilityParams{Colors: []string{"W"}}); err == nil {
		t.Error("a second activation this turn was allowed")
	}
}

// --- Jeskai Monument -----------------------------------------------------------

func TestTdmJeskaiMonumentFindsABasicIslandMountainOrPlains(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedSearchLibrary(me,
		searchTestLand("Swamp", "Basic Land — Swamp"),
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Volcanic Island", "Land — Island Mountain"),
		searchTestLand("Island", "Basic Land — Island"),
	)
	castCatalogSpell(t, g, "Jeskai Monument", "Artifact", tdmJeskaiMonumentOracle, nil)
	passPriorityAroundTable(t, g)
	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the Monument's entry asked for no search")
	}
	for _, bad := range []string{"Swamp", "Forest", "Volcanic Island"} {
		if searchOptionNamed(g, c, bad) != uuid.Nil {
			t.Errorf("%s was offered; only a basic Island, Mountain, or Plains may be found", bad)
		}
	}
	answerSearchNamed(t, g, me.ID, "Island")
	if _, ok := handCardNamed(me, "Island"); !ok {
		t.Error("the Island is not in hand")
	}
}

func TestTdmJeskaiMonumentMakesTwoFlyingBirds(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monument := pushCatalogPermanent(g, me.ID, "Jeskai Monument", "Artifact", tdmJeskaiMonumentOracle, false)
	advanceToMain(t, g)
	fillPoolColored(me, "U", 2)
	fillPoolColored(me, "R", 1)
	fillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, monument, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(monument) {
		t.Fatal("the Monument was not sacrificed as the cost")
	}
	passPriorityAroundTable(t, g)
	birds := socB8Named(g, me.ID, "Bird")
	if len(birds) != 2 {
		t.Fatalf("%d Birds, want 2", len(birds))
	}
	for _, b := range birds {
		if b.Power != 1 || b.Toughness != 1 || len(b.Colors) != 1 || b.Colors[0] != "W" {
			t.Errorf("Bird %d/%d %v, want a 1/1 white Bird", b.Power, b.Toughness, b.Colors)
		}
		if !hasEffectiveKeyword(t, g, b.InstanceID, "flying") {
			t.Error("a Bird has no flying")
		}
	}
}

func TestTdmJeskaiMonumentActivatesOnlyAsASorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monument := pushCatalogPermanent(g, me.ID, "Jeskai Monument", "Artifact", tdmJeskaiMonumentOracle, false)
	advanceTo(t, g, game.StepBeginCombat)
	fillPoolColored(me, "U", 2)
	fillPoolColored(me, "R", 1)
	fillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, monument, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the Monument was activated outside a main phase")
	}
}

// --- Kin-Tree Nurturer ---------------------------------------------------------

func TestTdmKinTreeNurturerHasLifelinkAndEnduresOne(t *testing.T) {
	g := newCatalogGame(t)
	nurturer := tdmCLEnduresOnEntry(t, g, "Kin-Tree Nurturer", "Creature — Human Druid", tdmKinTreeNurturerOracle, 2, 1, 1, 1)
	if !hasEffectiveKeyword(t, g, nurturer, "lifelink") {
		t.Error("Kin-Tree Nurturer has no lifelink")
	}
	g2 := newCatalogGame(t)
	tdmCLEnduresOnEntry(t, g2, "Kin-Tree Nurturer", "Creature — Human Druid", tdmKinTreeNurturerOracle, 2, 1, 1, 0)
	me := g2.Seats[g2.Turn.ActiveSeat]
	if spirits := socB8Named(g2, me.ID, "Spirit"); len(spirits) == 1 && hasEffectiveKeyword(t, g2, spirits[0].InstanceID, "lifelink") {
		t.Error("the Spirit has lifelink; only the Nurturer does")
	}
}

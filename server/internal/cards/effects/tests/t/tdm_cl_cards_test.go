package testst

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestTdmCLEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		effectstest.TdmChampionOfDusanOracle:    "Champion of Dusan",
		effectstest.TdmCoriMountainStalwartOrcl: "Cori Mountain Stalwart",
		effectstest.TdmDalkovanPackbeastsOracle: "Dalkovan Packbeasts",
		effectstest.TdmDevotedDuelistOracle:     "Devoted Duelist",
		effectstest.TdmDragonbackLancerOracle:   "Dragonback Lancer",
		effectstest.TdmDusyutEarthcarverOracle:  "Dusyut Earthcarver",
		effectstest.TdmFortressKinGuardOracle:   "Fortress Kin-Guard",
		effectstest.TdmIceridgeSerpentOracle:    "Iceridge Serpent",
		effectstest.TdmInspiritedVanguardOracle: "Inspirited Vanguard",
		effectstest.TdmJeskaiDevoteeOracle:      "Jeskai Devotee",
		effectstest.TdmJeskaiMonumentOracle:     "Jeskai Monument",
		effectstest.TdmKinTreeNurturerOracle:    "Kin-Tree Nurturer",
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
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.TdmPlaceWarriors(t, g, me.ID)
	warriors := effectstest.SocB8Named(g, me.ID, "Warrior")
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
	effectstest.AdvanceToEndStepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if left := len(effectstest.SocB8Named(g, me.ID, "Warrior")); left != 0 {
		t.Errorf("%d Warriors survived the end step", left)
	}
}

// tdmCLEnduresOnEntry casts a creature whose entry trigger endures n, and
// answers with `pick` (0 the Spirit, 1 the counters).
func tdmCLEnduresOnEntry(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness, n, pick int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := tdmSTCast(t, g, name, typeLine, oracle, power, toughness)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		t.Fatalf("%s: no endure choice on entry", name)
	}
	effectstest.AnswerOptionPick(t, g, me.ID, pick)
	spirits := effectstest.SocB8Named(g, me.ID, "Spirit")
	if pick == 1 {
		if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != n {
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
	if got := effectstest.CountersOn(g, id, game.CounterPlusOne); got != 0 {
		t.Errorf("%s: %d +1/+1 counters after choosing the Spirit", name, got)
	}
	return id
}

// --- Champion of Dusan ---------------------------------------------------------

func TestTdmChampionOfDusanHasTrampleAndRenewsACounterAndTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	champion := effectstest.B12Push(g, me.ID, "Champion of Dusan", "Creature — Human Warrior", effectstest.TdmChampionOfDusanOracle, 4, 2)
	if !effectstest.HasEffectiveKeyword(t, g, champion, "trample") {
		t.Error("Champion of Dusan has no trample")
	}
	bear := effectstest.SeedBear(g, me.ID)
	if effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Fatal("the Bear starts with trample")
	}
	id, _ := effectstest.ActivateFromGraveyard(t, g, "Champion of Dusan", "Creature — Human Warrior",
		effectstest.TdmChampionOfDusanOracle, 4, 2, "{G}{G}", game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if !g.Exile.Contains(id) {
		t.Fatal("renew's cost did not exile the Champion")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters, want 1", got)
	}
	if got := effectstest.CountersOn(g, bear, game.CounterTrample); got != 1 {
		t.Errorf("%d trample counters, want 1", got)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "trample") {
		t.Error("the trample counter did not give trample")
	}
}

func TestTdmChampionOfDusanRenewCostsOneGFromTheGraveyardOnly(t *testing.T) {
	abilities := game.ActivatedAbilitiesForCard(game.Card{OracleID: effectstest.TdmChampionOfDusanOracle})
	if len(abilities) != 1 || abilities[0].Cost.Mana != "{1}{G}" ||
		!game.AbilityFunctionsFromZone(abilities[0], game.ZoneGraveyard) ||
		game.AbilityFunctionsFromZone(abilities[0], game.ZoneBattlefield) {
		t.Fatalf("abilities %+v, want one {1}{G} ability that works only from the graveyard", abilities)
	}
}

// --- Cori Mountain Stalwart ----------------------------------------------------

func TestTdmCoriMountainStalwartDrainsTwoOnTheSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.B12Push(g, me.ID, "Cori Mountain Stalwart", "Creature — Human Monk", effectstest.TdmCoriMountainStalwartOrcl, 3, 3)
	effectstest.ToMainForCost(t, g)
	before := tdmCLLives(g)
	effectstest.TdmMPCastFreeSpell(t, g, me, "First")
	for _, p := range g.Seats {
		if p.Life != before[p.ID] {
			t.Fatalf("life moved on the first spell: %d -> %d", before[p.ID], p.Life)
		}
	}
	effectstest.TdmMPCastFreeSpell(t, g, me, "Second")
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
	effectstest.TdmMPCastFreeSpell(t, g, me, "Third")
	for _, p := range g.Seats {
		if p.Life != after[p.ID] {
			t.Errorf("life moved on the third spell: %d -> %d", after[p.ID], p.Life)
		}
	}
}

// --- Dalkovan Packbeasts -------------------------------------------------------

func TestTdmDalkovanPackbeastsHasVigilanceAndMobilizesThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ox := effectstest.B12Push(g, me.ID, "Dalkovan Packbeasts", "Creature — Ox", effectstest.TdmDalkovanPackbeastsOracle, 0, 4)
	if !effectstest.HasEffectiveKeyword(t, g, ox, "vigilance") {
		t.Error("Dalkovan Packbeasts has no vigilance")
	}
	tdmCLMobilizes(t, g, ox, 3)
}

// --- Devoted Duelist -----------------------------------------------------------

func TestTdmDevotedDuelistPingsEachOpponentOnTheSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	duelist := effectstest.B12Push(g, me.ID, "Devoted Duelist", "Creature — Goblin Monk", effectstest.TdmDevotedDuelistOracle, 2, 1)
	if !effectstest.HasEffectiveKeyword(t, g, duelist, "haste") {
		t.Error("Devoted Duelist has no haste")
	}
	effectstest.ToMainForCost(t, g)
	before := tdmCLLives(g)
	effectstest.TdmMPCastFreeSpell(t, g, me, "First")
	effectstest.TdmMPCastFreeSpell(t, g, me, "Second")
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	lancer := effectstest.B12Push(g, me.ID, "Dragonback Lancer", "Creature — Human Soldier", effectstest.TdmDragonbackLancerOracle, 3, 3)
	if !effectstest.HasEffectiveKeyword(t, g, lancer, "flying") {
		t.Error("Dragonback Lancer has no flying")
	}
	tdmCLMobilizes(t, g, lancer, 1)
}

// --- Dusyut Earthcarver --------------------------------------------------------

func TestTdmDusyutEarthcarverEnduresThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	carver := tdmCLEnduresOnEntry(t, g, "Dusyut Earthcarver", "Creature — Elephant Druid", effectstest.TdmDusyutEarthcarverOracle, 4, 4, 3, 1)
	if !effectstest.HasEffectiveKeyword(t, g, carver, "reach") {
		t.Error("Dusyut Earthcarver has no reach")
	}
}

func TestTdmDusyutEarthcarverCanMakeAThreeThreeSpirit(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	tdmCLEnduresOnEntry(t, g, "Dusyut Earthcarver", "Creature — Elephant Druid", effectstest.TdmDusyutEarthcarverOracle, 4, 4, 3, 0)
}

// --- Fortress Kin-Guard --------------------------------------------------------

func TestTdmFortressKinGuardEnduresOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	tdmCLEnduresOnEntry(t, g, "Fortress Kin-Guard", "Creature — Dog Soldier", effectstest.TdmFortressKinGuardOracle, 1, 2, 1, 1)
	g2 := effectstest.NewCatalogGame(t)
	tdmCLEnduresOnEntry(t, g2, "Fortress Kin-Guard", "Creature — Dog Soldier", effectstest.TdmFortressKinGuardOracle, 1, 2, 1, 0)
}

// --- Iceridge Serpent ----------------------------------------------------------

func TestTdmIceridgeSerpentBouncesAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	serpent := effectstest.CastHoldCreature(t, g, "Iceridge Serpent", "Creature — Serpent", effectstest.TdmIceridgeSerpentOracle, 3, 3)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	for _, bad := range []uuid.UUID{mine, serpent} {
		if err := g.ResolvePickTarget(p.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bad}); err == nil {
			t.Fatalf("%s was accepted as a target", bad)
		}
	}
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) || !opp.Hand.Contains(theirs) {
		t.Error("the opponent's creature was not returned to its owner's hand")
	}
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(serpent) {
		t.Error("your creature or the Serpent left the battlefield")
	}
}

// --- Inspirited Vanguard -------------------------------------------------------

func TestTdmInspiritedVanguardEnduresTwoOnEntry(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	tdmCLEnduresOnEntry(t, g, "Inspirited Vanguard", "Creature — Human Soldier", effectstest.TdmInspiritedVanguardOracle, 3, 2, 2, 1)
	g2 := effectstest.NewCatalogGame(t)
	tdmCLEnduresOnEntry(t, g2, "Inspirited Vanguard", "Creature — Human Soldier", effectstest.TdmInspiritedVanguardOracle, 3, 2, 2, 0)
}

func TestTdmInspiritedVanguardEnduresTwoWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	vanguard := effectstest.B12Push(g, me.ID, "Inspirited Vanguard", "Creature — Human Soldier", effectstest.TdmInspiritedVanguardOracle, 3, 2)
	effectstest.DeclareAttack(t, g, opp.ID, vanguard)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		t.Fatal("no endure choice when the Vanguard attacked")
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	if got := effectstest.CountersOn(g, vanguard, game.CounterPlusOne); got != 2 {
		t.Errorf("%d +1/+1 counters, want 2", got)
	}
	if n := len(effectstest.SocB8Named(g, me.ID, "Spirit")); n != 0 {
		t.Errorf("%d Spirits after choosing the counters", n)
	}
}

// --- Jeskai Devotee ------------------------------------------------------------

func TestTdmJeskaiDevoteeGrowsOnTheSecondSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	devotee := effectstest.B12Push(g, me.ID, "Jeskai Devotee", "Creature — Orc Monk", effectstest.TdmJeskaiDevoteeOracle, 2, 2)
	effectstest.ToMainForCost(t, g)
	effectstest.TdmMPCastFreeSpell(t, g, me, "First")
	if p := effectstest.EffectivePower(t, g, devotee); p != 2 {
		t.Fatalf("power %d after the first spell, want 2", p)
	}
	effectstest.TdmMPCastFreeSpell(t, g, me, "Second")
	if p, tough := effectstest.EffectivePower(t, g, devotee), effectstest.EffectiveToughness(t, g, devotee); p != 3 || tough != 3 {
		t.Errorf("the Devotee is %d/%d after the second spell, want 3/3", p, tough)
	}
}

func TestTdmJeskaiDevoteeFiltersOnceEachTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	devotee := effectstest.B12Push(g, me.ID, "Jeskai Devotee", "Creature — Orc Monk", effectstest.TdmJeskaiDevoteeOracle, 2, 2)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { me.ManaPool.EmptyPool() })
	effectstest.FillPoolColored(me, "G", 2)
	if err := g.ActivateManaAbility(me.ID, devotee, 0, game.ManaAbilityParams{Colors: []string{"B"}}); err == nil {
		t.Fatal("the Devotee made {B}; it adds {U}, {R}, or {W}")
	}
	if err := g.ActivateManaAbility(me.ID, devotee, 0, game.ManaAbilityParams{Colors: []string{"U"}}); err != nil {
		t.Fatalf("{1}: add {U}: %v", err)
	}
	if got := effectstest.McPoolCount(me, "U"); got != 1 {
		t.Errorf("blue in the pool %d, want 1", got)
	}
	if err := g.ActivateManaAbility(me.ID, devotee, 0, game.ManaAbilityParams{Colors: []string{"W"}}); err == nil {
		t.Error("a second activation this turn was allowed")
	}
}

// --- Jeskai Monument -----------------------------------------------------------

func TestTdmJeskaiMonumentFindsABasicIslandMountainOrPlains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Swamp", "Basic Land — Swamp"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Volcanic Island", "Land — Island Mountain"),
		effectstest.SearchTestLand("Island", "Basic Land — Island"),
	)
	effectstest.CastCatalogSpell(t, g, "Jeskai Monument", "Artifact", effectstest.TdmJeskaiMonumentOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("the Monument's entry asked for no search")
	}
	for _, bad := range []string{"Swamp", "Forest", "Volcanic Island"} {
		if effectstest.SearchOptionNamed(g, c, bad) != uuid.Nil {
			t.Errorf("%s was offered; only a basic Island, Mountain, or Plains may be found", bad)
		}
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Island")
	if _, ok := effectstest.HandCardNamed(me, "Island"); !ok {
		t.Error("the Island is not in hand")
	}
}

func TestTdmJeskaiMonumentMakesTwoFlyingBirds(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monument := effectstest.PushCatalogPermanent(g, me.ID, "Jeskai Monument", "Artifact", effectstest.TdmJeskaiMonumentOracle, false)
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "U", 2)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.FillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, monument, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(monument) {
		t.Fatal("the Monument was not sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	birds := effectstest.SocB8Named(g, me.ID, "Bird")
	if len(birds) != 2 {
		t.Fatalf("%d Birds, want 2", len(birds))
	}
	for _, b := range birds {
		if b.Power != 1 || b.Toughness != 1 || len(b.Colors) != 1 || b.Colors[0] != "W" {
			t.Errorf("Bird %d/%d %v, want a 1/1 white Bird", b.Power, b.Toughness, b.Colors)
		}
		if !effectstest.HasEffectiveKeyword(t, g, b.InstanceID, "flying") {
			t.Error("a Bird has no flying")
		}
	}
}

func TestTdmJeskaiMonumentActivatesOnlyAsASorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	monument := effectstest.PushCatalogPermanent(g, me.ID, "Jeskai Monument", "Artifact", effectstest.TdmJeskaiMonumentOracle, false)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FillPoolColored(me, "U", 2)
	effectstest.FillPoolColored(me, "R", 1)
	effectstest.FillPoolColored(me, "W", 1)
	if err := g.ActivateCatalogAbility(me.ID, monument, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("the Monument was activated outside a main phase")
	}
}

// --- Kin-Tree Nurturer ---------------------------------------------------------

func TestTdmKinTreeNurturerHasLifelinkAndEnduresOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	nurturer := tdmCLEnduresOnEntry(t, g, "Kin-Tree Nurturer", "Creature — Human Druid", effectstest.TdmKinTreeNurturerOracle, 2, 1, 1, 1)
	if !effectstest.HasEffectiveKeyword(t, g, nurturer, "lifelink") {
		t.Error("Kin-Tree Nurturer has no lifelink")
	}
	g2 := effectstest.NewCatalogGame(t)
	tdmCLEnduresOnEntry(t, g2, "Kin-Tree Nurturer", "Creature — Human Druid", effectstest.TdmKinTreeNurturerOracle, 2, 1, 1, 0)
	me := g2.Seats[g2.Turn.ActiveSeat]
	if spirits := effectstest.SocB8Named(g2, me.ID, "Spirit"); len(spirits) == 1 && effectstest.HasEffectiveKeyword(t, g2, spirits[0].InstanceID, "lifelink") {
		t.Error("the Spirit has lifelink; only the Nurturer does")
	}
}

package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// avatar_b30_c_cards_test.go — Avatar: The Last Airbender Eternal batches
// 30–33, group c (#5075).

const (
	avB30cMandrillOracle    = "8695c7e2-464e-464c-9d3f-a7cf025f3eb2"
	avB30cExemplarsOracle   = "4bb0596f-df1d-4977-bb8b-4d2e6648dcb3"
	avB30cSolidGroundOracle = "8609ed1a-f202-483e-9299-408ef6e84ad6"
	avB30cTophOracle        = "ad5d3e7e-bf8c-4fa3-a81e-1523aec2585f"
	avB30cAangOracle        = "854278bc-2bcf-45dd-abd5-6aca7a1e5fc9"
	avB30cHookSwordsOracle  = "610e5fda-ff6c-408d-9d58-ba2b02c5101f"
	avB30cKataraOracle      = "124cc815-0877-40f2-98c8-e7c4f8cd6598"
	avB30cStandUnitedOracle = "5a05ee6e-1812-4792-ad00-6df8fa4fd85f"
	avB30cZukoOracle        = "6e737c52-cbe0-493f-ae46-7bf068e04882"
	avB30cTectonicOracle    = "b71c3c1e-f927-4192-ac3a-cf6a442ffd00"
)

func TestAvatarB30cCardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		avB30cMandrillOracle:    "Elephant-Mandrill",
		avB30cExemplarsOracle:   "Kyoshi Warrior Exemplars",
		avB30cSolidGroundOracle: "Solid Ground",
		avB30cTophOracle:        "Toph, Earthbending Master",
		avB30cAangOracle:        "Aang, A Lot to Learn",
		avB30cHookSwordsOracle:  "Hook Swords",
		avB30cKataraOracle:      "Katara, Seeking Revenge",
		avB30cStandUnitedOracle: "Stand United",
		avB30cZukoOracle:        "Zuko, Seeking Honor",
		avB30cTectonicOracle:    "Tectonic Split",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// avB30cArtifactsOf counts the artifacts a player controls by name.
func avB30cCount(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

// avB30cCastALesson casts a Lesson sorcery from p's hand and resolves it,
// so it reaches the graveyard by a real zone move (which refreshes the
// layers, as a hand-placed graveyard card does not).
func avB30cCastALesson(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	advanceToMain(t, g)
	castFromHandForTest(t, g, p, "A Lesson", "Sorcery — Lesson", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
}

// --- Elephant-Mandrill ----------------------------------------------------------

func TestAvatarB30ElephantMandrillFeedsTheTableAndGrowsOnOpponentsArtifacts(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	pushTypedCard(g, opp.ID, "Their Rock", "Artifact", "{1}")
	pushTypedCard(g, me.ID, "My Rock", "Artifact", "{1}")
	mandrill := avB14Enter(t, g, me, game.Card{Name: "Elephant-Mandrill", TypeLine: "Creature — Elephant Monkey",
		OracleID: avB30cMandrillOracle, Power: 3, Toughness: 2})
	passPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if got := avB30cCount(g, p.ID, "Food"); got != 1 {
			t.Errorf("seat %s has %d Foods, want 1", p.ID, got)
		}
	}
	if !hasEffectiveKeyword(t, g, mandrill, "reach") {
		t.Error("Elephant-Mandrill has no reach")
	}
	if p, th := effectivePTOf(t, g, mandrill); p != 3 || th != 2 {
		t.Fatalf("before combat the Mandrill is %d/%d, want 3/2", p, th)
	}
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	// Three opponents' Foods and their rock; my own Food and rock don't count.
	if p, th := effectivePTOf(t, g, mandrill); p != 7 || th != 6 {
		t.Errorf("at the beginning of combat the Mandrill is %d/%d, want 7/6", p, th)
	}
	// The bonus is locked in: a new opposing artifact changes nothing.
	pushTypedCard(g, opp.ID, "Late Rock", "Artifact", "{1}")
	if p, _ := effectivePTOf(t, g, mandrill); p != 7 {
		t.Errorf("a later artifact changed the bonus: power %d, want 7", p)
	}
	advanceToNextSeatsTurn(t, g)
	if p, th := effectivePTOf(t, g, mandrill); p != 3 || th != 2 {
		t.Errorf("after the turn the Mandrill is %d/%d, want 3/2", p, th)
	}
}

// --- Kyoshi Warrior Exemplars ---------------------------------------------------

func avB30cExemplarsAttack(t *testing.T, lands int) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	for i := 0; i < lands; i++ {
		pushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	}
	exemplars := pushCatalogPermanent(g, me.ID, "Kyoshi Warrior Exemplars", "Creature — Human Warrior Ally", avB30cExemplarsOracle, false)
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	advanceToMain(t, g)
	declareAttack(t, g, opp.ID, exemplars)
	return g, me, exemplars, mine, theirs
}

func TestAvatarB30KyoshiWarriorExemplarsPumpsTheTeamWithEightLands(t *testing.T) {
	g, _, exemplars, mine, theirs := avB30cExemplarsAttack(t, 8)
	passPriorityAroundTable(t, g)
	if p, th := effectivePTOf(t, g, exemplars); p != 3 || th != 3 {
		t.Errorf("the Exemplars are %d/%d, want 3/3", p, th)
	}
	if p, th := effectivePTOf(t, g, mine); p != 4 || th != 4 {
		t.Errorf("my other creature is %d/%d, want 4/4", p, th)
	}
	if p, th := effectivePTOf(t, g, theirs); p != 2 || th != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, th)
	}
}

func TestAvatarB30KyoshiWarriorExemplarsNeedsEightLandsBothTimes(t *testing.T) {
	g, _, _, mine, _ := avB30cExemplarsAttack(t, 7)
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, mine); p != 2 {
		t.Errorf("seven lands pumped the team: power %d, want 2", p)
	}

	// Eight as it triggers, seven as it resolves: the ability does nothing (CR 603.4).
	g, me, _, mine, _ := avB30cExemplarsAttack(t, 8)
	var land uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsLand() {
			land = c.InstanceID
		}
	}
	killCreature(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	if p, _ := effectivePTOf(t, g, mine); p != 2 {
		t.Errorf("a land lost in response still pumped the team: power %d, want 2", p)
	}
}

// --- Solid Ground ---------------------------------------------------------------

func TestAvatarB30SolidGroundEarthbendsFourAndAddsACounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	land := avB14Forest(g, me.ID)
	avB14Enter(t, g, me, game.Card{Name: "Solid Ground", TypeLine: "Enchantment", OracleID: avB30cSolidGroundOracle})
	pickTriggerTarget(t, g, me.ID, land)
	passPriorityAroundTable(t, g)
	// Earthbend 3 under its own replacement: four counters.
	avB14Earthbent(t, g, land, 4)

	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if err := g.AddCounter(mine, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if err := g.AddCounter(theirs, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := countersOn(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("my creature has %d counters, want 2", got)
	}
	if got := countersOn(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("an opponent's creature has %d counters, want 1", got)
	}
}

// --- Toph, Earthbending Master --------------------------------------------------

func TestAvatarB30TophGainsExperienceOnLandfallAndEarthbendsByIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	land := avB14Forest(g, me.ID)
	toph := pushCatalogPermanent(g, me.ID, "Toph, Earthbending Master", "Legendary Creature — Human Warrior Ally", avB30cTophOracle, false)
	advanceToMain(t, g)

	avB14LandEnters(t, g, opp)
	passPriorityAroundTable(t, g)
	if got := experienceCounters(g, me.ID); got != 0 {
		t.Fatalf("an opponent's land gave %d experience counters", got)
	}
	avB14LandEnters(t, g, me)
	passPriorityAroundTable(t, g)
	avB14LandEnters(t, g, me)
	passPriorityAroundTable(t, g)
	if got := experienceCounters(g, me.ID); got != 2 {
		t.Fatalf("two lands gave %d experience counters, want 2", got)
	}

	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	declareAttack(t, g, opp.ID, toph, bear)
	// One trigger for the whole attack (CR 508.3d).
	pickTriggerTarget(t, g, me.ID, land)
	if latestPickTarget(g, me.ID) != nil {
		t.Fatal("a two-creature attack triggered Toph twice")
	}
	passPriorityAroundTable(t, g)
	avB14Earthbent(t, g, land, 2)
}

// --- Aang, A Lot to Learn -------------------------------------------------------

func TestAvatarB30AangHasVigilanceWithALessonAndGrowsOnDeaths(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	aang := pushCatalogPermanent(g, me.ID, "Aang, A Lot to Learn", "Legendary Creature — Human Avatar Ally", avB30cAangOracle, false)
	if hasEffectiveKeyword(t, g, aang, "vigilance") {
		t.Fatal("Aang has vigilance with no Lesson in the graveyard")
	}
	avB4Lesson(g, opp)
	seedCreature(g, "Nudge", me.ID) // a battlefield change refreshes the layer cache
	if hasEffectiveKeyword(t, g, aang, "vigilance") {
		t.Fatal("an opponent's Lesson gave Aang vigilance")
	}
	avB30cCastALesson(t, g, me)
	if !hasEffectiveKeyword(t, g, aang, "vigilance") {
		t.Fatal("Aang has no vigilance with a Lesson in your graveyard")
	}

	theirs := pushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	killCreature(t, g, opp.ID, theirs)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := countersOn(g, aang, game.CounterPlusOne); got != 0 {
		t.Fatalf("an opponent's creature dying gave Aang %d counters", got)
	}
	mine := pushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	killCreature(t, g, me.ID, mine)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := countersOn(g, aang, game.CounterPlusOne); got != 1 {
		t.Errorf("Aang has %d counters after another creature you control died, want 1", got)
	}
}

// --- Hook Swords ----------------------------------------------------------------

func TestAvatarB30HookSwordsMakesAnAllyAndArmsItOnYourTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	swords := avB14Enter(t, g, me, game.Card{Name: "Hook Swords", TypeLine: "Artifact — Equipment", OracleID: avB30cHookSwordsOracle})
	passPriorityAroundTable(t, g)
	ally := findBattlefieldByName(g, "Ally")
	if ally == uuid.Nil {
		t.Fatal("no Ally token")
	}
	c := avB1Card(t, g, swords)
	if !c.IsAttachedTo(ally) {
		t.Fatal("Hook Swords is not attached to the Ally")
	}
	if p, th := effectivePTOf(t, g, ally); p != 2 || th != 2 {
		t.Errorf("on your turn the Ally is %d/%d, want 2/2", p, th)
	}
	if !hasEffectiveKeyword(t, g, ally, "first strike") {
		t.Error("on your turn the Ally has no first strike")
	}
	advanceToNextSeatsTurn(t, g)
	if p, th := effectivePTOf(t, g, ally); p != 1 || th != 1 {
		t.Errorf("on another turn the Ally is %d/%d, want 1/1", p, th)
	}
	if hasEffectiveKeyword(t, g, ally, "first strike") {
		t.Error("on another turn the Ally has first strike")
	}
}

// --- Katara, Seeking Revenge ----------------------------------------------------

func avB30cKatara() game.Card {
	return game.Card{Name: "Katara, Seeking Revenge", TypeLine: "Legendary Creature — Human Warrior Ally",
		OracleID: avB30cKataraOracle, Power: 3, Toughness: 3}
}

func TestAvatarB30KataraLootsWhenTheWaterbendWasNotPaid(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	avB10Fillers(me, 3)
	katara := avB14Enter(t, g, me, avB30cKatara())
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d after the draw, want %d", me.Hand.Size(), hand+1)
	}
	if discardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt when the waterbend wasn't paid")
	}
	discardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}

	avB30cCastALesson(t, g, me)
	avB30cCastALesson(t, g, me)
	if p, th := effectivePTOf(t, g, katara); p != 5 || th != 5 {
		t.Errorf("with two Lessons Katara is %d/%d, want 5/5", p, th)
	}
}

func TestAvatarB30KataraOnlyDrawsWhenTheWaterbendWasPaid(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	avB10Fillers(me, 3)
	a := pushVanillaCreature(g, me.ID, "Tapper A", 1, 1)
	b := pushVanillaCreature(g, me.ID, "Tapper B", 1, 1)

	short := avB1Hand(me, avB30cKatara())
	if err := g.CastSpell(me.ID, short, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, TapIDs: []uuid.UUID{a}}); err == nil {
		t.Fatal("one tap and no mana paid waterbend {2}")
	}
	if err := g.CastSpell(me.ID, short, game.CastSpellParams{Strict: true, OptionalCosts: []int{0}, TapIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("waterbent cast: %v", err)
	}
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		if err := g.PassPriority(); err != nil {
			break
		}
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d after the draw, want %d", me.Hand.Size(), hand+1)
	}
	if discardChoiceFor(g, me.ID) != nil {
		t.Error("a discard prompt after the waterbend was paid")
	}
}

// --- Stand United ---------------------------------------------------------------

func TestAvatarB30StandUnitedScriesOnlyForAnAlly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := pushVanillaCreature(g, me.ID, "Bear", 2, 2)
	ally := b21Push(g, me.ID, "Ally Bear", "Creature — Bear Ally", "", 2, 2)

	castCatalogSpell(t, g, "Stand United", "Instant", avB30cStandUnitedOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	passPriorityAroundTable(t, g)
	if p, th := effectivePTOf(t, g, bear); p != 4 || th != 4 {
		t.Errorf("the bear is %d/%d, want 4/4", p, th)
	}
	if scryChoiceFor(g, me.ID) != nil {
		t.Fatal("Stand United scried for a non-Ally")
	}

	castCatalogSpell(t, g, "Stand United", "Instant", avB30cStandUnitedOracle, []game.TargetRef{{Kind: game.TargetCard, ID: ally}})
	passPriorityAroundTable(t, g)
	if p, th := effectivePTOf(t, g, ally); p != 4 || th != 4 {
		t.Errorf("the Ally is %d/%d, want 4/4", p, th)
	}
	c := scryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt %+v, want one of two cards", c)
	}
}

// --- Zuko, Seeking Honor --------------------------------------------------------

func TestAvatarB30ZukoFirebendsGainsFirstStrikeAndGrows(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	zuko := pushCatalogPermanent(g, me.ID, "Zuko, Seeking Honor", "Legendary Creature — Human Noble", avB30cZukoOracle, false)

	castCatalogSpell(t, g, "Bear Cub", "Creature — Bear", "", nil)
	passPriorityAroundTable(t, g)
	if hasEffectiveKeyword(t, g, zuko, "first strike") {
		t.Fatal("a creature spell gave Zuko first strike")
	}
	castCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, zuko, "first strike") {
		t.Fatal("a noncreature spell didn't give Zuko first strike")
	}

	g.WithWriteLock(func() { me.ManaPool = nil })
	declareAttack(t, g, opp.ID, zuko)
	passPriorityAroundTable(t, g)
	if got := sortedPool(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("after the attack the pool is %v, want {R}", got)
	}

	dealCombatDamageToPlayer(g, zuko, opp.ID, 1)
	g.RunStateChecksForTest()
	passPriorityAroundTable(t, g)
	if got := countersOn(g, zuko, game.CounterPlusOne); got != 1 {
		t.Errorf("Zuko has %d counters after combat damage to a player, want 1", got)
	}
}

// --- Tectonic Split -------------------------------------------------------------

func TestAvatarB30TectonicSplitSacrificesHalfTheLandsRoundedUp(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	var lands []uuid.UUID
	for i := 0; i < 5; i++ {
		lands = append(lands, pushLandFor(g, me.ID, "Forest", "Basic Land — Forest"))
	}
	theirs := pushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	split := avB1Hand(me, game.Card{Name: "Tectonic Split", TypeLine: "Enchantment", OracleID: avB30cTectonicOracle})

	// Five lands: three, no more and no fewer, and only your own.
	for _, bad := range [][]uuid.UUID{lands[:2], lands[:4], {lands[0], lands[1], theirs}} {
		if err := g.CastSpell(me.ID, split, game.CastSpellParams{Strict: true, SacrificeIDs: bad}); err == nil {
			t.Fatalf("Tectonic Split was cast sacrificing %v", bad)
		}
	}
	if err := g.CastSpell(me.ID, split, game.CastSpellParams{Strict: true, SacrificeIDs: lands[:3]}); err != nil {
		t.Fatalf("cast sacrificing three of five lands: %v", err)
	}
	for _, id := range lands[:3] {
		if g.Battlefield.Contains(id) {
			t.Fatalf("a sacrificed land %s is still on the battlefield with the spell on the stack", id)
		}
	}
	passPriorityAroundTable(t, g)
	split = findBattlefieldByName(g, "Tectonic Split")
	if split == uuid.Nil {
		t.Fatal("Tectonic Split didn't resolve")
	}
	if !hasEffectiveKeyword(t, g, split, "hexproof") {
		t.Error("Tectonic Split has no hexproof")
	}
	if n := grantedManaRowsOf(g, theirs); n != 0 {
		t.Errorf("an opponent's land has %d granted mana rows, want 0", n)
	}
	g.WithWriteLock(func() { me.ManaPool = nil })
	if err := tapGrantedMana(t, g, me.ID, lands[3], "U"); err != nil {
		t.Fatalf("tap a land for the granted mana: %v", err)
	}
	if got := batch01PoolColors(me); len(got) != 3 || got[0] != "U" || got[1] != "U" || got[2] != "U" {
		t.Errorf("pool %v, want three blue", got)
	}
}

func TestAvatarB30TectonicSplitWithNoLandsSacrificesNone(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	split := avB1Hand(me, game.Card{Name: "Tectonic Split", TypeLine: "Enchantment", OracleID: avB30cTectonicOracle})
	if err := g.CastSpell(me.ID, split, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast with no lands: %v", err)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldByName(g, "Tectonic Split") == uuid.Nil {
		t.Fatal("Tectonic Split didn't resolve")
	}
}

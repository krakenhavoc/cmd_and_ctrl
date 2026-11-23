package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
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
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
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
	effectstest.AdvanceToMain(t, g)
	effectstest.CastFromHandForTest(t, g, p, "A Lesson", "Sorcery — Lesson", "", "", game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
}

// --- Elephant-Mandrill ----------------------------------------------------------

func TestAvatarB30ElephantMandrillFeedsTheTableAndGrowsOnOpponentsArtifacts(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.PushTypedCard(g, opp.ID, "Their Rock", "Artifact", "{1}")
	effectstest.PushTypedCard(g, me.ID, "My Rock", "Artifact", "{1}")
	mandrill := avB14Enter(t, g, me, game.Card{Name: "Elephant-Mandrill", TypeLine: "Creature — Elephant Monkey",
		OracleID: avB30cMandrillOracle, Power: 3, Toughness: 2})
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if got := avB30cCount(g, p.ID, "Food"); got != 1 {
			t.Errorf("seat %s has %d Foods, want 1", p.ID, got)
		}
	}
	if !effectstest.HasEffectiveKeyword(t, g, mandrill, "reach") {
		t.Error("Elephant-Mandrill has no reach")
	}
	if p, th := effectstest.EffectivePTOf(t, g, mandrill); p != 3 || th != 2 {
		t.Fatalf("before combat the Mandrill is %d/%d, want 3/2", p, th)
	}
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.PassPriorityAroundTable(t, g)
	// Three opponents' Foods and their rock; my own Food and rock don't count.
	if p, th := effectstest.EffectivePTOf(t, g, mandrill); p != 7 || th != 6 {
		t.Errorf("at the beginning of combat the Mandrill is %d/%d, want 7/6", p, th)
	}
	// The bonus is locked in: a new opposing artifact changes nothing.
	effectstest.PushTypedCard(g, opp.ID, "Late Rock", "Artifact", "{1}")
	if p, _ := effectstest.EffectivePTOf(t, g, mandrill); p != 7 {
		t.Errorf("a later artifact changed the bonus: power %d, want 7", p)
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, mandrill); p != 3 || th != 2 {
		t.Errorf("after the turn the Mandrill is %d/%d, want 3/2", p, th)
	}
}

// --- Kyoshi Warrior Exemplars ---------------------------------------------------

func avB30cExemplarsAttack(t *testing.T, lands int) (*game.Game, *game.Player, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	for i := 0; i < lands; i++ {
		effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest")
	}
	exemplars := effectstest.PushCatalogPermanent(g, me.ID, "Kyoshi Warrior Exemplars", "Creature — Human Warrior Ally", avB30cExemplarsOracle, false)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, exemplars)
	return g, me, exemplars, mine, theirs
}

func TestAvatarB30KyoshiWarriorExemplarsPumpsTheTeamWithEightLands(t *testing.T) {
	g, _, exemplars, mine, theirs := avB30cExemplarsAttack(t, 8)
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, exemplars); p != 3 || th != 3 {
		t.Errorf("the Exemplars are %d/%d, want 3/3", p, th)
	}
	if p, th := effectstest.EffectivePTOf(t, g, mine); p != 4 || th != 4 {
		t.Errorf("my other creature is %d/%d, want 4/4", p, th)
	}
	if p, th := effectstest.EffectivePTOf(t, g, theirs); p != 2 || th != 2 {
		t.Errorf("an opponent's creature is %d/%d, want 2/2", p, th)
	}
}

func TestAvatarB30KyoshiWarriorExemplarsNeedsEightLandsBothTimes(t *testing.T) {
	g, _, _, mine, _ := avB30cExemplarsAttack(t, 7)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, mine); p != 2 {
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
	effectstest.KillCreature(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.EffectivePTOf(t, g, mine); p != 2 {
		t.Errorf("a land lost in response still pumped the team: power %d, want 2", p)
	}
}

// --- Solid Ground ---------------------------------------------------------------

func TestAvatarB30SolidGroundEarthbendsFourAndAddsACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	land := effectstest.AvB14Forest(g, me.ID)
	avB14Enter(t, g, me, game.Card{Name: "Solid Ground", TypeLine: "Enchantment", OracleID: avB30cSolidGroundOracle})
	effectstest.PickTriggerTarget(t, g, me.ID, land)
	effectstest.PassPriorityAroundTable(t, g)
	// Earthbend 3 under its own replacement: four counters.
	avB14Earthbent(t, g, land, 4)

	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if err := g.AddCounter(mine, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if err := g.AddCounter(theirs, game.CounterPlusOne, 1); err != nil {
		t.Fatalf("AddCounter: %v", err)
	}
	if got := effectstest.CountersOn(g, mine, game.CounterPlusOne); got != 2 {
		t.Errorf("my creature has %d counters, want 2", got)
	}
	if got := effectstest.CountersOn(g, theirs, game.CounterPlusOne); got != 1 {
		t.Errorf("an opponent's creature has %d counters, want 1", got)
	}
}

// --- Toph, Earthbending Master --------------------------------------------------

func TestAvatarB30TophGainsExperienceOnLandfallAndEarthbendsByIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	land := effectstest.AvB14Forest(g, me.ID)
	toph := effectstest.PushCatalogPermanent(g, me.ID, "Toph, Earthbending Master", "Legendary Creature — Human Warrior Ally", avB30cTophOracle, false)
	effectstest.AdvanceToMain(t, g)

	avB14LandEnters(t, g, opp)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effects.ExperienceCounters(g, me.ID); got != 0 {
		t.Fatalf("an opponent's land gave %d experience counters", got)
	}
	avB14LandEnters(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	avB14LandEnters(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effects.ExperienceCounters(g, me.ID); got != 2 {
		t.Fatalf("two lands gave %d experience counters, want 2", got)
	}

	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, toph, bear)
	// One trigger for the whole attack (CR 508.3d).
	effectstest.PickTriggerTarget(t, g, me.ID, land)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Fatal("a two-creature attack triggered Toph twice")
	}
	effectstest.PassPriorityAroundTable(t, g)
	avB14Earthbent(t, g, land, 2)
}

// --- Aang, A Lot to Learn -------------------------------------------------------

func TestAvatarB30AangHasVigilanceWithALessonAndGrowsOnDeaths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	aang := effectstest.PushCatalogPermanent(g, me.ID, "Aang, A Lot to Learn", "Legendary Creature — Human Avatar Ally", avB30cAangOracle, false)
	if effectstest.HasEffectiveKeyword(t, g, aang, "vigilance") {
		t.Fatal("Aang has vigilance with no Lesson in the graveyard")
	}
	avB4Lesson(g, opp)
	effectstest.SeedCreature(g, "Nudge", me.ID) // a battlefield change refreshes the layer cache
	if effectstest.HasEffectiveKeyword(t, g, aang, "vigilance") {
		t.Fatal("an opponent's Lesson gave Aang vigilance")
	}
	avB30cCastALesson(t, g, me)
	if !effectstest.HasEffectiveKeyword(t, g, aang, "vigilance") {
		t.Fatal("Aang has no vigilance with a Lesson in your graveyard")
	}

	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	effectstest.KillCreature(t, g, opp.ID, theirs)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, aang, game.CounterPlusOne); got != 0 {
		t.Fatalf("an opponent's creature dying gave Aang %d counters", got)
	}
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	effectstest.KillCreature(t, g, me.ID, mine)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, aang, game.CounterPlusOne); got != 1 {
		t.Errorf("Aang has %d counters after another creature you control died, want 1", got)
	}
}

// --- Hook Swords ----------------------------------------------------------------

func TestAvatarB30HookSwordsMakesAnAllyAndArmsItOnYourTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	swords := avB14Enter(t, g, me, game.Card{Name: "Hook Swords", TypeLine: "Artifact — Equipment", OracleID: avB30cHookSwordsOracle})
	effectstest.PassPriorityAroundTable(t, g)
	ally := effectstest.FindBattlefieldByName(g, "Ally")
	if ally == uuid.Nil {
		t.Fatal("no Ally token")
	}
	c := effectstest.AvB1Card(t, g, swords)
	if !c.IsAttachedTo(ally) {
		t.Fatal("Hook Swords is not attached to the Ally")
	}
	if p, th := effectstest.EffectivePTOf(t, g, ally); p != 2 || th != 2 {
		t.Errorf("on your turn the Ally is %d/%d, want 2/2", p, th)
	}
	if !effectstest.HasEffectiveKeyword(t, g, ally, "first strike") {
		t.Error("on your turn the Ally has no first strike")
	}
	effectstest.AdvanceToNextSeatsTurn(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, ally); p != 1 || th != 1 {
		t.Errorf("on another turn the Ally is %d/%d, want 1/1", p, th)
	}
	if effectstest.HasEffectiveKeyword(t, g, ally, "first strike") {
		t.Error("on another turn the Ally has first strike")
	}
}

// --- Katara, Seeking Revenge ----------------------------------------------------

func avB30cKatara() game.Card {
	return game.Card{Name: "Katara, Seeking Revenge", TypeLine: "Legendary Creature — Human Warrior Ally",
		OracleID: avB30cKataraOracle, Power: 3, Toughness: 3}
}

func TestAvatarB30KataraLootsWhenTheWaterbendWasNotPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	avB10Fillers(me, 3)
	katara := avB14Enter(t, g, me, avB30cKatara())
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Fatalf("hand %d after the draw, want %d", me.Hand.Size(), hand+1)
	}
	if effectstest.DiscardChoiceFor(g, me.ID) == nil {
		t.Fatal("no discard prompt when the waterbend wasn't paid")
	}
	effectstest.DiscardFromHand(t, g, me.ID)
	if me.Hand.Size() != hand {
		t.Errorf("hand %d after the discard, want %d", me.Hand.Size(), hand)
	}

	avB30cCastALesson(t, g, me)
	avB30cCastALesson(t, g, me)
	if p, th := effectstest.EffectivePTOf(t, g, katara); p != 5 || th != 5 {
		t.Errorf("with two Lessons Katara is %d/%d, want 5/5", p, th)
	}
}

func TestAvatarB30KataraOnlyDrawsWhenTheWaterbendWasPaid(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	avB10Fillers(me, 3)
	a := effectstest.PushVanillaCreature(g, me.ID, "Tapper A", 1, 1)
	b := effectstest.PushVanillaCreature(g, me.ID, "Tapper B", 1, 1)

	short := effectstest.AvB1Hand(me, avB30cKatara())
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
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d after the draw, want %d", me.Hand.Size(), hand+1)
	}
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("a discard prompt after the waterbend was paid")
	}
}

// --- Stand United ---------------------------------------------------------------

func TestAvatarB30StandUnitedScriesOnlyForAnAlly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	ally := effectstest.B21Push(g, me.ID, "Ally Bear", "Creature — Bear Ally", "", 2, 2)

	effectstest.CastCatalogSpell(t, g, "Stand United", "Instant", avB30cStandUnitedOracle, []game.TargetRef{{Kind: game.TargetCard, ID: bear}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, bear); p != 4 || th != 4 {
		t.Errorf("the bear is %d/%d, want 4/4", p, th)
	}
	if effectstest.ScryChoiceFor(g, me.ID) != nil {
		t.Fatal("Stand United scried for a non-Ally")
	}

	effectstest.CastCatalogSpell(t, g, "Stand United", "Instant", avB30cStandUnitedOracle, []game.TargetRef{{Kind: game.TargetCard, ID: ally}})
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePTOf(t, g, ally); p != 4 || th != 4 {
		t.Errorf("the Ally is %d/%d, want 4/4", p, th)
	}
	c := effectstest.ScryChoiceFor(g, me.ID)
	if c == nil || len(c.ScryCards) != 2 {
		t.Fatalf("scry prompt %+v, want one of two cards", c)
	}
}

// --- Zuko, Seeking Honor --------------------------------------------------------

func TestAvatarB30ZukoFirebendsGainsFirstStrikeAndGrows(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	zuko := effectstest.PushCatalogPermanent(g, me.ID, "Zuko, Seeking Honor", "Legendary Creature — Human Noble", avB30cZukoOracle, false)

	effectstest.CastCatalogSpell(t, g, "Bear Cub", "Creature — Bear", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasEffectiveKeyword(t, g, zuko, "first strike") {
		t.Fatal("a creature spell gave Zuko first strike")
	}
	effectstest.CastCatalogSpell(t, g, "Plain Sorcery", "Sorcery", "", nil)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, zuko, "first strike") {
		t.Fatal("a noncreature spell didn't give Zuko first strike")
	}

	g.WithWriteLock(func() { me.ManaPool = nil })
	effectstest.DeclareAttack(t, g, opp.ID, zuko)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.SortedPool(me); len(got) != 1 || got[0] != "R" {
		t.Errorf("after the attack the pool is %v, want {R}", got)
	}

	effectstest.DealCombatDamageToPlayer(g, zuko, opp.ID, 1)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CountersOn(g, zuko, game.CounterPlusOne); got != 1 {
		t.Errorf("Zuko has %d counters after combat damage to a player, want 1", got)
	}
}

// --- Tectonic Split -------------------------------------------------------------

func TestAvatarB30TectonicSplitSacrificesHalfTheLandsRoundedUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	var lands []uuid.UUID
	for i := 0; i < 5; i++ {
		lands = append(lands, effectstest.PushLandFor(g, me.ID, "Forest", "Basic Land — Forest"))
	}
	theirs := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	split := effectstest.AvB1Hand(me, game.Card{Name: "Tectonic Split", TypeLine: "Enchantment", OracleID: avB30cTectonicOracle})

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
	effectstest.PassPriorityAroundTable(t, g)
	split = effectstest.FindBattlefieldByName(g, "Tectonic Split")
	if split == uuid.Nil {
		t.Fatal("Tectonic Split didn't resolve")
	}
	if !effectstest.HasEffectiveKeyword(t, g, split, "hexproof") {
		t.Error("Tectonic Split has no hexproof")
	}
	if n := effectstest.GrantedManaRowsOf(g, theirs); n != 0 {
		t.Errorf("an opponent's land has %d granted mana rows, want 0", n)
	}
	g.WithWriteLock(func() { me.ManaPool = nil })
	if err := effectstest.TapGrantedMana(t, g, me.ID, lands[3], "U"); err != nil {
		t.Fatalf("tap a land for the granted mana: %v", err)
	}
	if got := effectstest.Batch01PoolColors(me); len(got) != 3 || got[0] != "U" || got[1] != "U" || got[2] != "U" {
		t.Errorf("pool %v, want three blue", got)
	}
}

func TestAvatarB30TectonicSplitWithNoLandsSacrificesNone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	split := effectstest.AvB1Hand(me, game.Card{Name: "Tectonic Split", TypeLine: "Enchantment", OracleID: avB30cTectonicOracle})
	if err := g.CastSpell(me.ID, split, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast with no lands: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.FindBattlefieldByName(g, "Tectonic Split") == uuid.Nil {
		t.Fatal("Tectonic Split didn't resolve")
	}
}

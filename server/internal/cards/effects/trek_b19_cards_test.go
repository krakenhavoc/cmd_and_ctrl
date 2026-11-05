package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Star Trek batch 19 (#4973): each card's printed text, driven through
// the engine.

const (
	trekB19AdversaryOracle = "37f3eda9-3242-4d62-846f-935543d8f4fd"
	trekB19AmokTimeOracle  = "98a752ff-f041-4032-9f3f-c6c4007f10f1"
	trekB19TerraformOracle = "c5a7a82e-3c13-4fa1-8bf5-bb85b2890f75"
	trekB19BranchingOracle = "6e13c619-706d-4c77-bc7d-187df0fd054a"
	trekB19ArcherOracle    = "85df3564-5134-4dc2-9021-40ec84ee6d08"
	trekB19JanewayOracle   = "d77d1895-ce02-4302-87c9-649e852fd0ea"
	trekB19PikeOracle      = "745633e6-8635-4d99-8a4e-edc5d500bb3d"
	trekB19CrewOracle      = "d5e9b04f-4a97-4da1-bf86-3715fc75ab23"
	trekB19CommonGoal      = "d4030d30-4c24-458c-87a9-c2f80bb68d27"
	trekB19CulberOracle    = "5b8e8ef2-2f04-4485-ae12-08f52dac81a2"
)

func TestTrekB19EveryCardIsRegisteredFull(t *testing.T) {
	want := map[string]string{
		trekB19AdversaryOracle: "Adversary by Algorithm",
		trekB19AmokTimeOracle:  "Amok Time",
		trekB19TerraformOracle: "Astute Terraformer",
		trekB19BranchingOracle: "Branching Timeline",
		trekB19ArcherOracle:    "Captain Jonathan Archer",
		trekB19JanewayOracle:   "Captain Kathryn Janeway",
		trekB19PikeOracle:      "Captain Pike, Fullness of Time",
		trekB19CrewOracle:      "Cold-Blooded Crew",
		trekB19CommonGoal:      "Common Goal",
		trekB19CulberOracle:    "Culber and Stamets, Star-Crossed",
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

// trekB19Lands puts an untapped-but-irrelevant land per name onto p's
// side, tapped so it pays for nothing.
func trekB19Lands(g *game.Game, p uuid.UUID, names ...string) {
	for _, n := range names {
		tmtB4Tap(g, b12Permanent(g, p, n, "Basic Land — "+n))
	}
}

func trekB19Ref(id uuid.UUID, slot int) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id, Slot: slot}
}

// trekB19Named returns the ids of the battlefield permanents with `name`.
func trekB19Named(t *testing.T, g *game.Game, name string) []game.Card {
	t.Helper()
	var out []game.Card
	for _, id := range battlefieldIDsNamed(g, name) {
		out = append(out, layeredCard(t, g, id))
	}
	return out
}

// --- Adversary by Algorithm ---------------------------------------------------

func TestTrekB19AdversaryMakesAColorlessIllusionArtifact(t *testing.T) {
	g, me, _ := trekB8Main(t)
	trekB8Float(t, g, me, "{2}{G}")
	castFromHandForTest(t, g, me, "Adversary by Algorithm", "Sorcery", "{2}{G}", trekB19AdversaryOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	tokens := trekB19Named(t, g, "Illusion")
	if len(tokens) != 1 {
		t.Fatalf("%d Illusions, want 1", len(tokens))
	}
	c := tokens[0]
	if !c.IsToken() || !c.IsArtifact() || !c.IsCreature() || len(c.Colors) != 0 || c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Errorf("Illusion = %+v, want a 3/3 colorless artifact creature token", c)
	}
}

func TestTrekB19AdversaryBeamsUpFromTheGraveyard(t *testing.T) {
	g, me, opp := trekB8Main(t)
	mine := b12Creature(g, me.ID, "Away Team", "Creature — Human", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	card := pushToGraveyard(g, me, game.Card{Name: "Adversary by Algorithm", TypeLine: "Sorcery", ManaCost: "{2}{G}", OracleID: trekB19AdversaryOracle})
	trekB8Float(t, g, me, "{3}{G}")
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyBeamMeUp, AltCostIDs: []uuid.UUID{theirs}, Strict: true,
	}); err == nil {
		t.Fatal("beam me up returned an opponent's creature")
	}
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Fatal("the card was cast from the graveyard without beam me up")
	}
	if err := g.CastSpell(me.ID, card, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: game.AltCostKeyBeamMeUp, AltCostIDs: []uuid.UUID{mine}, Strict: true,
	}); err != nil {
		t.Fatalf("beam me up cast: %v", err)
	}
	if !me.Hand.Contains(mine) {
		t.Fatal("the creature was not returned to hand as the cost")
	}
	passPriorityAroundTable(t, g)
	if len(trekB19Named(t, g, "Illusion")) != 1 {
		t.Error("the beamed-up spell made no Illusion")
	}
	if me.Graveyard.Contains(card) || !g.Exile.Contains(card) {
		t.Error("a beamed-up spell must be exiled as it leaves the stack")
	}
}

// --- Amok Time ----------------------------------------------------------------

func trekB19AmokTime(g *game.Game, owner uuid.UUID, lore int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Amok Time", TypeLine: "Enchantment — Saga",
		OracleID: trekB19AmokTimeOracle, ManaCost: "{2}{G}", Colors: []string{"G"},
		Owner: owner, Controller: owner, Counters: map[string]int{game.CounterLore: lore},
	})
}

func TestTrekB19AmokTimeChapterOneFights(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Vulcan", "Creature — Vulcan", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	trekB19AmokTime(g, me.ID, 0)
	advanceToMain(t, g)
	passPriorityAroundTable(t, g)
	// The first clause is a creature you control: theirs is refused.
	if err := sgAnswerTargets(t, g, theirs); !errors.Is(err, game.ErrIllegalTarget) {
		t.Fatalf("an opponent's creature as the first target: err = %v, want ErrIllegalTarget", err)
	}
	pickCard(t, g, me.ID, mine)
	pickCard(t, g, me.ID, theirs)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(theirs) {
		t.Error("the 2/2 survived a fight with a 3/3")
	}
	if c := layeredCard(t, g, mine); c.DamageMarked != 2 {
		t.Errorf("the 3/3 took %d damage, want 2", c.DamageMarked)
	}
}

func TestTrekB19AmokTimeChapterOneWithNoOpponentsCreatureFightsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "Vulcan", "Creature — Vulcan", 3, 3)
	trekB19AmokTime(g, me.ID, 0)
	advanceToMain(t, g)
	passPriorityAroundTable(t, g)
	pickCard(t, g, me.ID, mine)
	if p := latestPickTarget(g, me.ID); p != nil {
		if err := g.ResolvePickTargets(p.ID, me.ID, nil); err != nil {
			t.Fatalf("declining the second target: %v", err)
		}
	}
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, mine); c.DamageMarked != 0 {
		t.Errorf("with no second target nothing fights, but the Vulcan took %d damage", c.DamageMarked)
	}
}

func TestTrekB19AmokTimeChapterTwoGainsThreeLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	trekB19AmokTime(g, me.ID, 1)
	life := me.Life
	advanceToMain(t, g)
	passPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

func TestTrekB19AmokTimeChapterThreeReturnsACreatureCardAndIsSacrificed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	saga := trekB19AmokTime(g, me.ID, 2)
	sorcery := batch01GraveyardCard(me, "My Spell", "Sorcery")
	theirs := batch01GraveyardCard(opp, "Their Bear", "Creature — Bear")
	creature := batch01GraveyardCard(me, "Dead Bear", "Creature — Bear")
	advanceToMain(t, g)
	passPriorityAroundTable(t, g)
	for _, bad := range []uuid.UUID{sorcery, theirs} {
		if err := sgAnswerTargets(t, g, bad); !errors.Is(err, game.ErrIllegalTarget) {
			t.Fatalf("%s: err = %v, want ErrIllegalTarget", cardNameFor(g, bad), err)
		}
	}
	if err := sgAnswerTargets(t, g, creature); err != nil {
		t.Fatalf("my creature card: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(creature) {
		t.Error("the creature card did not return to hand")
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after chapter III")
	}
}

// --- Astute Terraformer -------------------------------------------------------

func TestTrekB19TerraformerCostsLessPerLandNameOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	vulcan := b27Push(g, me.ID, "Astute Terraformer", "Creature — Vulcan Scientist", trekB19TerraformOracle, "{1}{G}", 2, 2, "G")
	// Three names (two Forests are one): {6} costs {3}.
	trekB19Lands(g, me.ID, "Forest", "Forest", "Island", "Plains")
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, vulcan, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated for {2} with three differently named lands")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, vulcan, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate for {3} with three differently named lands: %v", err)
	}
	if me.ManaPool.Total() != 0 {
		t.Errorf("pool %d after paying {3}, want 0", me.ManaPool.Total())
	}
	passPriorityAroundTable(t, g)
	if p, tt := effectivePTOf(t, g, vulcan); p != 4 || tt != 4 {
		t.Errorf("Terraformer is %d/%d, want 4/4", p, tt)
	}
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, vulcan, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated a second time in one turn")
	}
}

// --- Branching Timeline -------------------------------------------------------

func TestTrekB19BranchingTimelinePutsPermanentsTappedAndTheRestOnTheBottom(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// Two names: X = 2.
	trekB19Lands(g, me.ID, "Forest", "Forest", "Island")
	deep := putOnTopForTest(me, "Deep Bear", "Creature — Bear")
	sorcery := putOnTopForTest(me, "Some Sorcery", "Sorcery")
	bear := putOnTopForTest(me, "Top Bear", "Creature — Bear")
	castCatalogSpell(t, g, "Branching Timeline", "Sorcery", trekB19BranchingOracle, nil)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the revealed creature card did not enter")
	}
	if !layeredCard(t, g, bear).Tapped {
		t.Error("the permanent card entered untapped")
	}
	if g.Battlefield.Contains(deep) {
		t.Error("a card below the top X was put onto the battlefield")
	}
	if !me.Library.Contains(sorcery) {
		t.Fatal("the revealed sorcery left the library")
	}
	if bottom, err := me.Library.Bottom(); err != nil || bottom.InstanceID != sorcery {
		t.Errorf("the revealed sorcery is not on the bottom (bottom %+v, %v)", bottom.Name, err)
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != deep {
		t.Errorf("top is %q, want the unrevealed Deep Bear", top.Name)
	}
}

func TestTrekB19BranchingTimelineWithNoLandsRevealsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := putOnTopForTest(me, "Top Bear", "Creature — Bear")
	castCatalogSpell(t, g, "Branching Timeline", "Sorcery", trekB19BranchingOracle, nil)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("X = 0 put a card onto the battlefield")
	}
	if top, err := me.Library.Top(); err != nil || top.InstanceID != bear {
		t.Error("X = 0 moved the top card")
	}
}

// --- Captain Jonathan Archer --------------------------------------------------

func TestTrekB19ArcherMakesPorthos(t *testing.T) {
	g := newCatalogGame(t)
	castHoldCreature(t, g, "Captain Jonathan Archer", "Legendary Creature — Human Officer", trekB19ArcherOracle, 3, 3)
	passPriorityAroundTable(t, g)
	dogs := trekB19Named(t, g, "Porthos")
	if len(dogs) != 1 {
		t.Fatalf("%d Porthos tokens, want 1", len(dogs))
	}
	c := dogs[0]
	if !c.IsToken() || !c.IsLegendary() || !c.HasSubtype("Dog") || c.CurrentPower() != 1 || c.CurrentToughness() != 1 ||
		len(c.Colors) != 1 || c.Colors[0] != "G" {
		t.Errorf("Porthos = %+v, want a legendary 1/1 green Dog token", c)
	}
}

func TestTrekB19ArcherPumpsByCreatureTypesAmongNonBorg(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	archer := b27Push(g, me.ID, "Captain Jonathan Archer", "Legendary Creature — Human Officer", trekB19ArcherOracle, "{2}{G}", 3, 3, "G")
	dog := b12Creature(g, me.ID, "Dog", "Creature — Dog", 1, 1)
	b12Creature(g, me.ID, "Drone", "Creature — Borg Elf", 1, 1)
	b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 3, 3)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: dog}}
	// Activate only as a sorcery: the draw step is not a main phase.
	if err := hobB4Activate(t, g, me.ID, archer, 0, "{G}", targets); err == nil {
		t.Fatal("activated outside a main phase")
	}
	advanceToMain(t, g)
	if err := hobB4Activate(t, g, me.ID, archer, 0, "{G}", []game.TargetRef{{Kind: game.TargetCard, ID: theirs}}); err == nil {
		t.Fatal("targeted an opponent's creature")
	}
	if err := hobB4Activate(t, g, me.ID, archer, 0, "", targets); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	// Human, Officer, Dog: the Borg Elf adds nothing, the Giant isn't yours.
	if p, tt := effectivePTOf(t, g, dog); p != 4 || tt != 4 {
		t.Errorf("the Dog is %d/%d, want 4/4", p, tt)
	}
	if !layeredCard(t, g, archer).Tapped {
		t.Error("the {T} cost did not tap Archer")
	}
}

// --- Captain Kathryn Janeway --------------------------------------------------

func TestTrekB19JanewayExploresHerselfAndGivesAnExtraLandDrop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	land := putOnTopForTest(me, "Forest", "Basic Land — Forest")
	castHoldCreature(t, g, "Captain Kathryn Janeway", "Legendary Creature — Human Officer", trekB19JanewayOracle, 2, 3)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(land) {
		t.Error("Janeway's own entry did not explore: the land is not in hand")
	}
	if got := g.LandDropsRemainingFor(me.ID); got != 2 {
		t.Errorf("land drops with Janeway out = %d, want 2", got)
	}
}

func TestTrekB19JanewayMakesAnotherCreatureExplore(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "Captain Kathryn Janeway", "Legendary Creature — Human Officer", trekB19JanewayOracle, "{3}{G}", 2, 3, "G")
	top := putOnTopForTest(me, "Some Sorcery", "Sorcery")
	bear := tdCastCreature(t, g, "Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the explorer has %d +1/+1 counters, want 1", got)
	}
	answerExploreQuestion(t, g, me.ID, true)
	if !me.Graveyard.Contains(top) {
		t.Error("the revealed nonland was not put into the graveyard")
	}
	// An opponent's creature entering is not yours.
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	passPriorityAroundTable(t, g)
	if g.Stack.Size() != 0 || len(g.PendingChoices) != 0 {
		t.Error("an opponent's creature triggered Janeway")
	}
}

// --- Captain Pike, Fullness of Time -------------------------------------------

func TestTrekB19PikeIsSizedByLandNamesAndPlaysLandsFromTheTop(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	pike := b27Push(g, me.ID, "Captain Pike, Fullness of Time", "Legendary Creature — Human Officer Scout", trekB19PikeOracle, "{1}{G}{G}", 0, 0, "G")
	trekB19Lands(g, me.ID, "Forest", "Forest", "Island")
	trekB19Lands(g, opp.ID, "Plains", "Swamp")
	if p, tt := effectivePTOf(t, g, pike); p != 2 || tt != 2 {
		t.Errorf("Pike is %d/%d, want 2/2 (two names you control)", p, tt)
	}
	if got := g.LibraryTopVisibilityForEffect(me.ID); got != game.LibraryTopOwner {
		t.Errorf("top-of-library visibility %v, want the owner's look", got)
	}
	toMain(t, g)
	top := seedTopOfLibrary(me, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "library"}); err != nil {
		t.Fatalf("play the land off the top: %v", err)
	}
	if !g.Battlefield.Contains(top) {
		t.Fatal("the land never reached the battlefield")
	}
	if p, _ := effectivePTOf(t, g, pike); p != 3 {
		t.Errorf("Pike's power %d after a third name, want 3", p)
	}
}

func TestTrekB19PikeDoesNotOpenASpellOnTop(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Captain Pike, Fullness of Time", "Legendary Creature — Human Officer Scout", trekB19PikeOracle, "{1}{G}{G}", 0, 0, "G")
	trekB19Lands(g, me.ID, "Forest")
	toMain(t, g)
	top := seedTopOfLibrary(me, game.Card{Name: "Top Sorcery", TypeLine: "Sorcery", ManaCost: "{G}"})
	floatMana(t, g, me, "{G}")
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{FromZone: "library", Strict: true}); err == nil {
		t.Error("a sorcery on top was castable under a lands-only permission")
	}
}

// --- Cold-Blooded Crew --------------------------------------------------------

func TestTrekB19CrewPutsTwoCountersOnEntry(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Cold-Blooded Crew", "Creature — Gorn Warrior", trekB19CrewOracle, 4, 4)
	answerPickTarget(t, g, bear)
	passPriorityAroundTable(t, g)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 2 {
		t.Errorf("the Bear has %d +1/+1 counters, want 2", got)
	}
}

func TestTrekB19CrewCantBeBlockedByPowerTwoOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	crew := b12Push(g, me.ID, "Cold-Blooded Crew", "Creature — Gorn Warrior", trekB19CrewOracle, 4, 4)
	small := b12Creature(g, opp.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Hill Giant", "Creature — Giant", 3, 3)
	brAttack(t, g, crew)
	brRefusal(t, g.DeclareBlocker(small, crew), game.BlockReasonCantBeBlockedBy)
	if err := g.DeclareBlocker(big, crew); err != nil {
		t.Fatalf("a power 3 creature blocks: %v", err)
	}
}

func TestTrekB19CrewBasicLandcyclingFindsABasicLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	id := pushCatalogHandCard(me, "Cold-Blooded Crew", "Creature — Gorn Warrior", trekB19CrewOracle)
	ids := seedSearchLibrary(me,
		game.Card{Name: "Bolt", TypeLine: "Instant"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
	)
	if err := hobB4Activate(t, g, me.ID, id, 0, "{C}", nil); err == nil {
		t.Fatal("landcycled for {1}")
	}
	if err := hobB4Activate(t, g, me.ID, id, 0, "{C}", nil); err != nil {
		t.Fatalf("landcycling for {2}: %v", err)
	}
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ids[1])
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(ids[1]) || !me.Graveyard.Contains(id) {
		t.Errorf("Forest in hand %v, Crew in graveyard %v", me.Hand.Contains(ids[1]), me.Graveyard.Contains(id))
	}
}

// --- Common Goal --------------------------------------------------------------

func TestTrekB19CommonGoalTwoCreaturesEachDealTheirPower(t *testing.T) {
	g, me, opp := trekB8Main(t)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 3, 3)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)
	trekB8Float(t, g, me, "{1}{G}")
	castFromHandForTest(t, g, me, "Common Goal", "Sorcery", "{1}{G}", trekB19CommonGoal, game.CastSpellParams{
		Targets: []game.TargetRef{trekB19Ref(a, 0), trekB19Ref(b, 0), trekB19Ref(ogre, 1)}, Strict: true,
	})
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ogre) {
		t.Error("2 + 3 damage did not destroy the 5/5")
	}
	for _, id := range []uuid.UUID{a, b} {
		if c := layeredCard(t, g, id); c.DamageMarked != 0 {
			t.Errorf("%s took %d damage; Common Goal is one-sided", c.Name, c.DamageMarked)
		}
	}
}

func TestTrekB19CommonGoalOneCreatureAndRefusedTargets(t *testing.T) {
	g, me, opp := trekB8Main(t)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	ogre := b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 5, 5)
	trekB8Float(t, g, me, "{1}{G}")
	bad := []game.TargetRef{trekB19Ref(theirs, 0), trekB19Ref(ogre, 1)}
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Common Goal", TypeLine: "Sorcery", ManaCost: "{1}{G}",
			OracleID: trekB19CommonGoal, Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: bad, Strict: true}); err == nil {
		t.Fatal("an opponent's creature was accepted as a damage source")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{trekB19Ref(a, 0), trekB19Ref(a, 1)}, Strict: true}); err == nil {
		t.Fatal("your own creature was accepted as the recipient")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: []game.TargetRef{trekB19Ref(a, 0), trekB19Ref(ogre, 1)}, Strict: true}); err != nil {
		t.Fatalf("cast with one source: %v", err)
	}
	passPriorityAroundTable(t, g)
	if c := layeredCard(t, g, ogre); c.DamageMarked != 2 {
		t.Errorf("the Ogre took %d damage, want 2", c.DamageMarked)
	}
}

// --- Culber and Stamets, Star-Crossed -----------------------------------------

func TestTrekB19CulberAndStametsCounterMode(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	castHoldCreature(t, g, "Culber and Stamets, Star-Crossed", "Legendary Creature — Human Doctor Artificer", trekB19CulberOracle, 3, 4)
	trekB6ModeThenTarget(t, g, me.ID, 0, bear)
	if got := countersOn(g, bear, game.CounterPlusOne); got != 1 {
		t.Errorf("the Bear has %d +1/+1 counters, want 1", got)
	}
}

func TestTrekB19CulberAndStametsAddsTwoManaOfOneColor(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castHoldCreature(t, g, "Culber and Stamets, Star-Crossed", "Legendary Creature — Human Doctor Artificer", trekB19CulberOracle, 3, 4)
	trekB6ModeThenTarget(t, g, me.ID, 1, uuid.Nil)
	ashlingPickColor(t, g, me.ID, "U")
	passPriorityAroundTable(t, g)
	if got := me.ManaPool.Total(); got != 2 {
		t.Fatalf("pool holds %d mana, want 2", got)
	}
	for _, tok := range me.ManaPool {
		if tok.Color != "U" {
			t.Errorf("pool token %+v, want blue", tok)
		}
	}
}

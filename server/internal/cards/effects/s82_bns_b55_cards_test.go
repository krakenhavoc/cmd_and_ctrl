package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_bns_b55_cards_test.go — Foundations Jumpstart, S82 BNS batch 55
// (#5929): Overflowing Insight; Overwhelm; Pigment Storm; Pyrophobia;
// Renegade Tactics; Roil Eruption; Saproling Migration; Shoulder to
// Shoulder; Silent Departure; Triplicate Spirits. (Overcome was built in
// batch 22.)

const (
	bnsB55Insight    = "ff03a063-9634-4746-a208-9dbcde5c6bdb"
	bnsB55Overwhelm  = "caeaad5e-2d68-481c-9a35-c2c3b345963d"
	bnsB55Pigment    = "89abdf4c-ba9d-4602-ae50-cf79c68ea08d"
	bnsB55Pyrophobia = "fa70903d-3c8b-49f7-8817-f708b5ba250f"
	bnsB55Renegade   = "7fbcd256-c132-406f-a490-df9709835504"
	bnsB55Roil       = "6798163a-864f-4844-96b1-77585a7e7ab4"
	bnsB55Saproling  = "bbbb80bd-76fa-4155-929b-4a3565d1cb35"
	bnsB55Shoulder   = "4ffa7cc8-4774-430d-b898-5bd26366e124"
	bnsB55Departure  = "c41641f0-1abf-4776-9dec-1882f2d9badd"
	bnsB55Triplicate = "48090c64-f41a-447b-ad7a-54e3194f759b"
)

func TestBnsB55CardsAreRegisteredFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		bnsB55Insight: "Overflowing Insight", bnsB55Overwhelm: "Overwhelm", bnsB55Pigment: "Pigment Storm",
		bnsB55Pyrophobia: "Pyrophobia", bnsB55Renegade: "Renegade Tactics", bnsB55Roil: "Roil Eruption",
		bnsB55Saproling: "Saproling Migration", bnsB55Shoulder: "Shoulder to Shoulder",
		bnsB55Departure: "Silent Departure", bnsB55Triplicate: "Triplicate Spirits",
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
	for _, oracle := range []string{bnsB55Overwhelm, bnsB55Triplicate} {
		if spec, _ := Lookup(oracle); spec.TapCost == nil {
			t.Errorf("%s: convoke is a TapCost", spec.Name)
		}
	}
}

func TestBnsB55OverflowingInsightTargetPlayerDrawsSeven(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(opp, "a", "b", "c", "d", "e", "f", "g", "h", "i")
	before, mine := handSize(opp), handSize(me)
	castCatalogSpell(t, g, "Overflowing Insight", "Sorcery", bnsB55Insight, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := handSize(opp) - before; got != 7 {
		t.Errorf("target player drew %d, want 7", got)
	}
	if handSize(me) != mine {
		t.Error("the caster drew")
	}
}

func TestBnsB55OverwhelmPumpsOnlyYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Overwhelm", "Sorcery", bnsB55Overwhelm, nil)
	passPriorityAroundTable(t, g)
	if p, tough := ptOf(t, g, mine); p != 5 || tough != 5 {
		t.Errorf("my creature is %d/%d, want 5/5", p, tough)
	}
	if p, tough := ptOf(t, g, theirs); p != 2 || tough != 2 {
		t.Errorf("their creature is %d/%d, want 2/2", p, tough)
	}
	later := b12Creature(g, me.ID, "Later", "Creature — Bear", 3, 3)
	if p, _ := ptOf(t, g, later); p != 3 {
		t.Errorf("a creature that arrived afterwards has power %d, want 3", p)
	}
}

func TestBnsB55PigmentStormSplitsLethalAndExcess(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	life := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Pigment Storm", "Sorcery", bnsB55Pigment, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, theirs) != nil {
		t.Error("the 2/2 survived five damage")
	}
	if got := life - lifeOf(g, opp.ID); got != 3 {
		t.Errorf("the controller lost %d, want the 3 excess", got)
	}
}

func TestBnsB55PigmentStormBigCreatureTakesItAll(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 6, 6)
	life := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Pigment Storm", "Sorcery", bnsB55Pigment, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, theirs); got != 5 {
		t.Errorf("damage %d, want 5", got)
	}
	if lifeOf(g, opp.ID) != life {
		t.Error("the controller took damage with no excess")
	}
}

func TestBnsB55PigmentStormCountsDamageAlreadyMarked(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 4, 4)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == theirs {
				g.Battlefield.Cards[i].DamageMarked = 3
			}
		}
	})
	life := lifeOf(g, opp.ID)
	castCatalogSpell(t, g, "Pigment Storm", "Sorcery", bnsB55Pigment, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	if got := life - lifeOf(g, opp.ID); got != 4 {
		t.Errorf("the controller lost %d, want 4 (1 lethal, 4 excess)", got)
	}
}

func TestBnsB55PyrophobiaDamagesAndStopsCowardsBlocking(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := pushSizedCreature(g, me.ID, "My Attacker", 3, 3)
	target := b12Creature(g, opp.ID, "Big", "Creature — Ogre", 5, 5)
	coward := b12Creature(g, opp.ID, "Skittish", "Creature — Human Coward", 1, 1)
	castCatalogSpell(t, g, "Pyrophobia", "Sorcery", bnsB55Pyrophobia, b16TargetCard(target))
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, target); got != 3 {
		t.Errorf("damage %d, want 3", got)
	}
	late := b12Creature(g, opp.ID, "Late Coward", "Creature — Coward", 1, 1)
	declareAttack(t, g, opp.ID, attacker)
	advanceTo(t, g, game.StepDeclareBlockers)
	for _, c := range []uuid.UUID{coward, late} {
		if err := g.DeclareBlocker(c, attacker); !errors.Is(err, game.ErrIllegalBlock) {
			t.Errorf("a Coward blocked: err = %v, want ErrIllegalBlock", err)
		}
	}
	if err := g.DeclareBlocker(target, attacker); err != nil {
		t.Errorf("a non-Coward may still block: %v", err)
	}
}

func TestBnsB55RenegadeTacticsStopsBlockAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	seedLibrary(me, "a", "b")
	theirs := pushVanillaCreature(g, opp.ID, "Theirs", 2, 2)
	other := pushVanillaCreature(g, opp.ID, "Other", 2, 2)
	before := handSize(me)
	castCatalogSpell(t, g, "Renegade Tactics", "Sorcery", bnsB55Renegade, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	if !midB03CantBlock(t, g, theirs) || midB03CantBlock(t, g, other) {
		t.Error("only the target should be unable to block")
	}
	if got := handSize(me) - before; got != 1 {
		t.Errorf("hand changed by %d, want 1 (spell added then cast, one drawn)", got)
	}
}

func TestBnsB55RoilEruptionThreeOrKickedFive(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	life := lifeOf(g, opp.ID)
	dmuB09Cast(t, g, "Roil Eruption", "Sorcery", bnsB55Roil, false, []game.TargetRef{playerTarget(opp.ID)})
	passPriorityAroundTable(t, g)
	if got := life - lifeOf(g, opp.ID); got != 3 {
		t.Errorf("unkicked dealt %d, want 3", got)
	}
	g2 := newCatalogGame(t)
	opp2 := g2.Seats[1]
	theirs := b12Creature(g2, opp2.ID, "Theirs", "Creature — Ogre", 6, 6)
	dmuB09Cast(t, g2, "Roil Eruption", "Sorcery", bnsB55Roil, true, b16TargetCard(theirs))
	passPriorityAroundTable(t, g2)
	if got := damageOn(t, g2, theirs); got != 5 {
		t.Errorf("kicked dealt %d, want 5", got)
	}
}

func TestBnsB55SaprolingMigrationTwoOrKickedFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dmuB09Cast(t, g, "Saproling Migration", "Sorcery", bnsB55Saproling, false, nil)
	passPriorityAroundTable(t, g)
	if n := tokensNamedControlledBy(g, "Saproling", me.ID); n != 2 {
		t.Errorf("unkicked made %d Saprolings, want 2", n)
	}
	g2 := newCatalogGame(t)
	dmuB09Cast(t, g2, "Saproling Migration", "Sorcery", bnsB55Saproling, true, nil)
	passPriorityAroundTable(t, g2)
	if n := tokensNamedControlledBy(g2, "Saproling", g2.Seats[0].ID); n != 4 {
		t.Errorf("kicked made %d Saprolings, want 4", n)
	}
}

func TestBnsB55ShoulderToShoulderSupportsTwoAndDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "a", "b")
	a := pushVanillaCreature(g, me.ID, "A", 2, 2)
	b := pushVanillaCreature(g, me.ID, "B", 2, 2)
	c := pushVanillaCreature(g, me.ID, "C", 2, 2)
	if err := castCatalogSpellErrStrict(t, g, "Shoulder to Shoulder", "Sorcery", bnsB55Shoulder,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}, {Kind: game.TargetCard, ID: c}}); err == nil {
		t.Error("support 2 took three targets")
	}
	g2 := newCatalogGame(t)
	me2 := g2.Seats[0]
	seedLibrary(me2, "a", "b")
	a = pushVanillaCreature(g2, me2.ID, "A", 2, 2)
	b = pushVanillaCreature(g2, me2.ID, "B", 2, 2)
	c = pushVanillaCreature(g2, me2.ID, "C", 2, 2)
	before := handSize(me2)
	castCatalogSpell(t, g2, "Shoulder to Shoulder", "Sorcery", bnsB55Shoulder,
		[]game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}})
	passPriorityAroundTable(t, g2)
	if plusOneCounters(g2, a) != 1 || plusOneCounters(g2, b) != 1 || plusOneCounters(g2, c) != 0 {
		t.Errorf("counters %d/%d/%d, want 1/1/0", plusOneCounters(g2, a), plusOneCounters(g2, b), plusOneCounters(g2, c))
	}
	if got := handSize(me2) - before; got != 1 {
		t.Errorf("hand changed by %d, want 1 (spell added then cast, one drawn)", got)
	}
}

func TestBnsB55ShoulderToShoulderWithNoTargetStillDraws(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	seedLibrary(me, "a", "b")
	before := handSize(me)
	castCatalogSpell(t, g, "Shoulder to Shoulder", "Sorcery", bnsB55Shoulder, nil)
	passPriorityAroundTable(t, g)
	if got := handSize(me) - before; got != 1 {
		t.Errorf("hand changed by %d, want 1 (spell added then cast, one drawn)", got)
	}
}

func TestBnsB55SilentDepartureBouncesAndFlashesBack(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[1]
	theirs := b12Creature(g, opp.ID, "Theirs", "Creature — Ogre", 2, 2)
	castCatalogSpell(t, g, "Silent Departure", "Sorcery", bnsB55Departure, b16TargetCard(theirs))
	passPriorityAroundTable(t, g)
	if findBattlefieldCardByID(g, theirs) != nil || !inHandOf(opp, theirs) {
		t.Error("the creature was not returned to its owner's hand")
	}
	g2 := newCatalogGame(t)
	foe := b12Creature(g2, g2.Seats[1].ID, "Theirs", "Creature — Ogre", 2, 2)
	id := midB03Flash(t, g2, "Silent Departure", "Sorcery", bnsB55Departure, b16TargetCard(foe))
	passPriorityAroundTable(t, g2)
	if findBattlefieldCardByID(g2, foe) != nil {
		t.Error("the flashback did not bounce the creature")
	}
	midB03Exiled(t, g2, id, "Silent Departure")
}

func TestBnsB55TriplicateSpiritsMakesThreeFlyingSpirits(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	castCatalogSpell(t, g, "Triplicate Spirits", "Sorcery", bnsB55Triplicate, nil)
	passPriorityAroundTable(t, g)
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsToken() && c.HasSubtype("Spirit") && game.HasKeyword(&c, "flying") && c.CurrentPower() == 1 && c.CurrentToughness() == 1 {
			n++
		}
	}
	if n != 3 {
		t.Errorf("%d 1/1 flying Spirits, want 3", n)
	}
}

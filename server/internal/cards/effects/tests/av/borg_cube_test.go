package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const borgCubeOracle = "a59b854c-0e04-4581-b9d3-1516263d7ace"

func TestBorgCubeIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(borgCubeOracle)
	if !ok {
		t.Fatal("Borg Cube is not registered")
	}
	if spec.Name != "Borg Cube" || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %s, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
}

// borgCubePush seeds a Borg Cube on the battlefield with `charge` charge
// counters.
func borgCubePush(g *game.Game, owner uuid.UUID, charge int) uuid.UUID {
	id := effectstest.B12Push(g, owner, "Borg Cube", "Artifact — Spacecraft", borgCubeOracle, 6, 6)
	if charge > 0 {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Counters = map[string]int{game.CounterCharge: charge}
			}
		}
		g.BumpLayerVersionForTest()
	}
	return id
}

// borgCubeRef is the Cube's current object, as an exile link names it.
func borgCubeRef(t *testing.T, g *game.Game, cube uuid.UUID) game.PermissionCardRef {
	t.Helper()
	c, ok := effectstest.BattlefieldCard(g, cube)
	if !ok {
		t.Fatal("the Cube is not on the battlefield")
	}
	return game.PermissionCardRef{ID: c.InstanceID, Epoch: c.ObjectEpoch}
}

func borgCubeExiledCard(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Exile.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}

// borgCubeOnBattlefield is the permanent named `name`: a card put onto
// the battlefield from exile is a new object with a new ID (CR 400.7).
func borgCubeOnBattlefield(t *testing.T, g *game.Game, name string) uuid.UUID {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c.InstanceID
		}
	}
	t.Fatalf("%s is not on the battlefield", name)
	return uuid.Nil
}

// borgCubeCheckAssimilated asserts `id` is a Borg artifact creature under
// `me` with one +1/+1 counter and no other creature type.
func borgCubeCheckAssimilated(t *testing.T, g *game.Game, me, id uuid.UUID, gone ...string) {
	t.Helper()
	got := effectstest.FindBattlefieldCardForTest(g, id)
	if got == nil {
		t.Fatal("the exiled creature card was not assimilated")
	}
	if got.Controller != me {
		t.Errorf("it entered under %v, want the Cube's controller", got.Controller)
	}
	if n := effectstest.PlusOneCounters(g, id); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1", n)
	}
	v := effectstest.GaLayered(t, g, id)
	if !v.IsArtifact() || !v.IsCreature() || !v.HasSubtype("Borg") {
		t.Errorf("it is %q, want a Borg artifact creature", v.TypeLine)
	}
	for _, s := range gone {
		if v.HasSubtype(s) {
			t.Errorf("it is %q, still a %s", v.TypeLine, s)
		}
	}
}

func TestBorgCubeCastWithAffinityExilesATargetLinkedToIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	// Without artifact creatures, {3}{W}{B} does not pay for {5}{W}{B}.
	if _, err := effectstest.Msc10Cast(t, g, me, "Borg Cube", "Artifact — Spacecraft", "{5}{W}{B}", borgCubeOracle, "{C}{C}{C}{W}{B}", game.CastSpellParams{}); err == nil {
		t.Fatal("cast for {3}{W}{B} with no artifact creatures was accepted")
	}
	g.Seats[0].ManaPool = game.ManaPool{}
	effectstest.B12Creature(g, me.ID, "Drone A", "Artifact Creature — Construct", 1, 1)
	effectstest.B12Creature(g, me.ID, "Drone B", "Artifact Creature — Construct", 1, 1)
	effectstest.B12Permanent(g, me.ID, "Mind Stone", "Artifact")
	effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Knight", "Creature — Human Knight", 3, 3)
	cube, err := effectstest.Msc10Cast(t, g, me, "Borg Cube", "Artifact — Spacecraft", "{5}{W}{B}", borgCubeOracle, "{C}{C}{C}{W}{B}", game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast for {3}{W}{B} with two artifact creatures: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	got, ok := borgCubeExiledCard(g, theirs)
	if !ok {
		t.Fatal("with nothing exiled with it, the enter trigger did not exile its target")
	}
	if want := borgCubeRef(t, g, cube); got.ExiledWith != want {
		t.Errorf("exiled with %v, want the Cube %v", got.ExiledWith, want)
	}
}

func TestBorgCubeAttackAssimilatesTheCardExiledWithItInsteadOfExiling(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cube := borgCubePush(g, me.ID, 12)
	knight := effectstest.ElPushExiled(g, opp.ID, borgCubeRef(t, g, cube), "Their Knight", "Enchantment Creature — Human Knight", "")
	other := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, cube)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, other)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := borgCubeExiledCard(g, knight); ok {
		t.Fatal("the card exiled with it is still in exile")
	}
	knight = borgCubeOnBattlefield(t, g, "Their Knight")
	borgCubeCheckAssimilated(t, g, me.ID, knight, "Human", "Knight")
	if !effectstest.GaLayered(t, g, knight).IsEnchantment() {
		t.Error("it lost its enchantment type (CR 205.1b)")
	}
	if !effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, knight), "flying") {
		t.Error("the assimilated artifact creature lacks flying from the 12+ line")
	}
	if effectstest.FindBattlefieldCardForTest(g, other) == nil {
		t.Error("it assimilated, yet the target was exiled too")
	}
}

func TestBorgCubeDoesNotAssimilateACardExiledByAnythingElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cube := borgCubePush(g, me.ID, 12)
	stray := effectstest.ElPushExiled(g, opp.ID, game.PermissionCardRef{}, "Stray Knight", "Creature — Knight", "")
	// A card linked to an earlier object of this card (CR 400.7).
	old := effectstest.ElPushExiled(g, opp.ID, game.PermissionCardRef{ID: cube, Epoch: borgCubeRef(t, g, cube).Epoch + 1}, "Old Knight", "Creature — Knight", "")
	noncreature := effectstest.ElPushExiled(g, opp.ID, borgCubeRef(t, g, cube), "Old Vehicle", "Artifact — Vehicle", "")
	target := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, cube)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, target)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{stray, old, noncreature} {
		if _, ok := borgCubeExiledCard(g, id); !ok {
			t.Errorf("%v left exile", id)
		}
	}
	got, ok := borgCubeExiledCard(g, target)
	if !ok {
		t.Fatal("with no creature card exiled with it, the target was not exiled")
	}
	if got.ExiledWith != borgCubeRef(t, g, cube) {
		t.Errorf("the target is exiled with %v, want the Cube", got.ExiledWith)
	}
}

func TestBorgCubeAsksWhichCardWhenSeveralAreExiledWithIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	cube := borgCubePush(g, me.ID, 12)
	ref := borgCubeRef(t, g, cube)
	first := effectstest.ElPushExiled(g, opp.ID, ref, "First Elf", "Creature — Elf", "")
	second := effectstest.ElPushExiled(g, opp.ID, ref, "Second Goblin", "Creature — Goblin", "")
	effectstest.DeclareAttack(t, g, opp.ID, cube)
	effectstest.B04WaitForPick(t, g, me.ID)
	effectstest.PickCard(t, g, me.ID, effectstest.B12Creature(g, opp.ID, "Spare", "Creature — Bear", 1, 1))
	for i := 0; i < 8 && effectstest.LatestChooseCardsFor(g, me.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	effectstest.EtChooseCards(t, g, me.ID, 2, second)
	effectstest.PassPriorityAroundTable(t, g)
	borgCubeCheckAssimilated(t, g, me.ID, borgCubeOnBattlefield(t, g, "Second Goblin"), "Goblin")
	if _, ok := borgCubeExiledCard(g, first); !ok {
		t.Error("the card not chosen left exile")
	}
}

func TestBorgCubeGivesArtifactCreaturesYouControlFlyingAtTwelve(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cube := borgCubePush(g, me.ID, 11)
	drone := effectstest.B12Creature(g, me.ID, "Drone", "Artifact Creature — Construct", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Drone", "Artifact Creature — Construct", 1, 1)
	if effectstest.GaLayered(t, g, cube).IsCreature() {
		t.Error("the Cube is a creature at 11 charge counters")
	}
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, drone), "flying") {
		t.Error("an artifact creature has flying at 11 charge counters")
	}
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(cube, game.CounterCharge, 1) })
	if lc := effectstest.LayeredCard(t, g, cube); !lc.IsCreature() || lc.CurrentPower() != 6 || lc.CurrentToughness() != 6 {
		t.Errorf("at 12 the Cube is %q %d/%d, want a 6/6 artifact creature", lc.TypeLine, lc.CurrentPower(), lc.CurrentToughness())
	}
	for id, want := range map[uuid.UUID]bool{cube: true, drone: true, bear: false, theirs: false} {
		if got := effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, id), "flying"); got != want {
			t.Errorf("%s flying %v, want %v", effectstest.FindBattlefieldCardForTest(g, id).Name, got, want)
		}
	}
}

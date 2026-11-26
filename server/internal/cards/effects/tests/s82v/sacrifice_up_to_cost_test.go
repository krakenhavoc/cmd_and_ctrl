package testss82v

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// #6121: a sacrifice clause with a floor of zero on an activated ability,
// beside a {T} that keeps it from being free. CR 602.2b applies CR 601.2b
// to an activation: the activator announces how many, zero included.

const (
	upToOracle   = "test-6121-sacrifice-up-to-three"
	anyNumOracle = "test-6121-sacrifice-self-and-any-number"
)

// registerUpToThree is Baba Lysaga's shape: "{T}, Sacrifice up to three
// permanents: If there were three or more card types among the sacrificed
// permanents, you gain 3 life." The types are read off the sacrificed
// objects as they last existed on the battlefield (CR 608.2h).
func registerUpToThree(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: upToOracle,
		Name:     upToOracle,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}, Sacrifice up to three permanents: gain 3 life if three card types",
			Cost:  effects.Plus(effects.TapCost(), effects.SacrificeUpTo(3, "up to three permanents", effects.Permanent())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				types := map[string]bool{}
				for _, p := range ctx.SacrificedPermanents() {
					for _, ty := range p.Characteristic.Types {
						types[ty] = true
					}
				}
				if len(types) < 3 {
					return nil
				}
				return effects.GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
			},
		}},
	})
}

// registerSelfAndAnyNumber is Sword of the Ages' cost: "{T}, Sacrifice this
// artifact and any number of creatures you control: you gain life equal to
// the total power of the creatures sacrificed this way."
func registerSelfAndAnyNumber(t *testing.T) {
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID: anyNumOracle,
		Name:     anyNumOracle,
		Activated: []effects.ActivatedAbility{{
			Label: "{T}, Sacrifice this artifact and any number of creatures you control: gain life",
			Cost:  effects.Plus(effects.TapCost(), effects.SacrificeThis(), effects.SacrificeAnyNumberOf("any number of creatures you control", effects.Creature())),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				total := 0
				for _, p := range ctx.SacrificedPermanents() {
					if effects.HasFold(p.Characteristic.Types, "creature") {
						total += p.Power
					}
				}
				return effects.GainLife{Player: item.Controller, Amount: total}.Apply(ctx)
			},
		}},
	})
}

func upToPermanent(g *game.Game, owner uuid.UUID, name, typeLine string, power int) uuid.UUID {
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine,
		Power: power, Toughness: 1, Owner: owner, Controller: owner,
	})
	return id
}

func TestSacrificeUpToBoundsAreTheRange(t *testing.T) {
	spec := effects.SacrificeUpTo(3, "up to three permanents").SacrificeOther
	lo, hi := game.SacrificeCostBounds(spec, 0)
	if lo != 0 || hi != 3 {
		t.Fatalf("bounds = %d/%d, want 0/3", lo, hi)
	}
	for n := 0; n <= 3; n++ {
		if !game.SacrificeCountLegal(spec, 0, n) {
			t.Errorf("naming %d is refused, want accepted", n)
		}
	}
	if game.SacrificeCountLegal(spec, 0, 4) {
		t.Error("naming 4 is accepted, want refused — the ceiling is three")
	}
	// The printed range Register already accepted ("one or two") reads
	// as a range too, not as its ceiling.
	rng := effects.SacrificeSpec("one or two creatures", effects.Creature()).WithCount(1, 2)
	if lo, hi := game.SacrificeCostBounds(rng, 0); lo != 1 || hi != 2 {
		t.Errorf("one or two: bounds = %d/%d, want 1/2", lo, hi)
	}
}

func TestSacrificeUpToThreeReadsTheCardTypesOfWhatWasPaid(t *testing.T) {
	registerUpToThree(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	src := effectstest.PushCatalogPermanent(g, me.ID, "Witch", "Legendary Creature — Human Warlock", upToOracle, false)
	art := upToPermanent(g, me.ID, "Rock", "Artifact", 0)
	land := upToPermanent(g, me.ID, "Swamp", "Basic Land — Swamp", 0)
	bear := upToPermanent(g, me.ID, "Bear", "Creature — Bear", 2)
	life := me.Life

	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{art, land, bear}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	for _, id := range []uuid.UUID{art, land, bear} {
		if g.Battlefield.Contains(id) {
			t.Error("a named permanent survived the payment")
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Errorf("life = %d, want %d — three card types were sacrificed", me.Life, life+3)
	}
}

func TestSacrificeUpToThreeAcceptsNoneAndRefusesFour(t *testing.T) {
	registerUpToThree(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	src := effectstest.PushCatalogPermanent(g, me.ID, "Witch", "Legendary Creature — Human Warlock", upToOracle, false)
	ids := []uuid.UUID{
		upToPermanent(g, me.ID, "A", "Artifact", 0),
		upToPermanent(g, me.ID, "B", "Artifact", 0),
		upToPermanent(g, me.ID, "C", "Artifact", 0),
		upToPermanent(g, me.ID, "D", "Artifact", 0),
	}
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{SacrificeIDs: ids}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("four sacrifices: err = %v, want ErrInvalidParam", err)
	}
	for _, id := range ids {
		if !g.Battlefield.Contains(id) {
			t.Fatal("a refused activation sacrificed a permanent")
		}
	}
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("zero sacrifices: %v (CR 601.2b: up to three includes none)", err)
	}
	if c, ok := g.LookupCardForEffect(src); !ok || !c.Tapped {
		t.Error("the {T} was not paid")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life = %d, want %d — nothing was sacrificed", me.Life, life)
	}
	// Tapped, it can't be activated again: the {T} is what the zero
	// floor stands on.
	if err := g.ActivateCatalogAbility(me.ID, src, 0, game.ActivateAbilityParams{}); err == nil {
		t.Error("a tapped source activated again")
	}
}

func TestSacrificeSelfAndAnyNumberSumsThePowerPaid(t *testing.T) {
	registerSelfAndAnyNumber(t)
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sword := effectstest.PushCatalogPermanent(g, me.ID, "Sword", "Artifact", anyNumOracle, false)
	a := upToPermanent(g, me.ID, "Ogre", "Creature — Ogre", 3)
	b := upToPermanent(g, me.ID, "Bear", "Creature — Bear", 2)
	rock := upToPermanent(g, me.ID, "Rock", "Artifact", 0)
	life := me.Life

	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{rock}}); err == nil {
		t.Fatal("a noncreature artifact paid \"any number of creatures\"")
	}
	if err := g.ActivateCatalogAbility(me.ID, sword, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{a, b}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(sword) || g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("the source and both creatures are sacrificed as the cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+5 {
		t.Errorf("life = %d, want %d — the creatures' total power", me.Life, life+5)
	}
}

// Without a component beside it, a zero floor is a free ability, and
// Register still refuses it.
func TestRegisterRefusesAZeroFloorWithNothingBesideIt(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("a bare \"sacrifice up to three\" ability registered")
		}
		delete(effects.Registry, "test-6121-free")
	}()
	effects.Register(effects.Spec{OracleID: "test-6121-free", Name: "test-6121-free", Activated: []effects.ActivatedAbility{{
		Label: "x", Cost: effects.SacrificeUpTo(3, "up to three permanents"),
		Effect: func(*game.Game, *game.StackItem) error { return nil },
	}}})
}

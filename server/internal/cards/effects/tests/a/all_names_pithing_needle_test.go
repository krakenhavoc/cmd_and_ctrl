package testsa

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// all_names_pithing_needle_test.go — #6257's acceptance: a creature
// equipped with Spy Kit ("Equipped creature ... has all names of
// nonlegendary creature cards in addition to its name") is a source with
// the name a Pithing Needle chose, when that is a nonlegendary creature
// card's name (CR 201.2a).

const testSpyKitOracle = "test-6257-spy-kit"

func withTestSpyKit(t *testing.T) {
	t.Helper()
	prevNames := game.NonlegendaryCreatureCardName
	t.Cleanup(func() { game.NonlegendaryCreatureCardName = prevNames })
	game.NonlegendaryCreatureCardName = func(name string) bool { return name == "Grizzly Bears" }
	prev := game.CatalogStaticAbilities
	t.Cleanup(func() { game.CatalogStaticAbilities = prev })
	game.CatalogStaticAbilities = func(key string) []game.StaticAbility {
		if key == testSpyKitOracle {
			return []game.StaticAbility{effects.EquippedCreatureIs(game.Layer3Text,
				"Equipped creature has all names of nonlegendary creature cards in addition to its name.",
				func(c *game.Characteristic) { c.AllNonlegendaryCreatureNames = true })}
		}
		return prev(key)
	}
}

func TestPithingNeedleReachesACreatureWithAllNames(t *testing.T) {
	withTestSpyKit(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := g.Seats[0], g.Seats[1]

	needle := effectstest.B21Push(g, them.ID, "Pithing Needle", "Artifact", effectstest.ActPithingNeedleOracle, 0, 0)
	g.WithWriteLock(func() {
		g.QueueCardNameChoiceForEffect(them.ID, needle, "Pithing Needle — choose a card name")
	})
	effectstest.ActNameFor(t, g, them.ID, "Grizzly Bears")

	equipped := effectstest.ActPushKrenko(g, me.ID)
	kit := effectstest.B21Push(g, me.ID, "Spy Kit", "Artifact — Equipment", testSpyKitOracle, 0, 0)
	effectstest.AttachForTest(t, g, kit, equipped)
	if err := g.ActivateCatalogAbility(me.ID, equipped, 0, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrCantActivate) {
		t.Errorf("the equipped creature with Grizzly Bears among its names: err = %v, want ErrCantActivate", err)
	}
	// An unequipped Krenko is not named Grizzly Bears: the Needle does
	// not stop it (whatever else may).
	plain := effectstest.ActPushKrenko(g, them.ID)
	if err := g.ActivateCatalogAbility(them.ID, plain, 0, game.ActivateAbilityParams{}); errors.Is(err, game.ErrCantActivate) {
		t.Errorf("an unequipped Krenko: err = %v, want the Needle not to stop it", err)
	}
}

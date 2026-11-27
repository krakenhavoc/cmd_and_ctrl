package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b588_cards_test.go — Saviors of Kamigawa, S82 older sets, batch 588
// (#5601). Each test plays the real catalog entry.

const (
	oldB588Moth      = "6501b3c7-fd21-4668-b647-db12f0cc23f7"
	oldB588Secret    = "709a8bf0-bd0b-4ee0-9c06-98b7551447b5"
	oldB588Akki      = "91f95d83-3637-4cf8-af6c-56bc4e4b2591"
	oldB588Okina     = "d1c222fe-0ca7-415b-a28a-3b704df53d66"
	oldB588Kitsune   = "d989d8aa-2b2a-4fd2-a8c9-1d182909987c"
	oldB588Nezumi    = "fede7174-261c-4724-b6b0-f2a71f8ff53f"
	oldB588InsectTyp = "Creature — Insect"
)

func TestOldB588EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB588Moth:    "Moonwing Moth",
		oldB588Secret:  "Secretkeeper",
		oldB588Akki:    "Akki Underling",
		oldB588Okina:   "Okina Nightwatch",
		oldB588Kitsune: "Kitsune Loreweaver",
		oldB588Nezumi:  "Deathmask Nezumi",
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

func oldB588HasKeyword(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	return effectstest.P1g2HasKeyword(g, id, kw)
}

// oldB588Setup is a fresh game with the card on the battlefield at its printed
// size and the hands set before anything reads the layers.
func oldB588Setup(t *testing.T, name, typeLine, oracle string, power, toughness, mine, theirs int) (*game.Game, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	id := effectstest.PushCatalogPermanent(g, g.Seats[0].ID, name, typeLine, oracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = power, toughness
			}
		}
	})
	oldB588Hands(g, mine, theirs)
	return g, id
}

// oldB588Hands sets the controller's hand and every opponent's.
func oldB588Hands(g *game.Game, mine, theirs int) {
	for i, p := range g.Seats {
		if i == 0 {
			effectstest.D2dSetHand(p, mine)
		} else {
			effectstest.D2dSetHand(p, theirs)
		}
	}
}

// --- Moonwing Moth -----------------------------------------------------

func TestOldB588MoonwingMothPumpsItsToughnessForWhiteMana(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	moth := effectstest.PushCatalogPermanent(g, me.ID, "Moonwing Moth", oldB588InsectTyp, oldB588Moth, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == moth {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 1
			}
		}
	})
	for i := 0; i < 2; i++ {
		me.ManaPool.AddMana(game.ManaToken{Color: "W"})
		if err := g.ActivateCatalogAbility(me.ID, moth, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, tough := effectstest.EffectivePTOf(t, g, moth); p != 2 || tough != 3 {
		t.Errorf("after two activations Moonwing Moth is %d/%d, want 2/3", p, tough)
	}
	effectstest.AdvanceTo(t, g, game.StepCleanup)
	if _, tough := effectstest.EffectivePTOf(t, g, moth); tough != 1 {
		t.Errorf("toughness after cleanup = %d, want 1", tough)
	}
}

// --- Secretkeeper and Okina Nightwatch ---------------------------------

func TestOldB588SecretkeeperNeedsStrictlyMoreCardsThanEachOpponent(t *testing.T) {
	for _, tc := range []struct {
		mine, theirs int
		big          bool
	}{{5, 4, true}, {4, 4, false}, {3, 4, false}, {1, 0, true}} {
		g, sk := oldB588Setup(t, "Secretkeeper", "Creature — Spirit", oldB588Secret, 2, 2, tc.mine, tc.theirs)
		p, tough := effectstest.EffectivePTOf(t, g, sk)
		fly := oldB588HasKeyword(t, g, sk, "flying")
		if tc.big && (p != 4 || tough != 4 || !fly) {
			t.Errorf("hand %d vs %d: %d/%d flying=%v, want 4/4 flying", tc.mine, tc.theirs, p, tough, fly)
		}
		if !tc.big && (p != 2 || tough != 2 || fly) {
			t.Errorf("hand %d vs %d: %d/%d flying=%v, want 2/2 with no flying", tc.mine, tc.theirs, p, tough, fly)
		}
	}
}

// One opposing hand as big as yours turns it off, whatever the others hold.
func TestOldB588SecretkeeperComparesEveryOpponentAndFollowsADraw(t *testing.T) {
	g, sk := oldB588Setup(t, "Secretkeeper", "Creature — Spirit", oldB588Secret, 2, 2, 5, 2)
	effectstest.D2dSetHand(g.Seats[2], 5)
	if p, _ := effectstest.EffectivePTOf(t, g, sk); p != 2 {
		t.Fatalf("an opponent holding as many cards left it at power %d, want 2", p)
	}

	g, sk = oldB588Setup(t, "Secretkeeper", "Creature — Spirit", oldB588Secret, 2, 2, 4, 4)
	me := g.Seats[0]
	effectstest.SeedLibrary(me, "A", "B")
	if p, _ := effectstest.EffectivePTOf(t, g, sk); p != 2 {
		t.Fatalf("tied hands: power %d, want 2", p)
	}
	effectstest.DrawN(g, me, 1) // a draw is a hand change: the layers are rebuilt
	if p, _ := effectstest.EffectivePTOf(t, g, sk); p != 4 {
		t.Errorf("after drawing to a bigger hand the power is %d, want 4", p)
	}
}

func TestOldB588OkinaNightwatchGetsPlusThreeWhileYourHandIsBigger(t *testing.T) {
	for _, tc := range []struct {
		mine, theirs, power, toughness int
	}{{3, 2, 7, 6}, {3, 3, 4, 3}, {2, 3, 4, 3}} {
		g, ok := oldB588Setup(t, "Okina Nightwatch", "Creature — Human Monk", oldB588Okina, 4, 3, tc.mine, tc.theirs)
		if p, tough := effectstest.EffectivePTOf(t, g, ok); p != tc.power || tough != tc.toughness {
			t.Errorf("hand %d vs %d: %d/%d, want %d/%d", tc.mine, tc.theirs, p, tough, tc.power, tc.toughness)
		}
	}
}

// --- Akki Underling and Deathmask Nezumi --------------------------------

func TestOldB588SevenCardCreaturesGetTheirBonusesAtSevenOrMore(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, typeLine, keyword string
		power, tough, bonusP, bonusT    int
	}{
		{"Akki Underling", oldB588Akki, "Creature — Goblin Warrior", "first strike", 2, 1, 2, 1},
		{"Deathmask Nezumi", oldB588Nezumi, "Creature — Rat Shaman", "fear", 2, 2, 2, 1},
	} {
		// Six cards: nothing. Seven, with every opponent holding more: the bonus.
		g, id := oldB588Setup(t, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough, 6, 0)
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != tc.power || tough != tc.tough || oldB588HasKeyword(t, g, id, tc.keyword) {
			t.Errorf("%s with six cards: %d/%d keyword %v", tc.name, p, tough, oldB588HasKeyword(t, g, id, tc.keyword))
		}
		g, id = oldB588Setup(t, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough, 7, 9)
		if p, tough := effectstest.EffectivePTOf(t, g, id); p != tc.power+tc.bonusP || tough != tc.tough+tc.bonusT || !oldB588HasKeyword(t, g, id, tc.keyword) {
			t.Errorf("%s with seven cards: %d/%d keyword %v, want +%d/+%d and %s", tc.name, p, tough, oldB588HasKeyword(t, g, id, tc.keyword), tc.bonusP, tc.bonusT, tc.keyword)
		}
		// A draw from six to seven rebuilds the layers.
		g, id = oldB588Setup(t, tc.name, tc.typeLine, tc.oracle, tc.power, tc.tough, 6, 0)
		effectstest.SeedLibrary(g.Seats[0], "A")
		if p, _ := effectstest.EffectivePTOf(t, g, id); p != tc.power {
			t.Fatalf("%s: setup power %d", tc.name, p)
		}
		effectstest.DrawN(g, g.Seats[0], 1)
		if p, _ := effectstest.EffectivePTOf(t, g, id); p != tc.power+tc.bonusP {
			t.Errorf("%s: power after drawing the seventh card is %d, want %d", tc.name, p, tc.power+tc.bonusP)
		}
	}
}

// --- Kitsune Loreweaver ------------------------------------------------

func TestOldB588KitsuneLoreweaverAddsToughnessEqualToTheHandOnce(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	k := effectstest.PushCatalogPermanent(g, me.ID, "Kitsune Loreweaver", "Creature — Fox Cleric", oldB588Kitsune, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == k {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 2, 1
			}
		}
	})
	effectstest.D2dSetHand(me, 4)
	me.ManaPool.AddMana(game.ManaToken{Color: "W"}, game.ManaToken{Color: "C"})
	if err := g.ActivateCatalogAbility(me.ID, k, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.EffectivePTOf(t, g, k); p != 2 || tough != 5 {
		t.Fatalf("Kitsune Loreweaver is %d/%d, want 2/5", p, tough)
	}
	// X was counted once: a bigger hand later changes nothing.
	effectstest.D2dSetHand(me, 9)
	if _, tough := effectstest.EffectivePTOf(t, g, k); tough != 5 {
		t.Errorf("toughness followed the hand to %d", tough)
	}
}

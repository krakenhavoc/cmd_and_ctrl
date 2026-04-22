package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// anthem_test.go covers Glorious Anthem (Layer 7c +1/+1 to your
// creatures) end-to-end through the S16 layer engine. The catalog
// blank-import pulls Glorious Anthem's init() into scope; pushing
// it onto the battlefield fires the layer listener which bumps
// layerVersion; the next snapshot resolves the layer, and
// Effective() reads the post-7c power/toughness.

const gloriousAnthemOracle = "fa9c2b75-0c93-4eaa-a5ad-6f8aff09a98c"

// pushBattlefieldCardWithTimestamp seeds a card directly on the
// battlefield AND fires EventZoneMove so the layer listener stamps
// EnteredBattlefieldAt + bumps layerVersion. Prefer this over
// raw PushTop for any test that exercises the layer engine.
func pushBattlefieldCardWithTimestamp(g *game.Game, c game.Card) uuid.UUID {
	g.Battlefield.PushTop(c)
	g.WithWriteLock(func() {
		g.EmitEvent(game.Event{
			Kind:    game.EventZoneMove,
			CardID:  c.InstanceID,
			OldZone: game.ZoneHand,
			NewZone: game.ZoneBattlefield,
		})
	})
	return c.InstanceID
}

// effectivePower walks the battlefield for the named card and
// returns its post-layer power. Forces a snapshot so the recompute
// runs. Fails the test if the card isn't on the battlefield.
func effectivePower(t *testing.T, g *game.Game, cardID uuid.UUID) int {
	t.Helper()
	var p int
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != cardID {
				continue
			}
			eff := c.Effective()
			p = eff.Power
			found = true
			return
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", cardID)
	}
	return p
}

func effectiveToughness(t *testing.T, g *game.Game, cardID uuid.UUID) int {
	t.Helper()
	var tg int
	var found bool
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID != cardID {
				continue
			}
			eff := c.Effective()
			tg = eff.Toughness
			found = true
			return
		}
	})
	if !found {
		t.Fatalf("card %s not on battlefield", cardID)
	}
	return tg
}

func TestSingleAnthemAddsPlusOne(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0].ID
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Glorious Anthem",
		TypeLine:   "Enchantment",
		OracleID:   gloriousAnthemOracle,
		Owner:      owner,
		Controller: owner,
	})

	if got := effectivePower(t, g, bear); got != 3 {
		t.Errorf("Bears effective power = %d, want 3 (printed 2 + anthem 1)", got)
	}
	if got := effectiveToughness(t, g, bear); got != 3 {
		t.Errorf("Bears effective toughness = %d, want 3", got)
	}
}

func TestTwoAnthemsStack(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0].ID
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	for i := 0; i < 2; i++ {
		pushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(),
			Name:       "Glorious Anthem",
			TypeLine:   "Enchantment",
			OracleID:   gloriousAnthemOracle,
			Owner:      owner,
			Controller: owner,
		})
	}

	if got := effectivePower(t, g, bear); got != 4 {
		t.Errorf("Bears with 2 anthems = %d/%d power, want 4", got, effectiveToughness(t, g, bear))
	}
	if got := effectiveToughness(t, g, bear); got != 4 {
		t.Errorf("Bears with 2 anthems toughness = %d, want 4", got)
	}
}

func TestAnthemRespectsControllerBoundary(t *testing.T) {
	g := newCatalogGame(t)
	mine := g.Seats[0].ID
	theirs := g.Seats[1].ID

	myBear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "My Bear",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      mine,
		Controller: mine,
	})
	theirBear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Their Bear",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      theirs,
		Controller: theirs,
	})
	pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Glorious Anthem",
		TypeLine:   "Enchantment",
		OracleID:   gloriousAnthemOracle,
		Owner:      mine,
		Controller: mine,
	})

	if got := effectivePower(t, g, myBear); got != 3 {
		t.Errorf("my Bear effective power = %d, want 3 (anthem applies)", got)
	}
	if got := effectivePower(t, g, theirBear); got != 2 {
		t.Errorf("their Bear effective power = %d, want 2 (anthem belongs to seat 0)", got)
	}
}

func TestRemoveAnthemRevertsCreatures(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[0].ID
	bear := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	anthem := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Glorious Anthem",
		TypeLine:   "Enchantment",
		OracleID:   gloriousAnthemOracle,
		Owner:      owner,
		Controller: owner,
	})

	if got := effectivePower(t, g, bear); got != 3 {
		t.Fatalf("Bears with anthem = %d, want 3 (sanity)", got)
	}

	// Move the Anthem to the graveyard. Remove from battlefield, push
	// to graveyard, fire the EventZoneMove that the listener responds
	// to (clears effective + bumps layerVersion).
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID != anthem {
				continue
			}
			card := g.Battlefield.Cards[i]
			g.Battlefield.Cards = append(g.Battlefield.Cards[:i], g.Battlefield.Cards[i+1:]...)
			g.Seats[0].Graveyard.PushTop(card)
			g.EmitEvent(game.Event{
				Kind:    game.EventZoneMove,
				CardID:  anthem,
				OldZone: game.ZoneBattlefield,
				NewZone: game.ZoneGraveyard,
			})
			return
		}
	})

	if got := effectivePower(t, g, bear); got != 2 {
		t.Errorf("Bears after anthem removed = %d, want 2 (revert to printed)", got)
	}
}

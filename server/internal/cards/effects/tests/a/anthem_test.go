package testsa

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestSingleAnthemAddsPlusOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0].ID
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Glorious Anthem",
		TypeLine:   "Enchantment",
		OracleID:   effectstest.GloriousAnthemOracle,
		Owner:      owner,
		Controller: owner,
	})

	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
		t.Errorf("Bears effective power = %d, want 3 (printed 2 + anthem 1)", got)
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 3 {
		t.Errorf("Bears effective toughness = %d, want 3", got)
	}
}

func TestTwoAnthemsStack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0].ID
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	for i := 0; i < 2; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
			InstanceID: uuid.New(),
			Name:       "Glorious Anthem",
			TypeLine:   "Enchantment",
			OracleID:   effectstest.GloriousAnthemOracle,
			Owner:      owner,
			Controller: owner,
		})
	}

	if got := effectstest.EffectivePower(t, g, bear); got != 4 {
		t.Errorf("Bears with 2 anthems = %d/%d power, want 4", got, effectstest.EffectiveToughness(t, g, bear))
	}
	if got := effectstest.EffectiveToughness(t, g, bear); got != 4 {
		t.Errorf("Bears with 2 anthems toughness = %d, want 4", got)
	}
}

func TestAnthemRespectsControllerBoundary(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	mine := g.Seats[0].ID
	theirs := g.Seats[1].ID

	myBear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "My Bear",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      mine,
		Controller: mine,
	})
	theirBear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Their Bear",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      theirs,
		Controller: theirs,
	})
	effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Glorious Anthem",
		TypeLine:   "Enchantment",
		OracleID:   effectstest.GloriousAnthemOracle,
		Owner:      mine,
		Controller: mine,
	})

	if got := effectstest.EffectivePower(t, g, myBear); got != 3 {
		t.Errorf("my Bear effective power = %d, want 3 (anthem applies)", got)
	}
	if got := effectstest.EffectivePower(t, g, theirBear); got != 2 {
		t.Errorf("their Bear effective power = %d, want 2 (anthem belongs to seat 0)", got)
	}
}

func TestRemoveAnthemRevertsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	owner := g.Seats[0].ID
	bear := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Grizzly Bears",
		TypeLine:   "Creature — Bear",
		Power:      2,
		Toughness:  2,
		Owner:      owner,
		Controller: owner,
	})
	anthem := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       "Glorious Anthem",
		TypeLine:   "Enchantment",
		OracleID:   effectstest.GloriousAnthemOracle,
		Owner:      owner,
		Controller: owner,
	})

	if got := effectstest.EffectivePower(t, g, bear); got != 3 {
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

	if got := effectstest.EffectivePower(t, g, bear); got != 2 {
		t.Errorf("Bears after anthem removed = %d, want 2 (revert to printed)", got)
	}
}

package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tokens_spawn_test.go — ADR 0075 §2.4, the token half of the table
// spawner.

// A key that is in both tables would make TokenTemplate's answer
// depend on which map it checked first, which is the kind of thing
// nobody finds until a spawned Treasure comes out as a 1/1.
func TestSpawnableTokenKeysAreUnique(t *testing.T) {
	for k := range behaviourTokens {
		if _, clash := tokenTable[k]; clash {
			t.Errorf("%q is in both behaviourTokens and tokenTable", k)
		}
	}
	lib := Tokens()
	keys := lib.TokenKeys()
	seen := make(map[string]bool, len(keys))
	for _, k := range keys {
		if seen[k] {
			t.Errorf("TokenKeys repeats %q", k)
		}
		seen[k] = true
		if _, ok := lib.TokenTemplate(k); !ok {
			t.Errorf("TokenKeys lists %q but TokenTemplate does not resolve it", k)
		}
	}
	if len(keys) != len(tokenTable)+len(behaviourTokens) {
		t.Errorf("TokenKeys returned %d keys, want %d", len(keys), len(tokenTable)+len(behaviourTokens))
	}
	if _, ok := lib.TokenTemplate("no such token"); ok {
		t.Error("TokenTemplate resolved a key that is in neither table")
	}
}

// Every template the spawner hands out must actually be a token, or
// game.SpawnCards would route it down the raw-insert path and the
// CR 704.5d zone restriction would never be checked.
func TestEverySpawnableTemplateIsAToken(t *testing.T) {
	lib := Tokens()
	for _, k := range lib.TokenKeys() {
		tmpl, _ := lib.TokenTemplate(k)
		if !tmpl.IsToken() {
			t.Errorf("template %q has type line %q, which is not a token", k, tmpl.TypeLine)
		}
		if tmpl.Name == "" {
			t.Errorf("template %q has no name; the log line would say \"a card\"", k)
		}
	}
}

// The point of routing a spawned token through CreateTokensForEffect
// rather than the raw insert: a spawned Treasure is a real Treasure.
// A Treasure-shaped artifact with no mana ability on it looks
// identical on the board and is the failure this test exists for.
func TestSpawnedTreasureCracksLikeARealOne(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]

	tmpl, ok := Tokens().TokenTemplate("Treasure")
	if !ok {
		t.Fatal("no Treasure template")
	}
	ids, err := g.SpawnCards(me.ID, me.ID, game.ZoneBattlefield, tmpl, 1)
	if err != nil {
		t.Fatalf("SpawnCards: %v", err)
	}
	if len(ids) != 1 {
		t.Fatalf("spawned %d tokens, want 1", len(ids))
	}
	if err := g.ActivateManaAbility(me.ID, ids[0], 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("ActivateManaAbility on a spawned Treasure: %v", err)
	}

	// The Treasure sacrificed itself to pay for its own ability, and
	// the colour pick is waiting — both of which only happen if the
	// token carries the printed ability.
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID == ids[0] {
			t.Error("the spawned Treasure is still on the battlefield; its sacrifice cost was not paid")
		}
	}
	if pick := riderLatestManaPick(g, me.ID); pick == nil {
		t.Error("no mana colour pick queued; the spawned Treasure had no mana ability")
	}
}

// CR 704.5d: a token anywhere but the battlefield ceases to exist at
// the next state-based check. Refusing the spawn is the only outcome
// that tells the host why nothing happened.
func TestSpawningATokenOutsideTheBattlefieldIsRefused(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tmpl, _ := Tokens().TokenTemplate("Treasure")

	for _, zone := range []game.ZoneKind{game.ZoneHand, game.ZoneGraveyard, game.ZoneExile, game.ZoneLibrary, game.ZoneCommand} {
		before := zoneSizeForSpawnTest(g, me, zone)
		if _, err := g.SpawnCards(uuid.Nil, me.ID, zone, tmpl, 1); err != game.ErrSpawnTokenZone {
			t.Errorf("spawn a Treasure into %s: err = %v, want ErrSpawnTokenZone", zone, err)
		}
		if after := zoneSizeForSpawnTest(g, me, zone); after != before {
			t.Errorf("%s grew from %d to %d on a refused spawn", zone, before, after)
		}
	}
}

func zoneSizeForSpawnTest(g *game.Game, p *game.Player, zone game.ZoneKind) int {
	switch zone {
	case game.ZoneHand:
		return len(p.Hand.Cards)
	case game.ZoneGraveyard:
		return len(p.Graveyard.Cards)
	case game.ZoneLibrary:
		return len(p.Library.Cards)
	case game.ZoneCommand:
		return len(p.Command.Cards)
	case game.ZoneExile:
		return len(g.Exile.Cards)
	default:
		return len(g.Battlefield.Cards)
	}
}

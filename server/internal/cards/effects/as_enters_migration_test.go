package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// as_enters_migration_test.go — #578 (Discussion #560): every printed
// "When ~ enters" that used to run from the direct entry hook now goes
// on the stack. The assertion that earns its keep is the response
// window: before priority passes the trigger is on the stack and its
// effect has NOT happened; after, it has.
//
// The cards enter from hand through MoveCardByID, the path a played
// land or a sandbox drop takes. That path runs the entry pipeline and
// drains harvested triggers onto the stack before it returns, which is
// what lets the test look at the stack between entry and resolution.

// enterFromHand puts a card in owner's hand and moves it onto the
// battlefield.
func enterFromHand(t *testing.T, g *game.Game, owner uuid.UUID, name, typeLine, oracle string) uuid.UUID {
	t.Helper()
	id := uuid.New()
	p := playerByIDForTest(g, owner)
	p.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: 2, Toughness: 1, Owner: owner, Controller: owner,
	})
	if err := g.MoveCardByID(
		game.ZoneRef{Kind: game.ZoneHand, Owner: owner},
		game.ZoneRef{Kind: game.ZoneBattlefield},
		id,
	); err != nil {
		t.Fatalf("enter %s: %v", name, err)
	}
	return id
}

func TestMigratedEntersTriggersUseTheStack(t *testing.T) {
	edictFired := func(g *game.Game, _ *game.Player, _ uuid.UUID) bool {
		return sacrificeChoiceFor(g, g.Seats[1].ID) != nil
	}
	opponentBear := func(_ *testing.T, g *game.Game, _ *game.Player) uuid.UUID {
		return seedCreature(g, "Bear", g.Seats[1].ID)
	}
	cases := []struct {
		name, typeLine, oracle string
		setup                  func(t *testing.T, g *game.Game, me *game.Player) uuid.UUID
		fired                  func(g *game.Game, me *game.Player, seeded uuid.UUID) bool
	}{
		{"Accursed Marauder", "Creature — Zombie Warrior", "d8ad23a1-0b43-48ea-9fbe-d89b29194509", opponentBear, edictFired},
		{"Fleshbag Marauder", "Creature — Zombie Warrior", fleshbagMarauderOracle, opponentBear, edictFired},
		{"Merciless Executioner", "Creature — Orc Warrior", "c3c45d50-9038-41df-bb2f-9bc40071845b", opponentBear, edictFired},
		{"Bastion of Remembrance", "Enchantment", "c7f33cea-2ec8-4081-9208-a5b1d86721b3", nil,
			func(g *game.Game, me *game.Player, _ uuid.UUID) bool {
				return countBattlefieldNamed(g, me.ID, "Human Soldier") == 1
			}},
		{"Imperial Recruiter", "Creature — Human Advisor", "4d6a1391-817a-4ddc-840d-886b138eeb3f",
			func(_ *testing.T, _ *game.Game, me *game.Player) uuid.UUID {
				return pushLibraryCardForTest(me, game.Card{Name: "Ragavan", TypeLine: "Creature — Monkey Pirate", Power: 2, Toughness: 1})
			},
			func(_ *game.Game, me *game.Player, seeded uuid.UUID) bool { return me.Hand.Contains(seeded) }},
		{"Wood Elves", "Creature — Elf Scout", "8973bd99-20f8-4867-90ef-50392147ee1b",
			func(_ *testing.T, _ *game.Game, me *game.Player) uuid.UUID {
				return pushLibraryCardForTest(me, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
			},
			func(g *game.Game, _ *game.Player, seeded uuid.UUID) bool {
				_, ok := battlefieldCard(g, seeded)
				return ok
			}},
		{"Ox of Agonas", "Creature — Ox", "22113051-c971-4108-953b-95356d21323a", nil,
			func(_ *game.Game, me *game.Player, _ uuid.UUID) bool { return me.Hand.Size() == 3 }},
		{"Temple of Silence", "Land", templeOfSilenceOracle, nil,
			func(g *game.Game, me *game.Player, _ uuid.UUID) bool { return scryChoiceFor(g, me.ID) != nil }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			var seeded uuid.UUID
			if tc.setup != nil {
				seeded = tc.setup(t, g, me)
			}
			if tc.fired(g, me, seeded) {
				t.Fatalf("%s: the 'fired' probe is true before the card even entered — the test is not measuring anything", tc.name)
			}

			id := enterFromHand(t, g, me.ID, tc.name, tc.typeLine, tc.oracle)

			if triggerOnStack(g, id) == nil {
				t.Fatalf("%s: no ETB trigger on the stack — the effect is running off the stack again", tc.name)
			}
			if tc.fired(g, me, seeded) {
				t.Fatalf("%s: the effect happened before anyone could respond", tc.name)
			}

			passPriorityAroundTable(t, g)

			if !tc.fired(g, me, seeded) {
				t.Errorf("%s: the trigger resolved but the effect did not happen", tc.name)
			}
		})
	}
}

// TestEnteringTappedLandsAreNeverUntapped — Bojuka Bog and Azorius
// Chancery moved off the tap-after-entry hook onto the CR 614
// self-replacement (#360, #578). Both end up tapped either way; the
// discriminator is that a replacement emits no tap event.
func TestEnteringTappedLandsAreNeverUntapped(t *testing.T) {
	for _, land := range []struct{ name, oracle string }{
		{"Bojuka Bog", bojukaBogOracle},
		{"Azorius Chancery", azoriusChanceryOracle},
	} {
		t.Run(land.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			id := enterFromHand(t, g, me.ID, land.name, "Land", land.oracle)
			card, ok := battlefieldCard(g, id)
			if !ok {
				t.Fatalf("%s never reached the battlefield", land.name)
			}
			if !card.Tapped {
				t.Errorf("%s entered untapped", land.name)
			}
			if n := tapEventsFor(g, id); n != 0 {
				t.Errorf("%d tap events for %s; it should have ENTERED tapped, not been tapped after", n, land.name)
			}
		})
	}
}

// passUntilSpellResolves passes priority until no spell card is left
// on the stack. Anything the resolution queued — an ETB trigger — has
// been drained onto StackMeta at that boundary and is still waiting,
// which is the moment these tests want to look at.
func passUntilSpellResolves(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 32; i++ {
		if g.Stack.Size() == 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority iter %d: %v", i, err)
		}
	}
	t.Fatalf("the spell did not resolve after 32 priority passes")
}

// TestMigratedAuraEntersTriggersUseTheStack — the two Auras among the
// migrated cards enter attached, through a real cast with a target,
// so they get their own table.
func TestMigratedAuraEntersTriggersUseTheStack(t *testing.T) {
	cases := []struct {
		name, oracle string
		fired        func(g *game.Game, me *game.Player, baseline int) bool
		baseline     func(me *game.Player) int
	}{
		{"Faith's Fetters", "2b2d76f5-4c9b-49dc-b202-68095e2d9b29",
			func(_ *game.Game, me *game.Player, base int) bool { return me.Life == base+4 },
			func(me *game.Player) int { return me.Life }},
		{"Kenrith's Transformation", "a492a323-df8a-40fd-bff0-50091baa7700",
			func(_ *game.Game, me *game.Player, base int) bool { return me.Hand.Size() == base+1 },
			func(me *game.Player) int { return me.Hand.Size() }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			host := seedCreature(g, "Bear", opp.ID)

			id := castCatalogSpell(t, g, tc.name, "Enchantment — Aura", tc.oracle,
				[]game.TargetRef{{Kind: game.TargetCard, ID: host}})
			base := tc.baseline(me)
			passUntilSpellResolves(t, g)

			if _, ok := battlefieldCard(g, id); !ok {
				t.Fatalf("%s did not resolve onto the battlefield", tc.name)
			}
			if triggerOnStack(g, id) == nil {
				t.Fatalf("%s: no ETB trigger on the stack — the effect is running off the stack again", tc.name)
			}
			if tc.fired(g, me, base) {
				t.Fatalf("%s: the effect happened before anyone could respond", tc.name)
			}

			passPriorityAroundTable(t, g)

			if !tc.fired(g, me, base) {
				t.Errorf("%s: the trigger resolved but the effect did not happen", tc.name)
			}
		})
	}
}

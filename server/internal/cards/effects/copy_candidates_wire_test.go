package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// #4027: Sakashima the Impostor "can only pick tokens". The engine's
// candidate list (CopyOptions) was right all along; the wire never
// carried it for a copy_target, and the per-viewer redaction then
// dropped every candidate its chooser was not a knower of, so a
// permanent that reached the battlefield without being stamped as known
// stayed on the board and off the picker.
//
// These tests ask the chooser's own view, the thing the client renders,
// not the engine's list.

// copyWireBoard puts the shared board of the issue on the table: the
// active seat's cast creature, a seeded creature and a token, and the
// opponent's seeded creature and token, plus a noncreature artifact.
type copyWireBoard struct {
	ownCast, ownSeeded, ownToken     uuid.UUID
	oppSeeded, oppToken, artifactOpp uuid.UUID
}

func buildCopyWireBoard(t *testing.T, g *game.Game) copyWireBoard {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	var b copyWireBoard

	// Cast from hand: the zone move stamps it as known to every seat.
	b.ownCast = castCatalogSpell(t, g, "Grizzly Bears", "Creature — Bear", "oracle-wire-bears", nil)
	passPriorityAroundTable(t, g)

	// Put on the battlefield without a knower stamp, the way a card
	// that arrives by a path other than a hand cast does.
	b.ownSeeded = seedCopyableCreature(g, active.ID, "Mishra, Eminent One", "Legendary Artifact Creature — Human Artificer", 5, 4)
	b.oppSeeded = seedCopyableCreature(g, opp.ID, "Goblin Rabblemaster", "Creature — Goblin Warrior", 2, 2)
	b.artifactOpp = pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: "oracle-wire-sol-ring", Owner: opp.ID, Controller: opp.ID,
	})

	tokenIDs := func(controller uuid.UUID, name string) uuid.UUID {
		before := len(g.Battlefield.Cards)
		if err := g.CreateTokenForEffect(controller, game.Card{
			Name: name, TypeLine: "Token Creature — " + name, Power: 1, Toughness: 1,
		}, 1); err != nil {
			t.Fatalf("token %s: %v", name, err)
		}
		if len(g.Battlefield.Cards) != before+1 {
			t.Fatalf("token %s did not arrive", name)
		}
		return g.Battlefield.Cards[len(g.Battlefield.Cards)-1].InstanceID
	}
	b.ownToken = tokenIDs(active.ID, "Servo")
	b.oppToken = tokenIDs(opp.ID, "Goblin")
	return b
}

// copyWireOptions casts `name` for the active seat and returns the ids
// the chooser's view offers in the copy_target prompt.
func copyWireOptions(t *testing.T, g *game.Game, name, typeLine, oracle string) []uuid.UUID {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	castCatalogSpell(t, g, name, typeLine, oracle, nil)
	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := copyPrompt(g)
	if pc == nil {
		t.Fatalf("%s asked no copy question", name)
	}
	v := protocol.ViewOfGameFor(g, active.ID.String())
	var ids []uuid.UUID
	found := false
	for _, c := range v.PendingChoices {
		if c.Kind != string(game.PendingChoiceCopyTarget) {
			continue
		}
		found = true
		for _, o := range c.Options {
			id, err := uuid.Parse(o.InstanceID)
			if err != nil {
				t.Fatalf("option id %q: %v", o.InstanceID, err)
			}
			ids = append(ids, id)
		}
	}
	if !found {
		t.Fatalf("%s: the chooser's view has no copy_target prompt", name)
	}
	return ids
}

func copyWireSameIDs(got []uuid.UUID, want ...uuid.UUID) bool {
	if len(got) != len(want) {
		return false
	}
	for _, w := range want {
		if !slices.Contains(got, w) {
			return false
		}
	}
	return true
}

// TestCopyPickerOffersEveryLegalPermanentOnTheWire: every clone that
// shares the entry-copy path is offered the whole legal set on the
// chooser's wire — tokens and nontokens, both players'.
func TestCopyPickerOffersEveryLegalPermanentOnTheWire(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		want                   func(b copyWireBoard) []uuid.UUID
	}{
		{"Sakashima the Impostor", "Legendary Creature — Human Rogue", oracleSakashima, anyCreature},
		{"Clone", "Creature — Shapeshifter", oracleClone, anyCreature},
		{"Phantasmal Image", "Creature — Illusion", "bde94af8-faea-41ff-8eed-ba642eac9968", anyCreature},
		{"Phyrexian Metamorph", "Artifact Creature — Phyrexian Shapeshifter", oracleMetamorph,
			func(b copyWireBoard) []uuid.UUID {
				return append(anyCreature(b), b.artifactOpp)
			}},
		{"Glasspool Mimic", "Creature — Shapeshifter", "c178953c-3888-4edd-9d0c-265bd82b1d24", ownCreature},
		{"Spark Double", "Creature — Clone", oracleSparkDoubl, ownCreature},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			b := buildCopyWireBoard(t, g)
			got := copyWireOptions(t, g, tc.name, tc.typeLine, tc.oracle)
			if want := tc.want(b); !copyWireSameIDs(got, want...) {
				t.Errorf("the chooser's picker offers %d permanents %v, want these %d: %v",
					len(got), got, len(want), want)
			}
		})
	}
}

func anyCreature(b copyWireBoard) []uuid.UUID {
	return []uuid.UUID{b.ownCast, b.ownSeeded, b.ownToken, b.oppSeeded, b.oppToken}
}

func ownCreature(b copyWireBoard) []uuid.UUID {
	return []uuid.UUID{b.ownCast, b.ownSeeded, b.ownToken}
}

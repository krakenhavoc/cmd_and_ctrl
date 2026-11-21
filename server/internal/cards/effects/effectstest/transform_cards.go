package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// transform_cards_test.go — the two #1112 cards that prove ADR 0079's
// two verbs, end to end on real imported cards.
//
// # Why these go through deck.ToGameCard
//
// Same reason siege_transform_test.go gives, and it is sharper here.
// A transform card's whole point is `Faces[1]`, which lives on
// Scryfall's `card_faces[1]` and reaches game.Card only through the
// importer. castCatalogSpell builds a flat fixture with no faces at
// all, so a test written on one would exercise a CanTransform that
// always answers false and pass by doing nothing.

// transformRow builds a Scryfall record for a two-faced `transform`
// card in the shape deck.ToGameCard consumes: a null top-level mana
// cost, the joined type line, and the real data on the faces. Verbatim
// how all 401 transform oracle IDs appear in the bulk dump.
func TransformRow(oracleID, front, frontType, frontCost, back, backType, power, toughness string, colors []string) cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     front + " // " + back,
		Layout:   "transform",
		TypeLine: frontType + " // " + backType,
		OracleID: uuid.MustParse(oracleID),
		CardFaces: []cards.CardFace{
			{Name: front, TypeLine: frontType, ManaCost: frontCost, Colors: colors},
			{Name: back, TypeLine: backType, Power: power, Toughness: toughness, Colors: colors},
		},
	}
}

// importToBattlefield runs a row down the import road and puts the
// result onto the battlefield THROUGH the entry pipeline, so the ETB
// hook fires — which is what gives a Saga its CR 714.3 entry lore
// counter. Returns the permanent's instance ID.
func ImportToBattlefield(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	id := ImportToHand(row, p)
	var entered uuid.UUID
	var err error
	g.WithWriteLock(func() {
		entered, err = g.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{})
	})
	if err != nil {
		t.Fatalf("put %q onto the battlefield: %v", row.Name, err)
	}
	if entered == uuid.Nil {
		t.Fatalf("%q never reached the battlefield", row.Name)
	}
	return entered
}

// importAndCast runs a row down the import road and CASTS it from the
// active seat's hand, the way castCatalogSpell does for a fixture.
//
// A Saga has to arrive this way rather than through
// importToBattlefield: chapter I triggers off the CR 714.3 entry lore
// counter, and with no spell on the stack the first priority pass
// advances the STEP instead of settling the trigger — which adds a
// second lore counter and leaves two chapters racing for one
// trigger-order prompt.
func ImportAndCast(t *testing.T, g *game.Game, row cards.Card, p *game.Player) uuid.UUID {
	t.Helper()
	id := ImportToHand(row, p)
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(p.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", row.Name, err)
	}
	return id
}

func StormTheVaultRow() cards.Card {
	return TransformRow(effects.StormTheVaultOracleID,
		"Storm the Vault", "Legendary Enchantment", "{2}{U}{R}",
		"Vault of Catlacan", "Legendary Land", "", "", []string{"U", "R"})
}

// assertFaceInvariant checks ADR 0034's invariant on every battlefield
// permanent: the flat printed fields are the active face's. game's own
// AssertFaceInvariant lives in a _test.go file and so is not importable
// from here, and the invariant is exactly what a transform can break.
func AssertFaceInvariant(t *testing.T, g *game.Game) {
	t.Helper()
	for _, c := range g.Battlefield.Cards {
		if len(c.Faces) == 0 {
			if c.ActiveFace != 0 {
				t.Errorf("%q has no faces but ActiveFace = %d", c.Name, c.ActiveFace)
			}
			continue
		}
		if c.ActiveFace < 0 || c.ActiveFace >= len(c.Faces) {
			t.Errorf("%q ActiveFace = %d, out of range for %d faces", c.Name, c.ActiveFace, len(c.Faces))
			continue
		}
		f := c.Faces[c.ActiveFace]
		if c.Name != f.Name || c.TypeLine != f.TypeLine || c.ManaCost != f.ManaCost ||
			c.Power != f.Power || c.Toughness != f.Toughness {
			t.Errorf("%q desynchronised from Faces[%d] (%q %q) — something wrote "+
				"ActiveFace without going through SetFace", c.Name, c.ActiveFace, f.Name, f.TypeLine)
		}
	}
}

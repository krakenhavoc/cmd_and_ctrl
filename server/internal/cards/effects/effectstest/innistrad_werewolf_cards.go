package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// innistrad_werewolf_cards_test.go — the original Innistrad werewolves
// and Immerwolf (#4071, ADR 0132 amendment 2026-10-16), on real
// imported cards. The two seams they stand on (the previous turn's
// spells by every player, and "can't transform") are pinned in
// game/werewolf_transform_test.go; these drive them through the catalog
// rows, the real upkeep and the real stack.

const (
	IwFrontText = "At the beginning of each upkeep, if no spells were cast last turn, transform this creature."
	IwBackText  = "At the beginning of each upkeep, if a player cast two or more spells last turn, transform this creature."

	IwImmerwolfOracle = "3c2445ee-84d5-47de-8b09-dbdc89ca2db7"
)

// iwCard is one Innistrad werewolf as the bulk dump has it.
type IwCard struct {
	Oracle, Front, FrontType, Cost, FrontText, Fp, Ft string
	Back, BackText, Bp, Bt                            string
	Colors, Keywords                                  []string
}

// row is the Scryfall `transform` record the deck importer reads.
func (w IwCard) Row() cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     w.Front + " // " + w.Back,
		Layout:   "transform",
		TypeLine: w.FrontType + " // Creature — Werewolf",
		OracleID: uuid.MustParse(w.Oracle),
		Keywords: append([]string{"Transform"}, w.Keywords...),
		CardFaces: []cards.CardFace{
			{Name: w.Front, TypeLine: w.FrontType, ManaCost: w.Cost, Colors: w.Colors, Power: w.Fp, Toughness: w.Ft, OracleText: w.FrontText},
			{Name: w.Back, TypeLine: "Creature — Werewolf", Colors: w.Colors, Power: w.Bp, Toughness: w.Bt, OracleText: w.BackText},
		},
	}
}

// iwPut puts the imported werewolf onto p's battlefield front face up,
// able to tap.
func IwPut(t *testing.T, g *game.Game, w IwCard, p *game.Player) uuid.UUID {
	t.Helper()
	id := ImportToHand(w.Row(), p)
	c, ok := CardInZone(p.Hand, id)
	if !ok {
		t.Fatalf("%s did not reach the hand", w.Front)
	}
	p.Hand.Remove(id)
	PushBattlefieldCardWithTimestamp(g, c)
	return id
}

func IwTransform(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	var err error
	g.WithWriteLock(func() { err = g.TransformPermanentForEffect(id) })
	if err != nil {
		t.Fatalf("TransformPermanentForEffect: %v", err)
	}
}

func IwFace(g *game.Game, id uuid.UUID) int {
	c, ok := BattlefieldCard(g, id)
	if !ok {
		return -1
	}
	return c.ActiveFace
}

// iwNextUpkeep ends the current turn with `casts` (seat -> spells cast
// this turn, standing in for real casts) as its record, walks to the
// next player's upkeep and settles what triggered there. It returns that
// player's seat.
func IwNextUpkeep(t *testing.T, g *game.Game, casts map[int]int) int {
	t.Helper()
	g.WithWriteLock(func() {
		g.SpellsCastThisTurn = map[uuid.UUID]game.CastTally{}
		for seat, n := range casts {
			g.SpellsCastThisTurn[g.Seats[seat].ID] = game.CastTally{Total: n}
		}
	})
	start := g.Turn.ActiveSeat
	for i := 0; i < 40; i++ {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Step == game.StepUpkeep && g.Turn.ActiveSeat != start {
			MonSettle(t, g)
			return g.Turn.ActiveSeat
		}
	}
	t.Fatal("never reached the next upkeep")
	return -1
}

// iwTable is a catalog game in the active seat's precombat main.
func IwTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := NewCatalogGame(t)
	AdvanceToMain(t, g)
	return g, g.Seats[g.Turn.ActiveSeat]
}

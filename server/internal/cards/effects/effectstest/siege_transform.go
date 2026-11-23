package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/deck"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// siege_transform_test.go — S32: "exile it, then cast it
// transformed" (CR 310.12b), end to end, on real cards.
//
// The engine half of the seam is pinned in
// server/internal/game/exile_play_face_test.go against a fixture with
// no catalog entry. What is here is everything above it: that a real
// Siege's defeated trigger stamps a back-face grant, that the back
// face is castable for nothing out of exile, that it resolves as the
// back face, and that the BACK FACE'S OWN abilities — registered
// under "<oracle_id>#1" — run once it lands.
//
// # Why these go through deck.ToGameCard
//
// The battles' first bug (#274's twin, in the previous comment on
// #92) was invisible to fixture-built tests for exactly one reason:
// the fixtures set a field the importer had no path to. A Siege's
// back face is the same shape of risk — its name, type line, colours
// and stats live on Scryfall's `card_faces[1]`, and the whole
// transformed cast is worthless if the importer does not carry them
// onto game.Card.Faces. So every card here is a cards.Card row driven
// down the production import road, the way the theme-deck smoke test
// does it.

// siegeRow builds the Scryfall record for a two-faced battle in the
// shape deck.ToGameCard consumes: a `transform` layout, a null
// top-level mana cost and defense, and the real data on the faces.
// That is verbatim how every one of the 36 printed battles appears in
// the bulk dump.
func SiegeRow(oracleID, front, back, frontCost, defense, backType, power, toughness string) cards.Card {
	return cards.Card{
		ID:       uuid.New(),
		Name:     front + " // " + back,
		Layout:   "transform",
		TypeLine: "Battle — Siege // " + backType,
		OracleID: uuid.MustParse(oracleID),
		CardFaces: []cards.CardFace{
			{
				Name:     front,
				TypeLine: "Battle — Siege",
				ManaCost: frontCost,
				Defense:  defense,
				Colors:   []string{"R"},
			},
			{
				Name:      back,
				TypeLine:  backType,
				Power:     power,
				Toughness: toughness,
				Colors:    []string{"R"},
			},
		},
	}
}

// importToHand runs a row down the import road and puts the result in
// a player's hand, returning its instance ID.
func ImportToHand(row cards.Card, p *game.Player) uuid.UUID {
	c := deck.ToGameCard(row, false)
	c.Owner = p.ID
	c.Controller = p.ID
	p.Hand.PushTop(c)
	return c.InstanceID
}

// defeatBattle takes the last defense counter off a battle and walks
// the game far enough for the CR 704.5v/w sweep and the defeated
// trigger to have run.
func DefeatBattle(t *testing.T, g *game.Game, id uuid.UUID, defense int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToCreatureForEffect(uuid.Nil, id, defense); err != nil {
			t.Fatalf("damage the battle: %v", err)
		}
	})
	// State-based actions fire at a priority BOUNDARY; with four seats
	// a single PassPriority only rotates. AdvanceStep crosses one
	// unconditionally, and the defeated trigger drains from the same
	// pass.
	if _, err := g.AdvanceStep(); err != nil {
		t.Fatalf("AdvanceStep: %v", err)
	}
	PassPriorityAroundTable(t, g)
}

// exileGrantFor returns the grant on an exiled card.
func ExileGrantFor(g *game.Game, id uuid.UUID) *game.CastPermission {
	if perm := g.CastPermissionOnCardByIDForEffect(id); perm != nil {
		return perm
	}
	return &game.CastPermission{}
}

func BattlefieldCardFor(g *game.Game, id uuid.UUID) *game.Card {
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	return nil
}

// toMainPhaseSameTurn advances to a main phase WITHOUT crossing a
// turn boundary. The boundary matters: a Siege's transformed-cast
// grant lapses at this turn's cleanup, so a helper that wandered into
// the next turn would test the expiry rather than the cast.
func ToMainPhaseSameTurn(t *testing.T, g *game.Game) {
	t.Helper()
	turn := g.Turn.Seq
	for i := 0; i < 12; i++ {
		if g.Turn.Step == game.StepPrecombatMain || g.Turn.Step == game.StepPostcombatMain {
			return
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
		if g.Turn.Seq != turn {
			t.Fatal("crossed a turn boundary before reaching a main phase")
		}
	}
	t.Fatal("never reached a main phase")
}

// answerPickTargetPlayer answers the first open pick_target prompt
// with a player.
func AnswerPickTargetPlayer(t *testing.T, g *game.Game, playerID uuid.UUID) {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c == nil || c.Kind != game.PendingChoicePickTarget {
			continue
		}
		pick := game.TargetRef{Kind: game.TargetPlayer, ID: playerID}
		if err := g.ResolvePickTarget(c.ID, c.Chooser, pick); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
		return
	}
	t.Fatal("no pick_target prompt was queued")
}

// pendingChooseCards returns the open choose-cards prompt owed by a
// seat.
func PendingChooseCards(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.Chooser == chooser {
			return c
		}
	}
	t.Fatal("no choose_cards prompt was queued")
	return nil
}

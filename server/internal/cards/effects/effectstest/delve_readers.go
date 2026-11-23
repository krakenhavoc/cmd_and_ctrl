package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// delve_readers_test.go — ADR 0100 sub-PR 2: the readers of the cards
// delve exiled (CR 607.2q) and Teval's granted delve. The payment
// itself is delve_test.go's.

const (
	SoulflayerOracle      = "d45cf9b1-d916-48bf-8b99-60cd09230a4b"
	EtherealForagerOracle = "a3f3a5b5-b931-4961-828c-501a33e5a0a0"
	TevalArbiterOracle    = "f6cf359a-91cb-4b3c-8837-be53e4ed9c93"
	TemporalTrespassOrcl  = "c216b924-88ac-4853-9e95-0c345c09eeb6"
)

// pushGraveyardTyped puts one card of the given type line, with the
// given printed keywords, into p's graveyard and returns its ID.
func PushGraveyardTyped(p *game.Player, name, typeLine string, keywords ...string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Keywords: keywords,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

// delveLinkedPermanent returns the live permanent with this ID.
func DelveLinkedPermanent(t *testing.T, g *game.Game, id uuid.UUID) *game.Card {
	t.Helper()
	for i := range g.Battlefield.Cards {
		if g.Battlefield.Cards[i].InstanceID == id {
			return &g.Battlefield.Cards[i]
		}
	}
	t.Fatalf("card %s is not on the battlefield", id)
	return nil
}

// castAndResolve casts a card with the given delve payment and passes
// priority until the stack is empty.
func CastAndResolve(t *testing.T, g *game.Game, name, typeLine, cost, oracle string, delve []uuid.UUID) uuid.UUID {
	t.Helper()
	id, err := CastWithTapParams(t, g, name, typeLine, cost, oracle, game.CastSpellParams{DelveIDs: delve})
	if err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	PassPriorityAroundTable(t, g)
	return id
}

// --- Ethereal Forager ------------------------------------------------

// pushLinkedForager puts an Ethereal Forager onto the battlefield whose
// spell delved the given exiled cards, and returns its ID.
func PushLinkedForager(t *testing.T, g *game.Game, owner uuid.UUID, linked ...game.Card) uuid.UUID {
	t.Helper()
	id := PushDiesCreatureForTest(g, owner, "Ethereal Forager", EtherealForagerOracle, "Creature — Elemental Whale", 3, 3)
	refs := make([]game.ObjectRef, 0, len(linked))
	for _, c := range linked {
		g.Exile.PushTop(c)
		refs = append(refs, game.ObjectRef{ID: c.InstanceID, Epoch: c.ObjectEpoch})
	}
	DelveLinkedPermanent(t, g, id).Provenance.Delved = refs
	return id
}

// chooseCardsPrompt returns the open choose_cards prompt.
func ChooseCardsPrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards {
			return c
		}
	}
	t.Fatal("no choose_cards prompt is open")
	return nil
}

package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cipher_test.go — CR 702.99 (#3867): a cipher spell may be exiled
// encoded on a creature its controller controls as it resolves; the
// creature then has "whenever this creature deals combat damage to a
// player, you may copy the encoded card and cast the copy without
// paying its mana cost"; the encoding ends when the creature leaves the
// battlefield, and the card stays in exile.

const (
	LastThoughtsOracle   = "649d983a-e861-4893-8ae8-fb0a22cfad0a"
	TraitDoctoringOracle = "3b5b926b-ee8f-4c7a-8737-1de7d18ca60d"
)

// cipherResolveToPrompt passes priority until `chooser` has an option
// pick open, or the stack is empty.
func CipherResolveToPrompt(t *testing.T, g *game.Game, chooser uuid.UUID) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 16; i++ {
		if c := LatestOptionPickFor(g, chooser); c != nil {
			return c
		}
		if StackFullyEmpty(g) {
			return nil
		}
		if err := g.PassPriority(); err != nil {
			return LatestOptionPickFor(g, chooser)
		}
	}
	return LatestOptionPickFor(g, chooser)
}

// encodedOn is the creature `card` is encoded on, if any.
func EncodedOn(g *game.Game, card uuid.UUID) (uuid.UUID, bool) {
	var ref game.PermissionCardRef
	var ok bool
	g.WithWriteLock(func() { ref, ok = g.EncodedOnForEffect(card) })
	return ref.ID, ok
}

// castAndEncode casts Last Thoughts and encodes it on `creature`.
func CastAndEncode(t *testing.T, g *game.Game, me *game.Player, creature uuid.UUID) uuid.UUID {
	t.Helper()
	id := CastCatalogSpell(t, g, "Last Thoughts", "Sorcery", LastThoughtsOracle, nil)
	pick := CipherResolveToPrompt(t, g, me.ID)
	if pick == nil {
		t.Fatal("cipher asks whether to encode the card")
	}
	idx := -1
	for i, o := range pick.PickOptions {
		if len(o.Cards) == 1 && o.Cards[0] == creature {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatalf("the creature is not offered: %+v", pick.PickOptions)
	}
	AnswerOptionPick(t, g, me.ID, idx)
	g.SettleResolution()
	return id
}

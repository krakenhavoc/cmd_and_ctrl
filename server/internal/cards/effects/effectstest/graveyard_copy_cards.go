package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// castGraveyardCopier puts the creature in the active player's hand with
// its printed body and casts it.
func CastGraveyardCopier(t *testing.T, g *game.Game, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: me.ID, Controller: me.ID})
	AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	return id
}

// graveyardCreature puts a vanilla creature card in p's graveyard.
func GraveyardCreature(p *game.Player, name, subtype string, power, toughness int) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — " + subtype,
		ManaCost: "{2}{G}", Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID})
	return id
}

// waitForCopyPrompt passes priority until the copy picker opens.
func WaitForCopyPrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	for i := 0; i < 16 && CopyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	p := CopyPrompt(g)
	if p == nil {
		t.Fatal("no copy-target prompt")
	}
	return p
}

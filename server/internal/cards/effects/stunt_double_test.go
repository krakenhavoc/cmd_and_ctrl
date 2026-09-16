package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const oracleStuntDouble = "a7ff1b64-8eab-4309-892c-17a619cd302e"

// stuntDoubleCastBy puts a catalog card in `caster`'s hand and casts
// it during seat 0's precombat main. The fixture carries NO Keywords,
// so whether the cast is allowed on someone else's turn depends only
// on the catalog entry.
func stuntDoubleCastBy(t *testing.T, g *game.Game, caster *game.Player, name, oracle string) (uuid.UUID, error) {
	t.Helper()
	aangAdvanceToMain(t, g, 0)
	id := uuid.New()
	caster.Hand.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Creature — Shapeshifter",
		OracleID: oracle, Owner: caster.ID, Controller: caster.ID,
	})
	return id, g.CastSpell(caster.ID, id, game.CastSpellParams{})
}

// TestStuntDoubleFlashesInOnAnOpponentsTurnAndCopies is the card: cast
// on an opponent's turn, it offers every creature on the battlefield —
// the opponent's included, a noncreature excluded — and enters as a
// copy under the caster's control. The copy does not keep flash: it
// has exactly what the copied creature prints (CR 707.2).
func TestStuntDoubleFlashesInOnAnOpponentsTurnAndCopies(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[0]
	me := g.Seats[1]
	angel := seedCopyableCreature(g, opp.ID, "Serra Angel", "Creature — Angel", 4, 4)
	rock := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Sol Ring", TypeLine: "Artifact",
		OracleID: "oracle-sol-ring", Owner: opp.ID, Controller: opp.ID,
	})

	id, err := stuntDoubleCastBy(t, g, me, "Stunt Double", oracleStuntDouble)
	if err != nil {
		t.Fatalf("Stunt Double on an opponent's turn: %v — flash should allow it", err)
	}

	for i := 0; i < 8 && copyPrompt(g) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pc := copyPrompt(g)
	if pc == nil {
		t.Fatal("no copy prompt")
	}
	if pc.Chooser != me.ID {
		t.Errorf("chooser = %s, want the caster %s", pc.Chooser, me.ID)
	}
	if len(pc.CopyOptions) != 1 || pc.CopyOptions[0] != angel {
		t.Errorf("CopyOptions = %v, want exactly [%s] (the artifact %s is not a legal choice)",
			pc.CopyOptions, angel, rock)
	}
	resolveWithCopyChoice(t, g, angel)

	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Serra Angel" || !got.IsCopy() {
		t.Errorf("name = %q, copy = %v, want a copy of Serra Angel", got.Name, got.IsCopy())
	}
	if got.CurrentPower() != 4 || got.CurrentToughness() != 4 {
		t.Errorf("copy is %d/%d, want 4/4", got.CurrentPower(), got.CurrentToughness())
	}
	if got.Controller != me.ID {
		t.Errorf("controller = %s, want the caster %s", got.Controller, me.ID)
	}
	if hasEffectiveKeyword(t, g, id, "flash") {
		t.Error("the copy kept Stunt Double's flash; a copy has only the copied card's keywords")
	}
}

// TestStuntDoubleFlashIsTheCardsNotTheGates is the control for the
// test above: Clone, cast the same way, is refused. Without it a lax
// timing gate would pass the flash test for free.
func TestStuntDoubleFlashIsTheCardsNotTheGates(t *testing.T) {
	g := newCatalogGame(t)
	if _, err := stuntDoubleCastBy(t, g, g.Seats[1], "Clone", oracleClone); err == nil {
		t.Fatal("Clone was cast on an opponent's turn; the sorcery-speed gate is not checking flash")
	}
}

// TestStuntDoubleDecliningLeavesItAsItsPrintedSelf mirrors Clone's
// decline test, including the pinned deviation: the 0/0 survives
// (see clone.go). It keeps its own printed flash.
func TestStuntDoubleDecliningLeavesItAsItsPrintedSelf(t *testing.T) {
	g := newCatalogGame(t)
	seedCopyableCreature(g, g.Seats[0].ID, "Grizzly Bears", "Creature — Bear", 2, 2)

	id, err := stuntDoubleCastBy(t, g, g.Seats[1], "Stunt Double", oracleStuntDouble)
	if err != nil {
		t.Fatalf("CastSpell: %v", err)
	}
	resolveWithCopyChoice(t, g, uuid.Nil)

	got := copyBattlefieldCard(t, g, id)
	if got.Name != "Stunt Double" || got.IsCopy() {
		t.Errorf("name = %q, copy = %v, want the uncopied Stunt Double", got.Name, got.IsCopy())
	}
	if got.CurrentToughness() != 0 {
		t.Errorf("toughness = %d, want the printed 0", got.CurrentToughness())
	}
	if !hasEffectiveKeyword(t, g, id, "flash") {
		t.Error("a declined Stunt Double lost its printed flash")
	}
}

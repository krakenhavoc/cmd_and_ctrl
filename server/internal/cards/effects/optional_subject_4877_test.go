package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// optional_subject_4877_test.go — #4877: Lazav, Dimir Mastermind's "you
// may" names the card it would copy, so two creature cards milled at
// once give two prompts a player can tell apart.

func wireSubjectOf(g *game.Game, prompt *game.PendingChoice) *protocol.CardView {
	for _, ch := range protocol.ViewOfGame(g).PendingChoices {
		if ch.ID == prompt.ID.String() {
			return ch.Subject
		}
	}
	return nil
}

// TestLazavPromptsNameEachMilledCreature — each prompt carries its own
// card as subject, on the engine and the wire, and "yes" on the second
// copies the second.
func TestLazavPromptsNameEachMilledCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lazav := dcSeed(g, me.ID, "Lazav, Dimir Mastermind", "Legendary Creature — Shapeshifter", oracleLazavDimirMastermind, 3, 3)
	names := map[uuid.UUID]string{}
	for _, c := range []struct {
		name, sub string
		p         int
	}{{"Craterhoof Behemoth", "Beast", 5}, {"Hill Giant", "Giant", 3}} {
		id := uuid.New()
		names[id] = c.name
		opp.Library.PushTop(game.Card{InstanceID: id, Name: c.name, TypeLine: "Creature — " + c.sub,
			OracleID: "oracle-" + c.name, Power: c.p, Toughness: c.p, Owner: opp.ID, Controller: opp.ID})
	}
	g.WithWriteLock(func() {
		if _, err := g.MillToZoneForEffect(opp.ID, 2, game.ZoneGraveyard); err != nil {
			t.Fatalf("mill: %v", err)
		}
	})
	answerAnyTriggerOrderPrompt(t, g, me.ID)

	first := awaitTriggerPrompt(t, g, me.ID)
	s1 := first.Subject()
	if names[s1] == "" {
		t.Fatalf("the first prompt's subject = %s, want one of the milled creatures", s1)
	}
	if w := wireSubjectOf(g, first); w == nil || w.Name != names[s1] {
		t.Errorf("the first prompt's wire subject = %+v, want %s", w, names[s1])
	}
	if err := g.ResolveTriggerPrompt(first.ID, me.ID, false); err != nil {
		t.Fatalf("decline the first: %v", err)
	}

	second := awaitTriggerPrompt(t, g, me.ID)
	s2 := second.Subject()
	if names[s2] == "" || s2 == s1 {
		t.Fatalf("the second prompt's subject = %s, want the other milled creature (first was %s)", s2, s1)
	}
	if w := wireSubjectOf(g, second); w == nil || w.Name != names[s2] {
		t.Errorf("the second prompt's wire subject = %+v, want %s", w, names[s2])
	}
	if err := g.ResolveTriggerPrompt(second.ID, me.ID, true); err != nil {
		t.Fatalf("accept the second: %v", err)
	}
	passPriorityAroundTable(t, g)

	c := dcCard(t, g, lazav)
	want := map[string]string{"Craterhoof Behemoth": "Beast", "Hill Giant": "Giant"}[names[s2]]
	if c.Name != "Lazav, Dimir Mastermind" || !c.HasSubtype(want) {
		t.Fatalf("Lazav is %q (Beast %v, Giant %v), want a copy of %s", c.Name, c.HasSubtype("Beast"), c.HasSubtype("Giant"), names[s2])
	}
}

// TestAnOptionalTriggerWithoutShowSubjectNamesNone — a "you may" that
// does not declare a subject carries none (Perplexing Chimera's prompt
// keeps only its trade_for).
func TestAnOptionalTriggerWithoutShowSubjectNamesNone(t *testing.T) {
	g := newCatalogGame(t)
	chimeraOwner := g.Seats[1]
	pushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Perplexing Chimera",
		TypeLine: "Enchantment Creature — Chimera", OracleID: perplexingChimeraOracle, Power: 3, Toughness: 3,
		Owner: chimeraOwner.ID, Controller: chimeraOwner.ID})
	castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: chimeraOwner.ID}})
	passPriorityAroundTable(t, g)
	prompt := latestChoiceOfKindFor(g, game.PendingChoiceTriggerPrompt, chimeraOwner.ID)
	if prompt == nil {
		t.Fatal("the Chimera's controller was not asked")
	}
	if s := prompt.Subject(); s != uuid.Nil {
		t.Errorf("subject = %s, want none", s)
	}
	if w := wireSubjectOf(g, prompt); w != nil {
		t.Errorf("wire subject = %+v, want none", w)
	}
}

// TestCreamOfTheCropPromptNamesTheEnteredCreature — a ShowSubject row
// whose Build names no object falls back to the triggering event's card.
func TestCreamOfTheCropPromptNamesTheEnteredCreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dcSeed(g, me.ID, "Cream of the Crop", "Enchantment", "b61a87e3-dc98-4db9-abed-b47b677d81ab", 0, 0)
	bears := dcEnter(t, g, me.ID, "Grizzly Bears", "Creature — Bear", 2, 2)
	prompt := awaitTriggerPrompt(t, g, me.ID)
	if s := prompt.Subject(); s != bears {
		t.Fatalf("subject = %s, want the Bears %s", s, bears)
	}
	if w := wireSubjectOf(g, prompt); w == nil || w.Name != "Grizzly Bears" {
		t.Errorf("wire subject = %+v, want the Bears", w)
	}
}

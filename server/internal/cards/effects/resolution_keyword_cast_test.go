package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// resolution_keyword_cast_test.go — cipher, cascade, discover and paradigm
// cast during the resolution that offers the cast (CR 608.2g, #4243; ADR
// 0066's 2026-10-23 amendment). Per mechanic: the cast happens before the
// resolving object finishes, declining works, and a table restored while
// the prompt is open resumes. (Paradigm's are in game/paradigm_test.go,
// where its delayed trigger is built.)

// ---- cascade --------------------------------------------------------

// castShardlessAgent casts a cascade-3 creature into a library whose top
// is two lands over one cheap instant, and passes priority until the
// cascade trigger has resolved as far as the caster's offer: the lands are
// the cards cascade passes over, waiting in exile on the prompt.
func castShardlessAgent(t *testing.T, g *game.Game, me *game.Player) *game.PendingChoice {
	t.Helper()
	seedCheapLibrary(g, me, 1)
	for i := 0; i < 2; i++ {
		land := game.NewCard("Test Mountain", me.ID)
		land.TypeLine = "Basic Land — Mountain"
		me.Library.PushTop(land)
	}
	castWithCost(t, g, "Shardless Agent", "Artifact Creature — Human Rogue", "{1}{G}{U}", shardlessAgentOracle)
	passPriorityAroundTable(t, g)
	c := resolutionCastFor(g, me.ID)
	if c == nil {
		t.Fatal("cascade did not offer the hit")
	}
	return c
}

// The hit is cast while the cascade trigger is still resolving: the
// trigger's resolution is held open, nobody can pass, and the cast spell
// lands on the stack above the cascading spell.
func TestCascadeCastsTheHitBeforeTheTriggerFinishes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := castShardlessAgent(t, g, me)

	if c.MayCastKeyword != game.MayCastKeywordCascade || len(c.DiscardOptions) != 1 {
		t.Fatalf("prompt = %+v, want cascade's offer of the one hit", c)
	}
	if !g.ResolutionPaused() {
		t.Fatal("the trigger's resolution is not held open by the offer")
	}
	hit := c.DiscardOptions[0]
	if g.Stack.Contains(hit) {
		t.Fatal("the hit is on the stack before it was cast")
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("PassPriority = %v, want the choice gate's refusal: nobody has priority mid-resolution", err)
	}
	if err := g.CastForResolutionChoice(me.ID, c.ID, hit, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the hit: %v", err)
	}
	if g.ResolutionPaused() || resolutionCastFor(g, me.ID) != nil {
		t.Error("the resolution is still held after the cast")
	}
	if len(g.Exile.Cards) != 0 || me.Library.Size() != 2 {
		t.Errorf("exile %d, library %d: the two cards cascade passed over go to the bottom when the hit is cast",
			len(g.Exile.Cards), me.Library.Size())
	}
	// The hit sits above the cascading spell (CR 608.2g: it becomes the
	// topmost object).
	n := len(g.Stack.Cards)
	if n < 2 || g.Stack.Cards[n-1].Name != "Cheap Spell" || g.Stack.Cards[n-2].Name != "Shardless Agent" {
		t.Fatalf("stack = %+v, want the Cheap Spell on top of the Shardless Agent", g.Stack.Cards)
	}
}

// Declining: the hit goes to the bottom with the rest, the offer is gone
// and nothing is left in exile.
func TestCascadeDeclineBottomsTheHitAndFinishesTheResolution(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := castShardlessAgent(t, g, me)
	hit := c.DiscardOptions[0]
	if len(c.CastPile) != 2 || len(g.Exile.Cards) != 3 {
		t.Fatalf("the prompt carries %d passed-over cards with %d in exile, want 2 and 3", len(c.CastPile), len(g.Exile.Cards))
	}
	libBefore := me.Library.Size()
	if err := g.ResolvePendingChoice(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if g.ResolutionPaused() || resolutionCastFor(g, me.ID) != nil {
		t.Error("the decline did not finish the resolution")
	}
	if inExile(g, hit) || !me.Library.Contains(hit) {
		t.Error("the declined hit did not go to the bottom of the library")
	}
	if me.Library.Size() != libBefore+3 || len(g.Exile.Cards) != 0 {
		t.Errorf("library %d, exile %d, want %d and 0: every exiled card comes back", me.Library.Size(), len(g.Exile.Cards), libBefore+3)
	}
	if g.CastPermissionOnCardByIDForEffect(hit) != nil {
		t.Error("the free-cast permission outlived the decline")
	}
}

// A table paused on the cascade offer is a restore point and resumes: the
// restored offer casts the hit, or the decline bottoms it.
func TestCascadeOfferSurvivesARestoreMidPrompt(t *testing.T) {
	for _, cast := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		c := castShardlessAgent(t, g, me)
		hit := c.DiscardOptions[0]

		restored := restoreRoundTrip(t, g, true)
		rme := restored.Seats[restored.Turn.ActiveSeat]
		rc := resolutionCastFor(restored, rme.ID)
		if rc == nil || len(rc.DiscardOptions) != 1 || rc.DiscardOptions[0] != hit || !restored.ResolutionPaused() {
			t.Fatalf("cast=%v: the restored table lost the offer: %+v", cast, rc)
		}
		if rc.MayCastKeyword != game.MayCastKeywordCascade || rc.DeclineLabel == "" {
			t.Errorf("cast=%v: the restored offer lost its words: %+v", cast, rc)
		}
		if cast {
			if err := restored.CastForResolutionChoice(rme.ID, rc.ID, hit, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
				t.Fatalf("cast after the restore: %v", err)
			}
			if !restored.Stack.Contains(hit) {
				t.Error("the restored cast did not reach the stack")
			}
		} else {
			if err := restored.ResolvePendingChoice(rc.ID, rme.ID, nil); err != nil {
				t.Fatalf("decline after the restore: %v", err)
			}
			if !rme.Library.Contains(hit) {
				t.Error("the restored decline did not bottom the hit")
			}
		}
		if restored.ResolutionPaused() {
			t.Errorf("cast=%v: the restored resolution did not finish", cast)
		}
		// The cards cascade passed over came through the restore on the
		// prompt, and are on the bottom now.
		if len(rc.CastPile) != 2 || len(restored.Exile.Cards) != 0 {
			t.Errorf("cast=%v: pile %d, exile %d after the restored answer, want 2 and 0", cast, len(rc.CastPile), len(restored.Exile.Cards))
		}
		// The table carries on: the rest of the stack resolves.
		passPriorityAroundTable(t, restored)
		onBattlefield := false
		for _, c := range restored.Battlefield.Cards {
			if c.Name == "Shardless Agent" {
				onBattlefield = true
			}
		}
		if !onBattlefield {
			t.Errorf("cast=%v: the restored table did not carry on to resolve the Shardless Agent", cast)
		}
	}
}

// ---- discover -------------------------------------------------------

// castHitTheMotherLode casts the sorcery into a library whose only card
// is a seven-drop artifact, and passes priority until the discover's offer
// is open.
func castHitTheMotherLode(t *testing.T, g *game.Game, me *game.Player) (*game.PendingChoice, uuid.UUID) {
	t.Helper()
	seven := discoverLibraryCard(me.ID, "Seven Drop", "Artifact", genericCost(7))
	me.Library.Cards = []game.Card{seven}
	castCatalogSpell(t, g, "Hit the Mother Lode", "Sorcery", hitMotherLodeOracle, nil)
	passPriorityAroundTable(t, g)
	c := resolutionCastFor(g, me.ID)
	if c == nil || c.MayCastKeyword != game.MayCastKeywordDiscover {
		t.Fatalf("the discover did not offer its card: %+v", c)
	}
	return c, seven.InstanceID
}

func tappedTreasures(g *game.Game, owner uuid.UUID) (n int) {
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.Name == "Treasure" && c.Tapped {
			n++
		}
	}
	return n
}

// The discovered card is cast during the spell's resolution, and the
// spell's later instruction (the Treasures) waits for it, then reads the
// discovered card's mana value.
func TestDiscoverCastsTheCardBeforeTheSpellFinishes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c, seven := castHitTheMotherLode(t, g, me)
	if !g.ResolutionPaused() || g.Stack.Contains(seven) {
		t.Fatal("the offer is not holding the resolution, or the card is already on the stack")
	}
	if n := tappedTreasures(g, me.ID); n != 0 {
		t.Fatalf("%d Treasures before the discover was answered", n)
	}
	if err := g.CastForResolutionChoice(me.ID, c.ID, seven, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the discovered card: %v", err)
	}
	if !g.Stack.Contains(seven) {
		t.Error("the discovered card is not on the stack")
	}
	if n := tappedTreasures(g, me.ID); n != 3 {
		t.Errorf("%d Treasures after the cast, want 10 - 7 = 3", n)
	}
	if g.ResolutionPaused() {
		t.Error("the resolution is still held after the cast")
	}
}

// Declining puts the discovered card into the hand, and the Treasures are
// made all the same: "if the discovered card's mana value is less than
// 10" does not ask whether it was cast.
func TestDiscoverDeclineGoesToHandAndTheRestOfTheSpellRuns(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	c, seven := castHitTheMotherLode(t, g, me)
	if err := g.ResolvePendingChoice(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if !me.Hand.Contains(seven) {
		t.Error("the declined discovered card is not in hand")
	}
	if n := tappedTreasures(g, me.ID); n != 3 {
		t.Errorf("%d Treasures after the decline, want 3", n)
	}
}

// A table paused on the discover offer is a restore point, and the later
// step (the Treasures) still runs after the answer on the restored table.
func TestDiscoverOfferSurvivesARestoreMidPrompt(t *testing.T) {
	for _, cast := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		c, seven := castHitTheMotherLode(t, g, me)
		_ = c

		restored := restoreRoundTrip(t, g, true)
		rme := restored.Seats[0]
		rc := resolutionCastFor(restored, rme.ID)
		if rc == nil || len(rc.DiscardOptions) != 1 || rc.DiscardOptions[0] != seven || !restored.ResolutionPaused() {
			t.Fatalf("cast=%v: the restored table lost the offer: %+v", cast, rc)
		}
		if n := tappedTreasures(restored, rme.ID); n != 0 {
			t.Fatalf("cast=%v: %d Treasures before the restored answer", cast, n)
		}
		if cast {
			if err := restored.CastForResolutionChoice(rme.ID, rc.ID, seven, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
				t.Fatalf("cast after the restore: %v", err)
			}
		} else if err := restored.ResolvePendingChoice(rc.ID, rme.ID, nil); err != nil {
			t.Fatalf("decline after the restore: %v", err)
		}
		if n := tappedTreasures(restored, rme.ID); n != 3 {
			t.Errorf("cast=%v: %d Treasures after the restored answer, want 3 (the rest of the spell ran)", cast, n)
		}
		if restored.ResolutionPaused() {
			t.Errorf("cast=%v: the restored resolution did not finish", cast)
		}
	}
}

// ---- cipher ---------------------------------------------------------

// reachCipherOffer encodes Last Thoughts on a bear, hits a player with it
// and passes until the copy is offered inside the trigger's resolution.
func reachCipherOffer(t *testing.T, g *game.Game) (me *game.Player, bear, original uuid.UUID, c *game.PendingChoice) {
	t.Helper()
	advanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear = seedBear(g, me.ID)
	original = castAndEncode(t, g, me, bear)
	dealCombatDamageToPlayer(g, bear, opp.ID, 2)
	passPriorityAroundTable(t, g)
	c = resolutionCastFor(g, me.ID)
	if c == nil {
		t.Fatal("cipher did not offer the copy")
	}
	return me, bear, original, c
}

// CR 702.99a, CR 707.12, CR 608.2g: the copy is cast while the cipher
// trigger resolves. It is on offer, not on the stack, with the resolution
// held; casting it puts it on the stack and finishes the resolution, and
// it draws when it resolves.
func TestCipherCopyIsCastBeforeTheTriggerFinishes(t *testing.T) {
	g := newCatalogGame(t)
	me, bear, original, c := reachCipherOffer(t, g)
	if len(c.DiscardOptions) != 1 || !g.ResolutionPaused() {
		t.Fatalf("prompt = %+v, want one copy and a held resolution", c)
	}
	cp := c.DiscardOptions[0]
	if cp == original || g.Stack.Contains(cp) {
		t.Fatal("the offer is the original, or the copy is already on the stack")
	}
	if err := g.PassPriority(); !errors.Is(err, game.ErrChoicePending) {
		t.Fatalf("PassPriority = %v, want the choice gate's refusal", err)
	}
	if n := len(cardCopies(g)); n != 1 {
		t.Fatalf("the refused pass left %d copies, want the one on offer", n)
	}
	hand := me.Hand.Size()
	if err := g.CastForResolutionChoice(me.ID, c.ID, cp, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
		t.Fatalf("cast the copy: %v", err)
	}
	if !g.Stack.Contains(cp) || g.ResolutionPaused() {
		t.Error("the copy is not on the stack, or the resolution is still held")
	}
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d → %d: the copy draws one", hand, me.Hand.Size())
	}
	if on, ok := encodedOn(g, original); !ok || on != bear {
		t.Error("the original stays exiled and encoded")
	}
}

// Declining: the copy ceases to exist and the original stays encoded.
func TestCipherDeclineKeepsTheOriginalEncoded(t *testing.T) {
	g := newCatalogGame(t)
	me, bear, original, c := reachCipherOffer(t, g)
	if err := g.ResolvePendingChoice(c.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	if g.ResolutionPaused() {
		t.Error("the decline did not finish the resolution")
	}
	if n := len(cardCopies(g)); n != 0 {
		t.Errorf("%d copies after the decline, want 0 (CR 704.5e)", n)
	}
	if on, ok := encodedOn(g, original); !ok || on != bear {
		t.Error("declining the copy cost the encoding")
	}
}

// A table paused on the cipher offer is a restore point, and the restored
// table resumes.
func TestCipherOfferSurvivesARestoreMidPrompt(t *testing.T) {
	for _, cast := range []bool{true, false} {
		g := newCatalogGame(t)
		_, bear, original, c := reachCipherOffer(t, g)
		cp := c.DiscardOptions[0]

		restored := restoreRoundTrip(t, g, true)
		rme := restored.Seats[restored.Turn.ActiveSeat]
		rc := resolutionCastFor(restored, rme.ID)
		if rc == nil || len(rc.DiscardOptions) != 1 || rc.DiscardOptions[0] != cp || !restored.ResolutionPaused() {
			t.Fatalf("cast=%v: the restored table lost the offer: %+v", cast, rc)
		}
		hand := rme.Hand.Size()
		if cast {
			if err := restored.CastForResolutionChoice(rme.ID, rc.ID, cp, game.CastSpellParams{FromZone: "exile", Strict: true}); err != nil {
				t.Fatalf("cast after the restore: %v", err)
			}
			passPriorityAroundTable(t, restored)
			if rme.Hand.Size() != hand+1 {
				t.Errorf("the restored copy did not draw")
			}
		} else {
			if err := restored.ResolvePendingChoice(rc.ID, rme.ID, nil); err != nil {
				t.Fatalf("decline after the restore: %v", err)
			}
			if n := len(cardCopies(restored)); n != 0 {
				t.Errorf("%d copies after the restored decline", n)
			}
		}
		if restored.ResolutionPaused() {
			t.Errorf("cast=%v: the restored resolution did not finish", cast)
		}
		if on, ok := encodedOn(restored, original); !ok || on != bear {
			t.Errorf("cast=%v: the encoding did not survive", cast)
		}
	}
}

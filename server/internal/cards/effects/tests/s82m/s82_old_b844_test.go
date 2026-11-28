package testss82m

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	b844DeathWard  = "0544b707-ec67-43e3-a25d-fd7005ab673d"
	b844WordUndo   = "2597a1ee-fe9e-44b5-a5e3-d34332c6c123"
	b844MindWarp   = "5109aa76-ca40-44d1-abb4-5e7a23a2c274"
	b844Trailblzr  = "53ae0359-aa3c-464b-bdbd-139ae17ebd76"
	b844Avalanche  = "a349e526-882c-454f-ba79-ad17340f8153"
	b844ForgLore   = "ca9f538d-4191-448a-870d-f645b7c37c1b"
	b844StuntedGro = "e3480c8f-15ad-451b-bb34-b129805569e1"
)

func b844Card(ids ...uuid.UUID) []game.TargetRef {
	var out []game.TargetRef
	for _, id := range ids {
		out = append(out, game.TargetRef{Kind: game.TargetCard, ID: id})
	}
	return out
}

// Death Ward gives the target a regeneration shield: a later destruction taps it and removes it from combat instead.
func TestB844DeathWardRegeneratesTheTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	bear := effectstest.B12Creature(g, g.Seats[1].ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.D1658Cast(t, g, "Death Ward", "Instant", b844DeathWard, game.CastSpellParams{Targets: b844Card(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	g.WithWriteLock(func() { g.DestroyPermanentsForEffect([]uuid.UUID{bear}) })
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the shielded creature was destroyed")
	}
	if !effectstest.B13Tapped(t, g, bear) {
		t.Error("the regenerated creature is not tapped")
	}
}

// Word of Undoing returns the creature and the white Auras its caster owns, and only those.
func TestB844WordOfUndoingReturnsTheCreatureAndMyWhiteAuras(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	mine := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "My White Aura", TypeLine: "Enchantment — Aura", Colors: []string{"W"}})
	blue := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "My Blue Aura", TypeLine: "Enchantment — Aura", Colors: []string{"U"}})
	theirs := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their White Aura", TypeLine: "Enchantment — Aura", Colors: []string{"W"}})
	g.WithWriteLock(func() {
		for _, a := range []uuid.UUID{mine, blue, theirs} {
			_ = g.AttachForEffect(a, game.TargetRef{Kind: game.TargetCard, ID: bear})
		}
	})
	effectstest.D1658Cast(t, g, "Word of Undoing", "Instant", b844WordUndo, game.CastSpellParams{Targets: b844Card(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || effectstest.B12ZoneOf(g, bear) != game.ZoneHand {
		t.Errorf("the creature is in %v, want its owner's hand", effectstest.B12ZoneOf(g, bear))
	}
	if effectstest.B12ZoneOf(g, mine) != game.ZoneHand {
		t.Errorf("my white Aura is in %v, want my hand", effectstest.B12ZoneOf(g, mine))
	}
	if effectstest.B12ZoneOf(g, blue) == game.ZoneHand || effectstest.B12ZoneOf(g, theirs) == game.ZoneHand {
		t.Error("an Aura that is not white and mine was returned")
	}
}

// Mind Warp: choose X cards from the target's hand to discard; X = 0 only looks.
func TestB844MindWarpDiscardsXChosenCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	hand := len(opp.Hand.Cards)
	picks := []uuid.UUID{opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID}
	effectstest.D1658Cast(t, g, "Mind Warp", "Sorcery", b844MindWarp, game.CastSpellParams{XValue: 2, Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	var c *game.PendingChoice
	for _, pc := range g.PendingChoices {
		if pc != nil && pc.Chooser == me.ID {
			c = pc
		}
	}
	if c == nil {
		t.Fatal("the caster was not asked to choose")
	}
	if c.Count != 2 {
		t.Errorf("choose %d cards, want exactly 2 (%+v)", c.Count, c.Kind)
	}
	if err := g.ResolvePendingChoice(c.ID, me.ID, picks); err != nil {
		t.Fatalf("choose: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if len(opp.Hand.Cards) != hand-2 {
		t.Errorf("hand %d -> %d, want -2", hand, len(opp.Hand.Cards))
	}
	for _, id := range picks {
		if opp.Hand.Contains(id) {
			t.Error("a chosen card is still in hand")
		}
	}

	g2 := effectstest.NewCatalogGame(t)
	opp2 := g2.Seats[1]
	hand2 := len(opp2.Hand.Cards)
	effectstest.D1658Cast(t, g2, "Mind Warp", "Sorcery", b844MindWarp, game.CastSpellParams{XValue: 0, Targets: effectstest.PlayerRefs(opp2.ID)})
	effectstest.PassPriorityAroundTable(t, g2)
	if len(g2.PendingChoices) != 0 || len(opp2.Hand.Cards) != hand2 {
		t.Error("X=0 chose or discarded something")
	}
}

// Trailblazer: the target can't be blocked this turn.
func TestB844TrailblazerMakesTheTargetUnblockable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, them := effectstest.SeatsForCombat(g)
	att := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	blocker := effectstest.B12Creature(g, them.ID, "Blocker", "Creature — Bear", 2, 2)
	effectstest.D1658Cast(t, g, "Trailblazer", "Instant", b844Trailblzr, game.CastSpellParams{Targets: b844Card(att)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AttackIntoBlocks(t, g, them.ID, att)
	if effectstest.BlockRefusal(t, g, blocker, att) == "" {
		t.Error("the Trailblazer target was blocked")
	}
}

// Avalanche destroys exactly X target snow lands.
func TestB844AvalancheDestroysXSnowLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	s1 := effectstest.PushLandFor(g, opp.ID, "Snow-Covered Forest", "Basic Snow Land — Forest")
	s2 := effectstest.PushLandFor(g, opp.ID, "Snow-Covered Island", "Basic Snow Land — Island")
	plain := effectstest.PushLandFor(g, opp.ID, "Forest", "Basic Land — Forest")
	if err := effectstest.D1658CastErr(t, g, "Avalanche", "Sorcery", b844Avalanche, game.CastSpellParams{XValue: 2, Targets: b844Card(s1)}); err == nil {
		t.Fatal("X=2 was cast with one target")
	}
	if err := effectstest.D1658CastErr(t, g, "Avalanche", "Sorcery", b844Avalanche, game.CastSpellParams{XValue: 1, Targets: b844Card(plain)}); err == nil {
		t.Fatal("a nonsnow land was targeted")
	}
	effectstest.D1658Cast(t, g, "Avalanche", "Sorcery", b844Avalanche, game.CastSpellParams{XValue: 2, Targets: b844Card(s1, s2)})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(s1) || g.Battlefield.Contains(s2) || !g.Battlefield.Contains(plain) {
		t.Error("Avalanche should destroy exactly the two snow lands")
	}
}

// Stunted Growth: the target player puts three cards from their hand on top of their library.
func TestB844StuntedGrowthPutsThreeCardsOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	hand := len(opp.Hand.Cards)
	lib := len(opp.Library.Cards)
	picks := []uuid.UUID{opp.Hand.Cards[0].InstanceID, opp.Hand.Cards[1].InstanceID, opp.Hand.Cards[2].InstanceID}
	effectstest.D1658Cast(t, g, "Stunted Growth", "Sorcery", b844StuntedGro, game.CastSpellParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatal("the target player was not asked to choose")
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, picks...)
	order := effectstest.PendingOfKind(g, game.PendingChoicePutInLibrary)
	if order == nil {
		t.Fatal("no ordering prompt")
	}
	if err := g.ResolvePutInLibrary(order.ID, opp.ID, nil, []uuid.UUID{picks[2], picks[1], picks[0]}); err != nil {
		t.Fatalf("order: %v", err)
	}
	if len(opp.Hand.Cards) != hand-3 || len(opp.Library.Cards) != lib+3 {
		t.Errorf("hand %d -> %d, library %d -> %d, want -3/+3", hand, len(opp.Hand.Cards), lib, len(opp.Library.Cards))
	}
	if top := opp.Library.Cards[len(opp.Library.Cards)-1].InstanceID; !slices.Contains(picks, top) {
		t.Error("the top card is not one of the three put back")
	}
}

// Forgotten Lore: the opponent chooses, you may pay {G} to make them choose again (not a chosen card), and the
// last chosen card goes to your hand.
func TestB844ForgottenLoreRepeatsAndTakesTheLastChosenCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	var yard []uuid.UUID
	for _, name := range []string{"A", "B", "C"} {
		id := uuid.New()
		me.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
		yard = append(yard, id)
	}
	effectstest.D1658Cast(t, g, "Forgotten Lore", "Sorcery", b844ForgLore, game.CastSpellParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	first := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if first == nil || len(first.ChooseCards) != 3 {
		t.Fatalf("first prompt: %+v", first)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, yard[0])
	effectstest.ApaMana(me, "G")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	second := effectstest.ChooseCardsChoiceFor(g, opp.ID)
	if second == nil || len(second.ChooseCards) != 2 || slices.Contains(second.ChooseCards, yard[0]) {
		t.Fatalf("second prompt should offer the two unchosen cards: %+v", second)
	}
	effectstest.AnswerChooseCards(t, g, opp.ID, yard[2])
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(yard[2]) {
		t.Error("the last chosen card is not in my hand")
	}
	if me.Hand.Contains(yard[0]) || me.Hand.Contains(yard[1]) {
		t.Error("an earlier card went to hand")
	}
}

func TestB844ForgottenLoreWithoutPayingTakesTheFirstChoice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a, b := uuid.New(), uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: a, Name: "A", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	me.Graveyard.PushTop(game.Card{InstanceID: b, Name: "B", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	effectstest.D1658Cast(t, g, "Forgotten Lore", "Sorcery", b844ForgLore, game.CastSpellParams{Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, opp.ID, a)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || me.Hand.Contains(b) {
		t.Error("the first chosen card should be the one taken")
	}
}

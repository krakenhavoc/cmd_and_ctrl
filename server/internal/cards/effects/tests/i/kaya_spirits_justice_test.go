package testsi

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	cardsj "github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/cards/j"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// kaya_spirits_justice_test.go — Kaya, Spirits' Justice (Murders at
// Karlov Manor, #4469), through the real catalog entry.

func TestKayaSpiritsJusticeIsRegisteredFull(t *testing.T) {
	spec, ok := effects.Lookup(cardsj.KayaSpiritsJusticeOracle)
	if !ok {
		t.Fatal("Kaya, Spirits' Justice is not registered")
	}
	if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 || spec.StartingLoyalty != 3 {
		t.Errorf("completeness %s, caveats %v, loyalty %d", spec.Completeness, spec.Caveats, spec.StartingLoyalty)
	}
	if len(spec.Triggered) != 1 || len(spec.Activated) != 3 {
		t.Errorf("%d triggers and %d loyalty abilities, want 1 and 3", len(spec.Triggered), len(spec.Activated))
	}
}

func kayaSJ(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	kaya := effectstest.PushCatalogWalkerWithLine(g, me.ID, "Kaya, Spirits' Justice", "Legendary Planeswalker — Kaya", cardsj.KayaSpiritsJusticeOracle, 3)
	effectstest.AdvanceToMain(t, g)
	return g, me, opp, kaya
}

// kayaToken creates a 1/1 Spirit token under owner and returns it.
func kayaToken(t *testing.T, g *game.Game, owner uuid.UUID) uuid.UUID {
	t.Helper()
	before := map[uuid.UUID]bool{}
	for _, c := range g.Battlefield.Cards {
		before[c.InstanceID] = true
	}
	g.WithWriteLock(func() {
		if err := g.CreateTokenForEffect(owner, effects.KayasGuileSpirit(), 1); err != nil {
			t.Fatal(err)
		}
	})
	for _, c := range g.Battlefield.Cards {
		if !before[c.InstanceID] {
			return c.InstanceID
		}
	}
	t.Fatal("no token was created")
	return uuid.Nil
}

// kayaCopyChoice is the pending "choose a creature card from among them".
func kayaCopyChoice(g *game.Game) *game.PendingChoice {
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards && c.PickThen == cardsj.KayaSpiritsJusticeCopyPick.Key() {
			return c
		}
	}
	return nil
}

// The −2 exiles my creature and the opponent's together; the trigger
// targets my token and offers only my exiled creature card; the token is
// a copy of it with flying until end of turn.
func TestKayaSpiritsJusticeMinusTwoExilesAndTheTokenCopiesMyCreature(t *testing.T) {
	g, me, opp, kaya := kayaSJ(t)
	mine := effectstest.PushVanillaCreature(g, me.ID, "Hill Giant", 3, 3)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	token := kayaToken(t, g, me.ID)
	if err := g.ActivateCatalogAbility(me.ID, kaya, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{effectstest.CardRef(mine), effectstest.CardRef(theirs)},
	}); err != nil {
		t.Fatalf("−2: %v", err)
	}
	if effectstest.LoyaltyCount(g, kaya) != 1 {
		t.Errorf("loyalty %d, want 1", effectstest.LoyaltyCount(g, kaya))
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, mine) || !effectstest.InExile(g, theirs) {
		t.Fatal("both creatures should be exiled")
	}
	g.RunStateChecksForTest()
	effectstest.AnswerPickTargetIfOpen(t, g, token)
	effectstest.PassPriorityAroundTable(t, g)
	c := kayaCopyChoice(g)
	if c == nil {
		t.Fatal("no choice of an exiled creature card")
	}
	if len(c.ChooseCards) != 1 || c.ChooseCards[0] != mine {
		t.Fatalf("offered %v, want only my exiled creature %s", c.ChooseCards, mine)
	}
	if err := g.ResolveChooseCards(c.ID, me.ID, []uuid.UUID{mine}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	tc := effectstest.DcCard(t, g, token)
	if tc.Name != "Hill Giant" || tc.Effective().Power != 3 || !game.HasKeyword(&tc, "flying") {
		t.Errorf("token is %q %d/%d flying=%v, want a flying Hill Giant", tc.Name, tc.Effective().Power, tc.Effective().Toughness, game.HasKeyword(&tc, "flying"))
	}
	effectstest.AdvancePastCleanupForTest(t, g)
	if tc := effectstest.DcCard(t, g, token); tc.Name != "Spirit" {
		t.Errorf("after cleanup the token is %q, want Spirit", tc.Name)
	}
}

// A creature card exiled from my graveyard triggers it; "you may" lets
// me choose nothing.
func TestKayaSpiritsJusticeGraveyardExileAndDecline(t *testing.T) {
	for _, take := range []bool{true, false} {
		g, me, _, _ := kayaSJ(t)
		token := kayaToken(t, g, me.ID)
		dead := effectstest.GyCard(me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")
		g.WithWriteLock(func() {
			if err := g.ExileCardForEffect(dead); err != nil {
				t.Fatal(err)
			}
		})
		g.RunStateChecksForTest()
		effectstest.AnswerPickTargetIfOpen(t, g, token)
		effectstest.PassPriorityAroundTable(t, g)
		c := kayaCopyChoice(g)
		if c == nil || len(c.ChooseCards) != 1 || c.ChooseCards[0] != dead {
			t.Fatalf("take=%v: choice %+v, want the exiled Bears", take, c)
		}
		var pick []uuid.UUID
		if take {
			pick = []uuid.UUID{dead}
		}
		if err := g.ResolveChooseCards(c.ID, me.ID, pick); err != nil {
			t.Fatalf("take=%v: %v", take, err)
		}
		want := "Spirit"
		if take {
			want = "Grizzly Bears"
		}
		if tc := effectstest.DcCard(t, g, token); tc.Name != want {
			t.Errorf("take=%v: token is %q, want %q", take, tc.Name, want)
		}
	}
}

// Not triggered by an opponent's creature or a noncreature card from my
// graveyard; and a card that left exile before resolution can't be
// chosen (the ruling), so nothing is asked.
func TestKayaSpiritsJusticeTriggerScope(t *testing.T) {
	g, me, opp, _ := kayaSJ(t)
	kayaToken(t, g, me.ID)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	sorcery := effectstest.GyCard(me, "Old Sorcery", "Sorcery", "{1}", "")
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(theirs); err != nil {
			t.Fatal(err)
		}
		if err := g.ExileCardForEffect(sorcery); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	if g.Stack.Size() != 0 || len(g.PendingChoices) != 0 {
		t.Fatalf("triggered on an opponent's creature or a sorcery card (stack %d, choices %d)", g.Stack.Size(), len(g.PendingChoices))
	}

	token := kayaToken(t, g, me.ID)
	dead := effectstest.GyCard(me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "")
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(dead); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
	effectstest.AnswerPickTargetIfOpen(t, g, token)
	g.WithWriteLock(func() {
		if _, err := game.MoveCard(g.Exile, me.Hand, dead); err != nil {
			t.Fatal(err)
		}
	})
	effectstest.PassPriorityAroundTable(t, g)
	if c := kayaCopyChoice(g); c != nil {
		t.Errorf("offered %v after the card left exile", c.ChooseCards)
	}
}

// +1: a 1/1 white and black Spirit with flying.
func TestKayaSpiritsJusticePlusOneMakesASpirit(t *testing.T) {
	g, me, _, kaya := kayaSJ(t)
	if err := g.ActivateCatalogAbility(me.ID, kaya, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+1: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var spirit *game.Card
	for i := range g.Battlefield.Cards {
		if c := &g.Battlefield.Cards[i]; c.Name == "Spirit" && c.Controller == me.ID {
			spirit = c
		}
	}
	if spirit == nil {
		t.Fatal("no Spirit token")
	}
	if spirit.Power != 1 || spirit.Toughness != 1 || !slices.Equal(spirit.Colors, []string{"W", "B"}) || !game.HasKeyword(spirit, "flying") {
		t.Errorf("token %+v, want a 1/1 white and black flier", spirit)
	}
	if effectstest.LoyaltyCount(g, kaya) != 4 {
		t.Errorf("loyalty %d, want 4", effectstest.LoyaltyCount(g, kaya))
	}
}

// +2: surveil 2, then exile a card from any graveyard.
func TestKayaSpiritsJusticePlusTwoSurveilsThenExiles(t *testing.T) {
	g, me, opp, kaya := kayaSJ(t)
	second := effectstest.TopLibraryCard(me, "Second", "Sorcery")
	top := effectstest.TopLibraryCard(me, "Top", "Sorcery")
	theirs := effectstest.GyCard(opp, "Their Card", "Instant", "{U}", "")
	if err := g.ActivateCatalogAbility(me.ID, kaya, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("+2: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	sv := effectstest.SurveilChoiceFor(g, me.ID)
	if sv == nil || len(sv.ScryCards) != 2 {
		t.Fatalf("surveil prompt %+v, want one over two cards", sv)
	}
	if err := g.ResolveSurveil(sv.ID, me.ID, []uuid.UUID{top}, []uuid.UUID{second}); err != nil {
		t.Fatalf("ResolveSurveil: %v", err)
	}
	var pick *game.PendingChoice
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseCards {
			pick = c
		}
	}
	if pick == nil || !slices.Contains(pick.ChooseCards, theirs) || !slices.Contains(pick.ChooseCards, top) || pick.ChooseMin != 1 {
		t.Fatalf("exile pick %+v, want a required pick over every graveyard", pick)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{theirs}); err != nil {
		t.Fatal(err)
	}
	if !effectstest.InExile(g, theirs) || !me.Graveyard.Contains(top) {
		t.Error("the chosen card is not exiled, or the surveilled card moved")
	}
	if effectstest.LoyaltyCount(g, kaya) != 5 {
		t.Errorf("loyalty %d, want 5", effectstest.LoyaltyCount(g, kaya))
	}
}

// The −2 needs a creature I control, and an opponent's slot takes only
// that opponent's creature.
func TestKayaSpiritsJusticeMinusTwoTargetRefusals(t *testing.T) {
	g, me, opp, kaya := kayaSJ(t)
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, kaya, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{effectstest.CardRef(theirs)}, Strict: true,
	}); err == nil {
		t.Error("−2 accepted an opponent's creature as my creature")
	}
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	other := effectstest.PushVanillaCreature(g, me.ID, "My Other Bear", 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, kaya, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{effectstest.CardRef(mine), effectstest.CardRef(other)}, Strict: true,
	}); err == nil {
		t.Error("−2 accepted my creature in an opponent's slot")
	}
	if effectstest.LoyaltyCount(g, kaya) != 3 || !effects.OnBattlefield(g, mine) || !effects.OnBattlefield(g, theirs) {
		t.Fatal("a refused activation paid or exiled something")
	}
	// Only my creature: the opponents' slots are "up to one".
	if err := g.ActivateCatalogAbility(me.ID, kaya, 2, game.ActivateAbilityParams{
		Targets: []game.TargetRef{effectstest.CardRef(mine)}, Strict: true,
	}); err != nil {
		t.Fatalf("−2 with only my creature: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, mine) || !effects.OnBattlefield(g, theirs) {
		t.Error("only my creature should be exiled")
	}
}

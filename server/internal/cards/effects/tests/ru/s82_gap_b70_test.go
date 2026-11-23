package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b70_test.go — Ravnica Remastered creatures, batch 70 (#5948). Each
// test plays the real catalog entry through the engine.

const (
	gapB70PackRatOracle    = "5632c2df-71e6-41b7-b8f5-c683812cc0e2"
	gapB70DenizenOracle    = "3dd2f3de-98cc-4263-bd16-b5803e48e892"
	gapB70GlowsporeOracle  = "ac7a9d73-dbdf-4e97-b382-67cd30ef5649"
	gapB70SavraOracle      = "641db975-cf94-44c6-a391-f540a952e0d8"
	gapB70GobhobblerOracle = "aa254e74-583d-4681-8169-5ac0dbdb10e5"
)

func TestGapB70EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB70PackRatOracle:    "Pack Rat",
		gapB70DenizenOracle:    "Shadow Alley Denizen",
		gapB70GlowsporeOracle:  "Glowspore Shaman",
		gapB70SavraOracle:      "Savra, Queen of the Golgari",
		gapB70GobhobblerOracle: "Gobhobbler Rats",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func gapB70Rats(g *game.Game, controller uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == "Pack Rat" {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func TestGapB70PackRatCountsRatsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rat := effectstest.PushCatalogPermanent(g, me.ID, "Pack Rat", "Creature — Rat", gapB70PackRatOracle, false)
	if p, tg := effectstest.EffectivePower(t, g, rat), effectstest.EffectiveToughness(t, g, rat); p != 1 || tg != 1 {
		t.Errorf("a lone Pack Rat is %d/%d, want 1/1", p, tg)
	}
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Other Rat", TypeLine: "Creature — Rat", Power: 1, Toughness: 1})
	effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Their Rat", TypeLine: "Creature — Rat", Power: 1, Toughness: 1})
	effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if p, tg := effectstest.EffectivePower(t, g, rat), effectstest.EffectiveToughness(t, g, rat); p != 2 || tg != 2 {
		t.Errorf("with two Rats Pack Rat is %d/%d, want 2/2 (an opponent's Rat and a Bear don't count)", p, tg)
	}
}

func TestGapB70PackRatDiscardsToMakeACopy(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rat := effectstest.PushCatalogPermanent(g, me.ID, "Pack Rat", "Creature — Rat", gapB70PackRatOracle, false)
	discard := effectstest.HandCardForTest(me, "Discarded", "Sorcery", "")
	effectstest.AdvanceToMain(t, g)
	effectstest.FillPoolColored(me, "B", 3)
	if err := g.ActivateCatalogAbility(me.ID, rat, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without paying the discard")
	}
	if err := g.ActivateCatalogAbility(me.ID, rat, 0, game.ActivateAbilityParams{Strict: true, DiscardIDs: []uuid.UUID{discard}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(discard) || !me.Graveyard.Contains(discard) {
		t.Error("the discarded card did not reach the graveyard")
	}
	rats := gapB70Rats(g, me.ID)
	if len(rats) != 2 {
		t.Fatalf("%d Pack Rats on the battlefield, want the original and a token copy", len(rats))
	}
	tokens := 0
	for _, id := range rats {
		if effectstest.FindBattlefieldCardForTest(g, id).IsToken() {
			tokens++
		}
		if p, tg := effectstest.EffectivePower(t, g, id), effectstest.EffectiveToughness(t, g, id); p != 2 || tg != 2 {
			t.Errorf("a Pack Rat is %d/%d, want 2/2 with two Rats", p, tg)
		}
	}
	if tokens != 1 {
		t.Errorf("%d tokens, want 1", tokens)
	}
}

func TestGapB70ShadowAlleyDenizenGivesIntimidateForABlackCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	denizen := effectstest.B21Push(g, me.ID, "Shadow Alley Denizen", "Creature — Vampire Rogue", gapB70DenizenOracle, 1, 1, "B")
	friend := effectstest.B21Push(g, me.ID, "Friend", "Creature — Bear", "", 2, 2, "G")
	// A green creature entering does nothing.
	r35B23Cast(t, g, "Green Bear", "Creature — Bear", "{G}", "", []string{"G"}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil || effectstest.HasKeywordOnBattlefield(t, g, friend, "intimidate") {
		t.Fatal("a green creature triggered the Denizen")
	}
	// A black one does, and the target is any creature.
	black := r35B23Cast(t, g, "Black Bear", "Creature — Bear", "{B}", "", []string{"B"}, nil)
	effectstest.B04WaitForPick(t, g, me.ID)
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil || !effectstest.HasID(p.PickTargetCards, friend) || !effectstest.HasID(p.PickTargetCards, denizen) || !effectstest.HasID(p.PickTargetCards, black) {
		t.Fatalf("target creature: offered %+v", p)
	}
	effectstest.PickCard(t, g, me.ID, friend)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, friend, "intimidate") {
		t.Error("the target did not gain intimidate")
	}
	// An opponent's black creature entering is not "you control".
	effectstest.AdvanceToMainOf(t, g, 1)
	r35B23Cast(t, g, "Their Bear", "Creature — Bear", "{B}", "", []string{"B"}, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestPickTarget(g, me.ID) != nil {
		t.Error("an opponent's creature triggered the Denizen")
	}
}

func TestGapB70GlowsporeShamanMillsThenPutsALandOnTop(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := effectstest.HornTopFive(me, []struct{ Name, TypeLine string }{
		{"Forest", "Basic Land — Forest"},
		{"Spell", "Sorcery"},
		{"Swamp", "Basic Land — Swamp"},
		{"Next", "Sorcery"},
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.EnterFromHand(t, g, me.ID, "Glowspore Shaman", "Creature — Elf Shaman", gapB70GlowsporeOracle)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if prompt == nil {
		t.Fatal("no land prompt after the mill")
	}
	if len(prompt.ChooseCards) < 2 || !effectstest.HasID(prompt.ChooseCards, top[0]) || !effectstest.HasID(prompt.ChooseCards, top[2]) || effectstest.HasID(prompt.ChooseCards, top[1]) {
		t.Fatalf("offered %v, want the two milled lands and no spell", prompt.ChooseCards)
	}
	if prompt.ChooseMin != 0 || prompt.ChooseMax != 1 {
		t.Errorf("min %d max %d, want an optional single pick", prompt.ChooseMin, prompt.ChooseMax)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, top[2])
	if cards := me.Library.Cards; len(cards) == 0 || cards[len(cards)-1].InstanceID != top[2] {
		t.Error("the Swamp is not on top of the library")
	}
	if me.Graveyard.Contains(top[2]) || !me.Graveyard.Contains(top[0]) || !me.Graveyard.Contains(top[1]) {
		t.Error("the other two milled cards should stay in the graveyard")
	}
}

func TestGapB70GlowsporeShamanMayDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	top := effectstest.HornTopFive(me, []struct{ Name, TypeLine string }{
		{"Forest", "Basic Land — Forest"}, {"A", "Sorcery"}, {"B", "Sorcery"},
	})
	effectstest.AdvanceToMain(t, g)
	effectstest.EnterFromHand(t, g, me.ID, "Glowspore Shaman", "Creature — Elf Shaman", gapB70GlowsporeOracle)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID)
	if !me.Graveyard.Contains(top[0]) {
		t.Error("declining should leave the land in the graveyard")
	}
}

func gapB70Savra(t *testing.T) (*game.Game, uuid.UUID, uuid.UUID, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	savra := effectstest.B21Push(g, me.ID, "Savra, Queen of the Golgari", "Legendary Creature — Elf Shaman", gapB70SavraOracle, 2, 2, "B", "G")
	black := effectstest.B21Push(g, me.ID, "Black Fodder", "Creature — Rat", "", 1, 1, "B")
	green := effectstest.B21Push(g, me.ID, "Green Fodder", "Creature — Elf", "", 1, 1, "G")
	var theirs []uuid.UUID
	for _, p := range g.Seats[1:] {
		theirs = append(theirs, effectstest.B21Push(g, p.ID, "Their Bear", "Creature — Bear", "", 2, 2, "G"))
	}
	return g, savra, black, green, theirs
}

func TestGapB70SavraPaysLifeForAnEdictWhenABlackCreatureIsSacrificed(t *testing.T) {
	g, _, black, _, theirs := gapB70Savra(t)
	me := g.Seats[0]
	life := me.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(black) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if prompt := effectstest.ChooseCardsChoiceFor(g, p.ID); prompt != nil {
			effectstest.AnswerChooseCards(t, g, p.ID, prompt.ChooseCards[0])
		}
	}
	if me.Life != life-2 {
		t.Errorf("life %d, want %d", me.Life, life-2)
	}
	for _, id := range theirs {
		if effectstest.FindBattlefieldCardForTest(g, id) != nil {
			t.Errorf("an opponent's creature survived the edict")
		}
	}
	if effectstest.FindBattlefieldCardForTest(g, black) != nil {
		t.Error("the sacrificed creature is still there")
	}
}

func TestGapB70SavraDeclinedLifePaymentMakesNoOneSacrifice(t *testing.T) {
	g, _, black, _, theirs := gapB70Savra(t)
	me := g.Seats[0]
	life := me.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(black) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life {
		t.Errorf("life %d, want %d", me.Life, life)
	}
	for _, id := range theirs {
		if effectstest.FindBattlefieldCardForTest(g, id) == nil {
			t.Error("an opponent sacrificed though the life was not paid")
		}
	}
}

func TestGapB70SavraGainsTwoLifeForAGreenCreature(t *testing.T) {
	g, _, _, green, theirs := gapB70Savra(t)
	me := g.Seats[0]
	life := me.Life
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(green) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	for _, id := range theirs {
		if effectstest.FindBattlefieldCardForTest(g, id) == nil {
			t.Error("a green sacrifice made an opponent sacrifice")
		}
	}
}

func TestGapB70SavraIgnoresAnOpponentsSacrificeAndOtherColors(t *testing.T) {
	g, _, _, _, theirs := gapB70Savra(t)
	me := g.Seats[0]
	red := effectstest.B21Push(g, me.ID, "Red Fodder", "Creature — Goblin", "", 1, 1, "R")
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(red) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("a red creature triggered Savra")
	}
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(theirs[0]) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Error("an opponent's sacrifice triggered Savra")
	}
}

func TestGapB70GobhobblerRatsAreHellbent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	rats := effectstest.B21Push(g, me.ID, "Gobhobbler Rats", "Creature — Rat", gapB70GobhobblerOracle, 2, 2, "B", "R")
	if me.Hand.Size() == 0 {
		t.Fatal("the test needs a hand to start from")
	}
	if p := effectstest.EffectivePower(t, g, rats); p != 2 {
		t.Errorf("with cards in hand the Rats have power %d, want 2", p)
	}
	effectstest.FillPoolColored(me, "B", 1)
	if err := g.ActivateCatalogAbility(me.ID, rats, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("regenerated without being hellbent")
	}
	for len(me.Hand.Cards) > 0 {
		id := me.Hand.Cards[0].InstanceID
		if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneHand, Owner: me.ID}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, id); err != nil {
			t.Fatalf("discard: %v", err)
		}
	}
	if p := effectstest.EffectivePower(t, g, rats); p != 3 {
		t.Errorf("hellbent the Rats have power %d, want 3", p)
	}
	if err := g.ActivateCatalogAbility(me.ID, rats, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("hellbent regeneration: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	// A card back in hand takes both away.
	r35B23Draw(g, me.ID, 1)
	if p := effectstest.EffectivePower(t, g, rats); p != 2 {
		t.Errorf("with a card back in hand the Rats have power %d, want 2", p)
	}
}

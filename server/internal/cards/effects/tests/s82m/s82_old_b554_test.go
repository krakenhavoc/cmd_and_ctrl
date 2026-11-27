package testss82m

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b554_test.go — Lorwyn, token makers, batch 554 (#5593). Every test
// plays the real catalog entry through the engine.

const (
	oldB554Rebellion  = "0e2a283b-c06a-4b54-a627-5b4959318b1a"
	oldB554Guardian   = "27d6b1fb-ce53-45d6-a601-9f5b9ee7ab2f"
	oldB554Eyes       = "3c59ab03-8744-4148-be06-eb1b0147f2cc"
	oldB554Benthicore = "68eb3566-c74d-42a6-89a6-b468677fc485"
	oldB554Hearthcage = "7e523d42-c7d8-47d1-8874-230920f03345"
	oldB554Pride      = "d311154b-4742-49e5-9b65-4de2d3e0d50b"
	oldB554Greed      = "60040023-7bf0-42e0-ab30-46a730675564"
	oldB554Bear       = "Creature — Bear"
)

func TestOldB554EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB554Rebellion: "Rebellion of the Flamekin", oldB554Guardian: "Guardian of Cloverdell", oldB554Eyes: "Eyes of the Wisent",
		oldB554Benthicore: "Benthicore", oldB554Hearthcage: "Hearthcage Giant", oldB554Pride: "Militia's Pride",
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

func oldB554TokensOf(g *game.Game, controller uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.IsToken() && c.Name == name {
			out = append(out, c)
		}
	}
	return out
}

func oldB554LibraryTop(p *game.Player, name, cost string) {
	c := game.NewCard(name, p.ID)
	c.TypeLine = "Sorcery"
	c.ManaCost = cost
	p.Library.PushTop(c)
}

// oldB554ClashOnce has the active seat cast Hoarder's Greed and clash with the
// opponent: my revealed card has mana value `mine`, theirs `theirs`. Both keep
// their card on top. It returns once the clash is done.
func oldB554ClashOnce(t *testing.T, g *game.Game, me, opp *game.Player, mine, theirs string) {
	t.Helper()
	oldB554LibraryTop(opp, "Their Card", theirs)
	oldB554LibraryTop(me, "My Card", mine)
	oldB554LibraryTop(me, "Draw Two", "{1}")
	oldB554LibraryTop(me, "Draw One", "{1}")
	effectstest.CastCatalogSpell(t, g, "Hoarder's Greed", "Sorcery", oldB554Greed, nil)
	effectstest.PassPriorityAroundTable(t, g)
	pick := func(c *game.PendingChoice) {
		for i, o := range c.PickOptions {
			if o.Player == opp.ID {
				if err := g.ResolveOptionPick(c.ID, me.ID, i); err != nil {
					t.Fatalf("choose the opponent: %v", err)
				}
				return
			}
		}
		t.Fatalf("no opponent offered: %q", c.Reason)
	}
	answer := func() {
		for _, c := range g.PendingChoices {
			if c == nil || c.Kind != game.PendingChoiceOptionPick {
				continue
			}
			if strings.Contains(c.PickOptions[len(c.PickOptions)-1].Label, "bottom") {
				if err := g.ResolveOptionPick(c.ID, c.Chooser, 0); err != nil {
					t.Fatalf("decide: %v", err)
				}
				return
			}
			pick(c)
			return
		}
	}
	for i := 0; i < 3; i++ {
		answer()
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func oldB554AnswerMay(t *testing.T, g *game.Game, who *game.Player, yes bool) {
	t.Helper()
	p := effectstest.LatestTriggerPromptFor(g, who.ID)
	if p == nil {
		t.Fatal("no optional trigger prompt is open")
	}
	if err := g.ResolveTriggerPrompt(p.ID, who.ID, yes); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
}

func oldB554PayIfAsked(t *testing.T, g *game.Game, who *game.Player, mana string, pay bool) bool {
	t.Helper()
	if !effectstest.HasPayUnlessFor(g, who.ID) {
		return false
	}
	if pay {
		effectstest.FloatForTest(g, who, mana)
	}
	effectstest.AnswerPayUnless(t, g, who.ID, pay)
	return true
}

func TestOldB554RebellionOfTheFlamekinMakesAHastyShamanOnlyWhenPaidAndWon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Rebellion of the Flamekin", "Kindred Enchantment — Elemental", oldB554Rebellion, false)
	// A lost clash: 1 against 5 makes a Shaman without haste for {1}.
	oldB554ClashOnce(t, g, me, opp, "{B}", "{3}{U}{U}")
	if !oldB554PayIfAsked(t, g, me, "R", true) {
		for _, c := range g.PendingChoices {
			t.Logf("pending %s %q chooser=%s", c.Kind, c.Reason, c.Chooser)
		}
		t.Fatal("no {1} offer after a clash")
	}
	effectstest.PassPriorityAroundTable(t, g)
	shamans := oldB554TokensOf(g, me.ID, "Elemental Shaman")
	if len(shamans) != 1 {
		t.Fatalf("%d Elemental Shaman tokens, want 1", len(shamans))
	}
	if effectstest.HasEffectiveKeyword(t, g, shamans[0].InstanceID, "haste") {
		t.Error("the Shaman has haste after a lost clash")
	}
	if p, tt := effectstest.PtOf(t, g, shamans[0].InstanceID); p != 3 || tt != 1 {
		t.Errorf("the token is %d/%d, want 3/1", p, tt)
	}
}

func TestOldB554RebellionOfTheFlamekinGivesHasteForAWonClashAndNothingWhenDeclined(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Rebellion of the Flamekin", "Kindred Enchantment — Elemental", oldB554Rebellion, false)
	// Hoarder's Greed repeats on a win; a second, lost clash ends it.
	oldB554LibraryTop(opp, "Their Two", "{1}{U}")
	oldB554LibraryTop(me, "My One", "{B}")
	oldB554LibraryTop(me, "Draw Four", "{1}")
	oldB554LibraryTop(me, "Draw Three", "{1}")
	oldB554LibraryTop(opp, "Their One", "{U}")
	oldB554LibraryTop(me, "My Five", "{3}{B}{B}")
	oldB554LibraryTop(me, "Draw Two", "{1}")
	oldB554LibraryTop(me, "Draw One", "{1}")
	effectstest.CastCatalogSpell(t, g, "Hoarder's Greed", "Sorcery", oldB554Greed, nil)
	effectstest.PassPriorityAroundTable(t, g)
	step := func() {
		for _, c := range g.PendingChoices {
			if c == nil || c.Kind != game.PendingChoiceOptionPick {
				continue
			}
			idx := 0
			if !strings.Contains(c.PickOptions[len(c.PickOptions)-1].Label, "bottom") {
				for i, o := range c.PickOptions {
					if o.Player == opp.ID {
						idx = i
					}
				}
			}
			if err := g.ResolveOptionPick(c.ID, c.Chooser, idx); err != nil {
				t.Fatalf("answer: %v", err)
			}
			return
		}
	}
	for i := 0; i < 8 && !effectstest.HasPayUnlessFor(g, me.ID); i++ {
		step()
	}
	var asked []string
	for i := 0; i < 12; i++ {
		if effectstest.HasPayUnlessFor(g, me.ID) {
			pay := len(asked) == 0 // pay for the first (won) clash, decline the second
			asked = append(asked, map[bool]string{true: "paid", false: "declined"}[pay])
			oldB554PayIfAsked(t, g, me, "R", pay)
			effectstest.PassPriorityAroundTable(t, g)
			continue
		}
		step()
		effectstest.PassPriorityAroundTable(t, g)
	}
	shamans := oldB554TokensOf(g, me.ID, "Elemental Shaman")
	if len(shamans) != 1 {
		t.Fatalf("%d Elemental Shaman tokens after %v, want 1 (one clash paid for)", len(shamans), asked)
	}
	if !effectstest.HasEffectiveKeyword(t, g, shamans[0].InstanceID, "haste") {
		t.Errorf("the Shaman from the won clash has no haste (%v)", asked)
	}
}

func TestOldB554GuardianOfCloverdellMakesThreeKithkinAndSacrificesOneForLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	guardian := effectstest.CastHoldCreature(t, g, "Guardian of Cloverdell", "Creature — Treefolk Shaman", oldB554Guardian, 4, 5)
	effectstest.PassPriorityAroundTable(t, g)
	kithkin := oldB554TokensOf(g, me.ID, "Kithkin Soldier")
	if len(kithkin) != 3 {
		t.Fatalf("%d Kithkin Soldier tokens, want 3", len(kithkin))
	}
	life := effectstest.LifeOf(g, me.ID)
	effectstest.FloatForTest(g, me, "G")
	if err := g.ActivateCatalogAbility(me.ID, guardian, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{guardian}}); err == nil {
		t.Fatal("sacrificed the Guardian, which is no Kithkin")
	}
	effectstest.FloatForTest(g, me, "G")
	if err := g.ActivateCatalogAbility(me.ID, guardian, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{kithkin[0].InstanceID}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life+1 {
		t.Errorf("life %d, want %d", got, life+1)
	}
	if effectstest.BattlefieldHas(g, kithkin[0].InstanceID) {
		t.Error("the sacrificed Kithkin is still on the battlefield")
	}
	if len(oldB554TokensOf(g, me.ID, "Kithkin Soldier")) != 2 {
		t.Error("expected two Kithkin Soldiers left")
	}
}

func TestOldB554EyesOfTheWisentMakesAnElementalForAnOpponentsBlueSpellOnYourTurnOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Eyes of the Wisent", "Kindred Enchantment — Elemental", oldB554Eyes, false)
	effectstest.AdvanceToMain(t, g)
	cast := func(caster *game.Player, colors []string) {
		id := uuid.New()
		caster.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell", TypeLine: "Instant", Colors: colors, Owner: caster.ID, Controller: caster.ID})
		if err := g.CastSpell(caster.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	cast(opp, []string{"R"})
	cast(me, []string{"U"})
	if n := len(oldB554TokensOf(g, me.ID, "Elemental")); n != 0 {
		t.Fatalf("%d Elementals from spells that should not count, want 0", n)
	}
	cast(opp, []string{"U"})
	oldB554AnswerMay(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(oldB554TokensOf(g, me.ID, "Elemental")); n != 0 {
		t.Fatalf("%d Elementals after declining, want 0", n)
	}
	cast(opp, []string{"U", "R"})
	oldB554AnswerMay(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	toks := oldB554TokensOf(g, me.ID, "Elemental")
	if len(toks) != 1 {
		t.Fatalf("%d Elementals, want 1", len(toks))
	}
	if p, tt := effectstest.PtOf(t, g, toks[0].InstanceID); p != 4 || tt != 4 {
		t.Errorf("the token is %d/%d, want 4/4", p, tt)
	}
}

func TestOldB554EyesOfTheWisentIgnoresASpellOnAnotherPlayersTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Eyes of the Wisent", "Kindred Enchantment — Elemental", oldB554Eyes, false)
	effectstest.AdvanceToNextSeatsTurn(t, g)
	effectstest.AdvanceToMain(t, g)
	third := g.Seats[2]
	id := uuid.New()
	third.Hand.PushTop(game.Card{InstanceID: id, Name: "Spell", TypeLine: "Instant", Colors: []string{"U"}, Owner: third.ID, Controller: third.ID})
	if err := g.CastSpell(third.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(oldB554TokensOf(g, me.ID, "Elemental")); n != 0 {
		t.Errorf("%d Elementals from a spell cast on another player's turn, want 0", n)
	}
}

func TestOldB554BenthicoreMakesTwoMerfolkAndUntapsWithShroudForTwoUntappedMerfolk(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	core := effectstest.CastHoldCreature(t, g, "Benthicore", "Creature — Elemental", oldB554Benthicore, 5, 5)
	effectstest.PassPriorityAroundTable(t, g)
	wizards := oldB554TokensOf(g, me.ID, "Merfolk Wizard")
	if len(wizards) != 2 {
		t.Fatalf("%d Merfolk Wizard tokens, want 2", len(wizards))
	}
	effectstest.TapForTest(t, g, core)
	if err := g.ActivateCatalogAbility(me.ID, core, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{wizards[0].InstanceID}}); err == nil {
		t.Fatal("paid with one Merfolk")
	}
	if err := g.ActivateCatalogAbility(me.ID, core, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{wizards[0].InstanceID, wizards[1].InstanceID}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.TappedNow(g, core) {
		t.Error("Benthicore is still tapped")
	}
	if !effectstest.TappedNow(g, wizards[0].InstanceID) || !effectstest.TappedNow(g, wizards[1].InstanceID) {
		t.Error("the Merfolk were not tapped for the cost")
	}
	if !effectstest.HasEffectiveKeyword(t, g, core, "shroud") {
		t.Error("Benthicore has no shroud")
	}
}

func TestOldB554HearthcageGiantMakesTwoShamansAndSacrificesAnElementalToPumpAGiant(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	giant := effectstest.CastHoldCreature(t, g, "Hearthcage Giant", "Creature — Giant Warrior", oldB554Hearthcage, 5, 5)
	effectstest.PassPriorityAroundTable(t, g)
	shamans := oldB554TokensOf(g, me.ID, "Elemental Shaman")
	if len(shamans) != 2 {
		t.Fatalf("%d Elemental Shaman tokens, want 2", len(shamans))
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB554Bear, 2, 2)
	if err := g.ActivateCatalogAbility(me.ID, giant, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{shamans[0].InstanceID}, Targets: effectstest.CardTargets(bear)}); err == nil {
		t.Fatal("targeted a non-Giant")
	}
	if err := g.ActivateCatalogAbility(me.ID, giant, 0, game.ActivateAbilityParams{Strict: true,
		SacrificeIDs: []uuid.UUID{shamans[0].InstanceID}, Targets: effectstest.CardTargets(giant)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, tt := effectstest.PtOf(t, g, giant); p != 8 || tt != 6 {
		t.Errorf("the Giant is %d/%d, want 8/6", p, tt)
	}
	if effectstest.BattlefieldHas(g, shamans[0].InstanceID) {
		t.Error("the sacrificed Elemental is still on the battlefield")
	}
}

func TestOldB554MilitiasPrideMakesATappedAttackingKithkinOnlyForAPaidNontokenAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Militia's Pride", "Kindred Enchantment — Kithkin", oldB554Pride, false)
	a := effectstest.B12Creature(g, me.ID, "Attacker", oldB554Bear, 2, 2)
	token := effectstest.PushToken(g, me.ID, effects.TokenCard("1/1 white Warrior"))
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.DeclareAttack(t, g, opp.ID, a, token)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB554PayIfAsked(t, g, me, "W", false) {
		t.Fatal("no {W} offer for the nontoken attacker")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Fatal("asked again for the token attacker")
	}
	if n := len(oldB554TokensOf(g, me.ID, "Kithkin Soldier")); n != 0 {
		t.Fatalf("%d Kithkin after declining, want 0", n)
	}
}

func TestOldB554MilitiasPridePaidMakesTheTokenAttackThePlayerTheCreatureAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.PushCatalogPermanent(g, me.ID, "Militia's Pride", "Kindred Enchantment — Kithkin", oldB554Pride, false)
	a := effectstest.B12Creature(g, me.ID, "Attacker", oldB554Bear, 2, 2)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.DeclareAttack(t, g, opp.ID, a)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB554PayIfAsked(t, g, me, "W", true) {
		t.Fatal("no {W} offer")
	}
	effectstest.PassPriorityAroundTable(t, g)
	toks := oldB554TokensOf(g, me.ID, "Kithkin Soldier")
	if len(toks) != 1 {
		t.Fatalf("%d Kithkin Soldier tokens, want 1", len(toks))
	}
	if !toks[0].Tapped || toks[0].AttackingTarget != opp.ID {
		t.Errorf("the token is tapped=%v attacking %s, want tapped and attacking %s", toks[0].Tapped, toks[0].AttackingTarget, opp.ID)
	}
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("the token's arrival asked for another payment")
	}
}

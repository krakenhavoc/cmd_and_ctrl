package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b587_cards_test.go — Saviors of Kamigawa, S82 older sets, batch 587
// (#5601). Each test plays the real catalog entry.

const (
	oldB587Akuta   = "29169659-309f-4e1c-86ba-1d36420ed41e"
	oldB587Iname   = "865e569b-6999-4a03-ad02-4bfd595d5820"
	oldB587Denied  = "94f23a93-cd98-4ee7-9e62-9b7b49ace6fd"
	oldB587Exile   = "bfa21cbc-8a04-4456-bf02-bae6b514d50a"
	oldB587Goryo   = "d0e8d9f0-86e9-46b7-ba6f-2976ff9148d6"
	oldB587Spirit  = "Creature — Spirit"
	oldB587Bear    = "Creature — Bear"
	oldB587Swamp   = "Basic Land — Swamp"
	oldB587Instant = "Instant — Arcane"
)

func TestOldB587EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB587Akuta:  "Akuta, Born of Ash",
		oldB587Iname:  "Iname as One",
		oldB587Denied: "Death Denied",
		oldB587Exile:  "Exile into Darkness",
		oldB587Goryo:  "Footsteps of the Goryo",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s (%s) is not registered", name, oracle)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB587ToUpkeep starts the active seat's upkeep with `mine` cards in its hand
// and `theirs` in every other hand.
func oldB587ToUpkeep(t *testing.T, g *game.Game, mine, theirs int) {
	t.Helper()
	for i, p := range g.Seats {
		if i == 0 {
			effectstest.D2dSetHand(p, mine)
		} else {
			effectstest.D2dSetHand(p, theirs)
		}
		for j := 0; j < 3; j++ {
			effectstest.SeedLibrary(p, "Card")
		}
	}
	g.Turn.Step = game.StepUntap
	effectstest.AdvanceTo(t, g, game.StepUpkeep)
}

// --- Akuta, Born of Ash ------------------------------------------------

func TestOldB587AkutaReturnsByTheSacrificeOfASwampWhenYouHaveTheBiggestHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	akuta := effectstest.GraveyardCardOf(g, me, "Akuta, Born of Ash", "Legendary Creature — Spirit", "{2}{B}{B}", oldB587Akuta)
	swamp := effectstest.CcLand(g, me.ID, "Swamp", oldB587Swamp)
	plains := effectstest.CcLand(g, me.ID, "Plains", "Basic Land — Plains")
	oldB587ToUpkeep(t, g, 5, 4)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil || pick.ChooseMin != 0 || pick.ChooseMax != 1 || len(pick.ChooseCards) != 1 || pick.ChooseCards[0] != swamp {
		t.Fatalf("sacrifice prompt = %+v, want a may-pick over the Swamp alone", pick)
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, []uuid.UUID{swamp}); err != nil {
		t.Fatalf("ResolveOwnPermanents: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, akuta) || effects.OnBattlefield(g, swamp) || !effects.OnBattlefield(g, plains) {
		t.Errorf("Akuta on battlefield %v, Swamp still there %v, Plains there %v", effects.OnBattlefield(g, akuta), effects.OnBattlefield(g, swamp), effects.OnBattlefield(g, plains))
	}
}

func TestOldB587AkutaStaysDeadWhenYouDeclineOrDoNotHaveMoreCards(t *testing.T) {
	// Declined: the may-pick takes none.
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	akuta := effectstest.GraveyardCardOf(g, me, "Akuta, Born of Ash", "Legendary Creature — Spirit", "{2}{B}{B}", oldB587Akuta)
	swamp := effectstest.CcLand(g, me.ID, "Swamp", oldB587Swamp)
	oldB587ToUpkeep(t, g, 5, 4)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID)
	if pick == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveOwnPermanents(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, akuta) || !effects.OnBattlefield(g, swamp) {
		t.Error("declining still returned Akuta or sacrificed the Swamp")
	}

	// A tied hand is not "more than each opponent": nothing triggers.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	akuta = effectstest.GraveyardCardOf(g, me, "Akuta, Born of Ash", "Legendary Creature — Spirit", "{2}{B}{B}", oldB587Akuta)
	effectstest.CcLand(g, me.ID, "Swamp", oldB587Swamp)
	oldB587ToUpkeep(t, g, 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil || effects.OnBattlefield(g, akuta) {
		t.Error("Akuta triggered with only as many cards as an opponent")
	}
}

// --- Iname as One ------------------------------------------------------

func TestOldB587InameSearchesForASpiritPermanentOnlyWhenCastFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	effectstest.SeedLibraryTop(me, "Bear Cub", oldB587Bear, "{1}{G}")
	effectstest.SeedLibraryTop(me, "Forest Kami", oldB587Spirit, "{3}{G}")
	effectstest.SeedLibraryTop(me, "Spirit Charm", "Instant — Arcane", "{W}")
	effectstest.CastAndResolveCreature(t, g, "Iname as One", "Legendary Creature — Spirit", oldB587Iname)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatalf("no search prompt: %+v", g.PendingChoices)
	}
	for _, name := range []string{"Bear Cub", "Spirit Charm"} {
		if effectstest.SearchOptionNamed(g, c, name) != uuid.Nil {
			t.Errorf("%s was offered as a Spirit permanent card", name)
		}
	}
	effectstest.AnswerSearchNamed(t, g, me.ID, "Forest Kami")
	found := false
	for _, card := range g.Battlefield.Cards {
		if card.Name == "Forest Kami" && card.Controller == me.ID {
			found = true
		}
	}
	if !found {
		t.Error("the searched Spirit did not enter the battlefield")
	}

	// Not cast from hand: it enters with no search.
	g = effectstest.NewCatalogGame(t)
	me = g.Seats[0]
	effectstest.SeedLibraryTop(me, "Forest Kami", oldB587Spirit, "{3}{G}")
	effectstest.EnterFromHand(t, g, me.ID, "Iname as One", "Legendary Creature — Spirit", oldB587Iname)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Error("Iname searched although it was not cast")
	}
}

func TestOldB587InameDiesExilesItselfToReturnASpiritFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	kami := effectstest.GraveyardCardOf(g, me, "Dead Kami", "Creature — Spirit", "{2}", "")
	bear := effectstest.GraveyardCardOf(g, me, "Dead Bear", oldB587Bear, "{2}", "")
	iname := effectstest.PushCatalogPermanent(g, me.ID, "Iname as One", "Legendary Creature — Spirit", oldB587Iname, false)
	effectstest.KillCreature(t, g, me.ID, iname)
	effectstest.PickCard(t, g, me.ID, kami)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ConfirmForTriggerPrompt(g, me.ID)
	if prompt == nil {
		t.Fatalf("no \"you may exile it\" prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveTriggerPrompt: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, kami) {
		t.Error("the targeted Spirit did not return")
	}
	if effects.OnBattlefield(g, bear) {
		t.Error("a non-Spirit returned")
	}
	if z := g.FindCardZoneForEffect(iname); z == nil || z.Kind != game.ZoneExile {
		t.Errorf("Iname is in %v, want exile", z)
	}
}

func TestOldB587InameDeclinedStaysInTheGraveyardAndReturnsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	kami := effectstest.GraveyardCardOf(g, me, "Dead Kami", "Creature — Spirit", "{2}", "")
	iname := effectstest.PushCatalogPermanent(g, me.ID, "Iname as One", "Legendary Creature — Spirit", oldB587Iname, false)
	effectstest.KillCreature(t, g, me.ID, iname)
	effectstest.PickCard(t, g, me.ID, kami)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ConfirmForTriggerPrompt(g, me.ID)
	if prompt == nil {
		t.Fatal("no prompt")
	}
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, false); err != nil {
		t.Fatalf("decline: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, kami) {
		t.Error("the Spirit returned although Iname was not exiled")
	}
	if z := g.FindCardZoneForEffect(iname); z == nil || z.Kind != game.ZoneGraveyard {
		t.Errorf("Iname is in %v, want the graveyard", z)
	}
}

// --- Death Denied ------------------------------------------------------

func TestOldB587DeathDeniedReturnsExactlyXCreatureCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	a := effectstest.GraveyardCardOf(g, me, "Bear A", oldB587Bear, "{2}", "")
	b := effectstest.GraveyardCardOf(g, me, "Bear B", oldB587Bear, "{2}", "")
	c := effectstest.GraveyardCardOf(g, me, "Bear C", oldB587Bear, "{2}", "")
	land := effectstest.GraveyardCardOf(g, me, "Swamp", oldB587Swamp, "", "")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Death Denied", TypeLine: oldB587Instant, ManaCost: "{X}{B}{B}", OracleID: oldB587Denied, Owner: me.ID, Controller: me.ID})
	me.ManaPool.AddMana(game.ManaToken{Color: "B"}, game.ManaToken{Color: "B"}, game.ManaToken{Color: "C"}, game.ManaToken{Color: "C"})
	refs := func(ids ...uuid.UUID) []game.TargetRef {
		var out []game.TargetRef
		for _, i := range ids {
			out = append(out, game.TargetRef{Kind: game.TargetCard, ID: i})
		}
		return out
	}
	// A land card is not a legal target, and one target short of X is refused.
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Strict: true, Targets: refs(a, land)}); err == nil {
		t.Fatal("a land card was accepted as a target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Strict: true, Targets: refs(a)}); err == nil {
		t.Fatal("one target was accepted for X = 2")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 2, Strict: true, Targets: refs(a, b)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, want := range []uuid.UUID{a, b} {
		if !me.Hand.Contains(want) {
			t.Error("a targeted creature card did not return to hand")
		}
	}
	if me.Hand.Contains(c) || me.Hand.Contains(land) {
		t.Error("an untargeted card returned")
	}
}

// --- Exile into Darkness -----------------------------------------------

func TestOldB587ExileIntoDarknessMakesTheTargetSacrificeACreatureOfManaValueThreeOrLess(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Small", TypeLine: oldB587Bear, ManaCost: "{2}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	small2 := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Small Two", TypeLine: oldB587Bear, ManaCost: "{3}", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	big := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Big", TypeLine: oldB587Bear, ManaCost: "{5}", Power: 5, Toughness: 5, Owner: opp.ID, Controller: opp.ID})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Exile into Darkness", TypeLine: "Sorcery", OracleID: oldB587Exile, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.SacrificeChoiceFor(g, opp.ID)
	if c == nil {
		t.Fatalf("no sacrifice prompt: %+v", g.PendingChoices)
	}
	opts := effectstest.SacrificeOptions(c)
	if !effectstest.HasID(opts, small) || !effectstest.HasID(opts, small2) || effectstest.HasID(opts, big) {
		t.Fatalf("options %v: want the two small creatures and not the big one", opts)
	}
	effectstest.AnswerSacrifice(t, g, opp.ID, small2)
	if effects.OnBattlefield(g, small2) || !effects.OnBattlefield(g, small) || !effects.OnBattlefield(g, big) {
		t.Error("the wrong creature was sacrificed")
	}
}

func TestOldB587ExileIntoDarknessComesBackFromTheGraveyardWithTheBiggestHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	card := effectstest.GraveyardCardOf(g, me, "Exile into Darkness", "Sorcery", "{4}{B}", oldB587Exile)
	oldB587ToUpkeep(t, g, 5, 4)
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ConfirmForTriggerPrompt(g, me.ID)
	if prompt == nil {
		t.Fatalf("no \"you may return\" prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveTriggerPrompt(prompt.ID, me.ID, true); err != nil {
		t.Fatalf("accept: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(card) {
		t.Error("the card did not return to hand")
	}
}

// --- Footsteps of the Goryo ---------------------------------------------

func TestOldB587FootstepsReturnsACreatureAndSacrificesItAtTheNextEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[0]
	dead := effectstest.GraveyardCardOf(g, me, "Dead Bear", oldB587Bear, "{2}", "")
	land := effectstest.GraveyardCardOf(g, me, "Swamp", oldB587Swamp, "", "")
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Footsteps of the Goryo", TypeLine: "Sorcery — Arcane", OracleID: oldB587Goryo, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: land}}}); err == nil {
		t.Fatal("a land card was accepted as the target")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: dead}}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effects.OnBattlefield(g, dead) {
		t.Fatal("the creature card did not return")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, dead) {
		t.Error("the creature was not sacrificed at the next end step")
	}
	if z := g.FindCardZoneForEffect(dead); z == nil || z.Kind != game.ZoneGraveyard {
		t.Errorf("the sacrificed creature is in %v, want the graveyard", z)
	}
}

package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	rprB15MistvaultOracle  = "33ee23bc-6327-4a54-a704-dfd83be36bb5"
	rprB15MuzzioOracle     = "fc2768a5-621b-48de-9160-4dbaeddb3de1"
	rprB15SpellbombOracle  = "b713e49f-1b13-42d1-91f2-cc7a579e7614"
	rprB15StoneOracle      = "2fd61c9f-44d2-48d7-a271-8e4208407cca"
	rprB15RelicOracle      = "6a9c3401-570c-41cb-a605-75dfa57d6ab7"
	rprB15SchematicOracle  = "6b8c6104-c537-4025-83e2-e89c8668f3ba"
	rprB15SharuumOracle    = "037e7fc9-3aa6-484c-a2c8-43009e45f1d8"
	rprB15SilasOracle      = "0dc78b1e-581e-4f28-bf35-e2082553d6a9"
	rprB15SlobadOracle     = "1f2d6cf4-b8f4-440f-ac08-ca47ada868ac"
	rprB15CaptainsOracle   = "46418fe4-065c-4dfa-b796-eee02c14f351"
	rprB15DragonWhelpOracl = "705a1985-ed39-4a4b-812e-a677170b596e"
)

func rprB15Graveyard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, Owner: p.ID, Controller: p.ID})
	return id
}

func TestRprB15EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		rprB15MistvaultOracle:  "Mistvault Bridge",
		rprB15MuzzioOracle:     "Muzzio, Visionary Architect",
		rprB15SpellbombOracle:  "Nihil Spellbomb",
		rprB15StoneOracle:      "Oblivion Stone",
		rprB15RelicOracle:      "Relic of Progenitus",
		rprB15SchematicOracle:  "Servo Schematic",
		rprB15SharuumOracle:    "Sharuum the Hegemon",
		rprB15SilasOracle:      "Silas Renn, Seeker Adept",
		rprB15SlobadOracle:     "Slobad, Goblin Tinkerer",
		rprB15CaptainsOracle:   "Captain's Call",
		rprB15DragonWhelpOracl: "Dragon Whelp",
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

func TestRprB15MistvaultBridgeEntersTappedIndestructibleAndTapsForEither(t *testing.T) {
	for _, colour := range []string{"U", "B"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		land := effectstest.B12PlayFromHand(t, g, "Mistvault Bridge", "Artifact Land", rprB15MistvaultOracle, game.CastSpellParams{})
		if !effectstest.B16Tapped(t, g, land) {
			t.Fatal("Mistvault Bridge enters tapped")
		}
		effectstest.AssertKeywords(t, g, land, "indestructible")
		g.WithWriteLock(func() { _ = g.UntapTargetForEffect(land) })
		effectstest.B28TapForMana(t, g, me.ID, land, colour)
		if got := effectstest.PoolColors(me); len(got) != 1 || got[0] != colour {
			t.Errorf("tapped for {%s}: pool %v", colour, got)
		}
	}
}

func TestRprB15CaptainsCallMakesThreeSoldiers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Captain's Call", "Sorcery", rprB15CaptainsOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Soldier", me.ID); n != 3 {
		t.Errorf("Soldier tokens: %d, want 3", n)
	}
}

func TestRprB15ServoSchematicMakesAServoWhenItEntersAndWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.CastCatalogSpell(t, g, "Servo Schematic", "Artifact", rprB15SchematicOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Servo", me.ID); n != 1 {
		t.Fatalf("Servos after entering: %d, want 1", n)
	}
	var schematic uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Servo Schematic" {
			schematic = c.InstanceID
		}
	}
	effectstest.B18Kill(t, g, schematic)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Servo", me.ID); n != 2 {
		t.Errorf("Servos after it went to the graveyard: %d, want 2", n)
	}
}

func TestRprB15ServoSchematicExiledMakesNoServo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sch := effectstest.B12Push(g, me.ID, "Servo Schematic", "Artifact", rprB15SchematicOracle, 0, 0)
	g.WithWriteLock(func() { _ = g.ExileCardForEffect(sch) })
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.TokensNamedControlledBy(g, "Servo", me.ID); n != 0 {
		t.Errorf("an exiled Schematic made %d Servos", n)
	}
}

func TestRprB15NihilSpellbombExilesAGraveyardAndDrawsOnlyIfYouPayB(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rprB15Graveyard(opp, "Old Bear", "Creature — Bear")
	rprB15Graveyard(opp, "Old Elk", "Creature — Elk")
	bomb := effectstest.B12Push(g, me.ID, "Nihil Spellbomb", "Artifact", rprB15SpellbombOracle, 0, 0)
	hand := me.Hand.Size()
	if err := g.ActivateCatalogAbility(me.ID, bomb, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !me.Graveyard.Contains(bomb) {
		t.Fatal("the sacrifice is a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != 0 {
		t.Errorf("opponent graveyard %d, want 0", opp.Graveyard.Size())
	}
	// The dies trigger asks for {B}: pay.
	effectstest.B06AddMana(me, "B")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("paid {B}: hand %d, want %d", got, hand+1)
	}

	// Declining draws nothing.
	bomb2 := effectstest.B12Push(g, me.ID, "Nihil Spellbomb", "Artifact", rprB15SpellbombOracle, 0, 0)
	hand = me.Hand.Size()
	effectstest.B06AddMana(me, "B")
	effectstest.B18Kill(t, g, bomb2)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("declined: hand %d, want %d", got, hand)
	}
}

func TestRprB15RelicOfProgenitusTargetPlayerChoosesTheCardAndTheSecondAbilityExilesAll(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	keep := rprB15Graveyard(opp, "Old Bear", "Creature — Bear")
	gone := rprB15Graveyard(opp, "Old Elk", "Creature — Elk")
	mine := rprB15Graveyard(me, "My Elk", "Creature — Elk")
	relic := effectstest.B12Push(g, me.ID, "Relic of Progenitus", "Artifact", rprB15RelicOracle, 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, relic, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsPromptFrom(g, relic)
	if prompt == nil || prompt.Chooser != opp.ID {
		t.Fatalf("the target player chooses: %+v", prompt)
	}
	if err := g.ResolveChooseCards(prompt.ID, opp.ID, []uuid.UUID{gone}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(gone) || !opp.Graveyard.Contains(keep) {
		t.Fatalf("exactly the chosen card should be exiled: exile %v keep %v", g.Exile.Contains(gone), opp.Graveyard.Contains(keep))
	}

	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(relic) })
	hand := me.Hand.Size()
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, relic, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate second: %v", err)
	}
	if !g.Exile.Contains(relic) {
		t.Fatal("exiling the Relic is a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Graveyard.Size() != 0 || me.Graveyard.Contains(mine) {
		t.Error("every graveyard, yours included, is exiled")
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("hand %d, want %d", got, hand+1)
	}
}

func TestRprB15ObliviionStoneWipesTheUnmarkedAndClearsTheCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	stone := effectstest.B12Push(g, me.ID, "Oblivion Stone", "Artifact", rprB15StoneOracle, 0, 0)
	marked := effectstest.B12Creature(g, me.ID, "Marked Bear", "Creature — Bear", 2, 2)
	doomed := effectstest.B12Creature(g, opp.ID, "Doomed Ogre", "Creature — Ogre", 3, 3)
	land := effectstest.B12Push(g, opp.ID, "Plains", "Basic Land — Plains", "", 0, 0)
	effectstest.B06AddMana(me, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, stone, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: marked}},
	}); err != nil {
		t.Fatalf("fate counter: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.BattlefieldCountersOf(g, marked, "fate"); got != 1 {
		t.Fatalf("fate counters on the marked creature: %d, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(stone) })
	effectstest.B06AddMana(me, "C", "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, stone, 1, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("wipe: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(marked) {
		t.Error("the creature with a fate counter survives")
	}
	if g.Battlefield.Contains(doomed) {
		t.Error("the unmarked creature is destroyed")
	}
	if !g.Battlefield.Contains(land) {
		t.Error("lands are not destroyed")
	}
	if got := effectstest.BattlefieldCountersOf(g, marked, "fate"); got != 0 {
		t.Errorf("fate counters should all be removed, %d left", got)
	}
}

func TestRprB15SharuumReturnsAnArtifactCardFromYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	art := rprB15Graveyard(me, "Old Signet", "Artifact")
	rprB15Graveyard(me, "Old Bear", "Creature — Bear")
	sh := castCatalogSpellStrictlyTargeting(t, g, art)
	_ = sh
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(art) {
		t.Error("the artifact card did not return to the battlefield")
	}
	if effectstest.B16Tapped(t, g, art) {
		t.Error("it returns untapped")
	}
}

func castCatalogSpellStrictlyTargeting(t *testing.T, g *game.Game, target uuid.UUID) uuid.UUID {
	t.Helper()
	id := effectstest.CastAndResolveCreature(t, g, "Sharuum the Hegemon", "Legendary Artifact Creature — Sphinx", rprB15SharuumOracle)
	if p := effectstest.LatestPickTarget(g, g.Seats[0].ID); p != nil {
		effectstest.AnswerFirstPickTarget(t, g)
	}
	return id
}

func TestRprB15SlobadSacrificesAnArtifactToProtectAnother(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	slobad := effectstest.B12Push(g, me.ID, "Slobad, Goblin Tinkerer", "Legendary Creature — Goblin Artificer", rprB15SlobadOracle, 1, 2)
	fodder := effectstest.B12Push(g, me.ID, "Fodder", "Artifact", "", 0, 0)
	keeper := effectstest.B12Push(g, me.ID, "Keeper", "Artifact", "", 0, 0)
	if err := g.ActivateCatalogAbility(me.ID, slobad, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: keeper}},
		SacrificeIDs: []uuid.UUID{fodder},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed artifact is gone")
	}
	effectstest.AssertKeywords(t, g, keeper, "indestructible")
	// A non-artifact is not a legal target.
	if err := g.ActivateCatalogAbility(me.ID, slobad, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: slobad}},
	}); err == nil {
		t.Error("a nonartifact creature is not a legal target")
	}
}

func TestRprB15DragonWhelpSacrificesAtEndStepAfterFourActivations(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	whelp := effectstest.B12Push(g, me.ID, "Dragon Whelp", "Creature — Dragon", rprB15DragonWhelpOracl, 2, 3)
	for i := 1; i <= 4; i++ {
		effectstest.B06AddMana(me, "R")
		if err := g.ActivateCatalogAbility(me.ID, whelp, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if p, _ := effectstest.EffectivePTOf(t, g, whelp); p != 6 {
		t.Errorf("power %d, want 6", p)
	}
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(whelp) {
		t.Error("the fourth activation sacrifices the Whelp at the next end step")
	}
}

func TestRprB15DragonWhelpSurvivesThreeActivations(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	whelp := effectstest.B12Push(g, me.ID, "Dragon Whelp", "Creature — Dragon", rprB15DragonWhelpOracl, 2, 3)
	for i := 0; i < 3; i++ {
		effectstest.B06AddMana(me, "R")
		if err := g.ActivateCatalogAbility(me.ID, whelp, 0, game.ActivateAbilityParams{}); err != nil {
			t.Fatalf("activation: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	for g.Turn.Step != game.StepEnd {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(whelp) {
		t.Error("three activations do not sacrifice the Whelp")
	}
}

func TestRprB15MuzzioPutsAnArtifactFromTheTopOntoTheBattlefield(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	muzzio := effectstest.B12Push(g, me.ID, "Muzzio, Visionary Architect", "Legendary Creature — Human Artificer", rprB15MuzzioOracle, 1, 3)
	effectstest.PushPermanent(g, me.ID, game.Card{Name: "Three Drop", TypeLine: "Artifact", ManaCost: "{3}"})
	top := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: top, Name: "Library Gizmo", TypeLine: "Artifact", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	effectstest.B06AddMana(me, "C", "C", "C", "U")
	if err := g.ActivateCatalogAbility(me.ID, muzzio, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	prompt := effectstest.ChooseCardsPromptFrom(g, muzzio)
	if prompt == nil {
		t.Fatal("no pick offered")
	}
	if err := g.ResolveChooseCards(prompt.ID, me.ID, []uuid.UUID{top}); err != nil {
		t.Fatalf("choose: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(top) {
		t.Error("the chosen artifact card did not enter")
	}
}

func TestRprB15SilasRennGrantsACastPermissionOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	art := rprB15Graveyard(me, "Old Signet", "Artifact")
	silas := effectstest.B12Push(g, me.ID, "Silas Renn, Seeker Adept", "Legendary Artifact Creature — Human", rprB15SilasOracle, 2, 2)
	effectstest.DealCombatDamageToPlayer(g, silas, opp.ID, 2)
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil {
		effectstest.AnswerFirstPickTarget(t, g)
		effectstest.PassPriorityAroundTable(t, g)
	}
	for g.Turn.Step != game.StepPrecombatMain && g.Turn.Step != game.StepPostcombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, art, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err != nil {
		t.Fatalf("casting the granted card: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(art) {
		t.Error("the artifact cast from the graveyard did not resolve onto the battlefield")
	}
}

func TestRprB15SilasRennGrantsNothingWithoutTheDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	art := rprB15Graveyard(me, "Old Signet", "Artifact")
	effectstest.B12Push(g, me.ID, "Silas Renn, Seeker Adept", "Legendary Artifact Creature — Human", rprB15SilasOracle, 2, 2)
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if err := g.CastSpell(me.ID, art, game.CastSpellParams{FromZone: "graveyard", Strict: true}); err == nil {
		t.Error("no permission without the combat damage")
	}
}

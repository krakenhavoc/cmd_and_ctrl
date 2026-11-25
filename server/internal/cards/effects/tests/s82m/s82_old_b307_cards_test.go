package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b307_cards_test.go — Battle for Zendikar Devoid cards, batch 307
// (#5547). Each test plays the real catalog entry and asserts what its text
// says.

const (
	oldB307Salvage   = "9b18bbf3-3b1c-4010-bc6f-e7eacb2ffb70"
	oldB307Awry      = "9cac23a1-a0b3-490c-aea2-dc2928f6dc9e"
	oldB307Butcher   = "a20edb8d-219d-4d0f-8346-0c31064eb49f"
	oldB307Grip      = "a7c8d026-b5b0-4912-a4a0-4c35df5f8899"
	oldB307Cruiser   = "ae7cfdb8-b0f2-46d7-b847-9a760ddbe028"
	oldB307Birthing  = "b25962ab-21b5-47f9-bdef-35b911b45671"
	oldB307Benthic   = "b2977807-5f46-424a-a6e9-14262f8a723a"
	oldB307Skyspawn  = "b5661289-a03c-40fc-a1e4-f602adb56f81"
	oldB307Watcher   = "ba047b0b-d451-4566-8656-00b00611de8c"
	oldB307Culling   = "bda0ca68-78a6-4b32-842f-0bee9a9c84ef"
	oldB307Shrivel   = "bfaf376f-90f9-45d9-bfbe-dd84a2a4688b"
	oldB307OpenSpell = "Creature — Bear"
)

func TestOldB307CardsAreRegisteredAndWhole(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB307Salvage: "Salvage Drone", oldB307Awry: "Horribly Awry", oldB307Butcher: "Brood Butcher",
		oldB307Grip: "Grip of Desolation", oldB307Cruiser: "Cryptic Cruiser", oldB307Birthing: "Grave Birthing",
		oldB307Benthic: "Benthic Infiltrator", oldB307Skyspawn: "Eldrazi Skyspawner", oldB307Watcher: "Eyeless Watcher",
		oldB307Culling: "Culling Drone", oldB307Shrivel: "Spell Shrivel",
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

func oldB307Scions(g *game.Game, owner uuid.UUID) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == owner && c.IsToken() && c.HasSubtype("Scion") {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func TestOldB307IngestCreaturesExileTheTopCardAndBenthicCantBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	culling := effectstest.B12Push(g, me.ID, "Culling Drone", "Creature — Eldrazi Drone", oldB307Culling, 2, 2)
	benthic := effectstest.B12Push(g, me.ID, "Benthic Infiltrator", "Creature — Eldrazi Drone", oldB307Benthic, 1, 4)
	wall := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
	effectstest.SeedLibrary(opp, "a", "b", "c", "d")
	lib := opp.Library.Size()
	effectstest.DeclareAttack(t, g, opp.ID, culling, benthic)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(wall, benthic); err == nil {
		t.Error("Benthic Infiltrator was blocked")
	}
	if err := g.DeclareBlocker(wall, culling); err != nil {
		t.Fatalf("the Wall could not block Culling Drone: %v", err)
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != lib-1 {
		t.Errorf("library %d -> %d, want one card exiled (only Benthic connected)", lib, opp.Library.Size())
	}
}

func TestOldB307SalvageDroneLootsWhenItDiesIfYouChoose(t *testing.T) {
	for _, yes := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		drone := effectstest.B12Push(g, me.ID, "Salvage Drone", "Creature — Eldrazi Drone", oldB307Salvage, 1, 1)
		effectstest.SeedLibrary(me, "x", "y")
		junk := effectstest.B13HandCard(me, "Junk", "Sorcery", "")
		hand := me.Hand.Size()
		effectstest.KillCreature(t, g, me.ID, drone)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, me.ID, yes)
		effectstest.PassPriorityAroundTable(t, g)
		if yes {
			if effectstest.DiscardChoiceFor(g, me.ID) != nil {
				effectstest.AnswerDiscard(t, g, me.ID, junk)
			}
			effectstest.PassPriorityAroundTable(t, g)
		}
		if me.Hand.Size() != hand {
			t.Errorf("yes=%v: hand %d -> %d, want no change (a draw and a discard, or neither)", yes, hand, me.Hand.Size())
		}
		if yes && !me.Graveyard.Contains(junk) {
			t.Error("the discard did not happen")
		}
	}
}

func TestOldB307HorriblyAwryExilesACheapCreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	cheap := effectstest.SocB17Hand(me, "Bear", oldB307OpenSpell, "{2}{G}", "")
	if err := g.CastSpell(me.ID, cheap, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.CastFromSeat(t, g, opp, "Horribly Awry", "Instant", oldB307Awry, effectstest.CardTargets(cheap))
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.ZoneOfCard(g, cheap); z != game.ZoneExile {
		t.Errorf("the countered spell is in %v, want exile", z)
	}
}

func TestOldB307HorriblyAwryRefusesBigSpellsAndNoncreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	big := effectstest.SocB17Hand(me, "Big Bear", oldB307OpenSpell, "{4}{G}", "")
	if err := g.CastSpell(me.ID, big, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	card := game.Card{InstanceID: uuid.New(), Name: "Horribly Awry", TypeLine: "Instant", OracleID: oldB307Awry, Owner: opp.ID, Controller: opp.ID}
	opp.Hand.PushTop(card)
	if err := g.CastSpell(opp.ID, card.InstanceID, game.CastSpellParams{Strict: true, Targets: effectstest.CardTargets(big)}); err == nil {
		t.Error("Horribly Awry targeted a mana value 5 creature spell")
	}
}

func TestOldB307SpellShrivelCountersAndExilesUnlessTheirControllerPaysFour(t *testing.T) {
	for _, pay := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.AdvanceToMain(t, g)
		spell := effectstest.SocB17Hand(me, "Sorcery", "Sorcery", "", "")
		if err := g.CastSpell(me.ID, spell, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast: %v", err)
		}
		effectstest.CastFromSeat(t, g, opp, "Spell Shrivel", "Instant", oldB307Shrivel, effectstest.CardTargets(spell))
		effectstest.PassUntilTaxed(t, g, me.ID)
		if pay {
			effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
		}
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		z := effectstest.ZoneOfCard(g, spell)
		if pay && z == game.ZoneExile {
			t.Error("a spell whose controller paid {4} was exiled")
		}
		if !pay && z != game.ZoneExile {
			t.Errorf("the unpaid spell is in %v, want exile", z)
		}
	}
}

func TestOldB307BroodButcherMakesAScionAndSacrificesForMinusTwo(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	butcher := effectstest.CastAndResolveCreature(t, g, "Brood Butcher", "Creature — Eldrazi Drone", oldB307Butcher)
	effectstest.PassPriorityAroundTable(t, g)
	scions := oldB307Scions(g, me.ID)
	if len(scions) != 1 {
		t.Fatalf("%d Scions, want 1", len(scions))
	}
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 3, 3)
	effectstest.FloatForTest(g, me, "BG")
	if err := g.ActivateCatalogAbility(me.ID, butcher, 0, game.ActivateAbilityParams{
		Targets: effectstest.CardRefs(victim), SacrificeIDs: []uuid.UUID{scions[0]}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(scions[0]) {
		t.Error("the Scion was not sacrificed")
	}
	if p, tough := effectstest.EffectivePower(t, g, victim), effectstest.EffectiveToughness(t, g, victim); p != 1 || tough != 1 {
		t.Errorf("target is %d/%d, want 1/1", p, tough)
	}
}

func TestOldB307SkyspawnerAndWatcherMakeScions(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	sky := effectstest.CastAndResolveCreature(t, g, "Eldrazi Skyspawner", "Creature — Eldrazi Drone", oldB307Skyspawn)
	effectstest.PassPriorityAroundTable(t, g)
	if len(oldB307Scions(g, me.ID)) != 1 {
		t.Errorf("%d Scions after the Skyspawner, want 1", len(oldB307Scions(g, me.ID)))
	}
	if !effectstest.HasEffectiveKeyword(t, g, sky, "flying") {
		t.Error("Skyspawner has no flying")
	}
	effectstest.CastAndResolveCreature(t, g, "Eyeless Watcher", "Creature — Eldrazi Drone", oldB307Watcher)
	effectstest.PassPriorityAroundTable(t, g)
	if len(oldB307Scions(g, me.ID)) != 3 {
		t.Errorf("%d Scions after the Watcher, want 3", len(oldB307Scions(g, me.ID)))
	}
}

func TestOldB307GripOfDesolationExilesACreatureAndALand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	bear := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	land := effectstest.PushPermanentForTest(g, opp.ID, "Forest", "", "Basic Land — Forest")
	other := effectstest.PushPermanentForTest(g, opp.ID, "Forest", "", "Basic Land — Forest")
	effectstest.CastCatalogSpell(t, g, "Grip of Desolation", "Instant", oldB307Grip, effectstest.CardRefs(bear, land))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) || g.Battlefield.Contains(land) {
		t.Error("a target is still on the battlefield")
	}
	if !g.Exile.Contains(bear) || !g.Exile.Contains(land) {
		t.Error("a target was not exiled")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("an untargeted land was exiled")
	}
}

func TestOldB307GripOfDesolationRefusesANonlandSecondTarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[1]
	a := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Elk", "Creature — Elk", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Grip of Desolation", "Instant", oldB307Grip, effectstest.CardRefs(a, b)); err == nil {
		t.Error("two creatures were accepted as a creature and a land")
	}
}

func TestOldB307CrypticCruiserProcessesACardToTapACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	cruiser := effectstest.B12Push(g, me.ID, "Cryptic Cruiser", "Creature — Eldrazi Processor", oldB307Cruiser, 3, 3)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	exiled := uuid.New()
	g.Exile.PushTop(game.Card{InstanceID: exiled, Name: "Exiled Card", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
	effectstest.FloatForTest(g, me, "CCU")
	if err := g.ActivateCatalogAbility(me.ID, cruiser, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)}); err == nil {
		t.Fatal("activated without naming a card to process")
	}
	if err := g.ActivateCatalogAbility(me.ID, cruiser, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim), ExileIDs: []uuid.UUID{exiled}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if !opp.Graveyard.Contains(exiled) {
		t.Error("the processed card did not go to its owner's graveyard")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, victim) {
		t.Error("the target creature was not tapped")
	}
}

func TestOldB307GraveBirthingExilesFromTheirGraveyardMakesAScionAndDraws(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := effectstest.PushGraveyardPermanent(opp, "Their Dead", "Creature — Bear", "{1}{G}")
	effectstest.SeedLibrary(me, "x", "y")
	hand := me.Hand.Size()
	effectstest.CastCatalogSpell(t, g, "Grave Birthing", "Instant", oldB307Birthing, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if prompt := effectstest.ChooseCardsChoiceFor(g, opp.ID); prompt != nil {
		if err := g.ResolveChooseCards(prompt.ID, opp.ID, []uuid.UUID{dead}); err != nil {
			t.Fatalf("pick: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
	}
	if !g.Exile.Contains(dead) {
		t.Error("their graveyard card was not exiled")
	}
	if len(oldB307Scions(g, me.ID)) != 1 {
		t.Errorf("%d Scions, want 1", len(oldB307Scions(g, me.ID)))
	}
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d -> %d, want a card drawn", hand, me.Hand.Size())
	}
}

func TestOldB307GraveBirthingStillMakesAScionWithAnEmptyGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.SeedLibrary(me, "x", "y")
	effectstest.CastCatalogSpell(t, g, "Grave Birthing", "Instant", oldB307Birthing, effectstest.PlayerRefs(opp.ID))
	effectstest.PassPriorityAroundTable(t, g)
	if len(oldB307Scions(g, me.ID)) != 1 {
		t.Errorf("%d Scions, want 1", len(oldB307Scions(g, me.ID)))
	}
}

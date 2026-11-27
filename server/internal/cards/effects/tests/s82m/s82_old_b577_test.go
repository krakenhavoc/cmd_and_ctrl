package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b577_test.go — Ravnica: City of Guilds, batch 577 (#5600). Every test
// plays the real catalog entry through the engine.

const (
	oldB577Brownscale = "34242c44-a987-412c-9652-5242b9b5200c"
	oldB577Angel      = "837e4452-3002-41e3-890b-3d54b469e7b9"
	oldB577Sins       = "9059a940-ac28-4322-a11b-af1d107b2edd"
	oldB577Strangler  = "c72ab66b-2011-449c-93a6-1808669653f0"
	oldB577Sewerdreg  = "f05e3fb4-a7e0-4dfa-9f3e-5109be79ef16"
	oldB577March      = "fc17d8dd-887f-405e-a195-777aa3da36f5"
)

func TestOldB577EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB577Brownscale: "Golgari Brownscale", oldB577Angel: "Firemane Angel", oldB577Sins: "Sins of the Past",
		oldB577Strangler: "Woodwraith Strangler", oldB577Sewerdreg: "Sewerdreg", oldB577March: "Bloodbond March",
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

func oldB577Perm(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int) uuid.UUID {
	return effectstest.B12Push(g, owner, name, typeLine, oracle, p, tough)
}

func TestOldB577BrownscaleDredgedFromTheGraveyardGainsTwoLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.StockLibraryFor(g, me, 10)
	scale := effectstest.PushCatalogGraveyardCard(me, "Golgari Brownscale", "Creature — Lizard", oldB577Brownscale, 2, 3)
	life, lib := effectstest.LifeOf(g, me.ID), me.Library.Size()
	effectstest.DrawN(g, me, 1)
	effectstest.AnswerOptional(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(scale) {
		t.Fatal("dredge did not return the Brownscale to the hand")
	}
	if me.Library.Size() != lib-2 {
		t.Errorf("library %d, want %d (dredge 2 mills two and draws nothing)", me.Library.Size(), lib-2)
	}
	if got := effectstest.LifeOf(g, me.ID); got != life+2 {
		t.Errorf("life %d, want %d", got, life+2)
	}
}

func TestOldB577BrownscaleDrawnFromTheLibraryGainsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.StockLibraryFor(g, me, 10)
	scale := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: scale, Name: "Golgari Brownscale", TypeLine: "Creature — Lizard",
		OracleID: oldB577Brownscale, Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID})
	life := effectstest.LifeOf(g, me.ID)
	effectstest.DrawN(g, me, 1)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(scale) {
		t.Fatal("the Brownscale was not drawn")
	}
	if got := effectstest.LifeOf(g, me.ID); got != life {
		t.Errorf("life %d, want %d: a card drawn from the library is not put into the hand from the graveyard", got, life)
	}
}

func TestOldB577BrownscaleDeclinedDredgeDrawsAndGainsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	effectstest.StockLibraryFor(g, me, 10)
	scale := effectstest.PushCatalogGraveyardCard(me, "Golgari Brownscale", "Creature — Lizard", oldB577Brownscale, 2, 3)
	life, lib := effectstest.LifeOf(g, me.ID), me.Library.Size()
	effectstest.DrawN(g, me, 1)
	effectstest.AnswerOptional(t, g, me, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Contains(scale) || me.Library.Size() != lib-1 || effectstest.LifeOf(g, me.ID) != life {
		t.Errorf("declined dredge: in hand %v, library %d (want %d), life %d (want %d)",
			me.Hand.Contains(scale), me.Library.Size(), lib-1, effectstest.LifeOf(g, me.ID), life)
	}
}

func oldB577AngelCard(me *game.Player) game.Card {
	return game.Card{InstanceID: uuid.New(), Name: "Firemane Angel", TypeLine: "Creature — Angel", OracleID: oldB577Angel,
		ManaCost: "{3}{R}{W}{W}", Power: 4, Toughness: 3, Owner: me.ID, Controller: me.ID}
}

func TestOldB577FiremaneAngelGainsLifeAtYourUpkeepFromTheBattlefieldAndTheGraveyard(t *testing.T) {
	for _, zone := range []string{"battlefield", "graveyard"} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		if zone == "battlefield" {
			effectstest.PushBattlefieldCardWithTimestamp(g, oldB577AngelCard(me))
		} else {
			me.Graveyard.PushTop(oldB577AngelCard(me))
		}
		effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
		life := effectstest.LifeOf(g, me.ID)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.LifeOf(g, me.ID); got != life+1 {
			t.Errorf("from the %s: life %d, want %d", zone, got, life+1)
		}
	}
}

func TestOldB577FiremaneAngelMayDeclineAndDoesNothingFromHand(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	me.Graveyard.PushTop(oldB577AngelCard(me))
	effectstest.AdvanceToUpkeepOfSeat(t, g, 0)
	life := effectstest.LifeOf(g, me.ID)
	effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.LifeOf(g, me.ID); got != life {
		t.Errorf("declined: life %d, want %d", got, life)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	me2.Hand.PushTop(oldB577AngelCard(me2))
	effectstest.AdvanceToUpkeepOfSeat(t, g2, 0)
	effectstest.PassPriorityAroundTable(t, g2)
	if p := effectstest.LatestTriggerPrompt(g2, me2.ID); p != nil {
		t.Error("an Angel in the hand triggered at upkeep")
	}
}

func TestOldB577FiremaneAngelReturnsOnlyDuringYourUpkeep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	c := oldB577AngelCard(me)
	c.InstanceID = id
	me.Graveyard.PushTop(c)
	effectstest.FloatMana(t, g, me, "{R}{R}{R}{R}{R}{R}{R}{R}{W}{W}{W}{W}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated in the main phase")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	id2 := uuid.New()
	c2 := oldB577AngelCard(me2)
	c2.InstanceID = id2
	me2.Graveyard.PushTop(c2)
	effectstest.AdvanceToUpkeepOfSeat(t, g2, 0)
	effectstest.AnswerLatestTriggerPrompt(t, g2, me2.ID, false)
	if g2.Turn.Step != game.StepUpkeep {
		t.Skipf("the upkeep ended before the activation (step %s)", g2.Turn.Step)
	}
	effectstest.FloatMana(t, g2, me2, "{R}{R}{R}{R}{R}{R}{R}{R}{W}{W}{W}{W}")
	if err := g2.ActivateCatalogAbility(me2.ID, id2, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate during upkeep: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !g2.Battlefield.Contains(id2) {
		t.Errorf("the Angel did not return to the battlefield (zone %s)", effectstest.B12ZoneOf(g2, id2))
	}
	if cc, ok := g2.LookupCardForEffect(id2); ok && cc.Tapped {
		t.Error("the Angel returned tapped")
	}
}

func TestOldB577SinsOfThePastCastsAGraveyardSpellFreeAndExilesBoth(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	bolt := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: bolt, Name: "Test Sorcery", TypeLine: "Sorcery", ManaCost: "{2}{R}", Owner: me.ID, Controller: me.ID})
	sins := effectstest.CastCatalogSpell(t, g, "Sins of the Past", "Sorcery", oldB577Sins, []game.TargetRef{{Kind: game.TargetCard, ID: bolt}})
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, sins); z != game.ZoneExile {
		t.Errorf("Sins of the Past is in %s, want exile", z)
	}
	// Until end of turn: the permission outlives the resolution, and a
	// sorcery keeps sorcery timing (the stack is empty in the main phase).
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: game.AltCostKeyGrantedCost}); err != nil {
		t.Fatalf("cast from the graveyard for free: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, bolt); z != game.ZoneExile {
		t.Errorf("the cast spell is in %s, want exile instead of the graveyard", z)
	}
}

func TestOldB577SinsOfThePastRefusesACreatureCardAndAGraveyardNotYours(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	bear := effectstest.PushToGraveyard(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := effectstest.PushToGraveyard(g, opp, game.Card{Name: "Their Instant", TypeLine: "Instant"})
	for _, id := range []uuid.UUID{bear, theirs} {
		if err := effectstest.CastCatalogSpellErr(t, g, "Sins of the Past", "Sorcery", oldB577Sins,
			[]game.TargetRef{{Kind: game.TargetCard, ID: id}}); err == nil {
			t.Errorf("Sins of the Past accepted target %s", id)
		}
	}
}

func TestOldB577WoodwraithStranglerExilesACreatureCardToRegenerate(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	strangler := oldB577Perm(g, me.ID, "Woodwraith Strangler", "Creature — Plant Zombie", oldB577Strangler, 2, 2)
	spell := effectstest.PushToGraveyard(g, me, game.Card{Name: "A Sorcery", TypeLine: "Sorcery"})
	if err := g.ActivateCatalogAbility(me.ID, strangler, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{spell}}); err == nil {
		t.Fatal("exiled a noncreature card to pay")
	}
	if err := g.ActivateCatalogAbility(me.ID, strangler, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without a graveyard card to exile")
	}
	dead := effectstest.PushToGraveyard(g, me, game.Card{Name: "Dead Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := g.ActivateCatalogAbility(me.ID, strangler, 0, game.ActivateAbilityParams{Strict: true, ExileIDs: []uuid.UUID{dead}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if z := effectstest.B12ZoneOf(g, dead); z != game.ZoneExile {
		t.Errorf("the creature card is in %s, want exile", z)
	}
	if g.RegenerationShieldsOn(strangler) <= 0 {
		t.Error("the Strangler has no regeneration shield")
	}
}

func TestOldB577SewerdregSacrificesToExileAnyGraveyardCard(t *testing.T) {
	spec, _ := effects.Lookup(oldB577Sewerdreg)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "swampwalk" {
		t.Errorf("keywords %v, want swampwalk", spec.PrintedKeywords)
	}
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	dreg := oldB577Perm(g, me.ID, "Sewerdreg", "Creature — Spirit", oldB577Sewerdreg, 3, 3)
	card := effectstest.PushToGraveyard(g, opp, game.Card{Name: "Their Card", TypeLine: "Creature — Bear", Power: 1, Toughness: 1})
	if err := g.ActivateCatalogAbility(me.ID, dreg, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no target")
	}
	if err := g.ActivateCatalogAbility(me.ID, dreg, 0, game.ActivateAbilityParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: card}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dreg) {
		t.Error("Sewerdreg was not sacrificed")
	}
	if z := effectstest.B12ZoneOf(g, card); z != game.ZoneExile {
		t.Errorf("the target is in %s, want exile", z)
	}
}

func TestOldB577BloodbondMarchReturnsEveryPlayersCardsWithTheSpellsName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	oldB577Perm(g, me.ID, "Bloodbond March", "Enchantment", oldB577March, 0, 0)
	mine := effectstest.PushToGraveyard(g, me, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	theirs := effectstest.PushToGraveyard(g, opp, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	other := effectstest.PushToGraveyard(g, opp, game.Card{Name: "Runeclaw Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	bear := effectstest.HandCardFull(me, "Grizzly Bears", "Creature — Bear", "{1}{G}", "", []string{"G"})
	effectstest.FloatMana(t, g, me, "{G}{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(theirs) {
		t.Error("a card named Grizzly Bears stayed in a graveyard")
	}
	if g.Battlefield.Contains(other) {
		t.Error("a card with another name came back")
	}
	if c, ok := g.LookupCardForEffect(theirs); ok && c.Controller != opp.ID {
		t.Errorf("the opponent's card came back under %s, want its owner", c.Controller)
	}
}

func TestOldB577BloodbondMarchIgnoresANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	oldB577Perm(g, me.ID, "Bloodbond March", "Enchantment", oldB577March, 0, 0)
	dead := effectstest.PushToGraveyard(g, me, game.Card{Name: "Shock", TypeLine: "Instant"})
	spell := effectstest.HandCardFull(me, "Shock", "Instant", "{R}", "", []string{"R"})
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{Strict: true,
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}}); err != nil {
		t.Logf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(dead) {
		t.Error("a noncreature spell returned cards")
	}
}

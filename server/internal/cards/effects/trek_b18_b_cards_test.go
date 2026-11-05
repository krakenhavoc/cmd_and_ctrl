package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b18_b_cards_test.go — Star Trek, batch 18 (#4973), part B. Each
// test plays the real catalog entry and asserts what its text says.

const (
	trekB18BTashaOracle       = "1a9607a2-58cb-4bb8-b454-4813528221f8"
	trekB18BTransporterOracle = "6eb9865e-513c-43c6-81e1-6ad000e9c48e"
	trekB18BRikerOracle       = "7c1fd7c6-2f3d-496f-9f09-660a13ef923f"
)

func TestTrekB18BEveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB18BTashaOracle:       "Tasha Yar, Out of Time",
		trekB18BTransporterOracle: "Transporter Malfunction",
		trekB18BRikerOracle:       "Will Riker, First in the Field",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// --- fixtures ------------------------------------------------------------------

// trekB18BCreature seeds a creature on the battlefield, catalog row or not.
func trekB18BCreature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int) uuid.UUID {
	return pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// trekB18BLibrary replaces p's library with cards, the first one on top.
func trekB18BLibrary(p *game.Player, cards ...game.Card) []uuid.UUID {
	p.Library.Cards = nil
	ids := make([]uuid.UUID, len(cards))
	for i := len(cards) - 1; i >= 0; i-- {
		c := cards[i]
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = p.ID, p.ID
		p.Library.PushTop(c)
		ids[i] = c.InstanceID
	}
	return ids
}

func trekB18BMountains(g *game.Game, owner uuid.UUID, n int) {
	for i := 0; i < n; i++ {
		pushLandFor(g, owner, "Mountain", "Basic Land — Mountain")
	}
}

// trekB18BNamed lists the battlefield permanents named name.
func trekB18BNamed(g *game.Game, name string) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func trekB18BDestroy(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DestroyPermanentForEffect(id); err != nil {
			t.Fatalf("destroy: %v", err)
		}
	})
}

// --- Tasha Yar, Out of Time -------------------------------------------------------

func TestTrekB18TashaYarHasHasteAndDiesIntoAPlayableTopCard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	top := trekB18BLibrary(me,
		game.Card{Name: "Spark", TypeLine: "Instant", ManaCost: "{R}"},
		game.Card{Name: "Second", TypeLine: "Instant", ManaCost: "{R}"})
	tasha := trekB18BCreature(g, me.ID, "Tasha Yar, Out of Time", "Legendary Creature — Human Officer", trekB18BTashaOracle, 2, 1)
	if !hasEffectiveKeyword(t, g, tasha, "haste") {
		t.Error("Tasha Yar has no haste")
	}
	trekB18BDestroy(t, g, tasha)
	passPriorityAroundTable(t, g)
	if !inExile(g, top[0]) {
		t.Fatal("the top card was not exiled when Tasha Yar died")
	}
	if !me.Library.Contains(top[1]) {
		t.Error("the dies trigger exiled more than the top card")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, top[0], game.CastSpellParams{FromZone: "exile", Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("casting the exiled card: %v", err)
	}
}

func TestTrekB18TashaYarMakesAHastyTokenCopySacrificedAtTheEndStep(t *testing.T) {
	g := newCatalogGame(t)
	id, me := activateFromGraveyard(t, g, "Tasha Yar, Out of Time", "Legendary Creature — Human Officer",
		trekB18BTashaOracle, 2, 1, "{C}{C}{R}", game.ActivateAbilityParams{Strict: true})
	if !inExile(g, id) {
		t.Fatal("the cost did not exile Tasha Yar from the graveyard")
	}
	top := trekB18BLibrary(me,
		game.Card{Name: "Top", TypeLine: "Sorcery", ManaCost: "{R}"},
		game.Card{Name: "Second", TypeLine: "Sorcery", ManaCost: "{R}"})
	passPriorityAroundTable(t, g)
	toks := trekB18BNamed(g, "Tasha Yar, Out of Time")
	if len(toks) != 1 {
		t.Fatalf("%d Tasha tokens, want 1", len(toks))
	}
	tok := findBattlefieldCardForTest(g, toks[0])
	if !tok.IsToken() || tok.Controller != me.ID || !tok.HasSupertype("Legendary") {
		t.Errorf("token %v under %v legendary %v, want a legendary token under the activator",
			tok.IsToken(), tok.Controller, tok.HasSupertype("Legendary"))
	}
	if !hasEffectiveKeyword(t, g, toks[0], "haste") {
		t.Error("the token copy has no haste")
	}

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if len(trekB18BNamed(g, "Tasha Yar, Out of Time")) != 0 {
		t.Fatal("the token survived the end step")
	}
	// The sacrificed copy died, so its own dies trigger exiled the top card.
	if !inExile(g, top[0]) || inExile(g, top[1]) {
		t.Error("the token copy's dies trigger did not exile exactly the top card")
	}
}

func TestTrekB18TashaYarGraveyardAbilityIsSorcerySpeedAndCosts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceTo(t, g, game.StepBeginCombat)
	id := pushCatalogGraveyardCard(me, "Tasha Yar, Out of Time", "Legendary Creature — Human Officer", trekB18BTashaOracle, 2, 1)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{C}{C}{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the graveyard ability was activated outside a main phase")
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("a refused activation exiled the card")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	advanceToMain(t, g2)
	id2 := pushCatalogGraveyardCard(me2, "Tasha Yar, Out of Time", "Legendary Creature — Human Officer", trekB18BTashaOracle, 2, 1)
	if err := g2.AddManaForEffect(me2.ID, uuid.Nil, "{C}{R}"); err != nil {
		t.Fatal(err)
	}
	if err := g2.ActivateCatalogAbility(me2.ID, id2, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("the graveyard ability was activated with {1}{R} floating, not {2}{R}")
	}
	if !me2.Graveyard.Contains(id2) {
		t.Error("an unpaid activation exiled the card")
	}
}

// --- Transporter Malfunction -------------------------------------------------------

func TestTrekB18TransporterMalfunctionDealsThreeFromHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceToMain(t, g)
	victim := trekB18BCreature(g, opp.ID, "Victim", "Creature — Bear", "", 3, 3)
	trekB18BMountains(g, me.ID, 2)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Transporter Malfunction", TypeLine: "Sorcery", ManaCost: "{1}{R}",
		OracleID: trekB18BTransporterOracle, Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Targets: cardRefs(victim), Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("cast from hand: %v", err)
	}
	passPriorityAroundTable(t, g)
	g.RunStateChecksForTest()
	if g.Battlefield.Contains(victim) {
		t.Error("3 damage did not kill the 3/3")
	}
	if !me.Graveyard.Contains(id) || inExile(g, id) {
		t.Error("cast from hand, the sorcery did not go to the graveyard")
	}
}

func TestTrekB18TransporterMalfunctionBeamsUpFromTheGraveyard(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	spell := seedGraveyardCard(t, g, "Transporter Malfunction", "Sorcery", trekB18BTransporterOracle)
	crew := trekB18BCreature(g, me.ID, "Crew", "Creature — Human", "", 1, 1)
	trekB18BMountains(g, me.ID, 6)
	life := opp.Life
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: trekB18AltCostKeyBeamMeUp, AltCostIDs: []uuid.UUID{crew},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}, Strict: true, AutoTap: true,
	}); err != nil {
		t.Fatalf("beam me up: %v", err)
	}
	if !me.Hand.Contains(crew) {
		t.Error("the returned creature is not in its owner's hand with the spell on the stack")
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("opponent life %d → %d, want -3", life, opp.Life)
	}
	if !inExile(g, spell) || me.Graveyard.Contains(spell) {
		t.Error("a beamed-up Transporter Malfunction was not exiled as it left the stack")
	}
}

func TestTrekB18TransporterMalfunctionBeamMeUpNeedsYourCreatureAndTheMana(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	spell := seedGraveyardCard(t, g, "Transporter Malfunction", "Sorcery", trekB18BTransporterOracle)
	theirs := trekB18BCreature(g, opp.ID, "Theirs", "Creature — Bear", "", 2, 2)
	trekB18BMountains(g, me.ID, 6)
	params := game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: trekB18AltCostKeyBeamMeUp,
		Targets: cardRefs(theirs), Strict: true, AutoTap: true,
	}
	if err := g.CastSpell(me.ID, spell, params); err == nil {
		t.Fatal("beam me up was claimed without returning a creature")
	}
	params.AltCostIDs = []uuid.UUID{theirs}
	if err := g.CastSpell(me.ID, spell, params); err == nil {
		t.Fatal("beam me up returned a creature the caster does not control")
	}
	if !me.Graveyard.Contains(spell) || !g.Battlefield.Contains(theirs) {
		t.Error("a refused cast moved a card")
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	spell2 := seedGraveyardCard(t, g2, "Transporter Malfunction", "Sorcery", trekB18BTransporterOracle)
	crew := trekB18BCreature(g2, me2.ID, "Crew", "Creature — Human", "", 1, 1)
	trekB18BMountains(g2, me2.ID, 5)
	if err := g2.CastSpell(me2.ID, spell2, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: trekB18AltCostKeyBeamMeUp, AltCostIDs: []uuid.UUID{crew},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g2.Seats[(g2.Turn.ActiveSeat+1)%4].ID}}, Strict: true, AutoTap: true,
	}); err == nil {
		t.Fatal("beam me up was cast with five lands for {3}{R}{R}{R}")
	}
	if !g2.Battlefield.Contains(crew) {
		t.Error("an unpaid beam me up returned the creature")
	}
}

// --- Will Riker, First in the Field -------------------------------------------------

func trekB18BRikerLibrary(p *game.Player) []uuid.UUID {
	return trekB18BLibrary(p,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
		game.Card{Name: "Island", TypeLine: "Basic Land — Island"},
		game.Card{Name: "Plains", TypeLine: "Basic Land — Plains"},
	)
}

func TestTrekB18WillRikerPutsTwoLandsAndSacrificesThemAtYourEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ids := trekB18BRikerLibrary(me)
	trekB18BCreature(g, me.ID, "Will Riker, First in the Field", "Legendary Creature — Human Officer", trekB18BRikerOracle, 4, 3)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)

	for _, land := range []uuid.UUID{ids[1], ids[3]} {
		c := findBattlefieldCardByID(g, land)
		if c == nil {
			t.Fatal("a revealed land is not on the battlefield")
		}
		if c.Tapped || c.Controller != me.ID {
			t.Errorf("%s entered tapped=%v under %v, want untapped under Riker's controller", c.Name, c.Tapped, c.Controller)
		}
	}
	if findBattlefieldCardByID(g, ids[4]) != nil {
		t.Error("the reveal went past the second land")
	}
	if top := me.Library.Cards[len(me.Library.Cards)-1]; top.InstanceID != ids[4] {
		t.Errorf("top of library is %s, want the unrevealed Plains", top.Name)
	}
	bottom := map[uuid.UUID]bool{me.Library.Cards[0].InstanceID: true, me.Library.Cards[1].InstanceID: true}
	if !bottom[ids[0]] || !bottom[ids[2]] {
		t.Error("the revealed nonland cards are not the bottom two")
	}

	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, land := range []uuid.UUID{ids[1], ids[3]} {
		if g.Battlefield.Contains(land) || !me.Graveyard.Contains(land) {
			t.Error("a land Riker put was not sacrificed at the beginning of your end step")
		}
	}
}

func TestTrekB18WillRikerWithOneLandInTheLibraryPutsIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ids := trekB18BLibrary(me,
		game.Card{Name: "Bear", TypeLine: "Creature — Bear"},
		game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"},
		game.Card{Name: "Shock", TypeLine: "Instant"},
	)
	trekB18BCreature(g, me.ID, "Will Riker, First in the Field", "Legendary Creature — Human Officer", trekB18BRikerOracle, 4, 3)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(ids[1]) {
		t.Fatal("the only land in the library did not enter")
	}
	if me.Library.Size() != 2 || !me.Library.Contains(ids[0]) || !me.Library.Contains(ids[2]) {
		t.Errorf("library holds %d cards, want the two nonland cards back", me.Library.Size())
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ids[1]) {
		t.Error("the land was not sacrificed at the end step")
	}
}

func TestTrekB18WillRikerSparesALandThatBecameANewObject(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	ids := trekB18BRikerLibrary(me)
	trekB18BCreature(g, me.ID, "Will Riker, First in the Field", "Legendary Creature — Human Officer", trekB18BRikerOracle, 4, 3)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	var back uuid.UUID
	g.WithWriteLock(func() {
		if err := g.ExileCardForEffect(ids[1]); err != nil {
			t.Fatalf("exile: %v", err)
		}
		var err error
		if back, err = g.ReturnToBattlefieldForEffect(ids[1], me.ID, false); err != nil {
			t.Fatalf("return: %v", err)
		}
	})
	if back == uuid.Nil || !g.Battlefield.Contains(back) {
		t.Fatal("the flickered Forest is not back")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(back) {
		t.Error("a land that left and came back (a new object, CR 400.7) was sacrificed")
	}
	if g.Battlefield.Contains(ids[3]) {
		t.Error("the untouched land was not sacrificed")
	}
}

func TestTrekB18WillRikerDoesNothingOnAnOpponentsCombat(t *testing.T) {
	g := newCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	advanceToMain(t, g)
	ids := trekB18BRikerLibrary(opp)
	trekB18BCreature(g, opp.ID, "Will Riker, First in the Field", "Legendary Creature — Human Officer", trekB18BRikerOracle, 4, 3)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(ids[1]) || opp.Library.Size() != 5 {
		t.Error("Riker's trigger fired at the beginning of another player's combat")
	}
}

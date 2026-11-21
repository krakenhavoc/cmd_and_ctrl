package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82DmuReady puts a creature that can attack (no summoning sickness) on
// the battlefield without an entry event.
func s82DmuReady(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return s82DmuPush(g, owner, name, typeLine, oracle, power, toughness, colors...)
}

// --- Rundvelt Hordemaster ----------------------------------------------

func TestS82DmuB12HordemasterPumpsOnlyOtherGoblinsYouControl(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	master := s82DmuReady(g, me.ID, "Rundvelt Hordemaster", s82DmuGoblinWarriorType, s82DmuHordemasterOracle, 1, 1, "R")
	goblin := s82DmuReady(g, me.ID, "Goblin Test", "Creature — Goblin", "", 1, 1, "R")
	bear := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	theirs := s82DmuReady(g, opp.ID, "Their Goblin", "Creature — Goblin", "", 1, 1, "R")
	for id, want := range map[uuid.UUID]int{master: 1, goblin: 2, bear: 2, theirs: 1} {
		if p := effectstest.EffectivePower(t, g, id); p != want {
			t.Errorf("power %d, want %d", p, want)
		}
	}
}

// hordemasterTable lets a Goblin die with `top` on the library, and
// returns the exiled card's ID.
func hordemasterTable(t *testing.T, topName, topType string, dier func(g *game.Game, me *game.Player, master uuid.UUID) uuid.UUID) (*game.Game, *game.Player, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	master := s82DmuReady(g, me.ID, "Rundvelt Hordemaster", s82DmuGoblinWarriorType, s82DmuHordemasterOracle, 1, 1, "R")
	id := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: id, Name: topName, TypeLine: topType, ManaCost: "{R}",
		Power: 1, Toughness: 1, Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	dead := dier(g, me, master)
	effectstest.KillCreature(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	return g, me, id
}

func TestS82DmuB12HordemasterDyingExilesAGoblinYouMayCast(t *testing.T) {
	g, me, top := hordemasterTable(t, "Goblin Card", "Creature — Goblin", func(_ *game.Game, _ *game.Player, m uuid.UUID) uuid.UUID { return m })
	if !g.Exile.Contains(top) {
		t.Fatal("the top card was not exiled")
	}
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast the exiled Goblin: %v", err)
	}
}

func TestS82DmuB12HordemasterAnotherGoblinDyingTriggersIt(t *testing.T) {
	g, me, top := hordemasterTable(t, "Goblin Card", "Creature — Goblin", func(g *game.Game, me *game.Player, _ uuid.UUID) uuid.UUID {
		return s82DmuReady(g, me.ID, "Goblin Test", "Creature — Goblin", "", 1, 1, "R")
	})
	if !g.Exile.Contains(top) {
		t.Fatal("another Goblin dying did not trigger the exile")
	}
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{Strict: true, FromZone: "exile"}); err != nil {
		t.Fatalf("cast: %v", err)
	}
}

func TestS82DmuB12HordemasterAnotherCreatureDyingDoesNothing(t *testing.T) {
	g, _, top := hordemasterTable(t, "Goblin Card", "Creature — Goblin", func(g *game.Game, me *game.Player, _ uuid.UUID) uuid.UUID {
		return effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	})
	if g.Exile.Contains(top) {
		t.Fatal("a Bear dying exiled the top card")
	}
}

func TestS82DmuB12HordemasterAGoblinCardOnlyPermitsThatCard(t *testing.T) {
	g, me, top := hordemasterTable(t, "Plain Bear", "Creature — Bear", func(_ *game.Game, _ *game.Player, m uuid.UUID) uuid.UUID { return m })
	if !g.Exile.Contains(top) {
		t.Fatal("the top card is exiled whatever it is")
	}
	effectstest.FloatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, top, game.CastSpellParams{Strict: true, FromZone: "exile"}); err == nil {
		t.Fatal("a non-Goblin card was castable")
	}
}

// --- Shadow-Rite Priest -------------------------------------------------

func TestS82DmuB12PriestPumpsOtherClericsAndTutorsABlackCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	priest := effectstest.PushCatalogPermanent(g, me.ID, "Shadow-Rite Priest", s82DmuHumanClericType, s82DmuPriestOracle, false)
	cleric := s82DmuReady(g, me.ID, "Test Cleric", "Creature — Human Cleric", "", 1, 1, "W")
	if p := effectstest.EffectivePower(t, g, cleric); p != 2 {
		t.Errorf("the other Cleric is %d power, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, priest); p != 1 {
		t.Errorf("the Priest is %d power, want 1 (it pumps others only)", p)
	}
	black := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: black, Name: "Black Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Colors: []string{"B"}, Owner: me.ID, Controller: me.ID})
	me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})

	if err := s82DmuActivate(t, g, me.ID, priest, 0, cleric); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{B}{B}")
	if err := s82DmuActivate(t, g, me.ID, priest, 0, priest); err == nil {
		t.Fatal("the Priest paid for its own ability by sacrificing itself")
	}
	if err := s82DmuActivate(t, g, me.ID, priest, 0, cleric); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(cleric) {
		t.Error("the other Cleric was not sacrificed as a cost")
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchNamed(t, g, me.ID, "Black Bear")
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(black) {
		t.Fatal("the black creature did not enter")
	}
}

func TestS82DmuB12PriestCannotSacrificeItself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	priest := effectstest.PushCatalogPermanent(g, me.ID, "Shadow-Rite Priest", s82DmuHumanClericType, s82DmuPriestOracle, false)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{B}{B}")
	if err := s82DmuActivate(t, g, me.ID, priest, 0, priest); err == nil {
		t.Fatal("activated with no other Cleric to sacrifice")
	}
	if !g.Battlefield.Contains(priest) {
		t.Fatal("the Priest sacrificed itself")
	}
}

// --- Ragefire Hellkite --------------------------------------------------

func TestS82DmuB12HellkiteSacrificesForDoubleStrike(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		kite := s82DmuReady(g, me.ID, "Ragefire Hellkite", s82DmuHellkiteType, s82DmuHellkiteOracle, 5, 3, "R")
		fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(kite, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		if pay {
			effectstest.AnswerOwnPermanents(t, g, me.ID, fodder)
		} else {
			effectstest.AnswerOwnPermanents(t, g, me.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := effectstest.HasEffectiveKeyword(t, g, kite, "double strike"); got != pay {
			t.Errorf("pay=%v: double strike %v", pay, got)
		}
		if g.Battlefield.Contains(fodder) == pay {
			t.Errorf("pay=%v: fodder on the battlefield = %v", pay, g.Battlefield.Contains(fodder))
		}
		if !effectstest.HasEffectiveKeyword(t, g, kite, "flying") {
			t.Error("no flying")
		}
	}
}

// --- Tori D'Avenant, Fury Rider ----------------------------------------

func TestS82DmuB12ToriBoostsGrantsTrampleAndUntaps(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tori := s82DmuReady(g, me.ID, "Tori D'Avenant, Fury Rider", s82DmuToriType, s82DmuToriOracle, 3, 3, "R", "W")
	red := s82DmuReady(g, me.ID, "Red Bear", "Creature — Bear", "", 2, 2, "R")
	white := s82DmuReady(g, me.ID, "White Bear", "Creature — Bear", "", 2, 2, "W")
	green := s82DmuReady(g, me.ID, "Green Bear", "Creature — Bear", "", 2, 2, "G")
	idle := s82DmuReady(g, me.ID, "Idle Bear", "Creature — Bear", "", 2, 2, "R")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{tori, red, white, green} {
		if err := g.DeclareAttacker(id, opp.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)

	for _, id := range []uuid.UUID{red, white, green} {
		if p := effectstest.EffectivePower(t, g, id); p != 3 {
			t.Errorf("an attacking Bear is %d power, want 3", p)
		}
	}
	if p := effectstest.EffectivePower(t, g, idle); p != 2 {
		t.Errorf("a non-attacker is %d power, want 2", p)
	}
	if p := effectstest.EffectivePower(t, g, tori); p != 3 {
		t.Errorf("Tori is %d power, want 3: it says all OTHER attacking creatures", p)
	}
	if !effectstest.HasEffectiveKeyword(t, g, red, "trample") {
		t.Error("the red attacker has no trample")
	}
	for name, id := range map[string]uuid.UUID{"white": white, "green": green, "idle": idle} {
		if effectstest.HasEffectiveKeyword(t, g, id, "trample") {
			t.Errorf("the %s creature gained trample", name)
		}
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, white); tapped {
		t.Error("the white attacker is still tapped")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, green); !tapped {
		t.Error("the green attacker was untapped")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, red); !tapped {
		t.Error("the red attacker was untapped")
	}
}

// --- The Raven Man ------------------------------------------------------

func TestS82DmuB12RavenManMakesABirdWhenAnyoneDiscarded(t *testing.T) {
	for _, discarded := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.PushCatalogPermanent(g, me.ID, "The Raven Man", s82DmuRavenManType, s82DmuRavenManOracle, false)
		if discarded {
			effectstest.PushCatalogHandCard(opp, "Spare Card", "Instant", "")
			g.WithWriteLock(func() {
				if err := g.DiscardRandomForEffect(opp.ID, 1); err != nil {
					t.Fatalf("discard: %v", err)
				}
			})
		}
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		bird := effectstest.FindBattlefieldByName(g, "Bird")
		if discarded != (bird != uuid.Nil) {
			t.Fatalf("discarded=%v: Bird %v", discarded, bird != uuid.Nil)
		}
		if bird == uuid.Nil {
			continue
		}
		if !effectstest.HasEffectiveKeyword(t, g, bird, "flying") {
			t.Error("the Bird has no flying")
		}
		effectstest.AssertRestrictions(t, g, bird, game.CantBlock)
		if got := effectstest.TokensNamedControlledBy(g, "Bird", me.ID); got != 1 {
			t.Errorf("%d Birds for the Raven Man's controller, want 1", got)
		}
	}
}

func TestS82DmuB12RavenManMakesOneBirdPerEndStepHoweverManyDiscarded(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "The Raven Man", s82DmuRavenManType, s82DmuRavenManOracle, false)
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogHandCard(opp, "Spare Card", "Instant", "")
	}
	g.WithWriteLock(func() { _ = g.DiscardRandomForEffect(opp.ID, 3) })
	effectstest.AdvanceTo(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TokensNamedControlledBy(g, "Bird", me.ID); got != 1 {
		t.Errorf("%d Birds, want 1", got)
	}
}

func TestS82DmuB12RavenManMakesEachOpponentDiscardAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	raven := effectstest.PushCatalogPermanent(g, me.ID, "The Raven Man", s82DmuRavenManType, s82DmuRavenManOracle, false)
	card := effectstest.PushCatalogHandCard(opp, "Spare Card", "Instant", "")
	if err := s82DmuActivate(t, g, me.ID, raven, 0); err == nil {
		t.Fatal("activated with no mana")
	}
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{B}")
	if err := s82DmuActivate(t, g, me.ID, raven, 0); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats[1:] {
		if effectstest.DiscardChoiceFor(g, p.ID) != nil {
			if p.ID == opp.ID {
				effectstest.AnswerDiscard(t, g, p.ID, card)
			} else if len(p.Hand.Cards) > 0 {
				effectstest.AnswerDiscard(t, g, p.ID, p.Hand.Cards[0].InstanceID)
			}
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Contains(card) {
		t.Error("the opponent kept the card")
	}
	// Not at instant speed: in the beginning of combat the ability is refused.
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{B}")
	if err := s82DmuActivate(t, g, me.ID, raven, 0); err == nil {
		t.Error("activated outside a main phase")
	}
}

// --- Nemata, Primeval Warden -------------------------------------------

func TestS82DmuB12NemataExilesOpponentCreaturesAndMakesSaprolings(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	s82DmuReady(g, me.ID, "Nemata, Primeval Warden", s82DmuNemataType, s82DmuNemataOracle, 3, 4, "B", "G")
	theirs := effectstest.PushVanillaCreature(g, opp.ID, "Their Bear", 2, 2)
	mine := effectstest.PushVanillaCreature(g, me.ID, "My Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(theirs) })
	if !g.Exile.Contains(theirs) || opp.Graveyard.Contains(theirs) {
		t.Fatal("their creature should be exiled, not in the graveyard")
	}
	if got := effectstest.TokensNamedControlledBy(g, "Saproling", me.ID); got != 0 {
		t.Fatalf("%d Saprolings before the trigger resolved: it uses the stack", got)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TokensNamedControlledBy(g, "Saproling", me.ID); got != 1 {
		t.Fatalf("%d Saprolings, want 1", got)
	}
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(mine) })
	if !me.Graveyard.Contains(mine) {
		t.Error("my own creature should die normally")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.TokensNamedControlledBy(g, "Saproling", me.ID); got != 1 {
		t.Errorf("%d Saprolings after my own creature died, want still 1", got)
	}
}

func TestS82DmuB12NemataSaprolingAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	nemata := effectstest.PushCatalogPermanent(g, me.ID, "Nemata, Primeval Warden", s82DmuNemataType, s82DmuNemataOracle, false)
	sap := func() uuid.UUID {
		return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Saproling",
			TypeLine: "Token Creature — Saproling", Power: 1, Toughness: 1, Colors: []string{"G"}, Owner: me.ID, Controller: me.ID})
	}
	a, b := sap(), sap()
	before := effectstest.EffectivePower(t, g, nemata)

	effectstest.FloatMana(t, g, me, "{G}")
	if err := s82DmuActivate(t, g, me.ID, nemata, 0, a); err != nil {
		t.Fatalf("pump: %v", err)
	}
	if g.Battlefield.Contains(a) {
		t.Error("the Saproling was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p := effectstest.EffectivePower(t, g, nemata); p != before+2 {
		t.Errorf("Nemata is %d power, want %d", p, before+2)
	}

	// Draw needs two Saprolings; only one is left.
	effectstest.FloatMana(t, g, me, "{C}{B}")
	if err := s82DmuActivate(t, g, me.ID, nemata, 1, b); err == nil {
		t.Fatal("drew with a single Saproling to sacrifice")
	}
	c := sap()
	hand := effectstest.HandSize(me)
	if err := s82DmuActivate(t, g, me.ID, nemata, 1, b, c); err != nil {
		t.Fatalf("draw: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(c) {
		t.Error("the Saprolings were not sacrificed")
	}
	if got := effectstest.HandSize(me); got != hand+1 {
		t.Errorf("hand %d -> %d, want one card drawn", hand, got)
	}
}

// --- Squee, Dubious Monarch --------------------------------------------

func TestS82DmuB12SqueeAttackMakesATappedAttackingGoblin(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	squee := s82DmuReady(g, me.ID, "Squee, Dubious Monarch", s82DmuSqueeType, s82DmuSqueeOracle, 2, 2, "R")
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(squee, opp.ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.PassAndSettleAttackSplit(t, g, me.ID)
	if got := effectstest.TokensNamedControlledBy(g, "Goblin", me.ID); got != 1 {
		t.Fatalf("%d Goblins, want 1", got)
	}
	gob, _ := effectstest.BattlefieldCard(g, effectstest.FindBattlefieldByName(g, "Goblin"))
	if !gob.Tapped || gob.AttackingTarget == uuid.Nil {
		t.Errorf("the Goblin is tapped=%v attacking=%v, want both", gob.Tapped, gob.AttackingTarget != uuid.Nil)
	}
	if !effectstest.HasEffectiveKeyword(t, g, squee, "haste") {
		t.Error("no haste")
	}
}

func s82DmuSqueeInGraveyard(g *game.Game, me *game.Player, others int) uuid.UUID {
	id := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: id, Name: "Squee, Dubious Monarch", TypeLine: s82DmuSqueeType,
		OracleID: s82DmuSqueeOracle, ManaCost: "{2}{R}", Power: 2, Toughness: 2, Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	for i := 0; i < others; i++ {
		effectstest.PushGraveyardCardTyped(me, "Filler", "Sorcery")
	}
	return id
}

// s82DmuFillers names the first n Filler cards in me's graveyard.
func s82DmuFillers(me *game.Player, n int) []uuid.UUID {
	var out []uuid.UUID
	for _, c := range me.Graveyard.Cards {
		if c.Name == "Filler" && len(out) < n {
			out = append(out, c.InstanceID)
		}
	}
	return out
}

func TestS82DmuB12SqueeCastsFromTheGraveyardExilingFourOthers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	squee := s82DmuSqueeInGraveyard(g, me, 5)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{R}")
	if err := g.CastSpell(me.ID, squee, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: "squee-dubious-monarch-graveyard", AltCostIDs: s82DmuFillers(me, 4)}); err != nil {
		t.Fatalf("cast from the graveyard: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(squee) {
		t.Fatal("Squee did not resolve onto the battlefield")
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left: {3}{R} is the whole price", n)
	}
	if got := len(me.Graveyard.Cards); got != 1 {
		t.Errorf("%d cards left in the graveyard, want 1 (five fillers, four exiled)", got)
	}
}

func TestS82DmuB12SqueeNeedsFourOtherCardsAndTheAlternativeCost(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	squee := s82DmuSqueeInGraveyard(g, me, 3)
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{R}")
	if err := g.CastSpell(me.ID, squee, game.CastSpellParams{Strict: true, FromZone: "graveyard",
		AlternativeCost: "squee-dubious-monarch-graveyard", AltCostIDs: s82DmuFillers(me, 3)}); err == nil {
		t.Fatal("cast with only three other cards in the graveyard")
	}
	// Four others but only the printed cost {2}{R} in the pool: not enough.
	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[0]
	effectstest.AdvanceToMain(t, g2)
	squee2 := s82DmuSqueeInGraveyard(g2, me2, 4)
	effectstest.FloatMana(t, g2, me2, "{C}{C}{R}")
	if err := g2.CastSpell(me2.ID, squee2, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("cast from the graveyard for its printed mana cost")
	}
}

// --- Phoenix Chick ------------------------------------------------------

func chickTable(t *testing.T, attackers int) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chick := uuid.New()
	me.Graveyard.PushTop(game.Card{InstanceID: chick, Name: "Phoenix Chick", TypeLine: s82DmuChickType, OracleID: s82DmuChickOracle,
		ManaCost: "{R}", Power: 1, Toughness: 1, Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	var ids []uuid.UUID
	for i := 0; i < attackers; i++ {
		ids = append(ids, s82DmuReady(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2, "R"))
	}
	return g, me, chick, ids
}

func chickAttack(t *testing.T, g *game.Game, defender uuid.UUID, ids []uuid.UUID) {
	t.Helper()
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range ids {
		if err := g.DeclareAttacker(id, defender); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	effectstest.LockInAttacks(t, g)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
}

func TestS82DmuB12PhoenixChickReturnsTappedAndAttackingWithACounter(t *testing.T) {
	g, me, chick, ids := chickTable(t, 3)
	chickAttack(t, g, g.Seats[1].ID, ids)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	if effectstest.LatestOptionPickFor(g, me.ID) == nil {
		t.Fatal("no \"what does it attack?\" question with three opponents to pick from")
	}
	effectstest.AnswerOptionPick(t, g, me.ID, 1)
	c, ok := effectstest.BattlefieldCard(g, chick)
	if !ok {
		t.Fatal("the Chick did not return to the battlefield")
	}
	if !c.Tapped || c.AttackingTarget == uuid.Nil {
		t.Errorf("the Chick is tapped=%v attacking=%v, want both", c.Tapped, c.AttackingTarget != uuid.Nil)
	}
	if c.AttackingTarget != g.Seats[2].ID {
		t.Errorf("the Chick attacks %v, want the second option (seat 2)", c.AttackingTarget)
	}
	if got := effectstest.EffectivePowerWithCounters(t, g, chick); got != 2 {
		t.Errorf("the Chick is %d power, want 2 (a 1/1 with a +1/+1 counter)", got)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left, want {R}{R} spent", n)
	}
}

func TestS82DmuB12PhoenixChickStaysIfYouDontPayOrAttackWithTwo(t *testing.T) {
	g, me, chick, ids := chickTable(t, 3)
	chickAttack(t, g, g.Seats[1].ID, ids)
	effectstest.FloatMana(t, g, me, "{R}{R}")
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	if g.Battlefield.Contains(chick) {
		t.Error("the Chick returned without being paid for")
	}

	g2, _, chick2, ids2 := chickTable(t, 2)
	chickAttack(t, g2, g2.Seats[1].ID, ids2)
	if g2.Battlefield.Contains(chick2) || len(g2.PendingChoices) != 0 {
		t.Error("two attackers triggered the Chick")
	}
}

func TestS82DmuB12PhoenixChickCantBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	chick := s82DmuReady(g, me.ID, "Phoenix Chick", s82DmuChickType, s82DmuChickOracle, 1, 1, "R")
	effectstest.AssertRestrictions(t, g, chick, game.CantBlock)
	for _, kw := range []string{"flying", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, chick, kw) {
			t.Errorf("no %s", kw)
		}
	}
}

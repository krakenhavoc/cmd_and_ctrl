package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b218_cards_test.go — S82 OLD batch 218 (#5531, Rivals of Ixalan):
// Trapjaw Tyrant, Release to the Wind, Blazing Hope, Bishop of Binding,
// Deadeye Rig-Hauler, Admiral's Order, Protean Raider, Siren Reaver, Hadana's
// Climb // Winged Temple of Orazca, Profane Procession // Tomb of the Dusk
// Rose, Azor's Gateway // Sanctum of the Sun and Path of Mettle // Metzali,
// Tower of Triumph. Each test plays the real catalog entry through the engine.

const (
	oldB218Tyrant   = "00605b83-faf6-4779-be9d-6a2c8f0ea722"
	oldB218Release  = "49d6a9be-1d10-4ce4-afa3-aa637f39b898"
	oldB218Hope     = "b8d87d15-0ab4-46e3-8258-8ca267f354fd"
	oldB218Bishop   = "c2dd6c78-e482-4a3c-ad9c-7dd71b3e34d2"
	oldB218Rig      = "1355424e-9909-41dc-8d43-b8d21b0919bf"
	oldB218Order    = "ae4fe5e0-408d-4398-84d7-a901b836df2a"
	oldB218Raider   = "b2f6a75e-5637-41a8-a32b-56666716492d"
	oldB218Reaver   = "bc1c9ca8-2f00-4761-bd85-1329287ae7ac"
	oldB218Climb    = "93b91d18-6acf-42e5-9a31-bc6e01f90c1f"
	oldB218Proc     = "a656ad7f-133f-4d93-919a-43bcf1f815f3"
	oldB218Gateway  = "c0cbb347-b060-43ce-a9c5-8c835be3cf1b"
	oldB218Path     = "db9ea3f9-c723-422f-98cc-a3ef7ca2c290"
	oldB218Legend   = "Legendary Enchantment"
	oldB218LandLine = "Legendary Land"
)

func TestOldB218EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB218Tyrant:         "Trapjaw Tyrant",
		oldB218Release:        "Release to the Wind",
		oldB218Hope:           "Blazing Hope",
		oldB218Bishop:         "Bishop of Binding",
		oldB218Rig:            "Deadeye Rig-Hauler",
		oldB218Order:          "Admiral's Order",
		oldB218Raider:         "Protean Raider",
		oldB218Reaver:         "Siren Reaver",
		oldB218Climb:          "Hadana's Climb",
		oldB218Climb + "#1":   "Winged Temple of Orazca",
		oldB218Proc:           "Profane Procession",
		oldB218Proc + "#1":    "Tomb of the Dusk Rose",
		oldB218Gateway:        "Azor's Gateway",
		oldB218Gateway + "#1": "Sanctum of the Sun",
		oldB218Path:           "Path of Mettle",
		oldB218Path + "#1":    "Metzali, Tower of Triumph",
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

func oldB218Seats(g *game.Game) (me, opp, third *game.Player) {
	n := len(g.Seats)
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%n], g.Seats[(g.Turn.ActiveSeat+2)%n]
}

func oldB218Row(oracle, front, frontType, frontCost, back, backType string, colors ...string) cards.Card {
	return effectstest.TransformRow(oracle, front, frontType, frontCost, back, backType, "", "", colors)
}

func oldB218Tap(g *game.Game, id uuid.UUID, tapped bool) {
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				g.Battlefield.Cards[i].Tapped = tapped
			}
		}
	})
}

func oldB218BattlefieldNamed(g *game.Game, name string) (game.Card, bool) {
	for _, c := range g.Battlefield.Cards {
		if c.Name == name {
			return c, true
		}
	}
	return game.Card{}, false
}

// --- Trapjaw Tyrant ----------------------------------------------------------

func TestOldB218TyrantExilesACreatureWhenDamagedUntilItLeaves(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := oldB218Seats(g)
	tyrant := effectstest.B12Push(g, me.ID, "Trapjaw Tyrant", "Creature — Dinosaur", oldB218Tyrant, 5, 5)
	pinger := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Bear", 1, 1)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	effectstest.Damage(t, g, pinger, tyrant, 1)
	g.RunStateChecksForTest()
	if p := effectstest.LatestPickTarget(g, me.ID); p != nil && effectstest.HasID(p.PickTargetCards, mine) {
		t.Errorf("your own creature is a legal target: %v", p.PickTargetCards)
	}
	effectstest.AnswerPickTarget(t, g, pinger)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(pinger) {
		t.Fatal("the opponent's creature was not exiled")
	}
	effectstest.RtaDestroy(t, g, tyrant)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByName(g, "Pinger") {
		t.Error("the exiled creature did not return when the Tyrant left")
	}
}

// --- Release to the Wind -----------------------------------------------------

func TestOldB218ReleaseLetsTheOwnerCastTheExiledCardForFree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := oldB218Seats(g)
	rock := effectstest.PushCatalogPermanent(g, opp.ID, "Shield of the Realm", "Artifact — Equipment", "d669ed30-efe9-4455-813a-c747b123333f", false)
	land := effectstest.B12Permanent(g, opp.ID, "Plains", "Basic Land — Plains")
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Release to the Wind", "Instant", oldB218Release, effectstest.CardRefs(land)); err == nil {
		t.Error("Release to the Wind targeted a land")
	}
	effectstest.CastCatalogSpell(t, g, "Release to the Wind", "Instant", oldB218Release, effectstest.CardRefs(rock))
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.B12ZoneOf(g, rock) != game.ZoneExile {
		t.Fatalf("the permanent is in %v, want exile", effectstest.B12ZoneOf(g, rock))
	}
	if !effectstest.B12ExiledPlayableBy(t, g, rock, opp.ID) {
		t.Error("the owner may not cast the exiled card")
	}
	if effectstest.B12ExiledPlayableBy(t, g, rock, me.ID) {
		t.Error("the spell's controller may cast the exiled card")
	}
}

// --- Blazing Hope ------------------------------------------------------------

func TestOldB218HopeNeedsPowerAtLeastYourLifeTotal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := oldB218Seats(g)
	g.WithWriteLock(func() { me.Life = 4 })
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 4, 4)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Blazing Hope", "Instant", oldB218Hope, effectstest.CardRefs(small)); err == nil {
		t.Error("Blazing Hope targeted a creature with power below your life total")
	}
	effectstest.CastCatalogSpell(t, g, "Blazing Hope", "Instant", oldB218Hope, effectstest.CardRefs(big))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(big) {
		t.Error("the creature with power equal to your life total was not exiled")
	}
}

// --- Bishop of Binding -------------------------------------------------------

func TestOldB218BishopExilesAndPumpsAVampireByTheExiledPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := oldB218Seats(g)
	bishop := effectstest.CastCatalogSpell(t, g, "Bishop of Binding", "Creature — Vampire Cleric", oldB218Bishop, nil)
	prey := effectstest.B12Creature(g, opp.ID, "Prey", "Creature — Ogre", 4, 4)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, prey)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(prey) {
		t.Fatal("the creature was not exiled")
	}
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == bishop {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	_ = me
	p0, t0 := effectstest.PtOf(t, g, bishop)
	effectstest.DeclareAttack(t, g, opp.ID, bishop)
	effectstest.AnswerPickTarget(t, g, bishop)
	effectstest.PassPriorityAroundTable(t, g)
	if p, to := effectstest.PtOf(t, g, bishop); p-p0 != 4 || to-t0 != 4 {
		t.Errorf("the Bishop went from %d/%d to %d/%d, want +4/+4 for the exiled 4-power creature", p0, t0, p, to)
	}
	effectstest.RtaDestroy(t, g, bishop)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByName(g, "Prey") {
		t.Error("the exiled creature did not return when the Bishop left")
	}
}

// --- Deadeye Rig-Hauler ------------------------------------------------------

func TestOldB218RigHaulerBouncesOnlyAfterYouAttacked(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := oldB218Seats(g)
		attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
		victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
		if attacked {
			effectstest.DeclareAttack(t, g, opp.ID, attacker)
			effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
		} else {
			effectstest.AdvanceToMain(t, g)
		}
		effectstest.CastCatalogSpell(t, g, "Deadeye Rig-Hauler", "Creature — Human Pirate", oldB218Rig, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if !attacked {
			if effectstest.LatestPickTarget(g, me.ID) != nil || !g.Battlefield.Contains(victim) {
				t.Error("the raid trigger fired without an attack")
			}
			continue
		}
		effectstest.AnswerPickTarget(t, g, victim)
		effectstest.AnswerLatestTriggerPrompt(t, g, me.ID, true)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(victim) || !opp.Hand.Contains(victim) {
			t.Error("the target creature was not returned to its owner's hand")
		}
	}
}

// --- Admiral's Order ---------------------------------------------------------

func TestOldB218AdmiralsOrderOffersTheRaidCostOnlyAfterAnAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := oldB218Seats(g)
	spec, ok := effects.Lookup(oldB218Order)
	if !ok || len(spec.AlternativeCosts) != 1 {
		t.Fatalf("Admiral's Order has %d alternative costs, want 1", len(spec.AlternativeCosts))
	}
	alt := spec.AlternativeCosts[0]
	if alt.ManaCost != "{U}" {
		t.Errorf("the raid cost is %q, want {U}", alt.ManaCost)
	}
	if alt.Condition(g, me.ID) {
		t.Error("the raid cost is offered before you attacked")
	}
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	if !alt.Condition(g, me.ID) {
		t.Error("the raid cost is not offered after you attacked")
	}
	if alt.Condition(g, opp.ID) {
		t.Error("the raid cost is offered to a player who did not attack")
	}
}

// --- Siren Reaver ------------------------------------------------------------

func TestOldB218ReaverCostsOneLessAfterYouAttacked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := oldB218Seats(g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
	card := game.Card{Name: "Siren Reaver", TypeLine: "Creature — Siren Pirate", ManaCost: "{3}{U}", OracleID: oldB218Reaver}
	q := game.CostQuery{Game: g, Card: card, Controller: me.ID, Source: card}
	spec, _ := effects.Lookup(oldB218Reaver)
	if len(spec.SelfCostModifiers) != 1 {
		t.Fatalf("%d self cost modifiers, want 1", len(spec.SelfCostModifiers))
	}
	mod := spec.SelfCostModifiers[0]
	if mod.AppliesTo != nil && mod.AppliesTo(q) {
		t.Error("the reduction applies before you attacked")
	}
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	if mod.AppliesTo == nil || !mod.AppliesTo(q) || mod.Amount(q) != 1 {
		t.Error("the reduction of 1 does not apply after you attacked")
	}
}

// --- Protean Raider ----------------------------------------------------------

func TestOldB218RaiderCopiesACreatureOnlyAfterYouAttacked(t *testing.T) {
	for _, attacked := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp, _ := oldB218Seats(g)
		attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 1)
		giant := effectstest.SeedCopyableCreature(g, opp.ID, "Giant", "Creature — Giant", 6, 6)
		if attacked {
			effectstest.DeclareAttack(t, g, opp.ID, attacker)
			effectstest.AdvanceTo(t, g, game.StepPostcombatMain)
		} else {
			effectstest.AdvanceToMain(t, g)
		}
		raider := effectstest.CastCatalogSpell(t, g, "Protean Raider", "Creature — Shapeshifter Pirate", oldB218Raider, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if attacked {
			if effectstest.CopyPrompt(g) == nil {
				t.Fatal("no copy prompt after an attack")
			}
			effectstest.ResolveWithCopyChoice(t, g, giant)
		} else if effectstest.CopyPrompt(g) != nil {
			t.Error("a copy was offered without an attack")
		}
		c := effectstest.LayeredCard(t, g, raider)
		if attacked && (c.Name != "Giant" || c.CurrentPower() != 6) {
			t.Errorf("the Raider is %q %d/%d, want a copy of the Giant", c.Name, c.CurrentPower(), c.CurrentToughness())
		}
		if !attacked && c.Name != "Protean Raider" {
			t.Errorf("the Raider is %q, want itself", c.Name)
		}
	}
}

// --- Hadana's Climb // Winged Temple of Orazca -------------------------------

func oldB218ClimbRow() cards.Card {
	return oldB218Row(oldB218Climb, "Hadana's Climb", oldB218Legend, "{1}{G}{U}", "Winged Temple of Orazca", oldB218LandLine, "G", "U")
}

func TestOldB218ClimbCountsUpAndTransformsAtThree(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, _, _ := oldB218Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.PutCounters(t, g, bear, game.CounterPlusOne, 1)
	effectstest.ImportToBattlefield(t, g, oldB218ClimbRow(), me)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := oldB218BattlefieldNamed(g, "Hadana's Climb"); !ok {
		t.Fatal("Hadana's Climb transformed with two counters")
	}
	if got := effectstest.B12Counter(t, g, bear, game.CounterPlusOne); got != 2 {
		t.Fatalf("the creature has %d counters, want 2", got)
	}
	// The next turn of mine: the third counter transforms it.
	oldB218Walk(t, g, me)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := oldB218BattlefieldNamed(g, "Winged Temple of Orazca"); !ok {
		t.Error("Hadana's Climb did not transform at three +1/+1 counters")
	}
}

// oldB218Walk passes turns until it is `p`'s turn again, at its precombat main.
func oldB218Walk(t *testing.T, g *game.Game, p *game.Player) {
	t.Helper()
	for i := 0; i < 2*len(g.Seats); i++ {
		effectstest.EndTurn(t, g)
		if g.Seats[g.Turn.ActiveSeat].ID == p.ID {
			return
		}
	}
	t.Fatal("never reached my turn again")
}

func TestOldB218WingedTemplePumpsByPowerAndGivesFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _, _ := oldB218Seats(g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	id := effectstest.ImportToBattlefield(t, g, oldB218ClimbRow(), me)
	g.WithWriteLock(func() { _ = g.TransformPermanentForEffect(id) })
	land := effectstest.RoonaOnly(t, g, "Winged Temple of Orazca").InstanceID
	effectstest.ApaMana(me, "C", "G", "U")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if p, to := effectstest.PtOf(t, g, bear); p != 6 || to != 6 {
		t.Errorf("the creature is %d/%d, want 6/6", p, to)
	}
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the creature did not gain flying")
	}
}

// --- Profane Procession // Tomb of the Dusk Rose -----------------------------

func oldB218ProcessionRow() cards.Card {
	return oldB218Row(oldB218Proc, "Profane Procession", oldB218Legend, "{1}{W}{B}", "Tomb of the Dusk Rose", oldB218LandLine, "W", "B")
}

func TestOldB218ProcessionTransformsAtThreeExiledAndTheTombReturnsOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp, _ := oldB218Seats(g)
	proc := effectstest.ImportToBattlefield(t, g, oldB218ProcessionRow(), me)
	var victims []uuid.UUID
	for _, n := range []string{"One", "Two", "Three"} {
		victims = append(victims, effectstest.B12Creature(g, opp.ID, n, "Creature — Bear", 2, 2))
	}
	for i, v := range victims {
		effectstest.ApaMana(me, "C", "C", "C", "W", "B")
		if err := g.ActivateCatalogAbility(me.ID, proc, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(v)}); err != nil {
			t.Fatalf("activate %d: %v", i, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(v) {
			t.Fatalf("creature %d was not exiled", i)
		}
		_, transformed := oldB218BattlefieldNamed(g, "Tomb of the Dusk Rose")
		if transformed != (i == 2) {
			t.Fatalf("after %d exiles, transformed = %v", i+1, transformed)
		}
	}
	tomb := effectstest.RoonaOnly(t, g, "Tomb of the Dusk Rose").InstanceID
	effectstest.ApaMana(me, "C", "C", "W", "B")
	if err := g.ActivateCatalogAbility(me.ID, tomb, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate the Tomb: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var picked uuid.UUID
	if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick != nil && len(pick.ChooseCards) > 0 {
		picked = pick.ChooseCards[0]
		effectstest.AnswerChooseCards(t, g, me.ID, picked)
	} else {
		t.Fatal("no choice among the exiled creature cards")
	}
	effectstest.PassPriorityAroundTable(t, g)
	var back int
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.IsCreature() {
			back++
		}
	}
	if back != 1 {
		t.Errorf("you control %d creatures after the Tomb, want the one returned", back)
	}
}

// --- Azor's Gateway // Sanctum of the Sun ------------------------------------

func oldB218GatewayRow() cards.Card {
	return oldB218Row(oldB218Gateway, "Azor's Gateway", "Legendary Artifact", "{2}", "Sanctum of the Sun", oldB218LandLine)
}

func TestOldB218GatewayTransformsOnFiveDifferentManaValues(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _, _ := oldB218Seats(g)
	gate := effectstest.ImportToBattlefield(t, g, oldB218GatewayRow(), me)
	effectstest.AwbPadLibrary(me)
	life := me.Life
	for n := 1; n <= 5; n++ {
		spell := effectstest.HandSpell(me, "Spell", "Sorcery", "{"+string(rune('0'+n))+"}")
		oldB218Tap(g, gate, false)
		effectstest.ApaMana(me, "C")
		if err := g.ActivateCatalogAbility(me.ID, gate, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", n, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID, spell)
		effectstest.PassPriorityAroundTable(t, g)
		_, transformed := oldB218BattlefieldNamed(g, "Sanctum of the Sun")
		if transformed != (n == 5) {
			t.Fatalf("after %d different mana values, transformed = %v", n, transformed)
		}
	}
	if got := me.Life - life; got != 5 {
		t.Errorf("gained %d life, want 5", got)
	}
	land, _ := oldB218BattlefieldNamed(g, "Sanctum of the Sun")
	if land.Tapped {
		t.Error("the Gateway was not untapped")
	}
}

func TestOldB218GatewayDoesNotTransformOnRepeatedManaValues(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, _, _ := oldB218Seats(g)
	gate := effectstest.ImportToBattlefield(t, g, oldB218GatewayRow(), me)
	effectstest.AwbPadLibrary(me)
	for n := 0; n < 5; n++ {
		spell := effectstest.HandSpell(me, "Spell", "Sorcery", "{2}")
		oldB218Tap(g, gate, false)
		effectstest.ApaMana(me, "C")
		if err := g.ActivateCatalogAbility(me.ID, gate, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
			t.Fatalf("activation %d: %v", n, err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID, spell)
		effectstest.PassPriorityAroundTable(t, g)
	}
	if _, transformed := oldB218BattlefieldNamed(g, "Sanctum of the Sun"); transformed {
		t.Error("five cards of the same mana value transformed the Gateway")
	}
}

// --- Path of Mettle // Metzali, Tower of Triumph -----------------------------

func oldB218PathRow() cards.Card {
	return oldB218Row(oldB218Path, "Path of Mettle", oldB218Legend, "{R}{W}", "Metzali, Tower of Triumph", oldB218LandLine, "R", "W")
}

func TestOldB218PathPingsCreaturesWithoutTheKeywords(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := oldB218Seats(g)
	plain := effectstest.B12Creature(g, opp.ID, "Plain", "Creature — Bear", 2, 3)
	fast := effectstest.Pr7bCreature(g, me.ID, "Hasty", 2, "{1}", nil, "haste")
	flyer := effectstest.Pr7bCreature(g, me.ID, "Flyer", 2, "{1}", nil, "flying")
	effectstest.ImportToBattlefield(t, g, oldB218PathRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, plain); got != 1 {
		t.Errorf("the plain creature took %d, want 1", got)
	}
	if got := effectstest.DamageOn(t, g, flyer); got != 1 {
		t.Errorf("the flyer took %d, want 1 (flying is not one of the keywords)", got)
	}
	if got := effectstest.DamageOn(t, g, fast); got != 0 {
		t.Errorf("the hasty creature took %d, want 0", got)
	}
}

func TestOldB218PathTransformsWhenYouAttackWithTwoKeywordCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := oldB218Seats(g)
	a := effectstest.Pr7bCreature(g, me.ID, "Hasty A", 2, "{1}", nil, "haste")
	b := effectstest.Pr7bCreature(g, me.ID, "Vigilant B", 2, "{1}", nil, "vigilance")
	effectstest.ImportToBattlefield(t, g, oldB218PathRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := oldB218BattlefieldNamed(g, "Metzali, Tower of Triumph"); !ok {
		t.Error("Path of Mettle did not transform after attacking with two keyword creatures")
	}
}

func TestOldB218PathStaysWhenOnlyOneAttackerHasAKeyword(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := oldB218Seats(g)
	a := effectstest.Pr7bCreature(g, me.ID, "Hasty A", 2, "{1}", nil, "haste")
	b := effectstest.B12Creature(g, me.ID, "Plain B", "Creature — Bear", 2, 5)
	effectstest.ImportToBattlefield(t, g, oldB218PathRow(), me)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.DeclareAttack(t, g, opp.ID, a, b)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := oldB218BattlefieldNamed(g, "Metzali, Tower of Triumph"); ok {
		t.Error("Path of Mettle transformed with only one keyword attacker")
	}
}

func TestOldB218MetzaliPingsEachOpponentAndDestroysARandomAttacker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, third := oldB218Seats(g)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 1, 5)
	id := effectstest.ImportToBattlefield(t, g, oldB218PathRow(), me)
	g.WithWriteLock(func() { _ = g.TransformPermanentForEffect(id) })
	land := effectstest.RoonaOnly(t, g, "Metzali, Tower of Triumph").InstanceID
	before, thirdBefore, mine := opp.Life, third.Life, me.Life
	effectstest.ApaMana(me, "C", "R")
	if err := g.ActivateCatalogAbility(me.ID, land, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate damage: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if before-opp.Life != 2 || thirdBefore-third.Life != 2 || me.Life != mine {
		t.Errorf("each opponent should lose 2: opp %d→%d third %d→%d me %d→%d", before, opp.Life, thirdBefore, third.Life, mine, me.Life)
	}
	oldB218Tap(g, land, false)
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.ApaMana(me, "C", "C", "W")
	if err := g.ActivateCatalogAbility(me.ID, land, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate destroy: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("the only creature that attacked was not destroyed")
	}
}

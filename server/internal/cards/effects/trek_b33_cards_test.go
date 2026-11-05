package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b33_cards_test.go — Star Trek Commander, batch 33 (#4974). Each
// test plays the real catalog entry through the engine and asserts what
// its text says.

const (
	trekB33MandateOracle    = "201269a1-8f11-485a-a9be-6447b189b4df"
	trekB33MRessOracle      = "a33f6ec8-5992-43a1-a16c-d8e83c4017fa"
	trekB33VengeanceOracle  = "7e019f06-7e87-419e-b51d-e2cb86a6c7f1"
	trekB33GalileoOracle    = "12e4e6b6-e684-49df-9373-d1e057664272"
	trekB33SpotOracle       = "25fe4df1-abf7-4f17-9e0d-57e414697400"
	trekB33EmulatorOracle   = "726cd303-f034-4dbe-8139-550feb59f8f3"
	trekB33NodeOracle       = "24296042-2963-405b-88d7-28808280ac00"
	trekB33TribbleOracle    = "1c1caf0f-9ee8-4eaa-b8e7-309ba9be9282"
	trekB33Type03Oracle     = "ff2ab17e-185b-4782-a4e2-e8ace40ef1bb"
	trekB33EnterpriseOracle = "fa667031-25cf-45c4-97ff-19a73cb30018"
)

func TestTrekB33EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		trekB33MandateOracle:    "Merciful Mandate",
		trekB33MRessOracle:      "M'Ress, Crew Coordinator",
		trekB33VengeanceOracle:  "Right of Vengeance",
		trekB33GalileoOracle:    "Shuttlecraft Galileo",
		trekB33SpotOracle:       "Spot, Felis catus",
		trekB33EmulatorOracle:   "Stratagem Emulator",
		trekB33NodeOracle:       "Supplementation Node",
		trekB33TribbleOracle:    "Tribble Cluster",
		trekB33Type03Oracle:     "Type-03 Borg Vessel",
		trekB33EnterpriseOracle: "U.S.S. Enterprise-E, Sovereign",
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

// --- shared test helpers ------------------------------------------------------

// trekB33Tokens is the tokens named `name` that p controls.
func trekB33Tokens(g *game.Game, p uuid.UUID, name string) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.IsToken() && c.Name == name && c.Controller == p {
			out = append(out, c)
		}
	}
	return out
}

// trekB33Charge puts n charge counters on a station card directly.
func trekB33Charge(g *game.Game, id uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(id, game.CounterCharge, n) })
}

// trekB33GivePriority hands priority to p's seat in the current step.
func trekB33GivePriority(g *game.Game, p *game.Player) {
	for i, s := range g.Seats {
		if s.ID == p.ID {
			g.WithWriteLock(func() { g.Turn.PriorityHolder = i })
		}
	}
}

// --- Merciful Mandate -----------------------------------------------------------

// trekB33MandateInCombat attacks seat 1 with two creatures and casts
// Merciful Mandate in the declare attackers step, up to its question.
func trekB33MandateInCombat(t *testing.T) (*game.Game, *game.Player, uuid.UUID, []uuid.UUID) {
	t.Helper()
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, a, b)
	trekB33GivePriority(g, me)
	floatMana(t, g, me, "{C}{W}")
	spell := castFromHandForTest(t, g, me, "Merciful Mandate", "Instant", "{1}{W}", trekB33MandateOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) == nil {
		t.Fatal("Merciful Mandate asked nothing during combat")
	}
	return g, me, spell, []uuid.UUID{a, b}
}

func TestTrekB33MandateUntapsAndEndsTheCombatPhase(t *testing.T) {
	g, me, spell, attackers := trekB33MandateInCombat(t)
	for _, a := range attackers {
		if isTapped(g, a) {
			t.Error("an attacker is still tapped after the untap")
		}
	}
	answerMayChoice(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if g.Turn.Step != game.StepPostcombatMain {
		t.Errorf("at %s, want the postcombat main phase", g.Turn.Step)
	}
	if n := len(AttackingCreatures(g)); n != 0 {
		t.Errorf("%d creatures still attacking after the combat phase ended", n)
	}
	if !g.Exile.Contains(spell) || me.Graveyard.Contains(spell) {
		t.Error("ending the combat phase should exile Merciful Mandate (CR 724.2b)")
	}
}

func TestTrekB33MandateDeclinedLeavesCombatAndGoesToTheGraveyard(t *testing.T) {
	g, me, spell, attackers := trekB33MandateInCombat(t)
	answerMayChoice(t, g, me.ID, false)
	if game.PhaseOf(g.Turn.Step) != game.PhaseCombat {
		t.Errorf("at %s; declining should leave the combat phase running", g.Turn.Step)
	}
	if n := len(AttackingCreatures(g)); n != len(attackers) {
		t.Errorf("%d attackers after declining, want %d", n, len(attackers))
	}
	if !me.Graveyard.Contains(spell) {
		t.Error("a declined Mandate should resolve to its owner's graveyard")
	}
}

func TestTrekB33MandateOutsideCombatOnlyUntapsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.TapTargetForEffect(mine)
		_ = g.TapTargetForEffect(theirs)
	})
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{W}")
	castFromHandForTest(t, g, me, "Merciful Mandate", "Instant", "{1}{W}", trekB33MandateOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	if latestConfirmFor(g, me.ID) != nil {
		t.Error("outside a combat phase there is nothing to ask")
	}
	if isTapped(g, mine) {
		t.Error("my creature was not untapped")
	}
	if !isTapped(g, theirs) {
		t.Error("an opponent's creature was untapped")
	}
}

// --- M'Ress, Crew Coordinator ------------------------------------------------------

func TestTrekB33MRessMakesTwoOfficersPerLand(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "M'Ress, Crew Coordinator", "Legendary Creature — Caitian Officer", trekB33MRessOracle, "{4}{W}", 4, 4, "W")
	playLandFromHand(t, g, "Plains", "")
	passPriorityAroundTable(t, g)
	officers := trekB33Tokens(g, me.ID, "Officer")
	if len(officers) != 2 {
		t.Fatalf("%d Officers after one land, want 2", len(officers))
	}
	o := officers[0]
	if !o.IsCreature() || o.Power != 1 || o.Toughness != 1 || len(o.Colors) != 1 || o.Colors[0] != "R" {
		t.Errorf("token = %s %d/%d %v, want a 1/1 red Officer", o.TypeLine, o.Power, o.Toughness, o.Colors)
	}
}

func TestTrekB33MRessIgnoresAnOpponentsLand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, opp.ID, "M'Ress, Crew Coordinator", "Legendary Creature — Caitian Officer", trekB33MRessOracle, "{4}{W}", 4, 4, "W")
	playLandFromHand(t, g, "Plains", "")
	passPriorityAroundTable(t, g)
	if n := len(trekB33Tokens(g, opp.ID, "Officer")) + len(trekB33Tokens(g, me.ID, "Officer")); n != 0 {
		t.Errorf("%d Officers from another player's land, want 0", n)
	}
}

// --- Right of Vengeance -------------------------------------------------------------

func TestTrekB33VengeanceDestroysEachCreatureThatHitsYou(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToUpkeepOf(t, g, 1)
	a := b12Creature(g, opp.ID, "Raider", "Creature — Klingon", 2, 2)
	b := b12Creature(g, opp.ID, "Other Raider", "Creature — Klingon", 3, 3)
	other := b12Creature(g, opp.ID, "Elsewhere", "Creature — Bear", 2, 2)
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b} {
		if err := g.DeclareAttacker(id, me.ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	if err := g.DeclareAttacker(other, g.Seats[2].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttacks(t, g)
	trekB33GivePriority(g, me)
	floatMana(t, g, me, "{C}{C}{C}{W}")
	castFromHandForTest(t, g, me, "Right of Vengeance", "Instant", "{3}{W}", trekB33VengeanceOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)

	if g.Battlefield.Contains(a) || g.Battlefield.Contains(b) {
		t.Error("a creature that dealt me combat damage survived")
	}
	if !g.Battlefield.Contains(other) {
		t.Error("a creature that hit another player was destroyed")
	}
	warriors := trekB33Tokens(g, me.ID, "Klingon Warrior")
	if len(warriors) != 2 {
		t.Fatalf("%d Klingon Warriors, want 2", len(warriors))
	}
	if k := warriors[0]; k.Power != 2 || k.Toughness != 1 || !hasEffectiveKeyword(t, g, k.InstanceID, "haste") {
		t.Errorf("token = %d/%d, want a 2/1 with haste", k.Power, k.Toughness)
	}
}

func TestTrekB33VengeanceEndsWithTheTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}{W}")
	castFromHandForTest(t, g, me, "Right of Vengeance", "Instant", "{3}{W}", trekB33VengeanceOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	advanceToUpkeepOf(t, g, 1)
	raider := b12Creature(g, opp.ID, "Raider", "Creature — Klingon", 2, 2)
	declareAttack(t, g, me.ID, raider)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(raider) || len(trekB33Tokens(g, me.ID, "Klingon Warrior")) != 0 {
		t.Error("Right of Vengeance outlasted the turn it was cast")
	}
}

// --- Shuttlecraft Galileo ---------------------------------------------------------

func TestTrekB33GalileoLandfallMakesAnOfficer(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ship := b27Push(g, me.ID, "Shuttlecraft Galileo", "Legendary Artifact — Spacecraft", trekB33GalileoOracle, "{3}{W}", 3, 3, "W")
	playLandFromHand(t, g, "Plains", "")
	passPriorityAroundTable(t, g)
	if n := len(trekB33Tokens(g, me.ID, "Officer")); n != 1 {
		t.Errorf("%d Officers, want 1", n)
	}
	if layeredCard(t, g, ship).IsCreature() {
		t.Error("the Galileo is a creature with no charge counters")
	}
}

func TestTrekB33GalileoAtSixFliesAndGivesAttackersFlying(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ship := b27Push(g, me.ID, "Shuttlecraft Galileo", "Legendary Artifact — Spacecraft", trekB33GalileoOracle, "{3}{W}", 3, 3, "W")
	crew := b12Creature(g, me.ID, "Crew", "Creature — Human", 5, 5)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	walker := b12Creature(g, me.ID, "Walker", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	station(t, g, me, ship, crew)
	passPriorityAroundTable(t, g)
	if layeredCard(t, g, ship).IsCreature() {
		t.Fatal("the Galileo is a creature at five charge counters")
	}
	trekB33Charge(g, ship, 1)
	c := layeredCard(t, g, ship)
	if !c.IsCreature() || !hasEffectiveKeyword(t, g, ship, "flying") {
		t.Fatal("at six charge counters the Galileo should be a flying creature")
	}
	if pw, tg := b38PT(t, g, ship); pw != 3 || tg != 3 {
		t.Errorf("the Galileo is %d/%d, want 3/3", pw, tg)
	}

	declareAttack(t, g, opp.ID, ship, bear, walker)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(bear)...)
	passPriorityAroundTable(t, g)
	if !hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the targeted attacker did not gain flying")
	}
	if hasEffectiveKeyword(t, g, walker, "flying") {
		t.Error("an untargeted attacker gained flying")
	}
	advanceToUpkeepOf(t, g, 1)
	if hasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("flying outlasted the turn")
	}
}

// --- Spot, Felis catus --------------------------------------------------------------

func TestTrekB33SpotMakesFoodForItselfAndNontokenLegends(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{W}")
	castFromHandForTest(t, g, me, "Spot, Felis catus", "Legendary Creature — Cat", "{2}{W}", trekB33SpotOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n := len(trekB33Tokens(g, me.ID, "Food")); n != 1 {
		t.Fatalf("%d Food after Spot entered, want 1", n)
	}
	castFromHandForTest(t, g, me, "Plain Legend", "Legendary Creature — Human", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	castFromHandForTest(t, g, me, "Plain Bear", "Creature — Bear", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if n := len(trekB33Tokens(g, me.ID, "Food")); n != 2 {
		t.Errorf("%d Food, want 2: the legend makes one, the Bear none", n)
	}
}

func TestTrekB33SpotSacrificesAFoodToDestroyAnArtifact(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	spot := b27Push(g, me.ID, "Spot, Felis catus", "Legendary Creature — Cat", trekB33SpotOracle, "{2}{W}", 1, 1, "W")
	rock := b12Permanent(g, opp.ID, "Rock", "Artifact")
	advanceToMain(t, g)

	// No Food: the cost can't be paid.
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, spot, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
	}); err == nil {
		t.Fatal("Spot activated without a Food to sacrifice")
	}

	var food uuid.UUID
	g.WithWriteLock(func() { _ = g.CreateTokenForEffect(me.ID, FoodToken(), 1) })
	for _, c := range trekB33Tokens(g, me.ID, "Food") {
		food = c.InstanceID
	}
	if err := g.ActivateCatalogAbility(me.ID, spot, 0, game.ActivateAbilityParams{
		Targets:      []game.TargetRef{{Kind: game.TargetCard, ID: rock}},
		SacrificeIDs: []uuid.UUID{food},
	}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(food) {
		t.Error("the Food was not sacrificed")
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not destroyed")
	}
}

// --- Stratagem Emulator ---------------------------------------------------------------

func TestTrekB33EmulatorDiesIntoBorgEqualToItsPower(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	emu := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Stratagem Emulator", TypeLine: "Artifact Creature — Borg", OracleID: trekB33EmulatorOracle,
		ManaCost: "{2}{W}", Colors: []string{"W"}, Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID,
		Counters: map[string]int{game.CounterPlusOne: 1},
	})
	b27Kill(g, emu)
	passPriorityAroundTable(t, g)
	borg := trekB33Tokens(g, me.ID, "Borg")
	if len(borg) != 3 {
		t.Fatalf("%d Borg, want 3 (power 2 plus a +1/+1 counter)", len(borg))
	}
	if c := borg[0]; !c.IsArtifact() || !c.IsCreature() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 0 {
		t.Errorf("token = %s %d/%d %v, want a 1/1 colorless Borg artifact creature", c.TypeLine, c.Power, c.Toughness, c.Colors)
	}
}

// --- Supplementation Node ----------------------------------------------------------------

func TestTrekB33NodeArtifactCreaturesGrowWhenTheyConnect(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	node := b27Push(g, me.ID, "Supplementation Node", "Artifact Creature — Borg", trekB33NodeOracle, "{2}{W}", 2, 2, "W")
	drone := b12Creature(g, me.ID, "Drone", "Artifact Creature — Construct", 1, 1)
	bear := b12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Drone", "Artifact Creature — Construct", 1, 1)
	declareAttack(t, g, opp.ID, node, drone, bear)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	answerAnyTriggerOrderPrompt(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if countersOn(g, node, game.CounterPlusOne) != 1 || countersOn(g, drone, game.CounterPlusOne) != 1 {
		t.Errorf("counters: Node %d, Drone %d, want 1 each", countersOn(g, node, game.CounterPlusOne), countersOn(g, drone, game.CounterPlusOne))
	}
	if countersOn(g, bear, game.CounterPlusOne) != 0 {
		t.Error("a non-artifact creature got a counter")
	}
	if countersOn(g, theirs, game.CounterPlusOne) != 0 {
		t.Error("an opponent's artifact creature got a counter")
	}
}

// --- Tribble Cluster -----------------------------------------------------------------------

func TestTrekB33TribbleClusterGrowsAndDiesIntoTribbles(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cluster := b27Push(g, me.ID, "Tribble Cluster", "Creature — Tribble", trekB33TribbleOracle, "{2}{W}", 1, 1, "W")
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, cluster, game.CounterPlusOne); n != 1 {
		t.Fatalf("%d +1/+1 counters at the beginning of combat, want 1", n)
	}
	b27Kill(g, cluster)
	passPriorityAroundTable(t, g)
	tribbles := trekB33Tokens(g, me.ID, "Tribble")
	if len(tribbles) != 2 {
		t.Fatalf("%d Tribbles, want 2 (its power as it died)", len(tribbles))
	}
	c := layeredCard(t, g, tribbles[0].InstanceID)
	if !c.IsCreature() || c.Power != 1 || c.Toughness != 1 || len(c.Colors) != 1 || c.Colors[0] != "W" || !c.HasSubtype("Tribble") {
		t.Errorf("token = %s %d/%d %v, want a 1/1 white Tribble", c.TypeLine, c.Power, c.Toughness, c.Colors)
	}
	if !game.Restricted(&c, game.CantBlock) {
		t.Error("the Tribble token can block")
	}
}

func TestTrekB33TribbleClusterDoesNotGrowOnAnOpponentsTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	cluster := b27Push(g, me.ID, "Tribble Cluster", "Creature — Tribble", trekB33TribbleOracle, "{2}{W}", 1, 1, "W")
	advanceToUpkeepOf(t, g, 1)
	before := countersOn(g, cluster, game.CounterPlusOne)
	advanceTo(t, g, game.StepBeginCombat)
	passPriorityAroundTable(t, g)
	if n := countersOn(g, cluster, game.CounterPlusOne); n != before {
		t.Errorf("%d counters after an opponent's beginning of combat, want %d", n, before)
	}
}

// --- Type-03 Borg Vessel ---------------------------------------------------------------------

func TestTrekB33Type03ReducesArtifactSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Type-03 Borg Vessel", "Artifact — Spacecraft", trekB33Type03Oracle, "{3}{W}", 5, 5, "W")
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}")
	castFromHandForTest(t, g, me, "Rock", "Artifact", "{2}", "", game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)

	// A non-artifact spell is not reduced.
	floatMana(t, g, me, "{C}")
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: id, Name: "Bear", TypeLine: "Creature — Bear", ManaCost: "{2}", Owner: me.ID, Controller: me.ID})
	})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("a creature spell was reduced by {1}")
	}
}

func TestTrekB33Type03MakesBorgOnlyAtSeven(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ship := b27Push(g, me.ID, "Type-03 Borg Vessel", "Artifact — Spacecraft", trekB33Type03Oracle, "{3}{W}", 5, 5, "W")
	advanceToMain(t, g)
	trekB33Charge(g, ship, 6)
	castFromHandForTest(t, g, me, "Plain Sorcery", "Sorcery", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := len(trekB33Tokens(g, me.ID, "Borg")); n != 0 {
		t.Fatalf("%d Borg below 7+, want 0", n)
	}
	if layeredCard(t, g, ship).IsCreature() {
		t.Error("the Vessel is a creature at six charge counters")
	}
	trekB33Charge(g, ship, 1)
	if !layeredCard(t, g, ship).IsCreature() || !hasEffectiveKeyword(t, g, ship, "flying") {
		t.Error("at seven charge counters the Vessel should be a flying creature")
	}
	castFromHandForTest(t, g, me, "Plain Sorcery", "Sorcery", "", "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	borg := trekB33Tokens(g, me.ID, "Borg")
	if len(borg) != 1 {
		t.Fatalf("%d Borg at 7+, want 1", len(borg))
	}
	if c := borg[0]; !c.IsArtifact() || c.Power != 1 || c.Toughness != 1 {
		t.Errorf("token = %s %d/%d, want a 1/1 Borg artifact creature", c.TypeLine, c.Power, c.Toughness)
	}
}

// --- U.S.S. Enterprise-E, Sovereign ---------------------------------------------------------

func TestTrekB33EnterpriseEDestroysAllTappedCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	myTapped := b12Creature(g, me.ID, "My Tapped", "Creature — Bear", 2, 2)
	theirTapped := b12Creature(g, opp.ID, "Their Tapped", "Creature — Bear", 2, 2)
	untapped := b12Creature(g, opp.ID, "Untapped", "Creature — Bear", 2, 2)
	tappedRock := b12Permanent(g, opp.ID, "Tapped Rock", "Artifact")
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{myTapped, theirTapped, tappedRock} {
			_ = g.TapTargetForEffect(id)
		}
	})
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{C}{C}{W}{W}")
	ship := castFromHandForTest(t, g, me, "U.S.S. Enterprise-E, Sovereign", "Legendary Artifact — Spacecraft", "{3}{W}{W}", trekB33EnterpriseOracle, game.CastSpellParams{Strict: true})
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(myTapped) || g.Battlefield.Contains(theirTapped) {
		t.Error("a tapped creature survived")
	}
	if !g.Battlefield.Contains(untapped) || !g.Battlefield.Contains(tappedRock) {
		t.Error("an untapped creature or a tapped noncreature was destroyed")
	}
	if !g.Battlefield.Contains(ship) {
		t.Fatal("the Enterprise-E is not on the battlefield")
	}
	trekB33Charge(g, ship, 7)
	if layeredCard(t, g, ship).IsCreature() {
		t.Error("the Enterprise-E is a creature at seven charge counters")
	}
	trekB33Charge(g, ship, 1)
	if !layeredCard(t, g, ship).IsCreature() || !hasEffectiveKeyword(t, g, ship, "flying") || !hasEffectiveKeyword(t, g, ship, "vigilance") {
		t.Error("at eight charge counters the Enterprise-E should be a creature with flying and vigilance")
	}
	if pw, tg := b38PT(t, g, ship); pw != 7 || tg != 7 {
		t.Errorf("the Enterprise-E is %d/%d, want 7/7", pw, tg)
	}
}

package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b571_cards_test.go — Guildpact's other creatures, OLD batch 571
// (#5599). Each test plays the real catalog entry.

const (
	oldB571EnforcerOracle    = "0da4988f-47ca-4786-ac33-edbc9cfd6785"
	oldB571InkTreaderOracle  = "3846b499-59c8-4e77-b46a-1d697bcbded7"
	oldB571SmogsteedOracle   = "a23cb875-2ef8-4457-81eb-e66b67a3c477"
	oldB571TraineeOracle     = "d8b6d9be-11a6-402c-a847-ec6cababe560"
	oldB571BioplasmOracle    = "f39777be-39af-4a97-a70d-824d713bf145"
	oldB571BureaucratsOracle = "21dc394e-ba82-4f27-8014-9236690d10a4"
	oldB571NodorogOracle     = "68fc3611-8d82-4f87-ab50-603790a27455"
	oldB571PlowOracle        = "6dba119e-1b08-42e0-8d0a-b606bd90af02"
	oldB571BonesOracle       = "b6c6e872-1363-480f-9a8d-c393361ccbc1"
	oldB571RabbleOracle      = "1d591c00-7086-4337-bdb8-a5f35568f04e"
	oldB571WitchMawOracle    = "26edbbb8-331e-42a6-a550-acc7e9efb32c"
	oldB571WoodKinOracle     = "3bc9649c-4350-4c6a-8441-dc76a5f833af"
)

func TestOldB571CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB571EnforcerOracle:    "Spelltithe Enforcer",
		oldB571InkTreaderOracle:  "Ink-Treader Nephilim",
		oldB571SmogsteedOracle:   "Smogsteed Rider",
		oldB571TraineeOracle:     "Skyrider Trainee",
		oldB571BioplasmOracle:    "Bioplasm",
		oldB571BureaucratsOracle: "Droning Bureaucrats",
		oldB571NodorogOracle:     "Gruul Nodorog",
		oldB571PlowOracle:        "Gruul War Plow",
		oldB571BonesOracle:       "Restless Bones",
		oldB571RabbleOracle:      "Rabble-Rouser",
		oldB571WitchMawOracle:    "Witch-Maw Nephilim",
		oldB571WoodKinOracle:     "Petrified Wood-Kin",
	} {
		def, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if def.Name != name || def.Completeness != effects.CompletenessFull || len(def.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, def.Name, def.Completeness, def.Caveats)
		}
	}
}

func oldB571PT(t *testing.T, g *game.Game, id uuid.UUID, what string, p, tough int) {
	t.Helper()
	if gp, gt := effectstest.PtOf(t, g, id); gp != p || gt != tough {
		t.Errorf("%s is %d/%d, want %d/%d", what, gp, gt, p, tough)
	}
}

func oldB571OnBattlefield(g *game.Game, id uuid.UUID) bool {
	found := false
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.InstanceID == id {
				found = true
			}
		}
	})
	return found
}

// --- Spelltithe Enforcer -------------------------------------------------

func TestOldB571SpelltitheEnforcerMakesAnOpponentPayOrSacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, opp.ID, "Spelltithe Enforcer", "Creature — Elephant Wizard", oldB571EnforcerOracle, 3, 3)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)

	// Paying {1} keeps everything.
	effectstest.AdvanceToMain(t, g)
	effectstest.ApaMana(me, "C")
	effectstest.CastNoncreature(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	if !oldB571OnBattlefield(g, mine) {
		t.Fatal("a payer sacrificed anyway")
	}

	// Declining sacrifices a permanent of the caster's choice.
	effectstest.CastNoncreature(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.AnswerSacrifice(t, g, me.ID, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if oldB571OnBattlefield(g, mine) {
		t.Error("the caster who did not pay kept every permanent")
	}
}

func TestOldB571SpelltitheEnforcerIgnoresItsControllersSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	enf := effectstest.B12Push(g, me.ID, "Spelltithe Enforcer", "Creature — Elephant Wizard", oldB571EnforcerOracle, 3, 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastNoncreature(t, g)
	if n := effectstest.TriggersOnStackFrom(g, enf); n != 0 {
		t.Errorf("its controller's spell triggered it %d times", n)
	}
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("its controller was asked to pay")
	}
}

// --- Ink-Treader Nephilim -------------------------------------------------

func TestOldB571InkTreaderCopiesASpellOnItForEveryOtherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	nephilim := effectstest.B12Push(g, me.ID, "Ink-Treader Nephilim", "Creature — Nephilim", oldB571InkTreaderOracle, 3, 3)
	a := effectstest.B12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, opp.ID, "Bear B", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", effectstest.RprB13GiantGrowthOracle, effectstest.CardTarget(nephilim))
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.ChooseCardsChoiceFor(g, me.ID); pick != nil {
		if len(pick.ChooseCards) != 2 || !effectstest.HasID(pick.ChooseCards, a) || !effectstest.HasID(pick.ChooseCards, b) {
			t.Fatalf("order prompt over %v, want the two other creatures", pick.ChooseCards)
		}
		effectstest.AnswerChooseCards(t, g, me.ID, a, b)
	} else {
		t.Fatal("no order prompt for the Nephilim's controller")
	}
	for i := 0; i < 5 && len(g.Stack.Cards) > 0; i++ {
		effectstest.PassPriorityAroundTable(t, g)
	}
	oldB571PT(t, g, nephilim, "the Nephilim", 6, 6)
	oldB571PT(t, g, a, "my bear", 5, 5)
	oldB571PT(t, g, b, "their bear", 5, 5)
}

func TestOldB571InkTreaderIgnoresASpellOnAnotherCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.AdvanceToMain(t, g)
	nephilim := effectstest.B12Push(g, me.ID, "Ink-Treader Nephilim", "Creature — Nephilim", oldB571InkTreaderOracle, 3, 3)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Giant Growth", "Instant", effectstest.RprB13GiantGrowthOracle, effectstest.CardTarget(bear))
	if n := effectstest.TriggersOnStackFrom(g, nephilim); n != 0 {
		t.Errorf("a spell on another creature triggered it %d times", n)
	}
}

// --- Smogsteed Rider ------------------------------------------------------

func TestOldB571SmogsteedRiderGivesOtherAttackersFear(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rider := effectstest.B12Push(g, me.ID, "Smogsteed Rider", "Creature — Human Wizard", oldB571SmogsteedOracle, 2, 2)
	other := effectstest.B12Creature(g, me.ID, "Other Bear", "Creature — Bear", 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Home Bear", "Creature — Bear", 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, rider, other)
	if !effectstest.HasKeywordOnBattlefield(t, g, other, "fear") {
		t.Error("the other attacker has no fear")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, rider, "fear") {
		t.Error("the Rider itself has fear")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, home, "fear") {
		t.Error("a creature that stayed home has fear")
	}
}

// --- Skyrider Trainee -----------------------------------------------------

func TestOldB571SkyriderTraineeFliesOnlyWhileEnchanted(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	trainee := effectstest.B12Push(g, me.ID, "Skyrider Trainee", "Creature — Human Soldier", oldB571TraineeOracle, 3, 3)
	if effectstest.HasKeywordOnBattlefield(t, g, trainee, "flying") {
		t.Fatal("an unenchanted Trainee flies")
	}
	aura := effectstest.AuraCast(t, g, "Guardian's Magemark", "95348ba4-038d-4f86-8be0-eb1a68a1f231", trainee)
	if !effectstest.HasKeywordOnBattlefield(t, g, trainee, "flying") {
		t.Error("an enchanted Trainee does not fly")
	}
	effectstest.ExileCreature(t, g, aura)
	g.RunStateChecksForTest()
	if effectstest.HasKeywordOnBattlefield(t, g, trainee, "flying") {
		t.Error("the Trainee still flies after the Aura left")
	}
}

// --- Bioplasm -------------------------------------------------------------

func TestOldB571BioplasmGetsTheExiledCreatureCardsPowerAndToughness(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bio := effectstest.B12Push(g, me.ID, "Bioplasm", "Creature — Ooze", oldB571BioplasmOracle, 4, 4)
	id := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: id, Name: "Top Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 5, Owner: me.ID, Controller: me.ID})
	effectstest.Pr7bAttack(t, g, opp.ID, bio)
	if !g.Exile.Contains(id) {
		t.Fatal("the top card was not exiled")
	}
	oldB571PT(t, g, bio, "the attacking Bioplasm", 7, 9)
}

func TestOldB571BioplasmExilesANoncreatureCardAndGetsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bio := effectstest.B12Push(g, me.ID, "Bioplasm", "Creature — Ooze", oldB571BioplasmOracle, 4, 4)
	id := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: id, Name: "Top Land", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	effectstest.Pr7bAttack(t, g, opp.ID, bio)
	if !g.Exile.Contains(id) {
		t.Fatal("the top card was not exiled")
	}
	oldB571PT(t, g, bio, "the attacking Bioplasm", 4, 4)
}

// --- Droning Bureaucrats --------------------------------------------------

func TestOldB571DroningBureaucratsHoldsEveryCreatureWithManaValueX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	bur := effectstest.B12Push(g, me.ID, "Droning Bureaucrats", "Creature — Human Advisor", oldB571BureaucratsOracle, 1, 4)
	two := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Two Drop", TypeLine: "Creature — Bear",
		ManaCost: "{1}{G}", Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID})
	three := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Three Drop", TypeLine: "Creature — Bear",
		ManaCost: "{2}{G}", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	effectstest.ApaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, bur, 0, game.ActivateAbilityParams{Strict: true, XValue: 2}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	r2 := effectstest.AuraRestrictions(t, g, two)
	if r2&game.CantAttackOrBlock != game.CantAttackOrBlock {
		t.Errorf("a creature with mana value 2 has restrictions %v", r2)
	}
	if r := effectstest.AuraRestrictions(t, g, three); r&game.CantAttackOrBlock != 0 {
		t.Errorf("a creature with mana value 3 has restrictions %v", r)
	}
	// The set is live: a creature with mana value 2 that enters later is held.
	later := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Late Two", TypeLine: "Creature — Bear",
		ManaCost: "{1}{R}", Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	if r := effectstest.AuraRestrictions(t, g, later); r&game.CantAttackOrBlock != game.CantAttackOrBlock {
		t.Errorf("a creature with mana value 2 that entered later has restrictions %v", r)
	}
}

// --- Gruul Nodorog ----------------------------------------------------------

func TestOldB571GruulNodorogGainsMenaceForR(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	nod := effectstest.B12Push(g, me.ID, "Gruul Nodorog", "Creature — Beast", oldB571NodorogOracle, 4, 4)
	if err := g.ActivateCatalogAbility(me.ID, nod, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated with no mana")
	}
	effectstest.ApaMana(me, "R")
	if err := g.ActivateCatalogAbility(me.ID, nod, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, nod, "menace") {
		t.Error("no menace")
	}
}

// --- Gruul War Plow ---------------------------------------------------------

func TestOldB571GruulWarPlowGivesTrampleAndBecomesAJuggernaut(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plow := effectstest.PushCatalogPermanent(g, me.ID, "Gruul War Plow", "Artifact", oldB571PlowOracle, false)
	mine := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	if !effectstest.HasKeywordOnBattlefield(t, g, mine, "trample") {
		t.Error("my creature has no trample")
	}
	if effectstest.HasKeywordOnBattlefield(t, g, theirs, "trample") {
		t.Error("an opponent's creature has trample")
	}
	effectstest.ApaMana(me, "C", "R", "G")
	if err := g.ActivateCatalogAbility(me.ID, plow, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	oldB571PT(t, g, plow, "the animated Plow", 4, 4)
	if !effectstest.HasKeywordOnBattlefield(t, g, plow, "trample") {
		t.Error("the animated Plow has no trample from itself")
	}
}

// --- Restless Bones -----------------------------------------------------------

func TestOldB571RestlessBonesGrantsSwampwalkAndRegenerates(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bones := effectstest.B12Push(g, me.ID, "Restless Bones", "Creature — Skeleton", oldB571BonesOracle, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.ApaMana(me, "B", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, bones, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(bear)}); err != nil {
		t.Fatalf("activate swampwalk: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasKeywordOnBattlefield(t, g, bear, "swampwalk") {
		t.Error("the target has no swampwalk")
	}
	// The Skeleton is tapped now, but regeneration costs no tap.
	effectstest.ApaMana(me, "B", "C")
	if err := g.ActivateCatalogAbility(me.ID, bones, 1, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate regenerate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.Damage(t, g, bear, bones, 3)
	g.RunStateChecksForTest()
	if !oldB571OnBattlefield(g, bones) {
		t.Error("the regenerating Skeleton was destroyed")
	}
}

// --- Rabble-Rouser ------------------------------------------------------------

func TestOldB571RabbleRouserBloodthirstAndAttackerPump(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.AdvanceToMain(t, g)
	rouser := effectstest.DftABCast(t, g, me, game.Card{Name: "Rabble-Rouser", TypeLine: "Creature — Goblin Shaman", OracleID: oldB571RabbleOracle,
		ManaCost: "{3}{R}", Power: 1, Toughness: 1})
	oldB571PT(t, g, rouser, "an undamaged-opponent Rabble-Rouser", 1, 1)

	effectstest.GptB1DamagePlayer(g, opp.ID, 1)
	second := effectstest.DftABCast(t, g, me, game.Card{Name: "Rabble-Rouser", TypeLine: "Creature — Goblin Shaman", OracleID: oldB571RabbleOracle,
		ManaCost: "{3}{R}", Power: 1, Toughness: 1})
	oldB571PT(t, g, second, "a Rabble-Rouser cast after damage", 2, 2)

	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 2, 2)
	home := effectstest.B12Creature(g, me.ID, "Home", "Creature — Bear", 2, 2)
	effectstest.Pr7bAttack(t, g, opp.ID, atk)
	effectstest.ApaMana(me, "R")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == second {
				g.Battlefield.Cards[i].SummonedThisTurn = false
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, second, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	oldB571PT(t, g, atk, "the attacker", 4, 2)
	oldB571PT(t, g, home, "a creature that stayed home", 2, 2)
}

// --- Witch-Maw Nephilim ---------------------------------------------------------

func TestOldB571WitchMawNephilimGrowsWhenYouCastASpellIfYouChoose(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	witch := effectstest.B12Push(g, me.ID, "Witch-Maw Nephilim", "Creature — Nephilim", oldB571WitchMawOracle, 1, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastNoncreature(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	oldB571PT(t, g, witch, "the Nephilim after saying yes", 3, 3)

	effectstest.CastNoncreature(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	oldB571PT(t, g, witch, "the Nephilim after saying no", 3, 3)
}

func TestOldB571WitchMawNephilimGainsTrampleOnlyAtTenPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := effectstest.B12Push(g, me.ID, "Witch-Maw Nephilim", "Creature — Nephilim", oldB571WitchMawOracle, 9, 9)
	effectstest.Pr7bAttack(t, g, opp.ID, small)
	if effectstest.HasKeywordOnBattlefield(t, g, small, "trample") {
		t.Error("a 9-power Nephilim gained trample")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	big := effectstest.B12Push(g2, me2.ID, "Witch-Maw Nephilim", "Creature — Nephilim", oldB571WitchMawOracle, 10, 10)
	effectstest.Pr7bAttack(t, g2, opp2.ID, big)
	if !effectstest.HasKeywordOnBattlefield(t, g2, big, "trample") {
		t.Error("a 10-power Nephilim did not gain trample")
	}
}

// --- Petrified Wood-Kin -----------------------------------------------------------

func TestOldB571PetrifiedWoodKinEntersWithCountersForAllDamageDealtToOpponents(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, opp2 := g.Seats[0], g.Seats[1], g.Seats[2]
	effectstest.AdvanceToMain(t, g)
	none := effectstest.DftABCast(t, g, me, game.Card{Name: "Petrified Wood-Kin", TypeLine: "Creature — Elemental Warrior", OracleID: oldB571WoodKinOracle,
		ManaCost: "{6}{G}", Power: 3, Toughness: 3})
	oldB571PT(t, g, none, "a Wood-Kin cast with no damage dealt", 3, 3)

	effectstest.GptB1DamagePlayer(g, opp.ID, 2)
	effectstest.GptB1DamagePlayer(g, opp2.ID, 3)
	effectstest.GptB1DamagePlayer(g, me.ID, 4) // damage to you is not damage to an opponent
	some := effectstest.DftABCast(t, g, me, game.Card{Name: "Petrified Wood-Kin", TypeLine: "Creature — Elemental Warrior", OracleID: oldB571WoodKinOracle,
		ManaCost: "{6}{G}", Power: 3, Toughness: 3})
	oldB571PT(t, g, some, "a Wood-Kin cast after 5 damage to opponents", 8, 8)
	if !effectstest.HasKeywordOnBattlefield(t, g, some, "protection from instants") {
		t.Error("no protection from instants")
	}
}

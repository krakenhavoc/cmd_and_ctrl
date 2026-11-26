package testss82m

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b353_test.go — S82 older sets, batch 353: Duel Decks: Speed vs.
// Cunning (#5555).

const (
	oldB353ArrowVolleyTrap  = "131dc114-cd48-43d9-93b3-88df85640e09"
	oldB353InfernoTrap      = "83a71000-942b-4472-9521-0e1070a3343b"
	oldB353FaerieImpostor   = "0d78ba2c-4de0-47d9-b1b9-13dc9de8844b"
	oldB353KathariBomber    = "1ad39f85-799a-42ae-852c-9117d689fe65"
	oldB353ScourgeDevil     = "3ee65210-547f-4375-8f9a-3484d963339f"
	oldB353MasterDecoy      = "94a16609-8529-4301-8a3d-47ced194f7cd"
	oldB353HellraiserGoblin = "c30e89fa-8993-468d-b2ee-093c3950f520"
	oldB353SwiftJustice     = "d363e135-0a76-4d0e-b3b7-3293944f392c"
	oldB353GoblinDeathrider = "d5142292-446a-413b-85a0-be8a59cec1f8"
)

func TestOldB353EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB353ArrowVolleyTrap: "Arrow Volley Trap", oldB353InfernoTrap: "Inferno Trap",
		oldB353FaerieImpostor: "Faerie Impostor", oldB353KathariBomber: "Kathari Bomber",
		oldB353ScourgeDevil: "Scourge Devil", oldB353MasterDecoy: "Master Decoy",
		oldB353HellraiserGoblin: "Hellraiser Goblin", oldB353SwiftJustice: "Swift Justice",
		oldB353GoblinDeathrider: "Goblin Deathraiders",
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

func TestOldB353GoblinDeathraidersHasTrample(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Goblin Deathraiders", "Creature — Goblin Warrior", oldB353GoblinDeathrider, 3, 1)
	effectstest.AdvanceToMain(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, id, "trample") {
		t.Error("no trample")
	}
}

func TestOldB353HellraiserGoblinGivesYourCreaturesHasteAndAMustAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	hell := effectstest.B12Push(g, me.ID, "Hellraiser Goblin", "Creature — Goblin Berserker", oldB353HellraiserGoblin, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	for _, id := range []uuid.UUID{hell, mine} {
		if !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
			t.Errorf("%s has no haste", id)
		}
	}
	if effectstest.HasEffectiveKeyword(t, g, theirs, "haste") {
		t.Error("an opponent's creature got haste")
	}
	// Every creature of mine is required to attack, the Goblin and the Bear.
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	var re *game.AttackRequirementError
	if err := g.PassPriority(); !errors.As(err, &re) {
		t.Fatalf("pass with everything home = %v, want an attack-requirement refusal", err)
	}
	if err := g.DeclareAttacker(mine, opp.ID); err != nil {
		t.Fatalf("declare: %v", err)
	}
	if err := g.PassPriority(); !errors.As(err, &re) || re.Attacker != hell {
		t.Fatalf("pass with the Goblin home = %v, want a refusal naming it", err)
	}
}

func TestOldB353MasterDecoyTapsATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	decoy := effectstest.B12Push(g, me.ID, "Master Decoy", "Creature — Human Soldier", oldB353MasterDecoy, 1, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "W")
	effectstest.B16Activate(t, g, me.ID, decoy, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim)})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.IsTapped(g, victim) {
		t.Error("the target is untapped")
	}
	if !effectstest.IsTapped(g, decoy) {
		t.Error("the Decoy did not tap for its cost")
	}
	// Without {W} the ability is refused.
	again := effectstest.B12Push(g, me.ID, "Master Decoy", "Creature — Human Soldier", oldB353MasterDecoy, 1, 2)
	if err := g.ActivateCatalogAbility(me.ID, again, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(victim)}); err == nil {
		t.Error("activated with no mana")
	}
}

func TestOldB353ScourgeDevilPumpsYourCreaturesAndUnearths(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	devil := effectstest.CastHoldCreature(t, g, "Scourge Devil", "Creature — Devil", oldB353ScourgeDevil, 3, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if p, _ := effectstest.PtOf(t, g, mine); p != 3 {
		t.Errorf("my creature has power %d, want 3", p)
	}
	if p, _ := effectstest.PtOf(t, g, devil); p != 4 {
		t.Errorf("the Devil has power %d, want 4", p)
	}
	if p, _ := effectstest.PtOf(t, g, theirs); p != 2 {
		t.Errorf("their creature has power %d, want 2", p)
	}

	g2 := effectstest.NewCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	effectstest.ToMainForCost(t, g2)
	gy := effectstest.PushCatalogGraveyardCard(me2, "Scourge Devil", "Creature — Devil", oldB353ScourgeDevil, 3, 3)
	idx := -1
	for i, ab := range game.ActivatedAbilitiesForCard(game.Card{OracleID: oldB353ScourgeDevil}) {
		if game.AbilityFunctionsFromZone(ab, game.ZoneGraveyard) {
			idx = i
		}
	}
	if idx < 0 {
		t.Fatal("no unearth")
	}
	effectstest.FloatForTest(g2, me2, "CCR")
	if err := g2.ActivateCatalogAbility(me2.ID, gy, idx, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("unearth: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	if !g2.Battlefield.Contains(gy) || !effectstest.EffectiveAbilitiesContain(t, g2, gy, "haste") {
		t.Fatal("unearthed Devil is not a hasty permanent")
	}
	effectstest.AdvanceTo(t, g2, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(gy) || !g2.Exile.Contains(gy) {
		t.Error("the unearthed Devil was not exiled at the end step")
	}
}

func TestOldB353KathariBomberMakesTwoGoblinsAndIsSacrificed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	bomber := effectstest.B12Push(g, me.ID, "Kathari Bomber", "Creature — Bird Shaman", oldB353KathariBomber, 2, 2)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	effectstest.DeclareAttack(t, g, opp.ID, bomber)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
	if g.Battlefield.Contains(bomber) {
		t.Error("the Bomber was not sacrificed")
	}
	goblins := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Goblin" && c.Controller == me.ID {
			goblins++
		}
	}
	if goblins != 2 {
		t.Errorf("%d Goblin tokens, want 2", goblins)
	}
}

func TestOldB353FaerieImpostorBouncesAnotherCreatureOrIsSacrificed(t *testing.T) {
	t.Run("alone", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		imp := effectstest.CastHoldCreature(t, g, "Faerie Impostor", "Creature — Faerie Rogue", oldB353FaerieImpostor, 2, 1)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(imp) {
			t.Error("the Impostor stayed with nothing to return")
		}
	})
	t.Run("return another creature", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
		imp := effectstest.CastHoldCreature(t, g, "Faerie Impostor", "Creature — Faerie Rogue", oldB353FaerieImpostor, 2, 1)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID, other)
		effectstest.PassPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(imp) || g.Battlefield.Contains(other) || !me.Hand.Contains(other) {
			t.Error("the creature was not returned and the Impostor kept")
		}
	})
	t.Run("decline", func(t *testing.T) {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		other := effectstest.B12Creature(g, me.ID, "Other", "Creature — Bear", 2, 2)
		imp := effectstest.CastHoldCreature(t, g, "Faerie Impostor", "Creature — Faerie Rogue", oldB353FaerieImpostor, 2, 1)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerChooseCards(t, g, me.ID)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Battlefield.Contains(imp) || !g.Battlefield.Contains(other) {
			t.Error("declining did not sacrifice the Impostor and keep the creature")
		}
	})
}

func TestOldB353SwiftJusticePumpsAndGrantsFirstStrikeAndLifelink(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	c := effectstest.B12Creature(g, me.ID, "Target", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	effectstest.CastCatalogSpell(t, g, "Swift Justice", "Instant", oldB353SwiftJustice, effectstest.CardTarget(c))
	effectstest.PassPriorityAroundTable(t, g)
	if p, tough := effectstest.PtOf(t, g, c); p != 3 || tough != 2 {
		t.Errorf("%d/%d, want 3/2", p, tough)
	}
	for _, kw := range []string{"first strike", "lifelink"} {
		if !effectstest.HasEffectiveKeyword(t, g, c, kw) {
			t.Errorf("no %s", kw)
		}
	}
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Swift Justice", "Instant", oldB353SwiftJustice, nil); err == nil {
		t.Error("cast with no target")
	}
}

// oldB353AttackAndDamage has the active seat attack the next seat with the
// attackers, through combat damage.
func oldB353AttackAndDamage(t *testing.T, g *game.Game, attackers ...uuid.UUID) (*game.Player, *game.Player) {
	t.Helper()
	atk, def := effectstest.SeatsForCombat(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.DeclareAttack(t, g, def.ID, attackers...)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	return atk, def
}

func TestOldB353InfernoTrapIsCheapOnlyAfterTwoCreaturesDamagedYou(t *testing.T) {
	cast := func(t *testing.T, attackers int) error {
		g := effectstest.NewCatalogGame(t)
		atk, def := effectstest.SeatsForCombat(g)
		var ids []uuid.UUID
		for i := 0; i < attackers; i++ {
			ids = append(ids, effectstest.B12Creature(g, atk.ID, "Raider", "Creature — Bear", 1, 1))
		}
		victim := effectstest.B12Creature(g, atk.ID, "Victim", "Creature — Ogre", 3, 3)
		if attackers > 0 {
			oldB353AttackAndDamage(t, g, ids...)
		}
		id := uuid.New()
		def.Hand.PushTop(game.Card{InstanceID: id, Name: "Inferno Trap", TypeLine: "Instant — Trap", ManaCost: "{3}{R}", OracleID: oldB353InfernoTrap, Owner: def.ID, Controller: def.ID})
		effectstest.FloatForTest(g, def, "R")
		return g.CastSpell(def.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "trap", Targets: effectstest.CardTarget(victim)})
	}
	if err := cast(t, 2); err != nil {
		t.Errorf("two damaging creatures: the trap cost was refused: %v", err)
	}
	if err := cast(t, 1); err == nil {
		t.Error("one damaging creature: the trap cost was offered")
	}
	if err := cast(t, 0); err == nil {
		t.Error("no damage: the trap cost was offered")
	}
}

func TestOldB353InfernoTrapDealsFourToACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Ogre", 5, 5)
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 2, 4)
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Inferno Trap", TypeLine: "Instant — Trap", ManaCost: "{3}{R}", OracleID: oldB353InfernoTrap, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "R", "C", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(small)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 4-toughness creature survived 4 damage")
	}
	if effectstest.DamageOn(t, g, big) != 0 {
		t.Error("the other creature was damaged")
	}
}

func TestOldB353ArrowVolleyTrapNeedsFourAttackersAndDividesFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	atk, def := effectstest.SeatsForCombat(g)
	var raiders []uuid.UUID
	for i := 0; i < 4; i++ {
		raiders = append(raiders, effectstest.B12Creature(g, atk.ID, "Raider", "Creature — Bear", 2, 3))
	}
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	def.Hand.PushTop(game.Card{InstanceID: id, Name: "Arrow Volley Trap", TypeLine: "Instant — Trap", ManaCost: "{3}{W}{W}", OracleID: oldB353ArrowVolleyTrap, Owner: def.ID, Controller: def.ID})
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: raiders[0]}, {Kind: game.TargetCard, ID: raiders[1]}}
	// Before anyone attacks the cheap cost is not offered.
	effectstest.FloatForTest(g, def, "CW")
	if err := g.CastSpell(def.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "trap", Targets: targets, Distribution: map[uuid.UUID]int{raiders[0]: 3, raiders[1]: 2}}); err == nil {
		t.Error("the trap cost was offered with no attackers")
	}
	effectstest.DeclareAttack(t, g, def.ID, raiders...)
	effectstest.FloatForTest(g, def, "CW")
	// A non-attacking creature is not a legal target.
	bystander := effectstest.B12Creature(g, atk.ID, "Bystander", "Creature — Bear", 2, 3)
	if err := g.CastSpell(def.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "trap", Targets: effectstest.CardTarget(bystander), Distribution: map[uuid.UUID]int{bystander: 5}}); err == nil {
		t.Error("a non-attacking creature was targeted")
	}
	if err := g.CastSpell(def.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: "trap", Targets: targets, Distribution: map[uuid.UUID]int{raiders[0]: 3, raiders[1]: 2}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(raiders[0]) {
		t.Error("the creature dealt 3 survived")
	}
	if !g.Battlefield.Contains(raiders[1]) || effectstest.DamageOn(t, g, raiders[1]) != 2 {
		t.Error("the creature dealt 2 should be alive with 2 damage")
	}
}

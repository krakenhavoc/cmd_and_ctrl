package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reality_fracture_creature_f2_test.go — slice fra-creature-f: the two
// cards that need a stack or a combat to be seen working.

// --- Venser, Fervent Forger -----------------------------------------

func TestVenserTokenCopyModeMakesTwoHastyCopiesThatAreSacrificedAtTheEnd(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	target := rfCrFPush(g, opp.ID, "Elk", "Creature — Elk", "", 3, 3)
	rfCrFCast(t, g, "Venser, Fervent Forger", "Legendary Creature — Human Sorcerer", oracleVenserFervent, 5, 3)
	passPriorityAroundTable(t, g)
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, me.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, me.ID, []int{1}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	pickCard(t, g, me.ID, target)
	passPriorityAroundTable(t, g)
	var tokens []uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Elk" && c.IsToken() {
			tokens = append(tokens, c.InstanceID)
			if c.Controller != me.ID {
				t.Fatalf("a copy is controlled by %s, want its creator", c.Controller)
			}
			if !rfCrFHasAbility(t, g, c.InstanceID, "haste") {
				t.Fatal("a copy lacks haste")
			}
		}
	}
	if len(tokens) != 2 {
		t.Fatalf("%d token copies, want 2", len(tokens))
	}
	if !g.Battlefield.Contains(target) {
		t.Fatal("the original left the battlefield")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	for _, id := range tokens {
		if g.Battlefield.Contains(id) {
			t.Fatal("a copy survived the end step")
		}
	}
}

func TestVenserCopyModeCopiesAnOpponentsSpellTwice(t *testing.T) {
	g := newCatalogGame(t)
	caster, venserOwner := g.Seats[0], g.Seats[1]
	casterLife, ownerLife := caster.Life, venserOwner.Life
	bolt := castCatalogSpell(t, g, "Lightning Bolt", "Instant", lightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: venserOwner.ID}})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	venser := uuid.New()
	venserOwner.Hand.PushTop(game.Card{
		InstanceID: venser, Name: "Venser, Fervent Forger", TypeLine: "Legendary Creature — Human Sorcerer",
		OracleID: oracleVenserFervent, Power: 5, Toughness: 3, Owner: venserOwner.ID, Controller: venserOwner.ID,
	})
	if err := g.CastSpell(venserOwner.ID, venser, game.CastSpellParams{}); err != nil {
		t.Fatalf("flash Venser: %v", err)
	}
	for i := 0; i < 8 && latestChoiceOfKindFor(g, game.PendingChoiceModePick, venserOwner.ID) == nil; i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	pick := latestChoiceOfKindFor(g, game.PendingChoiceModePick, venserOwner.ID)
	if pick == nil {
		t.Fatalf("no mode prompt: %+v", g.PendingChoices)
	}
	if err := g.ResolveModePick(pick.ID, venserOwner.ID, []int{0}); err != nil {
		t.Fatalf("ResolveModePick: %v", err)
	}
	pickCard(t, g, venserOwner.ID, bolt)
	// Settle, aiming every copy at the Bolt's caster.
	for i := 0; i < 40 && (g.Stack.Size() > 0 || len(g.PendingChoices) > 0); i++ {
		if p := latestPickTarget(g, venserOwner.ID); p != nil {
			if err := g.ResolvePickTarget(p.ID, venserOwner.ID, game.TargetRef{Kind: game.TargetPlayer, ID: caster.ID}); err != nil {
				t.Fatalf("retarget: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	if got := caster.Life; got != casterLife-6 {
		t.Fatalf("caster life %d, want %d: two copies aimed at the caster", got, casterLife-6)
	}
	if got := venserOwner.Life; got != ownerLife-3 {
		t.Fatalf("Venser's controller life %d, want %d: only the original Bolt hits them", got, ownerLife-3)
	}
}

// --- Yuriko, Blade of the Mighty ------------------------------------

func TestYurikoBladeOfTheMightyBansCastsAndActivationsOnlyDuringCombat(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rfCrFPush(g, me.ID, "Yuriko, Blade of the Mighty", "Legendary Creature — Human Samurai", oracleYurikoBlade, 2, 3)
	witness := rfCrFPush(g, me.ID, "Undulating Witness", "Creature — Serpent", oracleUndulatingWit, 3, 5)
	toMain(t, g)
	rfCrFMana(t, g, me, "{C}{C}{C}{C}")

	advanceTo(t, g, game.StepDeclareAttackers)
	id := rfCrFHand(g, "Idle Trick", "Instant", rfCrFNoncreatureOracle, "", 0, 0)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err == nil {
		t.Fatal("a spell was cast during combat")
	}
	if err := g.ActivateCatalogAbility(me.ID, witness, 0, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("an ability was activated during combat")
	}

	advanceTo(t, g, game.StepPostcombatMain)
	if err := g.ActivateCatalogAbility(me.ID, witness, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activation after combat: %v", err)
	}
	passPriorityAroundTable(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast after combat: %v", err)
	}
}

func TestYurikoBladeOfTheMightyIgnoresALoneAttackerAimedAtAPlaneswalker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	rfCrFPush(g, me.ID, "Yuriko, Blade of the Mighty", "Legendary Creature — Human Samurai", oracleYurikoBlade, 2, 3)
	a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
	walker := pushWalkerForTest(g, opp.ID, "Walker", "", 4)
	advanceTo(t, g, game.StepDeclareAttackers)
	if _, err := g.DeclareAttackers([]game.AttackDeclaration{{Attacker: a, Target: walker}}); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	passPriorityAroundTable(t, g)
	if rfCrFHasAbility(t, g, a, "double strike") {
		t.Fatal("double strike for an attack on a planeswalker")
	}
}

func TestYurikoBladeOfTheMightyGivesDoubleStrikeToALoneAttackerOnly(t *testing.T) {
	for _, tc := range []struct {
		name      string
		attackers int
		want      bool
	}{
		{"alone", 1, true},
		{"two attackers", 2, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
			rfCrFPush(g, me.ID, "Yuriko, Blade of the Mighty", "Legendary Creature — Human Samurai", oracleYurikoBlade, 2, 3)
			a := pushVanillaCreature(g, me.ID, "Bear A", 2, 2)
			b := pushVanillaCreature(g, me.ID, "Bear B", 2, 2)
			advanceTo(t, g, game.StepDeclareAttackers)
			var decls []game.AttackDeclaration
			for _, id := range []uuid.UUID{a, b}[:tc.attackers] {
				decls = append(decls, game.AttackDeclaration{Attacker: id, Target: opp.ID})
			}
			if _, err := g.DeclareAttackers(decls); err != nil {
				t.Fatalf("DeclareAttackers: %v", err)
			}
			passPriorityAroundTable(t, g)
			if got := rfCrFHasAbility(t, g, a, "double strike"); got != tc.want {
				t.Fatalf("double strike on the attacker = %v, want %v", got, tc.want)
			}
		})
	}
}

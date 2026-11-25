package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b116_test.go — Ikoria: Lair of Behemoths, older-sets batch 116
// (#5512). Every test plays the real catalog entry through the engine.

const (
	oldB116MythosOfVadrok          = "0b72fdeb-80b2-4691-93b9-8df5307b4641"
	oldB116SpringjawTrap           = "3a635b73-692f-4345-9628-cb4f9f0cddad"
	oldB116SavaiThundermane        = "620a2037-8bf5-48e0-af94-f4f7d3650c60"
	oldB116ChargeOfTheForeverBeast = "937f0c60-003e-446d-8889-059f7b4a2dd6"
	oldB116ZenithFlare             = "ae005f00-3817-40ca-b8be-069a4085cfee"
	oldB116FlameSpill              = "de6ebf55-f1df-4c3c-bba5-8655b7daf488"
	oldB116WeaponizeTheMonsters    = "fa2154fb-6756-4cf8-85a6-68264fb7a0b2"
	oldB116ImposingVantasaur       = "a0d5fce0-a439-496a-9fcf-9271503f3ce8"
)

func TestOldB116EveryCardIsRegisteredFull(t *testing.T) {
	effectstest.NewCatalogGame(t)
	for _, id := range []string{oldB116MythosOfVadrok, oldB116SpringjawTrap, oldB116SavaiThundermane, oldB116ChargeOfTheForeverBeast, oldB116ZenithFlare, oldB116FlameSpill, oldB116WeaponizeTheMonsters} {
		spec, ok := effects.Lookup(id)
		if !ok {
			t.Errorf("%s is not in the catalog", id)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v caveats %v", spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOldB116SpringjawTrapFlashesInAndShoots(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	trap := effectstest.PushCatalogPermanent(g, me.ID, "Springjaw Trap", "Artifact", oldB116SpringjawTrap, false)
	life := opp.Life
	effectstest.ApaMana(me, "C", "C", "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, trap, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
	if g.Battlefield.Contains(trap) {
		t.Error("the Trap was not sacrificed")
	}
}

func TestOldB116SpringjawTrapCastsWithFlash(t *testing.T) {
	effectstest.NewCatalogGame(t)
	spec, ok := effects.Lookup(oldB116SpringjawTrap)
	if !ok || len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flash" {
		t.Errorf("Springjaw Trap's printed keywords %v, want flash", spec.PrintedKeywords)
	}
}

func TestOldB116WeaponizeTheMonstersSacrificesACreatureToShoot(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	ench := effectstest.PushCatalogPermanent(g, me.ID, "Weaponize the Monsters", "Enchantment", oldB116WeaponizeTheMonsters, false)
	fodder := effectstest.B12Push(g, me.ID, "Fodder", "Creature — Bear", "", 1, 1)
	life := opp.Life
	effectstest.ApaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, ench, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d", opp.Life, life-2)
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the creature was not sacrificed")
	}
}

func TestOldB116WeaponizeTheMonstersNeedsACreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	ench := effectstest.PushCatalogPermanent(g, me.ID, "Weaponize the Monsters", "Enchantment", oldB116WeaponizeTheMonsters, false)
	effectstest.ApaMana(me, "C", "C")
	if err := g.ActivateCatalogAbility(me.ID, ench, 0, game.ActivateAbilityParams{Strict: true, Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}}}); err == nil {
		t.Error("the ability was activated with no creature to sacrifice")
	}
}

func TestOldB116FlameSpillSendsTheExcessToTheController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	bear := effectstest.B12Push(g, opp.ID, "Bear", "Creature — Bear", "", 2, 2)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Flame Spill", "Instant", oldB116FlameSpill, effectstest.CardTarget(bear))
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(bear) {
		t.Error("the Bear survived 2 of 4 damage")
	}
	if opp.Life != life-2 {
		t.Errorf("life %d, want %d: 2 of the 4 was excess", opp.Life, life-2)
	}
	_ = me
}

func TestOldB116FlameSpillOnABigCreatureSendsNothing(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	opp := g.Seats[1]
	wall := effectstest.B12Push(g, opp.ID, "Wall", "Creature — Wall", "", 0, 6)
	life := opp.Life
	effectstest.CastCatalogSpell(t, g, "Flame Spill", "Instant", oldB116FlameSpill, effectstest.CardTarget(wall))
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(wall) {
		t.Error("the Wall died to 4 damage")
	}
	if opp.Life != life {
		t.Errorf("life %d, want %d", opp.Life, life)
	}
}

func TestOldB116ZenithFlareCountsCyclingCardsInYourGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	for i := 0; i < 2; i++ {
		me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Imposing Vantasaur", TypeLine: "Creature — Dinosaur", OracleID: oldB116ImposingVantasaur, Owner: me.ID, Controller: me.ID})
	}
	me.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Plain Bear", TypeLine: "Creature — Bear", Owner: me.ID, Controller: me.ID})
	// A cycling card in an opponent's graveyard is not counted.
	opp.Graveyard.PushTop(game.Card{InstanceID: uuid.New(), Name: "Imposing Vantasaur", TypeLine: "Creature — Dinosaur", OracleID: oldB116ImposingVantasaur, Owner: opp.ID, Controller: opp.ID})
	oppLife, myLife := opp.Life, me.Life
	effectstest.CastCatalogSpell(t, g, "Zenith Flare", "Instant", oldB116ZenithFlare, []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 || me.Life != myLife+2 {
		t.Errorf("opponent %d (want %d), me %d (want %d)", opp.Life, oppLife-2, me.Life, myLife+2)
	}
}

func TestOldB116SavaiThundermanePaysTwoToShootOnACycle(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Savai Thundermane", "Creature — Elemental Cat", oldB116SavaiThundermane, false)
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Bear", "", 2, 3)
	effectstest.CycleFromHand(t, g, "Imposing Vantasaur", "Creature — Dinosaur", oldB116ImposingVantasaur, "{C}")
	life := me.Life
	effectstest.ApaMana(me, "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	// The trigger asks whether to pay {2}.
	effectstest.AnswerPayUnless(t, g, me.ID, true)
	effectstest.AnswerPickTarget(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+2 {
		t.Errorf("life %d, want %d", me.Life, life+2)
	}
	if c := effectstest.DamageOn(t, g, victim); c != 2 {
		t.Errorf("%d damage on the victim, want 2", c)
	}
}

func TestOldB116SavaiThundermaneDoesNothingIfYouDeclineToPay(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Savai Thundermane", "Creature — Elemental Cat", oldB116SavaiThundermane, false)
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Bear", "", 2, 3)
	effectstest.CycleFromHand(t, g, "Imposing Vantasaur", "Creature — Dinosaur", oldB116ImposingVantasaur, "{C}")
	life := me.Life
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPayUnless(t, g, me.ID, false)
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life || effectstest.DamageOn(t, g, victim) != 0 {
		t.Error("something happened without paying")
	}
}

func TestOldB116ChargeOfTheForeverBeastReadsTheRevealedCardsPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Bear", "", 2, 5)
	big := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: big, Name: "Big Beast", TypeLine: "Creature — Beast", Power: 4, Toughness: 4, Owner: me.ID, Controller: me.ID})
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Charge of the Forever-Beast", TypeLine: "Sorcery", ManaCost: "{2}{G}", OracleID: oldB116ChargeOfTheForeverBeast, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "C", "G")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(victim), RevealIDs: []uuid.UUID{big}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageOn(t, g, victim); got != 4 {
		t.Errorf("%d damage on the victim, want 4 (the revealed card's power)", got)
	}
	if !me.Hand.Contains(big) {
		t.Error("the revealed card left the hand")
	}
}

func TestOldB116ChargeOfTheForeverBeastNeedsACreatureCardToReveal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	victim := effectstest.B12Push(g, opp.ID, "Victim", "Creature — Bear", "", 2, 5)
	// Empty the hand of creature cards.
	for _, c := range append([]game.Card(nil), me.Hand.Cards...) {
		me.Hand.Remove(c.InstanceID)
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Charge of the Forever-Beast", TypeLine: "Sorcery", ManaCost: "{2}{G}", OracleID: oldB116ChargeOfTheForeverBeast, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "C", "C", "G")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(victim)}); err == nil {
		t.Error("the spell was cast with no creature card to reveal")
	}
}

func TestOldB116MythosOfVadrokDividesFiveDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	a := effectstest.B12Push(g, opp.ID, "A", "Creature — Bear", "", 2, 3)
	b := effectstest.B12Push(g, opp.ID, "B", "Creature — Bear", "", 2, 3)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mythos of Vadrok", TypeLine: "Sorcery", ManaCost: "{2}{R}{R}", OracleID: oldB116MythosOfVadrok, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "R", "R", "C", "C")
	targets := []game.TargetRef{{Kind: game.TargetCard, ID: a}, {Kind: game.TargetCard, ID: b}}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: targets, Distribution: map[uuid.UUID]int{a: 3, b: 2}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(a) {
		t.Error("A took 3 and should be dead")
	}
	if !g.Battlefield.Contains(b) || effectstest.DamageOn(t, g, b) != 2 {
		t.Errorf("B has %d damage, want 2 and alive", effectstest.DamageOn(t, g, b))
	}
	// Without {W}{U} spent, B can still attack and block.
	if effectstest.RestrictionsOf(t, g, b)&(game.CantAttackOrBlock|game.CantActivate) != 0 {
		t.Error("B was locked without {W}{U} being spent")
	}
}

func TestOldB116MythosOfVadrokLocksThePermanentsWhenWhiteAndBlueWereSpent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	opp := g.Seats[1]
	b := effectstest.B12Push(g, opp.ID, "B", "Creature — Bear", "", 2, 9)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Mythos of Vadrok", TypeLine: "Sorcery", ManaCost: "{2}{R}{R}", OracleID: oldB116MythosOfVadrok, Owner: me.ID, Controller: me.ID})
	effectstest.ApaMana(me, "R", "R", "W", "U")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(b), Distribution: map[uuid.UUID]int{b: 5}}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.RestrictionsOf(t, g, b)&(game.CantAttackOrBlock|game.CantActivate) == 0 {
		t.Error("B was not locked although {W}{U} was spent")
	}
}

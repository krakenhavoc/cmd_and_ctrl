package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b417_cards_test.go — Return to Ravnica, batch 417 (S82 older sets,
// #5566): nine cards of damage. Every test plays the real catalog entry
// through the engine.

const (
	oldB417Return     = "52d1d575-3e7f-4df1-8c89-aee0cac03aca"
	oldB417Courier    = "82184f0a-54a1-4ae0-9b99-2b3450c6e544"
	oldB417Lobber     = "84dffb4b-02d7-48e4-9a95-43f88cd52bf3"
	oldB417Pyro       = "8da5c156-3688-4d6e-b5c4-95c082d1b426"
	oldB417Rig        = "ba47a07e-890c-411a-9fd0-947c2f738da2"
	oldB417SkullRend  = "d29d59f1-e00c-4217-8185-2085a4142e3b"
	oldB417Staticast  = "d63adbed-07d3-44e4-83f8-1cbdfc186df4"
	oldB417Backlash   = "f51ed0b4-c144-458e-b97c-618b9e1cf25d"
	oldB417Ringleader = "f8ce8573-7c87-4f47-a0d4-14bd335f2c1d"
)

func TestOldB417EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB417Return: "Rakdos's Return", oldB417Courier: "Crosstown Courier", oldB417Lobber: "Lobber Crew",
		oldB417Pyro: "Pyroconvergence", oldB417Rig: "Volatile Rig", oldB417SkullRend: "Skull Rend",
		oldB417Staticast: "Izzet Staticaster", oldB417Backlash: "Essence Backlash", oldB417Ringleader: "Rakdos Ringleader",
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

func oldB417Hand(p *game.Player, n int) {
	for i := 0; i < n; i++ {
		p.Hand.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Instant", Owner: p.ID, Controller: p.ID})
	}
}

func oldB417CastX(t *testing.T, g *game.Game, x int, targets []game.TargetRef) {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Rakdos's Return", TypeLine: "Sorcery", OracleID: oldB417Return,
		Owner: me.ID, Controller: me.ID})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: x, Targets: targets}); err != nil {
		t.Fatalf("cast Rakdos's Return: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB417RakdossReturnDamagesAnOpponentAndTheyDiscardX(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	oldB417Hand(opp, 5)
	life, hand := opp.Life, opp.Hand.Size()
	oldB417CastX(t, g, 3, []game.TargetRef{effectstest.PlayerTarget(opp.ID)})
	_ = me
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d", opp.Life, life-3)
	}
	if got := effectstest.DiscardOwed(g, opp.ID); got != 3 {
		t.Fatalf("the opponent owes %d discards, want 3", got)
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	if opp.Hand.Size() != hand-3 {
		t.Errorf("hand %d, want %d", opp.Hand.Size(), hand-3)
	}
}

func TestOldB417RakdossReturnAtAPlaneswalkerMakesItsControllerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	_, opp := effectstest.SeatsForCombat(g)
	oldB417Hand(opp, 4)
	pw := effectstest.P1g2Planeswalker(g, opp.ID, 3)
	hand := opp.Hand.Size()
	oldB417CastX(t, g, 2, effectstest.CardTarget(pw))
	if effectstest.DiscardOwed(g, opp.ID) != 2 {
		t.Fatalf("the planeswalker's controller owes %d discards, want 2", effectstest.DiscardOwed(g, opp.ID))
	}
	effectstest.DiscardFromHand(t, g, opp.ID)
	if opp.Hand.Size() != hand-2 {
		t.Errorf("hand %d, want %d", opp.Hand.Size(), hand-2)
	}
}

func TestOldB417RakdossReturnRefusesYourselfAsATarget(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := effectstest.HandCardFull(me, "Rakdos's Return", "Sorcery", "{X}{B}{R}", oldB417Return, []string{"B", "R"})
	effectstest.AdvanceToMain(t, g)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, XValue: 1, Targets: []game.TargetRef{effectstest.PlayerTarget(me.ID)}}); err == nil {
		t.Error("Rakdos's Return targeted its own caster")
	}
}

func TestOldB417CrosstownCourierMillsThatMany(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	courier := effectstest.B12Push(g, me.ID, "Crosstown Courier", "Creature — Vedalken", oldB417Courier, 2, 1)
	for i := 0; i < 6; i++ {
		effectstest.PushLibraryCardForTest(opp, game.Card{Name: "Lib", TypeLine: "Instant"})
	}
	lib, yard := opp.Library.Size(), opp.Graveyard.Size()
	effectstest.Pr7bAttack(t, g, opp.ID, courier)
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if got := opp.Library.Size(); got != lib-2 {
		t.Errorf("library %d, want %d", got, lib-2)
	}
	if got := opp.Graveyard.Size(); got != yard+2 {
		t.Errorf("graveyard %d, want %d", got, yard+2)
	}
}

func oldB417Multicolored(t *testing.T, g *game.Game, me *game.Player, colors ...string) {
	t.Helper()
	id := effectstest.HandCardFull(me, "Gold Guy", "Creature — Test", "", "", colors)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB417LobberCrewPingsAndUntapsOnAMulticoloredSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	crew := effectstest.B12Push(g, me.ID, "Lobber Crew", "Creature — Goblin Warrior", oldB417Lobber, 0, 4)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	if err := g.ActivateCatalogAbility(me.ID, crew, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-1 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-1)
	}
	if !effectstest.B16Tapped(t, g, crew) {
		t.Fatal("the Crew did not tap")
	}
	oldB417Multicolored(t, g, me, "R")
	if !effectstest.B16Tapped(t, g, crew) {
		t.Error("a monocolored spell untapped the Crew")
	}
	oldB417Multicolored(t, g, me, "R", "G")
	if effectstest.B16Tapped(t, g, crew) {
		t.Error("a multicolored spell did not untap the Crew")
	}
	if !effectstest.HasEffectiveKeyword(t, g, crew, "defender") {
		t.Error("the Crew has no defender")
	}
}

func TestOldB417PyroconvergenceDamagesOnAMulticoloredSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	effectstest.B12Push(g, me.ID, "Pyroconvergence", "Enchantment", oldB417Pyro, 0, 0)
	effectstest.AdvanceToMain(t, g)
	life := opp.Life
	id := effectstest.HandCardFull(me, "Gold Guy", "Creature — Test", "", "", []string{"R", "W"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	if p := effectstest.FirstPendingTarget(g); p != nil {
		if err := g.ResolvePickTarget(p.ID, me.ID, effectstest.PlayerTarget(opp.ID)); err != nil {
			t.Fatalf("pick target: %v", err)
		}
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-2)
	}
	id = effectstest.HandCardFull(me, "Mono Guy", "Creature — Test", "", "", []string{"R"})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("a monocolored spell dealt damage: life %d", opp.Life)
	}
}

func TestOldB417SkullRendHitsEveryOpponentAndTheyDiscardTwoAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	for _, p := range g.Seats {
		oldB417Hand(p, 4)
	}
	lives := map[uuid.UUID]int{}
	hands := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID], hands[p.ID] = p.Life, p.Hand.Size()
	}
	effectstest.CastCatalogSpell(t, g, "Skull Rend", "Sorcery", oldB417SkullRend, nil)
	effectstest.PassPriorityAroundTable(t, g)
	for _, p := range g.Seats {
		if p.ID == me.ID {
			if p.Life != lives[p.ID] || p.Hand.Size() != hands[p.ID] {
				t.Errorf("the caster: life %d hand %d", p.Life, p.Hand.Size())
			}
			continue
		}
		if p.Life != lives[p.ID]-2 {
			t.Errorf("opponent life %d, want %d", p.Life, lives[p.ID]-2)
		}
		if p.Hand.Size() != hands[p.ID]-2 {
			t.Errorf("opponent hand %d, want %d", p.Hand.Size(), hands[p.ID]-2)
		}
	}
}

func TestOldB417IzzetStaticasterPingsTheTargetAndEveryCreatureWithItsName(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	st := effectstest.B12Push(g, me.ID, "Izzet Staticaster", "Creature — Human Wizard", oldB417Staticast, 0, 3)
	a := effectstest.B12Creature(g, opp.ID, "Goblin Token", "Creature — Goblin", 1, 3)
	b := effectstest.B12Creature(g, opp.ID, "Goblin Token", "Creature — Goblin", 1, 3)
	c := effectstest.B12Creature(g, me.ID, "Goblin Token", "Creature — Goblin", 1, 3)
	other := effectstest.B12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, st, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardTarget(a)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{a, b, c} {
		if got := effectstest.DamageOn(t, g, id); got != 1 {
			t.Errorf("a Goblin has %d damage, want 1", got)
		}
	}
	if got := effectstest.DamageOn(t, g, other); got != 0 {
		t.Errorf("the Bear has %d damage, want 0", got)
	}
	for _, kw := range []string{"flash", "haste"} {
		if !effectstest.HasEffectiveKeyword(t, g, st, kw) {
			t.Errorf("the Staticaster has no %s", kw)
		}
	}
}

func TestOldB417EssenceBacklashCountersACreatureSpellAndDamagesItsControllerByPower(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	caster, them := g.Seats[0], g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := uuid.New()
	caster.Hand.PushTop(game.Card{InstanceID: victim, Name: "Big Guy", TypeLine: "Creature — Giant", Power: 3, Toughness: 3, Owner: caster.ID, Controller: caster.ID})
	if err := g.CastSpell(caster.ID, victim, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast Big Guy: %v", err)
	}
	back := effectstest.HandCardFull(them, "Essence Backlash", "Instant", "{2}{U}{R}", oldB417Backlash, []string{"U", "R"})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	life := caster.Life
	if err := g.CastSpell(them.ID, back, game.CastSpellParams{Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("cast Essence Backlash: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the creature spell resolved")
	}
	if want := life - 3; caster.Life != want {
		t.Errorf("caster life %d, want %d", caster.Life, want)
	}
}

func TestOldB417EssenceBacklashRefusesANoncreatureSpell(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	them := g.Seats[1]
	effectstest.ToMainForCost(t, g)
	victim := effectstest.CastCatalogSpell(t, g, "A Sorcery", "Sorcery", "", nil)
	back := effectstest.HandCardFull(them, "Essence Backlash", "Instant", "{2}{U}{R}", oldB417Backlash, []string{"U", "R"})
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if err := g.CastSpell(them.ID, back, game.CastSpellParams{Strict: true, Targets: effectstest.CardRefs(victim)}); err == nil {
		t.Error("Essence Backlash targeted a noncreature spell")
	}
}

func TestOldB417RakdosRingleaderMakesTheDamagedPlayerDiscardAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.SeatsForCombat(g)
	rl := effectstest.B12Push(g, me.ID, "Rakdos Ringleader", "Creature — Skeleton Warrior", oldB417Ringleader, 3, 1)
	oldB417Hand(opp, 3)
	hand, yard := opp.Hand.Size(), opp.Graveyard.Size()
	effectstest.Pr7bAttack(t, g, opp.ID, rl)
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Hand.Size() != hand-1 || opp.Graveyard.Size() != yard+1 {
		t.Errorf("hand %d graveyard %d, want %d and %d", opp.Hand.Size(), opp.Graveyard.Size(), hand-1, yard+1)
	}
	if !effectstest.HasEffectiveKeyword(t, g, rl, "first strike") {
		t.Error("the Ringleader has no first strike")
	}
}

func TestOldB417RakdosRingleaderRegeneratesForB(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rl := effectstest.B12Push(g, me.ID, "Rakdos Ringleader", "Creature — Skeleton Warrior", oldB417Ringleader, 3, 1)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatMana(t, g, me, "{B}")
	if err := g.ActivateCatalogAbility(me.ID, rl, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, rl)
	if !g.Battlefield.Contains(rl) {
		t.Error("the Ringleader did not regenerate")
	}
}

func TestOldB417VolatileRigFlipsWhenDamagedAndWhenItDies(t *testing.T) {
	sawKeep, sawLose := false, false
	for i := 0; i < 80 && !(sawKeep && sawLose); i++ {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		rig := effectstest.B12Push(g, me.ID, "Volatile Rig", "Artifact Creature — Construct", oldB417Rig, 4, 4)
		effectstest.Pr6Damage(t, g, effectstest.B12Creature(g, g.Seats[1].ID, "Pinger", "Creature — Test", 1, 1), rig, 1)
		effectstest.PassPriorityAroundTable(t, g)
		flip := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if flip == nil {
			t.Fatal("no coin call when the Rig was dealt damage")
		}
		if err := g.ResolveCoinCall(flip.ID, me.ID, "heads"); err != nil {
			t.Fatalf("coin: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		flips := effectstest.RandomEvents(g, game.EventFlipCoin)
		won := len(flips) > 0 && flips[len(flips)-1].Won
		if won {
			sawKeep = true
			if !g.Battlefield.Contains(rig) {
				t.Error("won the flip, but the Rig was sacrificed")
			}
		} else {
			sawLose = true
			if g.Battlefield.Contains(rig) {
				t.Error("lost the flip, but the Rig is still on the battlefield")
			}
		}
	}
	if !sawKeep || !sawLose {
		t.Fatalf("never saw both outcomes (keep %v, lose %v)", sawKeep, sawLose)
	}
}

func TestOldB417VolatileRigDeathBlastHitsEveryCreatureAndPlayerOnALostFlip(t *testing.T) {
	sawBlast, sawQuiet := false, false
	for i := 0; i < 80 && !(sawBlast && sawQuiet); i++ {
		g := effectstest.NewCatalogGame(t)
		me, opp := effectstest.SeatsForCombat(g)
		rig := effectstest.B12Push(g, me.ID, "Volatile Rig", "Artifact Creature — Construct", oldB417Rig, 4, 4)
		bystander := effectstest.B12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 9)
		life := []int{me.Life, opp.Life}
		effectstest.P1Destroy(t, g, rig)
		effectstest.PassPriorityAroundTable(t, g)
		flip := effectstest.LatestChoiceOfKind(g, game.PendingChoiceCoinCall)
		if flip == nil {
			t.Fatal("no coin call when the Rig died")
		}
		if err := g.ResolveCoinCall(flip.ID, me.ID, "heads"); err != nil {
			t.Fatalf("coin: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		flips := effectstest.RandomEvents(g, game.EventFlipCoin)
		won := len(flips) > 0 && flips[len(flips)-1].Won
		if won {
			sawQuiet = true
			if me.Life != life[0] || opp.Life != life[1] || effectstest.DamageOn(t, g, bystander) != 0 {
				t.Error("won the flip, but something was damaged")
			}
		} else {
			sawBlast = true
			if me.Life != life[0]-4 || opp.Life != life[1]-4 || effectstest.DamageOn(t, g, bystander) != 4 {
				t.Errorf("lost the flip: life %d/%d, bystander damage %d", me.Life, opp.Life, effectstest.DamageOn(t, g, bystander))
			}
		}
	}
	if !sawBlast || !sawQuiet {
		t.Fatalf("never saw both outcomes (blast %v, quiet %v)", sawBlast, sawQuiet)
	}
}

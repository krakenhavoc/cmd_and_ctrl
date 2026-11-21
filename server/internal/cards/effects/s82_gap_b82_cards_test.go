package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	gapB82DinoDNA      = "26e988da-52cf-4382-8646-a27c5ad1b92d"
	gapB82Dilophosaur  = "19f2c2c1-cd29-4e66-b5d9-e85e1979d843"
	gapB82CompySwarm   = "f5bd96dd-0d1f-4923-a014-6c8dcb3b2266"
	gapB82Giganotosaur = "1c197b3c-f540-4766-aeaa-ae09d22214cb"
	gapB82Indoraptor   = "0dd4bc7b-2ddf-4ca4-912c-59685c41c3b5"
	gapB82Owen         = "4eda40a2-39e5-4b0e-ba6d-8a6f89d6f55f"
	gapB82Blue         = "4f13ed2e-7790-499f-87bf-07480b567811"
	gapB82Velociraptor = "bca33fb6-cd37-4259-8c0b-7a4675754a00"
	gapB82Pteranodon   = "0b1f6e77-55db-4638-9030-3f4e4d37eb1b"
	gapB82Permission   = "95aae5ba-7634-4cad-a0be-496c3398fa3c"
	gapB82SavageOrder  = "8dd5b7d5-3d92-40e2-9572-bf70ae591bed"
)

func TestGapB82EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB82DinoDNA:      "Dino DNA",
		gapB82Dilophosaur:  "Spitting Dilophosaurus",
		gapB82CompySwarm:   "Compy Swarm",
		gapB82Giganotosaur: "Grim Giganotosaurus",
		gapB82Indoraptor:   "Indoraptor, the Perfect Hybrid",
		gapB82Owen:         "Owen Grady, Raptor Trainer",
		gapB82Blue:         "Blue, Loyal Raptor",
		gapB82Velociraptor: "Hunting Velociraptor",
		gapB82Pteranodon:   "Swooping Pteranodon",
		gapB82Permission:   "Permission Denied",
		gapB82SavageOrder:  "Savage Order",
	} {
		spec, ok := Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v, %+v", name, ok, spec)
		}
	}
}

// gapB82Settle passes priority until the stack is empty, answering any target
// prompt `me` owes with `picks`.
func gapB82Settle(t *testing.T, g *game.Game, me uuid.UUID, picks ...game.TargetRef) {
	t.Helper()
	next := 0
	for i := 0; i < 64; i++ {
		if p := latestPickTarget(g, me); p != nil {
			var err error
			if next < len(picks) {
				err = g.ResolvePickTarget(p.ID, me, picks[next])
				next++
			} else {
				err = g.ResolvePickTargets(p.ID, me, nil)
			}
			if err != nil {
				t.Fatalf("answering the target prompt: %v", err)
			}
			continue
		}
		if stackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never emptied")
}

func gapB82Cast(t *testing.T, g *game.Game, me *game.Player, c game.Card, mana string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	advanceToMain(t, g)
	id := avB24Hand(g, me, c)
	if mana != "" {
		floatForTest(g, me, mana)
	}
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	return id
}

func TestGapB82PermissionDeniedCountersAndBansOpposingNoncreatureCasts(t *testing.T) {
	g := newCatalogGame(t)
	caster := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceToMain(t, g)
	spell := avB24Hand(g, caster, game.Card{Name: "Divination", TypeLine: "Sorcery"})
	if err := g.CastSpell(caster.ID, spell, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	pd := avB24Hand(g, opp, game.Card{Name: "Permission Denied", TypeLine: "Instant", OracleID: gapB82Permission})
	if err := g.CastSpell(opp.ID, pd, game.CastSpellParams{Targets: b16TargetCard(spell)}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !caster.Graveyard.Contains(spell) || g.Stack.Contains(spell) {
		t.Error("the noncreature spell was not countered")
	}
	// The caster is the opponent of Permission Denied's controller: no more noncreature spells.
	next := avB24Hand(g, caster, game.Card{Name: "Another", TypeLine: "Instant"})
	if err := g.CastSpell(caster.ID, next, game.CastSpellParams{Strict: true}); err == nil {
		t.Error("an opponent cast a noncreature spell after Permission Denied")
	}
	mine := avB24Hand(g, opp, game.Card{Name: "Mine", TypeLine: "Instant"})
	if err := g.CastSpell(opp.ID, mine, game.CastSpellParams{Strict: true}); err != nil {
		t.Errorf("Permission Denied's controller was barred: %v", err)
	}
	// Illegal target: a creature spell.
	crit := avB24Hand(g, caster, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	_ = crit
}

func TestGapB82SpittingDilophosaurusPutsACounterAndStopsBlockers(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	shrunk := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Shrunk", TypeLine: "Creature — Bear", Power: 3, Toughness: 3,
		Counters: map[string]int{game.CounterMinusOne: 1}})
	plain := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Plain", TypeLine: "Creature — Bear", Power: 3, Toughness: 3})
	mine := apaPush(g, me.ID, me.ID, game.Card{Name: "Mine", TypeLine: "Creature — Bear", Power: 3, Toughness: 3,
		Counters: map[string]int{game.CounterMinusOne: 1}})
	dilo := apaPush(g, me.ID, me.ID, game.Card{Name: "Spitting Dilophosaurus", TypeLine: "Creature — Dinosaur", OracleID: gapB82Dilophosaur, Power: 3, Toughness: 2})
	_ = dilo
	cant := func(id uuid.UUID) bool {
		c, _ := g.LookupCardForEffect(id)
		return c.Effective().Restrictions&game.CantBlock != 0
	}
	if !cant(shrunk) || cant(plain) || cant(mine) {
		t.Errorf("can't block: shrunk %v plain %v mine %v, want true false false", cant(shrunk), cant(plain), cant(mine))
	}
	// Entering puts a -1/-1 counter on up to one target creature.
	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	opp2 := g2.Seats[(g2.Turn.ActiveSeat+1)%len(g2.Seats)]
	victim := apaPush(g2, opp2.ID, opp2.ID, game.Card{Name: "Victim", TypeLine: "Creature — Bear", Power: 3, Toughness: 3})
	gapB82Cast(t, g2, me2, game.Card{Name: "Spitting Dilophosaurus", TypeLine: "Creature — Dinosaur", ManaCost: "{2}{B}", OracleID: gapB82Dilophosaur, Power: 3, Toughness: 2}, "BCC", game.CastSpellParams{})
	gapB82Settle(t, g2, me2.ID, b16TargetCard(victim)[0])
	if got := counterOnBattlefieldCard(g2, victim, game.CounterMinusOne); got != 1 {
		t.Errorf("-1/-1 counters on the target: %d, want 1", got)
	}
}

func TestGapB82CompySwarmCopiesItselfTappedOnlyIfACreatureDied(t *testing.T) {
	for _, died := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		apaPush(g, me.ID, me.ID, game.Card{Name: "Compy Swarm", TypeLine: "Creature — Dinosaur", OracleID: gapB82CompySwarm, Power: 2, Toughness: 2})
		if died {
			small := apaPush(g, me.ID, me.ID, game.Card{Name: "Mouse", TypeLine: "Creature — Mouse", Power: 1, Toughness: 1})
			gapB82Cast(t, g, me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", OracleID: pr7bLightningBolt}, "", game.CastSpellParams{Targets: b16TargetCard(small)})
			passPriorityAroundTable(t, g)
		}
		advanceTo(t, g, game.StepEnd)
		passPriorityAroundTable(t, g)
		n, tapped := 0, 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Compy Swarm" {
				n++
				if c.Tapped && c.IsToken() {
					tapped++
				}
			}
		}
		want := 1
		if died {
			want = 2
		}
		if n != want || (died && tapped != 1) {
			t.Errorf("creature died %v: %d Compy Swarms (%d tapped tokens), want %d", died, n, tapped, want)
		}
	}
}

func TestGapB82IndoraptorBloodthirstXAndEnrage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	gptB1DamagePlayer(g, opp.ID, 3)
	id := gapB82Cast(t, g, me, game.Card{Name: "Indoraptor, the Perfect Hybrid", TypeLine: "Legendary Creature — Dinosaur Mutant",
		ManaCost: "{1}{B/G}{R}", OracleID: gapB82Indoraptor, Power: 3, Toughness: 1}, "BRC", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if got := counterOnBattlefieldCard(g, id, game.CounterPlusOne); got != 3 {
		t.Fatalf("Indoraptor entered with %d counters, want 3 (damage dealt to opponents this turn)", got)
	}
	// Enrage: damage to Indoraptor makes a random opponent take damage equal to its power.
	before := 0
	for _, p := range g.Seats {
		if p.ID != me.ID {
			before += p.Life
		}
	}
	bolt := avB24Hand(g, me, game.Card{Name: "Lightning Bolt", TypeLine: "Instant", OracleID: pr7bLightningBolt})
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{Targets: b16TargetCard(id)}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	// The chosen opponent controls the Pinger (a nontoken creature) and so is asked; they take the damage.
	for _, p := range g.Seats {
		if p.ID != me.ID && latestOptionPickFor(g, p.ID) != nil {
			bbdAnswerOption(t, g, p.ID, gapB82IndoraptorTakeLabel)
		}
	}
	passPriorityAroundTable(t, g)
	after := 0
	for _, p := range g.Seats {
		if p.ID != me.ID {
			after += p.Life
		}
	}
	if before-after != 6 {
		t.Errorf("opponents lost %d life, want 6 (Indoraptor's power as the trigger resolved)", before-after)
	}
}

func TestGapB82GrimGiganotosaurusMonstrosityCostsLessAndWipes(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	giga := apaPush(g, me.ID, me.ID, game.Card{Name: "Grim Giganotosaurus", TypeLine: "Creature — Dinosaur", OracleID: gapB82Giganotosaur, Power: 10, Toughness: 10})
	big1 := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Big One", TypeLine: "Creature — Giant", Power: 4, Toughness: 4})
	big2 := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Big Two", TypeLine: "Creature — Giant", Power: 5, Toughness: 5})
	small := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Small", TypeLine: "Creature — Mouse", Power: 3, Toughness: 3})
	rock := apaPush(g, me.ID, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	land := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	advanceToMain(t, g)
	// {10}{B}{G} less {2} for two creatures with power 4 or greater: ten mana.
	floatForTest(g, me, "BGCCCCCCCC")
	if err := g.ActivateCatalogAbility(me.ID, giga, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate with the reduced cost: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := counterOnBattlefieldCard(g, giga, game.CounterPlusOne); got != 10 {
		t.Errorf("%d +1/+1 counters, want 10", got)
	}
	for _, id := range []uuid.UUID{big1, big2, small, rock} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%s survived the wipe", id)
		}
	}
	if !g.Battlefield.Contains(giga) || !g.Battlefield.Contains(land) {
		t.Error("the Giganotosaurus or the land was destroyed")
	}
}

func TestGapB82OwenGradyPutsYourChoiceOfCounterOnADinosaur(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	owen := apaPush(g, me.ID, me.ID, game.Card{Name: "Owen Grady, Raptor Trainer", TypeLine: "Legendary Creature — Human Soldier Scientist", OracleID: gapB82Owen, Power: 3, Toughness: 2})
	raptor := apaPush(g, me.ID, me.ID, game.Card{Name: "Raptor", TypeLine: "Creature — Dinosaur", Power: 2, Toughness: 2})
	bear := apaPush(g, me.ID, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	advanceToMain(t, g)
	if err := g.ActivateCatalogAbility(me.ID, owen, 0, game.ActivateAbilityParams{Targets: b16TargetCard(bear)}); err == nil {
		t.Fatal("a non-Dinosaur was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, owen, 0, game.ActivateAbilityParams{Targets: b16TargetCard(raptor)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	bbdAnswerOption(t, g, me.ID, "A trample counter")
	passPriorityAroundTable(t, g)
	if got := counterOnBattlefieldCard(g, raptor, game.CounterTrample); got != 1 {
		t.Errorf("trample counters on the Dinosaur: %d, want 1", got)
	}
}

func TestGapB82BlueGivesEachOtherDinosaurItsKindsOfCounters(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	apaPush(g, me.ID, me.ID, game.Card{Name: "Blue, Loyal Raptor", TypeLine: "Legendary Creature — Dinosaur", OracleID: gapB82Blue, Power: 5, Toughness: 4,
		Counters: map[string]int{game.CounterReach: 1, game.CounterPlusOne: 2}})
	dino := gapB82Cast(t, g, me, game.Card{Name: "Raptor", TypeLine: "Creature — Dinosaur", Power: 2, Toughness: 2}, "", game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if counterOnBattlefieldCard(g, dino, game.CounterReach) != 1 || counterOnBattlefieldCard(g, dino, game.CounterPlusOne) != 1 {
		t.Errorf("Dinosaur counters: reach %d, +1/+1 %d, want one of each kind Blue has",
			counterOnBattlefieldCard(g, dino, game.CounterReach), counterOnBattlefieldCard(g, dino, game.CounterPlusOne))
	}
	bear := avB24Hand(g, me, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatal(err)
	}
	passPriorityAroundTable(t, g)
	if counterOnBattlefieldCard(g, bear, game.CounterReach) != 0 {
		t.Error("a non-Dinosaur entered with a counter")
	}
}

func TestGapB82HuntingVelociraptorGrantsProwlToDinosaurSpells(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	apaPush(g, me.ID, me.ID, game.Card{Name: "Hunting Velociraptor", TypeLine: "Creature — Dinosaur", OracleID: gapB82Velociraptor, Power: 3, Toughness: 2})
	attacker := apaPush(g, me.ID, me.ID, game.Card{Name: "Raptor", TypeLine: "Creature — Dinosaur", Power: 2, Toughness: 2})
	spell := avB24Hand(g, me, game.Card{Name: "Big Dino", TypeLine: "Creature — Dinosaur", ManaCost: "{5}{G}{G}", Power: 7, Toughness: 7})
	advanceToMain(t, g)
	floatForTest(g, me, "RCC")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{AlternativeCost: GrantedAltCostProwl, Strict: true}); err == nil {
		t.Fatal("prowl was available before any combat damage")
	}
	pr7bAttack(t, g, opp.ID, attacker)
	pr7bDamageStep(t, g)
	advanceTo(t, g, game.StepPostcombatMain)
	floatForTest(g, me, "RCC")
	if err := g.CastSpell(me.ID, spell, game.CastSpellParams{AlternativeCost: GrantedAltCostProwl, Strict: true}); err != nil {
		t.Fatalf("prowl after dealing combat damage with a Dinosaur: %v", err)
	}
}

func TestGapB82SavageOrderSacrificesAPowerFourCreatureAndFetchesADinosaur(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	small := apaPush(g, me.ID, me.ID, game.Card{Name: "Small", TypeLine: "Creature — Bear", Power: 3, Toughness: 3})
	big := apaPush(g, me.ID, me.ID, game.Card{Name: "Big", TypeLine: "Creature — Bear", Power: 4, Toughness: 4})
	dino := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: dino, Name: "Library Dino", TypeLine: "Creature — Dinosaur", Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	advanceToMain(t, g)
	id := avB24Hand(g, me, game.Card{Name: "Savage Order", TypeLine: "Sorcery", OracleID: gapB82SavageOrder})
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{SacrificeIDs: []uuid.UUID{small}, Strict: true}); err == nil {
		t.Fatal("a power-3 creature paid the sacrifice")
	}
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{SacrificeIDs: []uuid.UUID{big}, Strict: true}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	for i := 0; i < 16 && !g.Battlefield.Contains(dino); i++ {
		if p := latestChoiceOfKindFor(g, game.PendingChoiceSearchLibrary, me.ID); p != nil {
			if err := g.ResolveSearchLibrary(p.ID, me.ID, []uuid.UUID{dino}); err != nil {
				t.Fatalf("search: %v", err)
			}
			continue
		}
		if err := g.PassPriority(); err != nil {
			t.Fatal(err)
		}
	}
	if g.Battlefield.Contains(big) || !g.Battlefield.Contains(dino) {
		t.Fatalf("big on battlefield %v, dinosaur on battlefield %v", g.Battlefield.Contains(big), g.Battlefield.Contains(dino))
	}
	found, _ := g.LookupCardForEffect(dino)
	if !game.HasKeyword(&found, "indestructible") {
		t.Error("the found Dinosaur is not indestructible")
	}
}

func TestGapB82DinoDNAExilesThenCopiesAsASixBySixDinosaur(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	dna := apaPush(g, me.ID, me.ID, game.Card{Name: "Dino DNA", TypeLine: "Artifact", OracleID: gapB82DinoDNA})
	target := avB24Graveyard(me, game.Card{Name: "Grizzly Bears", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, Colors: []string{"G"}})
	advanceToMain(t, g)
	floatForTest(g, me, "C")
	if err := g.ActivateCatalogAbility(me.ID, dna, 0, game.ActivateAbilityParams{Targets: b16TargetCard(target)}); err != nil {
		t.Fatalf("imprint: %v", err)
	}
	passPriorityAroundTable(t, g)
	if me.Graveyard.Contains(target) {
		t.Fatal("the creature card was not exiled")
	}
	floatForTest(g, me, "CCCCCC")
	if err := g.ActivateCatalogAbility(me.ID, dna, 1, game.ActivateAbilityParams{Targets: b16TargetCard(target)}); err != nil {
		t.Fatalf("copy: %v", err)
	}
	passPriorityAroundTable(t, g)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Grizzly Bears" && c.IsToken() {
			found = true
			v := layeredCard(t, g, c.InstanceID)
			if v.CurrentPower() != 6 || v.CurrentToughness() != 6 || !v.HasSubtype("Dinosaur") || v.HasSubtype("Bear") || !game.HasKeyword(&v, "trample") {
				t.Errorf("token is %d/%d dinosaur=%v bear=%v trample=%v", v.CurrentPower(), v.CurrentToughness(), v.HasSubtype("Dinosaur"), v.HasSubtype("Bear"), game.HasKeyword(&v, "trample"))
			}
		}
	}
	if !found {
		t.Error("no token copy was created")
	}
}

func TestGapB82SwoopingPteranodonStealsAndTheLandBurnsItAtEndStep(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	victim := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Victim", TypeLine: "Creature — Bear", Power: 2, Toughness: 3, Tapped: true})
	land := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	gapB82Cast(t, g, me, game.Card{Name: "Swooping Pteranodon", TypeLine: "Creature — Dinosaur", ManaCost: "{3}{R}{W}", OracleID: gapB82Pteranodon,
		Power: 3, Toughness: 3, Keywords: []string{"flying", "haste"}}, "RWCCC", game.CastSpellParams{})
	gapB82Settle(t, g, me.ID, game.TargetRef{Kind: game.TargetCard, ID: victim}, game.TargetRef{Kind: game.TargetCard, ID: land})
	v, _ := g.LookupCardForEffect(victim)
	if v.Controller != me.ID || v.Tapped {
		t.Fatalf("victim controller %v (want me), tapped %v (want untapped)", v.Controller == me.ID, v.Tapped)
	}
	layered := layeredCard(t, g, victim)
	if !game.HasKeyword(&layered, "flying") || !game.HasKeyword(&layered, "haste") {
		t.Error("the stolen creature lacks flying or haste")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(victim) {
		t.Error("the land's 3 damage did not kill the 2/3 creature")
	}
}

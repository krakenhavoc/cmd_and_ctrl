package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b101_cards_test.go — S82 OLD batch 101 (#5510, Jumpstart): graveyard
// recursion. Each test plays the real catalog entry.

const (
	oldB101Accomplice  = "16616aa2-5490-4fab-9de4-f37ca8c8be08"
	oldB101Archaeo     = "5a4553fc-f945-4094-b231-2d378ed6609e"
	oldB101Gravewaker  = "7ceec1c7-c1f9-4d47-a203-6ee98c5d30d2"
	oldB101Phoenix     = "e4e0e23b-f172-4e1c-99d0-bb6f66afbc1c"
	oldB101Attendant   = "e9195f10-8cc2-482a-9e28-0be9a37099b4"
	oldB101Scarecrone  = "f45716dc-c613-4b14-865d-419395fb02f6"
	oldB101ArtifactCrt = "Artifact Creature — Construct"
)

func TestOldB101EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB101Accomplice: "Ghoulcaller's Accomplice",
		oldB101Archaeo:    "Archaeomender",
		oldB101Gravewaker: "Gravewaker",
		oldB101Phoenix:    "Lightning Phoenix",
		oldB101Attendant:  "Dutiful Attendant",
		oldB101Scarecrone: "Scarecrone",
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

func TestOldB101AccompliceMakesAZombieFromTheGraveyardAtSorcerySpeed(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := effectstest.Top100ActiveSeat(g)
	id := effectstest.SeedGraveyardCard(t, g, "Ghoulcaller's Accomplice", "Creature — Human Rogue", oldB101Accomplice)
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated without the mana")
	}
	effectstest.FloatMana(t, g, me, "{B}{B}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, id, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("the Accomplice was not exiled")
	}
	zombies := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Zombie" && c.Controller == me.ID && c.Power == 2 && c.Toughness == 2 {
			zombies++
		}
	}
	if zombies != 1 {
		t.Errorf("%d 2/2 Zombie tokens, want 1", zombies)
	}
}

func TestOldB101ArchaeomenderReturnsAnArtifactCardOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	rock := effectstest.PushGraveyardCardTyped(me, "Rock", "Artifact")
	bear := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.DftABCast(t, g, me, game.Card{Name: "Archaeomender", TypeLine: "Creature — Human Wizard",
		OracleID: oldB101Archaeo, Power: 2, Toughness: 3})
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the enters trigger asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, bear) {
		t.Error("a creature card was offered")
	}
	effectstest.PickCard(t, g, me.ID, rock)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, rock) || effectstest.HandHasID(me, bear) {
		t.Error("the artifact card did not come back alone")
	}
}

func TestOldB101GravewakerReturnsACreatureTapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gw := effectstest.B12Push(g, me.ID, "Gravewaker", "Creature — Bird Spirit", oldB101Gravewaker, 5, 5)
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.FloatMana(t, g, me, "{B}{B}{B}{B}{B}{B}{B}")
	if err := g.ActivateCatalogAbility(me.ID, gw, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dead)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, dead)
	if !ok {
		t.Fatal("the creature did not return")
	}
	if !c.Tapped {
		t.Error("the creature returned untapped")
	}
}

func TestOldB101LightningPhoenixReturnsOnlyAfterThreeDamageToAnOpponent(t *testing.T) {
	for _, tc := range []struct {
		dmg  int
		back bool
	}{{2, false}, {3, true}} {
		g := effectstest.NewCatalogGame(t)
		seat := g.Turn.ActiveSeat
		me, opp := g.Seats[seat], g.Seats[(seat+1)%len(g.Seats)]
		ph := effectstest.PushCatalogGraveyardCard(me, "Lightning Phoenix", "Creature — Phoenix", oldB101Phoenix, 2, 2)
		src := effectstest.SeedCreature(g, "Bear", me.ID)
		effectstest.DealCombatDamageToPlayer(g, src, opp.ID, tc.dmg)
		effectstest.AdvanceToEndStepOf(t, g, seat)
		effectstest.FillPoolColored(me, "R", 1)
		effectstest.PassPriorityAroundTable(t, g)
		pay := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayUnless)
		if !tc.back {
			if pay != nil || g.Battlefield.Contains(ph) {
				t.Errorf("%d damage: the Phoenix offered a return", tc.dmg)
			}
			continue
		}
		if pay == nil {
			t.Fatalf("%d damage: no {R} offer", tc.dmg)
		}
		if err := g.ResolvePayUnless(pay.ID, me.ID, true); err != nil {
			t.Fatal(err)
		}
		if !g.Battlefield.Contains(ph) {
			t.Error("the Phoenix did not return")
		}
	}
}

func TestOldB101LightningPhoenixCannotBlock(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	ph := effectstest.B12Push(g, opp.ID, "Lightning Phoenix", "Creature — Phoenix", oldB101Phoenix, 2, 2)
	effectstest.AssertRestrictions(t, g, ph, game.CantBlock)
}

func TestOldB101DutifulAttendantReturnsAnotherCreatureCardWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	att := effectstest.B12Push(g, me.ID, "Dutiful Attendant", "Creature — Human Warrior", oldB101Attendant, 1, 2)
	dead := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.KillCreature(t, g, me.ID, att)
	g.RunStateChecksForTest()
	p := effectstest.LatestPickTarget(g, me.ID)
	if p == nil {
		t.Fatal("the dies trigger asked for no target")
	}
	if effectstest.HasID(p.PickTargetCards, att) {
		t.Error("the Attendant offered itself as 'another'")
	}
	effectstest.PickCard(t, g, me.ID, dead)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HandHasID(me, dead) {
		t.Error("the creature card did not return to hand")
	}
}

func TestOldB101ScarecroneSacrificesAScarecrowToDrawAndReanimatesAnArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sc := effectstest.B12Push(g, me.ID, "Scarecrone", "Artifact Creature — Scarecrow", oldB101Scarecrone, 1, 2)
	effectstest.SeedLibraryTop(me, "Top", "Basic Land — Forest", "")
	before := me.Hand.Size()
	effectstest.FloatMana(t, g, me, "{C}")
	effectstest.Pr7Activate(t, g, me.ID, sc, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{sc}})
	if g.Battlefield.Contains(sc) || me.Hand.Size() != before+1 {
		t.Errorf("sacrifice draw: on battlefield %v, hand %d (was %d)", g.Battlefield.Contains(sc), me.Hand.Size(), before)
	}
}

func TestOldB101ScarecroneReturnsAnArtifactCreatureCard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	sc := effectstest.B12Push(g, me.ID, "Scarecrone", "Artifact Creature — Scarecrow", oldB101Scarecrone, 1, 2)
	golem := effectstest.PushGraveyardCardTyped(me, "Dead Golem", oldB101ArtifactCrt)
	bear := effectstest.PushGraveyardCardTyped(me, "Dead Bear", "Creature — Bear")
	effectstest.FloatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, sc, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear), Strict: true}); err == nil {
		t.Fatal("a non-artifact creature card was a legal target")
	}
	if err := g.ActivateCatalogAbility(me.ID, sc, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(golem)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(golem) {
		t.Error("the artifact creature did not return")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, sc); !tapped {
		t.Error("Scarecrone did not tap for its cost")
	}
}

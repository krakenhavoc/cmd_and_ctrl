package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b821_test.go — Homelands, "Activated abilities", batch 821 (#5635).
// Every test plays the real catalog entry through the engine.

const (
	o821Gnomes    = "21d269f2-c11c-4094-b526-500f0785f483"
	o821Paladin   = "2fdbbca7-d1d0-4834-84df-2c2d9953f83a"
	o821Maze      = "44c65424-9c44-4bd8-9022-7ef5bb5b898e"
	o821Priestess = "630eab2b-c1fd-46bb-8ae0-ff837e46db1c"
	o821Kelp      = "64722c32-975b-43b9-9b4a-9511c75740d9"
	o821Walkers   = "68dfc626-66d5-4a40-9410-2b15676ebcec"
	o821Carriage  = "6bf5dcf1-5c93-42f4-adaf-a6f2e5807596"
	o821Tools     = "b1935140-b038-4a81-98f7-28d2c137124a"
	o821Didg      = "e9a2cb10-6526-4a5c-bda7-e2f328b377dc"
	o821Pony      = "f47f1c96-0b09-4277-8d61-cffa56d81be7"
	o821Troll     = "feabe30a-3c28-43a1-b538-b0207758520c"
)

func TestS82OldB821EveryCardIsRegisteredFull(t *testing.T) {
	for _, oracle := range []string{o821Gnomes, o821Paladin, o821Maze, o821Priestess, o821Kelp, o821Walkers, o821Carriage, o821Tools, o821Didg, o821Pony, o821Troll} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: completeness %v with %d caveats, want Full with none", spec.Name, spec.Completeness, len(spec.Caveats))
		}
	}
}

func o821Perm(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func TestS82OldB821ClockworkGnomesRegenerateOnlyAnArtifactCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	gnomes := o821Perm(g, me.ID, "Clockwork Gnomes", "Artifact Creature — Gnome", o821Gnomes, 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	golem := effectstest.B12Creature(g, me.ID, "Golem", "Artifact Creature — Golem", 3, 3)
	effectstest.FloatForTest(g, me, "CCC")
	err := g.ActivateCatalogAbility(me.ID, gnomes, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)})
	if err == nil {
		t.Fatal("the Gnomes targeted a nonartifact creature")
	}
	effectstest.ApaActivate(t, g, me, gnomes, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(golem)})
	if n := effectstest.RegenShieldsOn(t, g, golem); n != 1 {
		t.Errorf("%d regeneration shields on the Golem, want 1", n)
	}
	if c, _ := effectstest.BattlefieldCard(g, gnomes); !c.Tapped {
		t.Error("the Gnomes were not tapped as a cost")
	}
}

func TestS82OldB821SerraPaladinPreventsOneDamageAndGrantsVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	pal := o821Perm(g, me.ID, "Serra Paladin", "Creature — Human Knight", o821Paladin, 2, 2)
	src := effectstest.B12Creature(g, opp.ID, "Source", "Creature — Ogre", 3, 3)
	victim := effectstest.B12Creature(g, me.ID, "Victim", "Creature — Bear", 4, 4)
	effectstest.ApaActivate(t, g, me, pal, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(victim)})
	effectstest.Cmdr2B08Damage(g, src, victim, 3)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.DamageMarkedOn(g, victim); n != 2 {
		t.Errorf("%d damage marked, want 2 (3 less the 1 prevented)", n)
	}

	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g2)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	pal2 := o821Perm(g2, me2.ID, "Serra Paladin", "Creature — Human Knight", o821Paladin, 2, 2)
	bear := effectstest.B12Creature(g2, me2.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g2, me2, "CWW")
	effectstest.ApaActivate(t, g2, me2, pal2, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(bear)})
	if !effectstest.HasEffectiveKeyword(t, g2, bear, "vigilance") {
		t.Error("the target did not gain vigilance")
	}
}

func TestS82OldB821DarkMazeAttacksThenIsExiledAtTheEndStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, _ := effectstest.BlbB1Seats(g)
	maze := effectstest.B12Push(g, me.ID, "Dark Maze", "Creature — Wall", o821Maze, 4, 5)
	effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(maze, opp.ID); err == nil {
		t.Fatal("a defender attacked without the waiver")
	}

	g2 := effectstest.NewCatalogGame(t)
	me2, opp2, _ := effectstest.BlbB1Seats(g2)
	maze2 := effectstest.B12Push(g2, me2.ID, "Dark Maze", "Creature — Wall", o821Maze, 4, 5)
	effectstest.B16Activate(t, g2, me2.ID, maze2, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AdvanceTo(t, g2, game.StepDeclareAttackers)
	if err := g2.DeclareAttacker(maze2, opp2.ID); err != nil {
		t.Fatalf("the Maze could not attack after {0}: %v", err)
	}
	if !g2.Battlefield.Contains(maze2) {
		t.Fatal("the Maze left before the end step")
	}
	effectstest.AdvanceToEndStep(t, g2)
	effectstest.PassPriorityAroundTable(t, g2)
	if g2.Battlefield.Contains(maze2) || !g2.Exile.Contains(maze2) {
		t.Error("the Maze was not exiled at the beginning of the end step")
	}
}

func TestS82OldB821DidgeridooAndWillowPriestessPutOnlyTheNamedType(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	didg := o821Perm(g, me.ID, "Didgeridoo", "Artifact", o821Didg, 0, 0)
	mino := effectstest.HandCardForTest(me, "Minotaur Tester", "Creature — Minotaur", "")
	bear := effectstest.HandCardForTest(me, "Bear", "Creature — Bear", "")
	effectstest.FloatForTest(g, me, "CCC")
	if err := g.ActivateCatalogAbility(me.ID, didg, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("Didgeridoo did not offer a card")
	}
	for _, id := range pick.ChooseCards {
		if id == bear {
			t.Error("a non-Minotaur was offered")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, mino)
	if !g.Battlefield.Contains(mino) {
		t.Error("the Minotaur did not enter")
	}

	g2 := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g2)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	pri := o821Perm(g2, me2.ID, "Willow Priestess", "Creature — Faerie Druid", o821Priestess, 2, 2)
	faerie := effectstest.HandCardForTest(me2, "Faerie Tester", "Creature — Faerie", "")
	if err := g2.ActivateCatalogAbility(me2.ID, pri, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g2)
	effectstest.AnswerChooseCards(t, g2, me2.ID, faerie)
	if !g2.Battlefield.Contains(faerie) {
		t.Error("the Faerie did not enter")
	}
}

func TestS82OldB821WillowPriestessGivesAGreenCreatureProtectionFromBlack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pri := o821Perm(g, me.ID, "Willow Priestess", "Creature — Faerie Druid", o821Priestess, 2, 2)
	green := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Green Bear", TypeLine: "Creature — Bear", Colors: []string{"G"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	red := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Red Bear", TypeLine: "Creature — Bear", Colors: []string{"R"},
		Power: 2, Toughness: 2, Owner: me.ID, Controller: me.ID,
	})
	effectstest.FloatForTest(g, me, "CCG")
	if err := g.ActivateCatalogAbility(me.ID, pri, 1, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(red)}); err == nil {
		t.Fatal("a nongreen creature was targeted")
	}
	effectstest.ApaActivate(t, g, me, pri, 1, game.ActivateAbilityParams{Targets: effectstest.CardRefs(green)})
	if !effectstest.HasEffectiveKeyword(t, g, green, game.ProtectionFromColor("B")) {
		t.Error("the green creature did not gain protection from black")
	}
}

func TestS82OldB821WallOfKelpMakesAKelp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	wall := o821Perm(g, me.ID, "Wall of Kelp", "Creature — Plant Wall", o821Kelp, 0, 3)
	effectstest.FloatForTest(g, me, "UU")
	effectstest.ApaActivate(t, g, me, wall, 0, game.ActivateAbilityParams{})
	if n := effectstest.BattlefieldCountNamed(g, me.ID, "Kelp"); n != 1 {
		t.Fatalf("%d Kelp tokens, want 1", n)
	}
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Kelp" {
			if c.Power != 0 || c.Toughness != 1 || !game.HasKeyword(&c, "defender") {
				t.Errorf("Kelp is %d/%d, want a 0/1 with defender", c.Power, c.Toughness)
			}
		}
	}
}

func TestS82OldB821BeastWalkersGainBandingForG(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bw := o821Perm(g, me.ID, "Beast Walkers", "Creature — Human Beast Soldier", o821Walkers, 2, 2)
	if effectstest.HasEffectiveKeyword(t, g, bw, "banding") {
		t.Fatal("the Walkers have banding before the ability")
	}
	effectstest.FloatForTest(g, me, "G")
	effectstest.ApaActivate(t, g, me, bw, 0, game.ActivateAbilityParams{})
	if !effectstest.HasEffectiveKeyword(t, g, bw, "banding") {
		t.Error("the Walkers did not gain banding")
	}
}

func TestS82OldB821DwarvenPonyGrantsMountainwalkToADwarf(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	pony := o821Perm(g, me.ID, "Dwarven Pony", "Creature — Horse", o821Pony, 1, 1)
	dwarf := effectstest.B12Creature(g, me.ID, "Dwarf", "Creature — Dwarf", 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatForTest(g, me, "CR")
	if err := g.ActivateCatalogAbility(me.ID, pony, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(bear)}); err == nil {
		t.Fatal("a non-Dwarf was targeted")
	}
	effectstest.ApaActivate(t, g, me, pony, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(dwarf)})
	if !effectstest.HasEffectiveKeyword(t, g, dwarf, "mountainwalk") {
		t.Error("the Dwarf did not gain mountainwalk")
	}
}

func TestS82OldB821JovensToolsLimitBlockersToWalls(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Ogre", 3, 3)
	plain := effectstest.B12Creature(g, opp.ID, "Plain Blocker", "Creature — Bear", 2, 2)
	wall := effectstest.B12Creature(g, opp.ID, "Wall Blocker", "Creature — Wall", 0, 4)
	effectstest.AdvanceToMain(t, g)
	tools := o821Perm(g, me.ID, "Joven's Tools", "Artifact", o821Tools, 0, 0)
	effectstest.FloatForTest(g, me, "CCCC")
	effectstest.ApaActivate(t, g, me, tools, 0, game.ActivateAbilityParams{Targets: effectstest.CardRefs(atk)})
	if c, _ := effectstest.BattlefieldCard(g, tools); !c.Tapped {
		t.Error("the Tools were not tapped as a cost")
	}
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(plain, atk); err == nil {
		t.Error("a non-Wall blocked the creature")
	}
	if err := g.DeclareBlocker(wall, atk); err != nil {
		t.Errorf("a Wall could not block it: %v", err)
	}
}

func o821Colored(g *game.Game, owner uuid.UUID, name, typeLine, color string, p, tough int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, Colors: []string{color},
		Power: p, Toughness: tough, Owner: owner, Controller: owner,
	})
}

func TestS82OldB821SeaTrollRegeneratesOnlyAfterABlueCombat(t *testing.T) {
	for _, tc := range []struct {
		name  string
		color string
		want  bool
	}{{"blue", "U", true}, {"red", "R", false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			atk := o821Colored(g, me.ID, "Attacker", "Creature — Ogre", tc.color, 0, 5)
			troll := o821Perm(g, opp.ID, "Sea Troll", "Creature — Troll", o821Troll, 2, 1)
			effectstest.AdvanceToMain(t, g)
			effectstest.FloatForTest(g, opp, "U")
			if err := g.ActivateCatalogAbility(opp.ID, troll, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
				t.Fatal("the Troll regenerated before any block")
			}
			effectstest.DeclareAttack(t, g, opp.ID, atk)
			effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
			if err := g.DeclareBlocker(troll, atk); err != nil {
				t.Fatalf("block: %v", err)
			}
			effectstest.LockInBlocks(t, g)
			effectstest.FloatForTest(g, opp, "U")
			err := g.ActivateCatalogAbility(opp.ID, troll, 0, game.ActivateAbilityParams{Strict: true})
			if (err == nil) != tc.want {
				t.Fatalf("activation after blocking a %s creature: err=%v, want allowed=%v", tc.name, err, tc.want)
			}
			if tc.want {
				effectstest.PassPriorityAroundTable(t, g)
				if n := effectstest.RegenShieldsOn(t, g, troll); n != 1 {
					t.Errorf("%d regeneration shields, want 1", n)
				}
			}
		})
	}
}

func TestS82OldB821BlackCarriageUntapsOnlyInYourUpkeepForASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	car := o821Perm(g, me.ID, "Black Carriage", "Creature — Horse", o821Carriage, 4, 4)
	fodder := effectstest.B12Creature(g, me.ID, "Fodder", "Creature — Bear", 1, 1)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == car {
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	if err := g.ActivateCatalogAbility(me.ID, car, 0, game.ActivateAbilityParams{Strict: true, SacrificeIDs: []uuid.UUID{fodder}}); err == nil {
		t.Fatal("the Carriage untapped outside the upkeep")
	}
	seat := g.Turn.ActiveSeat
	effectstest.AdvanceToUpkeepOf(t, g, 1)
	effectstest.AdvanceToUpkeepOf(t, g, seat)
	if c, _ := effectstest.BattlefieldCard(g, car); !c.Tapped {
		t.Fatal("the Carriage untapped in the untap step")
	}
	effectstest.ApaActivate(t, g, g.Seats[seat], car, 0, game.ActivateAbilityParams{SacrificeIDs: []uuid.UUID{fodder}})
	if c, _ := effectstest.BattlefieldCard(g, car); c.Tapped {
		t.Error("the Carriage did not untap")
	}
	if g.Battlefield.Contains(fodder) {
		t.Error("the sacrificed creature is still on the battlefield")
	}
}

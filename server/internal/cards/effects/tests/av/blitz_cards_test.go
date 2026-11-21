package testsav

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Every blitz card offers "blitz" at its printed price, carrying the
// clause; the two graveyard cards also offer "blitz-graveyard", bound to
// the graveyard, and bind the hand offer to the hand.
func TestBlitzCardsOfferTheirPrintedBlitzCost(t *testing.T) {
	for _, tc := range []struct {
		name, oracle, cost string
		graveyard          bool
	}{
		{"Sabin, Master Monk", sabinOracle, "{2}{R}{R}", true},
		{"Tenacious Underdog", effectstest.TenaciousUnderdogOracle, "{2}{B}{B}", true},
		{"Riveteers Requisitioner", effectstest.RequisitionerOracle, "{2}{R}", false},
		{"Girder Goons", effectstest.GirderGoonsOracle, "{3}{B}", false},
		{"Workshop Warchief", effectstest.WorkshopWarchiefOracle, "{4}{G}{G}", false},
		{"Caldaia Strongarm", effectstest.CaldaiaStrongarmOracle, "{3}{G}", false},
		{"Mayhem Patrol", effectstest.MayhemPatrolOracle, "{1}{R}", false},
		{"Plasma Jockey", effectstest.PlasmaJockeyOracle, "{2}{R}", false},
		{"Night Clubber", effectstest.NightClubberOracle, "{2}{B}", false},
		{"Riveteers Decoy", effectstest.RiveteersDecoyOracle, "{3}{G}", false},
		{"Mezzio Mugger", effectstest.MezzioMuggerOracle, "{2}{R}", false},
		{"Pugnacious Pugilist", effectstest.PugnaciousPugilistOracle, "{3}{R}", false},
		{"Caldaia Guardian", effectstest.CaldaiaGuardianOracle, "{2}{G}", false},
		{"Wave of Rats", effectstest.WaveOfRatsOracle, "{4}{B}", false},
		{"Star Athlete", effectstest.StarAthleteOracle, "{3}{R}", false},
		{"Jaxis, the Troublemaker", effectstest.JaxisOracle, "{1}{R}", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := game.AlternativeCostByKey(tc.oracle, "blitz")
			if b == nil || !b.Blitz || b.ManaCost != tc.cost {
				t.Fatalf("blitz offer = %+v, want Blitz %s", b, tc.cost)
			}
			gy := game.AlternativeCostByKey(tc.oracle, "blitz-graveyard")
			if !tc.graveyard {
				if gy != nil || b.FromZone != "" {
					t.Fatalf("a hand-only blitz card offers %+v from %q", gy, b.FromZone)
				}
				return
			}
			if gy == nil || !gy.Blitz || gy.FromZone != game.ZoneGraveyard || gy.ManaCost != tc.cost || b.FromZone != game.ZoneHand {
				t.Fatalf("graveyard offer = %+v, hand offer bound to %q", gy, b.FromZone)
			}
		})
	}
}

// blitzCreatureFromHand casts a creature card from hand with its P/T set,
// claiming `alt` ("" for the printed cost) from a floated pool.
func blitzCreatureFromHand(t *testing.T, g *game.Game, me *game.Player, name, typeLine, manaCost, oracle string, power, toughness int, alt, pool string) uuid.UUID {
	t.Helper()
	effectstest.ToMainForCost(t, g)
	id := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: manaCost, OracleID: oracle,
			Owner: me.ID, Controller: me.ID, Power: power, Toughness: toughness,
		})
	})
	effectstest.FloatForTest(g, me, pool)
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, AlternativeCost: alt}); err != nil {
		t.Fatalf("cast %s (%q): %v", name, alt, err)
	}
	return id
}

// Tenacious Underdog from the graveyard: {2}{B}{B} and 2 life, haste,
// and back to the graveyard with a card drawn at the end step.
func TestTenaciousUnderdogBlitzesFromTheGraveyardForLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.ToMainForCost(t, g)
	id := effectstest.GraveyardCardOf(g, me, "Tenacious Underdog", "Creature — Human Warrior", "{1}{B}", effectstest.TenaciousUnderdogOracle)
	g.WithWriteLock(func() {
		for i := range me.Graveyard.Cards {
			if me.Graveyard.Cards[i].InstanceID == id {
				me.Graveyard.Cards[i].Power, me.Graveyard.Cards[i].Toughness = 3, 2
			}
		}
	})
	effectstest.FloatForTest(g, me, "BB")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard"}); err == nil {
		t.Fatal("the Underdog cannot be cast from the graveyard for {1}{B}")
	}
	effectstest.FloatForTest(g, me, "CC")
	life := me.Life
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "blitz-graveyard"}); err != nil {
		t.Fatalf("blitz from the graveyard: %v", err)
	}
	if me.Life != life-2 {
		t.Fatalf("blitz pays 2 life: %d want %d", me.Life, life-2)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.OnBattlefieldByID(g, id) || !effectstest.HasEffectiveKeyword(t, g, id, "haste") {
		t.Fatal("the Underdog enters with haste")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	hand := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(id) || len(me.Hand.Cards) != hand+1 {
		t.Errorf("sacrificed and drew (graveyard=%v hand %d want %d)", me.Graveyard.Contains(id), len(me.Hand.Cards), hand+1)
	}
}

// Riveteers Requisitioner blitzed: at the end step it dies, and both its
// own Treasure and the blitz draw happen.
func TestRiveteersRequisitionerBlitzedLeavesATreasureAndACard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := blitzCreatureFromHand(t, g, me, "Riveteers Requisitioner", "Creature — Lizard Rogue", "{1}{R}", effectstest.RequisitionerOracle, 3, 1, "blitz", "CCR")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEnd)
	hand := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, id) {
		t.Fatal("the blitzed Requisitioner is sacrificed")
	}
	// Two dies triggers of one controller at once: CR 603.3b ordering.
	if !effectstest.AnswerAnyTriggerOrderPrompt(t, g, me.ID) {
		t.Fatal("the two dies triggers are ordered by their controller")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DgTreasuresOf(g, me.ID) != 1 || len(me.Hand.Cards) != hand+1 {
		t.Errorf("Treasures %d want 1, hand %d want %d", effectstest.DgTreasuresOf(g, me.ID), len(me.Hand.Cards), hand+1)
	}
}

// Girder Goons: the Rogue enters tapped.
func TestGirderGoonsLeavesATappedRogue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B21Push(g, me.ID, "Girder Goons", "Creature — Ogre Warrior", effectstest.GirderGoonsOracle, 4, 4, "B")
	effectstest.B18Kill(t, g, id)
	var rogues, tapped int
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Rogue" && c.Controller == me.ID {
			rogues++
			if c.Tapped {
				tapped++
			}
		}
	}
	if rogues != 1 || tapped != 1 {
		t.Errorf("rogues %d tapped %d, want one tapped Rogue", rogues, tapped)
	}
}

// Workshop Warchief: 3 life on entry, a 4/4 Rhino Warrior on death.
func TestWorkshopWarchiefGainsLifeAndLeavesARhino(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	life := me.Life
	id := blitzCreatureFromHand(t, g, me, "Workshop Warchief", "Creature — Rhino Warrior", "{3}{G}{G}", effectstest.WorkshopWarchiefOracle, 5, 3, "", "CCCGG")
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life+3 {
		t.Fatalf("life %d want %d", me.Life, life+3)
	}
	effectstest.B18Kill(t, g, id)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Rhino Warrior" && c.Controller == me.ID && c.Power == 4 && c.Toughness == 4 {
			found = true
		}
	}
	if !found {
		t.Error("a 4/4 Rhino Warrior token")
	}
}

// Caldaia Strongarm: two +1/+1 counters on the chosen creature.
func TestCaldaiaStrongarmPutsTwoCounters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	bear := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	blitzCreatureFromHand(t, g, me, "Caldaia Strongarm", "Creature — Human Warrior", "{4}{G}", effectstest.CaldaiaStrongarmOracle, 2, 3, "blitz", "CCCG")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.CounterCount(g, bear, "+1/+1"); n != 2 {
		t.Errorf("+1/+1 counters %d want 2", n)
	}
}

// Mayhem Patrol: menace, and +1/+0 on the target when it attacks.
func TestMayhemPatrolPumpsTheTargetWhenItAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushDiesCreatureForTest(g, me.ID, "Mayhem Patrol", effectstest.MayhemPatrolOracle, "Creature — Devil Warrior", 1, 2)
	if !effectstest.HasEffectiveKeyword(t, g, id, "menace") {
		t.Fatal("Mayhem Patrol has menace")
	}
	effectstest.AttackWithTrigger(t, g, id, opp.ID, id)
	if p := effectstest.EffectivePower(t, g, id); p != 2 {
		t.Errorf("power %d want 2", p)
	}
}

// Plasma Jockey: the opponent's creature can't block.
func TestPlasmaJockeyStopsABlocker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushDiesCreatureForTest(g, me.ID, "Plasma Jockey", effectstest.PlasmaJockeyOracle, "Creature — Lizard Warrior", 3, 1)
	wall := effectstest.B16Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4, "W")
	effectstest.AttackWithTrigger(t, g, id, opp.ID, wall)
	var restricted bool
	g.ReadSnapshot(func() {
		if c, ok := g.LookupCardForEffect(wall); ok {
			restricted = c.Effective().Restrictions&game.CantBlock != 0
		}
	})
	if !restricted {
		t.Error("the targeted creature can't block this turn")
	}
}

// Night Clubber: opponents' creatures get -1/-1; yours do not.
func TestNightClubberShrinksOpponentsCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B16Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2, "G")
	elf := effectstest.B16Creature(g, opp.ID, "Their Elf", "Creature — Elf", 1, 1, "G")
	mine := effectstest.B16Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2, "G")
	blitzCreatureFromHand(t, g, me, "Night Clubber", "Creature — Human Warrior", "{1}{B}{B}", effectstest.NightClubberOracle, 2, 2, "", "CBB")
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.EffectivePower(t, g, theirs) != 1 || effectstest.EffectiveToughness(t, g, theirs) != 1 {
		t.Error("the opponent's Bear is 1/1")
	}
	if effectstest.OnBattlefieldByID(g, elf) {
		t.Error("the opponent's 1/1 dies")
	}
	if effectstest.EffectivePower(t, g, mine) != 2 {
		t.Error("your creatures are untouched")
	}
}

// Riveteers Decoy must be blocked if able.
func TestRiveteersDecoyMustBeBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := effectstest.Seats1684(g)
	decoy := effectstest.ReqCreature(g, me.ID, "Riveteers Decoy", effectstest.RiveteersDecoyOracle, 3, 1, "haste")
	effectstest.ReqCreature(g, opp.ID, "Spiked Wall", "", 3, 5)
	effectstest.ReqToBlockers(t, g, decoy)
	br := effectstest.ReqRefusal(t, g.PassPriority())
	if got := br.Sentence(opp.ID); got != "Riveteers Decoy must be blocked if able." {
		t.Errorf("sentence = %q", got)
	}
}

// Mezzio Mugger: each player's top card is exiled and playable this turn.
func TestMezzioMuggerExilesEachTopCardPlayable(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushDiesCreatureForTest(g, me.ID, "Mezzio Mugger", effectstest.MezzioMuggerOracle, "Creature — Lizard Rogue", 3, 3)
	tops := map[uuid.UUID]uuid.UUID{}
	for _, p := range g.Seats {
		if len(p.Library.Cards) > 0 {
			tops[p.ID] = p.Library.Cards[len(p.Library.Cards)-1].InstanceID
		}
	}
	effectstest.AttackWithTrigger(t, g, id, opp.ID, uuid.Nil)
	if len(tops) == 0 {
		t.Fatal("setup: libraries are empty")
	}
	for owner, card := range tops {
		var zone game.ZoneKind
		g.ReadSnapshot(func() {
			if z := g.FindCardZoneForEffect(card); z != nil {
				zone = z.Kind
			}
		})
		perm := effectstest.GrantedPermissionOn(g, me.ID, card, game.ZoneExile)
		if zone != game.ZoneExile || perm == nil || perm.CastOnly || !perm.AnyColor {
			t.Errorf("player %s's top card: zone %v permission %+v", owner, zone, perm)
		}
	}
}

// Pugnacious Pugilist: a tapped Devil attacking the same player.
func TestPugnaciousPugilistMakesAnAttackingDevil(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushDiesCreatureForTest(g, me.ID, "Pugnacious Pugilist", effectstest.PugnaciousPugilistOracle, "Creature — Ogre Warrior", 4, 4)
	effectstest.AttackWithTrigger(t, g, id, opp.ID, uuid.Nil)
	found := false
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Devil" && c.Controller == me.ID && c.Tapped && c.AttackingTarget == opp.ID {
			found = true
		}
	}
	if !found {
		t.Error("a tapped Devil attacking the Pugilist's defender")
	}
}

// Caldaia Guardian: itself, or another creature you control with mana
// value 4+, dying makes two Citizens; a 2-drop does not.
func TestCaldaiaGuardianCitizens(t *testing.T) {
	citizens := func(g *game.Game, owner uuid.UUID) int {
		n := 0
		for _, c := range g.Battlefield.Cards {
			if c.Name == "Citizen" && c.Controller == owner {
				n++
			}
		}
		return n
	}
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	guardian := effectstest.B21Push(g, me.ID, "Caldaia Guardian", "Creature — Human Soldier", effectstest.CaldaiaGuardianOracle, 4, 3, "G")
	small := effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2, "G")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == small {
				g.Battlefield.Cards[i].ManaCost = "{1}{G}"
			}
		}
	})
	big := effectstest.B16Creature(g, me.ID, "Big", "Creature — Giant", 5, 5, "G")
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == big {
				g.Battlefield.Cards[i].ManaCost = "{3}{G}{G}"
			}
		}
	})
	effectstest.B18Kill(t, g, small)
	if n := citizens(g, me.ID); n != 0 {
		t.Fatalf("a mana value 2 creature makes no Citizens, got %d", n)
	}
	effectstest.B18Kill(t, g, big)
	if n := citizens(g, me.ID); n != 2 {
		t.Fatalf("a mana value 5 creature makes two Citizens, got %d", n)
	}
	effectstest.B18Kill(t, g, guardian)
	if n := citizens(g, me.ID); n != 4 {
		t.Errorf("the Guardian itself makes two more, got %d", n)
	}
}

// Wave of Rats: returns only if it dealt combat damage to a player this
// turn.
func TestWaveOfRatsReturnsOnlyAfterConnecting(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	stay := effectstest.PushDiesCreatureForTest(g, me.ID, "Wave of Rats", effectstest.WaveOfRatsOracle, "Creature — Rat", 4, 2)
	effectstest.B18Kill(t, g, stay)
	if !me.Graveyard.Contains(stay) {
		t.Fatal("a Wave of Rats that dealt no combat damage stays dead")
	}

	g = effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	id := effectstest.PushDiesCreatureForTest(g, me.ID, "Wave of Rats", effectstest.WaveOfRatsOracle, "Creature — Rat", 4, 2)
	life := opp.Life
	effectstest.AttackWith(t, g, opp.ID, id)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Fatalf("setup: the Rats connected for 4, life %d want %d", opp.Life, life-4)
	}
	effectstest.B18Kill(t, g, id)
	if !effectstest.OnBattlefieldByID(g, id) {
		t.Error("a Wave of Rats that connected this turn returns")
	}
}

// Star Athlete: the target's controller sacrifices, or takes 5.
func TestStarAthletePunisher(t *testing.T) {
	for _, sacrifice := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		id := effectstest.PushDiesCreatureForTest(g, me.ID, "Star Athlete", effectstest.StarAthleteOracle, "Creature — Human Warrior", 3, 2)
		rock := effectstest.B21Push(g, opp.ID, "Rock", "Artifact", "", 0, 0)
		life := opp.Life
		effectstest.AdvanceTo(t, g, game.StepDeclareAttackers)
		if err := g.DeclareAttacker(id, g.Seats[2].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
		effectstest.LockInAttacks(t, g)
		effectstest.AnswerPickTarget(t, g, rock)
		effectstest.PassPriorityAroundTable(t, g)
		effectstest.AnswerMayChoice(t, g, opp.ID, sacrifice)
		effectstest.PassPriorityAroundTable(t, g)
		if sacrifice {
			if effectstest.OnBattlefieldByID(g, rock) || opp.Life != life {
				t.Errorf("sacrificed: rock on battlefield %v, life %d want %d", effectstest.OnBattlefieldByID(g, rock), opp.Life, life)
			}
		} else if !effectstest.OnBattlefieldByID(g, rock) || opp.Life != life-5 {
			t.Errorf("declined: rock on battlefield %v, life %d want %d", effectstest.OnBattlefieldByID(g, rock), opp.Life, life-5)
		}
	}
}

// Jaxis: the copy has haste and the draw, and is sacrificed at the end
// step, drawing a card.
func TestJaxisCopyHasBlitzRiders(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	jaxis := effectstest.PushDiesCreatureForTest(g, me.ID, "Jaxis, the Troublemaker", effectstest.JaxisOracle, "Legendary Creature — Human Warrior", 2, 3)
	bear := effectstest.PushDiesCreatureForTest(g, me.ID, "Bear", "", "Creature — Bear", 2, 2)
	effectstest.ToMainForCost(t, g)
	filler := effectstest.HandCardFull(me, "Filler", "Sorcery", "{1}", "", nil)
	effectstest.FloatForTest(g, me, "R")
	if err := g.ActivateCatalogAbility(me.ID, jaxis, 0, game.ActivateAbilityParams{
		Targets:    []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
		DiscardIDs: []uuid.UUID{filler},
	}); err != nil {
		t.Fatalf("activate Jaxis: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bear" && c.InstanceID != bear {
			token = c.InstanceID
		}
	}
	if token == uuid.Nil {
		t.Fatal("Jaxis makes a token copy of the Bear")
	}
	if !effectstest.HasEffectiveKeyword(t, g, token, "haste") || effectstest.GrantedTriggerCount(g, token) != 1 {
		t.Fatal("the copy has haste and \"When this token dies, draw a card.\"")
	}
	effectstest.AdvanceTo(t, g, game.StepEnd)
	hand := len(me.Hand.Cards)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.OnBattlefieldByID(g, token) || len(me.Hand.Cards) != hand+1 {
		t.Errorf("the copy is sacrificed and draws (on battlefield %v, hand %d want %d)", effectstest.OnBattlefieldByID(g, token), len(me.Hand.Cards), hand+1)
	}
	if !effectstest.OnBattlefieldByID(g, bear) {
		t.Error("the original stays")
	}
}

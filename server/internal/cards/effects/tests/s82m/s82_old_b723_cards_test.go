package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b723_cards_test.go — Urza's Saga damage, S82 OLD batch 723 (#5624).
// Every test plays the real catalog entry through the engine.

const (
	b723Wildfire        = "0da68ca2-07fd-4359-95a9-fb2f815e6ed0"
	b723Scald           = "1de0b522-f12a-4597-9d5f-9a246443487e"
	b723Retromancer     = "2c34949d-0765-455d-9ef4-81bc162b48d9"
	b723SulfuricVapors  = "40b14aaa-f29e-4e84-9bdc-bc2ac9762ce9"
	b723PurgingScythe   = "4410e3c4-b8c5-46d7-9418-7490bb69d8c4"
	b723SleeperAgent    = "627c9546-85b0-4b4d-93c6-e2ca8c3adf4f"
	b723Antagonism      = "68f81fcf-2dcf-41b9-ad7c-588149feb12a"
	b723PegasusCharger  = "81a5ac8d-b904-4755-a5bc-650e95f4138f"
	b723Somnophore      = "86541bba-7db8-4893-addf-8a4a42168ce4"
	b723SteamBlast      = "92f4bdfd-1c84-4541-bc46-2a65c39de717"
	b723OrderofYawgmoth = "c1795e9c-31c9-4892-a796-7656e09257b3"
	b723FireAnts        = "c23543b7-f637-4a7c-9b72-ec411cd5488e"
	b723FaultLine       = "ed743b88-0866-4f88-9f30-d49d2cf36b99"
	b723TerrorOracle    = "b81f041d-98db-4408-9472-c483e4a502bc"
)

func TestB723EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		b723Wildfire:        "Wildfire",
		b723Scald:           "Scald",
		b723Retromancer:     "Retromancer",
		b723SulfuricVapors:  "Sulfuric Vapors",
		b723PurgingScythe:   "Purging Scythe",
		b723SleeperAgent:    "Sleeper Agent",
		b723Antagonism:      "Antagonism",
		b723PegasusCharger:  "Pegasus Charger",
		b723Somnophore:      "Somnophore",
		b723SteamBlast:      "Steam Blast",
		b723OrderofYawgmoth: "Order of Yawgmoth",
		b723FireAnts:        "Fire Ants",
		b723FaultLine:       "Fault Line",
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

func b723Lands(g *game.Game, id uuid.UUID) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == id && c.IsLand() {
			n++
		}
	}
	return n
}

func TestB723WildfireTakesFourLandsFromEachPlayerThenDamagesEveryCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	var mine []uuid.UUID
	for i := 0; i < 6; i++ {
		mine = append(mine, effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"}))
	}
	effectstest.PushLands(g, opp.ID, 4)
	big := effectstest.B12Creature(g, me.ID, "Big Bear", "Creature — Bear", 5, 5)
	small := effectstest.B12Creature(g, opp.ID, "Small Bear", "Creature — Bear", 4, 4)
	effectstest.CastWithCost(t, g, "Wildfire", "Sorcery", "{4}{R}{R}", b723Wildfire)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSacrifice(t, g, me.ID, mine[:4]...)
	effectstest.PassPriorityAroundTable(t, g)
	if b723Lands(g, me.ID) != 2 || b723Lands(g, opp.ID) != 0 {
		t.Errorf("lands left: me %d (want 2), opponent %d (want 0)", b723Lands(g, me.ID), b723Lands(g, opp.ID))
	}
	if _, ok := effectstest.BattlefieldCard(g, small); ok {
		t.Error("the 4/4 survived 4 damage")
	}
	if got := effectstest.DamageMarkedOn(g, big); got != 4 {
		t.Errorf("the 5/5 has %d damage, want 4", got)
	}
}

func TestB723WildfireDamagesEvenWhenNobodyHasLands(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.CastWithCost(t, g, "Wildfire", "Sorcery", "{4}{R}{R}", b723Wildfire)
	effectstest.PassPriorityAroundTable(t, g)
	if _, ok := effectstest.BattlefieldCard(g, bear); ok {
		t.Error("the Bear survived with nobody having a land to sacrifice")
	}
}

func TestB723ScaldPingsWhoeverTapsAnIslandOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Scald", "Enchantment", b723Scald, 0, 0)
	forest := effectstest.SeedLandOnBattlefield(g, me.ID, "Forest", "Basic Land — Forest")
	mine := effectstest.SeedLandOnBattlefield(g, me.ID, "Island", "Basic Land — Island")
	theirs := effectstest.SeedLandOnBattlefield(g, opp.ID, "Island", "Basic Land — Island")
	effectstest.AdvanceToMain(t, g)
	myLife, oppLife := me.Life, opp.Life
	if err := g.ActivateManaAbility(me.ID, forest, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap the Forest: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife {
		t.Errorf("tapping a Forest cost %d life", myLife-me.Life)
	}
	if err := g.ActivateManaAbility(me.ID, mine, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap my Island: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != myLife-1 {
		t.Errorf("my life %d → %d, want -1: the controller is a player too", myLife, me.Life)
	}
	if err := g.ActivateManaAbility(opp.ID, theirs, 0, game.ManaAbilityParams{}); err != nil {
		t.Fatalf("tap their Island: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-1 || me.Life != myLife-1 {
		t.Errorf("life me %d opp %d, want the tapper (the opponent) to lose 1 and me to stay", me.Life, opp.Life)
	}
}

func TestB723RetromancerDealsThreeToTheControllerOfWhateverTargetsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	retro := effectstest.B12Push(g, opp.ID, "Retromancer", "Creature — Lizard Shaman", b723Retromancer, 3, 3)
	life := me.Life
	effectstest.CastCatalogSpell(t, g, "Terror", "Instant", b723TerrorOracle, effectstest.CardTarget(retro))
	effectstest.PassPriorityAroundTable(t, g)
	if me.Life != life-3 {
		t.Errorf("the caster's life %d → %d, want 3 damage from the Retromancer even as it died", life, me.Life)
	}
}

func TestB723SulfuricVapors(t *testing.T) {
	for _, red := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		effectstest.B12Push(g, me.ID, "Sulfuric Vapors", "Enchantment", b723SulfuricVapors, 0, 0)
		big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 6, 6)
		life := opp.Life
		cost := "{2}{R}"
		if !red {
			cost = "{3}"
		}
		effectstest.CastWithCost(t, g, "Steam Blast", "Sorcery", cost, b723SteamBlast)
		effectstest.PassPriorityAroundTable(t, g)
		want := 2
		if red {
			want = 3
		}
		if got := effectstest.DamageMarkedOn(g, big); got != want {
			t.Errorf("red=%v: the bear took %d, want %d", red, got, want)
		}
		if opp.Life != life-want {
			t.Errorf("red=%v: the opponent lost %d, want %d", red, life-opp.Life, want)
		}
	}
}

func TestB723SulfuricVaporsDoesNotBoostARedPermanent(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.B12Push(g, me.ID, "Sulfuric Vapors", "Enchantment", b723SulfuricVapors, 0, 0)
	ants := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Fire Ants", OracleID: b723FireAnts, TypeLine: "Creature — Insect",
		ManaCost: "{2}{R}", Colors: []string{"R"}, Power: 2, Toughness: 1, Owner: me.ID, Controller: me.ID})
	big := effectstest.B12Creature(g, opp.ID, "Big Bear", "Creature — Bear", 6, 6)
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, ants).SummonedThisTurn = false })
	if err := g.ActivateCatalogAbility(me.ID, ants, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate Fire Ants: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, big); got != 1 {
		t.Errorf("a red permanent's damage was boosted: the bear took %d, want 1", got)
	}
}

func b723AdvanceToUpkeepOf(t *testing.T, g *game.Game, seat int) {
	t.Helper()
	start := g.Turn.Seq
	for g.Turn.ActiveSeat != seat || g.Turn.Step != game.StepUpkeep || g.Turn.Seq <= start {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func TestB723PurgingScytheHitsTheLeastToughnessCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Purging Scythe", "Artifact", b723PurgingScythe, false)
	tough := effectstest.B12Creature(g, me.ID, "Tough", "Creature — Bear", 1, 4)
	frail := effectstest.B12Creature(g, opp.ID, "Frail", "Creature — Bear", 1, 3)
	b723AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, frail); got != 2 {
		t.Errorf("the least-toughness creature took %d, want 2", got)
	}
	if got := effectstest.DamageMarkedOn(g, tough); got != 0 {
		t.Errorf("the other creature took %d, want 0", got)
	}
}

func TestB723PurgingScytheLetsYouChooseAmongATie(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Purging Scythe", "Artifact", b723PurgingScythe, false)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 1, 3)
	effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 1, 5)
	b723AdvanceToUpkeepOf(t, g, 0)
	effectstest.PassPriorityAroundTable(t, g)
	c := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if c == nil || len(c.ChooseCards) != 2 {
		t.Fatalf("the tie between two creatures was not offered to the controller: %+v", c)
	}
	effectstest.AnswerChooseCards(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.DamageMarkedOn(g, theirs) != 2 || effectstest.DamageMarkedOn(g, mine) != 0 {
		t.Errorf("damage: theirs %d, mine %d; want 2 and 0", effectstest.DamageMarkedOn(g, theirs), effectstest.DamageMarkedOn(g, mine))
	}
}

func TestB723SleeperAgentChangesHandsThenHurtsItsNewController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	agent := effectstest.CastWithCost(t, g, "Sleeper Agent", "Creature — Phyrexian Minion", "{B}", b723SleeperAgent)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayerIfOpen(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ApaLive(g, agent).Controller; got != opp.ID {
		t.Fatalf("the Agent is controlled by %s, want the targeted opponent", got)
	}
	myLife, oppLife := me.Life, opp.Life
	b723AdvanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-2 {
		t.Errorf("the new controller's life %d → %d, want 2 damage", oppLife, opp.Life)
	}
	if me.Life != myLife {
		t.Errorf("the old controller lost %d life", myLife-me.Life)
	}
}

func TestB723AntagonismHurtsThePlayerUnlessAnOpponentWasDamaged(t *testing.T) {
	for _, damaged := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		effectstest.B12Push(g, opp.ID, "Antagonism", "Enchantment", b723Antagonism, 0, 0)
		src := effectstest.B12Creature(g, me.ID, "Pinger", "Creature — Bear", 1, 1)
		effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepPostcombatMain)
		if damaged {
			g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(src, opp.ID, 1) })
		}
		life := me.Life
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		want := life - 2
		if damaged {
			want = life
		}
		if me.Life != want {
			t.Errorf("damaged=%v: my life %d → %d, want %d (Antagonism belongs to my opponent but hits each player)", damaged, life, me.Life, want)
		}
	}
}

func TestB723OrderOfYawgmothMakesThePlayerDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	order := effectstest.B12Push(g, me.ID, "Order of Yawgmoth", "Creature — Phyrexian Zombie Knight", b723OrderofYawgmoth, 2, 2)
	effectstest.GiveHandCards(opp, 3)
	hand := opp.Hand.Size()
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(order, opp.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, opp.Hand.Cards[0].InstanceID)
	if opp.Hand.Size() != hand-1 {
		t.Errorf("hand %d → %d, want one discard", hand, opp.Hand.Size())
	}
	if !effectstest.HasEffectiveKeyword(t, g, order, "fear") {
		t.Error("Order of Yawgmoth lacks fear")
	}
}

func TestB723SomnophoreTapsAndHoldsOnlyWhileItRemains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	som := effectstest.B12Push(g, me.ID, "Somnophore", "Creature — Illusion", b723Somnophore, 2, 2)
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	victim := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { _ = g.DealDamageToPlayerForEffect(som, opp.ID, 2) })
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.ChooseCardsChoiceFor(g, me.ID) != nil || effectstest.LatestPickTarget(g, me.ID) != nil {
		effectstest.AnswerPickTarget(t, g, victim)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, victim); !tapped {
		t.Fatal("the targeted creature was not tapped")
	}
	if tapped, _ := effectstest.BattlefieldCardTapped(g, mine); tapped {
		t.Error("a creature the damaged player does not control was tapped")
	}
	effectstest.Batch01AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, victim); !tapped {
		t.Error("the creature untapped while Somnophore remained")
	}
	effectstest.KillCreature(t, g, me.ID, som)
	effectstest.Batch01AdvanceToStepOf(t, g, 0, game.StepUpkeep)
	effectstest.Batch01AdvanceToStepOf(t, g, 1, game.StepUpkeep)
	if tapped, _ := effectstest.BattlefieldCardTapped(g, victim); tapped {
		t.Error("the creature stayed tapped after Somnophore left")
	}
}

func TestB723SteamBlastHitsEveryCreatureAndPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"}, Power: 1, Toughness: 4, Owner: me.ID, Controller: me.ID})
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	effectstest.CastWithCost(t, g, "Steam Blast", "Sorcery", "{2}{R}", b723SteamBlast)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, flyer); got != 2 {
		t.Errorf("the flyer took %d, want 2 (Steam Blast spares nothing)", got)
	}
	for _, p := range g.Seats {
		if p.Life != lives[p.ID]-2 {
			t.Errorf("a player's life %d → %d, want -2", lives[p.ID], p.Life)
		}
	}
}

func TestB723FireAntsPingEveryOtherCreatureWithoutFlying(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	ants := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Fire Ants", OracleID: b723FireAnts, TypeLine: "Creature — Insect",
		Power: 2, Toughness: 4, Owner: me.ID, Controller: me.ID})
	mine := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 3)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 3)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"}, Power: 1, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	effectstest.AdvanceToMain(t, g)
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, ants).SummonedThisTurn = false })
	if err := g.ActivateCatalogAbility(me.ID, ants, 0, game.ActivateAbilityParams{Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	for name, c := range map[string]struct {
		id   uuid.UUID
		want int
	}{"mine": {mine, 1}, "theirs": {theirs, 1}, "flyer": {flyer, 0}, "the Ants": {ants, 0}} {
		if got := effectstest.DamageMarkedOn(g, c.id); got != c.want {
			t.Errorf("%s took %d, want %d", name, got, c.want)
		}
	}
	if err := g.ActivateCatalogAbility(me.ID, ants, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Error("activated while tapped")
	}
}

func TestB723FaultLineDamagesByXAndSparesFlyers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 5)
	flyer := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Bird", TypeLine: "Creature — Bird", Keywords: []string{"flying"}, Power: 1, Toughness: 5, Owner: me.ID, Controller: me.ID})
	lives := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		lives[p.ID] = p.Life
	}
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: "Fault Line", TypeLine: "Instant", OracleID: b723FaultLine, ManaCost: "{X}{R}{R}", Colors: []string{"R"}, Owner: me.ID, Controller: me.ID})
	effectstest.FloatForTest(g, me, "{R}{R}{C}{C}{C}")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{XValue: 3, Strict: true}); err != nil {
		t.Fatalf("cast Fault Line for X=3: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.DamageMarkedOn(g, bear); got != 3 {
		t.Errorf("the bear took %d, want 3", got)
	}
	if got := effectstest.DamageMarkedOn(g, flyer); got != 0 {
		t.Errorf("the flyer took %d, want 0", got)
	}
	for _, p := range g.Seats {
		if p.Life != lives[p.ID]-3 {
			t.Errorf("a player's life %d → %d, want -3", lives[p.ID], p.Life)
		}
	}
}

func TestB723PegasusChargerHasFlyingAndFirstStrike(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	id := effectstest.B12Push(g, me.ID, "Pegasus Charger", "Creature — Pegasus", b723PegasusCharger, 2, 1)
	for _, kw := range []string{"flying", "first strike"} {
		if !effectstest.HasEffectiveKeyword(t, g, id, kw) {
			t.Errorf("Pegasus Charger lacks %s", kw)
		}
	}
}

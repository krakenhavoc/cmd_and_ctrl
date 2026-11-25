package testsu

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// znr_old_b091_test.go — Zendikar Rising damage cards, S82 OLD batch 91
// (#5508). Every test plays the real catalog entry through the engine.

const (
	oldB091AngelOracle      = "02f46271-c90d-4318-9004-74e7bee64a89"
	oldB091StingerOracle    = "214562b1-fdf0-4d5d-93f0-594b31b80bb3"
	oldB091BarrageOracle    = "26a55f91-91a1-4a1e-86fe-6551fd64ea87"
	oldB091ChargerOracle    = "4d83ad77-dbac-4a03-8034-43d950d0a18c"
	oldB091TacticsOracle    = "612c9079-adae-4fe5-9a72-8efc09ddfe5c"
	oldB091ZagrasOracle     = "a2375a23-a1f0-4514-87a0-e699f91acd7d"
	oldB091ZarethOracle     = "a316b0ce-1f63-4a8e-9a6e-bfbe03ca4e9e"
	oldB091HellionOracle    = "b2151d51-39f6-49cc-a5cc-8202dd028336"
	oldB091SpellcraftOracle = "bfdb13b9-7050-43a3-ac5f-d427942b1616"
	oldB091GuideBeastOracle = "c752e001-56cc-46dc-acb3-3fca6c0bdd2e"
	oldB091SorcererOracle   = "d8b7b78e-91b2-4ad5-aeb8-6ebdca363e71"
	oldB091BearType         = "Creature — Bear"
	oldB091ClericType       = "Creature — Human Cleric"
	oldB091RogueType        = "Creature — Human Rogue"
	oldB091WarriorType      = "Creature — Human Warrior"
	oldB091WizardType       = "Creature — Human Wizard"
	oldB091ForestType       = "Basic Land — Forest"
)

func TestOldB091EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB091AngelOracle: "Angel of Destiny", oldB091StingerOracle: "Shadow Stinger",
		oldB091BarrageOracle: "Sizzling Barrage", oldB091ChargerOracle: "Fireblade Charger",
		oldB091TacticsOracle: "Practiced Tactics", oldB091ZagrasOracle: "Zagras, Thief of Heartbeats",
		oldB091ZarethOracle: "Zareth San, the Trickster", oldB091HellionOracle: "Pyroclastic Hellion",
		oldB091SpellcraftOracle: "Synchronized Spellcraft", oldB091GuideBeastOracle: "Wayward Guide-Beast",
		oldB091SorcererOracle: "Rockslide Sorcerer",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok {
			t.Errorf("%s is not in the catalog", name)
			continue
		}
		if spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %v, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// oldB091CastFree casts a card from the active player's hand for a {0} cost.
func oldB091CastFree(t *testing.T, g *game.Game, name, typeLine, oracle string, params game.CastSpellParams) uuid.UUID {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	id := effectstest.MayhemHandCard(g, me, name, typeLine, "{0}", oracle)
	if err := g.CastSpell(me.ID, id, params); err != nil {
		t.Fatalf("CastSpell %s: %v", name, err)
	}
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	return id
}

func oldB091Party(g *game.Game, owner uuid.UUID, types ...string) {
	for i, ty := range types {
		effectstest.B12Creature(g, owner, "Party member", ty, 1, 1+i)
	}
}

func TestOldB091AngelOfDestinyHasKeywordsAndSharesLifeFromCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	angel := effectstest.B12Push(g, me.ID, "Angel of Destiny", "Creature — Angel Cleric", oldB091AngelOracle, 2, 6)
	for _, kw := range []string{"flying", "double strike"} {
		if !effectstest.HasEffectiveKeyword(t, g, angel, kw) {
			t.Errorf("no %s", kw)
		}
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB091BearType, 3, 3)
	lifeMe, lifeOpp := me.Life, opp.Life
	effectstest.AttackWith(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != lifeOpp-3+3 || me.Life != lifeMe+3 {
		t.Errorf("life me %d opp %d, want %d / %d (the Bear's 3 damage, then each gains 3)", me.Life, opp.Life, lifeMe+3, lifeOpp)
	}
}

func TestOldB091AngelOfDestinyMakesTheAttackedPlayersLoseWithFifteenOverStartingLife(t *testing.T) {
	for _, enough := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		angel := effectstest.B12Push(g, me.ID, "Angel of Destiny", "Creature — Angel Cleric", oldB091AngelOracle, 2, 6)
		// The Angel's own double strike gains 2 + 2 life before the end step.
		me.Life = effects.StartingLifeOf(g) + 10
		if enough {
			me.Life = effects.StartingLifeOf(g) + 11
		}
		effectstest.AttackWith(t, g, opp.ID, angel)
		effectstest.AdvanceTo(t, g, game.StepEnd)
		effectstest.PassPriorityAroundTable(t, g)
		if lost := opp.Eliminated; lost != enough {
			t.Errorf("enough=%v: the attacked player eliminated = %v", enough, lost)
		}
	}
}

func TestOldB091ShadowStingerMillsAndGrantsDeathtouchByTappingARogue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	effectstest.Msc18Library(opp, 10)
	stinger := effectstest.B12Push(g, me.ID, "Shadow Stinger", "Creature — Vampire Rogue", oldB091StingerOracle, 1, 4)
	if err := g.ActivateCatalogAbility(me.ID, stinger, 0, game.ActivateAbilityParams{Strict: true}); err == nil {
		t.Fatal("activated with no other Rogue to tap")
	}
	rogue := effectstest.B12Creature(g, me.ID, "Rogue", oldB091RogueType, 1, 1)
	if err := g.ActivateCatalogAbility(me.ID, stinger, 0, game.ActivateAbilityParams{Strict: true, TapIDs: []uuid.UUID{rogue}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, stinger, "deathtouch") {
		t.Error("no deathtouch")
	}
	library := opp.Library.Size()
	effectstest.AttackWith(t, g, opp.ID, stinger)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Library.Size() != library-3 {
		t.Errorf("library %d -> %d, want 3 milled", library, opp.Library.Size())
	}
}

func TestOldB091SizzlingBarrageOnlyHitsACreatureThatBlocked(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB091BearType, 1, 1)
	blocker := effectstest.B12Creature(g, opp.ID, "Blocker", oldB091BearType, 0, 4)
	idle := effectstest.B12Creature(g, opp.ID, "Idle", oldB091BearType, 0, 4)
	effectstest.BrAttack(t, g, attacker)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("block: %v", err)
	}
	effectstest.FinishBlockDeclarations(t, g)
	cast := func(target uuid.UUID) error {
		id := effectstest.MayhemHandCard(g, me, "Sizzling Barrage", "Instant", "{0}", oldB091BarrageOracle)
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(target)})
		if err != nil {
			_, _ = me.Hand.Remove(id)
		}
		return err
	}
	if err := cast(idle); err == nil {
		t.Fatal("targeted a creature that did not block")
	}
	if err := cast(blocker); err != nil {
		t.Fatalf("a creature that blocked is a legal target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(blocker) {
		t.Error("4 damage did not kill the 0/4 blocker")
	}
}

func TestOldB091FirebladeChargerDealsItsPowerToAnyTargetWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	charger := effectstest.B12Push(g, me.ID, "Fireblade Charger", "Creature — Goblin Warrior", oldB091ChargerOracle, 1, 1)
	effectstest.PutCounters(t, g, charger, game.CounterPlusOne, 2)
	life := opp.Life
	effectstest.SbondDie(t, g, me.ID, charger)
	g.RunStateChecksForTest()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d (its power was 3 as it last existed)", opp.Life, life-3)
	}
}

func TestOldB091PracticedTacticsDealsTwicePartySizeToAnAttackingOrBlockingCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	oldB091Party(g, me.ID, oldB091ClericType, oldB091RogueType)
	attacker := effectstest.B12Creature(g, me.ID, "Attacker", oldB091BearType, 1, 4)
	idle := effectstest.B12Creature(g, me.ID, "Idle", oldB091BearType, 1, 4)
	effectstest.BrAttack(t, g, attacker)
	cast := func(target uuid.UUID) error {
		id := effectstest.MayhemHandCard(g, me, "Practiced Tactics", "Instant", "{0}", oldB091TacticsOracle)
		err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, Targets: effectstest.CardTarget(target)})
		if err != nil {
			_, _ = me.Hand.Remove(id)
		}
		return err
	}
	if err := cast(idle); err == nil {
		t.Fatal("targeted a creature that is neither attacking nor blocking")
	}
	if err := cast(attacker); err != nil {
		t.Fatalf("an attacking creature is a legal target: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(attacker) {
		t.Error("twice a party of two is 4 damage, which kills the 1/4")
	}
}

func TestOldB091SynchronizedSpellcraftHitsTheCreatureAndItsControllerForPartySize(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB091Party(g, me.ID, oldB091ClericType, oldB091RogueType, oldB091WizardType)
	ogre := effectstest.B12Creature(g, opp.ID, "Ogre", "Creature — Ogre", 3, 4)
	life := opp.Life
	oldB091CastFree(t, g, "Synchronized Spellcraft", "Instant", oldB091SpellcraftOracle, game.CastSpellParams{Targets: effectstest.CardTarget(ogre)})
	if g.Battlefield.Contains(ogre) {
		t.Error("4 damage did not kill the 3/4")
	}
	if opp.Life != life-3 {
		t.Errorf("life %d, want %d (a party of three)", opp.Life, life-3)
	}
}

func TestOldB091ZagrasCostsLessGivesDeathtouchAndKillsDamagedPlaneswalkers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	oldB091Party(g, me.ID, oldB091ClericType, oldB091WizardType)
	zagras := effectstest.HandCardFull(me, "Zagras, Thief of Heartbeats", "Legendary Creature — Vampire Rogue", "{4}{B}{R}", oldB091ZagrasOracle, []string{"B", "R"})
	// A party of two: {4}{B}{R} costs {2}{B}{R}.
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{B}{R}{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, zagras, game.CastSpellParams{Strict: true}); err == nil {
		t.Fatal("cast for three mana with a party of two")
	}
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{G}"); err != nil {
		t.Fatal(err)
	}
	if err := g.CastSpell(me.ID, zagras, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast for {2}{B}{R}: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	bear := effectstest.B12Creature(g, me.ID, "Bear", oldB091BearType, 2, 2)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "deathtouch") {
		t.Error("another creature you control has no deathtouch")
	}
	if effectstest.HasEffectiveKeyword(t, g, effectstest.B12Creature(g, opp.ID, "Theirs", oldB091BearType, 1, 1), "deathtouch") {
		t.Error("an opponent's creature has deathtouch")
	}
	walker := effectstest.P1g2Planeswalker(g, opp.ID, 7)
	effectstest.AdvanceToDeclareAttackersOf(t, g, g.Turn.ActiveSeat)
	if err := g.DeclareAttacker(bear, walker); err != nil {
		t.Fatalf("attack the planeswalker: %v", err)
	}
	effectstest.LockInAttacks(t, g)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(walker) {
		t.Error("the damaged planeswalker survived")
	}
}

func TestOldB091ZarethSanEntersAttackingForAnUnblockedRogue(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	rogue := effectstest.B12Creature(g, me.ID, "Sneaky Rogue", oldB091RogueType, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Sneaky Bear", oldB091BearType, 1, 1)
	zareth := effectstest.PushNinjaToHand(me, "Zareth San, the Trickster", "Legendary Creature — Merfolk Rogue", oldB091ZarethOracle, 4, 4)
	effectstest.DeclareAttack(t, g, opp.ID, rogue, bear)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.AddManaForEffect(me.ID, uuid.Nil, "{U}{U}{B}{B}"); err != nil {
		t.Fatal(err)
	}
	if err := g.ActivateCatalogAbility(me.ID, zareth, 0, game.ActivateAbilityParams{ReturnIDs: []uuid.UUID{bear}}); err == nil {
		t.Fatal("returned an attacking Bear for a Rogue's cost")
	}
	if err := g.ActivateCatalogAbility(me.ID, zareth, 0, game.ActivateAbilityParams{ReturnIDs: []uuid.UUID{rogue}}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Hand.Contains(rogue) {
		t.Error("the Rogue was not returned to hand")
	}
	c, ok := g.LookupCardForEffect(zareth)
	if !ok || !g.Battlefield.Contains(zareth) {
		t.Fatal("Zareth San did not enter")
	}
	if !c.Tapped || c.AttackingTarget != opp.ID {
		t.Errorf("tapped=%v attacking=%v, want tapped and attacking the same player", c.Tapped, c.AttackingTarget)
	}
}

func TestOldB091ZarethSanReturnsAPermanentFromTheDamagedPlayersGraveyard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	zareth := effectstest.B12Push(g, me.ID, "Zareth San, the Trickster", "Legendary Creature — Merfolk Rogue", oldB091ZarethOracle, 4, 4)
	dead := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: dead, Name: "Their Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3, Owner: opp.ID, Controller: opp.ID})
	spell := uuid.New()
	opp.Graveyard.PushTop(game.Card{InstanceID: spell, Name: "Their Bolt", TypeLine: "Instant", Owner: opp.ID, Controller: opp.ID})
	effectstest.AttackWith(t, g, opp.ID, zareth)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTarget(t, g, dead)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerConfirmFor(t, g, me, true)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := g.LookupCardForEffect(dead)
	if !ok || !g.Battlefield.Contains(dead) || c.Controller != me.ID {
		t.Error("the card did not come back under the Trickster's controller")
	}
	if !opp.Graveyard.Contains(spell) {
		t.Error("the untargeted card left the graveyard")
	}
}

func TestOldB091PyroclasticHellionReturnsALandThenDamagesEachOpponent(t *testing.T) {
	for _, accept := range []bool{false, true} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		land := effectstest.B12Permanent(g, me.ID, "Forest", oldB091ForestType)
		effectstest.CastHoldCreature(t, g, "Pyroclastic Hellion", "Creature — Hellion", oldB091HellionOracle, 4, 5)
		life := map[uuid.UUID]int{}
		for _, p := range g.Seats {
			life[p.ID] = p.Life
		}
		if accept {
			effectstest.AnswerOwnPermanents(t, g, me.ID, land)
		} else {
			effectstest.AnswerOwnPermanents(t, g, me.ID)
		}
		g.SettleResolution()
		for _, p := range g.Seats {
			want := life[p.ID]
			if accept && p.ID != me.ID {
				want -= 2
			}
			if p.Life != want {
				t.Errorf("accept=%v: %s life %d, want %d", accept, p.Name, p.Life, want)
			}
		}
		if got := me.Hand.Contains(land); got != accept {
			t.Errorf("accept=%v: land in hand = %v", accept, got)
		}
	}
}

func TestOldB091WaywardGuideBeastReturnsALandOnCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	effectstest.AdvanceToMain(t, g)
	beast := effectstest.B12Push(g, me.ID, "Wayward Guide-Beast", "Creature — Beast", oldB091GuideBeastOracle, 2, 2)
	land := effectstest.B12Permanent(g, me.ID, "Forest", oldB091ForestType)
	effectstest.AttackWith(t, g, opp.ID, beast)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID) != nil {
		effectstest.AnswerOwnPermanents(t, g, me.ID, land)
		g.SettleResolution()
	}
	if !me.Hand.Contains(land) {
		t.Error("the land did not return to hand")
	}
}

func TestOldB091RockslideSorcererPingsOnInstantSorceryAndWizardSpellsOnly(t *testing.T) {
	for _, c := range []struct {
		name, typeLine string
		want           bool
	}{
		{"Quick Instant", "Instant", true},
		{"Slow Sorcery", "Sorcery", true},
		{"Other Wizard", oldB091WizardType, true},
		{"Plain Bear", oldB091BearType, false},
	} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
		effectstest.B12Push(g, me.ID, "Rockslide Sorcerer", "Creature — Human Wizard Sorcerer", oldB091SorcererOracle, 3, 3)
		life := opp.Life
		id := effectstest.MayhemHandCard(g, me, c.name, c.typeLine, "{0}", "")
		effectstest.AdvanceToMain(t, g)
		if err := g.CastSpell(me.ID, id, game.CastSpellParams{}); err != nil {
			t.Fatalf("cast %s: %v", c.name, err)
		}
		g.RunStateChecksForTest()
		if c.want {
			effectstest.AnswerPickTargetPlayer(t, g, opp.ID)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if got := opp.Life == life-1; got != c.want {
			t.Errorf("%s: pinged = %v, want %v (life %d -> %d)", c.name, got, c.want, life, opp.Life)
		}
	}
}

package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b708_cards_test.go — S82 OLD batch 708 (#5622, Portal Three Kingdoms):
// damage. Each test plays the real catalog entry.

const (
	oldB708Rockslide   = "13958d6f-25f9-4a6c-a16c-e7a20e1b9117"
	oldB708Sandstorm   = "230bf3a4-68a8-4553-b8d7-ad340cbd7371"
	oldB708ZhangLiao   = "244541f6-cbfc-4e89-9156-64e412f745a7"
	oldB708Xinye       = "3a3cec13-4766-40d2-88e0-41bcdd5249e9"
	oldB708Rolling     = "46b96552-830e-4732-9560-9872aeec858b"
	oldB708Eunuchs     = "77e604fe-d6ca-495f-ba22-a7d8ba88c8b5"
	oldB708EastWind    = "79334732-8db1-4393-9fed-3cde3c572c5c"
	oldB708DongZhou    = "a1da64bc-8cff-4c89-9b9c-db0ddff5abc4"
	oldB708Fields      = "b3b58231-d532-423b-91d0-78d9c0f66f0f"
	oldB708Kongming    = "c3327ffe-f47f-4f5b-b929-a7706ef4932d"
	oldB708StolenGrain = "dc17836d-8edf-44bd-a9ee-346d6af25dbf"
)

func TestOldB708EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		oldB708Rockslide: "Rockslide Ambush", oldB708Sandstorm: "Desert Sandstorm", oldB708ZhangLiao: "Zhang Liao, Hero of Hefei",
		oldB708Xinye: "Burning of Xinye", oldB708Rolling: "Rolling Earthquake", oldB708Eunuchs: "Corrupt Eunuchs",
		oldB708EastWind: "Borrowing the East Wind", oldB708DongZhou: "Dong Zhou, the Tyrant", oldB708Fields: "Burning Fields",
		oldB708Kongming: "Kongming's Contraptions", oldB708StolenGrain: "Stolen Grain",
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

func oldB708Seats(g *game.Game) (*game.Player, *game.Player) {
	return g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
}

func oldB708Bear(g *game.Game, owner uuid.UUID, name string, toughness int, keywords ...string) uuid.UUID {
	return effectstest.PushTestPermanent(g, owner, game.Card{Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: toughness, Keywords: keywords})
}

// --- Rockslide Ambush --------------------------------------------------------

func TestOldB708RockslideAmbushDealsDamageEqualToYourMountains(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB708Seats(g)
	for i := 0; i < 3; i++ {
		effectstest.PushCatalogPermanent(g, me.ID, "Mountain", "Basic Land — Mountain", "", false)
	}
	effectstest.PushCatalogPermanent(g, opp.ID, "Mountain", "Basic Land — Mountain", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Forest", "Basic Land — Forest", "", false)
	big := oldB708Bear(g, opp.ID, "Big Bear", 6)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RR")
	effectstest.CastFromHandForTest(t, g, me, "Rockslide Ambush", "Sorcery", "{1}{R}", oldB708Rockslide,
		game.CastSpellParams{Strict: true, Targets: effectstest.Pr7bTargets(big)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1g2WantDamage(t, g, big, 3, "the Bear (three Mountains of mine, one of theirs)")
}

func TestOldB708RockslideAmbushRefusesANonCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB708Seats(g)
	rock := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RR")
	if _, err := effectstest.CastFromHandErr(g, me, "Rockslide Ambush", "Sorcery", "{1}{R}", oldB708Rockslide,
		game.CastSpellParams{Strict: true, Targets: effectstest.Pr7bTargets(rock)}); err == nil {
		t.Error("Rockslide Ambush targeted an artifact")
	}
}

// --- Desert Sandstorm --------------------------------------------------------

func TestOldB708DesertSandstormDamagesEveryCreatureByOneAndNobodyElse(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB708Seats(g)
	small := oldB708Bear(g, opp.ID, "Small", 1)
	mine := oldB708Bear(g, me.ID, "Mine", 3)
	theirs := oldB708Bear(g, opp.ID, "Theirs", 3)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RRR")
	life, oppLife := me.Life, opp.Life
	effectstest.CastFromHandForTest(t, g, me, "Desert Sandstorm", "Sorcery", "{2}{R}", oldB708Sandstorm, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the 2/1 survived 1 damage")
	}
	effectstest.P1g2WantDamage(t, g, mine, 1, "my creature")
	effectstest.P1g2WantDamage(t, g, theirs, 1, "their creature")
	if me.Life != life || opp.Life != oppLife {
		t.Error("a player took damage from Desert Sandstorm")
	}
}

// --- Rolling Earthquake and Borrowing the East Wind --------------------------

func oldB708HorsemanshipBoard(g *game.Game) (me, opp *game.Player, horse, plain uuid.UUID) {
	me, opp = oldB708Seats(g)
	horse = oldB708Bear(g, opp.ID, "Horse Bear", 5, "horsemanship")
	plain = oldB708Bear(g, opp.ID, "Plain Bear", 5)
	return
}

func TestOldB708RollingEarthquakeSparesHorsemanshipAndHitsEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, horse, plain := oldB708HorsemanshipBoard(g)
	life, oppLife := me.Life, opp.Life
	effectstest.Pr8CastX(t, g, "Rolling Earthquake", "Sorcery", oldB708Rolling, 3, nil)
	effectstest.P1g2WantDamage(t, g, horse, 0, "the horsemanship creature")
	effectstest.P1g2WantDamage(t, g, plain, 3, "the plain creature")
	if me.Life != life-3 || opp.Life != oppLife-3 {
		t.Errorf("life %d and %d, want both players down 3 (%d, %d)", me.Life, opp.Life, life-3, oppLife-3)
	}
}

func TestOldB708BorrowingTheEastWindHitsOnlyHorsemanshipAndEveryPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp, horse, plain := oldB708HorsemanshipBoard(g)
	life, oppLife := me.Life, opp.Life
	effectstest.Pr8CastX(t, g, "Borrowing the East Wind", "Sorcery", oldB708EastWind, 2, nil)
	effectstest.P1g2WantDamage(t, g, horse, 2, "the horsemanship creature")
	effectstest.P1g2WantDamage(t, g, plain, 0, "the plain creature")
	if me.Life != life-2 || opp.Life != oppLife-2 {
		t.Errorf("life %d and %d, want both players down 2", me.Life, opp.Life)
	}
}

// --- Burning Fields and Stolen Grain -----------------------------------------

func TestOldB708BurningFieldsDealsFiveToTheOpponentAndRefusesYourself(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB708Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RRRRR")
	if _, err := effectstest.CastFromHandErr(g, me, "Burning Fields", "Sorcery", "{4}{R}", oldB708Fields,
		game.CastSpellParams{Strict: true, Targets: effectstest.P8Player(me.ID)}); err == nil {
		t.Error("Burning Fields targeted its own controller")
	}
	oppLife := opp.Life
	effectstest.CastFromHandForTest(t, g, me, "Burning Fields", "Sorcery", "{4}{R}", oldB708Fields,
		game.CastSpellParams{Strict: true, Targets: effectstest.P8Player(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-5 {
		t.Errorf("opponent life %d, want %d", opp.Life, oppLife-5)
	}
}

func TestOldB708BurningFieldsHitsAPlaneswalker(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB708Seats(g)
	pw := effectstest.P1g2Planeswalker(g, opp.ID, 7)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RRRRR")
	effectstest.CastFromHandForTest(t, g, me, "Burning Fields", "Sorcery", "{4}{R}", oldB708Fields,
		game.CastSpellParams{Strict: true, Targets: effectstest.Pr7bTargets(pw)})
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, pw, game.CounterLoyalty); got != 2 {
		t.Errorf("loyalty = %d, want 2 after 5 damage to a 7-loyalty planeswalker", got)
	}
}

func TestOldB708StolenGrainDealsFiveAndGainsFive(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB708Seats(g)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "BBBBBB")
	life, oppLife := me.Life, opp.Life
	effectstest.CastFromHandForTest(t, g, me, "Stolen Grain", "Sorcery", "{4}{B}{B}", oldB708StolenGrain,
		game.CastSpellParams{Strict: true, Targets: effectstest.P8Player(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != oppLife-5 || me.Life != life+5 {
		t.Errorf("life %d / %d, want the opponent at %d and the caster at %d", opp.Life, me.Life, oppLife-5, life+5)
	}
}

// --- Corrupt Eunuchs and Dong Zhou -------------------------------------------

func oldB708CastCreature(t *testing.T, g *game.Game, me *game.Player, name, typeLine, cost, oracle, mana string, target uuid.UUID) {
	t.Helper()
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, mana)
	effectstest.CastFromHandForTest(t, g, me, name, typeLine, cost, oracle, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.LatestPickTarget(g, me.ID); pick != nil {
		effectstest.PickCard(t, g, me.ID, target)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB708CorruptEunuchsDealsTwoToTargetCreatureWhenItEnters(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB708Seats(g)
	victim := oldB708Bear(g, opp.ID, "Victim", 4)
	oldB708CastCreature(t, g, me, "Corrupt Eunuchs", "Creature — Human Advisor", "{3}{R}", oldB708Eunuchs, "RRRR", victim)
	effectstest.P1g2WantDamage(t, g, victim, 2, "the target")
}

func TestOldB708DongZhouMakesTheOpponentsCreatureHitItsController(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB708Seats(g)
	big := effectstest.PushTestPermanent(g, opp.ID, game.Card{Name: "Their Ogre", TypeLine: "Creature — Ogre", Power: 4, Toughness: 4})
	mine := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "My Ogre", TypeLine: "Creature — Ogre", Power: 5, Toughness: 5})
	life, oppLife := me.Life, opp.Life
	oldB708CastCreature(t, g, me, "Dong Zhou, the Tyrant", "Legendary Creature — Human Soldier", "{4}{R}", oldB708DongZhou, "RRRRR", big)
	if opp.Life != oppLife-4 {
		t.Errorf("their life %d, want %d: the Ogre deals its power 4 to its controller", opp.Life, oppLife-4)
	}
	if me.Life != life {
		t.Error("my life changed")
	}
	effectstest.P1g2WantDamage(t, g, big, 0, "the Ogre: it dealt the damage, it took none")
	_ = mine
}

// --- Zhang Liao --------------------------------------------------------------

func TestOldB708ZhangLiaoMakesTheDamagedOpponentDiscard(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB708Seats(g)
	zhang := effectstest.PushTestPermanent(g, me.ID, game.Card{Name: "Zhang Liao, Hero of Hefei", OracleID: oldB708ZhangLiao,
		TypeLine: "Legendary Creature — Human Soldier", Power: 3, Toughness: 3})
	effectstest.Unsick(g, zhang)
	effectstest.EmptyHand(g, opp)
	keep := effectstest.EoHand(opp, "Keep", "Sorcery")
	pitch := effectstest.EoHand(opp, "Pitch", "Sorcery")
	effectstest.Pr7bAttack(t, g, opp.ID, zhang)
	effectstest.Pr7bDamageStep(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerDiscard(t, g, opp.ID, pitch)
	effectstest.PassPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(pitch) || opp.Graveyard.Contains(keep) {
		t.Error("the opponent did not discard the card they chose")
	}
}

// --- Kongming's Contraptions -------------------------------------------------

func oldB708KongmingTable(t *testing.T, attackSeat int) (*game.Game, *game.Player, uuid.UUID, uuid.UUID) {
	t.Helper()
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	defender := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	other := g.Seats[(g.Turn.ActiveSeat+2)%len(g.Seats)]
	k := effectstest.PushTestPermanent(g, defender.ID, game.Card{Name: "Kongming's Contraptions", OracleID: oldB708Kongming,
		TypeLine: "Creature — Human Soldier", Power: 2, Toughness: 4})
	effectstest.Unsick(g, k)
	attacker := effectstest.PushTestPermanent(g, active.ID, game.Card{Name: "Attacker", TypeLine: "Creature — Bear", Power: 2, Toughness: 4})
	effectstest.Unsick(g, attacker)
	target := defender
	if attackSeat == 2 {
		target = other
	}
	effectstest.SfAttackHold(t, g, target.ID, attacker)
	return g, defender, k, attacker
}

func TestOldB708KongmingsContraptionsShootsAnAttackerWhenYouAreAttacked(t *testing.T) {
	g, defender, k, attacker := oldB708KongmingTable(t, 1)
	if g.Turn.Step != game.StepDeclareAttackers {
		t.Fatalf("step = %v, want the declare attackers step", g.Turn.Step)
	}
	if err := g.ActivateCatalogAbility(defender.ID, k, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.Pr7bTargets(attacker)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1g2WantDamage(t, g, attacker, 2, "the attacker")
	if c, _ := effectstest.BattlefieldCard(g, k); !c.Tapped {
		t.Error("the Contraptions were not tapped for the cost")
	}
}

func TestOldB708KongmingsContraptionsNeedsYouToBeTheOneAttacked(t *testing.T) {
	g, defender, k, attacker := oldB708KongmingTable(t, 2)
	if err := g.ActivateCatalogAbility(defender.ID, k, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.Pr7bTargets(attacker)}); err == nil {
		t.Error("activated while another player was the one attacked")
	}
}

func TestOldB708KongmingsContraptionsCannotBeUsedOutsideTheDeclareAttackersStep(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	defender := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	k := effectstest.PushTestPermanent(g, defender.ID, game.Card{Name: "Kongming's Contraptions", OracleID: oldB708Kongming,
		TypeLine: "Creature — Human Soldier", Power: 2, Toughness: 4})
	effectstest.Unsick(g, k)
	attacker := effectstest.PushTestPermanent(g, active.ID, game.Card{Name: "Attacker", TypeLine: "Creature — Bear", Power: 2, Toughness: 4})
	effectstest.Unsick(g, attacker)
	effectstest.Pr7bAttack(t, g, defender.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.ActivateCatalogAbility(defender.ID, k, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.Pr7bTargets(attacker)}); err == nil {
		t.Error("activated after the declare attackers step")
	}
}

// --- Burning of Xinye --------------------------------------------------------

func oldB708Lands(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	var ids []uuid.UUID
	for i := 0; i < n; i++ {
		ids = append(ids, effectstest.PushCatalogPermanent(g, owner, "Forest", "Basic Land — Forest", "", false))
	}
	return ids
}

func TestOldB708BurningOfXinyeEachPlayerChoosesFourLandsThenEveryCreatureTakesFour(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB708Seats(g)
	mine := oldB708Lands(g, me.ID, 6)
	theirs := oldB708Lands(g, opp.ID, 5)
	big := oldB708Bear(g, opp.ID, "Big", 5)
	small := oldB708Bear(g, me.ID, "Small", 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RRRRRR")
	effectstest.CastFromHandForTest(t, g, me, "Burning of Xinye", "Sorcery", "{4}{R}{R}", oldB708Xinye,
		game.CastSpellParams{Strict: true, Targets: effectstest.P8Player(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	// The caster chooses first: four of six.
	effectstest.AnswerChooseCards(t, g, me.ID, mine[0], mine[1], mine[2], mine[3])
	effectstest.PassPriorityAroundTable(t, g)
	// Then the opponent: four of five.
	effectstest.AnswerChooseCards(t, g, opp.ID, theirs[1], theirs[2], theirs[3], theirs[4])
	effectstest.PassPriorityAroundTable(t, g)
	for i, id := range mine {
		if g.Battlefield.Contains(id) == (i < 4) {
			t.Errorf("my land %d: on the battlefield = %v, want %v", i, g.Battlefield.Contains(id), i >= 4)
		}
	}
	for i, id := range theirs {
		if g.Battlefield.Contains(id) == (i >= 1 && i <= 4) {
			t.Errorf("their land %d: on the battlefield = %v, want %v", i, g.Battlefield.Contains(id), i == 0)
		}
	}
	if g.Battlefield.Contains(small) {
		t.Error("a creature with toughness 4 survived 4 damage")
	}
	effectstest.P1g2WantDamage(t, g, big, 4, "the big creature")
}

func TestOldB708BurningOfXinyeDestroysEverythingWhenAPlayerHasFourOrFewer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := oldB708Seats(g)
	mine := oldB708Lands(g, me.ID, 3)
	theirs := oldB708Lands(g, opp.ID, 4)
	effectstest.AdvanceToMain(t, g)
	effectstest.FloatForTest(g, me, "RRRRRR")
	effectstest.CastFromHandForTest(t, g, me, "Burning of Xinye", "Sorcery", "{4}{R}{R}", oldB708Xinye,
		game.CastSpellParams{Strict: true, Targets: effectstest.P8Player(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.PendingOfKind(g, game.PendingChoiceChooseCards) != nil {
		t.Fatal("asked to choose among four or fewer lands")
	}
	for _, id := range append(mine, theirs...) {
		if g.Battlefield.Contains(id) {
			t.Error("a land survived")
		}
	}
}

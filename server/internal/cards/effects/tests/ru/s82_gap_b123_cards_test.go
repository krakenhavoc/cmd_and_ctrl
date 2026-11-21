package testsru

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b123_cards_test.go — The Lord of the Rings: Tales of Middle-earth
// creatures (S82 GAP batch 123, #5948). Every test plays the real catalog
// entry through the engine.

const (
	gapB123FellBeastOracle    = "2e037c5e-3e07-4ffc-a388-7976459f6ff7"
	gapB123GorbagOracle       = "e77e97d7-58ff-4454-ba49-66182e550f12"
	gapB123LobeliaOracle      = "19f5a25a-97fc-41e6-bb26-7c35f4b4e37c"
	gapB123ShagratOracle      = "23db92aa-d4db-49e3-a5bc-7b120e02822f"
	gapB123TrebuchetOracle    = "8a8e0d0a-1101-45b6-8f70-1de2c07bc369"
	gapB123WillowOracle       = "96c96842-42c2-4aa3-95e2-6186d4ab5693"
	gapB123WargOracle         = "978a7a96-ca79-4f73-a4c6-99d662605969"
	gapB123BalrogDurinOracle  = "d113af33-d75d-44e4-a8dc-31f8984edc42"
	gapB123BalrogFlameOracle  = "de12ca82-4824-4975-9a31-309a6d9101a7"
	gapB123GothmogOracle      = "f3817f63-7bde-46c5-93cb-b4756f5f821e"
	gapB123LegendaryOrcType   = "Legendary Creature — Orc Soldier"
	gapB123OrcType            = "Creature — Orc Soldier"
	gapB123LegendaryHumanType = "Legendary Creature — Human"
)

func TestGapB123EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		gapB123FellBeastOracle:   "Voracious Fell Beast",
		gapB123GorbagOracle:      "Gorbag of Minas Morgul",
		gapB123LobeliaOracle:     "Lobelia Sackville-Baggins",
		gapB123ShagratOracle:     "Shagrat, Loot Bearer",
		gapB123TrebuchetOracle:   "Mordor Trebuchet",
		gapB123WillowOracle:      "Old Man Willow",
		gapB123WargOracle:        "Snarling Warg",
		gapB123BalrogDurinOracle: "The Balrog, Durin's Bane",
		gapB123BalrogFlameOracle: "The Balrog, Flame of Udûn",
		gapB123GothmogOracle:     "Gothmog, Morgul Lieutenant",
	} {
		effectstest.RequireFullNoCaveats(t, oracle, name)
	}
}

func gapB123Push(g *game.Game, owner uuid.UUID, name, typeLine string, power, toughness int) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: name, TypeLine: typeLine,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner})
}

func gapB123Library(g *game.Game, p *game.Player) {
	effectstest.StockLibraryFor(g, p, 8)
}

func TestGapB123VoraciousFellBeastEachOpponentSacrificesAndYouGetFoodForEach(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	o1, o2, o3 := g.Seats[1], g.Seats[2], g.Seats[3]
	a := effectstest.PushVanillaCreature(g, o1.ID, "One A", 2, 2)
	b := effectstest.PushVanillaCreature(g, o1.ID, "One B", 3, 3)
	only := effectstest.PushVanillaCreature(g, o2.ID, "Two Only", 1, 1)
	_ = o3 // no creature: nothing to sacrifice, no Food

	effectstest.TdCastFromHand(t, g, game.Card{Name: "Voracious Fell Beast", TypeLine: "Creature — Drake Beast", OracleID: gapB123FellBeastOracle,
		ManaCost: "{4}{B}{B}", Power: 4, Toughness: 4}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if pick := effectstest.ChooseCardsChoiceFor(g, o1.ID); pick != nil {
		effectstest.AnswerChooseCards(t, g, o1.ID, b)
	} else {
		t.Fatal("the opponent with a choice was not asked")
	}
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, b) || effectstest.BattlefieldHas(g, only) {
		t.Error("a chosen creature was not sacrificed")
	}
	if !effectstest.BattlefieldHas(g, a) {
		t.Error("the unchosen creature was sacrificed")
	}
	foods := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Food" && c.Controller == me.ID {
			foods++
		}
	}
	if foods != 2 {
		t.Errorf("%d Food tokens for the caster, want 2 (one per creature sacrificed)", foods)
	}
	spec, _ := effects.Lookup(gapB123FellBeastOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "flying" {
		t.Errorf("keywords = %v, want flying", spec.PrintedKeywords)
	}
}

func TestGapB123SnarlingWargGetsPlusOneOnlyWithYourOwnGoblinOrOrc(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	warg := effectstest.PushCatalogPermanent(g, me.ID, "Snarling Warg", "Creature — Wolf", gapB123WargOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == warg {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 3, 4
			}
		}
	})
	power := func() int { c := effectstest.LayeredCard(t, g, warg); return c.CurrentPower() }
	if got := power(); got != 3 {
		t.Fatalf("power = %d with no Goblin or Orc, want 3", got)
	}
	gapB123Push(g, opp.ID, "Their Orc", gapB123OrcType, 1, 1)
	if got := power(); got != 3 {
		t.Errorf("power = %d with only an opponent's Orc, want 3", got)
	}
	gapB123Push(g, me.ID, "My Goblin", "Creature — Goblin", 1, 1)
	if got := power(); got != 4 {
		t.Errorf("power = %d with my Goblin, want 4", got)
	}
}

func TestGapB123GothmogAmassesAndGivesCreatureTokensDeathtouch(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	plain := effectstest.PushVanillaCreature(g, me.ID, "Plain", 2, 2)
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Gothmog, Morgul Lieutenant", TypeLine: gapB123LegendaryHumanType, OracleID: gapB123GothmogOracle,
		ManaCost: "{3}{B}", Power: 3, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	var army uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Army") {
			army = c.InstanceID
		}
	}
	if army == uuid.Nil {
		t.Fatal("no Army was created")
	}
	if n := effectstest.CountersOn(g, army, game.CounterPlusOne); n != 1 {
		t.Errorf("Army has %d +1/+1 counters, want 1", n)
	}
	if !effectstest.EffectiveAbilitiesContain(t, g, army, "deathtouch") {
		t.Error("the Army token lacks deathtouch")
	}
	if effectstest.EffectiveAbilitiesContain(t, g, plain, "deathtouch") {
		t.Error("a nontoken creature gained deathtouch")
	}
}

func TestGapB123BalrogFlamePutsItselfOnTheBottomWhenAnOpponentsLegendDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	gapB123Library(g, me)
	balrog := effectstest.PushCatalogPermanent(g, me.ID, "The Balrog, Flame of Udûn", "Legendary Creature — Avatar Demon", gapB123BalrogFlameOracle, false)
	plain := effectstest.PushVanillaCreature(g, opp.ID, "Their Plain", 2, 2)
	mine := gapB123Push(g, me.ID, "My Legend", gapB123LegendaryHumanType, 1, 1)
	theirs := gapB123Push(g, opp.ID, "Their Legend", gapB123LegendaryHumanType, 1, 1)

	effectstest.P1Destroy(t, g, plain)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.P1Destroy(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, balrog) {
		t.Fatal("the Balrog left for a nonlegendary creature or its own side's legend")
	}
	effectstest.P1Destroy(t, g, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, balrog) {
		t.Fatal("the Balrog stayed after an opponent's legendary creature died")
	}
	if c, ok := me.Library.Cards[0], len(me.Library.Cards) > 0; !ok || c.InstanceID != balrog {
		t.Error("the Balrog is not on the bottom of its owner's library")
	}
}

func TestGapB123BalrogDurinsBaneCostsOneLessPerPermanentSacrificedThisTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	spec, _ := effects.Lookup(gapB123BalrogDurinOracle)
	if len(spec.SelfCostModifiers) != 1 {
		t.Fatalf("%d self cost modifiers, want 1", len(spec.SelfCostModifiers))
	}
	discount := func() int { return spec.SelfCostModifiers[0].Amount(game.CostQuery{Game: g, Controller: me.ID}) }
	if got := discount(); got != 0 {
		t.Fatalf("discount = %d with nothing sacrificed, want 0", got)
	}
	a := effectstest.PushVanillaCreature(g, me.ID, "Fodder A", 1, 1)
	b := effectstest.PushVanillaCreature(g, me.ID, "Fodder B", 1, 1)
	g.WithWriteLock(func() { _ = g.SacrificePermanentForEffect(a); _ = g.SacrificePermanentForEffect(b) })
	if got := discount(); got != 2 {
		t.Errorf("discount = %d after two sacrifices, want 2", got)
	}
}

func TestGapB123BalrogDurinsBaneCanOnlyBeBlockedByLegendaryCreatures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	balrog := effectstest.PushCatalogPermanent(g, me.ID, "The Balrog, Durin's Bane", "Legendary Creature — Avatar Demon", gapB123BalrogDurinOracle, false)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == balrog {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 7, 5
			}
		}
	})
	plainBlocker := effectstest.PushVanillaCreature(g, opp.ID, "Plain Blocker", 2, 2)
	legendBlocker := gapB123Push(g, opp.ID, "Legend Blocker", gapB123LegendaryHumanType, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, balrog)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(plainBlocker, balrog); err == nil {
		t.Error("a nonlegendary creature blocked The Balrog")
	}
	if err := g.DeclareBlocker(legendBlocker, balrog); err != nil {
		t.Errorf("a legendary creature could not block The Balrog: %v", err)
	}
	spec, _ := effects.Lookup(gapB123BalrogDurinOracle)
	if len(spec.PrintedKeywords) != 1 || spec.PrintedKeywords[0] != "haste" {
		t.Errorf("keywords = %v, want haste", spec.PrintedKeywords)
	}
}

func TestGapB123BalrogDurinsBaneDestroysATargetWhenItDies(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	balrog := effectstest.PushCatalogPermanent(g, me.ID, "The Balrog, Durin's Bane", "Legendary Creature — Avatar Demon", gapB123BalrogDurinOracle, false)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 1, 1)
	effectstest.P1Destroy(t, g, balrog)
	effectstest.PassPriorityAroundTable(t, g)
	answerPendingTargetIfAny(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, victim) {
		t.Error("the dies trigger did not destroy the target")
	}
}

// answerPendingTargetIfAny picks `pick` for an open target prompt of `who`,
// if the engine asked (it does not when only one target is legal).
func answerPendingTargetIfAny(t *testing.T, g *game.Game, who, pick uuid.UUID) {
	t.Helper()
	if p := effectstest.LatestPickTarget(g, who); p != nil {
		if err := g.ResolvePickTarget(p.ID, who, game.TargetRef{Kind: game.TargetCard, ID: pick}); err != nil {
			t.Fatalf("ResolvePickTarget: %v", err)
		}
	}
}

func TestGapB123LobeliaExilesAnOpponentsCreatureThatDiedThisTurnForTreasures(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	victim := gapB123Push(g, opp.ID, "Dead Bear", "Creature — Bear", 3, 3)
	mine := gapB123Push(g, me.ID, "My Dead Bear", "Creature — Bear", 5, 5)
	effectstest.P1Destroy(t, g, victim)
	effectstest.P1Destroy(t, g, mine)
	effectstest.PassPriorityAroundTable(t, g)
	older := uuid.New()
	g.WithWriteLock(func() {
		opp.Graveyard.PushTop(game.Card{InstanceID: older, Name: "Old Corpse", TypeLine: "Creature — Bear", Power: 4, Toughness: 4, Owner: opp.ID, Controller: opp.ID})
	})
	effectstest.TdCastFromHand(t, g, game.Card{Name: "Lobelia Sackville-Baggins", TypeLine: "Legendary Creature — Halfling Citizen", OracleID: gapB123LobeliaOracle,
		ManaCost: "{2}{B}", Power: 2, Toughness: 3}, game.CastSpellParams{})
	effectstest.PassPriorityAroundTable(t, g)
	answerPendingTargetIfAny(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, victim) {
		t.Fatal("the opponent's creature card that died this turn was not exiled")
	}
	if effectstest.GraveyardHas(opp, older) == false || effectstest.InExile(g, older) {
		t.Error("a creature card that did not die this turn was exiled")
	}
	if effectstest.InExile(g, mine) {
		t.Error("my own creature card was exiled")
	}
	treasures := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Treasure" && c.Controller == me.ID {
			treasures++
		}
	}
	if treasures != 3 {
		t.Errorf("%d Treasure tokens, want 3 (the exiled card's power)", treasures)
	}
	spec, _ := effects.Lookup(gapB123LobeliaOracle)
	if len(spec.PrintedKeywords) != 2 {
		t.Errorf("keywords = %v, want flash and menace", spec.PrintedKeywords)
	}
}

func TestGapB123ShagratAttachesAnEquipmentAndAmassesByTheNumberAttached(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	shagrat := effectstest.PushCatalogPermanent(g, me.ID, "Shagrat, Loot Bearer", gapB123LegendaryOrcType, gapB123ShagratOracle, false)
	// An Equipment already on Shagrat, and one an opponent controls.
	held := effectstest.PushCatalogPermanent(g, me.ID, "Held Blade", "Artifact — Equipment", "", false)
	theirs := effectstest.PushCatalogPermanent(g, opp.ID, "Their Blade", "Artifact — Equipment", "", false)
	g.WithWriteLock(func() {
		_ = g.AttachForEffect(held, game.TargetRef{Kind: game.TargetCard, ID: shagrat})
	})
	effectstest.DeclareAttack(t, g, opp.ID, shagrat)
	answerPendingTargetIfAny(t, g, me.ID, theirs)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.AttachedTo(t, g, theirs, shagrat) {
		t.Fatal("the opponent's Equipment was not attached to Shagrat")
	}
	if c, _ := effectstest.BattlefieldCard(g, theirs); c.Controller != opp.ID {
		t.Error("attaching changed control of the Equipment")
	}
	var army uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && c.HasSubtype("Army") {
			army = c.InstanceID
		}
	}
	if army == uuid.Nil {
		t.Fatal("no Army was created")
	}
	if n := effectstest.CountersOn(g, army, game.CounterPlusOne); n != 2 {
		t.Errorf("Army has %d counters, want 2 (two Equipment attached)", n)
	}
}

func TestGapB123MordorTrebuchetIgnoresAnAttackWithNoGoblinOrOrc(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Mordor Trebuchet", "Artifact Creature — Wall", gapB123TrebuchetOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Ballistic Boulder")); n != 0 {
		t.Fatalf("%d Boulders for a Bear attack, want 0", n)
	}
}

func TestGapB123MordorTrebuchetMakesABoulderWhenAGoblinOrOrcAttacks(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Mordor Trebuchet", "Artifact Creature — Wall", gapB123TrebuchetOracle, false)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	orc := gapB123Push(g, me.ID, "Orc Raider", gapB123OrcType, 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear, orc)
	answerAttackSplitIfAny(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	answerAttackSplitIfAny(t, g, me)
	effectstest.PassPriorityAroundTable(t, g)
	boulders := effectstest.BattlefieldIDsNamed(g, "Ballistic Boulder")
	if len(boulders) != 1 {
		t.Fatalf("%d Boulders for one declaration with an Orc, want 1", len(boulders))
	}
	c, _ := effectstest.BattlefieldCard(g, boulders[0])
	if !c.Tapped || c.AttackingTarget == uuid.Nil || c.Power != 2 || c.Toughness != 1 {
		t.Errorf("Boulder tapped=%v attacking=%v %d/%d, want tapped, attacking 2/1", c.Tapped, c.AttackingTarget != uuid.Nil, c.Power, c.Toughness)
	}
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.BattlefieldHas(g, boulders[0]) {
		t.Error("the Boulder was not sacrificed at end of combat")
	}
}

// answerAttackSplitIfAny answers the "which defender does it attack?" prompts
// of a token that enters attacking.
func answerAttackSplitIfAny(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	for c := effectstest.LatestChoiceOfKind(g, game.PendingChoicePayAmount); c != nil; c = effectstest.LatestChoiceOfKind(g, game.PendingChoicePayAmount) {
		if err := g.ResolvePayAmount(c.ID, me.ID, c.PayAmount.Max); err != nil {
			t.Fatalf("place the token: %v", err)
		}
	}
}

func TestGapB123GorbagSacrificesAnOrcThatConnectsForACardOrATreasure(t *testing.T) {
	for _, mode := range []int{0, 1} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
		gapB123Library(g, me)
		effectstest.PushCatalogPermanent(g, me.ID, "Gorbag of Minas Morgul", gapB123LegendaryOrcType, gapB123GorbagOracle, false)
		orc := gapB123Push(g, me.ID, "Orc Raider", gapB123OrcType, 2, 2)
		before := me.Hand.Size()
		effectstest.DeclareAttack(t, g, opp.ID, orc)
		effectstest.AdvanceTo(t, g, game.StepCombatDamage)
		effectstest.PassPriorityAroundTable(t, g)
		ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
		if ask == nil {
			t.Fatalf("mode %d: Gorbag's sacrifice was not offered", mode)
		}
		if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
			t.Fatalf("ResolveConfirm: %v", err)
		}
		if effectstest.BattlefieldHas(g, orc) {
			t.Fatalf("mode %d: the Orc was not sacrificed", mode)
		}
		effectstest.PassPriorityAroundTable(t, g)
		mp := effectstest.ModePickChoiceFor(g, me.ID)
		if mp == nil {
			t.Fatalf("mode %d: the reflexive trigger did not ask for a mode", mode)
		}
		if err := g.ResolveModePick(mp.ID, me.ID, []int{mode}); err != nil {
			t.Fatalf("ResolveModePick: %v", err)
		}
		effectstest.PassPriorityAroundTable(t, g)
		if mode == 0 && me.Hand.Size() != before+1 {
			t.Errorf("hand grew by %d, want 1", me.Hand.Size()-before)
		}
		if mode == 1 {
			if n := len(effectstest.BattlefieldIDsNamed(g, "Treasure")); n != 1 {
				t.Errorf("%d Treasure tokens, want 1", n)
			}
		}
	}
}

func TestGapB123GorbagDecliningLeavesTheCreatureAlone(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	effectstest.PushCatalogPermanent(g, me.ID, "Gorbag of Minas Morgul", gapB123LegendaryOrcType, gapB123GorbagOracle, false)
	orc := gapB123Push(g, me.ID, "Orc Raider", gapB123OrcType, 2, 2)
	bear := effectstest.PushVanillaCreature(g, me.ID, "Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, bear, orc)
	effectstest.AdvanceTo(t, g, game.StepCombatDamage)
	effectstest.PassPriorityAroundTable(t, g)
	// The Bear is no Goblin or Orc: one question, for the Orc only.
	ask := effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm)
	if ask == nil {
		t.Fatal("Gorbag's sacrifice was not offered for the Orc")
	}
	if err := g.ResolveConfirm(ask.ID, me.ID, false); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, orc) {
		t.Error("the Orc was sacrificed after declining")
	}
	if effectstest.LatestChoiceOfKind(g, game.PendingChoiceConfirm) != nil {
		t.Error("a second sacrifice question was asked for the Bear")
	}
}

func TestGapB123OldManWillowIsAsBigAsYourLandsAndShrinksAnOpponentsCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	willow := effectstest.PushCatalogPermanent(g, me.ID, "Old Man Willow", "Legendary Creature — Treefolk", gapB123WillowOracle, false)
	for i := 0; i < 3; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	if c := effectstest.LayeredCard(t, g, willow); c.CurrentPower() != 3 || c.CurrentToughness() != 3 {
		t.Fatalf("Willow is %d/%d with three lands, want 3/3", c.CurrentPower(), c.CurrentToughness())
	}
	fodder := effectstest.PushVanillaCreature(g, me.ID, "Fodder", 1, 1)
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)

	effectstest.DeclareAttack(t, g, opp.ID, willow)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("the optional sacrifice was not offered")
	}
	for _, id := range pick.ChooseCards {
		if id == willow {
			t.Error("Willow offered itself as the sacrifice")
		}
	}
	effectstest.AnswerChooseCards(t, g, me.ID, fodder)
	if effectstest.BattlefieldHas(g, fodder) {
		t.Fatal("the creature was not sacrificed")
	}
	effectstest.PassPriorityAroundTable(t, g)
	answerPendingTargetIfAny(t, g, me.ID, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.LayeredCard(t, g, victim); c.CurrentPower() != 1 || c.CurrentToughness() != 1 {
		t.Errorf("the opponent's creature is %d/%d, want 1/1", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestGapB123OldManWillowCanSacrificeATokenAndDecline(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[1]
	willow := effectstest.PushCatalogPermanent(g, me.ID, "Old Man Willow", "Legendary Creature — Treefolk", gapB123WillowOracle, false)
	for i := 0; i < 2; i++ {
		effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest", Owner: me.ID, Controller: me.ID})
	}
	treasure := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Treasure", TypeLine: "Token Artifact — Treasure",
		Owner: me.ID, Controller: me.ID})
	victim := effectstest.PushVanillaCreature(g, opp.ID, "Victim", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, willow)
	effectstest.PassPriorityAroundTable(t, g)
	pick := effectstest.ChooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("a token was not offered as a sacrifice")
	}
	found := false
	for _, id := range pick.ChooseCards {
		found = found || id == treasure
	}
	if !found {
		t.Error("the Treasure token was not among the choices")
	}
	// Declining: no sacrifice, no -2/-2.
	effectstest.AnswerChooseCards(t, g, me.ID)
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.BattlefieldHas(g, treasure) {
		t.Error("the token was sacrificed after declining")
	}
	if c := effectstest.LayeredCard(t, g, victim); c.CurrentPower() != 3 {
		t.Error("the opponent's creature shrank without a sacrifice")
	}
}

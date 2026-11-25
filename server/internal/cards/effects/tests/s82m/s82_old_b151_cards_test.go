package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b151_cards_test.go — S82 older sets, Commander 2019 batch 151
// (#5517): Kadena's Silencer; Volrath, the Shapestealer; Silumgar Assassin;
// Divine Reckoning; Desperate Ravings; Devil's Play; Tahngarth, First Mate;
// Emmara Tandris; Wall of Stolen Identity. (Rayami, First of the Fallen is
// skipped; see the commit message.)

const (
	c19B151SilencerOracle = "02095128-82cd-4531-b47e-471307746f81"
	c19B151VolrathOracle  = "a70d8e3d-a62d-4c4c-890b-e502e48c4475"
	c19B151AssassinOracle = "f496e8cb-1d79-4a78-be15-fdb7e3d0de71"
	c19B151ReckoningOracl = "15d8b339-7520-43c9-b2dc-4a936357042b"
	c19B151RavingsOracle  = "8b96c334-82c7-4511-805e-f04c27514d7c"
	c19B151PlayOracle     = "ee9f0b29-8a54-4cb8-8e2c-7bd67c2184ba"
	c19B151TahngarthOracl = "1a8a41a1-6fc5-4dc8-bd28-9bf404c777d7"
	c19B151EmmaraOracle   = "38a71429-e673-4de4-80a8-f59e5cf98acc"
	c19B151WallOracle     = "92c58354-e106-402f-8348-6c7cdb67c6df"
	c19B151ShadowOracle   = "928c62fe-9c4d-4e89-bd7a-0d3b6a81f393"
)

func TestC19B151CardsAreRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		c19B151SilencerOracle: "Kadena's Silencer",
		c19B151VolrathOracle:  "Volrath, the Shapestealer",
		c19B151AssassinOracle: "Silumgar Assassin",
		c19B151ReckoningOracl: "Divine Reckoning",
		c19B151RavingsOracle:  "Desperate Ravings",
		c19B151PlayOracle:     "Devil's Play",
		c19B151TahngarthOracl: "Tahngarth, First Mate",
		c19B151EmmaraOracle:   "Emmara Tandris",
		c19B151WallOracle:     "Wall of Stolen Identity",
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

func TestC19B151SilencerCountersOnlyOpponentAbilitiesWhenTurnedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	shadow := effectstest.PushCatalogPermanent(g, opp.ID, "Misery's Shadow", "Creature — Shade", c19B151ShadowOracle, false)
	id := effectstest.S36CastFaceDown(t, g, "Kadena's Silencer", "Creature — Snake Wizard", c19B151SilencerOracle, "megamorph", 2, 1)
	g.WithWriteLock(func() { opp.ManaPool.AddMana(effectstest.ManaTokens("C")...) })
	if err := g.ActivateCatalogAbility(opp.ID, shadow, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("opponent's pump: %v", err)
	}
	before := len(g.StackMeta)
	effectstest.S36TurnUp(t, g, me, id, "U", "C")
	effectstest.PassPriorityAroundTable(t, g)
	if before == 0 {
		t.Fatal("the opponent's ability never reached the stack")
	}
	if p, _ := effectstest.PtOf(t, g, shadow); p != 1 {
		t.Errorf("Misery's Shadow has power %d, want its base 1: its pump should have been countered", p)
	}
	if got := effectstest.CounterCount(g, id, game.CounterPlusOne); got != 1 {
		t.Errorf("%d +1/+1 counters on the megamorphed creature, want 1", got)
	}
}

func TestC19B151AssassinCantBeBlockedByGreaterPower(t *testing.T) {
	for _, tc := range []struct {
		name    string
		power   int
		refused bool
	}{{"power 3 blocker", 3, true}, {"power 2 blocker", 2, false}, {"power 1 blocker", 1, false}} {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			assassin := effectstest.B12Push(g, me.ID, "Silumgar Assassin", "Creature — Human Assassin", c19B151AssassinOracle, 2, 1)
			blocker := effectstest.B12Creature(g, opp.ID, "Blocker", "Creature — Bear", tc.power, 4)
			effectstest.AdvanceToMain(t, g)
			refused := effectstest.Msc13BlockRefused(t, g, opp.ID, blocker, assassin)
			if refused[assassin] != tc.refused {
				t.Errorf("block refused = %v, want %v", refused[assassin], tc.refused)
			}
		})
	}
}

func TestC19B151AssassinDestroysASmallOpposingCreatureWhenTurnedFaceUp(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	small := effectstest.B12Creature(g, opp.ID, "Small", "Creature — Bear", 3, 3)
	big := effectstest.B12Creature(g, opp.ID, "Big", "Creature — Bear", 4, 4)
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 1, 1)
	id := effectstest.S36CastFaceDown(t, g, "Silumgar Assassin", "Creature — Human Assassin", c19B151AssassinOracle, "megamorph", 2, 1)
	effectstest.S36TurnUp(t, g, me, id, "B", "C", "C")
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerPickTargetIfOpen(t, g, small)
	effectstest.PassPriorityAroundTable(t, g)
	if g.Battlefield.Contains(small) {
		t.Error("the power-3 creature survived")
	}
	if !g.Battlefield.Contains(big) || !g.Battlefield.Contains(mine) {
		t.Error("an illegal target was destroyed")
	}
}

func TestC19B151EmmaraPreventsDamageToYourCreatureTokensOnly(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.PushCatalogPermanent(g, me.ID, "Emmara Tandris", "Legendary Creature — Elf Shaman", c19B151EmmaraOracle, false)
	effectstest.AdvanceToMain(t, g)
	if err := (effects.CreateToken{Controller: me.ID, Template: effects.TokenCard("2/2 green Bear"), N: 1}).Apply(effects.KeyedContext(g, me.ID, uuid.Nil)); err != nil {
		t.Fatalf("token: %v", err)
	}
	var token uuid.UUID
	for _, c := range g.Battlefield.Cards {
		if c.Controller == me.ID && effects.IsToken(c) {
			token = c.InstanceID
		}
	}
	mineNonToken := effectstest.B12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirToken := effectstest.B12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	src := effectstest.B12Creature(g, opp.ID, "Pinger", "Creature — Bear", 1, 1)
	effectstest.DealDamageTo(t, g, src, token, 1)
	effectstest.DealDamageTo(t, g, src, mineNonToken, 1)
	effectstest.DealDamageTo(t, g, src, theirToken, 1)
	if got := effectstest.DamageMarkedOn(g, token); got != 0 {
		t.Errorf("my token has %d damage, want 0", got)
	}
	if got := effectstest.DamageMarkedOn(g, mineNonToken); got != 1 {
		t.Errorf("my nontoken creature has %d damage, want 1", got)
	}
	if got := effectstest.DamageMarkedOn(g, theirToken); got != 1 {
		t.Errorf("their creature has %d damage, want 1", got)
	}
}

func TestC19B151DevilsPlayDealsXDamageAndFlashesBack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	effectstest.B06AddMana(me, "R", "C", "C", "C")
	life := opp.Life
	id := effectstest.CastFromHandForTest(t, g, me, "Devil's Play", "Sorcery", "{X}{R}", c19B151PlayOracle,
		game.CastSpellParams{Strict: true, XValue: 3, Targets: effectstest.PlayerRefs(opp.ID)})
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-3 {
		t.Fatalf("opponent at %d, want %d", opp.Life, life-3)
	}
	if !me.Graveyard.Contains(id) {
		t.Fatal("Devil's Play is not in the graveyard")
	}
	effectstest.B06AddMana(me, "R", "R", "R", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback", XValue: 2,
		Targets: effectstest.PlayerRefs(opp.ID)}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if opp.Life != life-5 {
		t.Errorf("opponent at %d, want %d after the flashback for X=2", opp.Life, life-5)
	}
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back card was not exiled")
	}
}

func TestC19B151DesperateRavingsDrawsTwoThenDiscardsAtRandom(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	effectstest.AdvanceToMain(t, g)
	for i := 0; i < 4; i++ {
		me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Sorcery", Owner: me.ID, Controller: me.ID})
	}
	effectstest.B06AddMana(me, "R", "C")
	hand, gy := me.Hand.Size(), me.Graveyard.Size()
	id := effectstest.CastFromHandForTest(t, g, me, "Desperate Ravings", "Instant", "{1}{R}", c19B151RavingsOracle, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+1 {
		t.Errorf("hand %d, want %d (drew two, discarded one at random)", me.Hand.Size(), hand+1)
	}
	if me.Graveyard.Size() != gy+2 {
		t.Errorf("graveyard %d, want %d (the spell and one discard)", me.Graveyard.Size(), gy+2)
	}
	if effectstest.DiscardChoiceFor(g, me.ID) != nil {
		t.Error("the random discard asked the player to choose")
	}
	effectstest.B06AddMana(me, "U", "C", "C")
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{Strict: true, FromZone: "graveyard", AlternativeCost: "flashback"}); err != nil {
		t.Fatalf("flashback: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !g.Exile.Contains(id) {
		t.Error("the flashed-back card was not exiled")
	}
}

func TestC19B151DivineReckoningKeepsOneCreaturePerPlayer(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.AdvanceToMain(t, g)
	m1 := effectstest.B12Creature(g, me.ID, "Mine 1", "Creature — Bear", 2, 2)
	m2 := effectstest.B12Creature(g, me.ID, "Mine 2", "Creature — Bear", 2, 2)
	o1 := effectstest.B12Creature(g, opp.ID, "Theirs 1", "Creature — Bear", 2, 2)
	o2 := effectstest.B12Creature(g, opp.ID, "Theirs 2", "Creature — Bear", 2, 2)
	o3 := effectstest.B12Creature(g, opp.ID, "Theirs 3", "Creature — Bear", 2, 2)
	lone := effectstest.B12Creature(g, g.Seats[(g.Turn.ActiveSeat+2)%4].ID, "Lone", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "W", "W", "C", "C")
	effectstest.CastFromHandForTest(t, g, me, "Divine Reckoning", "Sorcery", "{2}{W}{W}", c19B151ReckoningOracl, game.CastSpellParams{Strict: true})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerChooseCards(t, g, me.ID, m1)
	effectstest.AnswerChooseCards(t, g, opp.ID, o2)
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{m1, o2, lone} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("a kept creature %s was destroyed", id)
		}
	}
	for _, id := range []uuid.UUID{m2, o1, o3} {
		if g.Battlefield.Contains(id) {
			t.Errorf("an unchosen creature %s survived", id)
		}
	}
}

func TestC19B151VolrathPutsAMinusCounterAtCombatAndCopiesACounteredCreature(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	volrath := effectstest.PushCatalogPermanent(g, me.ID, "Volrath, the Shapestealer", "Legendary Creature — Phyrexian Shapeshifter", c19B151VolrathOracle, false)
	victim := effectstest.B12Creature(g, opp.ID, "Victim", "Creature — Bear", 3, 3)
	effectstest.AdvanceTo(t, g, game.StepBeginCombat)
	effectstest.AnswerPickTargetIfOpen(t, g, victim)
	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.CounterCount(g, victim, game.CounterMinusOne); got != 1 {
		t.Fatalf("%d -1/-1 counters on the target, want 1", got)
	}
	effectstest.B06AddMana(me, "C")
	if err := g.ActivateCatalogAbility(me.ID, volrath, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(volrath)}); err == nil {
		t.Fatal("copied a creature with no counter on it")
	}
	if err := g.ActivateCatalogAbility(me.ID, volrath, 0, game.ActivateAbilityParams{Strict: true, Targets: effectstest.CardRefs(victim)}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if c := effectstest.FindCardAnywhere(t, g, volrath); c.CurrentPower() != 7 || c.CurrentToughness() != 5 {
		t.Errorf("Volrath is %d/%d as a copy, want 7/5", c.CurrentPower(), c.CurrentToughness())
	}
}

func TestC19B151WallOfStolenIdentityCopiesTapsAndHoldsTheOriginal(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	original := effectstest.B12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	effectstest.AdvanceToMain(t, g)
	wall := effectstest.CastCatalogSpell(t, g, "Wall of Stolen Identity", "Creature — Shapeshifter Wall", c19B151WallOracle, nil)
	_ = me
	effectstest.ResolveWithCopyChoice(t, g, original)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, wall)
	if !ok {
		t.Fatal("the Wall did not enter")
	}
	if c.CurrentPower() != 4 || !c.HasSubtype("Wall") || !c.HasSubtype("Ogre") {
		t.Errorf("the copy is %d/%d %q, want a 4/4 Ogre Wall", c.CurrentPower(), c.CurrentToughness(), c.TypeLine)
	}
	if !effectstest.HasKeywordOnBattlefield(t, g, wall, "defender") {
		t.Error("the copy lacks defender")
	}
	if !effectstest.IsTapped(g, original) {
		t.Error("the copied creature was not tapped")
	}
}

func TestC19B151TahngarthChangesSidesOnAnOpponentsAttack(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tahngarth := effectstest.PushCatalogPermanent(g, me.ID, "Tahngarth, First Mate", "Legendary Creature — Minotaur Warrior", c19B151TahngarthOracl, false)
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == tahngarth {
				g.Battlefield.Cards[i].Power, g.Battlefield.Cards[i].Toughness = 5, 5
				g.Battlefield.Cards[i].Tapped = true
			}
		}
	})
	// The opponent's turn: advance the table to seat 1's declare attackers.
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.DeclareAttack(t, g, me.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerMayChoice(t, g, me.ID, true)
	effectstest.PassPriorityAroundTable(t, g)
	c, ok := effectstest.BattlefieldCard(g, tahngarth)
	if !ok || c.Controller != opp.ID {
		t.Fatalf("Tahngarth did not change controllers: %+v", c.Controller)
	}
	if c.AttackingTarget != me.ID {
		t.Errorf("Tahngarth attacks %v, want the only attacked player %v", c.AttackingTarget, me.ID)
	}
}

func TestC19B151TahngarthDoesNothingWhileUntapped(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tahngarth := effectstest.PushCatalogPermanent(g, me.ID, "Tahngarth, First Mate", "Legendary Creature — Minotaur Warrior", c19B151TahngarthOracl, false)
	attacker := effectstest.B12Creature(g, opp.ID, "Attacker", "Creature — Bear", 2, 2)
	effectstest.AdvanceToMainOf(t, g, 1)
	effectstest.DeclareAttack(t, g, me.ID, attacker)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.LatestConfirmFor(g, me.ID) != nil {
		t.Fatal("Tahngarth asked about changing sides while untapped")
	}
	if c, _ := effectstest.BattlefieldCard(g, tahngarth); c.Controller != me.ID {
		t.Error("Tahngarth changed controllers while untapped")
	}
}

package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	o334Avatar      = "15fd66db-f9dd-40ab-92bc-9e0575bd7489"
	o334Sentinels   = "5c0c01c1-c207-4348-aaec-fbefa77cfcd4"
	o334Quarter     = "9398fbb4-a2b9-4ae4-9e0f-2553eb570501"
	o334Dance       = "5ab0ab9b-c143-4315-8983-f645fb21f8b3"
	o334LoseCalm    = "73d71731-f694-471c-b2ad-a36854bb06a1"
	o334Wind        = "905a6d6c-5ef8-4afa-89bb-9d85a854743f"
	o334Teachings   = "95a5e444-39f6-4f42-a83f-befa76350bc1"
	o334Necromaster = "b03fc155-2a70-4106-853c-33bdc88d548c"
	o334Descent     = "d32c6f64-c4b0-4451-8986-a0abc1fc2bb3"
	o334Hunter      = "11c8e218-0cd5-4356-89fd-f39f989e487b"
	o334Herald      = "dae8dc90-8047-434c-9e3a-a2eaa6aba575"
)

func TestO334CardsAreFullWithNoCaveats(t *testing.T) {
	for oracle, name := range map[string]string{
		o334Avatar: "Avatar of the Resolute", o334Sentinels: "Scaleguard Sentinels", o334Quarter: "Salt Road Quartermasters",
		o334Dance: "Dance of the Skywise", o334LoseCalm: "Lose Calm", o334Wind: "Revealing Wind",
		o334Teachings: "Skywise Teachings", o334Necromaster: "Necromaster Dragon", o334Descent: "Descent of the Dragons",
		o334Hunter: "Dragon Hunter", o334Herald: "Herald of Dromoka",
	} {
		spec, ok := effects.Lookup(oracle)
		if !ok || spec.Name != name || spec.Completeness != effects.CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: registered %v as %q (%v, caveats %v)", name, ok, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func o334WithCounter(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: "Creature — Bear", Power: 2, Toughness: 2,
		Counters: map[string]int{game.CounterPlusOne: 1}, Owner: owner, Controller: owner,
	})
}

func TestO334AvatarEntersWithACounterPerOtherCreatureWithACounter(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	o334WithCounter(g, me.ID, "A")
	o334WithCounter(g, me.ID, "B")
	o334WithCounter(g, opp.ID, "Theirs")
	effectstest.B12Creature(g, me.ID, "Plain", "Creature — Bear", 2, 2)
	avatar := effectstest.CastCatalogSpell(t, g, "Avatar of the Resolute", "Creature — Avatar", o334Avatar, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldPlusOnes(g, avatar); n != 2 {
		t.Errorf("the Avatar entered with %d counters, want 2 (not its own, not an opponent's, not a creature with none)", n)
	}
}

func TestO334SentinelsEnterWithACounterOnlyWithADragonRevealedOrControlled(t *testing.T) {
	for _, tc := range []string{"nothing", "controlled", "revealed"} {
		t.Run(tc, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			effectstest.AdvanceToMain(t, g)
			params := game.CastSpellParams{}
			switch tc {
			case "controlled":
				effectstest.B12Creature(g, me.ID, "Dragon", "Creature — Dragon", 4, 4)
			case "revealed":
				dragon := effectstest.PutInHand(me, game.Card{Name: "Hand Dragon", TypeLine: "Creature — Dragon", Power: 4, Toughness: 4})
				params = game.CastSpellParams{OptionalCosts: []int{0}, RevealIDs: []uuid.UUID{dragon}}
			}
			id := uuid.New()
			me.Hand.PushTop(game.Card{InstanceID: id, Name: "Scaleguard Sentinels", TypeLine: "Creature — Human Soldier", OracleID: o334Sentinels,
				ManaCost: "{G}{G}", Power: 2, Toughness: 3, Owner: me.ID, Controller: me.ID})
			effectstest.B06AddMana(me, "G", "G")
			if err := g.CastSpell(me.ID, id, params); err != nil {
				t.Fatalf("cast: %v", err)
			}
			effectstest.PassPriorityAroundTable(t, g)
			want := 1
			if tc == "nothing" {
				want = 0
			}
			if n := effectstest.BattlefieldPlusOnes(g, id); n != want {
				t.Errorf("%s: %d counters, want %d", tc, n, want)
			}
		})
	}
}

func TestO334QuartermastersEnterWithTwoCountersAndMoveOneAtATime(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	qm := effectstest.CastCatalogSpell(t, g, "Salt Road Quartermasters", "Creature — Human Soldier", o334Quarter, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldPlusOnes(g, qm); n != 2 {
		t.Fatalf("entered with %d counters, want 2", n)
	}
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.B06AddMana(me, "G", "G", "G")
	effectstest.B16Activate(t, g, me.ID, qm, 0, game.ActivateAbilityParams{Strict: true, Targets: o334Card(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	if n := effectstest.BattlefieldPlusOnes(g, qm); n != 1 {
		t.Errorf("Quartermasters has %d counters after the activation, want 1", n)
	}
	if n := effectstest.BattlefieldPlusOnes(g, bear); n != 1 {
		t.Errorf("the target has %d counters, want 1", n)
	}
	effectstest.B06AddMana(me, "G", "G", "G")
	effectstest.B16Activate(t, g, me.ID, qm, 0, game.ActivateAbilityParams{Strict: true, Targets: o334Card(bear)})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B06AddMana(me, "G", "G", "G")
	if err := g.ActivateCatalogAbility(me.ID, qm, 0, game.ActivateAbilityParams{Strict: true, Targets: o334Card(bear)}); err == nil {
		t.Error("activated with no +1/+1 counter left to remove")
	}
}

func TestO334DanceMakesAFourFourFlyingDragonIllusionWithNoOtherAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.BrCreature(g, me.ID, "Trampler", "Creature — Bear", 2, 2, "trample", "haste")
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	if err := effectstest.CastCatalogSpellErrStrict(t, g, "Dance of the Skywise", "Instant", o334Dance, o334Card(theirs)); err == nil {
		t.Fatal("Dance of the Skywise targeted a creature its caster doesn't control")
	}
	effectstest.CastCatalogSpell(t, g, "Dance of the Skywise", "Instant", o334Dance, o334Card(mine))
	effectstest.PassPriorityAroundTable(t, g)
	if p, th := effectstest.EffectivePower(t, g, mine), effectstest.EffectiveToughness(t, g, mine); p != 4 || th != 4 {
		t.Errorf("the creature is %d/%d, want 4/4", p, th)
	}
	abilities := effectstest.EffectiveAbilities(t, g, mine)
	if !effectstest.HasAbility(abilities, "flying") || effectstest.HasAbility(abilities, "trample") || effectstest.HasAbility(abilities, "haste") {
		t.Errorf("abilities %v, want flying and nothing else", abilities)
	}
}

func TestO334LoseCalmStealsUntapsAndGivesHasteAndMenace(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	theirs := effectstest.B12Creature(g, opp.ID, "Theirs", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() { g.Battlefield.Cards[0].Tapped = true })
	effectstest.CastCatalogSpell(t, g, "Lose Calm", "Sorcery", o334LoseCalm, o334Card(theirs))
	effectstest.PassPriorityAroundTable(t, g)
	var c game.Card
	for _, bc := range g.Battlefield.Cards {
		if bc.InstanceID == theirs {
			c = bc
		}
	}
	if c.Controller != me.ID {
		t.Fatal("I did not gain control of the creature")
	}
	if effectstest.IsTapped(g, theirs) {
		t.Error("the creature stayed tapped")
	}
	effectstest.AssertKeywords(t, g, theirs, "haste", "menace")
}

func TestO334RevealingWindPreventsAllCombatDamage(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	atk := effectstest.B12Creature(g, me.ID, "Attacker", "Creature — Bear", 3, 3)
	effectstest.DeclareAttack(t, g, opp.ID, atk)
	life := opp.Life
	if err := o334CastInHand(g, me, "Revealing Wind", "Instant", o334Wind, nil); err != nil {
		t.Fatalf("cast: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AdvanceTo(t, g, game.StepEndCombat)
	if opp.Life != life {
		t.Errorf("the defender lost %d life; all combat damage should have been prevented", life-opp.Life)
	}
}

func TestO334SkywiseTeachingsMakesADjinnMonkOnlyWhenPaid(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Skywise Teachings", "Enchantment", o334Teachings, false)
		effectstest.CastCatalogSpell(t, g, "Revealing Wind", "Instant", o334Wind, nil)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("no payment was offered for the noncreature spell")
		}
		effectstest.B06AddMana(me, "U", "U")
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		n := len(effectstest.BattlefieldIDsNamed(g, "Djinn Monk"))
		if pay && n != 1 || !pay && n != 0 {
			t.Errorf("paid=%v: %d Djinn Monk tokens", pay, n)
		}
	}
}

func TestO334SkywiseTeachingsIgnoresCreatureSpells(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	effectstest.PushCatalogPermanent(g, me.ID, "Skywise Teachings", "Enchantment", o334Teachings, false)
	effectstest.CastCatalogSpell(t, g, "Dragon Hunter", "Creature — Human Warrior", o334Hunter, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if effectstest.HasPayUnlessFor(g, me.ID) {
		t.Error("a creature spell offered the payment")
	}
}

func TestO334NecromasterDragonPaysTwoForAZombieAndTheMill(t *testing.T) {
	for _, pay := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		dragon := effectstest.B12Push(g, me.ID, "Necromaster Dragon", "Creature — Dragon", o334Necromaster, 4, 4)
		lib := opp.Library.Size()
		effectstest.DeclareAttack(t, g, opp.ID, dragon)
		effectstest.AdvanceTo(t, g, game.StepCombatDamage)
		effectstest.PassPriorityAroundTable(t, g)
		if !effectstest.HasPayUnlessFor(g, me.ID) {
			t.Fatal("no payment was offered after combat damage to a player")
		}
		effectstest.B06AddMana(me, "U", "B")
		effectstest.AnswerPayUnless(t, g, me.ID, pay)
		effectstest.PassPriorityAroundTable(t, g)
		zombies := len(effectstest.BattlefieldIDsNamed(g, "Zombie"))
		milled := lib - opp.Library.Size()
		if pay && (zombies != 1 || milled != 2) || !pay && (zombies != 0 || milled != 0) {
			t.Errorf("paid=%v: %d Zombies, %d cards milled", pay, zombies, milled)
		}
	}
}

func TestO334DescentDestroysEachTargetAndGivesItsControllerADragon(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := effectstest.B12Creature(g, me.ID, "Mine", "Creature — Bear", 2, 2)
	t1 := effectstest.B12Creature(g, opp.ID, "T1", "Creature — Bear", 2, 2)
	t2 := effectstest.B12Creature(g, opp.ID, "T2", "Creature — Bear", 2, 2)
	untouched := effectstest.B12Creature(g, opp.ID, "Untouched", "Creature — Bear", 2, 2)
	effectstest.CastCatalogSpell(t, g, "Descent of the Dragons", "Sorcery", o334Descent, effectstest.CardRefs(mine, t1, t2))
	effectstest.PassPriorityAroundTable(t, g)
	for _, id := range []uuid.UUID{mine, t1, t2} {
		if g.Battlefield.Contains(id) {
			t.Errorf("%v survived", id)
		}
	}
	if !g.Battlefield.Contains(untouched) {
		t.Error("an untargeted creature was destroyed")
	}
	dragons := effectstest.BattlefieldIDsNamed(g, "Dragon")
	byMe, byOpp := 0, 0
	for _, bc := range g.Battlefield.Cards {
		if bc.Name == "Dragon" && bc.Controller == me.ID {
			byMe++
		}
		if bc.Name == "Dragon" && bc.Controller == opp.ID {
			byOpp++
		}
	}
	if len(dragons) != 3 || byMe != 1 || byOpp != 2 {
		t.Errorf("%d Dragons (mine %d, theirs %d), want 3 (1 and 2)", len(dragons), byMe, byOpp)
	}
	if p := effectstest.EffectivePower(t, g, dragons[0]); p != 4 {
		t.Errorf("a Dragon token has power %d, want 4", p)
	}
}

func TestO334DescentMayBeCastWithNoTargets(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.CastCatalogSpell(t, g, "Descent of the Dragons", "Sorcery", o334Descent, nil)
	effectstest.PassPriorityAroundTable(t, g)
	if n := len(effectstest.BattlefieldIDsNamed(g, "Dragon")); n != 0 {
		t.Errorf("%d Dragons from a spell that destroyed nothing", n)
	}
}

func TestO334DragonHunterBlocksAFlyingDragonAndNoOtherFlier(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dragon := effectstest.BrCreature(g, me.ID, "Drake Dragon", "Creature — Dragon", 3, 3, "flying")
	bird := effectstest.BrCreature(g, me.ID, "Bird", "Creature — Bird", 1, 1, "flying")
	hunter := effectstest.B12Push(g, opp.ID, "Dragon Hunter", "Creature — Human Warrior", o334Hunter, 2, 1)
	effectstest.BrAttack(t, g, dragon, bird)
	if err := g.DeclareBlocker(hunter, dragon); err != nil {
		t.Errorf("Dragon Hunter could not block a flying Dragon: %v", err)
	}
	effectstest.BrRefusal(t, g.DeclareBlocker(hunter, bird), game.BlockReasonFlying)
}

func TestO334HeraldGivesOtherWarriorsVigilance(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	herald := effectstest.B12Push(g, me.ID, "Herald of Dromoka", "Creature — Human Warrior", o334Herald, 2, 2)
	warrior := effectstest.B12Creature(g, me.ID, "Warrior", "Creature — Human Warrior", 2, 2)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	theirs := effectstest.B12Creature(g, opp.ID, "Their Warrior", "Creature — Human Warrior", 2, 2)
	effectstest.AssertKeywords(t, g, herald, "vigilance")
	effectstest.AssertKeywords(t, g, warrior, "vigilance")
	if effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, bear), "vigilance") || effectstest.HasAbility(effectstest.EffectiveAbilities(t, g, theirs), "vigilance") {
		t.Error("vigilance reached a non-Warrior or an opponent's Warrior")
	}
}

func o334Card(id uuid.UUID) []game.TargetRef { return effectstest.CardRefs(id) }

func o334CastInHand(g *game.Game, p *game.Player, name, typeLine, oracle string, targets []game.TargetRef) error {
	id := effectstest.PutInHand(p, game.Card{Name: name, TypeLine: typeLine, OracleID: oracle})
	return g.CastSpell(p.ID, id, game.CastSpellParams{Targets: targets, Strict: true})
}

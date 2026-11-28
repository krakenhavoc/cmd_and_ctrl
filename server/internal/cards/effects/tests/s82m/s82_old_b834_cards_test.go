package testss82m

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	oldB834AegisOfTheMeek      = "0cf78877-bf08-413c-8ecb-8c3265ac1c4a"
	oldB834BattleFrenzy        = "0fca46ec-ec1a-4e6c-bb0e-cdb834b38c92"
	oldB834Rally               = "274f8860-c8c3-4aeb-ba87-94a5139397d8"
	oldB834Aurochs             = "3961ef7c-4eb4-482e-9cda-d49d6a29c5a9"
	oldB834KrovikanElementalst = "43e6fc44-8b23-4bff-91da-936019efe84b"
	oldB834Stampede            = "49c997d5-7b41-4faf-b898-58db19055671"
	oldB834ChaosMoon           = "728e2960-1d85-4b3f-8b5b-915f9b9ffd2f"
	oldB834WoollySpider        = "811c6f52-03cb-49a9-a3c7-7f4ae40e7155"
	oldB834MartonStromgald     = "8364cf93-c323-46f4-854f-ae872e484438"
	oldB834VibratingSphere     = "8a8fe94f-81aa-4725-ace6-ed56a1e1ed91"
	oldB834PhantasmalMount     = "8c7723a5-04c7-4241-b0ba-6a821449ca43"
	oldB834KnightOfStromgald   = "8cf57287-21b0-4194-8da3-62d00160234f"
)

var oldB834Cards = map[string]string{
	oldB834AegisOfTheMeek: "Aegis of the Meek", oldB834BattleFrenzy: "Battle Frenzy", oldB834Rally: "Rally",
	oldB834Aurochs: "Aurochs", oldB834KrovikanElementalst: "Krovikan Elementalist", oldB834Stampede: "Stampede",
	oldB834ChaosMoon: "Chaos Moon", oldB834WoollySpider: "Woolly Spider", oldB834MartonStromgald: "Márton Stromgald",
	oldB834VibratingSphere: "Vibrating Sphere", oldB834PhantasmalMount: "Phantasmal Mount",
	oldB834KnightOfStromgald: "Knight of Stromgald",
}

func TestOldB834EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range oldB834Cards {
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

func oldB834Creature(g *game.Game, owner uuid.UUID, name, typeLine, oracle string, power, toughness int, colors ...string) uuid.UUID {
	return effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle,
		Power: power, Toughness: toughness, Owner: owner, Controller: owner, Colors: colors,
	})
}

func oldB834PT(t *testing.T, g *game.Game, id uuid.UUID, wantP, wantT int, what string) {
	t.Helper()
	if p, tough := effectstest.PtOf(t, g, id); p != wantP || tough != wantT {
		t.Errorf("%s is %d/%d, want %d/%d", what, p, tough, wantP, wantT)
	}
}

// oldB834CastNow casts a spell from the hand of `who` at the current step.
func oldB834CastNow(t *testing.T, g *game.Game, who *game.Player, name, oracle, mana string) {
	t.Helper()
	id := uuid.New()
	who.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: "Instant", OracleID: oracle, Owner: who.ID, Controller: who.ID})
	effectstest.FloatMana(t, g, who, mana)
	if err := g.CastSpell(who.ID, id, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast %s: %v", name, err)
	}
	effectstest.PassPriorityAroundTable(t, g)
}

func TestOldB834AegisOfTheMeekNeedsAOneOne(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	aegis := effectstest.B12Push(g, me.ID, "Aegis of the Meek", "Artifact", oldB834AegisOfTheMeek, 0, 0)
	small := effectstest.B12Creature(g, me.ID, "Mite", "Creature — Insect", 1, 1)
	big := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{G}")
	err := g.ActivateCatalogAbility(me.ID, aegis, 0, game.ActivateAbilityParams{
		Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}},
	})
	if err == nil {
		t.Fatal("a 2/2 was a legal target")
	}
	effectstest.B16Activate(t, g, me.ID, aegis, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	oldB834PT(t, g, small, 2, 3, "the 1/1")
}

func TestOldB834BattleFrenzyPumpsGreenAndNongreenDifferently(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me, opp := g.Seats[g.Turn.ActiveSeat], g.Seats[(g.Turn.ActiveSeat+1)%4]
	green := oldB834Creature(g, me.ID, "Green Bear", "Creature — Bear", "", 2, 2, "G")
	red := oldB834Creature(g, me.ID, "Red Bear", "Creature — Bear", "", 2, 2, "R")
	theirs := oldB834Creature(g, opp.ID, "Their Bear", "Creature — Bear", "", 2, 2, "G")
	oldB834CastNow(t, g, me, "Battle Frenzy", oldB834BattleFrenzy, "{R}{R}{R}")
	oldB834PT(t, g, green, 3, 3, "the green creature")
	oldB834PT(t, g, red, 3, 2, "the nongreen creature")
	oldB834PT(t, g, theirs, 2, 2, "an opponent's creature")
}

func TestOldB834StampedePumpsOnlyAttackers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := oldB834Creature(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2, "G")
	idle := oldB834Creature(g, me.ID, "Idle", "Creature — Bear", "", 2, 2, "G")
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	oldB834CastNow(t, g, me, "Stampede", oldB834Stampede, "{G}{G}{G}")
	oldB834PT(t, g, attacker, 3, 2, "the attacker")
	oldB834PT(t, g, idle, 2, 2, "the idle creature")
	if !effectstest.HasEffectiveKeyword(t, g, attacker, "trample") {
		t.Error("the attacker has no trample")
	}
	if effectstest.HasEffectiveKeyword(t, g, idle, "trample") {
		t.Error("the idle creature has trample")
	}
}

func TestOldB834RallyPumpsOnlyBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	attacker := oldB834Creature(g, me.ID, "Attacker", "Creature — Bear", "", 2, 2, "G")
	blocker := oldB834Creature(g, opp.ID, "Blocker", "Creature — Bear", "", 2, 6, "W")
	idle := oldB834Creature(g, opp.ID, "Idle", "Creature — Bear", "", 2, 2, "W")
	effectstest.DeclareAttack(t, g, opp.ID, attacker)
	effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	effectstest.LockInBlocks(t, g)
	oldB834CastNow(t, g, opp, "Rally", oldB834Rally, "{W}{W}")
	oldB834PT(t, g, blocker, 3, 7, "the blocker")
	oldB834PT(t, g, idle, 2, 2, "the creature that did not block")
}

func TestOldB834AurochsCountsOtherAttackingAurochs(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	a := effectstest.TypedCreature(g, me.ID, "Aurochs", "Creature — Aurochs", oldB834Aurochs, 2, 3, "haste", "trample")
	b := effectstest.TypedCreature(g, me.ID, "Aurochs", "Creature — Aurochs", oldB834Aurochs, 2, 3, "haste", "trample")
	effectstest.AttackAndTrigger(t, g, a, b)
	effectstest.AnswerTriggerOrderInOfferedOrder(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	oldB834PT(t, g, a, 3, 3, "the first Aurochs")
	oldB834PT(t, g, b, 3, 3, "the second Aurochs")
}

func TestOldB834KrovikanElementalistPumpsAndGivesFlyingForASacrifice(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	effectstest.AdvanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	el := effectstest.B12Push(g, me.ID, "Krovikan Elementalist", "Creature — Human Wizard", oldB834KrovikanElementalst, 1, 1)
	bear := effectstest.B12Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	effectstest.FloatMana(t, g, me, "{R}{R}{R}")
	effectstest.B16Activate(t, g, me.ID, el, 0, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	oldB834PT(t, g, bear, 3, 2, "the pumped Bear")
	effectstest.FloatMana(t, g, me, "{U}{U}")
	effectstest.B16Activate(t, g, me.ID, el, 1, game.ActivateAbilityParams{
		Targets: []game.TargetRef{{Kind: game.TargetCard, ID: bear}},
	})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, bear, "flying") {
		t.Error("the Bear did not gain flying")
	}
	if !effects.OnBattlefield(g, bear) {
		t.Fatal("the Bear was sacrificed before the end step")
	}
	effectstest.AdvanceToStepInTurn(t, g, game.StepEnd)
	effectstest.PassPriorityAroundTable(t, g)
	if effects.OnBattlefield(g, bear) {
		t.Error("the Bear was not sacrificed at the beginning of the end step")
	}
}

func TestOldB834WoollySpiderGrowsOnlyBlockingAFlyer(t *testing.T) {
	for _, flyer := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me, opp := g.Seats[0], g.Seats[1]
		var attacker uuid.UUID
		if flyer {
			attacker = effectstest.Pr7bCreature(g, me.ID, "Bird", 1, "{1}", []string{"U"}, "flying")
		} else {
			attacker = effectstest.B16Creature(g, me.ID, "Bear", "Creature — Bear", 1, 1, "G")
		}
		spider := effectstest.TypedCreature(g, opp.ID, "Woolly Spider", "Creature — Spider", oldB834WoollySpider, 2, 3, "reach")
		effectstest.DeclareAttack(t, g, opp.ID, attacker)
		effectstest.AdvanceTo(t, g, game.StepDeclareBlockers)
		if err := g.DeclareBlocker(spider, attacker); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
		effectstest.LockInBlocks(t, g)
		effectstest.PassPriorityAroundTable(t, g)
		want := 3
		if flyer {
			want = 5
		}
		oldB834PT(t, g, spider, 2, want, "the Spider")
	}
}

func TestOldB834MartonPumpsTheOtherAttackersAndBlockers(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	marton := effectstest.B12Push(g, me.ID, "Márton Stromgald", "Legendary Creature — Human Knight", oldB834MartonStromgald, 1, 1)
	a := effectstest.B12Creature(g, me.ID, "A", "Creature — Bear", 2, 2)
	b := effectstest.B12Creature(g, me.ID, "B", "Creature — Bear", 2, 2)
	idle := effectstest.B12Creature(g, me.ID, "Idle", "Creature — Bear", 2, 2)
	effectstest.DeclareAttack(t, g, opp.ID, marton, a, b)
	effectstest.LockInAttacks(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	oldB834PT(t, g, marton, 1, 1, "Márton attacking")
	oldB834PT(t, g, a, 4, 4, "an attacker (two others attack)")
	oldB834PT(t, g, b, 4, 4, "the other attacker")
	oldB834PT(t, g, idle, 2, 2, "the idle creature")

	// Blocking: Márton and two other creatures block three attackers.
	g2 := effectstest.NewCatalogGame(t)
	me2, opp2 := g2.Seats[0], g2.Seats[1]
	m2 := effectstest.B12Push(g2, opp2.ID, "Márton Stromgald", "Legendary Creature — Human Knight", oldB834MartonStromgald, 1, 1)
	b1 := effectstest.B12Creature(g2, opp2.ID, "B1", "Creature — Bear", 2, 2)
	b2 := effectstest.B12Creature(g2, opp2.ID, "B2", "Creature — Bear", 2, 2)
	a1 := effectstest.B12Creature(g2, me2.ID, "A1", "Creature — Bear", 1, 1)
	a2 := effectstest.B12Creature(g2, me2.ID, "A2", "Creature — Bear", 1, 1)
	a3 := effectstest.B12Creature(g2, me2.ID, "A3", "Creature — Bear", 1, 1)
	effectstest.DeclareAttack(t, g2, opp2.ID, a1, a2, a3)
	effectstest.AdvanceTo(t, g2, game.StepDeclareBlockers)
	for _, pair := range [][2]uuid.UUID{{m2, a1}, {b1, a2}, {b2, a3}} {
		if err := g2.DeclareBlocker(pair[0], pair[1]); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	effectstest.LockInBlocks(t, g2)
	effectstest.PassPriorityAroundTable(t, g2)
	oldB834PT(t, g2, m2, 1, 1, "Márton blocking")
	oldB834PT(t, g2, b1, 4, 4, "a blocker (two others block)")
	oldB834PT(t, g2, a1, 1, 1, "an attacker")
}

func TestOldB834VibratingSphereFollowsTheTurn(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	active := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%4]
	effectstest.PushCatalogPermanent(g, active.ID, "Vibrating Sphere", "Artifact", oldB834VibratingSphere, false)
	mine := effectstest.B12Push(g, active.ID, "Bear", "Creature — Bear", "", 2, 3)
	theirs := effectstest.B12Push(g, other.ID, "Their Bear", "Creature — Bear", "", 2, 3)
	oldB834PT(t, g, mine, 4, 3, "my creature on my turn")
	oldB834PT(t, g, theirs, 2, 3, "an opponent's creature on my turn")

	g2 := effectstest.NewCatalogGame(t)
	act2 := g2.Seats[g2.Turn.ActiveSeat]
	oth2 := g2.Seats[(g2.Turn.ActiveSeat+1)%4]
	effectstest.PushCatalogPermanent(g2, oth2.ID, "Vibrating Sphere", "Artifact", oldB834VibratingSphere, false)
	m2 := effectstest.B12Push(g2, oth2.ID, "Bear", "Creature — Bear", "", 2, 3)
	o2 := effectstest.B12Push(g2, act2.ID, "Their Bear", "Creature — Bear", "", 2, 3)
	oldB834PT(t, g2, m2, 2, 1, "my creature on another player's turn")
	oldB834PT(t, g2, o2, 2, 3, "the active player's creature")
}

func TestOldB834PhantasmalMountSacrificesTheOtherWhenOneLeaves(t *testing.T) {
	for _, mountLeaves := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		effectstest.AdvanceToMain(t, g)
		me := g.Seats[g.Turn.ActiveSeat]
		mount := effectstest.TypedCreature(g, me.ID, "Phantasmal Mount", "Creature — Illusion Horse", oldB834PhantasmalMount, 1, 1, "flying")
		small := effectstest.B12Creature(g, me.ID, "Small", "Creature — Bear", 1, 2)
		big := effectstest.B12Creature(g, me.ID, "Big", "Creature — Bear", 3, 3)
		err := g.ActivateCatalogAbility(me.ID, mount, 0, game.ActivateAbilityParams{
			Strict: true, Targets: []game.TargetRef{{Kind: game.TargetCard, ID: big}},
		})
		if err == nil {
			t.Fatal("a creature with toughness 3 was a legal target")
		}
		effectstest.B16Activate(t, g, me.ID, mount, 0, game.ActivateAbilityParams{
			Targets: []game.TargetRef{{Kind: game.TargetCard, ID: small}},
		})
		effectstest.PassPriorityAroundTable(t, g)
		oldB834PT(t, g, small, 2, 3, "the target")
		if !effectstest.HasEffectiveKeyword(t, g, small, "flying") {
			t.Error("the target did not gain flying")
		}
		gone, stays := small, mount
		if mountLeaves {
			gone, stays = mount, small
		}
		g.WithWriteLock(func() {
			if err := g.DestroyPermanentForEffect(gone); err != nil {
				t.Fatalf("destroy: %v", err)
			}
		})
		g.RunStateChecksForTest()
		effectstest.PassPriorityAroundTable(t, g)
		if effects.OnBattlefield(g, stays) {
			t.Errorf("mount leaves %v: the other creature was not sacrificed", mountLeaves)
		}
	}
}

func TestOldB834KnightOfStromgaldAbilities(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	k := effectstest.TypedCreature(g, me.ID, "Knight of Stromgald", "Creature — Human Knight", oldB834KnightOfStromgald, 2, 1, "protection from white")
	effectstest.FloatMana(t, g, me, "{B}{B}{B}")
	effectstest.B16Activate(t, g, me.ID, k, 0, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.B16Activate(t, g, me.ID, k, 1, game.ActivateAbilityParams{})
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.HasEffectiveKeyword(t, g, k, "first strike") {
		t.Error("no first strike")
	}
	oldB834PT(t, g, k, 3, 1, "the Knight")
}

func TestOldB834ChaosMoonOddAndEven(t *testing.T) {
	for _, odd := range []bool{true, false} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		effectstest.PushCatalogPermanent(g, me.ID, "Chaos Moon", "Enchantment", oldB834ChaosMoon, false)
		red := oldB834Creature(g, me.ID, "Red Bear", "Creature — Bear", "", 2, 2, "R")
		mountain := effectstest.Basics(g, me, "Mountain", 1)[0]
		if !odd {
			// Chaos Moon, the red creature and the Mountain are three permanents.
			oldB834Creature(g, me.ID, "Filler", "Creature — Bear", "", 1, 1, "G")
		}
		effectstest.AdvanceToStepInTurn(t, g, game.StepUpkeep)
		effectstest.PassPriorityAroundTable(t, g)
		if g.Turn.Step != game.StepUpkeep {
			// The next turn's upkeep, whoever's it is.
			effectstest.AdvanceTo(t, g, game.StepUpkeep)
			effectstest.PassPriorityAroundTable(t, g)
		}
		if odd {
			oldB834PT(t, g, red, 3, 3, "the red creature on an odd count")
		} else {
			oldB834PT(t, g, red, 1, 1, "the red creature on an even count")
		}
		before := effectstest.PoolSize(g, me.ID)
		if err := g.ActivateManaAbility(me.ID, mountain, 0, game.ManaAbilityParams{}); err != nil {
			t.Fatalf("tap Mountain: %v", err)
		}
		got := effectstest.PoolSize(g, me.ID) - before
		if odd && got != 2 {
			t.Errorf("odd: a Mountain made %d mana, want 2 ({R} and an additional {R})", got)
		}
		if !odd && got != 1 {
			t.Errorf("even: a Mountain made %d mana, want 1 (colorless)", got)
		}
		if !odd {
			for _, tok := range g.Seats[0].ManaPool {
				if tok.Color == "R" {
					t.Error("even: the Mountain still made red mana")
				}
			}
		}
	}
}

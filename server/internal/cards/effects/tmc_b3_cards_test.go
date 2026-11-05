package effects

import (
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmc_b3_cards_test.go — Teenage Mutant Ninja Turtles Commander, batch 3
// (#5011). Each test plays the real catalog entry and asserts what its text
// says.

const (
	tmcB3SpecialMoveOracle = "7564afbb-b29e-4234-a3b2-3354c6f5c90a"
	tmcB3TempestraOracle   = "12401ad1-b5c9-43dc-9389-04aaf5db8b26"
	tmcB3TokkaOracle       = "8d857d79-bdb2-4ca3-abd6-7db8f6b3bcfa"
	tmcB3HighScoreOracle   = "d81fc181-ecb0-43a5-88e8-c61aca3428d8"
	tmcB3LeatherheadOracle = "1e9679e8-bef1-495c-8251-0d87d971e32a"
	tmcB3LevelUpOracle     = "88cd43ce-b168-45a4-8d29-6cac5ab42857"
	tmcB3DoubleJumpOracle  = "cb3072c7-393c-43d9-a2c9-f918bd0fd7be"
	tmcB3ArcadeOracle      = "a8e55287-f0df-46f2-a41d-27bea90216cc"
	tmcB3BarrelOracle      = "82800275-1edf-4aa0-af9e-59dea82d7c6a"
	tmcB3MoleModuleOracle  = "ba64d3a9-08b4-4b12-b011-e41b00815277"
)

func TestTmcB3EveryCardIsRegisteredFull(t *testing.T) {
	for oracle, name := range map[string]string{
		tmcB3SpecialMoveOracle:       "Special Move",
		tmcB3TempestraOracle:         "Tempestra, Dame of Games",
		tmcB3TokkaOracle:             "Tokka & Rahzar, Unsupervised",
		tmcB3HighScoreOracle:         "High Score",
		tmcB3LeatherheadOracle:       "Leatherhead, Iron Gator",
		tmcB3LevelUpOracle:           "Level Up",
		tmcB3DoubleJumpOracle:        "Double Jump",
		tmcB3DoubleJumpOracle + "#1": "Flying Kick",
		tmcB3ArcadeOracle:            "Arcade Cabinet",
		tmcB3BarrelOracle:            "Exploding Barrel",
		tmcB3MoleModuleOracle:        "Mole Module",
	} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", name)
			continue
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// tmcB3CastModal casts a modal spell from the active seat's hand, strictly,
// with the mana floated first, and returns the announce error.
func tmcB3CastModal(t *testing.T, g *game.Game, name, typeLine, oracle, mana string, modes []int, targets ...game.TargetRef) error {
	t.Helper()
	me := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: id, Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: mana, Owner: me.ID, Controller: me.ID})
	floatMana(t, g, me, strings.NewReplacer("{1}", "{C}", "{2}", "{C}{C}").Replace(mana))
	return g.CastSpell(me.ID, id, game.CastSpellParams{Modes: modes, Targets: targets, Strict: true})
}

// tmcB3ModeSlot is one target announced for clause `slot` of mode
// occurrence `occ`.
func tmcB3ModeSlot(id uuid.UUID, occ, slot int) game.TargetRef {
	return game.TargetRef{Kind: game.TargetCard, ID: id, Mode: occ, Slot: slot}
}

// --- Special Move -----------------------------------------------------------

func TestTmcB3SpecialMoveJumpKickAndFootToss(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	thrower := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	victim := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)

	if err := tmcB3CastModal(t, g, "Special Move", "Instant", tmcB3SpecialMoveOracle, "{2}{R}", []int{0, 2},
		p1g3ModeCard(rock, 0), tmcB3ModeSlot(thrower, 1, 0), tmcB3ModeSlot(victim, 1, 1)); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(rock) {
		t.Error("Jump Kick did not destroy the artifact")
	}
	if got := tmtB1Card(t, g, victim).DamageMarked; got != 3 {
		t.Errorf("Foot Toss dealt %d, want the thrower's power 3", got)
	}
	if g.Battlefield.Contains(thrower) {
		t.Error("the thrower was not sacrificed")
	}
}

func TestTmcB3SpecialMoveFootTossNeverThrowsAtItselfAndMustChooseTwo(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	thrower := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 3, 3)
	if err := tmcB3CastModal(t, g, "Special Move", "Instant", tmcB3SpecialMoveOracle, "{2}{R}", []int{0, 2},
		p1g3ModeCard(rock, 0), tmcB3ModeSlot(thrower, 1, 0), tmcB3ModeSlot(thrower, 1, 1)); err == nil {
		t.Error("Foot Toss's \"any other target\" accepted the thrower itself")
	}
	if err := tmcB3CastModal(t, g, "Special Move", "Instant", tmcB3SpecialMoveOracle, "{2}{R}", []int{0},
		p1g3ModeCard(rock, 0)); err == nil {
		t.Error("a single mode was accepted for \"choose two\"")
	}
}

func TestTmcB3SpecialMoveDashAttackNeedsAnAttackingOrBlockingCreature(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	if err := tmcB3CastModal(t, g, "Special Move", "Instant", tmcB3SpecialMoveOracle, "{2}{R}", []int{0, 1},
		p1g3ModeCard(rock, 0), p1g3ModeCard(bear, 1)); err == nil {
		t.Fatal("Dash Attack accepted a creature that is neither attacking nor blocking")
	}

	declareAttack(t, g, opp.ID, bear)
	if err := tmcB3CastModal(t, g, "Special Move", "Instant", tmcB3SpecialMoveOracle, "{2}{R}", []int{0, 1},
		p1g3ModeCard(rock, 0), p1g3ModeCard(bear, 1)); err != nil {
		t.Fatalf("cast during combat: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, bear); got != 2 {
		t.Errorf("the attacking Bear has %d +1/+1 counters, want 2", got)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("Jump Kick did not destroy the artifact")
	}
}

// --- Tempestra, Dame of Games -----------------------------------------------

func TestTmcB3TempestraMakesANonlegendaryHastyCopyAndSacrificesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	tempestra := b27Push(g, me.ID, "Tempestra, Dame of Games", "Legendary Creature — Elemental Illusion", tmcB3TempestraOracle, "{1}{R}", 1, 3, "R")
	hero := b12Creature(g, me.ID, "Big Hero", "Legendary Creature — Human Hero", 4, 4)
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")

	floatMana(t, g, me, "{C}{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, tempestra, 0, game.ActivateAbilityParams{
		Targets: cardRefs(tempestra), SacrificeIDs: []uuid.UUID{rock}, Strict: true}); err == nil {
		t.Fatal("Tempestra copied herself; the target must be another creature")
	}
	if err := g.ActivateCatalogAbility(me.ID, tempestra, 0, game.ActivateAbilityParams{
		Targets: cardRefs(hero), SacrificeIDs: []uuid.UUID{rock}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if g.Battlefield.Contains(rock) {
		t.Error("the artifact was not sacrificed as a cost")
	}
	passPriorityAroundTable(t, g)
	var token uuid.UUID
	for _, id := range tmtB1Named(g, "Big Hero") {
		if id != hero {
			token = id
		}
	}
	if token == uuid.Nil {
		t.Fatal("no token copy of the Hero (the legend rule took it, or it was never made)")
	}
	c := tmtB1Card(t, g, token)
	if !c.IsToken() || c.IsLegendary() || !effectiveAbilitiesContain(t, g, token, "haste") {
		t.Errorf("the copy: token %v, legendary %v, haste %v; want a nonlegendary token with haste",
			c.IsToken(), c.IsLegendary(), effectiveAbilitiesContain(t, g, token, "haste"))
	}
	if !g.Battlefield.Contains(hero) {
		t.Error("the original legend left the battlefield")
	}
	advanceTo(t, g, game.StepEnd)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(token) {
		t.Error("the token was not sacrificed at the beginning of the end step")
	}
}

func TestTmcB3TempestraNeedsAnArtifactToSacrifice(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	tempestra := b27Push(g, me.ID, "Tempestra, Dame of Games", "Legendary Creature — Elemental Illusion", tmcB3TempestraOracle, "{1}{R}", 1, 3, "R")
	hero := b12Creature(g, me.ID, "Big Hero", "Creature — Human Hero", 4, 4)
	floatMana(t, g, me, "{C}{C}{R}")
	if err := g.ActivateCatalogAbility(me.ID, tempestra, 0, game.ActivateAbilityParams{
		Targets: cardRefs(hero), SacrificeIDs: []uuid.UUID{hero}, Strict: true}); err == nil {
		t.Error("a creature that is not an artifact paid \"Sacrifice an artifact\"")
	}
}

// --- Tokka & Rahzar, Unsupervised -------------------------------------------

func TestTmcB3TokkaAndRahzarGrowsAndMakesATreasureOnceEachTurn(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	tr := b27Push(g, me.ID, "Tokka & Rahzar, Unsupervised", "Legendary Creature — Turtle Wolf Mutant", tmcB3TokkaOracle, "{2}{R}", 3, 3, "R")
	assertKeywords(t, g, tr, "first strike")
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	tmtB3MakeToken(g, me.ID, "2/2 red Mutant")
	mutant := tmtB1Named(g, "Mutant")[0]
	bear1 := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	bear2 := b12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)

	b27Kill(g, theirs)
	passPriorityAroundTable(t, g)
	b27Kill(g, mutant)
	passPriorityAroundTable(t, g)
	if plusOnes(g, tr) != 0 || tmtB1Count(g, me.ID, "Treasure") != 0 {
		t.Fatal("an opponent's creature or a token leaving triggered Tokka & Rahzar")
	}
	b27Kill(g, bear1)
	passPriorityAroundTable(t, g)
	if plusOnes(g, tr) != 1 || tmtB1Count(g, me.ID, "Treasure") != 1 {
		t.Fatalf("after my nontoken creature died: %d counters, %d Treasures; want 1 and 1",
			plusOnes(g, tr), tmtB1Count(g, me.ID, "Treasure"))
	}
	g.WithWriteLock(func() { _ = g.BounceToHandForEffect(bear2) })
	passPriorityAroundTable(t, g)
	if plusOnes(g, tr) != 1 || tmtB1Count(g, me.ID, "Treasure") != 1 {
		t.Errorf("the ability triggered a second time this turn: %d counters, %d Treasures",
			plusOnes(g, tr), tmtB1Count(g, me.ID, "Treasure"))
	}
}

// --- High Score -------------------------------------------------------------

func TestTmcB3HighScoreAddsOneCounterOnYourCreaturesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "High Score", "Enchantment", tmcB3HighScoreOracle, "{2}{G}", 0, 0, "G")
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(mine, game.CounterPlusOne, 1)
		_ = g.AddCounterForEffect(theirs, game.CounterPlusOne, 1)
	})
	if plusOnes(g, mine) != 2 || plusOnes(g, theirs) != 1 {
		t.Errorf("counters: mine %d, theirs %d; want 2 and 1", plusOnes(g, mine), plusOnes(g, theirs))
	}
}

func TestTmcB3HighScoreDrawsAtYourEndStepWithTheGreatestPower(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "High Score", "Enchantment", tmcB3HighScoreOracle, "{2}{G}", 0, 0, "G")
	b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 4, 4)
	b12Creature(g, opp.ID, "Their Ogre", "Creature — Ogre", 4, 4)
	hand := me.Hand.Size()
	advanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("tied for the greatest power: hand %d, want %d", got, hand+1)
	}
}

func TestTmcB3HighScoreDoesNotDrawWhenOutsized(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b27Push(g, me.ID, "High Score", "Enchantment", tmcB3HighScoreOracle, "{2}{G}", 0, 0, "G")
	b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 4, 4)
	b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 5, 5)
	hand := me.Hand.Size()
	advanceToStepOf(t, g, 0, game.StepEnd)
	passPriorityAroundTable(t, g)
	if got := me.Hand.Size(); got != hand {
		t.Errorf("outsized: hand %d, want %d", got, hand)
	}
}

// --- Leatherhead, Iron Gator ------------------------------------------------

func TestTmcB3LeatherheadPutsTwoCountersOnEachCreatureYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	lh := b27Push(g, me.ID, "Leatherhead, Iron Gator", "Legendary Creature — Crocodile Mutant Rogue", tmcB3LeatherheadOracle, "{5}{G}{G}", 5, 5, "G")
	assertKeywords(t, g, lh, "trample", "haste")
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, opp.ID, lh)
	passPriorityAroundTable(t, g)
	if plusOnes(g, lh) != 2 || plusOnes(g, mine) != 2 || plusOnes(g, theirs) != 0 {
		t.Errorf("counters: Leatherhead %d, mine %d, theirs %d; want 2, 2, 0", plusOnes(g, lh), plusOnes(g, mine), plusOnes(g, theirs))
	}
}

// --- Level Up ---------------------------------------------------------------

// tmcB3LevelUp casts Level Up on `host` and resolves it and its enters
// trigger.
func tmcB3LevelUp(t *testing.T, g *game.Game, host uuid.UUID) uuid.UUID {
	t.Helper()
	aura := castCatalogSpell(t, g, "Level Up", "Enchantment — Aura", tmcB3LevelUpOracle, cardRefs(host))
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	return aura
}

func TestTmcB3LevelUpCountersThenDoublesOnAttack(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 5, 5)
	tmcB3LevelUp(t, g, host)
	if got := plusOnes(g, host); got != 1 {
		t.Fatalf("enters trigger: %d counters, want 1", got)
	}
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, host)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, host); got != 2 {
		t.Errorf("after attacking: %d counters, want 2", got)
	}
	if got := me.Hand.Size(); got != hand {
		t.Errorf("power 7 drew: hand %d, want %d", got, hand)
	}
}

func TestTmcB3LevelUpDrawsAtPowerTen(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	host := b12Creature(g, me.ID, "My Giant", "Creature — Giant", 8, 8)
	tmcB3LevelUp(t, g, host)
	hand := me.Hand.Size()
	declareAttack(t, g, opp.ID, host)
	passPriorityAroundTable(t, g)
	if got := plusOnes(g, host); got != 2 {
		t.Errorf("after attacking: %d counters, want 2", got)
	}
	if got := me.Hand.Size(); got != hand+1 {
		t.Errorf("power 10: hand %d, want %d", got, hand+1)
	}
}

// --- Double Jump // Flying Kick ---------------------------------------------

func tmcB3DoubleJumpCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   tmcB3DoubleJumpOracle,
		Layout:     game.LayoutSplit,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "Double Jump", TypeLine: "Instant", ManaCost: "{1}{U}",
				OracleText: "Put a flying counter on target creature you control. Until end of turn, it has base power and toughness 5/5.\nFuse (You may cast one or both halves of this card from your hand.)"},
			{Name: "Flying Kick", TypeLine: "Instant", ManaCost: "{1}{R}",
				OracleText: "Target creature you control deals damage equal to its power to target creature an opponent controls.\nFuse (You may cast one or both halves of this card from your hand.)"},
		},
	}
	c.SettleImported()
	return c
}

func TestTmcB3DoubleJumpGivesFlyingAndBaseFiveFiveUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	c := tmcB3DoubleJumpCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("cast Double Jump: %v", err)
	}
	passPriorityAroundTable(t, g)
	if countersOn(g, bear, game.CounterFlying) != 1 || !effectiveAbilitiesContain(t, g, bear, "flying") {
		t.Error("no flying counter")
	}
	if p, tt := effectivePower(t, g, bear), effectiveToughness(t, g, bear); p != 5 || tt != 5 {
		t.Errorf("Bear is %d/%d, want 5/5", p, tt)
	}
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	if p := effectivePower(t, g, bear); p != 2 || !effectiveAbilitiesContain(t, g, bear, "flying") {
		t.Errorf("next turn: power %d, flying %v; want 2 and still flying", p, effectiveAbilitiesContain(t, g, bear, "flying"))
	}
}

func TestTmcB3DoubleJumpFlyingKickFusedBitesForFive(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)
	c := tmcB3DoubleJumpCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{
		Fuse: true,
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: bear, Slot: 0},
			{Kind: game.TargetCard, ID: bear, Slot: 1},
			{Kind: game.TargetCard, ID: wall, Slot: 2},
		},
	}); err != nil {
		t.Fatalf("cast fused: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := tmtB1Card(t, g, wall).DamageMarked; got != 5 {
		t.Errorf("Flying Kick dealt %d, want the new base power 5", got)
	}
}

func TestTmcB3FlyingKickRefusesAnOpponentsCreatureAsTheKicker(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 3, 3)
	wall := b12Creature(g, opp.ID, "Their Wall", "Creature — Wall", 0, 9)
	c := tmcB3DoubleJumpCard(me.ID)
	me.Hand.PushTop(c)
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: wall, Slot: 0}, {Kind: game.TargetCard, ID: bear, Slot: 1},
	}}); err == nil {
		t.Fatal("Flying Kick accepted the clauses the wrong way round")
	}
	if err := g.CastSpell(me.ID, c.InstanceID, game.CastSpellParams{Face: 1, Targets: []game.TargetRef{
		{Kind: game.TargetCard, ID: bear, Slot: 0}, {Kind: game.TargetCard, ID: wall, Slot: 1},
	}}); err != nil {
		t.Fatalf("cast Flying Kick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := tmtB1Card(t, g, wall).DamageMarked; got != 3 {
		t.Errorf("Flying Kick dealt %d, want 3", got)
	}
}

// --- Arcade Cabinet ---------------------------------------------------------

func TestTmcB3ArcadeCabinetCountersUpToFourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	castCatalogSpell(t, g, "Arcade Cabinet", "Artifact", tmcB3ArcadeOracle, nil)
	passPriorityAroundTable(t, g)
	pickTargetRefs(t, g, me.ID, nil, cardRefs(a, b)...)
	passPriorityAroundTable(t, g)
	if plusOnes(g, a) != 1 || plusOnes(g, b) != 1 {
		t.Errorf("counters: %d and %d, want 1 and 1", plusOnes(g, a), plusOnes(g, b))
	}
}

func TestTmcB3ArcadeCabinetSacrificesATokenToDoubleEachKindOfCounter(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	advanceToMain(t, g)
	cab := b27Push(g, me.ID, "Arcade Cabinet", "Artifact", tmcB3ArcadeOracle, "{3}", 0, 0)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.AddCounterForEffect(bear, game.CounterPlusOne, 2)
		_ = g.AddCounterForEffect(bear, game.CounterFlying, 1)
	})
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	floatMana(t, g, me, "{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, cab, 0, game.ActivateAbilityParams{
		Targets: cardRefs(bear), SacrificeIDs: []uuid.UUID{rock}, Strict: true}); err == nil {
		t.Fatal("a nontoken artifact paid \"Sacrifice a token\"")
	}
	tmtB3MakeToken(g, me.ID, "2/2 red Mutant")
	mutant := tmtB1Named(g, "Mutant")[0]
	if err := g.ActivateCatalogAbility(me.ID, cab, 0, game.ActivateAbilityParams{
		Targets: cardRefs(bear), SacrificeIDs: []uuid.UUID{mutant}, Strict: true}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	passPriorityAroundTable(t, g)
	if plusOnes(g, bear) != 4 || countersOn(g, bear, game.CounterFlying) != 2 {
		t.Errorf("counters: +1/+1 %d, flying %d; want 4 and 2", plusOnes(g, bear), countersOn(g, bear, game.CounterFlying))
	}
}

// --- Exploding Barrel -------------------------------------------------------

func TestTmcB3ExplodingBarrelTapsForAnyColorAndGainsPressure(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	barrel := b27Push(g, me.ID, "Exploding Barrel", "Artifact", tmcB3BarrelOracle, "{3}", 0, 0)
	tapForMana(t, g, me.ID, barrel)
	if n := b10ResolveAllManaPicks(t, g, me.ID, "B"); n != 1 {
		t.Errorf("%d colour picks, want 1", n)
	}
	if got := poolColors(me); len(got) != 1 || got[0] != "B" {
		t.Errorf("pool %v, want [B]", got)
	}
	if got := countersOn(g, barrel, "pressure"); got != 1 {
		t.Errorf("%d pressure counters, want 1", got)
	}
}

func TestTmcB3ExplodingBarrelCostsOneLessPerPressureCounter(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	barrel := b27Push(g, me.ID, "Exploding Barrel", "Artifact", tmcB3BarrelOracle, "{3}", 0, 0)
	giant := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 10, 30)
	g.WithWriteLock(func() { _ = g.AddCounterForEffect(barrel, "pressure", 3) })
	floatMana(t, g, me, "{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, barrel, 0, game.ActivateAbilityParams{Targets: cardRefs(giant), Strict: true}); err == nil {
		t.Fatal("{4} paid an {8} cost reduced by only three counters")
	}
	floatMana(t, g, me, "{C}")
	if err := g.ActivateCatalogAbility(me.ID, barrel, 0, game.ActivateAbilityParams{Targets: cardRefs(giant), Strict: true}); err != nil {
		t.Fatalf("activate for {5}: %v", err)
	}
	if g.Battlefield.Contains(barrel) {
		t.Error("the Barrel was not sacrificed")
	}
	passPriorityAroundTable(t, g)
	if got := tmtB1Card(t, g, giant).DamageMarked; got != 20 {
		t.Errorf("damage %d, want 20", got)
	}
}

func TestTmcB3ExplodingBarrelIsSorcerySpeed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	barrel := b27Push(g, me.ID, "Exploding Barrel", "Artifact", tmcB3BarrelOracle, "{3}", 0, 0)
	giant := b12Creature(g, opp.ID, "Their Giant", "Creature — Giant", 10, 30)
	advanceTo(t, g, game.StepUpkeep)
	floatMana(t, g, me, "{C}{C}{C}{C}{C}{C}{C}{C}")
	if err := g.ActivateCatalogAbility(me.ID, barrel, 0, game.ActivateAbilityParams{Targets: cardRefs(giant), Strict: true}); err == nil {
		t.Error("the Barrel was activated outside a main phase")
	}
}

// --- Mole Module ------------------------------------------------------------

func TestTmcB3MoleModuleMillsFourAndPutsAPermanentOntoTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mole := pushVehicleForTest(g, me.ID, "Mole Module", tmcB3MoleModuleOracle, 6, 6)
	assertKeywords(t, g, mole, "menace")
	crew := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 2, 2)
	var milled []uuid.UUID
	for _, c := range []game.Card{
		{Name: "Bolt", TypeLine: "Instant"},
		{Name: "Dug-Up Golem", TypeLine: "Artifact Creature — Golem", Power: 3, Toughness: 3},
		{Name: "Shock", TypeLine: "Instant"},
		{Name: "Forest", TypeLine: "Basic Land — Forest"},
	} {
		c.InstanceID = uuid.New()
		c.Owner, c.Controller = me.ID, me.ID
		me.Library.PushTop(c)
		milled = append(milled, c.InstanceID)
	}
	if err := g.ActivateCatalogAbility(me.ID, mole, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, mole)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("no permanent-card pick after the mill")
	}
	if hasID(pick.ChooseCards, milled[0]) || hasID(pick.ChooseCards, milled[2]) {
		t.Error("an instant was offered as a permanent card")
	}
	if !hasID(pick.ChooseCards, milled[1]) || !hasID(pick.ChooseCards, milled[3]) {
		t.Errorf("offered %v, want the Golem and the Forest", pick.ChooseCards)
	}
	answerChooseCards(t, g, me.ID, milled[1])
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(milled[1]) {
		t.Error("the Golem did not enter the battlefield")
	}
	for _, id := range []uuid.UUID{milled[0], milled[2], milled[3]} {
		if !me.Graveyard.Contains(id) {
			t.Errorf("milled card %s is not in the graveyard", id)
		}
	}
}

func TestTmcB3MoleModuleMayDecline(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	advanceToMain(t, g)
	mole := pushVehicleForTest(g, me.ID, "Mole Module", tmcB3MoleModuleOracle, 6, 6)
	crew := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 2, 2)
	golem := uuid.New()
	me.Library.PushTop(game.Card{InstanceID: golem, Name: "Dug-Up Golem", TypeLine: "Artifact Creature — Golem",
		Power: 3, Toughness: 3, Owner: me.ID, Controller: me.ID})
	if err := g.ActivateCatalogAbility(me.ID, mole, 0, game.ActivateAbilityParams{CrewIDs: []uuid.UUID{crew}}); err != nil {
		t.Fatalf("crew: %v", err)
	}
	passPriorityAroundTable(t, g)
	declareAttack(t, g, opp.ID, mole)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(golem) || !me.Graveyard.Contains(golem) {
		t.Error("declining still put the Golem onto the battlefield")
	}
}

func TestTmcB3SpecialMoveFootTossCanHitAPlayer(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	rock := b12Permanent(g, opp.ID, "Their Rock", "Artifact")
	thrower := b12Creature(g, me.ID, "My Ogre", "Creature — Ogre", 4, 4)
	life := opp.Life
	if err := tmcB3CastModal(t, g, "Special Move", "Instant", tmcB3SpecialMoveOracle, "{2}{R}", []int{0, 2},
		p1g3ModeCard(rock, 0), tmcB3ModeSlot(thrower, 1, 0),
		game.TargetRef{Kind: game.TargetPlayer, ID: opp.ID, Mode: 1, Slot: 1}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life-4 {
		t.Errorf("opponent life %d, want %d", opp.Life, life-4)
	}
	if g.Battlefield.Contains(thrower) {
		t.Error("the thrower was not sacrificed")
	}
}

func TestTmcB3TokkaAndRahzarTriggersOnceForASimultaneousDeparture(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tr := b27Push(g, me.ID, "Tokka & Rahzar, Unsupervised", "Legendary Creature — Turtle Wolf Mutant", tmcB3TokkaOracle, "{2}{R}", 3, 3, "R")
	bear1 := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	bear2 := b12Creature(g, me.ID, "My Other Bear", "Creature — Bear", 2, 2)
	g.WithWriteLock(func() {
		_ = g.DestroyPermanentForEffect(bear1)
		_ = g.DestroyPermanentForEffect(bear2)
	})
	passPriorityAroundTable(t, g)
	if plusOnes(g, tr) != 1 || tmtB1Count(g, me.ID, "Treasure") != 1 {
		t.Errorf("two creatures left at once: %d counters, %d Treasures; want 1 and 1",
			plusOnes(g, tr), tmtB1Count(g, me.ID, "Treasure"))
	}
}

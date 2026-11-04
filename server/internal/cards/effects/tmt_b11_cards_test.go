package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tmt_b11_cards_test.go — Teenage Mutant Ninja Turtles, batch 11 (#4457):
// the sneak cards (CR 702.190). Each test plays the real catalog entry,
// most of them cast for their sneak cost in the declare blockers step.

const (
	tmtB11DarkLeoOracle     = "c3ea6af5-48b9-45d2-9576-46cacb2db5b0"
	tmtB11DonTechOracle     = "194284f7-cf92-4d9f-a8ce-e4dc3505ab7a"
	tmtB11DonGadgetOracle   = "bb6aef0d-052b-44e6-b4cd-2d2eafd84969"
	tmtB11FootNinjasOracle  = "306652a8-9a9c-4b33-96d7-b59234d148e9"
	tmtB11JennikaTechOracle = "90b1226f-4c90-47fd-98f3-ad4edd3b3518"
	tmtB11KaraiTechOracle   = "e53fb128-bf90-4af6-b998-ef9c250001cb"
	tmtB11KaraiOracle       = "7246638e-0361-40e3-8019-3c10210a593e"
	tmtB11KitsuneTechOracle = "3e0f8ab3-75a6-49a2-a228-fa720cbde3d8"
	tmtB11LeoTechOracle     = "e98c2af9-1393-487d-b8e9-f49a67bb7296"
	tmtB11LeoBigOracle      = "6496eea7-38f5-4428-a379-b08b0eeb5239"
	tmtB11LeoCuttingOracle  = "933fe6e5-8d00-4411-aa54-7780382c1ea6"
	tmtB11LeoBlueOracle     = "348c7e22-2179-41b8-a5d2-8dd488e7e66d"
	tmtB11MikeyTechOracle   = "42f12942-9ac1-497e-a8b7-9f66a5afd0ae"
	tmtB11TurtleType        = "Legendary Creature — Mutant Ninja Turtle"
	tmtB11SorceryType       = "Sorcery"
	tmtB11InstantType       = "Instant"
	tmtB11FootNinjasType    = "Creature — Human Ninja"
	tmtB11KaraiType         = "Legendary Creature — Human Ninja"
	tmtB11DarkLeoType       = "Legendary Creature — Mutant Ninja Turtle Human"
)

func TestTmtB11EveryCardIsRegisteredFullWithItsSneakCost(t *testing.T) {
	for _, tc := range []struct{ oracle, name, sneak string }{
		{tmtB11DarkLeoOracle, "Dark Leo & Shredder", "{W}{B}"},
		{tmtB11DonTechOracle, "Donatello's Technique", "{U}"},
		{tmtB11DonGadgetOracle, "Donatello, Gadget Master", "{1}{U}"},
		{tmtB11FootNinjasOracle, "Foot Ninjas", "{3}{W/B}"},
		{tmtB11JennikaTechOracle, "Jennika's Technique", "{R}"},
		{tmtB11KaraiTechOracle, "Karai's Technique", "{W}{B}"},
		{tmtB11KaraiOracle, "Karai, Future of the Foot", "{2}{W}{B}"},
		{tmtB11KitsuneTechOracle, "Kitsune's Technique", "{1}{U}"},
		{tmtB11LeoTechOracle, "Leonardo's Technique", "{1}{W}"},
		{tmtB11LeoBigOracle, "Leonardo, Big Brother", "{W}"},
		{tmtB11LeoCuttingOracle, "Leonardo, Cutting Edge", "{W}"},
		{tmtB11LeoBlueOracle, "Leonardo, Leader in Blue", "{3}{W}{W}"},
		{tmtB11MikeyTechOracle, "Michelangelo's Technique", "{3}{G}"},
	} {
		spec, ok := Lookup(tc.oracle)
		if !ok {
			t.Errorf("%s is not registered", tc.name)
			continue
		}
		if spec.Name != tc.name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", tc.name, spec.Name, spec.Completeness, spec.Caveats)
		}
		alt := game.AlternativeCostByKey(tc.oracle, "sneak")
		if alt == nil || !alt.Sneak || alt.ManaCost != tc.sneak || alt.ReturnToHand == nil {
			t.Errorf("%s: sneak offer %+v, want Sneak %s", tc.name, alt, tc.sneak)
		}
	}
}

// --- shared setup --------------------------------------------------------

// tmtB11Attack declares my 1/1 attacking seat 1, lets every defender
// finish blocking with no blocks, and stops in MY declare blockers step:
// the sneak window (CR 702.190a). `extra` attack too.
func tmtB11Attack(t *testing.T, g *game.Game, extra ...uuid.UUID) uuid.UUID {
	t.Helper()
	me, opp := g.Seats[0], g.Seats[1]
	attacker := b12Creature(g, me.ID, "Sneaky Rat", "Creature — Rat", 1, 1)
	declareAttack(t, g, opp.ID, append([]uuid.UUID{attacker}, extra...)...)
	advanceTo(t, g, game.StepDeclareBlockers)
	finishBlockDeclarations(t, g)
	if g.Turn.Step != game.StepDeclareBlockers {
		t.Fatalf("left the declare blockers step (at %s)", g.Turn.Step)
	}
	return attacker
}

// tmtB11Hand puts a catalog card into p's hand.
func tmtB11Hand(g *game.Game, p *game.Player, name, typeLine, oracle, cost string, power, toughness int) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{
			InstanceID: id, Name: name, TypeLine: typeLine, ManaCost: cost, OracleID: oracle,
			Power: power, Toughness: toughness, Owner: p.ID, Controller: p.ID,
			KnownBy: map[uuid.UUID]bool{p.ID: true},
		})
	})
	return id
}

// tmtB11Sneak floats `mana` and casts `card` for its sneak cost,
// returning `returned`; params carries any targets or modes.
func tmtB11Sneak(t *testing.T, g *game.Game, card, returned uuid.UUID, mana string, params game.CastSpellParams) {
	t.Helper()
	me := g.Seats[0]
	floatMana(t, g, me, mana)
	params.AlternativeCost = "sneak"
	params.AltCostIDs = []uuid.UUID{returned}
	if err := g.CastSpell(me.ID, card, params); err != nil {
		t.Fatalf("sneak cast: %v", err)
	}
	if !me.Hand.Contains(returned) {
		t.Fatal("the unblocked attacker was not returned to hand as the cost")
	}
}

// tmtB11AssertSneakedIn checks CR 702.190b: tapped and attacking seat 1.
func tmtB11AssertSneakedIn(t *testing.T, g *game.Game, id uuid.UUID) {
	t.Helper()
	c := tmtB1Card(t, g, id)
	if !c.Tapped || c.AttackingTarget != g.Seats[1].ID {
		t.Fatalf("tapped=%v attacking %s, want tapped and attacking %s", c.Tapped, c.AttackingTarget, g.Seats[1].ID)
	}
}

// --- Dark Leo & Shredder -------------------------------------------------

func TestTmtB11DarkLeoGivesAttackingNinjasDeathtouch(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	b27Push(g, me.ID, "Dark Leo & Shredder", tmtB11DarkLeoType, tmtB11DarkLeoOracle, "{W}{B}", 1, 3, "W", "B")
	ninja := b12Creature(g, me.ID, "Attacking Ninja", "Creature — Human Ninja", 2, 2)
	home := b12Creature(g, me.ID, "Home Ninja", "Creature — Human Ninja", 2, 2)
	bear := b12Creature(g, me.ID, "Attacking Bear", "Creature — Bear", 2, 2)
	declareAttack(t, g, g.Seats[1].ID, ninja, bear)
	if !effectiveAbilitiesContain(t, g, ninja, "deathtouch") {
		t.Error("an attacking Ninja has no deathtouch")
	}
	if effectiveAbilitiesContain(t, g, home, "deathtouch") {
		t.Error("a Ninja that is not attacking has deathtouch")
	}
	if effectiveAbilitiesContain(t, g, bear, "deathtouch") {
		t.Error("an attacking non-Ninja has deathtouch")
	}
}

func TestTmtB11DarkLeoSneaksInAndMakesANinjaOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leo := tmtB11Hand(g, me, "Dark Leo & Shredder", tmtB11DarkLeoType, tmtB11DarkLeoOracle, "{W}{B}", 1, 3)
	rat := tmtB11Attack(t, g)
	tmtB11Sneak(t, g, leo, rat, "{W}{B}", game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, nil, true)
	tmtB11AssertSneakedIn(t, g, leo)
	life := opp.Life
	advanceTo(t, g, game.StepCombatDamage)
	assistCardSettle(t, g, me.ID, nil, true)
	if n := tmtB1Count(g, me.ID, "Ninja"); n != 1 {
		t.Errorf("%d Ninja tokens, want 1", n)
	}
	// Two Ninjas (Leo and the token): no halving, just the 1 damage.
	if opp.Life != life-1 {
		t.Errorf("defender life %d, want %d", opp.Life, life-1)
	}
}

func TestTmtB11DarkLeoHalvesTheDamagedPlayersLifeWithFiveNinjas(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leo := b27Push(g, me.ID, "Dark Leo & Shredder", tmtB11DarkLeoType, tmtB11DarkLeoOracle, "{W}{B}", 1, 3, "W", "B")
	for i := 0; i < 3; i++ {
		b12Creature(g, me.ID, "Home Ninja", "Creature — Human Ninja", 1, 1)
	}
	declareAttack(t, g, opp.ID, leo)
	advanceTo(t, g, game.StepCombatDamage)
	assistCardSettle(t, g, me.ID, nil, true)
	// 40 - 1 combat damage = 39; Leo, three Ninjas and the new token
	// make five, so 39 loses 20 (rounded up).
	if opp.Life != 19 {
		t.Errorf("defender life %d, want 19", opp.Life)
	}
}

// --- Donatello's Technique -----------------------------------------------

func TestTmtB11DonatellosTechniqueSneaksASorceryInCombatAndDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tech := tmtB11Hand(g, me, "Donatello's Technique", tmtB11SorceryType, tmtB11DonTechOracle, "{2}{U}", 0, 0)
	rat := tmtB11Attack(t, g)
	hand := me.Hand.Size()
	tmtB11Sneak(t, g, tech, rat, "{U}", game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, nil, true)
	// The Technique left, the rat came back, two were drawn.
	if got := me.Hand.Size(); got != hand-1+1+2 {
		t.Errorf("hand %d, want %d", got, hand+2)
	}
	if !me.Graveyard.Contains(tech) {
		t.Error("the Technique did not resolve to the graveyard")
	}
}

func TestTmtB11DonatellosTechniqueIsStillASorceryForItsPrintedCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tech := tmtB11Hand(g, me, "Donatello's Technique", tmtB11SorceryType, tmtB11DonTechOracle, "{2}{U}", 0, 0)
	tmtB11Attack(t, g)
	floatMana(t, g, me, "{C}{C}{U}")
	if err := g.CastSpell(me.ID, tech, game.CastSpellParams{}); !errors.Is(err, game.ErrSorcerySpeedRequired) {
		t.Errorf("printed-cost cast in combat: %v, want ErrSorcerySpeedRequired", err)
	}
}

// --- Donatello, Gadget Master --------------------------------------------

func TestTmtB11DonatelloGadgetMasterCopiesAnArtifactOnCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	rock := b12Permanent(g, me.ID, "My Rock", "Artifact")
	don := tmtB11Hand(g, me, "Donatello, Gadget Master", tmtB11TurtleType, tmtB11DonGadgetOracle, "{2}{U}", 3, 2)
	rat := tmtB11Attack(t, g)
	tmtB11Sneak(t, g, don, rat, "{C}{U}", game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, nil, true)
	tmtB11AssertSneakedIn(t, g, don)
	advanceTo(t, g, game.StepCombatDamage)
	pick := cardRef(rock)
	assistCardSettle(t, g, me.ID, &pick, true)
	if n := tmtB1Count(g, me.ID, "My Rock"); n != 2 {
		t.Errorf("%d My Rock on the battlefield, want the original and a token copy", n)
	}
}

// --- Foot Ninjas ---------------------------------------------------------

func TestTmtB11FootNinjasSneaksInForAHybridCostAndGainsThree(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ninjas := tmtB11Hand(g, me, "Foot Ninjas", tmtB11FootNinjasType, tmtB11FootNinjasOracle, "{4}{W/B}{W/B}", 5, 5)
	rat := tmtB11Attack(t, g)
	life := me.Life
	tmtB11Sneak(t, g, ninjas, rat, "{B}{B}{B}{B}", game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, nil, true)
	tmtB11AssertSneakedIn(t, g, ninjas)
	if me.Life != life+3 {
		t.Errorf("life %d, want %d", me.Life, life+3)
	}
}

// --- Jennika's Technique -------------------------------------------------

func TestTmtB11JennikasTechniqueDealsTwoToEachCreature(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	small := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	big := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 3)
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	tech := tmtB11Hand(g, me, "Jennika's Technique", tmtB11InstantType, tmtB11JennikaTechOracle, "{2}{R}", 0, 0)
	rat := tmtB11Attack(t, g)
	tmtB11Sneak(t, g, tech, rat, "{R}", game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, nil, true)
	if g.Battlefield.Contains(small) || g.Battlefield.Contains(mine) {
		t.Error("a 2/2 survived 2 damage")
	}
	if !g.Battlefield.Contains(big) {
		t.Error("the 3/3 died")
	}
	if !me.Hand.Contains(rat) {
		t.Error("the returned attacker should be safe in hand")
	}
}

// --- Karai's Technique ---------------------------------------------------

func TestTmtB11KaraisTechniquePumpsOneAndShrinksAnother(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Ox", "Creature — Ox", 3, 3)
	tech := tmtB11Hand(g, me, "Karai's Technique", tmtB11SorceryType, tmtB11KaraiTechOracle, "{1}{W}{B}", 0, 0)
	rat := tmtB11Attack(t, g)
	tmtB11Sneak(t, g, tech, rat, "{W}{B}", game.CastSpellParams{
		Modes: []int{0, 1}, Targets: []game.TargetRef{escCardRef(mine, 0), escCardRef(theirs, 1)},
	})
	assistCardSettle(t, g, me.ID, nil, true)
	if p := effectivePower(t, g, mine); p != 5 {
		t.Errorf("+3/+3 on a 2/2: power %d, want 5", p)
	}
	if g.Battlefield.Contains(theirs) {
		t.Error("-3/-3 should kill a 3/3")
	}
}

func TestTmtB11KaraisTechniqueOneModeOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	tech := tmtB11Hand(g, me, "Karai's Technique", tmtB11SorceryType, tmtB11KaraiTechOracle, "{1}{W}{B}", 0, 0)
	rat := tmtB11Attack(t, g)
	tmtB11Sneak(t, g, tech, rat, "{W}{B}", game.CastSpellParams{
		Modes: []int{1}, Targets: []game.TargetRef{escCardRef(mine, 0)},
	})
	assistCardSettle(t, g, me.ID, nil, true)
	if g.Battlefield.Contains(mine) {
		t.Error("-3/-3 alone should kill a 2/2")
	}
}

// --- Karai, Future of the Foot -------------------------------------------

func TestTmtB11KaraiSneakedInReturnsACreatureToTheBattlefield(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	dead := gyCard(me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	karai := tmtB11Hand(g, me, "Karai, Future of the Foot", tmtB11KaraiType, tmtB11KaraiOracle, "{1}{W}{B}", 3, 3)
	rat := tmtB11Attack(t, g)
	tmtB11Sneak(t, g, karai, rat, "{C}{C}{W}{B}", game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, nil, true)
	tmtB11AssertSneakedIn(t, g, karai)
	advanceTo(t, g, game.StepCombatDamage)
	pick := cardRef(dead)
	assistCardSettle(t, g, me.ID, &pick, true)
	if !g.Battlefield.Contains(dead) {
		t.Error("her sneak cost was paid this turn, but the creature card did not return to the battlefield")
	}
}

func TestTmtB11KaraiCastNormallyReturnsACreatureToHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := gyCard(me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	karai := b27Push(g, me.ID, "Karai, Future of the Foot", tmtB11KaraiType, tmtB11KaraiOracle, "{1}{W}{B}", 3, 3, "W", "B")
	declareAttack(t, g, opp.ID, karai)
	advanceTo(t, g, game.StepCombatDamage)
	pick := cardRef(dead)
	assistCardSettle(t, g, me.ID, &pick, true)
	if !me.Hand.Contains(dead) || g.Battlefield.Contains(dead) {
		t.Error("without the sneak cost, the creature card should go to hand")
	}
}

// A Karai sneaked in on an earlier turn still has a sneak cast record, but
// it was not paid THIS turn.
func TestTmtB11KaraiSneakedOnAnEarlierTurnReturnsToHand(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	dead := gyCard(me, "Dead Bear", "Creature — Bear", "{1}{G}", "")
	karai := pushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: "Karai, Future of the Foot", TypeLine: tmtB11KaraiType,
		ManaCost: "{1}{W}{B}", OracleID: tmtB11KaraiOracle, Power: 3, Toughness: 3,
		Owner: me.ID, Controller: me.ID, Provenance: game.CastProvenance{AltCost: "sneak"},
	})
	declareAttack(t, g, opp.ID, karai)
	advanceTo(t, g, game.StepCombatDamage)
	pick := cardRef(dead)
	assistCardSettle(t, g, me.ID, &pick, true)
	if !me.Hand.Contains(dead) {
		t.Error("a sneak paid on an earlier turn should not put the card onto the battlefield")
	}
}

// --- Kitsune's Technique -------------------------------------------------

func TestTmtB11KitsunesTechniqueMillsHalfRoundedUp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	g.WithWriteLock(func() {
		opp.Library.Cards = nil
		for i := 0; i < 5; i++ {
			opp.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Creature — Bear", Owner: opp.ID, Controller: opp.ID})
		}
	})
	tech := tmtB11Hand(g, me, "Kitsune's Technique", tmtB11InstantType, tmtB11KitsuneTechOracle, "{4}{U}{U}", 0, 0)
	rat := tmtB11Attack(t, g)
	gy := opp.Graveyard.Size()
	tmtB11Sneak(t, g, tech, rat, "{C}{U}", game.CastSpellParams{
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}},
	})
	assistCardSettle(t, g, me.ID, nil, true)
	if got := opp.Library.Size(); got != 2 {
		t.Errorf("library %d, want 2 (half of 5, rounded up, is 3)", got)
	}
	if got := opp.Graveyard.Size(); got != gy+3 {
		t.Errorf("graveyard %d, want %d", got, gy+3)
	}
}

func TestTmtB11KitsunesTechniqueTargetsOnlyAnOpponent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	tech := tmtB11Hand(g, me, "Kitsune's Technique", tmtB11InstantType, tmtB11KitsuneTechOracle, "{4}{U}{U}", 0, 0)
	rat := tmtB11Attack(t, g)
	floatMana(t, g, me, "{C}{U}")
	if err := g.CastSpell(me.ID, tech, game.CastSpellParams{
		AlternativeCost: "sneak", AltCostIDs: []uuid.UUID{rat},
		Targets: []game.TargetRef{{Kind: game.TargetPlayer, ID: me.ID}},
	}); err == nil {
		t.Error("Kitsune's Technique targeted its own caster")
	}
}

// --- Leonardo's Technique ------------------------------------------------

func TestTmtB11LeonardosTechniqueReturnsTwoSmallCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	a := gyCard(me, "Bear A", "Creature — Bear", "{1}{G}", "")
	b := gyCard(me, "Ox B", "Creature — Ox", "{2}{G}", "")
	tech := tmtB11Hand(g, me, "Leonardo's Technique", tmtB11SorceryType, tmtB11LeoTechOracle, "{3}{W}", 0, 0)
	rat := tmtB11Attack(t, g)
	tmtB11Sneak(t, g, tech, rat, "{C}{W}", game.CastSpellParams{Targets: cardRefs(a, b)})
	assistCardSettle(t, g, me.ID, nil, true)
	if !g.Battlefield.Contains(a) || !g.Battlefield.Contains(b) {
		t.Error("both creature cards should return to the battlefield")
	}
}

func TestTmtB11LeonardosTechniqueRefusesManaValueFour(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	big := gyCard(me, "Big Beast", "Creature — Beast", "{3}{G}", "")
	tech := tmtB11Hand(g, me, "Leonardo's Technique", tmtB11SorceryType, tmtB11LeoTechOracle, "{3}{W}", 0, 0)
	rat := tmtB11Attack(t, g)
	floatMana(t, g, me, "{C}{W}")
	if err := g.CastSpell(me.ID, tech, game.CastSpellParams{
		AlternativeCost: "sneak", AltCostIDs: []uuid.UUID{rat}, Targets: cardRefs(big),
	}); err == nil {
		t.Error("a mana value 4 creature card was accepted as a target")
	}
}

// --- Leonardo, Big Brother -----------------------------------------------

func TestTmtB11LeonardoBigBrotherCountsOtherCreaturesYouControl(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	leo := b27Push(g, me.ID, "Leonardo, Big Brother", tmtB11TurtleType, tmtB11LeoBigOracle, "{2}{W}", 1, 3, "W")
	if p := effectivePower(t, g, leo); p != 1 {
		t.Errorf("alone: power %d, want 1", p)
	}
	b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	b12Creature(g, me.ID, "My Ox", "Creature — Ox", 3, 3)
	b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	b12Permanent(g, me.ID, "My Rock", "Artifact")
	if p, tough := effectivePower(t, g, leo), effectiveToughness(t, g, leo); p != 3 || tough != 3 {
		t.Errorf("with two other creatures: %d/%d, want 3/3", p, tough)
	}
}

func TestTmtB11LeonardoBigBrotherSneaksIn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	leo := tmtB11Hand(g, me, "Leonardo, Big Brother", tmtB11TurtleType, tmtB11LeoBigOracle, "{2}{W}", 1, 3)
	rat := tmtB11Attack(t, g)
	tmtB11Sneak(t, g, leo, rat, "{W}", game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, nil, true)
	tmtB11AssertSneakedIn(t, g, leo)
}

// --- Leonardo, Cutting Edge ----------------------------------------------

func TestTmtB11LeonardoCuttingEdgeGrowsWhenYouGainLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	leo := tmtB11Hand(g, me, "Leonardo, Cutting Edge", tmtB11TurtleType, tmtB11LeoCuttingOracle, "{1}{W}", 1, 1)
	rat := tmtB11Attack(t, g)
	tmtB11Sneak(t, g, leo, rat, "{W}", game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, nil, true)
	tmtB11AssertSneakedIn(t, g, leo)
	if !effectiveAbilitiesContain(t, g, leo, "lifelink") {
		t.Error("Leonardo has no lifelink")
	}
	life := me.Life
	advanceTo(t, g, game.StepCombatDamage)
	assistCardSettle(t, g, me.ID, nil, true)
	if me.Life != life+1 {
		t.Errorf("life %d, want %d from lifelink", me.Life, life+1)
	}
	if n := plusOnes(g, leo); n != 1 {
		t.Errorf("%d +1/+1 counters, want 1 from the life gain", n)
	}
}

// --- Leonardo, Leader in Blue --------------------------------------------

func TestTmtB11LeonardoLeaderInBlueSneakedPumpsYourCreatures(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	theirs := b12Creature(g, opp.ID, "Their Bear", "Creature — Bear", 2, 2)
	leo := tmtB11Hand(g, me, "Leonardo, Leader in Blue", tmtB11TurtleType, tmtB11LeoBlueOracle, "{W}", 2, 1)
	rat := tmtB11Attack(t, g)
	tmtB11Sneak(t, g, leo, rat, "{C}{C}{C}{W}{W}", game.CastSpellParams{})
	assistCardSettle(t, g, me.ID, nil, true)
	tmtB11AssertSneakedIn(t, g, leo)
	if p := effectivePower(t, g, mine); p != 4 {
		t.Errorf("my bear power %d, want 4", p)
	}
	if p := effectivePower(t, g, leo); p != 4 {
		t.Errorf("Leonardo power %d, want 4", p)
	}
	if p := effectivePower(t, g, theirs); p != 2 {
		t.Errorf("their bear power %d, want 2", p)
	}
}

func TestTmtB11LeonardoLeaderInBlueCastNormallyPumpsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mine := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	advanceToMain(t, g)
	leo := tmtB11Hand(g, me, "Leonardo, Leader in Blue", tmtB11TurtleType, tmtB11LeoBlueOracle, "{W}", 2, 1)
	floatMana(t, g, me, "{W}")
	if err := g.CastSpell(me.ID, leo, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast for the printed cost: %v", err)
	}
	assistCardSettle(t, g, me.ID, nil, true)
	if !g.Battlefield.Contains(leo) {
		t.Fatal("Leonardo did not resolve")
	}
	if p := effectivePower(t, g, mine); p != 2 {
		t.Errorf("my bear power %d, want 2: his sneak cost was not paid", p)
	}
}

func TestTmtB11LeonardoLeaderInBlueGainsFirstStrike(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	leo := b27Push(g, me.ID, "Leonardo, Leader in Blue", tmtB11TurtleType, tmtB11LeoBlueOracle, "{W}", 2, 1, "W")
	advanceToMain(t, g)
	floatMana(t, g, me, "{C}{W}")
	if err := g.ActivateCatalogAbility(me.ID, leo, 0, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("ActivateAbility: %v", err)
	}
	assistCardSettle(t, g, me.ID, nil, true)
	if !effectiveAbilitiesContain(t, g, leo, "first strike") {
		t.Error("Leonardo did not gain first strike")
	}
}

// --- Michelangelo's Technique --------------------------------------------

func TestTmtB11MichelangelosTechniquePutsUpToTwoCreaturesTotalSixOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// Eight cards, top first once looked at.
	one := plTop(me, "One Drop", "Creature — Bear", "{1}")
	two := plTop(me, "Two Drop", "Creature — Bear", "{2}")
	three := plTop(me, "Three Drop", "Creature — Bear", "{3}")
	four := plTop(me, "Four Drop", "Creature — Bear", "{4}")
	five := plTop(me, "Five Drop", "Creature — Bear", "{5}")
	rock := plTop(me, "Rock", "Artifact", "{1}")
	plTop(me, "Wastes", "Basic Land — Wastes", "")
	plTop(me, "Six Drop", "Creature — Bear", "{6}")
	libBefore := me.Library.Size()

	tech := tmtB11Hand(g, me, "Michelangelo's Technique", tmtB11SorceryType, tmtB11MikeyTechOracle, "{4}{G}", 0, 0)
	rat := tmtB11Attack(t, g)
	tmtB11Sneak(t, g, tech, rat, "{C}{C}{C}{G}", game.CastSpellParams{})
	passPriorityAroundTable(t, g)

	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no pick of creature cards")
	}
	for _, id := range pick.ChooseCards {
		if id == rock {
			t.Error("a noncreature card was offered")
		}
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{two, five}); !errors.Is(err, game.ErrChoiceSetRejected) {
		t.Fatalf("2 + 5 = 7 is over budget: got %v, want ErrChoiceSetRejected", err)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{one, two, three}); err == nil {
		t.Fatal("three creature cards were accepted for \"up to two\"")
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{two, four}); err != nil {
		t.Fatalf("2 + 4 = 6: %v", err)
	}
	for _, id := range []uuid.UUID{two, four} {
		if !g.Battlefield.Contains(id) {
			t.Errorf("%s did not enter", id)
		}
	}
	if got := me.Library.Size(); got != libBefore-2 {
		t.Errorf("library %d, want %d", got, libBefore-2)
	}
	bottom := plBottomIDs(me, 6)
	for _, id := range []uuid.UUID{one, three, five, rock} {
		if !bottom[id] {
			t.Errorf("%s is not among the bottom six", id)
		}
	}
}

package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// zinnia_valleys_voice_test.go — Zinnia, Valley's Voice and the granted
// offspring seam (#5428).

const (
	zinniaOracle     = "d8c16a3a-dd9d-410c-a5d5-06339a89970f"
	zinniaTypeLine   = "Legendary Creature — Bird Bard"
	zinniaBearName   = "Zinnia Test Bear"
	zinniaTwinOracle = "test-zinnia-second-offspring-grant-5428"
)

func TestZinniaIsRegisteredAndWhole(t *testing.T) {
	spec, ok := Lookup(zinniaOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Completeness = %v, Caveats = %v", spec.Completeness, spec.Caveats)
	}
}

// zinniaOnBattlefield walks the active seat to its main phase with a
// Zinnia it controls.
func zinniaOnBattlefield(t *testing.T, g *game.Game) (me *game.Player, zinnia uuid.UUID) {
	t.Helper()
	me = g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	zinnia = pushCatalogPermanent(g, me.ID, "Zinnia, Valley's Voice", zinniaTypeLine, zinniaOracle, false)
	return me, zinnia
}

// zinniaBearInHand puts a vanilla 2/2 creature card costing {G}{G} in
// `p`'s hand.
func zinniaBearInHand(g *game.Game, p *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		p.Hand.PushTop(game.Card{InstanceID: id, Name: zinniaBearName, TypeLine: "Creature — Bear", ManaCost: "{G}{G}",
			Power: 2, Toughness: 2, Colors: []string{"G"}, Owner: p.ID, Controller: p.ID})
	})
	return id
}

// zinniaSettle passes priority until the stack is empty, answering a
// trigger-order prompt in the order offered.
func zinniaSettle(t *testing.T, g *game.Game, me *game.Player) {
	t.Helper()
	for i := 0; i < 24; i++ {
		if triggerOrderPrompt(g) != nil {
			answerTriggerOrderInOfferedOrder(t, g)
			continue
		}
		if stackFullyEmpty(g) && len(g.PendingChoices) == 0 {
			return
		}
		passPriorityAroundTable(t, g)
	}
	t.Fatal("the stack did not settle")
}

// zinniaBears are the test bears on the battlefield: the tokens and the
// cards.
func zinniaBears(g *game.Game) (tokens, cards []game.Card) {
	return copiesOnBattlefield(g, zinniaBearName)
}

func TestZinniaGetsPlusOnePowerPerOtherCreatureYouControlWithBasePowerOne(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zinnia := pushDiesCreatureForTest(g, me.ID, "Zinnia, Valley's Voice", zinniaOracle, zinniaTypeLine, 1, 3)
	pushVanillaCreature(g, me.ID, "One A", 1, 1)
	pumped := pushVanillaCreature(g, me.ID, "One With A Counter", 1, 1)
	shrunk := pushVanillaCreature(g, me.ID, "Zero With A Counter", 0, 1)
	pushVanillaCreature(g, me.ID, "Two", 2, 2)
	pushVanillaCreature(g, opp.ID, "Their One", 1, 1)
	g.WithWriteLock(func() {
		for _, id := range []uuid.UUID{pumped, shrunk} {
			if err := g.AddCounterForEffect(id, "+1/+1", 1); err != nil {
				t.Fatalf("AddCounterForEffect: %v", err)
			}
		}
	})
	g.RunStateChecksForTest()
	c, _ := battlefieldCard(g, zinnia)
	// Base power 1: "One A" and "One With A Counter" (a counter is not
	// base power). Not the 0/1 a counter made a 1/2, the 2/2, the
	// opponent's 1/1, or Zinnia herself.
	if e := c.Effective(); e.Power != 3 || e.Toughness != 3 {
		t.Errorf("Zinnia is %d/%d, want 3/3", e.Power, e.Toughness)
	}
}

func TestZinniaGivesACreatureSpellOffspringAndItMakesAOneOneCopy(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := zinniaOnBattlefield(t, g)
	bear := zinniaBearInHand(g, me)
	floatMana(t, g, me, "{G}{G}{G}{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{OptionalCosts: []int{0}, Strict: true}); err != nil {
		t.Fatalf("cast with offspring: %v", err)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left, want 0: offspring {2} is paid on top of {1}{G}", n)
	}
	zinniaSettle(t, g, me)
	if !g.Battlefield.Contains(bear) {
		t.Fatal("the bear did not resolve")
	}
	tokens, cards := zinniaBears(g)
	if len(tokens) != 1 || len(cards) != 1 {
		t.Fatalf("%d tokens and %d cards, want 1 and 1 (the token is not cast, so it makes none of its own)", len(tokens), len(cards))
	}
	if e := tokens[0].Effective(); e.Power != 1 || e.Toughness != 1 {
		t.Errorf("the token is %d/%d, want 1/1", e.Power, e.Toughness)
	}
	if tokens[0].Controller != me.ID {
		t.Error("the token is not under its creator's control")
	}
}

func TestZinniaOffspringNeedsItsManaAndIsNotForcedOnTheCaster(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := zinniaOnBattlefield(t, g)
	bear := zinniaBearInHand(g, me)
	floatMana(t, g, me, "{G}{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{OptionalCosts: []int{0}, Strict: true}); err == nil {
		t.Fatal("offspring {2} was paid from a pool holding only {1}{G}")
	}
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("cast without offspring: %v", err)
	}
	zinniaSettle(t, g, me)
	if tokens, cards := zinniaBears(g); len(tokens) != 0 || len(cards) != 1 {
		t.Errorf("%d tokens and %d cards, want 0 and 1: the offspring cost was not paid", len(tokens), len(cards))
	}
}

func TestZinniaGivesOffspringOnlyToCreatureSpellsItsControllerCasts(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	bear := zinniaBearInHand(g, me)
	floatMana(t, g, me, "{G}{G}{G}{G}")
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{OptionalCosts: []int{0}, Strict: true}); err == nil {
		t.Fatal("offspring was paid with no Zinnia")
	}
	// An opponent's Zinnia gives this player's spells nothing.
	pushCatalogPermanent(g, g.Seats[1].ID, "Zinnia, Valley's Voice", zinniaTypeLine, zinniaOracle, false)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{OptionalCosts: []int{0}, Strict: true}); err == nil {
		t.Fatal("offspring was paid under an opponent's Zinnia")
	}
	pushCatalogPermanent(g, me.ID, "Zinnia, Valley's Voice", zinniaTypeLine, zinniaOracle, false)
	bolt := uuid.New()
	g.WithWriteLock(func() {
		me.Hand.PushTop(game.Card{InstanceID: bolt, Name: "Lightning Bolt", TypeLine: "Instant", ManaCost: "{R}",
			OracleID: lightningBoltOracle, Owner: me.ID, Controller: me.ID})
	})
	floatMana(t, g, me, "{R}")
	if err := g.CastSpell(me.ID, bolt, game.CastSpellParams{
		Targets:       []game.TargetRef{{Kind: game.TargetPlayer, ID: g.Seats[1].ID}},
		OptionalCosts: []int{0}, Strict: true,
	}); err == nil {
		t.Error("an instant spell was given offspring")
	}
}

// The spell gained offspring as it was cast (CR 610.5), so the permanent
// it becomes makes its token though Zinnia has left the battlefield.
func TestZinniaOffspringPaysOffAfterZinniaLeaves(t *testing.T) {
	g := newCatalogGame(t)
	me, zinnia := zinniaOnBattlefield(t, g)
	bear := zinniaBearInHand(g, me)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast with offspring: %v", err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneGraveyard, Owner: me.ID}, zinnia); err != nil {
		t.Fatalf("move Zinnia: %v", err)
	}
	zinniaSettle(t, g, me)
	if tokens, cards := zinniaBears(g); len(tokens) != 1 || len(cards) != 1 {
		t.Errorf("%d tokens and %d cards, want 1 and 1", len(tokens), len(cards))
	}
}

// CR 702.175b: "If a spell has multiple instances of offspring, each is
// paid separately and triggers based on the payments made for it, not any
// other instances of offspring." A creature that prints offspring has its
// own cost (index 0) and Zinnia's (index 1).
func TestZinniaOffspringIsPaidBesideAPrintedOffspring(t *testing.T) {
	for _, tc := range []struct {
		name     string
		optional []int
		tokens   int
	}{
		{"both", []int{0, 1}, 2},
		{"printed only", []int{0}, 1},
		{"granted only", []int{1}, 1},
		{"neither", nil, 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, _ := zinniaOnBattlefield(t, g)
			rabbit := uuid.New()
			g.WithWriteLock(func() {
				me.Hand.PushTop(game.Card{InstanceID: rabbit, Name: "Intrepid Rabbit", TypeLine: blbGLRabbitTypeLine,
					OracleID: blbGLRabbitOracle, ManaCost: "{3}{W}", Power: 3, Toughness: 2, Owner: me.ID, Controller: me.ID})
			})
			if err := g.CastSpell(me.ID, rabbit, game.CastSpellParams{OptionalCosts: tc.optional}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			for i := 0; i < 12; i++ {
				blbGLSettleTriggerOrder(g, me.ID)
				if p := latestPickTarget(g, me.ID); p != nil {
					answerPickTarget(t, g, rabbit)
					continue
				}
				if stackFullyEmpty(g) && len(g.PendingChoices) == 0 {
					break
				}
				passPriorityAroundTable(t, g)
			}
			tokens, cards := copiesOnBattlefield(g, "Intrepid Rabbit")
			if len(tokens) != tc.tokens || len(cards) != 1 {
				t.Errorf("%d tokens and %d cards, want %d and 1", len(tokens), len(cards), tc.tokens)
			}
		})
	}
}

// Two permanents granting offspring give the spell two instances (CR
// 702.175b), each paid by itself.
func TestTwoOffspringGrantsAreTwoInstances(t *testing.T) {
	registerForTest(t, Spec{
		OracleID: zinniaTwinOracle, Name: "Second Voice", Completeness: CompletenessFull,
		GrantedOptionalCosts: []game.GrantedOptionalCost{
			OffspringForSpellsYouCast("Creature spells you cast gain offspring {2} as you cast them.", Creature(), "{2}"),
		},
	})
	for _, tc := range []struct {
		name     string
		optional []int
		mana     string
		tokens   int
	}{
		{"both", []int{0, 1}, "{G}{G}{G}{G}{G}{G}", 2},
		{"one", []int{1}, "{G}{G}{G}{G}", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, _ := zinniaOnBattlefield(t, g)
			pushCatalogPermanent(g, me.ID, "Second Voice", "Creature — Bird", zinniaTwinOracle, false)
			bear := zinniaBearInHand(g, me)
			floatMana(t, g, me, tc.mana)
			if err := g.CastSpell(me.ID, bear, game.CastSpellParams{OptionalCosts: tc.optional, Strict: true}); err != nil {
				t.Fatalf("cast: %v", err)
			}
			if n := me.ManaPool.Total(); n != 0 {
				t.Errorf("%d mana left, want 0", n)
			}
			zinniaSettle(t, g, me)
			if tokens, cards := zinniaBears(g); len(tokens) != tc.tokens || len(cards) != 1 {
				t.Errorf("%d tokens and %d cards, want %d and 1", len(tokens), len(cards), tc.tokens)
			}
		})
	}
	// One instance is paid once: naming it twice is refused.
	g := newCatalogGame(t)
	me, _ := zinniaOnBattlefield(t, g)
	bear := zinniaBearInHand(g, me)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{OptionalCosts: []int{0, 0}}); err == nil {
		t.Error("one offspring instance was paid twice")
	}
}

// CR 707.10: a copy of a spell copies the additional costs paid for it,
// so the token a copy of the creature spell becomes has the offspring
// trigger too.
func TestACopyOfASpellWithGrantedOffspringPaidMakesItsOwnToken(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := zinniaOnBattlefield(t, g)
	bear := zinniaBearInHand(g, me)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast with offspring: %v", err)
	}
	copyTheSpell(t, g, me, bear)
	zinniaSettle(t, g, me)
	tokens, cards := zinniaBears(g)
	if len(tokens) != 3 || len(cards) != 1 {
		t.Fatalf("%d tokens and %d cards, want 3 and 1: the copy's token, and an offspring token each", len(tokens), len(cards))
	}
	ones := 0
	for _, tok := range tokens {
		if e := tok.Effective(); e.Power == 1 && e.Toughness == 1 {
			ones++
		}
	}
	if ones != 2 {
		t.Errorf("%d 1/1 tokens, want 2", ones)
	}
}

// And a copy of a spell cast without it has none.
func TestACopyOfASpellWithoutGrantedOffspringMakesNoToken(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := zinniaOnBattlefield(t, g)
	bear := zinniaBearInHand(g, me)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{}); err != nil {
		t.Fatalf("cast: %v", err)
	}
	copyTheSpell(t, g, me, bear)
	zinniaSettle(t, g, me)
	if tokens, cards := zinniaBears(g); len(tokens) != 1 || len(cards) != 1 {
		t.Errorf("%d tokens and %d cards, want 1 and 1", len(tokens), len(cards))
	}
}

// A creature that leaves and comes back is a new object (CR 400.7): the
// offspring cost of the spell it once was is not paid for it.
func TestZinniaOffspringDoesNotFollowAPermanentThatReturns(t *testing.T) {
	g := newCatalogGame(t)
	me, _ := zinniaOnBattlefield(t, g)
	bear := zinniaBearInHand(g, me)
	if err := g.CastSpell(me.ID, bear, game.CastSpellParams{OptionalCosts: []int{0}}); err != nil {
		t.Fatalf("cast with offspring: %v", err)
	}
	zinniaSettle(t, g, me)
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneBattlefield}, game.ZoneRef{Kind: game.ZoneExile}, bear); err != nil {
		t.Fatalf("exile: %v", err)
	}
	if err := g.MoveCardByID(game.ZoneRef{Kind: game.ZoneExile}, game.ZoneRef{Kind: game.ZoneBattlefield}, bear); err != nil {
		t.Fatalf("return: %v", err)
	}
	zinniaSettle(t, g, me)
	if tokens, _ := zinniaBears(g); len(tokens) != 1 {
		t.Errorf("%d tokens, want 1: the returned creature was not cast with offspring", len(tokens))
	}
}

// The move enumerator offers the offspring cast only under Zinnia.
func TestZinniaOffspringIsEnumerated(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	bear := zinniaBearInHand(g, me)
	floatMana(t, g, me, "{G}{G}{G}{G}")
	offspringMoves := func() int {
		n := 0
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type != "cast_spell" || m.Source != bear {
				continue
			}
			var p struct {
				OptionalCosts []int `json:"optional_costs"`
			}
			if err := json.Unmarshal(m.Params, &p); err != nil {
				t.Fatal(err)
			}
			if len(p.OptionalCosts) > 0 {
				n++
			}
		}
		return n
	}
	if n := offspringMoves(); n != 0 {
		t.Errorf("%d offspring moves with no Zinnia", n)
	}
	pushCatalogPermanent(g, me.ID, "Zinnia, Valley's Voice", zinniaTypeLine, zinniaOracle, false)
	if n := offspringMoves(); n == 0 {
		t.Error("no offspring move under Zinnia")
	}
}

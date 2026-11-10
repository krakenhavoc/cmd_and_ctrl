package effects

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// chocobo_kick_test.go — #5358: "Kicker—Return a land you control to its
// owner's hand" (CR 702.33a, 601.2b, 601.2h), the return-to-hand cost
// #5247 put on a cast's mandatory slot, here on the optional one. The land
// rides cast_spell's return_ids only when the kicker is announced.

const chocoboKickOracle = "aa00c18b-8719-4095-940a-8e35a57b46c2"

func TestChocoboKickIsRegistered(t *testing.T) {
	spec, ok := Lookup(chocoboKickOracle)
	if !ok {
		t.Fatal("Chocobo Kick is not registered")
	}
	if spec.Name != "Chocobo Kick" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("name %q, completeness %v, caveats %v", spec.Name, spec.Completeness, spec.Caveats)
	}
	oc := game.OptionalCostsFor(chocoboKickOracle)
	if len(oc) != 1 || oc[0].Key != game.KickerKey || oc[0].ManaCost != "" || oc[0].ReturnToHand.Empty() ||
		oc[0].ReturnToHand.Count != 1 || oc[0].Label != "Kicker—Return a land you control to its owner's hand" {
		t.Fatalf("optional costs %+v, want one land-returning kicker", oc)
	}
	if game.AdditionalCostFor(chocoboKickOracle) != nil {
		t.Error("Chocobo Kick has a mandatory additional cost")
	}
}

// ckTable is a main phase on seat 0's turn: a 3/3 of mine, a 10/10 of
// the opponent's, and {1}{G} in the pool.
func ckTable(t *testing.T) (g *game.Game, me, opp *game.Player, mine, theirs uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	me, opp = g.Seats[0], g.Seats[1]
	mine = b12Creature(g, me.ID, "Chocobo", "Creature — Bird", 3, 3)
	theirs = b12Creature(g, opp.ID, "Behemoth", "Creature — Beast", 10, 10)
	advanceToMain(t, g)
	return g, me, opp, mine, theirs
}

// ckHand puts a Chocobo Kick into seat 0's hand.
func ckHand(g *game.Game, me *game.Player) uuid.UUID {
	return handCardOf(g, me, "Chocobo Kick", "Sorcery", "{1}{G}", chocoboKickOracle)
}

func ckCast(g *game.Game, me *game.Player, kick uuid.UUID, mine, theirs uuid.UUID, kicked bool, returns ...uuid.UUID) error {
	params := game.CastSpellParams{Targets: cardRefs(mine, theirs), ReturnIDs: returns, Strict: true}
	if kicked {
		params.OptionalCosts = []int{0}
	}
	return g.CastSpell(me.ID, kick, params)
}

func TestChocoboKickUnkickedDealsItsPower(t *testing.T) {
	g, me, _, mine, theirs := ckTable(t)
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	kick := ckHand(g, me)
	b06AddMana(me, "G", "C")
	if err := ckCast(g, me, kick, mine, theirs, false); err != nil {
		t.Fatalf("unkicked Chocobo Kick: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, theirs); got != 3 {
		t.Errorf("the Behemoth has %d damage, want 3", got)
	}
	if got := damageOn(t, g, mine); got != 0 {
		t.Errorf("my creature took %d damage — it is one-sided", got)
	}
	if !g.Battlefield.Contains(forest) {
		t.Error("an unkicked cast returned the Forest")
	}
}

func TestChocoboKickKickedReturnsALandAndDealsTwiceItsPower(t *testing.T) {
	g, me, _, mine, theirs := ckTable(t)
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	kick := ckHand(g, me)
	b06AddMana(me, "G", "C")
	if err := ckCast(g, me, kick, mine, theirs, true, forest); err != nil {
		t.Fatalf("kicked Chocobo Kick: %v", err)
	}
	if !g.Stack.Contains(kick) || g.Battlefield.Contains(forest) || !me.Hand.Contains(forest) {
		t.Fatal("the Forest was not returned to its owner's hand with the spell on the stack")
	}
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, theirs); got != 6 {
		t.Errorf("the Behemoth has %d damage, want 6 (twice the power, instead)", got)
	}
}

// A land you control but don't own goes to its OWNER's hand (CR 400.3).
func TestChocoboKickReturnsABorrowedLandToItsOwner(t *testing.T) {
	g, me, opp, mine, theirs := ckTable(t)
	borrowed := apaPush(g, opp.ID, me.ID, game.Card{Name: "Borrowed Forest", TypeLine: "Basic Land — Forest"})
	kick := ckHand(g, me)
	b06AddMana(me, "G", "C")
	if err := ckCast(g, me, kick, mine, theirs, true, borrowed); err != nil {
		t.Fatalf("kicked Chocobo Kick returning a borrowed land: %v", err)
	}
	if !opp.Hand.Contains(borrowed) || me.Hand.Contains(borrowed) {
		t.Error("the borrowed land did not go to its owner's hand")
	}
}

// CR 601.2g before 601.2h: the auto-tapper taps both lands for {1}{G} and
// the kicker then returns one of them. Two lands and nothing else pay the
// kicked cast in full.
func TestChocoboKickMayReturnALandTappedForItsMana(t *testing.T) {
	g, me, _, mine, theirs := ckTable(t)
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	kick := ckHand(g, me)
	err := g.CastSpell(me.ID, kick, game.CastSpellParams{Targets: cardRefs(mine, theirs), OptionalCosts: []int{0},
		ReturnIDs: []uuid.UUID{forest}, Strict: true, AutoTap: true})
	if err != nil {
		t.Fatalf("kicked Chocobo Kick off two Forests, returning one: %v", err)
	}
	if !me.Hand.Contains(forest) || !g.Stack.Contains(kick) {
		t.Fatal("the Forest was not returned with the spell on the stack")
	}
	passPriorityAroundTable(t, g)
	if got := damageOn(t, g, theirs); got != 6 {
		t.Errorf("the Behemoth has %d damage, want 6", got)
	}
}

// With no land the kicker can't be paid: the view's offer is short, the
// enumerator offers the unkicked cast only, and a kicked cast is refused
// with nothing paid. The unkicked cast is still legal.
func TestChocoboKickCannotBeKickedWithoutALand(t *testing.T) {
	g, me, _, mine, theirs := ckTable(t)
	kick := ckHand(g, me)
	b06AddMana(me, "G", "C")

	offer := ckKickerView(t, g, me.ID, kick)
	if offer.ReturnLabel != "a land you control" || offer.ReturnOptions == nil ||
		offer.ReturnOptions.Min != 1 || len(offer.ReturnOptions.Cards) != 0 {
		t.Fatalf("with no land, the kicker offer is %+v / %+v, want a short 1/1 list", offer, offer.ReturnOptions)
	}
	for _, m := range foiCastMoves(g, me.ID, kick) {
		var p struct {
			OptionalCosts []int `json:"optional_costs"`
		}
		_ = json.Unmarshal(m.Params, &p)
		if len(p.OptionalCosts) > 0 {
			t.Fatalf("the enumerator offers a kicked cast with no land: %s", m.Label)
		}
	}
	if len(foiCastMoves(g, me.ID, kick)) == 0 {
		t.Error("the enumerator offers no unkicked cast")
	}
	if err := ckCast(g, me, kick, mine, theirs, true); err == nil {
		t.Fatal("Chocobo Kick was kicked with nothing returned")
	}
	if err := ckCast(g, me, kick, mine, theirs, true, mine); err == nil {
		t.Fatal("Chocobo Kick was kicked by returning a creature")
	}
	if !me.Hand.Contains(kick) || !g.Battlefield.Contains(mine) {
		t.Fatal("a refused cast paid something")
	}
	if err := ckCast(g, me, kick, mine, theirs, false); err != nil {
		t.Fatalf("the unkicked cast with no land: %v", err)
	}
}

// Strict refusals with a land on the table: an opponent's land, two lands,
// the same land twice, and return_ids on an unkicked cast. Nothing is
// paid. Then the kicked cast with one land of mine goes through.
func TestChocoboKickRefusals(t *testing.T) {
	g, me, opp, mine, theirs := ckTable(t)
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	other := b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	theirLand := b12Permanent(g, opp.ID, "Island", "Basic Land — Island")
	kick := ckHand(g, me)
	b06AddMana(me, "G", "C")

	for name, c := range map[string]struct {
		kicked bool
		ids    []uuid.UUID
	}{
		"an opponent's land":             {true, []uuid.UUID{theirLand}},
		"two lands":                      {true, []uuid.UUID{forest, other}},
		"the same land twice":            {true, []uuid.UUID{forest, forest}},
		"the spell itself":               {true, []uuid.UUID{kick}},
		"return_ids on an unkicked cast": {false, []uuid.UUID{forest}},
	} {
		if err := ckCast(g, me, kick, mine, theirs, c.kicked, c.ids...); err == nil {
			t.Fatalf("%s was accepted", name)
		}
	}
	if !me.Hand.Contains(kick) || !g.Battlefield.Contains(forest) || !g.Battlefield.Contains(other) || !g.Battlefield.Contains(theirLand) {
		t.Fatal("a refused cast paid something")
	}
	// Kicked twice is not a thing (CR 702.33d): one kicker, paid once.
	if err := g.CastSpell(me.ID, kick, game.CastSpellParams{Targets: cardRefs(mine, theirs), OptionalCosts: []int{0, 0},
		ReturnIDs: []uuid.UUID{forest, other}, Strict: true}); err == nil {
		t.Fatal("the kicker was paid twice")
	}
	// Targets the clauses refuse: my own creature as the victim, theirs as the biter.
	if err := ckCast(g, me, kick, theirs, mine, true, forest); err == nil {
		t.Fatal("the target clauses were swapped and accepted")
	}
	if err := ckCast(g, me, kick, mine, theirs, true, forest); err != nil {
		t.Fatalf("the kicked cast: %v", err)
	}
}

// If either target is illegal as the spell resolves, no damage is dealt
// (the card's ruling, CR 608.2b).
func TestChocoboKickDealsNothingIfOneTargetIsIllegal(t *testing.T) {
	t.Run("my creature left", func(t *testing.T) {
		g, me, _, mine, theirs := ckTable(t)
		forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
		kick := ckHand(g, me)
		b06AddMana(me, "G", "C")
		if err := ckCast(g, me, kick, mine, theirs, true, forest); err != nil {
			t.Fatalf("kicked Chocobo Kick: %v", err)
		}
		g.WithWriteLock(func() { _, _ = g.Battlefield.Remove(mine) })
		passPriorityAroundTable(t, g)
		if got := damageOn(t, g, theirs); got != 0 {
			t.Errorf("the Behemoth has %d damage with the biter gone, want 0", got)
		}
	})
	t.Run("their creature changed sides", func(t *testing.T) {
		g, me, _, mine, theirs := ckTable(t)
		forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
		kick := ckHand(g, me)
		b06AddMana(me, "G", "C")
		if err := ckCast(g, me, kick, mine, theirs, true, forest); err != nil {
			t.Fatalf("kicked Chocobo Kick: %v", err)
		}
		// Still on the battlefield, but no longer "a creature an opponent
		// controls": an illegal target, so nothing is dealt to it.
		g.WithWriteLock(func() { g.GainControlForEffect(uuid.Nil, theirs, me.ID, game.IndefiniteDuration(), "theft") })
		passPriorityAroundTable(t, g)
		if got := damageOn(t, g, theirs); got != 0 {
			t.Errorf("the Behemoth has %d damage as an illegal target, want 0", got)
		}
	})
}

// A copy of a kicked spell is kicked too (CR 707.10): the copy deals twice
// the power, and so does the original — 12 to a 20/20.
func TestChocoboKickACopyOfAKickedSpellIsKicked(t *testing.T) {
	g, me, opp, mine, _ := ckTable(t)
	theirs := b12Creature(g, opp.ID, "Leviathan", "Creature — Leviathan", 20, 20)
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	kick := ckHand(g, me)
	b06AddMana(me, "G", "C")
	if err := ckCast(g, me, kick, mine, theirs, true, forest); err != nil {
		t.Fatalf("kicked Chocobo Kick: %v", err)
	}
	g.WithWriteLock(func() {
		if err := g.CopySpellForEffect(kick, me.ID, false, nil); err != nil {
			t.Fatalf("CopySpellForEffect: %v", err)
		}
	})
	if g.Stack.Size() != 2 {
		t.Fatalf("stack size %d after the copy, want 2", g.Stack.Size())
	}
	for i := 0; i < 8 && g.Stack.Size() > 0; i++ {
		passPriorityAroundTable(t, g)
	}
	if got := damageOn(t, g, theirs); got != 12 {
		t.Errorf("after the copy and the spell resolved the Leviathan has %d damage, want 12 (6 each: both kicked)", got)
	}
}

// The bot: with a land on the table the enumerator offers a kicked cast
// naming one land on return_ids, labelled as returning it, and the engine
// accepts that move.
func TestChocoboKickKickedIsOfferedAndPaidByTheEnumerator(t *testing.T) {
	g, me, _, _, _ := ckTable(t)
	forest := b12Permanent(g, me.ID, "Forest", "Basic Land — Forest")
	kick := ckHand(g, me)
	b06AddMana(me, "G", "C")

	if opts := ckKickerView(t, g, me.ID, kick).ReturnOptions; opts == nil || opts.Min != 1 || opts.Max != 1 ||
		len(opts.Cards) != 1 || opts.Cards[0] != forest.String() {
		t.Fatalf("return_options = %+v, want 1/1 over the Forest", opts)
	}
	for _, m := range foiCastMoves(g, me.ID, kick) {
		var p struct {
			OptionalCosts []int    `json:"optional_costs"`
			ReturnIDs     []string `json:"return_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatal(err)
		}
		if len(p.OptionalCosts) == 0 {
			if len(p.ReturnIDs) != 0 {
				t.Fatalf("an unkicked move names return_ids: %s", m.Label)
			}
			continue
		}
		if len(p.ReturnIDs) != 1 || p.ReturnIDs[0] != forest.String() {
			t.Fatalf("kicked move %q returns %v, want the Forest", m.Label, p.ReturnIDs)
		}
		if !strings.Contains(m.Label, "returning ") {
			t.Errorf("move label %q does not name the returned land", m.Label)
		}
		if err := actions.Dispatch(g, actions.Action{Type: actions.Type(m.Type), Player: me.ID, Caller: me.ID, Params: m.Params}); err != nil {
			t.Fatalf("the enumerated kicked cast is refused: %v", err)
		}
		if !g.Stack.Contains(kick) || !me.Hand.Contains(forest) {
			t.Fatal("the enumerated kicked cast did not return the Forest")
		}
		return
	}
	t.Fatal("the enumerator offers no kicked Chocobo Kick")
}

// Register refuses a second return-to-hand component on one card: a cast
// names its returns on one return_ids list.
func TestRegisterRefusesTwoReturnToHandCosts(t *testing.T) {
	both := Spec{
		OracleID:       "fixture-5358-both",
		Name:           "Fixture Both",
		AdditionalCost: ReturnPermanentToHandCost("a permanent you control"),
		OptionalCosts:  []game.AdditionalCost{KickerReturnToHand("a land you control", Land())},
	}
	if msg := registerPanics(both); !strings.Contains(msg, "return-to-hand cost components") {
		t.Errorf("a mandatory and an optional return registered (%q)", msg)
	}
	bad := KickerReturnToHand("a land you control", Land())
	bad.ReturnToHand.Count = 0
	if msg := registerPanics(Spec{OracleID: "fixture-5358-count", Name: "Fixture Count",
		OptionalCosts: []game.AdditionalCost{bad}}); !strings.Contains(msg, "returns 0 permanents") {
		t.Errorf("a kicker returning nothing registered (%q)", msg)
	}
	repeat := KickerReturnToHand("a land you control", Land())
	repeat.Repeat = 2
	if msg := registerPanics(Spec{OracleID: "fixture-5358-repeat", Name: "Fixture Repeat",
		OptionalCosts: []game.AdditionalCost{repeat}}); !strings.Contains(msg, "only a mana-only cost may repeat") {
		t.Errorf("a repeating return kicker registered (%q)", msg)
	}
}

// ckKickerView is the kicker offer the view ships on the card in hand.
func ckKickerView(t *testing.T, g *game.Game, seat, card uuid.UUID) protocol.OptionalCostView {
	t.Helper()
	c := foiHandView(g, seat, card)
	if c == nil || len(c.OptionalCosts) != 1 {
		t.Fatalf("the hand card ships no single optional cost: %+v", c)
	}
	return c.OptionalCosts[0]
}

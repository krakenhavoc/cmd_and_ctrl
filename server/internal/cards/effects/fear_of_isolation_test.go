package effects

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// fear_of_isolation_test.go — #5247: "As an additional cost to cast this
// spell, return a permanent you control to its owner's hand" (CR 601.2b,
// 601.2h), the activated abilities' ReturnToHandCost on a cast's
// mandatory slot. The permanent rides cast_spell's return_ids; it is
// chosen as the spell is cast, returned with the spell on the stack to
// its OWNER's hand (CR 400.3), and with nothing to return the spell can't
// be cast (CR 118.3).

const fearOfIsolationOracle = "0ee28549-5909-4372-8376-34010b7b6654"

func TestFearOfIsolationIsRegistered(t *testing.T) {
	spec, ok := Lookup(fearOfIsolationOracle)
	if !ok {
		t.Fatal("Fear of Isolation is not registered")
	}
	if spec.Completeness != CompletenessFull {
		t.Errorf("completeness %v, want Full", spec.Completeness)
	}
	ac := game.AdditionalCostFor(fearOfIsolationOracle)
	if ac == nil || ac.ReturnToHand.Empty() || ac.ReturnToHand.Count != 1 || ac.Label != "Return a permanent you control to its owner's hand" {
		t.Fatalf("additional cost %+v, want one return of a permanent you control", ac)
	}
}

// foiTable is a main phase on seat 0's turn with Fear of Isolation in an
// otherwise empty hand.
func foiTable(t *testing.T) (*game.Game, *game.Player, *game.Player, uuid.UUID) {
	t.Helper()
	g, me, opp := tapTable(t)
	fear := handCardOf(g, me, "Fear of Isolation", "Enchantment Creature — Nightmare", "{1}{U}", fearOfIsolationOracle)
	g.WithWriteLock(func() {
		for i := range me.Hand.Cards {
			if c := &me.Hand.Cards[i]; c.InstanceID == fear {
				c.Power, c.Toughness = 2, 3
			}
		}
	})
	return g, me, opp, fear
}

func foiCast(g *game.Game, p *game.Player, card uuid.UUID, returns ...uuid.UUID) error {
	return g.CastSpell(p.ID, card, game.CastSpellParams{ReturnIDs: returns, Strict: true})
}

// Paid: the returned permanent goes to its owner's hand while the spell is
// on the stack, and the spell resolves into a 2/3 flyer. A tapped land and
// a hexproof creature are both legal picks (returning is not tapping, and
// a cost does not target).
func TestFearOfIsolationReturnsAPermanentAndResolves(t *testing.T) {
	g, me, _, fear := foiTable(t)
	bear := pr7Creature(g, me.ID, "Bear", 2, "G")
	island := b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, bear).Keywords = []string{"hexproof"}
		findBattlefieldCardForTest(g, island).Tapped = true
		fillPoolColored(me, "U", 2)
	})
	if err := foiCast(g, me, fear, bear); err != nil {
		t.Fatalf("Fear of Isolation returning a hexproof Bear: %v", err)
	}
	if !g.Stack.Contains(fear) || g.Battlefield.Contains(bear) || !me.Hand.Contains(bear) {
		t.Fatal("the Bear was not returned to its owner's hand with the spell on the stack")
	}
	if !g.Battlefield.Contains(island) {
		t.Error("the Island was returned too")
	}
	passPriorityAroundTable(t, g)
	if !g.Battlefield.Contains(fear) {
		t.Fatal("Fear of Isolation did not resolve onto the battlefield")
	}
	if !slices.Contains(effectiveAbilities(t, g, fear), "flying") {
		t.Error("Fear of Isolation has no flying")
	}

	// A second copy, now paid with the tapped Island.
	again := handCardOf(g, me, "Fear of Isolation", "Enchantment Creature — Nightmare", "{1}{U}", fearOfIsolationOracle)
	g.WithWriteLock(func() { fillPoolColored(me, "U", 2) })
	if err := foiCast(g, me, again, island); err != nil {
		t.Fatalf("Fear of Isolation returning a tapped Island: %v", err)
	}
	if !me.Hand.Contains(island) {
		t.Error("the tapped Island was not returned")
	}
}

// Refused: with nothing to return the spell can't be cast (the view says
// so and no move is offered), and a cast naming nothing, an opponent's
// permanent, two permanents or the same permanent twice is refused with
// nothing paid. return_ids on a spell with no such cost is refused too.
func TestFearOfIsolationIsRefusedWithoutAPermanentToReturn(t *testing.T) {
	g, me, opp, fear := foiTable(t)
	g.WithWriteLock(func() { fillPoolColored(me, "U", 2) })
	theirs := pr7Creature(g, opp.ID, "Their Bear", 2, "G")

	if err := foiCast(g, me, fear); err == nil {
		t.Fatal("Fear of Isolation was cast with nothing to return")
	}
	if opts := foiReturnOptions(t, g, me.ID, fear); opts == nil || opts.Min != 1 || len(opts.Cards) != 0 {
		t.Fatalf("with nothing to return, return_options = %+v, want a short 1/1 list", opts)
	}
	if ms := foiCastMoves(g, me.ID, fear); len(ms) != 0 {
		t.Fatalf("Fear of Isolation is offered with nothing to return: %v", ms)
	}
	if foiCastableHere(t, g, me.ID, fear) {
		t.Error("the view stamps Fear of Isolation castable with nothing to return")
	}

	mine := pr7Creature(g, me.ID, "My Bear", 2, "G")
	other := b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	for name, ids := range map[string][]uuid.UUID{
		"nothing":                  nil,
		"an opponent's permanent":  {theirs},
		"two permanents":           {mine, other},
		"the same permanent twice": {mine, mine},
		"the spell itself":         {fear},
	} {
		if err := foiCast(g, me, fear, ids...); err == nil {
			t.Fatalf("%s paid Fear of Isolation", name)
		}
	}
	if !me.Hand.Contains(fear) || !g.Battlefield.Contains(mine) || !g.Battlefield.Contains(other) || !g.Battlefield.Contains(theirs) {
		t.Fatal("a refused cast paid something")
	}
	if !foiCastableHere(t, g, me.ID, fear) {
		t.Error("the view does not stamp Fear of Isolation castable with a permanent to return")
	}

	// A spell with no return cost refuses return_ids rather than
	// ignoring them.
	exposure := handCardOf(g, me, "Fear of Exposure", "Enchantment Creature — Nightmare", "{2}{G}", fearOfExposureOracle)
	taps := []uuid.UUID{other, mine}
	if err := g.CastSpell(me.ID, exposure, game.CastSpellParams{ReturnIDs: []uuid.UUID{mine}, TeamworkIDs: taps}); err == nil {
		t.Error("return_ids on a spell with no return cost were accepted")
	}
	if err := g.CastSpell(me.ID, exposure, game.CastSpellParams{TeamworkIDs: taps}); err != nil {
		t.Errorf("the same cast without return_ids: %v", err)
	}
}

// A permanent you control but don't own goes to its OWNER's hand (CR
// 400.3), not the caster's.
func TestFearOfIsolationReturnsAStolenPermanentToItsOwner(t *testing.T) {
	g, me, opp, fear := foiTable(t)
	stolen := apaPush(g, opp.ID, me.ID, game.Card{Name: "Borrowed Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})
	g.WithWriteLock(func() { fillPoolColored(me, "U", 2) })
	if err := foiCast(g, me, fear, stolen); err != nil {
		t.Fatalf("Fear of Isolation returning a permanent you control but don't own: %v", err)
	}
	if !opp.Hand.Contains(stolen) || me.Hand.Contains(stolen) || g.Battlefield.Contains(stolen) {
		t.Fatal("the stolen permanent did not go to its owner's hand")
	}
}

// CR 601.2g before 601.2h: the auto-tapper may tap an Island for the
// spell's mana and the cost then returns that same Island. With two
// Islands and nothing else, the cast is paid in full.
func TestFearOfIsolationMayReturnALandTappedForItsMana(t *testing.T) {
	g, me, _, fear := foiTable(t)
	a := b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	if err := g.CastSpell(me.ID, fear, game.CastSpellParams{ReturnIDs: []uuid.UUID{a}, Strict: true, AutoTap: true}); err != nil {
		t.Fatalf("Fear of Isolation off two Islands, returning one: %v", err)
	}
	if !me.Hand.Contains(a) || !g.Stack.Contains(fear) {
		t.Fatal("the Island was not returned with the spell on the stack")
	}
}

// The bot: the legal-move list offers the cast with one permanent named on
// return_ids and says so in its label, and the engine accepts that move.
func TestFearOfIsolationIsOfferedAndPaidByTheEnumerator(t *testing.T) {
	g, me, _, fear := foiTable(t)
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	b12Permanent(g, me.ID, "Island", "Basic Land — Island")
	pr7Creature(g, me.ID, "Bear", 2, "G")

	if opts := foiReturnOptions(t, g, me.ID, fear); opts == nil || opts.Min != 1 || opts.Max != 1 || len(opts.Cards) != 3 {
		t.Fatalf("return_options = %+v, want 1/1 over the two Islands and the Bear", opts)
	}
	ms := foiCastMoves(g, me.ID, fear)
	if len(ms) == 0 {
		t.Fatal("the enumerator offers no Fear of Isolation cast")
	}
	var p struct {
		ReturnIDs []string `json:"return_ids"`
	}
	if err := json.Unmarshal(ms[0].Params, &p); err != nil || len(p.ReturnIDs) != 1 {
		t.Fatalf("enumerated return payment %v (%v), want one permanent", p.ReturnIDs, err)
	}
	returned, err := uuid.Parse(p.ReturnIDs[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(ms[0].Label, "returning ") {
		t.Errorf("move label %q does not name the returned permanent", ms[0].Label)
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(ms[0].Type), Player: me.ID, Caller: me.ID, Params: ms[0].Params}); err != nil {
		t.Fatalf("the enumerated Fear of Isolation cast is refused: %v", err)
	}
	if !g.Stack.Contains(fear) || !me.Hand.Contains(returned) {
		t.Fatal("the enumerated cast did not return its permanent with the spell on the stack")
	}
}

// A restore taken mid-cast keeps the choice: Fear of Isolation cast by
// returning a commander parks on its owner's CR 903.9b question with
// nothing paid; restored strictly while it is open, the answer re-makes
// the cast on the restored table with the same pick.
func TestAFearOfIsolationCommanderQuestionIsARestorePoint(t *testing.T) {
	for _, apply := range []bool{false, true} {
		g, me, _, fear := foiTable(t)
		cmd := pushBattlefieldCardWithTimestamp(g, costCommander(me.ID, "My Commander"))
		g.WithWriteLock(func() { fillPoolColored(me, "U", 2) })
		if err := foiCast(g, me, fear, cmd); err != nil {
			t.Fatalf("casting Fear of Isolation: %v", err)
		}
		if g.Stack.Contains(fear) || !g.Battlefield.Contains(cmd) {
			t.Fatal("Fear of Isolation was cast before the commander's owner answered")
		}

		g = s13StrictRestore(t, g, game.CostCommanderAnnounceKey)
		me = g.Seats[0]
		if got := g.PendingChoices[0].CommanderHeadedFor(); got != game.ZoneHand {
			t.Errorf("restored CommanderHeadedFor() = %q, want %q", got, game.ZoneHand)
		}
		answerCostCommander(t, g, me.ID, apply)
		if !g.Stack.Contains(fear) || g.Battlefield.Contains(cmd) {
			t.Fatalf("apply=%v: the restored answer did not cast Fear of Isolation returning the commander", apply)
		}
		if apply && !me.Command.Contains(cmd) {
			t.Error("the commander did not take the command zone on the restored table")
		}
		if !apply && !me.Hand.Contains(cmd) {
			t.Error("the declined commander did not go to its owner's hand on the restored table")
		}
		passPriorityAroundTable(t, g)
		if !g.Battlefield.Contains(fear) {
			t.Errorf("apply=%v: Fear of Isolation did not resolve on the restored table", apply)
		}
	}
}

// foiReturnOptions is the return picker the view ships on the card in hand.
func foiReturnOptions(t *testing.T, g *game.Game, seat, card uuid.UUID) *protocol.LegalTargetsView {
	t.Helper()
	if c := foiHandView(g, seat, card); c != nil && c.AdditionalCost != nil {
		if c.AdditionalCost.ReturnLabel != "a permanent you control" {
			t.Errorf("return_label %q", c.AdditionalCost.ReturnLabel)
		}
		return c.AdditionalCost.ReturnOptions
	}
	return nil
}

// foiCastableHere is the cast gate the view's castability reads (CR
// 601.2h): whether the card's additional cost can be paid at all.
func foiCastableHere(t *testing.T, g *game.Game, seat, card uuid.UUID) bool {
	t.Helper()
	ok, found := false, false
	g.WithWriteLock(func() {
		for _, p := range g.Seats {
			for _, c := range p.Hand.Cards {
				if c.InstanceID == card {
					found = true
					ok = g.AnyAdditionalCostBranchPayableLocked(seat, c)
				}
			}
		}
	})
	if !found {
		t.Fatal("the card is not in a hand")
	}
	return ok
}

func foiHandView(g *game.Game, seat, card uuid.UUID) *protocol.CardView {
	v := protocol.ViewOfGameFor(g, seat.String())
	for _, s := range v.Seats {
		for i := range s.Hand.Cards {
			if c := &s.Hand.Cards[i]; c.InstanceID == card.String() {
				return c
			}
		}
	}
	return nil
}

func foiCastMoves(g *game.Game, seat, card uuid.UUID) []legal.Move {
	var out []legal.Move
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == legal.TypeCastSpell && m.Source == card {
			out = append(out, m)
		}
	}
	return out
}

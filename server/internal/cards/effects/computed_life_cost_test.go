package effects

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// computed_life_cost_test.go — #1594, ADR 0020 Decision 47: a life
// component whose amount is a registered count (AbilityCost.LifeFrom),
// read once at announce (CR 601.2f–g via CR 602.2b) by the one function
// the engine, the legal enumerator and the view share.

const (
	warRoomOracle           = "71c52bf5-2a5d-488e-8b15-7ef290e4b77d"
	murderousBetrayalOracle = "a2fe1e2b-b621-4406-b5e0-f2eebd12c2ba"
	lurkingEvilOracle       = "81d6b42e-9c41-4d1e-bd28-4662e1fde572"
)

// warRoomDraw is the index of War Room's draw ability in its
// Activated list (the {C} ability is a mana ability, listed apart).
const warRoomDraw = 0

// clrCommander puts a commander `owner` owns into their command zone
// with the given colour identity. A real imported card (an OracleID),
// so an empty identity is an authoritative "colourless" rather than
// missing data (CR 903.4f, #844).
func clrCommander(p *game.Player, name string, identity ...string) uuid.UUID {
	id := uuid.New()
	p.Command.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: "Legendary Creature — Human",
		OracleID: "test-commander-" + name, IsCommander: true,
		ColorIdentity: identity, Owner: p.ID, Controller: p.ID,
	})
	return id
}

// clrWarRoom seeds War Room under `p`, walks to main and floats {3}.
func clrWarRoom(t *testing.T, g *game.Game, p *game.Player) uuid.UUID {
	t.Helper()
	room := b12Push(g, p.ID, "War Room", "Land", warRoomOracle, 0, 0)
	advanceToMain(t, g)
	fillPool(p, 3)
	return room
}

// clrWarRoomCosts activates War Room's draw at `life` and asserts the
// payment and the draw.
func clrWarRoomCosts(t *testing.T, g *game.Game, me *game.Player, room uuid.UUID, want int) {
	t.Helper()
	life, hand := me.Life, me.Hand.Size()
	b16Activate(t, g, me.ID, room, warRoomDraw, game.ActivateAbilityParams{})
	if got := life - me.Life; got != want {
		t.Errorf("War Room cost %d life, want %d", got, want)
	}
	if got := me.Hand.Size() - hand; got != 1 {
		t.Errorf("drew %d, want 1", got)
	}
}

func TestWarRoomMonoColourCommanderCostsOneLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	clrCommander(me, "Mono", "G")
	room := clrWarRoom(t, g, me)
	clrWarRoomCosts(t, g, me, room, 1)
}

// "Your commanders'" is the COMBINED identity (CR 903.4): a partner
// pair contributes the union of both halves, and a colour both share
// counts once — W/U + U/B/R is four colours, not five.
func TestWarRoomFourColourPartnersCostFourLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	clrCommander(me, "Left", "W", "U")
	clrCommander(me, "Right", "U", "B", "R")
	room := clrWarRoom(t, g, me)
	clrWarRoomCosts(t, g, me, room, 4)
}

// A colourless commander has zero colours, so the cost is free — and
// with no commander at all it is zero too. Neither is refused.
func TestWarRoomColourlessOrNoCommanderCostsNoLife(t *testing.T) {
	t.Run("colourless", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		clrCommander(me, "Karn")
		room := clrWarRoom(t, g, me)
		clrWarRoomCosts(t, g, me, room, 0)
	})
	t.Run("no commander", func(t *testing.T) {
		g := newCatalogGame(t)
		me := g.Seats[0]
		room := clrWarRoom(t, g, me)
		clrWarRoomCosts(t, g, me, room, 0)
	})
}

// Only the ACTIVATOR's commanders count: an opponent's five-colour
// commander does not raise your War Room's price.
func TestWarRoomCountsOnlyYourCommanders(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	clrCommander(me, "Mine", "B")
	clrCommander(opp, "Theirs", "W", "U", "B", "R", "G")
	room := clrWarRoom(t, g, me)
	clrWarRoomCosts(t, g, me, room, 1)
}

// CR 119.4: at 2 life a three-colour War Room cannot be paid. The
// engine refuses it with nothing spent, and the enumerator does not
// offer it (#695's rule, applied to the computed amount). At 3 life it
// is offered, and the move carries the computed 3 on Cost.Life — the
// number a bot prices.
func TestWarRoomUnaffordableLifeIsRefusedAndNotOffered(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	clrCommander(me, "Three", "U", "B", "R")
	room := clrWarRoom(t, g, me)
	me.Life = 2
	pool := len(me.ManaPool)

	if err := g.ActivateCatalogAbility(me.ID, room, warRoomDraw, game.ActivateAbilityParams{}); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("at 2 life a 3-life cost must be refused, got %v", err)
	}
	if me.Life != 2 || len(me.ManaPool) != pool || b16Tapped(t, g, room) {
		t.Errorf("a refused activation spent something: life %d, pool %d→%d, tapped %v",
			me.Life, pool, len(me.ManaPool), b16Tapped(t, g, room))
	}
	if m, ok := clrWarRoomMove(g, me.ID, room); ok {
		t.Errorf("the enumerator offered an unaffordable War Room: %q", m.Label)
	}

	me.Life = 3
	m, ok := clrWarRoomMove(g, me.ID, room)
	if !ok {
		t.Fatal("at 3 life the activation is affordable and must be offered")
	}
	if m.Cost == nil || m.Cost.Life != 3 {
		t.Errorf("move Cost = %+v, want Life = the computed 3", m.Cost)
	}
}

func clrWarRoomMove(g *game.Game, seat, room uuid.UUID) (legal.Move, bool) {
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == legal.TypeActivateAbility && m.Source == room {
			return m, true
		}
	}
	return legal.Move{}, false
}

// CR 601.2g: the total cost is locked in before it is paid. A
// commander that changes after the announcement — here a partner that
// widens the identity from one colour to four — changes nothing about
// the activation already on the stack: 1 life was paid, 1 is recorded,
// and resolving charges nothing more.
func TestWarRoomAmountIsFixedAtAnnounce(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	clrCommander(me, "Mono", "G")
	room := clrWarRoom(t, g, me)
	life := me.Life
	if err := g.ActivateCatalogAbility(me.ID, room, warRoomDraw, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	if got := life - me.Life; got != 1 {
		t.Fatalf("announce charged %d, want 1", got)
	}
	var paid int
	for _, meta := range g.StackMeta {
		if meta != nil && meta.SourceCardID == room {
			paid = meta.Paid.LifePaid
		}
	}
	if paid != 1 {
		t.Errorf("stack item records LifePaid %d, want 1", paid)
	}

	clrCommander(me, "Partner", "W", "U", "B")
	passPriorityAroundTable(t, g)
	if got := life - me.Life; got != 1 {
		t.Errorf("after the commander changed and the ability resolved, total paid %d, want 1", got)
	}
}

// The preview: the ability row's life_cost is the computed amount for
// the controller, and it follows the board.
func TestWarRoomPreviewShowsTheComputedLife(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	clrCommander(me, "Pair", "W", "B")
	room := clrWarRoom(t, g, me)
	if got := activatedRowOf(t, g, me.ID, room, warRoomDraw).LifeCost; got != 2 {
		t.Errorf("life_cost = %d, want 2", got)
	}
	clrCommander(me, "Background", "G")
	if got := activatedRowOf(t, g, me.ID, room, warRoomDraw).LifeCost; got != 3 {
		t.Errorf("after a third colour joined, life_cost = %d, want 3", got)
	}
}

// Undo (Clone / RestoreFrom) and a snapshot round trip: the paid amount
// rides the stack item, and the COUNT rides the cost as a key, so a
// restored game prices the next activation from the same function.
func TestWarRoomSurvivesUndoAndSnapshot(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	clrCommander(me, "Pair", "U", "R")
	room := clrWarRoom(t, g, me)
	before := g.Clone()
	life := me.Life

	if err := g.ActivateCatalogAbility(me.ID, room, warRoomDraw, game.ActivateAbilityParams{}); err != nil {
		t.Fatalf("activate: %v", err)
	}
	blob, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var round game.GameSnapshot
	if err := json.Unmarshal(blob, &round); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	restored, err := round.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	rme := restored.Seats[0]
	if got := life - rme.Life; got != 2 {
		t.Errorf("restored life paid %d, want 2", got)
	}
	hand := rme.Hand.Size()
	passPriorityAroundTable(t, restored)
	if rme.Hand.Size() != hand+1 || life-rme.Life != 2 {
		t.Errorf("restored resolution: drew %d, paid %d; want 1 and 2", rme.Hand.Size()-hand, life-rme.Life)
	}

	// Undo the activation, then activate again from the rewound game:
	// the key still prices it.
	g.RestoreFrom(before)
	me = g.Seats[0]
	if me.Life != life {
		t.Fatalf("undo left life at %d, want %d", me.Life, life)
	}
	clrWarRoomCosts(t, g, me, room, 2)
}

// Murderous Betrayal: half your life, rounded up, read at announce —
// 20 → 10, then 10 → 5, and at 1 life it still costs the last 1.
func TestMurderousBetrayalPaysHalfYourLifeRoundedUp(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	betrayal := b12Push(g, me.ID, "Murderous Betrayal", "Enchantment", murderousBetrayalOracle, 0, 0)
	advanceToMain(t, g)
	me.Life = 20
	for i, want := range []int{10, 5} {
		bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
		b06AddMana(me, "B")
		b06AddMana(me, "B")
		life := me.Life
		b16Activate(t, g, me.ID, betrayal, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)})
		if got := life - me.Life; got != want {
			t.Errorf("activation %d cost %d life, want %d", i+1, got, want)
		}
		if _, still := battlefieldCard(g, bear); still {
			t.Errorf("activation %d: the bear survived", i+1)
		}
	}
	me.Life = 1
	bear := b12Creature(g, opp.ID, "Bear", "Creature — Bear", 2, 2)
	b06AddMana(me, "B")
	b06AddMana(me, "B")
	if err := g.ActivateCatalogAbility(me.ID, betrayal, 0, game.ActivateAbilityParams{Targets: cardRefs(bear)}); err != nil {
		t.Fatalf("at 1 life the half (1) is payable: %v", err)
	}
	if me.Life != 0 {
		t.Errorf("life %d, want 0", me.Life)
	}
}

// Lurking Evil: the life is half, and the enchantment becomes a 4/4
// flying Phyrexian Horror creature — and stops being an enchantment
// (CR 205.1a: "becomes a … creature" sets the card type).
func TestLurkingEvilBecomesAFourFourFlyingHorror(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	evil := b12Push(g, me.ID, "Lurking Evil", "Enchantment", lurkingEvilOracle, 0, 0)
	advanceToMain(t, g)
	me.Life = 15
	b16Activate(t, g, me.ID, evil, 0, game.ActivateAbilityParams{})
	if me.Life != 7 {
		t.Errorf("life %d, want 7 (15 − 8)", me.Life)
	}
	if p, tough := effectivePower(t, g, evil), effectiveToughness(t, g, evil); p != 4 || tough != 4 {
		t.Errorf("P/T %d/%d, want 4/4", p, tough)
	}
	types := effectiveTypes(t, g, evil)
	if !clrHas(types, "Creature") || clrHas(types, "Enchantment") {
		t.Errorf("types %v, want a creature and no longer an enchantment", types)
	}
	subs := effectiveSubtypes(t, g, evil)
	if !clrHas(subs, "Phyrexian") || !clrHas(subs, "Horror") {
		t.Errorf("subtypes %v, want Phyrexian Horror", subs)
	}
	if !clrHas(effectiveAbilities(t, g, evil), "flying") {
		t.Errorf("abilities %v, want flying", effectiveAbilities(t, g, evil))
	}
}

func clrHas(xs []string, want string) bool {
	for _, x := range xs {
		if x == want {
			return true
		}
	}
	return false
}

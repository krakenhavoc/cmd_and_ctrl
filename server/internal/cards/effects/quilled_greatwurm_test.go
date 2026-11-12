package effects

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// quilled_greatwurm_test.go — #5390: a cast's "remove N counters from
// among creatures you control", and Quilled Greatwurm, the card that
// waited on it.

const quilledGreatwurmOracle = "01635d21-aeda-4754-9d94-68fcd0cd6556"

const qgOfferKey = "remove_counters_from_graveyard"

// qgTable is a main phase on seat 0's turn with Quilled Greatwurm in
// seat 0's graveyard and its {4}{G}{G} floating.
func qgTable(t *testing.T) (g *game.Game, me, opp *game.Player, wurm uuid.UUID) {
	t.Helper()
	g = newCatalogGame(t)
	advanceToMain(t, g)
	me, opp = g.Seats[0], g.Seats[1]
	wurm = graveyardCardOf(g, me, "Quilled Greatwurm", "Creature — Wurm", "{4}{G}{G}", quilledGreatwurmOracle)
	payMana(t, g, me, "{C}{C}{C}{C}{G}{G}")
	return g, me, opp, wurm
}

// qgCreature puts a creature with the given counters onto the battlefield
// under owner's control.
func qgCreature(g *game.Game, owner uuid.UUID, name string, counters map[string]int) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{
		Name: name, TypeLine: "Creature — Test", Power: 1, Toughness: 1, Counters: counters,
	})
}

func qgCounters(g *game.Game, id uuid.UUID) map[string]int {
	c := findBattlefieldCardForTest(g, id)
	if c == nil {
		return nil
	}
	return c.Counters
}

// qgCast casts the Greatwurm out of the graveyard claiming its offer,
// paying `counts[i]` counters of `kinds[i]` off `ids[i]`.
func qgCast(g *game.Game, me *game.Player, wurm uuid.UUID, ids []uuid.UUID, counts []int, kinds []string) error {
	return g.CastSpell(me.ID, wurm, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: qgOfferKey, Strict: true,
		CounterSourceIDs: ids, CounterCounts: counts, CounterKinds: kinds,
	})
}

func TestQuilledGreatwurmDeclaresItsGraveyardCast(t *testing.T) {
	spec, ok := Lookup(quilledGreatwurmOracle)
	if !ok {
		t.Fatal("Quilled Greatwurm is not in the catalog")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("Quilled Greatwurm: %s %v, want Full with no caveats", spec.Completeness, spec.Caveats)
	}
	if !game.CardCastableFromZone(quilledGreatwurmOracle, game.ZoneGraveyard) {
		t.Fatal("Quilled Greatwurm does not open the graveyard")
	}
	offers := game.AlternativeCostsOfferedFromZone(quilledGreatwurmOracle, game.ZoneGraveyard)
	if len(offers) != 1 || offers[0].Key != qgOfferKey {
		t.Fatalf("graveyard offers = %+v, want the one counter-removal offer", offers)
	}
	o := offers[0]
	if o.ManaCost != "{4}{G}{G}" {
		t.Errorf("the graveyard cast pays %q, want the printed {4}{G}{G}", o.ManaCost)
	}
	rc := o.RemoveCounters
	if rc == nil || rc.N != 6 || !rc.Among || rc.Counter != "" || rc.Variable || rc.From == nil {
		t.Fatalf("the offer's counter removal = %+v, want six counters of any kinds from among creatures", rc)
	}
	if o.PaysCards() || o.ExileOnLeavingStack {
		t.Errorf("the offer names a card or exiles the Greatwurm: %+v", o)
	}
	if got := game.AlternativeCostsOfferedFromZone(quilledGreatwurmOracle, game.ZoneHand); len(got) != 0 {
		t.Errorf("offers from hand = %+v, want none", got)
	}
}

// Six counters of two kinds, split across two creatures as the caster
// chooses, pay for the cast: exactly six come off, with the spell on the
// stack, and the Greatwurm resolves onto the battlefield.
func TestQuilledGreatwurmGraveyardCastRemovesExactlySixSplitAcrossCreatures(t *testing.T) {
	g, me, _, wurm := qgTable(t)
	a := qgCreature(g, me.ID, "Bear A", map[string]int{game.CounterPlusOne: 4})
	b := qgCreature(g, me.ID, "Bear B", map[string]int{game.CounterPlusOne: 3, "oil": 2})

	if err := qgCast(g, me, wurm, []uuid.UUID{a, b, b}, []int{3, 1, 2}, []string{game.CounterPlusOne, game.CounterPlusOne, "oil"}); err != nil {
		t.Fatalf("graveyard cast: %v", err)
	}
	if got := qgCounters(g, a)[game.CounterPlusOne]; got != 1 {
		t.Errorf("Bear A has %d +1/+1 counters, want 1", got)
	}
	if got := qgCounters(g, b); got[game.CounterPlusOne] != 2 || got["oil"] != 0 {
		t.Errorf("Bear B counters = %v, want two +1/+1 and no oil", got)
	}
	if g.Stack.Size() != 1 || inGraveyard(me, wurm) {
		t.Fatalf("the Greatwurm is not on the stack (stack %d)", g.Stack.Size())
	}
	if item := g.StackMeta[wurm]; item == nil || item.Paid.CountersRemoved != 6 {
		t.Errorf("paid record = %+v, want six counters removed", item)
	}
	if len(me.ManaPool) != 0 {
		t.Errorf("the mana was not paid: pool %v", me.ManaPool)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, wurm) == nil {
		t.Fatal("the Greatwurm did not resolve onto the battlefield")
	}
}

// Five counters in all can't pay six (CR 118.3): the offer is withheld
// from the view and the enumerator, and a cast that claims it anyway is
// refused with nothing removed and nothing spent.
func TestQuilledGreatwurmGraveyardCastIsRefusedWithFiveCounters(t *testing.T) {
	g, me, _, wurm := qgTable(t)
	a := qgCreature(g, me.ID, "Bear A", map[string]int{game.CounterPlusOne: 3})
	b := qgCreature(g, me.ID, "Bear B", map[string]int{"oil": 2})

	if err := qgCast(g, me, wurm, []uuid.UUID{a, b}, []int{3, 2}, []string{game.CounterPlusOne, "oil"}); err == nil {
		t.Fatal("five counters paid a six-counter cost")
	}
	if !inGraveyard(me, wurm) || qgCounters(g, a)[game.CounterPlusOne] != 3 || qgCounters(g, b)["oil"] != 2 || len(me.ManaPool) == 0 {
		t.Fatal("a refused cast moved the card, removed counters or spent mana")
	}
	v := graveyardViewOf(t, g, me.ID, wurm)
	for _, o := range v.AlternativeCosts {
		if o.Key == qgOfferKey {
			t.Errorf("the view offers the graveyard cast with five counters: %+v", o)
		}
	}
	if moves := retraceMoves(g, me.ID, wurm); len(moves) != 0 {
		t.Errorf("the enumerator offers %d graveyard casts with five counters", len(moves))
	}
}

// Every way to get the payment wrong is refused, and none of them removes
// a counter, moves the card or spends the mana.
func TestQuilledGreatwurmStrictRefusals(t *testing.T) {
	cases := []struct {
		name string
		pay  func(g *game.Game, me *game.Player, wurm, a, b, theirs, rock uuid.UUID) error
	}{
		{"five of six", func(g *game.Game, me *game.Player, wurm, a, b, _, _ uuid.UUID) error {
			return qgCast(g, me, wurm, []uuid.UUID{a, b}, []int{4, 1}, []string{game.CounterPlusOne, game.CounterPlusOne})
		}},
		{"seven of six", func(g *game.Game, me *game.Player, wurm, a, b, _, _ uuid.UUID) error {
			return qgCast(g, me, wurm, []uuid.UUID{a, b}, []int{4, 3}, []string{game.CounterPlusOne, game.CounterPlusOne})
		}},
		{"more than a creature holds", func(g *game.Game, me *game.Player, wurm, a, _, _, _ uuid.UUID) error {
			return qgCast(g, me, wurm, []uuid.UUID{a}, []int{6}, []string{game.CounterPlusOne})
		}},
		{"a kind it does not hold", func(g *game.Game, me *game.Player, wurm, a, b, _, _ uuid.UUID) error {
			return qgCast(g, me, wurm, []uuid.UUID{a, b}, []int{4, 2}, []string{game.CounterPlusOne, "oil"})
		}},
		{"one part named twice", func(g *game.Game, me *game.Player, wurm, a, b, _, _ uuid.UUID) error {
			return qgCast(g, me, wurm, []uuid.UUID{a, a, b}, []int{2, 2, 2}, []string{game.CounterPlusOne, game.CounterPlusOne, game.CounterPlusOne})
		}},
		{"an opponent's creature", func(g *game.Game, me *game.Player, wurm, a, _, theirs, _ uuid.UUID) error {
			return qgCast(g, me, wurm, []uuid.UUID{a, theirs}, []int{4, 2}, []string{game.CounterPlusOne, game.CounterPlusOne})
		}},
		{"a noncreature permanent", func(g *game.Game, me *game.Player, wurm, a, _, _, rock uuid.UUID) error {
			return qgCast(g, me, wurm, []uuid.UUID{a, rock}, []int{4, 2}, []string{game.CounterPlusOne, "charge"})
		}},
		{"no counters named", func(g *game.Game, me *game.Player, wurm, _, _, _, _ uuid.UUID) error {
			return qgCast(g, me, wurm, nil, nil, nil)
		}},
		{"counts without kinds", func(g *game.Game, me *game.Player, wurm, a, b, _, _ uuid.UUID) error {
			return qgCast(g, me, wurm, []uuid.UUID{a, b}, []int{4, 2}, nil)
		}},
		{"a zero part", func(g *game.Game, me *game.Player, wurm, a, b, _, _ uuid.UUID) error {
			return qgCast(g, me, wurm, []uuid.UUID{a, b, b}, []int{4, 2, 0}, []string{game.CounterPlusOne, game.CounterPlusOne, "oil"})
		}},
		{"the offer not claimed", func(g *game.Game, me *game.Player, wurm, a, b, _, _ uuid.UUID) error {
			return g.CastSpell(me.ID, wurm, game.CastSpellParams{
				FromZone: "graveyard", Strict: true,
				CounterSourceIDs: []uuid.UUID{a, b}, CounterCounts: []int{4, 2}, CounterKind: game.CounterPlusOne,
			})
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g, me, opp, wurm := qgTable(t)
			a := qgCreature(g, me.ID, "Bear A", map[string]int{game.CounterPlusOne: 4})
			b := qgCreature(g, me.ID, "Bear B", map[string]int{game.CounterPlusOne: 3})
			theirs := qgCreature(g, opp.ID, "Their Bear", map[string]int{game.CounterPlusOne: 5})
			rock := apaPush(g, me.ID, me.ID, game.Card{Name: "Charged Rock", TypeLine: "Artifact", Counters: map[string]int{"charge": 5}})
			if err := tc.pay(g, me, wurm, a, b, theirs, rock); err == nil {
				t.Fatal("the payment was accepted")
			}
			if !inGraveyard(me, wurm) || g.Stack.Size() != 0 || len(me.ManaPool) == 0 {
				t.Fatal("a refused cast moved the card or spent mana")
			}
			if qgCounters(g, a)[game.CounterPlusOne] != 4 || qgCounters(g, b)[game.CounterPlusOne] != 3 ||
				qgCounters(g, theirs)[game.CounterPlusOne] != 5 || qgCounters(g, rock)["charge"] != 5 {
				t.Fatal("a refused cast removed counters")
			}
		})
	}
}

// One kind for the whole payment may ride counter_kind alone, as an
// activation's does.
func TestQuilledGreatwurmPaymentOfOneKindRidesCounterKind(t *testing.T) {
	g, me, _, wurm := qgTable(t)
	a := qgCreature(g, me.ID, "Bear A", map[string]int{game.CounterPlusOne: 4})
	b := qgCreature(g, me.ID, "Bear B", map[string]int{game.CounterPlusOne: 2})
	if err := g.CastSpell(me.ID, wurm, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: qgOfferKey, Strict: true,
		CounterSourceIDs: []uuid.UUID{a, b}, CounterCounts: []int{4, 2}, CounterKind: game.CounterPlusOne,
	}); err != nil {
		t.Fatalf("graveyard cast: %v", err)
	}
	if qgCounters(g, a)[game.CounterPlusOne] != 0 || qgCounters(g, b)[game.CounterPlusOne] != 0 {
		t.Errorf("counters left: %v %v", qgCounters(g, a), qgCounters(g, b))
	}
}

// The hand cast is the printed cost and nothing else: no counters, and a
// counter payment sent with it is a stray and refused.
func TestQuilledGreatwurmHandCastNeedsNoCounters(t *testing.T) {
	g := newCatalogGame(t)
	advanceToMain(t, g)
	me := g.Seats[0]
	a := qgCreature(g, me.ID, "Bear A", map[string]int{game.CounterPlusOne: 6})
	wurm := uuid.New()
	me.Hand.PushTop(game.Card{InstanceID: wurm, Name: "Quilled Greatwurm", TypeLine: "Creature — Wurm",
		ManaCost: "{4}{G}{G}", OracleID: quilledGreatwurmOracle, Power: 7, Toughness: 7,
		Owner: me.ID, Controller: me.ID, KnownBy: map[uuid.UUID]bool{me.ID: true}})
	payMana(t, g, me, "{C}{C}{C}{C}{G}{G}")

	err := g.CastSpell(me.ID, wurm, game.CastSpellParams{
		Strict: true, CounterSourceIDs: []uuid.UUID{a}, CounterCounts: []int{6}, CounterKind: game.CounterPlusOne,
	})
	if err == nil {
		t.Fatal("a hand cast accepted a counter payment it has no cost for")
	}
	if !errors.Is(err, game.ErrInvalidParam) {
		t.Errorf("stray counter payment: %v, want ErrInvalidParam", err)
	}
	if err := g.CastSpell(me.ID, wurm, game.CastSpellParams{Strict: true}); err != nil {
		t.Fatalf("hand cast: %v", err)
	}
	if got := qgCounters(g, a)[game.CounterPlusOne]; got != 6 {
		t.Errorf("a hand cast removed counters: %d left of 6", got)
	}
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, wurm) == nil {
		t.Fatal("the hand-cast Greatwurm did not resolve")
	}
}

// The view ships the offer with the activation's counter block, so the
// client opens its counter picker.
func TestQuilledGreatwurmViewShipsTheCounterPicker(t *testing.T) {
	g, me, opp, wurm := qgTable(t)
	a := qgCreature(g, me.ID, "Bear A", map[string]int{game.CounterPlusOne: 4})
	b := qgCreature(g, me.ID, "Bear B", map[string]int{"oil": 2, game.CounterPlusOne: 1})
	qgCreature(g, opp.ID, "Their Bear", map[string]int{game.CounterPlusOne: 9})

	v := graveyardViewOf(t, g, me.ID, wurm)
	var found bool
	for _, o := range v.AlternativeCosts {
		if o.Key != qgOfferKey {
			continue
		}
		found = true
		if o.CounterCostN != 6 || !o.CounterCostAmong || o.CounterCostKind != "" || o.CounterCostSelf || o.CounterCostLabel != "creatures you control" {
			t.Errorf("counter block = %+v", o)
		}
		if o.ManaCost != "{4}{G}{G}" || o.PayOptions != nil || o.SacrificeOptions != nil || o.TapOptions != nil {
			t.Errorf("the offer = %+v, want the printed mana and no card picker", o)
		}
		ids := map[string]int{}
		for _, opt := range o.CounterCostOptions {
			for _, k := range opt.Kinds {
				ids[opt.CardID] += k.Count
			}
		}
		if len(ids) != 2 || ids[a.String()] != 4 || ids[b.String()] != 3 {
			t.Errorf("counter options = %+v, want my two creatures and their counters", o.CounterCostOptions)
		}
	}
	if !found {
		t.Fatalf("the graveyard view does not offer the cast: %+v", v.AlternativeCosts)
	}
}

type qgMove struct {
	FromZone         string   `json:"from_zone"`
	AlternativeCost  string   `json:"alternative_cost"`
	CounterSourceIDs []string `json:"counter_source_ids"`
	CounterCounts    []int    `json:"counter_counts"`
	CounterKind      string   `json:"counter_kind"`
	CounterKinds     []string `json:"counter_kinds"`
}

// The bot is offered the graveyard cast with one payment of exactly six
// counters, priced on the move, and the payment it is offered is one the
// engine accepts.
func TestQuilledGreatwurmBotPath(t *testing.T) {
	g, me, _, wurm := qgTable(t)
	a := qgCreature(g, me.ID, "Bear A", map[string]int{game.CounterPlusOne: 4})
	b := qgCreature(g, me.ID, "Bear B", map[string]int{"oil": 3})

	var move *legal.Move
	var p qgMove
	for _, m := range legal.EnumerateFor(g, me.ID) {
		if m.Type != legal.TypeCastSpell || m.Source != wurm {
			continue
		}
		var mp qgMove
		if err := json.Unmarshal(m.Params, &mp); err != nil {
			t.Fatal(err)
		}
		if mp.FromZone != "graveyard" {
			continue
		}
		if move != nil {
			t.Fatalf("two graveyard casts enumerated: %s and %s", move.Label, m.Label)
		}
		m := m
		move, p = &m, mp
	}
	if move == nil {
		t.Fatal("the enumerator offers no graveyard cast")
	}
	if p.AlternativeCost != qgOfferKey || len(p.CounterSourceIDs) == 0 || len(p.CounterSourceIDs) != len(p.CounterCounts) {
		t.Fatalf("move params = %+v", p)
	}
	total := 0
	for _, n := range p.CounterCounts {
		total += n
	}
	if total != 6 {
		t.Errorf("the enumerated payment removes %d counters, want 6", total)
	}
	priced := 0
	if move.Cost != nil {
		for _, c := range move.Cost.Counters {
			priced += c.N
		}
	}
	if priced != 6 {
		t.Errorf("the move prices %d counters, want 6: %+v", priced, move.Cost)
	}

	params := game.CastSpellParams{FromZone: "graveyard", AlternativeCost: p.AlternativeCost, Strict: true,
		CounterCounts: p.CounterCounts, CounterKind: p.CounterKind, CounterKinds: p.CounterKinds}
	for _, raw := range p.CounterSourceIDs {
		params.CounterSourceIDs = append(params.CounterSourceIDs, uuid.MustParse(raw))
	}
	if err := g.CastSpell(me.ID, wurm, params); err != nil {
		t.Fatalf("the enumerated payment was refused: %v", err)
	}
	left := qgCounters(g, a)[game.CounterPlusOne] + qgCounters(g, b)["oil"]
	if left != 1 {
		t.Errorf("%d counters left of 7, want 1", left)
	}
}

// --- the combat trigger ---------------------------------------------

func qgGreatwurmOnBattlefield(g *game.Game, owner uuid.UUID) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{
		Name: "Quilled Greatwurm", OracleID: quilledGreatwurmOracle, TypeLine: "Creature — Wurm",
		Power: 7, Toughness: 7,
	})
}

// An unblocked creature you control grows by the damage it dealt.
func TestQuilledGreatwurmGrowsACreatureThatConnects(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	qgGreatwurmOnBattlefield(g, me.ID)
	bear := spm20Creature(g, me.ID, "Bear", "Creature — Bear", 3, 3)
	declareAttack(t, g, opp.ID, bear)
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if got := spm20Counters(g, bear); got != 3 {
		t.Errorf("the bear has %d +1/+1 counters, want 3", got)
	}
}

// A trampler's damage to its blocker and to the player is one trigger for
// the total (CR 510.2). Hardened Scales tells them apart: one trigger for
// 7 puts 8 counters on, two triggers for 5 and 2 would put 9.
func TestQuilledGreatwurmTrampleIsOneTriggerForTheTotal(t *testing.T) {
	for _, scales := range []bool{false, true} {
		name := "plain"
		if scales {
			name = "under Hardened Scales"
		}
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			wurm := apaPush(g, me.ID, me.ID, game.Card{
				Name: "Quilled Greatwurm", OracleID: quilledGreatwurmOracle, TypeLine: "Creature — Wurm",
				Power: 7, Toughness: 7, Keywords: []string{"trample"},
			})
			want := 7
			if scales {
				pushCatalogPermanent(g, me.ID, "Hardened Scales", "Enchantment", qgHardenedScalesOracle, false)
				want = 8
			}
			blocker := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 1, Toughness: 5})
			life := opp.Life
			spm20Combat(t, g, wurm, opp, []uuid.UUID{blocker}, 5, 2, false)
			passPriorityAroundTable(t, g)
			if got := spm20Counters(g, wurm); got != want {
				t.Errorf("the Greatwurm has %d +1/+1 counters, want %d", got, want)
			}
			if opp.Life != life-2 {
				t.Errorf("opponent life %d -> %d, want -2", life, opp.Life)
			}
		})
	}
}

const qgHardenedScalesOracle = "a1f3da21-af6d-450e-bf0b-985d158418e6"

// "It must survive": a creature that dies to the combat damage gets
// nothing, and the trigger resolves quietly.
func TestQuilledGreatwurmCreatureThatDiesGetsNothing(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	qgGreatwurmOnBattlefield(g, me.ID)
	bear := spm20Creature(g, me.ID, "Bear", "Creature — Bear", 2, 2)
	blocker := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Ogre", TypeLine: "Creature — Ogre", Power: 3, Toughness: 3})
	spm20Combat(t, g, bear, opp, []uuid.UUID{blocker}, 2, 0, false)
	passPriorityAroundTable(t, g)
	if findBattlefieldCardForTest(g, bear) != nil {
		t.Fatal("the bear survived a 3-power blocker")
	}
	if c, ok := g.LookupCardForEffect(bear); ok && c.Counters[game.CounterPlusOne] != 0 {
		t.Errorf("the dead bear got %d counters", c.Counters[game.CounterPlusOne])
	}
	if got := findBattlefieldCardForTest(g, blocker).Counters[game.CounterPlusOne]; got != 0 {
		t.Errorf("the opponent's blocker grew by %d", got)
	}
}

// "During your turn": your creature blocking on an opponent's turn deals
// combat damage and grows nothing; the attacker is not yours, so it grows
// nothing either.
func TestQuilledGreatwurmDoesNothingOnAnotherPlayersTurn(t *testing.T) {
	g := newCatalogGame(t)
	attackerSeat, defender := g.Seats[0], g.Seats[1]
	qgGreatwurmOnBattlefield(g, defender.ID)
	blocker := apaPush(g, defender.ID, defender.ID, game.Card{Name: "Wall", TypeLine: "Creature — Wall", Power: 2, Toughness: 5})
	bear := spm20Creature(g, attackerSeat.ID, "Bear", "Creature — Bear", 2, 3)
	spm20Combat(t, g, bear, defender, []uuid.UUID{blocker}, 2, 0, false)
	passPriorityAroundTable(t, g)
	if got := spm20Counters(g, blocker); got != 0 {
		t.Errorf("the blocker grew by %d on an opponent's turn", got)
	}
	if got := spm20Counters(g, bear); got != 0 {
		t.Errorf("the opponent's attacker grew by %d", got)
	}
}

// Register holds a cast's counter removal to the one shape the cast path
// pays: a "from among …" count, fixed or X of one kind with a floor (#5679),
// with no card component beside it.
func TestRegisterRefusesMalformedCastCounterRemoval(t *testing.T) {
	spec := func(name string, mut func(ac *game.AlternativeCost)) Spec {
		ac := CastFromGraveyardByRemovingCounters("{1}", 2, "creatures you control", Creature())
		mut(&ac)
		return Spec{OracleID: "fixture-5390-" + name, Name: "Fixture " + name,
			CastableZones: []game.ZoneKind{game.ZoneGraveyard}, AlternativeCosts: []game.AlternativeCost{ac}}
	}
	cases := []struct {
		name string
		mut  func(ac *game.AlternativeCost)
		want string
	}{
		{"self", func(ac *game.AlternativeCost) { ac.RemoveCounters.From = nil }, "removes counters from itself"},
		{"other", func(ac *game.AlternativeCost) { ac.RemoveCounters.From.ExcludeSource = true }, "names nothing"},
		{"single", func(ac *game.AlternativeCost) { ac.RemoveCounters.Among = false; ac.RemoveCounters.Counter = "+1/+1" }, "the only counter removal a cast pays"},
		{"x of any kind", func(ac *game.AlternativeCost) { ac.RemoveCounters.Variable = true }, "variable number of counters of any kind"},
		{"x from one permanent", func(ac *game.AlternativeCost) { ac.RemoveCounters.Variable, ac.RemoveCounters.OnePermanent = true, true }, "the only counter removal a cast pays"},
		{"any number", func(ac *game.AlternativeCost) {
			ac.RemoveCounters.Variable, ac.RemoveCounters.N, ac.RemoveCounters.Counter = true, 0, "+1/+1"
		}, "floor 1"},
		{"zero", func(ac *game.AlternativeCost) { ac.RemoveCounters.N = 0 }, "removes 0 counters"},
		{"beside a card", func(ac *game.AlternativeCost) { ac.DiscardFromHand = CardInYourHand("a card") }, "beside a card-shaped payment"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if msg := registerPanics(spec(strings.ReplaceAll(tc.name, " ", "-"), tc.mut)); !strings.Contains(msg, tc.want) {
				t.Errorf("registered, or refused for another reason: %q, want %q", msg, tc.want)
			}
		})
	}
}

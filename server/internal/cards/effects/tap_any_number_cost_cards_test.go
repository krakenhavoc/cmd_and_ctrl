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

// tap_any_number_cost_cards_test.go — #3961: "As an additional cost to
// cast this spell, [you may] tap any number of untapped creatures you
// control" (CR 601.2b / 601.2f / 601.2h). #3868's TapOthers cost with the
// count the caster's: the creatures ride cast_spell's teamwork_ids, none
// included, are validated as #3868's are and tapped with the spell on the
// stack. Burn at the Stake reads the count at resolution; Explosive
// Singularity's discount reads it at CR 601.2f, through the one pricer.

const (
	burnAtTheStakeOracle       = "eb23bbb0-dd12-44f1-8645-d92f46cbe8ff"
	explosiveSingularityOracle = "30d79f61-f9d9-441a-b600-83721af8553b"
)

func TestTapAnyNumberCostCardsAreFull(t *testing.T) {
	for _, oracle := range []string{burnAtTheStakeOracle, explosiveSingularityOracle} {
		spec, ok := Lookup(oracle)
		if !ok {
			t.Errorf("%s is not registered", oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
		ac := game.AdditionalCostFor(oracle)
		if ac == nil || !game.TapOthersAnyNumber(ac.TapOthers) || ac.TapOthers.Count != 0 {
			t.Errorf("%s additional cost %+v, want one tapping any number", spec.Name, ac)
		}
	}
}

// taoCreatures puts n untapped creatures onto `owner`'s battlefield.
func taoCreatures(g *game.Game, owner uuid.UUID, n int) []uuid.UUID {
	out := make([]uuid.UUID, n)
	for i := range out {
		out[i] = pr7Creature(g, owner, "Soldier", 1, "R")
	}
	return out
}

// Burn at the Stake: refused for a tapped creature, an opponent's
// creature, a noncreature and one creature named twice, with nothing
// paid; cast with none it deals no damage; cast with three (one arrived
// this turn) they are tapped with the spell on the stack and it deals
// nine, even after one of them has died in response.
func TestBurnAtTheStakeDealsThreeTimesTheCreaturesTapped(t *testing.T) {
	g, me, opp := tapTable(t)
	burn := handCardOf(g, me, "Burn at the Stake", "Sorcery", "{2}{R}{R}{R}", burnAtTheStakeOracle)
	team := taoCreatures(g, me.ID, 3)
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, team[0]).SummonedThisTurn = true })
	tapped := pr7Creature(g, me.ID, "Tapped", 1, "R")
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, tapped).Tapped = true })
	theirs := pr7Creature(g, opp.ID, "Theirs", 1, "G")
	rock := apaPush(g, me.ID, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	face := []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
	pay := func() {
		g.WithWriteLock(func() {
			fillPoolColored(me, "R", 3)
			fillPool(me, 2)
		})
	}
	pay()

	for name, ids := range map[string][]uuid.UUID{
		"a tapped creature":      {team[1], tapped},
		"an opponent's creature": {team[1], theirs},
		"an artifact":            {rock},
		"one creature twice":     {team[1], team[1]},
	} {
		if err := tacCast(g, me, burn, face, ids...); err == nil {
			t.Fatalf("%s paid Burn at the Stake", name)
		}
	}
	for _, id := range team {
		if tappedForTest(t, g, id) {
			t.Fatal("a refused cast tapped a creature")
		}
	}
	if !me.Hand.Contains(burn) {
		t.Fatal("a refused cast moved the spell")
	}

	life := opp.Life
	if err := tacCast(g, me, burn, face); err != nil {
		t.Fatalf("Burn at the Stake tapping none: %v", err)
	}
	passPriorityAroundTable(t, g)
	if opp.Life != life {
		t.Errorf("tapping none dealt %d damage, want 0", life-opp.Life)
	}

	again := handCardOf(g, me, "Burn at the Stake", "Sorcery", "{2}{R}{R}{R}", burnAtTheStakeOracle)
	pay()
	if err := tacCast(g, me, again, face, team...); err != nil {
		t.Fatalf("Burn at the Stake tapping three: %v", err)
	}
	if g.Stack.Size() != 1 {
		t.Fatalf("stack %d, want the spell alone", g.Stack.Size())
	}
	for _, id := range team {
		if !tappedForTest(t, g, id) {
			t.Error("a creature named to the cost was not tapped")
		}
	}
	if item := g.StackMeta[again]; item == nil || len(item.Paid.TappedOthers) != 3 {
		t.Fatalf("the payment does not record three creatures: %+v", g.StackMeta[again])
	}
	// A creature tapped this way that dies in response was still tapped
	// this way.
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(team[2]) })
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 9 {
		t.Errorf("tapping three dealt %d damage, want 9", got)
	}
}

// Explosive Singularity costs {1} less for each creature tapped: three
// creatures and {5}{R}{R} cast it, two creatures and the same mana do
// not, and it deals 10 damage.
func TestExplosiveSingularityCostsOneLessPerCreatureTapped(t *testing.T) {
	g, me, opp := tapTable(t)
	boom := handCardOf(g, me, "Explosive Singularity", "Sorcery", "{8}{R}{R}", explosiveSingularityOracle)
	team := taoCreatures(g, me.ID, 3)
	victim := pr7Creature(g, opp.ID, "Victim", 5, "G")
	g.WithWriteLock(func() {
		fillPoolColored(me, "R", 2)
		fillPool(me, 5)
	})
	if err := tacCast(g, me, boom, cardRefs(victim), team[:2]...); err == nil {
		t.Fatal("two creatures and seven mana paid {8}{R}{R}")
	}
	if tappedForTest(t, g, team[0]) {
		t.Fatal("the refused cast tapped a creature")
	}
	if err := tacCast(g, me, boom, cardRefs(victim), team...); err != nil {
		t.Fatalf("Explosive Singularity tapping three for {5}{R}{R}: %v", err)
	}
	if n := me.ManaPool.Total(); n != 0 {
		t.Errorf("%d mana left in the pool, want the whole {5}{R}{R} spent", n)
	}
	passPriorityAroundTable(t, g)
	if !opp.Graveyard.Contains(victim) {
		t.Error("10 damage did not destroy the 5/4")
	}
}

// The view ships the picker min 0, max the untapped creatures the caster
// controls, and the cast is offered with none (none is a legal payment):
// the cost never makes the spell uncastable.
func TestTapAnyNumberShipsMinZeroAndIsCastableWithNone(t *testing.T) {
	g, me, _ := tapTable(t)
	burn := handCardOf(g, me, "Burn at the Stake", "Sorcery", "{2}{R}{R}{R}", burnAtTheStakeOracle)
	for i := 0; i < 5; i++ {
		b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	tapOptions := func() *protocol.LegalTargetsView {
		v := protocol.ViewOfGameFor(g, me.ID.String())
		for _, s := range v.Seats {
			for i := range s.Hand.Cards {
				if c := &s.Hand.Cards[i]; c.InstanceID == burn.String() && c.AdditionalCost != nil {
					if c.AdditionalCost.TapOthersLabel != "any number of untapped creatures you control" {
						t.Errorf("tap_others_label %q", c.AdditionalCost.TapOthersLabel)
					}
					return c.AdditionalCost.TapOthersOptions
				}
			}
		}
		return nil
	}
	if to := tapOptions(); to == nil || to.Min != 0 || len(to.Cards) != 0 {
		t.Fatalf("with no creature, tap_others_options = %+v, want an empty min-0 list", to)
	}
	if len(burnCasts(t, g, me.ID, burn)) == 0 {
		t.Fatal("Burn at the Stake is not offered with no creature to tap")
	}
	team := taoCreatures(g, me.ID, 2)
	if to := tapOptions(); to == nil || to.Min != 0 || to.Max != 2 || len(to.Cards) != 2 {
		t.Fatalf("tap_others_options = %+v, want min 0, max 2 over the two creatures", to)
	}
	_ = team
}

// burnCasts is the enumerated casts of `card`, with each one's tap count.
func burnCasts(t *testing.T, g *game.Game, seat, card uuid.UUID) map[int][]legal.Move {
	t.Helper()
	out := map[int][]legal.Move{}
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type != legal.TypeCastSpell || m.Source != card {
			continue
		}
		var p struct {
			TeamworkIDs []string `json:"teamwork_ids"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("params %s: %v", m.Params, err)
		}
		out[len(p.TeamworkIDs)] = append(out[len(p.TeamworkIDs)], m)
	}
	return out
}

// The enumerator offers none, the small counts and every creature, each
// its own move: Burn at the Stake with five creatures is offered at 0, 1,
// 2, 3 and 5, and the five-creature cast at the opponent dispatches and
// deals 15.
func TestTheEnumeratorOffersBurnAtTheStakeAtEachCount(t *testing.T) {
	g, me, opp := tapTable(t)
	burn := handCardOf(g, me, "Burn at the Stake", "Sorcery", "{2}{R}{R}{R}", burnAtTheStakeOracle)
	for i := 0; i < 5; i++ {
		b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	team := taoCreatures(g, me.ID, 5)
	byCount := burnCasts(t, g, me.ID, burn)
	var counts []int
	for n := range byCount {
		counts = append(counts, n)
	}
	slices.Sort(counts)
	if !slices.Equal(counts, []int{0, 1, 2, 3, 5}) {
		t.Fatalf("tap counts offered %v, want [0 1 2 3 5]", counts)
	}
	var face *legal.Move
	for i, m := range byCount[5] {
		if strings.Contains(string(m.Params), opp.ID.String()) {
			face = &byCount[5][i]
			break
		}
	}
	if face == nil {
		t.Fatal("no five-creature cast at the opponent")
	}
	if !strings.Contains(face.Label, "tapping Soldier") {
		t.Errorf("move label %q does not name the tapped creatures", face.Label)
	}
	life := opp.Life
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(face.Type), Player: me.ID, Caller: me.ID, Params: face.Params}); err != nil {
		t.Fatalf("the enumerated cast is refused: %v", err)
	}
	for _, id := range team {
		if !tappedForTest(t, g, id) {
			t.Error("a creature the move named was not tapped")
		}
	}
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 15 {
		t.Errorf("five creatures dealt %d, want 15", got)
	}
}

// The enumerator prices each count with its own discount: with six
// Mountains and four creatures Explosive Singularity is affordable only
// with all four tapped, so that is the only count offered, and it
// dispatches.
func TestTheEnumeratorPricesExplosiveSingularityPerCount(t *testing.T) {
	g, me, opp := tapTable(t)
	boom := handCardOf(g, me, "Explosive Singularity", "Sorcery", "{8}{R}{R}", explosiveSingularityOracle)
	for i := 0; i < 6; i++ {
		b12Permanent(g, me.ID, "Mountain", "Basic Land — Mountain")
	}
	team := taoCreatures(g, me.ID, 4)
	byCount := burnCasts(t, g, me.ID, boom)
	if len(byCount) != 1 || len(byCount[4]) == 0 {
		var counts []int
		for n := range byCount {
			counts = append(counts, n)
		}
		t.Fatalf("tap counts offered %v, want only 4", counts)
	}
	var face *legal.Move
	for i, m := range byCount[4] {
		if strings.Contains(string(m.Params), opp.ID.String()) {
			face = &byCount[4][i]
			break
		}
	}
	if face == nil {
		t.Fatal("no cast at the opponent")
	}
	life := opp.Life
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(face.Type), Player: me.ID, Caller: me.ID, Params: face.Params}); err != nil {
		t.Fatalf("the enumerated cast is refused: %v", err)
	}
	for _, id := range team {
		if !tappedForTest(t, g, id) {
			t.Error("a creature the move named was not tapped")
		}
	}
	passPriorityAroundTable(t, g)
	if got := life - opp.Life; got != 10 {
		t.Errorf("Explosive Singularity dealt %d, want 10", got)
	}
}

// The auto-tap preview's price reads the count too: the pricer's
// TappingToPay is the length of teamwork_ids.
func TestExplosiveSingularityPriceReadsTheTapCount(t *testing.T) {
	g, me, _ := tapTable(t)
	boom := handCardOf(g, me, "Explosive Singularity", "Sorcery", "{8}{R}{R}", explosiveSingularityOracle)
	team := taoCreatures(g, me.ID, 3)
	var card game.Card
	g.WithWriteLock(func() {
		for _, c := range me.Hand.Cards {
			if c.InstanceID == boom {
				card = c
			}
		}
	})
	for n, want := range map[int]int{0: 8, 2: 6, 3: 5} {
		var price game.CastPrice
		var err error
		g.WithWriteLock(func() { price, err = g.PriceCastForEffect(me.ID, card, game.CastSpellParams{TeamworkIDs: team[:n]}) })
		if err != nil {
			t.Fatalf("price with %d tapped: %v", n, err)
		}
		if price.Total.Generic != want {
			t.Errorf("%d tapped: generic %d, want %d", n, price.Total.Generic, want)
		}
	}
}

// effects.Register keeps "any number" to a spell's mandatory slot, with
// no count beside it, and off activated abilities and alternative costs.
func TestRegisterRefusesMalformedTapAnyNumberCosts(t *testing.T) {
	counted := TapAnyNumberUntappedCost("any number of untapped creatures you control", Creature())
	counted.TapOthers.Count = 2
	branch := EitherCost(TapAnyNumberUntappedCost("any number of untapped artifacts you control", Artifact()).Keyed("tap"), DiscardCost(1).Keyed("discard"))
	anyTap := *TapAnUntapped("an untapped creature you control", Creature()).TapOthers
	anyTap.Count, anyTap.AnyNumber = 0, true
	for _, c := range []struct {
		name string
		spec Spec
		want string
	}{
		{"counted", Spec{AdditionalCost: counted}, "beside a count"},
		{"branch", Spec{AdditionalCost: branch}, "only the mandatory slot"},
		{"ability", Spec{Activated: []ActivatedAbility{{
			Label: "Tap any number of untapped creatures you control: Draw a card.",
			Cost:  game.AbilityCost{TapOthers: &anyTap},
		}}}, "only a spell's additional cost"},
	} {
		c.spec.OracleID, c.spec.Name = "test-3961-tap-"+c.name, "Test "+c.name
		if msg := registerPanics(c.spec); !strings.Contains(msg, c.want) {
			t.Errorf("%s: panic %q, want one mentioning %q", c.name, msg, c.want)
		}
	}
}

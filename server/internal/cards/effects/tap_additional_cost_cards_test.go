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

// tap_additional_cost_cards_test.go — #3868: "As an additional cost to
// cast this spell, tap an untapped Vampire you control" (CR 601.2h), the
// activated abilities' TapOthersCost on a cast's mandatory slot. The
// permanents ride cast_spell's teamwork_ids; they are untapped, yours,
// matched without targeting, with no summoning-sickness check (CR 302.6),
// and tapped with the spell on the stack.

const (
	newBloodOracle        = "021d654e-b731-40d0-a18e-eaff43f9fdb9"
	sharedDiscoveryOracle = "dd8e9f28-39da-4336-9259-0e34358749d9"
	gazeOfJusticeOracle   = "07c7e4d4-6452-437b-9f17-b0ec1fd9f702"
	swallowWholeOracle    = "54d4c632-1e7c-42f1-938c-9b4bd9b9a9b9"
	guardianGreatDoor     = "18ba355d-a25f-4a5d-b734-11739b0b86af"
	warlordsEliteOracle   = "ee2c8714-6e85-4eb5-9bd6-057b6abfb136"
	fearOfExposureOracle  = "abec56eb-7b3d-4bd8-93c7-bd7049c859c3"
	disruptionProtocol    = "4ca3306b-06f7-4a09-b904-f5540d11ddf1"
)

func TestTapAdditionalCostCardsAreFull(t *testing.T) {
	for _, c := range []struct {
		oracle string
		n      int
	}{
		{newBloodOracle, 1},
		{sharedDiscoveryOracle, 4},
		{gazeOfJusticeOracle, 3},
		{swallowWholeOracle, 1},
		{guardianGreatDoor, 4},
		{warlordsEliteOracle, 2},
		{fearOfExposureOracle, 2},
	} {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.oracle)
			continue
		}
		if spec.Completeness != CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
		ac := game.AdditionalCostFor(c.oracle)
		if ac == nil || ac.TapOthers == nil || ac.TapOthers.Count != c.n || !strings.HasPrefix(ac.Label, "Tap ") {
			t.Errorf("%s additional cost %+v, want one tapping %d", spec.Name, ac, c.n)
		}
	}
}

// tacVampire puts an untapped Vampire onto `owner`'s battlefield.
func tacVampire(g *game.Game, owner uuid.UUID, name string) uuid.UUID {
	return apaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Vampire", Power: 2, Toughness: 2, Colors: []string{"B"}})
}

func tacCast(g *game.Game, p *game.Player, card uuid.UUID, targets []game.TargetRef, taps ...uuid.UUID) error {
	return g.CastSpell(p.ID, card, game.CastSpellParams{Targets: targets, TeamworkIDs: taps, Strict: true})
}

// New Blood: refused with nothing named, a non-Vampire, a tapped Vampire,
// an opponent's Vampire and two Vampires, and nothing is paid by a refused
// cast; accepted for a Vampire that arrived this turn and has hexproof
// (not the {T} symbol, and a cost does not target), which is tapped with
// the spell on the stack and recorded on the payment. Resolved, the
// caster controls the creature and its chosen type is now Vampire;
// Vampire itself is not offered as the type to replace.
func TestNewBloodTapsAVampireToStealAndRetype(t *testing.T) {
	g, me, opp := tapTable(t)
	blood := handCardOf(g, me, "New Blood", "Sorcery", "{2}{B}{B}", newBloodOracle)
	fresh := tacVampire(g, me.ID, "Fresh")
	g.WithWriteLock(func() {
		c := findBattlefieldCardForTest(g, fresh)
		c.SummonedThisTurn = true
		c.Keywords = []string{"hexproof"}
	})
	tapped := tacVampire(g, me.ID, "Tapped")
	g.WithWriteLock(func() { findBattlefieldCardForTest(g, tapped).Tapped = true })
	other := tacVampire(g, me.ID, "Other")
	human := pr7Creature(g, me.ID, "Human", 2, "W")
	theirVamp := tacVampire(g, opp.ID, "Their Vampire")
	goblin := apaPush(g, opp.ID, opp.ID, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin Warrior", Power: 2, Toughness: 2, Colors: []string{"R"}})
	target := cardRefs(goblin)
	g.WithWriteLock(func() {
		fillPoolColored(me, "B", 2)
		fillPool(me, 2)
	})

	for name, ids := range map[string][]uuid.UUID{
		"nothing":               nil,
		"a non-Vampire":         {human},
		"a tapped Vampire":      {tapped},
		"an opponent's Vampire": {theirVamp},
		"two Vampires":          {fresh, other},
	} {
		if err := tacCast(g, me, blood, target, ids...); err == nil {
			t.Fatalf("%s paid New Blood", name)
		}
	}
	if tappedForTest(t, g, fresh) || tappedForTest(t, g, other) || tappedForTest(t, g, human) || !me.Hand.Contains(blood) {
		t.Fatal("a refused cast paid something")
	}
	if err := tacCast(g, me, blood, target, fresh); err != nil {
		t.Fatalf("New Blood for a Vampire that arrived this turn: %v", err)
	}
	if g.Stack.Size() != 1 || !tappedForTest(t, g, fresh) {
		t.Fatal("the Vampire was not tapped with New Blood on the stack")
	}
	if tappedForTest(t, g, other) {
		t.Error("another Vampire was tapped too")
	}
	if item := g.StackMeta[blood]; item == nil || len(item.Paid.TappedOthers) != 1 || item.Paid.TappedOthers[0].ID != fresh {
		t.Errorf("the payment does not record the tapped Vampire: %+v", g.StackMeta[blood])
	}

	passPriorityAroundTable(t, g)
	if got := controllerOf(t, g, goblin); got != me.ID {
		t.Errorf("the Goblin is controlled by %s, want the caster", got)
	}
	var offered []string
	for _, m := range legal.EnumerateFor(g, me.ID) {
		var p struct {
			CreatureType string `json:"creature_type"`
		}
		if json.Unmarshal(m.Params, &p) == nil && p.CreatureType != "" {
			offered = append(offered, p.CreatureType)
		}
	}
	if len(offered) == 0 || slices.Contains(offered, "Vampire") {
		t.Fatalf("offered %v as the type to replace; want a list without Vampire", offered)
	}
	answerCreatureType(t, g, me.ID, "Goblin")
	if subs := effectiveSubtypes(t, g, goblin); !slices.Equal(subs, []string{"Vampire", "Warrior"}) {
		t.Errorf("the creature is a %v now, want [Vampire Warrior]", subs)
	}
}

// The wire and the bot: New Blood ships tap_others_options (the untapped
// Vampires the caster controls, min = max = 1), the enumerator names one
// of them on teamwork_ids and the engine accepts that move, and with no
// untapped Vampire the picker is short and no cast is offered (CR 118.3).
func TestNewBloodShipsTapOptionsAndTheEnumeratorPaysIt(t *testing.T) {
	g, me, opp := tapTable(t)
	blood := handCardOf(g, me, "New Blood", "Sorcery", "{2}{B}{B}", newBloodOracle)
	for i := 0; i < 4; i++ {
		b12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	}
	apaPush(g, opp.ID, opp.ID, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Power: 2, Toughness: 2})
	pr7Creature(g, me.ID, "Human", 2, "W")

	tapOptions := func() *protocol.LegalTargetsView {
		v := protocol.ViewOfGameFor(g, me.ID.String())
		for _, s := range v.Seats {
			for i := range s.Hand.Cards {
				if c := &s.Hand.Cards[i]; c.InstanceID == blood.String() && c.AdditionalCost != nil {
					if c.AdditionalCost.TapOthersLabel != "an untapped Vampire you control" {
						t.Errorf("tap_others_label %q", c.AdditionalCost.TapOthersLabel)
					}
					return c.AdditionalCost.TapOthersOptions
				}
			}
		}
		return nil
	}
	type castMove struct {
		TeamworkIDs []string `json:"teamwork_ids"`
	}
	casts := func() []legal.Move {
		var out []legal.Move
		for _, m := range legal.EnumerateFor(g, me.ID) {
			if m.Type == legal.TypeCastSpell && m.Source == blood {
				out = append(out, m)
			}
		}
		return out
	}

	if to := tapOptions(); to == nil || to.Min != 1 || len(to.Cards) != 0 {
		t.Fatalf("with no Vampire, tap_others_options = %+v, want a short 1/1 list", to)
	}
	if ms := casts(); len(ms) != 0 {
		t.Fatalf("New Blood is offered with no Vampire to tap: %v", ms)
	}

	vamp := tacVampire(g, me.ID, "Vamp")
	if to := tapOptions(); to == nil || to.Min != 1 || to.Max != 1 || len(to.Cards) != 1 || to.Cards[0] != vamp.String() {
		t.Fatalf("tap_others_options = %+v, want 1/1 over the one Vampire", to)
	}
	ms := casts()
	if len(ms) == 0 {
		t.Fatal("the enumerator offers no New Blood cast")
	}
	var p castMove
	if err := json.Unmarshal(ms[0].Params, &p); err != nil || len(p.TeamworkIDs) != 1 || p.TeamworkIDs[0] != vamp.String() {
		t.Fatalf("enumerated tap payment %v (%v), want the Vampire", p.TeamworkIDs, err)
	}
	if !strings.Contains(ms[0].Label, "tapping Vamp") {
		t.Errorf("move label %q does not name the tapped Vampire", ms[0].Label)
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(ms[0].Type), Player: me.ID, Caller: me.ID, Params: ms[0].Params}); err != nil {
		t.Fatalf("the enumerated New Blood cast is refused: %v", err)
	}
	if !tappedForTest(t, g, vamp) {
		t.Error("the Vampire was not tapped")
	}
	tappedSwamps := 0
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Swamp" && c.Tapped {
			tappedSwamps++
		}
	}
	if tappedSwamps != 4 {
		t.Errorf("%d Swamps tapped for {2}{B}{B}, want 4", tappedSwamps)
	}
}

// Shared Discovery taps four creatures, not three, and draws three.
func TestSharedDiscoveryTapsFourCreaturesToDrawThree(t *testing.T) {
	g, me, _ := tapTable(t)
	disc := handCardOf(g, me, "Shared Discovery", "Sorcery", "{U}", sharedDiscoveryOracle)
	var team []uuid.UUID
	for i := 0; i < 4; i++ {
		team = append(team, pr7Creature(g, me.ID, "Ally", 1, "U"))
	}
	g.WithWriteLock(func() {
		fillPoolColored(me, "U", 1)
		for i := 0; i < 5; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
		}
	})
	if err := tacCast(g, me, disc, nil, team[:3]...); err == nil {
		t.Fatal("three creatures paid Shared Discovery")
	}
	if err := tacCast(g, me, disc, nil, team...); err != nil {
		t.Fatalf("Shared Discovery for four creatures: %v", err)
	}
	for _, id := range team {
		if !tappedForTest(t, g, id) {
			t.Error("a creature named to the cost was not tapped")
		}
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d after resolving, want %d", me.Hand.Size(), hand+3)
	}
}

// Gaze of Justice taps three WHITE creatures, and its flashback pays the
// additional cost too (CR 702.34a: "its flashback cost and any additional
// costs").
func TestGazeOfJusticeTapsWhiteCreaturesFromHandAndOnFlashback(t *testing.T) {
	g, me, opp := tapTable(t)
	gaze := handCardOf(g, me, "Gaze of Justice", "Sorcery", "{W}", gazeOfJusticeOracle)
	var whites []uuid.UUID
	for i := 0; i < 3; i++ {
		whites = append(whites, pr7Creature(g, me.ID, "Knight", 2, "W"))
	}
	green := pr7Creature(g, me.ID, "Elf", 1, "G")
	first := pr7Creature(g, opp.ID, "First", 3, "R")
	second := pr7Creature(g, opp.ID, "Second", 3, "R")
	g.WithWriteLock(func() { fillPoolColored(me, "W", 1) })
	if err := tacCast(g, me, gaze, cardRefs(first), whites[0], whites[1], green); err == nil {
		t.Fatal("a green creature paid Gaze of Justice")
	}
	if err := tacCast(g, me, gaze, cardRefs(first), whites...); err != nil {
		t.Fatalf("Gaze of Justice for three white creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !inExile(g, first) {
		t.Fatal("the target was not exiled")
	}
	if !me.Graveyard.Contains(gaze) {
		t.Fatal("Gaze of Justice is not in the graveyard")
	}
	g.WithWriteLock(func() {
		for _, id := range whites {
			findBattlefieldCardForTest(g, id).Tapped = false
		}
		fillPoolColored(me, "W", 1)
		fillPool(me, 5)
	})
	flash := func(taps ...uuid.UUID) error {
		return g.CastSpell(me.ID, gaze, game.CastSpellParams{
			AlternativeCost: "flashback", FromZone: "graveyard", Targets: cardRefs(second), TeamworkIDs: taps, Strict: true,
		})
	}
	if err := flash(); err == nil {
		t.Fatal("Gaze of Justice was flashed back without its additional cost")
	}
	if err := flash(whites...); err != nil {
		t.Fatalf("Gaze of Justice's flashback with three white creatures: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !inExile(g, second) || !inExile(g, gaze) {
		t.Error("the flashback did not exile its target and itself")
	}
}

// Swallow Whole exiles a tapped creature and puts the counter on the
// creature that paid, which was untapped when the target was chosen.
func TestSwallowWholeCountsTheCreatureThatPaid(t *testing.T) {
	g, me, opp := tapTable(t)
	swallow := handCardOf(g, me, "Swallow Whole", "Sorcery", "{W}", swallowWholeOracle)
	payer := pr7Creature(g, me.ID, "Payer", 2, "W")
	untapped := pr7Creature(g, opp.ID, "Untapped", 3, "R")
	prey := pr7Creature(g, opp.ID, "Prey", 3, "R")
	g.WithWriteLock(func() {
		findBattlefieldCardForTest(g, prey).Tapped = true
		fillPoolColored(me, "W", 1)
	})
	if err := tacCast(g, me, swallow, cardRefs(untapped), payer); err == nil {
		t.Fatal("Swallow Whole targeted an untapped creature")
	}
	if err := tacCast(g, me, swallow, cardRefs(prey), payer); err != nil {
		t.Fatalf("Swallow Whole: %v", err)
	}
	passPriorityAroundTable(t, g)
	if !inExile(g, prey) {
		t.Fatal("the tapped creature was not exiled")
	}
	if n := plusOneCounters(g, payer); n != 1 {
		t.Errorf("the creature that paid has %d +1/+1 counters, want 1", n)
	}
}

// The cast path's two lists stay apart: a creature can't pay the tap
// cost and convoke on one cast (CR 118.3).
func TestATapAdditionalCostCreatureCannotAlsoConvoke(t *testing.T) {
	const oracle = "test-3868-tap-and-convoke"
	registerForTest(t, Spec{OracleID: oracle, Name: "Test Tap And Convoke", Completeness: CompletenessFull,
		TapCost:        Convoke(),
		AdditionalCost: TapUntappedCost(1, "an untapped creature you control", Creature()),
	})
	g, me, _ := tapTable(t)
	spell := handCardOf(g, me, "Test Tap And Convoke", "Sorcery", "{1}", oracle)
	bear := pr7Creature(g, me.ID, "Bear", 2, "W")
	cast := func(convoke uuid.UUID) error {
		return g.CastSpell(me.ID, spell, game.CastSpellParams{
			TeamworkIDs: []uuid.UUID{bear}, TapIDs: []uuid.UUID{convoke}, Strict: true,
		})
	}
	if err := cast(bear); err == nil {
		t.Fatal("one creature paid a tap cost and convoke")
	}
	if tappedForTest(t, g, bear) {
		t.Error("the refused cast tapped the creature")
	}
	elf := pr7Creature(g, me.ID, "Elf", 1, "G")
	if err := cast(elf); err != nil {
		t.Fatalf("a second creature convoking the {1}: %v", err)
	}
	if !tappedForTest(t, g, bear) || !tappedForTest(t, g, elf) {
		t.Error("both creatures were not tapped")
	}
}

// effects.Register holds a tap additional cost to the shape the cast path
// reads: the mandatory slot or a branch of it, a fixed count, no
// "another", and never beside teamwork, which rides the same list.
func TestRegisterRefusesMalformedTapAdditionalCosts(t *testing.T) {
	another := TapUntappedCost(1, "another untapped creature you control", Creature())
	another.TapOthers.ExcludeSource = true
	xTap := TapUntappedCost(1, "X untapped creatures you control", Creature())
	xTap.TapOthers.Count = 0
	xTap.TapOthers.Filter.CountFromX = true
	optional := *TapUntappedCost(1, "an untapped creature you control", Creature())
	optional.Optional, optional.Key = true, "kicker"
	branchX := TapUntappedCost(1, "X untapped artifacts you control", Artifact())
	branchX.TapOthers.Count = 0
	branchX.TapOthers.Filter.CountFromX = true
	branch := EitherCost(branchX.Keyed("tap"), DiscardCost(1).Keyed("discard"))
	for _, c := range []struct {
		name string
		spec Spec
		want string
	}{
		{"another", Spec{AdditionalCost: another}, "another"},
		{"X", Spec{AdditionalCost: xTap}, "no fixed number"},
		{"optional", Spec{OptionalCosts: []game.AdditionalCost{optional}}, "only the mandatory slot"},
		{"branch X", Spec{AdditionalCost: branch}, "either/or branch 0 taps no fixed number"},
		{"teamwork", Spec{AdditionalCost: TapUntappedCost(1, "an untapped creature you control", Creature()),
			OptionalCosts: []game.AdditionalCost{Teamwork(2)}}, "teamwork"},
	} {
		c.spec.OracleID, c.spec.Name = "test-3868-tap-"+c.name, "Test "+c.name
		if msg := registerPanics(c.spec); !strings.Contains(msg, c.want) {
			t.Errorf("%s: panic %q, want one mentioning %q", c.name, msg, c.want)
		}
	}
}

// Disruption Protocol's either/or cost: branch 0 taps an untapped
// artifact (named on teamwork_ids, refused with nothing or with a
// creature), branch 1 pays {1}. The view marks the tap branch unpayable
// with no untapped artifact (CR 118.3), and the spell counters its target.
func TestDisruptionProtocolTapsAnArtifactOrPaysOne(t *testing.T) {
	g, me, _ := tapTable(t)
	if spec, ok := Lookup(disruptionProtocol); !ok || spec.Completeness != CompletenessFull {
		t.Fatalf("Disruption Protocol is not registered Full: %+v", spec)
	}
	proto := handCardOf(g, me, "Disruption Protocol", "Instant", "{U}{U}", disruptionProtocol)
	tapBranch := func() *protocol.AdditionalCostView {
		v := protocol.ViewOfGameFor(g, me.ID.String())
		for _, s := range v.Seats {
			for i := range s.Hand.Cards {
				if c := &s.Hand.Cards[i]; c.InstanceID == proto.String() && c.AdditionalCost != nil && len(c.AdditionalCost.Branches) == 2 {
					return &c.AdditionalCost.Branches[0]
				}
			}
		}
		t.Fatal("Disruption Protocol ships no two-branch additional cost")
		return nil
	}
	if b := tapBranch(); b.Payable || b.TapOthersOptions == nil {
		t.Errorf("with no artifact the tap branch is %+v, want unpayable with a short picker", b)
	}
	rock := apaPush(g, me.ID, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	bear := pr7Creature(g, me.ID, "Bear", 2, "U")
	if b := tapBranch(); !b.Payable {
		t.Error("the tap branch is not payable with an untapped artifact")
	}

	// Something to counter: the caster's own sorcery, still on the stack.
	victim := castCatalogSpell(t, g, "Ambition's Cost", "Sorcery", ambitionsCostOracle, nil)
	g.WithWriteLock(func() { fillPoolColored(me, "U", 2) })
	branch := 0
	cast := func(taps ...uuid.UUID) error {
		return g.CastSpell(me.ID, proto, game.CastSpellParams{
			Targets: cardRefs(victim), CostBranch: &branch, TeamworkIDs: taps, Strict: true,
		})
	}
	if err := cast(bear); err == nil {
		t.Fatal("a creature paid \"tap an untapped artifact\"")
	}
	if err := cast(); err == nil {
		t.Fatal("the tap branch was paid with nothing named")
	}
	if err := cast(rock); err != nil {
		t.Fatalf("Disruption Protocol tapping an artifact: %v", err)
	}
	if !tappedForTest(t, g, rock) {
		t.Error("the artifact was not tapped")
	}
	hand := me.Hand.Size()
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(victim) || me.Hand.Size() != hand {
		t.Errorf("the target was not countered (in graveyard %v, hand %d → %d)", me.Graveyard.Contains(victim), hand, me.Hand.Size())
	}
}

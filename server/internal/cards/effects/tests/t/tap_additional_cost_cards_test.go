package testst

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

func TestTapAdditionalCostCardsAreFull(t *testing.T) {
	for _, c := range []struct {
		oracle string
		n      int
	}{
		{effectstest.NewBloodOracle, 1},
		{effectstest.SharedDiscoveryOracle, 4},
		{effectstest.GazeOfJusticeOracle, 3},
		{effectstest.SwallowWholeOracle, 1},
		{effectstest.GuardianGreatDoor, 4},
		{effectstest.WarlordsEliteOracle, 2},
		{effectstest.FearOfExposureOracle, 2},
	} {
		spec, ok := effects.Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.oracle)
			continue
		}
		if spec.Completeness != effects.CompletenessFull {
			t.Errorf("%s: completeness %v, want Full", spec.Name, spec.Completeness)
		}
		ac := game.AdditionalCostFor(c.oracle)
		if ac == nil || ac.TapOthers == nil || ac.TapOthers.Count != c.n || !strings.HasPrefix(ac.Label, "Tap ") {
			t.Errorf("%s additional cost %+v, want one tapping %d", spec.Name, ac, c.n)
		}
	}
}

// New Blood: refused with nothing named, a non-Vampire, a tapped Vampire,
// an opponent's Vampire and two Vampires, and nothing is paid by a refused
// cast; accepted for a Vampire that arrived this turn and has hexproof
// (not the {T} symbol, and a cost does not target), which is tapped with
// the spell on the stack and recorded on the payment. Resolved, the
// caster controls the creature and its chosen type is now Vampire;
// Vampire itself is not offered as the type to replace.
func TestNewBloodTapsAVampireToStealAndRetype(t *testing.T) {
	g, me, opp := effectstest.TapTable(t)
	blood := effectstest.HandCardOf(g, me, "New Blood", "Sorcery", "{2}{B}{B}", effectstest.NewBloodOracle)
	fresh := effectstest.TacVampire(g, me.ID, "Fresh")
	g.WithWriteLock(func() {
		c := effectstest.FindBattlefieldCardForTest(g, fresh)
		c.SummonedThisTurn = true
		c.Keywords = []string{"hexproof"}
	})
	tapped := effectstest.TacVampire(g, me.ID, "Tapped")
	g.WithWriteLock(func() { effectstest.FindBattlefieldCardForTest(g, tapped).Tapped = true })
	other := effectstest.TacVampire(g, me.ID, "Other")
	human := effectstest.Pr7Creature(g, me.ID, "Human", 2, "W")
	theirVamp := effectstest.TacVampire(g, opp.ID, "Their Vampire")
	goblin := effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin Warrior", Power: 2, Toughness: 2, Colors: []string{"R"}})
	target := effectstest.CardRefs(goblin)
	g.WithWriteLock(func() {
		effectstest.FillPoolColored(me, "B", 2)
		effectstest.FillPool(me, 2)
	})

	for name, ids := range map[string][]uuid.UUID{
		"nothing":               nil,
		"a non-Vampire":         {human},
		"a tapped Vampire":      {tapped},
		"an opponent's Vampire": {theirVamp},
		"two Vampires":          {fresh, other},
	} {
		if err := effectstest.TacCast(g, me, blood, target, ids...); err == nil {
			t.Fatalf("%s paid New Blood", name)
		}
	}
	if effectstest.TappedForTest(t, g, fresh) || effectstest.TappedForTest(t, g, other) || effectstest.TappedForTest(t, g, human) || !me.Hand.Contains(blood) {
		t.Fatal("a refused cast paid something")
	}
	if err := effectstest.TacCast(g, me, blood, target, fresh); err != nil {
		t.Fatalf("New Blood for a Vampire that arrived this turn: %v", err)
	}
	if g.Stack.Size() != 1 || !effectstest.TappedForTest(t, g, fresh) {
		t.Fatal("the Vampire was not tapped with New Blood on the stack")
	}
	if effectstest.TappedForTest(t, g, other) {
		t.Error("another Vampire was tapped too")
	}
	if item := g.StackMeta[blood]; item == nil || len(item.Paid.TappedOthers) != 1 || item.Paid.TappedOthers[0].ID != fresh {
		t.Errorf("the payment does not record the tapped Vampire: %+v", g.StackMeta[blood])
	}

	effectstest.PassPriorityAroundTable(t, g)
	if got := effectstest.ControllerOf(t, g, goblin); got != me.ID {
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
	effectstest.AnswerCreatureType(t, g, me.ID, "Goblin")
	if subs := effectstest.EffectiveSubtypes(t, g, goblin); !slices.Equal(subs, []string{"Vampire", "Warrior"}) {
		t.Errorf("the creature is a %v now, want [Vampire Warrior]", subs)
	}
}

// The wire and the bot: New Blood ships tap_others_options (the untapped
// Vampires the caster controls, min = max = 1), the enumerator names one
// of them on teamwork_ids and the engine accepts that move, and with no
// untapped Vampire the picker is short and no cast is offered (CR 118.3).
func TestNewBloodShipsTapOptionsAndTheEnumeratorPaysIt(t *testing.T) {
	g, me, opp := effectstest.TapTable(t)
	blood := effectstest.HandCardOf(g, me, "New Blood", "Sorcery", "{2}{B}{B}", effectstest.NewBloodOracle)
	for i := 0; i < 4; i++ {
		effectstest.B12Permanent(g, me.ID, "Swamp", "Basic Land — Swamp")
	}
	effectstest.ApaPush(g, opp.ID, opp.ID, game.Card{Name: "Goblin", TypeLine: "Creature — Goblin", Power: 2, Toughness: 2})
	effectstest.Pr7Creature(g, me.ID, "Human", 2, "W")

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

	vamp := effectstest.TacVampire(g, me.ID, "Vamp")
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
	if !effectstest.TappedForTest(t, g, vamp) {
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
	g, me, _ := effectstest.TapTable(t)
	disc := effectstest.HandCardOf(g, me, "Shared Discovery", "Sorcery", "{U}", effectstest.SharedDiscoveryOracle)
	var team []uuid.UUID
	for i := 0; i < 4; i++ {
		team = append(team, effectstest.Pr7Creature(g, me.ID, "Ally", 1, "U"))
	}
	g.WithWriteLock(func() {
		effectstest.FillPoolColored(me, "U", 1)
		for i := 0; i < 5; i++ {
			me.Library.PushTop(game.Card{InstanceID: uuid.New(), Name: "Filler", TypeLine: "Land", Owner: me.ID, Controller: me.ID})
		}
	})
	if err := effectstest.TacCast(g, me, disc, nil, team[:3]...); err == nil {
		t.Fatal("three creatures paid Shared Discovery")
	}
	if err := effectstest.TacCast(g, me, disc, nil, team...); err != nil {
		t.Fatalf("Shared Discovery for four creatures: %v", err)
	}
	for _, id := range team {
		if !effectstest.TappedForTest(t, g, id) {
			t.Error("a creature named to the cost was not tapped")
		}
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	if me.Hand.Size() != hand+3 {
		t.Errorf("hand %d after resolving, want %d", me.Hand.Size(), hand+3)
	}
}

// Gaze of Justice taps three WHITE creatures, and its flashback pays the
// additional cost too (CR 702.34a: "its flashback cost and any additional
// costs").
func TestGazeOfJusticeTapsWhiteCreaturesFromHandAndOnFlashback(t *testing.T) {
	g, me, opp := effectstest.TapTable(t)
	gaze := effectstest.HandCardOf(g, me, "Gaze of Justice", "Sorcery", "{W}", effectstest.GazeOfJusticeOracle)
	var whites []uuid.UUID
	for i := 0; i < 3; i++ {
		whites = append(whites, effectstest.Pr7Creature(g, me.ID, "Knight", 2, "W"))
	}
	green := effectstest.Pr7Creature(g, me.ID, "Elf", 1, "G")
	first := effectstest.Pr7Creature(g, opp.ID, "First", 3, "R")
	second := effectstest.Pr7Creature(g, opp.ID, "Second", 3, "R")
	g.WithWriteLock(func() { effectstest.FillPoolColored(me, "W", 1) })
	if err := effectstest.TacCast(g, me, gaze, effectstest.CardRefs(first), whites[0], whites[1], green); err == nil {
		t.Fatal("a green creature paid Gaze of Justice")
	}
	if err := effectstest.TacCast(g, me, gaze, effectstest.CardRefs(first), whites...); err != nil {
		t.Fatalf("Gaze of Justice for three white creatures: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, first) {
		t.Fatal("the target was not exiled")
	}
	if !me.Graveyard.Contains(gaze) {
		t.Fatal("Gaze of Justice is not in the graveyard")
	}
	g.WithWriteLock(func() {
		for _, id := range whites {
			effectstest.FindBattlefieldCardForTest(g, id).Tapped = false
		}
		effectstest.FillPoolColored(me, "W", 1)
		effectstest.FillPool(me, 5)
	})
	flash := func(taps ...uuid.UUID) error {
		return g.CastSpell(me.ID, gaze, game.CastSpellParams{
			AlternativeCost: "flashback", FromZone: "graveyard", Targets: effectstest.CardRefs(second), TeamworkIDs: taps, Strict: true,
		})
	}
	if err := flash(); err == nil {
		t.Fatal("Gaze of Justice was flashed back without its additional cost")
	}
	if err := flash(whites...); err != nil {
		t.Fatalf("Gaze of Justice's flashback with three white creatures: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, second) || !effectstest.InExile(g, gaze) {
		t.Error("the flashback did not exile its target and itself")
	}
}

// Swallow Whole exiles a tapped creature and puts the counter on the
// creature that paid, which was untapped when the target was chosen.
func TestSwallowWholeCountsTheCreatureThatPaid(t *testing.T) {
	g, me, opp := effectstest.TapTable(t)
	swallow := effectstest.HandCardOf(g, me, "Swallow Whole", "Sorcery", "{W}", effectstest.SwallowWholeOracle)
	payer := effectstest.Pr7Creature(g, me.ID, "Payer", 2, "W")
	untapped := effectstest.Pr7Creature(g, opp.ID, "Untapped", 3, "R")
	prey := effectstest.Pr7Creature(g, opp.ID, "Prey", 3, "R")
	g.WithWriteLock(func() {
		effectstest.FindBattlefieldCardForTest(g, prey).Tapped = true
		effectstest.FillPoolColored(me, "W", 1)
	})
	if err := effectstest.TacCast(g, me, swallow, effectstest.CardRefs(untapped), payer); err == nil {
		t.Fatal("Swallow Whole targeted an untapped creature")
	}
	if err := effectstest.TacCast(g, me, swallow, effectstest.CardRefs(prey), payer); err != nil {
		t.Fatalf("Swallow Whole: %v", err)
	}
	effectstest.PassPriorityAroundTable(t, g)
	if !effectstest.InExile(g, prey) {
		t.Fatal("the tapped creature was not exiled")
	}
	if n := effectstest.PlusOneCounters(g, payer); n != 1 {
		t.Errorf("the creature that paid has %d +1/+1 counters, want 1", n)
	}
}

// The cast path's two lists stay apart: a creature can't pay the tap
// cost and convoke on one cast (CR 118.3).
func TestATapAdditionalCostCreatureCannotAlsoConvoke(t *testing.T) {
	const oracle = "test-3868-tap-and-convoke"
	effectstest.RegisterForTest(t, effects.Spec{OracleID: oracle, Name: "Test Tap And Convoke", Completeness: effects.CompletenessFull,
		TapCost:        effects.Convoke(),
		AdditionalCost: effects.TapUntappedCost(1, "an untapped creature you control", effects.Creature()),
	})
	g, me, _ := effectstest.TapTable(t)
	spell := effectstest.HandCardOf(g, me, "Test Tap And Convoke", "Sorcery", "{1}", oracle)
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "W")
	cast := func(convoke uuid.UUID) error {
		return g.CastSpell(me.ID, spell, game.CastSpellParams{
			TeamworkIDs: []uuid.UUID{bear}, TapIDs: []uuid.UUID{convoke}, Strict: true,
		})
	}
	if err := cast(bear); err == nil {
		t.Fatal("one creature paid a tap cost and convoke")
	}
	if effectstest.TappedForTest(t, g, bear) {
		t.Error("the refused cast tapped the creature")
	}
	elf := effectstest.Pr7Creature(g, me.ID, "Elf", 1, "G")
	if err := cast(elf); err != nil {
		t.Fatalf("a second creature convoking the {1}: %v", err)
	}
	if !effectstest.TappedForTest(t, g, bear) || !effectstest.TappedForTest(t, g, elf) {
		t.Error("both creatures were not tapped")
	}
}

// effects.Register holds a tap additional cost to the shape the cast path
// reads: the mandatory slot or a branch of it, a fixed count, no
// "another", and never beside teamwork, which rides the same list.
func TestRegisterRefusesMalformedTapAdditionalCosts(t *testing.T) {
	another := effects.TapUntappedCost(1, "another untapped creature you control", effects.Creature())
	another.TapOthers.ExcludeSource = true
	xTap := effects.TapUntappedCost(1, "X untapped creatures you control", effects.Creature())
	xTap.TapOthers.Count = 0
	xTap.TapOthers.Filter.CountFromX = true
	optional := *effects.TapUntappedCost(1, "an untapped creature you control", effects.Creature())
	optional.Optional, optional.Key = true, "buyback"
	branchX := effects.TapUntappedCost(1, "X untapped artifacts you control", effects.Artifact())
	branchX.TapOthers.Count = 0
	branchX.TapOthers.Filter.CountFromX = true
	branch := effects.EitherCost(branchX.Keyed("tap"), effects.DiscardCost(1).Keyed("discard"))
	for _, c := range []struct {
		name string
		spec effects.Spec
		want string
	}{
		{"another", effects.Spec{AdditionalCost: another}, "another"},
		{"X", effects.Spec{AdditionalCost: xTap}, "no fixed number"},
		{"optional", effects.Spec{OptionalCosts: []game.AdditionalCost{optional}}, "only the mandatory slot"},
		{"branch X", effects.Spec{AdditionalCost: branch}, "either/or branch 0 taps no fixed number"},
		{"teamwork", effects.Spec{AdditionalCost: effects.TapUntappedCost(1, "an untapped creature you control", effects.Creature()),
			OptionalCosts: []game.AdditionalCost{effects.Teamwork(2)}}, "teamwork"},
	} {
		c.spec.OracleID, c.spec.Name = "test-3868-tap-"+c.name, "Test "+c.name
		if msg := effectstest.RegisterPanics(c.spec); !strings.Contains(msg, c.want) {
			t.Errorf("%s: panic %q, want one mentioning %q", c.name, msg, c.want)
		}
	}
}

// Disruption Protocol's either/or cost: branch 0 taps an untapped
// artifact (named on teamwork_ids, refused with nothing or with a
// creature), branch 1 pays {1}. The view marks the tap branch unpayable
// with no untapped artifact (CR 118.3), and the spell counters its target.
func TestDisruptionProtocolTapsAnArtifactOrPaysOne(t *testing.T) {
	g, me, _ := effectstest.TapTable(t)
	if spec, ok := effects.Lookup(effectstest.DisruptionProtocol); !ok || spec.Completeness != effects.CompletenessFull {
		t.Fatalf("Disruption Protocol is not registered Full: %+v", spec)
	}
	proto := effectstest.HandCardOf(g, me, "Disruption Protocol", "Instant", "{U}{U}", effectstest.DisruptionProtocol)
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
	rock := effectstest.ApaPush(g, me.ID, me.ID, game.Card{Name: "Rock", TypeLine: "Artifact"})
	bear := effectstest.Pr7Creature(g, me.ID, "Bear", 2, "U")
	if b := tapBranch(); !b.Payable {
		t.Error("the tap branch is not payable with an untapped artifact")
	}

	// Something to counter: the caster's own sorcery, still on the stack.
	victim := effectstest.CastCatalogSpell(t, g, "Ambition's Cost", "Sorcery", effectstest.AmbitionsCostOracle, nil)
	g.WithWriteLock(func() { effectstest.FillPoolColored(me, "U", 2) })
	branch := 0
	cast := func(taps ...uuid.UUID) error {
		return g.CastSpell(me.ID, proto, game.CastSpellParams{
			Targets: effectstest.CardRefs(victim), CostBranch: &branch, TeamworkIDs: taps, Strict: true,
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
	if !effectstest.TappedForTest(t, g, rock) {
		t.Error("the artifact was not tapped")
	}
	hand := me.Hand.Size()
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.PassPriorityAroundTable(t, g)
	if !me.Graveyard.Contains(victim) || me.Hand.Size() != hand {
		t.Errorf("the target was not countered (in graveyard %v, hand %d → %d)", me.Graveyard.Contains(victim), hand, me.Hand.Size())
	}
}

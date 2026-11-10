package effects

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// tewCard is the double-faced card as the deck importer builds it,
// front face up.
func tewCard(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(),
		OracleID:   theEverflowingWellOracleID,
		Layout:     game.LayoutTransform,
		Owner:      owner,
		Controller: owner,
		Faces: []game.Face{
			{Name: "The Everflowing Well", TypeLine: "Legendary Artifact", ManaCost: "{2}{U}", Colors: []string{"U"}},
			{Name: "The Myriad Pools", TypeLine: "Legendary Artifact Land"},
		},
	}
	c.SetFace(0)
	return c
}

// tewPools puts The Myriad Pools on the battlefield, back face up.
func tewPools(g *game.Game, owner uuid.UUID) uuid.UUID {
	c := tewCard(owner)
	c.SetFace(1)
	return pushBattlefieldCardWithTimestamp(g, c)
}

// tewPoolsPick is the trigger's target prompt for `who`, or nil.
func tewPoolsPick(g *game.Game, who uuid.UUID) *game.PendingChoice {
	return latestChoiceOfKindFor(g, game.PendingChoicePickTarget, who)
}

func TestTheEverflowingWellIsRegisteredFull(t *testing.T) {
	for key, name := range map[string]string{
		theEverflowingWellOracleID:        "The Everflowing Well",
		theEverflowingWellOracleID + "#1": "The Myriad Pools",
	} {
		spec, ok := Lookup(key)
		if !ok {
			t.Fatalf("%s is not registered", name)
		}
		if spec.Name != name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

// The entry trigger mills two, then draws two.
func TestTheEverflowingWellMillsTwoThenDrawsTwo(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	me.Library.Cards = nil
	for _, n := range []string{"L1", "L2", "L3", "L4", "L5"} {
		ssCard(me.Library, me.ID, n, "Instant", 0, 0)
	}
	hand, yard := me.Hand.Size(), me.Graveyard.Size()

	castCatalogSpell(t, g, "The Everflowing Well", "Legendary Artifact", theEverflowingWellOracleID, nil)
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)

	if got := me.Graveyard.Size() - yard; got != 2 {
		t.Errorf("milled %d, want 2", got)
	}
	if got := me.Hand.Size() - hand; got != 2 {
		t.Errorf("drew %d, want 2", got)
	}
	if got := me.Library.Size(); got != 1 {
		t.Errorf("library %d, want 1 (two milled, two drawn)", got)
	}
}

// Descend 8: with eight permanent cards in your graveyard the upkeep
// trigger transforms it in place; with seven it does not trigger.
func TestTheEverflowingWellDescend8TransformsInUpkeep(t *testing.T) {
	for _, tc := range []struct {
		name       string
		permanents int
		transforms bool
	}{
		{"eight permanent cards", 8, true},
		{"seven permanent cards and an instant", 7, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[1]
			well := pushBattlefieldCardWithTimestamp(g, tewCard(me.ID))
			for i := 0; i < tc.permanents; i++ {
				ssCard(me.Graveyard, me.ID, "Dead Bear", "Creature — Bear", 2, 2)
			}
			ssCard(me.Graveyard, me.ID, "Spent Bolt", "Instant", 0, 0)

			advanceToUpkeepOf(t, g, 1)
			if got := triggersOnStackFrom(g, well) > 0; got != tc.transforms {
				t.Fatalf("triggered %v, want %v", got, tc.transforms)
			}
			passPriorityAroundTable(t, g)

			got, ok := battlefieldCard(g, well)
			if !ok {
				t.Fatal("the permanent stays on the battlefield")
			}
			if tc.transforms && (got.Name != "The Myriad Pools" || !got.IsLand()) {
				t.Errorf("it transformed into The Myriad Pools: %q %q", got.Name, got.TypeLine)
			}
			if !tc.transforms && got.Name != "The Everflowing Well" {
				t.Errorf("still the front face: %q", got.Name)
			}
		})
	}
}

// A permanent spell paid with the Pools' mana triggers; the chosen
// permanent becomes a copy of the spell before the spell resolves, and
// is itself again after the turn ends.
func TestTheMyriadPoolsTargetBecomesACopyOfTheSpellUntilEndOfTurn(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pools := tewPools(g, me.ID)
	bear := ssCard(g.Battlefield, me.ID, "Bear", "Creature — Bear", 2, 2)
	activateManaFor(t, g, me.ID, pools, 0, game.ManaAbilityParams{})

	spell := castFromHandForTest(t, g, me, "Wall of Water", "Artifact Creature — Wall", "{U}", "", strict)
	pick := tewPoolsPick(g, me.ID)
	if pick == nil {
		t.Fatal("the trigger asks for its target as it goes on the stack")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range pick.PickTargetCards {
		offered[id] = true
	}
	if !offered[bear] || offered[pools] {
		t.Fatalf("offered %v: want the Bear and not The Myriad Pools itself", offered)
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	resolveTopOfStack(t, g)
	if !g.Stack.Contains(spell) {
		t.Fatal("the trigger resolves before the spell")
	}
	got, _ := battlefieldCard(g, bear)
	if got.Name != "Wall of Water" || got.TypeLine != "Artifact Creature — Wall" {
		t.Fatalf("the Bear became a copy of the spell: %q %q", got.Name, got.TypeLine)
	}
	passPriorityAroundTable(t, g)

	advanceToUpkeepOf(t, g, (g.Turn.ActiveSeat+1)%len(g.Seats))
	got, _ = battlefieldCard(g, bear)
	if got.Name != "Bear" || got.TypeLine != "Creature — Bear" {
		t.Errorf("the copy ends with the turn: %q %q", got.Name, got.TypeLine)
	}
}

// The spell countered in response is still copied, from its last-known
// information (CR 608.2h).
func TestTheMyriadPoolsCopiesACounteredSpellFromLastKnownInformation(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pools := tewPools(g, me.ID)
	bear := ssCard(g.Battlefield, me.ID, "Bear", "Creature — Bear", 2, 2)
	activateManaFor(t, g, me.ID, pools, 0, game.ManaAbilityParams{})

	spell := castFromHandForTest(t, g, me, "Glass Sphinx", "Artifact Creature — Sphinx", "{U}", "", strict)
	pick := tewPoolsPick(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTarget(pick.ID, me.ID, game.TargetRef{Kind: game.TargetCard, ID: bear}); err != nil {
		t.Fatalf("ResolvePickTarget: %v", err)
	}
	if counterByEffect(t, g, spell) {
		t.Fatal("setup: the spell is countered")
	}
	passPriorityAroundTable(t, g)
	got, _ := battlefieldCard(g, bear)
	if got.Name != "Glass Sphinx" {
		t.Errorf("the countered spell is still copied: %q", got.Name)
	}
}

// Up to one: choosing no target is legal, and the trigger does nothing.
func TestTheMyriadPoolsUpToOneAllowsNoTarget(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pools := tewPools(g, me.ID)
	bear := ssCard(g.Battlefield, me.ID, "Bear", "Creature — Bear", 2, 2)
	activateManaFor(t, g, me.ID, pools, 0, game.ManaAbilityParams{})

	spell := castFromHandForTest(t, g, me, "Wall of Water", "Artifact Creature — Wall", "{U}", "", strict)
	pick := tewPoolsPick(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}
	if err := g.ResolvePickTargets(pick.ID, me.ID, nil); err != nil {
		t.Fatalf("no target is a legal answer: %v", err)
	}
	passPriorityAroundTable(t, g)
	passPriorityAroundTable(t, g)
	if got, _ := battlefieldCard(g, bear); got.Name != "Bear" {
		t.Errorf("nothing was chosen, so nothing changes: %q", got.Name)
	}
	if !g.Battlefield.Contains(spell) {
		t.Error("the spell still resolves")
	}
}

// Only the Pools' mana spent on a permanent spell triggers it: an
// instant paid with it, and a permanent spell paid with other mana, do
// not.
func TestTheMyriadPoolsTriggersOnlyOnAPermanentSpellItsManaPays(t *testing.T) {
	for _, tc := range []struct {
		name, typeLine string
		poolsMana      bool
	}{
		{"an instant paid with the Pools", "Instant", true},
		{"a creature paid with other mana", "Creature — Merfolk", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			advanceToMain(t, g)
			pools := tewPools(g, me.ID)
			ssCard(g.Battlefield, me.ID, "Bear", "Creature — Bear", 2, 2)
			if tc.poolsMana {
				activateManaFor(t, g, me.ID, pools, 0, game.ManaAbilityParams{})
			} else {
				floatForTest(g, me, "U")
			}
			castFromHandForTest(t, g, me, "Spell", tc.typeLine, "{U}", "", strict)
			if tewPoolsPick(g, me.ID) != nil || triggersOnStackFrom(g, pools) != 0 {
				t.Fatal("the Pools triggered")
			}
		})
	}
}

// The bot path: the enumerator offers the target prompt's answers, the
// Pools itself is not among them, and an enumerated answer is accepted.
func TestTheMyriadPoolsEnumeratorAnswersTheTargetPrompt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceToMain(t, g)
	pools := tewPools(g, me.ID)
	bear := ssCard(g.Battlefield, me.ID, "Bear", "Creature — Bear", 2, 2)
	activateManaFor(t, g, me.ID, pools, 0, game.ManaAbilityParams{})
	castFromHandForTest(t, g, me, "Wall of Water", "Artifact Creature — Wall", "{U}", "", strict)
	pick := tewPoolsPick(g, me.ID)
	if pick == nil {
		t.Fatal("no target prompt")
	}

	var bearMove *legal.Move
	answers := 0
	moves := legal.EnumerateFor(g, me.ID)
	for i, m := range moves {
		if m.Type != legal.TypeResolveChoice {
			continue
		}
		var p struct {
			ChoiceID string `json:"choice_id"`
			Target   *struct {
				ID string `json:"id"`
			} `json:"target"`
		}
		if err := json.Unmarshal(m.Params, &p); err != nil || p.ChoiceID != pick.ID.String() {
			continue
		}
		answers++
		if p.Target == nil {
			continue
		}
		switch uuid.MustParse(p.Target.ID) {
		case bear:
			bearMove = &moves[i]
		case pools:
			t.Error("the enumerator offers The Myriad Pools as its own target")
		}
	}
	if bearMove == nil {
		t.Fatalf("the enumerator offers no answer naming the Bear (%d answers)", answers)
	}
	if err := actions.Dispatch(g, actions.Action{Type: actions.Type(bearMove.Type), Player: me.ID, Caller: me.ID, Params: bearMove.Params}); err != nil {
		t.Fatalf("the enumerated answer is refused: %v", err)
	}
	passPriorityAroundTable(t, g)
	if got, _ := battlefieldCard(g, bear); got.Name != "Wall of Water" {
		t.Errorf("the enumerated answer took effect: %q", got.Name)
	}
}

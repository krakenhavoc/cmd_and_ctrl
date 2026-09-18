package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// modal_clauses_view_test.go — #764 / ADR 0065 §7: what the wire says
// about a multi-clause statement, a repeatable mode clause, a
// mode_pick prompt and a modal item on the stack.

// TestClausesStampedForAMultiClauseCard — the per-clause legal sets a
// two-step picker walks. `legal_targets` stays clause 0, so every
// pre-#764 client path keeps working; `clauses` is the whole list and
// is absent for the single-clause card that is nearly every card.
func TestClausesStampedForAMultiClauseCard(t *testing.T) {
	g := buildActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	const oracle = "test-view-two-clauses"

	prev := game.CatalogTargetSpec
	game.CatalogTargetSpec = func(id string) *game.TargetSpec {
		if id != oracle {
			return nil
		}
		mine := &game.TargetSpec{
			Mode: "creature", Label: "target creature you control",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(_ *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				return c.IsCreature() && c.Controller == caster
			},
			Min: 1, Max: 1,
		}
		theirs := game.TargetClause{
			Mode: "creature", Label: "target creature you don't control",
			Zones: []game.ZoneKind{game.ZoneBattlefield},
			CardOK: func(_ *game.Game, caster uuid.UUID, c game.Card, _ game.ZoneKind) bool {
				return c.IsCreature() && c.Controller != caster
			},
			Min: 1, Max: 1, Distinct: true,
		}
		mine.Rest = []game.TargetClause{theirs}
		return mine
	}
	t.Cleanup(func() { game.CatalogTargetSpec = prev })

	myBear := game.NewCard("My Bear", me.ID)
	myBear.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(myBear)
	theirBear := game.NewCard("Their Bear", opp.ID)
	theirBear.TypeLine = "Creature — Bear"
	g.Battlefield.PushTop(theirBear)

	spell := game.NewCard("Bite", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = oracle
	spell.KnownBy = map[uuid.UUID]bool{me.ID: true, opp.ID: true}
	me.Hand.PushTop(spell)

	mine := ViewOfGameFor(g, me.ID.String())
	var found *CardView
	for i := range mine.Seats[0].Hand.Cards {
		if mine.Seats[0].Hand.Cards[i].InstanceID == spell.InstanceID.String() {
			found = &mine.Seats[0].Hand.Cards[i]
		}
	}
	if found == nil {
		t.Fatal("the spell is in the owner's hand view")
	}
	if len(found.Clauses) != 2 {
		t.Fatalf("two clauses on the wire: %+v", found.Clauses)
	}
	if found.Clauses[0].Label != "target creature you control" {
		t.Errorf("clause 0 carries its printed wording: %q", found.Clauses[0].Label)
	}
	if len(found.Clauses[0].Cards) != 1 || found.Clauses[0].Cards[0] != myBear.InstanceID.String() {
		t.Errorf("clause 0's legal set is ITS predicate: %v", found.Clauses[0].Cards)
	}
	if len(found.Clauses[1].Cards) != 1 || found.Clauses[1].Cards[0] != theirBear.InstanceID.String() {
		t.Errorf("clause 1's legal set is ITS predicate: %v", found.Clauses[1].Cards)
	}
	if !found.Clauses[1].Distinct {
		t.Error("clause 1 is marked distinct, so the picker can grey what is taken")
	}
	// The compatibility half: legal_targets is still clause 0.
	if found.LegalTargets == nil || len(found.LegalTargets.Cards) != 1 ||
		found.LegalTargets.Cards[0] != myBear.InstanceID.String() {
		t.Errorf("legal_targets stays clause 0: %+v", found.LegalTargets)
	}

	// And an opponent sees neither.
	theirs := ViewOfGameFor(g, opp.ID.String())
	for _, c := range theirs.Seats[0].Hand.Cards {
		if c.InstanceID == spell.InstanceID.String() && (c.Clauses != nil || c.LegalTargets != nil) {
			t.Error("an opponent's view of a hand card carries no clause data")
		}
	}
}

func TestSingleClauseCardShipsNoClauseList(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-view-one-clause"
	prev := game.CatalogTargetSpec
	game.CatalogTargetSpec = func(id string) *game.TargetSpec {
		if id != oracle {
			return nil
		}
		return &game.TargetSpec{
			Mode: "creature", Zones: []game.ZoneKind{game.ZoneBattlefield},
			Min: 1, Max: 1,
		}
	}
	t.Cleanup(func() { game.CatalogTargetSpec = prev })

	spell := game.NewCard("Bolt", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = oracle
	spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(spell)

	mine := ViewOfGameFor(g, me.ID.String())
	for _, c := range mine.Seats[0].Hand.Cards {
		if c.InstanceID == spell.InstanceID.String() && c.Clauses != nil {
			t.Error("one clause: the list is omitted and legal_targets alone says everything")
		}
	}
}

// CR 700.2d rides the wire so the picker offers a count per bullet
// rather than a toggle.
func TestRepeatableModeSpecReachesTheWire(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	const oracle = "test-view-repeatable"
	prev := game.CatalogModeSpec
	game.CatalogModeSpec = func(id string) *game.ModeSpec {
		if id != oracle {
			return nil
		}
		return &game.ModeSpec{
			Prompt: "Choose three", Min: 3, Max: 3, Repeatable: true,
			Options: []game.ModeOption{{Label: "Draw a card."}},
		}
	}
	t.Cleanup(func() { game.CatalogModeSpec = prev })

	spell := game.NewCard("Confluence", me.ID)
	spell.TypeLine = "Instant"
	spell.OracleID = oracle
	spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(spell)

	mine := ViewOfGameFor(g, me.ID.String())
	for _, c := range mine.Seats[0].Hand.Cards {
		if c.InstanceID != spell.InstanceID.String() {
			continue
		}
		if c.Modes == nil || !c.Modes.Repeatable {
			t.Fatalf("repeatable on the wire: %+v", c.Modes)
		}
		if c.Modes.Min != 3 || c.Modes.Max != 3 {
			t.Errorf("bounds: %+v", c.Modes)
		}
	}
}

// The stack overlay reads the chosen bullets, not their indexes.
func TestStackItemCarriesModeLabelsAndSlottedTargets(t *testing.T) {
	item := &game.StackItem{
		ID: uuid.New(), Kind: game.StackItemSpell,
		Controller: uuid.New(), Owner: uuid.New(), SourceCardID: uuid.New(),
		Modes: []int{1, 1},
		Targets: []game.TargetRef{
			{Kind: game.TargetCard, ID: uuid.New(), Mode: 0, Slot: 0},
			{Kind: game.TargetCard, ID: uuid.New(), Mode: 1, Slot: 0},
		},
	}
	v := viewOfStackItem(item)
	if len(v.Targets) != 2 {
		t.Fatalf("both picks: %+v", v.Targets)
	}
	if v.Targets[0].Mode != 0 || v.Targets[1].Mode != 1 {
		t.Errorf("each pick names its mode occurrence: %+v", v.Targets)
	}
	// No ModeSpec stamped on the item, so no labels — the field is
	// omitted rather than guessed.
	if len(v.ModeLabels) != 0 {
		t.Errorf("no spec, no labels: %v", v.ModeLabels)
	}
}

// The mode_pick prompt's offer reaches the chooser whole.
func TestModePickPromptProjection(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind:            game.PendingChoiceModePick,
			Chooser:         me.ID,
			Count:           1,
			Reason:          "Choose one",
			ModeOptionIndex: []int{0, 2},
			ModeOptionLabel: []string{"Draw a card.", "You gain 2 life."},
			ModeMin:         1,
			ModeMax:         1,
		})
	})
	v := ViewOfGameFor(g, me.ID.String())
	var pc *PendingChoiceView
	for i := range v.PendingChoices {
		if v.PendingChoices[i].Kind == string(game.PendingChoiceModePick) {
			pc = &v.PendingChoices[i]
		}
	}
	if pc == nil {
		t.Fatal("the prompt reaches its chooser")
	}
	if len(pc.ModeOptions) != 2 || pc.ModeOptions[1] != "You gain 2 life." {
		t.Errorf("the bullets ride verbatim: %v", pc.ModeOptions)
	}
	if len(pc.ModeIndexes) != 2 || pc.ModeIndexes[1] != 2 {
		t.Errorf("the ModeSpec index each label belongs to: %v", pc.ModeIndexes)
	}
	if pc.ModeMin != 1 || pc.ModeMax != 1 || pc.ModeRepeatable {
		t.Errorf("bounds: %+v", pc)
	}
}

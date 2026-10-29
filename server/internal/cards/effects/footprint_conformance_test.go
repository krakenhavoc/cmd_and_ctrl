package effects

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// footprint_conformance_test.go — ADR 0018's #3013 amendment, sub-PR 1.
//
// A row's footprint is derived from the primitives its Do runs
// (footprint.go), so it cannot drift from the card's code. What can go
// wrong is a primitive's footprint() saying less than its Apply does: a
// step that changes a player or an object, or emits an event kind, that
// its declaration leaves out would let the #2884 check skip a prompt on
// a batch whose order shows. This test resolves every primitive that
// declares a footprint on a real board and fails on anything it changes
// or emits outside what game.FootprintDeclarationForTest — the check's
// own reading of the step — declares.

// conformanceCase is one footprinter, resolved once.
type conformanceCase struct {
	step Applier
	// target puts an opponent's vanilla creature on the item as its
	// one card target (ExileChosenTarget).
	target bool
	// subject gives the item a triggering event about a 2/3 creature of
	// its controller's that just entered (GainLifeEqualToSubject).
	subject bool
	// targetPlayer puts an opponent on the item as its one player
	// target (TargetPlayerLosesLife).
	targetPlayer bool
	// departed destroys the item's source before the probe, so a damage
	// step deals its damage as the source last existed (CR 608.2h).
	departed bool
	// movedOn then exiles it from the graveyard, so no last-known record
	// matches the card any more (a commander gone to the command zone).
	movedOn bool
	// setup arranges the board before the probe, under the write lock:
	// an If's condition made to hold (#3033).
	setup func(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID)
	// x is the item's XValue, as XFromCast stamps it (CreateXByXToken).
	x int
}

// ifCase is an If case: its condition is made to hold by setup, and its
// Then step must emit something for the probe to prove anything.
func ifCase(cond Condition, then Applier, setup func(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID)) conformanceCase {
	return conformanceCase{step: If{Cond: cond, Then: []Applier{then}}, setup: setup}
}

// footprintConformanceCases has one case per type that implements
// footprinter, keyed by the type's name.
// TestEveryFootprinterHasAConformanceCase keeps the two lists equal.
var footprintConformanceCases = map[string]conformanceCase{
	"GainLife":           {step: GainLife{Amount: 3}},
	"DrawCards":          {step: DrawCards{N: 2}},
	"MillCards":          {step: MillCards{N: 2}},
	"Scry":               {step: Scry{N: 2}},
	"Surveil":            {step: Surveil{N: 2}},
	"GetEnergy":          {step: GetEnergy{N: 2}},
	"CreateToken":        {step: CreateToken{Template: TokenCard("1/1 green Elf Warrior"), N: 2}},
	"BecomeTheMonarch":   {step: BecomeTheMonarch{}},
	"CounterOnThis":      {step: CounterOnThis{Kind: game.CounterPlusOne, N: 2}},
	"DamageEachOpponent": {step: DamageEachOpponent{N: 2}},
	"DamageEachOpponent:departed": {
		step: DamageEachOpponent{N: 2}, departed: true,
	},
	"DamageController": {step: DamageController{N: 2}},
	"DamageEachOpponent:moved-on": {
		step: DamageEachOpponent{N: 2}, departed: true, movedOn: true,
	},
	"DamageController:departed": {step: DamageController{N: 2}, departed: true},
	"TargetPlayerLosesLife":     {step: TargetPlayerLosesLife{N: 2}, targetPlayer: true},
	"EachOpponentLosesLife":     {step: EachOpponentLosesLife{N: 2}},
	"ExileChosenTarget":         {step: ExileChosenTarget{}, target: true},
	"GainLifeEqualToSubject": {
		step: GainLifeEqualToSubject{Stat: SubjectToughness}, subject: true,
	},
	// A variant of a type is keyed "<Type>:<variant>".
	"GainLifeEqualToSubject:power": {
		step: GainLifeEqualToSubject{Stat: SubjectPower}, subject: true,
	},
	"CreateTokenAdvanced:tapped": {
		step: CreateTokenAdvanced{Spec: Token(TreasureToken()).EntersTapped(), N: 2},
	},
	// An intervening "if" (#3033), one case per Condition type
	// (TestEveryConditionHasAConformanceCase). The condition reads and
	// writes nothing, so each case proves the If declares its Then.
	"If:life-gained": ifCase(CondLifeGained{AtLeast: 3}, DrawCards{N: 1},
		func(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID) {
			if err := g.ChangePlayerLifeForEffect(src, me.ID, 3); err != nil {
				t.Fatal(err)
			}
		}),
	"If:life-lost": ifCase(CondLifeLost{}, GainLife{Amount: 2},
		func(t *testing.T, g *game.Game, me *game.Player, src uuid.UUID) {
			if err := g.ChangePlayerLifeForEffect(src, me.ID, -1); err != nil {
				t.Fatal(err)
			}
		}),
	"If:cards-drawn": ifCase(CondCardsDrawn{AtLeast: 2}, GainLife{Amount: 2},
		func(t *testing.T, g *game.Game, me *game.Player, _ uuid.UUID) {
			if err := g.DrawNForEffect(me.ID, 2); err != nil {
				t.Fatal(err)
			}
		}),
	"If:creature-died": ifCase(CondCreatureDied{}, MillCards{N: 2},
		func(t *testing.T, g *game.Game, me *game.Player, _ uuid.UUID) {
			if err := g.DestroyPermanentForEffect(pushBattlefieldCreature(g, me.ID)); err != nil {
				t.Fatal(err)
			}
		}),
	"If:citys-blessing": ifCase(CondCitysBlessing{}, CounterOnThis{Kind: game.CounterPlusOne, N: 1},
		func(_ *testing.T, _ *game.Game, me *game.Player, _ uuid.UUID) { me.CitysBlessing = true }),
	"If:no-spells-cast": ifCase(CondNoSpellsCast{}, CounterOnThis{Kind: game.CounterPlusOne, N: 2}, nil),
	"If:monarch": ifCase(CondMonarch{}, CreateToken{Template: TokenCard("1/1 green Elf Warrior"), N: 1},
		func(_ *testing.T, g *game.Game, me *game.Player, _ uuid.UUID) { g.Monarch = me.ID }),
	"If:source": ifCase(CondFewerCountersOnThis{Kind: game.CounterPlusOne, N: 3}, CounterOnThis{Kind: game.CounterPlusOne, N: 1}, nil),
	"If:power":  ifCase(CondPowerAtLeast{N: 2}, DamageEachOpponent{N: 2}, nil),
	"If:board": ifCase(CondBoard{Name: "a creature", Test: func(ctx *Context) bool {
		return len(ctx.Game.Battlefield.Cards) > 0
	}}, EachOpponentLosesLife{N: 2}, nil),
	// Counted and X amounts (#3033).
	"CreateXByXToken": {step: CreateXByXToken{Template: game.Card{Name: "Treefolk", TypeLine: "Token Creature — Treefolk"}}, x: 3},
	// ADR 0208: Goblin Airbrusher's count reads the triggering placement.
	"CreateTokensForPlacedSticker": {step: CreateTokensForPlacedSticker{Template: TokenCard("1/1 green Elf Warrior"), N: 1, ArtN: 2}},
	"CreateTokensForEach": {step: CreateTokensForEach{
		Template: TokenCard("1/1 white Rabbit"), Name: "creature",
		Count: func(ctx *Context) int { return len(ctx.Game.Battlefield.Cards) },
	}},
	"If:nested": {step: If{Cond: CondNoSpellsCast{}, Then: []Applier{
		GainLife{Amount: 1},
		If{Cond: CondMonarch{}, Then: []Applier{DrawCards{N: 1}}},
	}}, setup: func(_ *testing.T, g *game.Game, me *game.Player, _ uuid.UUID) { g.Monarch = me.ID }},
}

// footprinterTypes lists the receiver types of every footprint() method
// in this package's non-test files.
func footprinterTypes(t *testing.T) []string {
	t.Helper()
	fset := token.NewFileSet()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, name, nil, parser.SkipObjectResolution)
		if err != nil {
			t.Fatal(err)
		}
		for _, d := range f.Decls {
			fn, ok := d.(*ast.FuncDecl)
			if !ok || fn.Recv == nil || (fn.Name.Name != "footprint" && fn.Name.Name != "footprints") || len(fn.Recv.List) != 1 {
				continue
			}
			typ := fn.Recv.List[0].Type
			if star, ok := typ.(*ast.StarExpr); ok {
				typ = star.X
			}
			if id, ok := typ.(*ast.Ident); ok {
				out = append(out, id.Name)
			}
		}
	}
	sort.Strings(out)
	return out
}

// A new primitive that declares a footprint fails here until it has a
// conformance case, and a case for a type that no longer declares one
// fails too.
func TestEveryFootprinterHasAConformanceCase(t *testing.T) {
	types := footprinterTypes(t)
	if len(types) == 0 {
		t.Fatal("found no footprint() methods: the scan is broken")
	}
	for _, name := range types {
		if _, ok := footprintConformanceCases[name]; !ok && !hasVariantCase(name) {
			t.Errorf("%s declares a footprint and has no case in footprintConformanceCases", name)
		}
	}
	for key, c := range footprintConformanceCases {
		name, _, _ := strings.Cut(key, ":")
		if !slices.Contains(types, name) {
			t.Errorf("footprintConformanceCases has %s, which declares no footprint", key)
		}
		if got := fmt.Sprintf("%T", c.step); got != "effects."+name {
			t.Errorf("the case keyed %s holds a %s", key, got)
		}
	}
}

// hasVariantCase reports a case keyed "<name>:<variant>".
func hasVariantCase(name string) bool {
	for key := range footprintConformanceCases {
		if base, _, ok := strings.Cut(key, ":"); ok && base == name {
			return true
		}
	}
	return false
}

// confPlayer is what a footprint can say about a player.
type confPlayer struct {
	life, energy             int
	counters                 string
	library, hand, graveyard []uuid.UUID
}

// confState is the part of the game a footprint describes: every
// player, every object's zone and rules-relevant state, the monarch.
type confState struct {
	players map[uuid.UUID]confPlayer
	// objects maps an object to "<zone>|<owner of the zone>|<state>".
	objects map[uuid.UUID]string
	monarch uuid.UUID
}

func countersKey(m map[string]int) string {
	keys := make([]string, 0, len(m))
	for k, v := range m {
		if v != 0 {
			keys = append(keys, fmt.Sprintf("%s=%d", k, v))
		}
	}
	sort.Strings(keys)
	return strings.Join(keys, ",")
}

func zoneIDs(z *game.Zone) []uuid.UUID {
	if z == nil {
		return nil
	}
	out := make([]uuid.UUID, 0, len(z.Cards))
	for i := range z.Cards {
		out = append(out, z.Cards[i].InstanceID)
	}
	return out
}

func captureConformance(g *game.Game) confState {
	s := confState{players: map[uuid.UUID]confPlayer{}, objects: map[uuid.UUID]string{}, monarch: g.Monarch}
	add := func(z *game.Zone, zone string, holder uuid.UUID) {
		if z == nil {
			return
		}
		for i := range z.Cards {
			c := &z.Cards[i]
			s.objects[c.InstanceID] = fmt.Sprintf("%s|%s|%s|%s|%v|%d|%v|%v",
				zone, holder, c.Controller, countersKey(c.Counters), c.Tapped, c.DamageMarked, c.AttachedTo, c.FaceDown)
		}
	}
	add(g.Battlefield, "battlefield", uuid.Nil)
	add(g.Exile, "exile", uuid.Nil)
	for _, p := range g.Seats {
		if p == nil {
			continue
		}
		s.players[p.ID] = confPlayer{
			life: p.Life, energy: p.Energy, counters: countersKey(p.Counters),
			library: zoneIDs(p.Library), hand: zoneIDs(p.Hand), graveyard: zoneIDs(p.Graveyard),
		}
		add(p.Library, "library", p.ID)
		add(p.Hand, "hand", p.ID)
		add(p.Graveyard, "graveyard", p.ID)
		add(p.Command, "command", p.ID)
	}
	return s
}

// cardZoneHolder is the player whose library, hand or graveyard an
// object fingerprint is in, or uuid.Nil.
func cardZoneHolder(fp string) uuid.UUID {
	parts := strings.SplitN(fp, "|", 3)
	if len(parts) < 2 {
		return uuid.Nil
	}
	switch parts[0] {
	case "library", "hand", "graveyard":
		id, _ := uuid.Parse(parts[1])
		return id
	}
	return uuid.Nil
}

func confZoneOf(fp string) string {
	zone, _, _ := strings.Cut(fp, "|")
	return zone
}

// answerConformancePrompts answers every scry or surveil prompt the
// step opened, sending every card away, so the move it makes is seen.
func answerConformancePrompts(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 8 && len(g.PendingChoices) > 0; i++ {
		c := g.PendingChoices[0]
		var err error
		switch c.Kind {
		case game.PendingChoiceScry:
			err = g.ResolveScry(c.ID, c.Chooser, c.ScryCards, nil)
		case game.PendingChoiceSurveil:
			err = g.ResolveSurveil(c.ID, c.Chooser, c.ScryCards, nil)
		default:
			t.Fatalf("the step opened a %q prompt this test cannot answer", c.Kind)
		}
		if err != nil {
			t.Fatalf("answering %q: %v", c.Kind, err)
		}
	}
}

func TestFootprintStepsDoOnlyWhatTheyDeclare(t *testing.T) {
	names := make([]string, 0, len(footprintConformanceCases))
	for name := range footprintConformanceCases {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		c := footprintConformanceCases[name]
		t.Run(name, func(t *testing.T) {
			g := newCatalogGame(t)
			me, opp := g.Seats[0], g.Seats[1]
			src := pushBattlefieldCreature(g, me.ID)
			var source game.Card
			for _, card := range g.Battlefield.Cards {
				if card.InstanceID == src {
					source = card
				}
			}
			item := game.NewTriggeredItem(&source, "Footprint conformance probe")
			item.ID = uuid.New()
			if c.target {
				victim := pushBattlefieldCreature(g, opp.ID)
				item.Targets = []game.TargetRef{{Kind: game.TargetCard, ID: victim}}
			}
			item.XValue = c.x
			if c.targetPlayer {
				item.Targets = []game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}}
			}
			if c.departed {
				g.WithWriteLock(func() {
					if err := g.DestroyPermanentForEffect(src); err != nil {
						t.Fatalf("destroying the source: %v", err)
					}
					if c.movedOn {
						if err := g.ExileCardForEffect(src); err != nil {
							t.Fatalf("exiling the source: %v", err)
						}
					}
				})
			}
			if c.subject {
				subject := b21Push(g, me.ID, "Conformance Subject", "Creature — Elf", "", 2, 3)
				var epoch int
				for _, card := range g.Battlefield.Cards {
					if card.InstanceID == subject {
						epoch = card.ObjectEpoch
					}
				}
				item.Trigger = &game.TriggerContext{
					Event:  game.Event{Kind: game.EventETB, CardID: subject, Actor: me.ID},
					Object: &game.ObjectSnapshot{ID: subject, Epoch: epoch},
				}
			}
			steps, ok := stepFootprints(c.step)
			if !ok {
				t.Fatalf("the case's %s declares no footprint: give it fields that do", name)
			}

			var decl game.FootprintDeclaration
			var before confState
			var events int
			var applyErr error
			g.WithWriteLock(func() {
				if c.setup != nil {
					c.setup(t, g, me, src)
				}
				decl, ok = footprintDeclarationOfSteps(g, item, steps)
				before = captureConformance(g)
				events = len(g.Events)
				applyErr = c.step.Apply(NewContext(g, item))
			})
			if !ok {
				t.Fatalf("the check refuses %s on this board", name)
			}
			if applyErr != nil {
				t.Fatalf("Apply: %v", applyErr)
			}
			answerConformancePrompts(t, g)
			after := captureConformance(g)

			// Event kinds.
			var emitted []string
			seen := map[game.EventKind]bool{}
			for _, ev := range g.Events[events:] {
				if seen[ev.Kind] {
					continue
				}
				seen[ev.Kind] = true
				emitted = append(emitted, string(ev.Kind))
				if !decl.Emits[ev.Kind] {
					t.Errorf("emitted %q, which the footprint does not declare (declared %v)", ev.Kind, decl.Emits)
				}
			}
			if len(emitted) == 0 {
				t.Errorf("emitted no event: the probe did nothing, so it proves nothing")
			}

			// Players.
			for id, b := range before.players {
				a := after.players[id]
				if a.life > b.life && !decl.LifeUp[id] {
					t.Errorf("player %s gained life (%d → %d) undeclared", id, b.life, a.life)
				}
				if a.life < b.life && !decl.LifeDown[id] {
					t.Errorf("player %s lost life (%d → %d) undeclared", id, b.life, a.life)
				}
				if (!slices.Equal(a.library, b.library) || !slices.Equal(a.hand, b.hand) || !slices.Equal(a.graveyard, b.graveyard)) &&
					!decl.Cards[id] && !touchesOnlyWrites(b, a, decl) {
					t.Errorf("player %s's library, hand or graveyard changed undeclared", id)
				}
				if (a.energy != b.energy || a.counters != b.counters) && !decl.Emits[game.EventPlayerCounterPlaced] {
					t.Errorf("player %s's counters changed (%d/%q → %d/%q) undeclared", id, b.energy, b.counters, a.energy, a.counters)
				}
			}
			if before.monarch != after.monarch && !decl.Monarch {
				t.Error("the monarch changed undeclared")
			}

			// Objects.
			for id, b := range before.objects {
				a, still := after.objects[id]
				if still && a == b {
					continue
				}
				if decl.Writes[id] {
					continue
				}
				if h := cardZoneHolder(b); h != uuid.Nil && decl.Cards[h] && (!still || cardZoneHolder(a) == h) {
					continue
				}
				t.Errorf("object %s changed undeclared: %q → %q", id, b, a)
			}
			for id, a := range after.objects {
				if _, was := before.objects[id]; was {
					continue
				}
				switch {
				case confZoneOf(a) == "battlefield" && decl.Emits[game.EventTokenCreated]:
				case confZoneOf(a) == "exile" && len(decl.Writes) > 0:
				default:
					t.Errorf("new object %s undeclared: %q", id, a)
				}
			}
		})
	}
}

// footprintDeclarationOfSteps is the union of the check's reading of
// each step, or false when it refuses one. Caller holds the write lock.
func footprintDeclarationOfSteps(g *game.Game, item *game.StackItem, steps []game.FootprintStep) (game.FootprintDeclaration, bool) {
	out := game.FootprintDeclaration{
		Reads: map[uuid.UUID]bool{}, Writes: map[uuid.UUID]bool{},
		LifeUp: map[uuid.UUID]bool{}, LifeDown: map[uuid.UUID]bool{},
		Cards: map[uuid.UUID]bool{}, Emits: map[game.EventKind]bool{},
	}
	for _, s := range steps {
		d, ok := g.FootprintDeclarationForTest(item, s)
		if !ok {
			return game.FootprintDeclaration{}, false
		}
		for _, pair := range []struct{ dst, src map[uuid.UUID]bool }{
			{out.Reads, d.Reads}, {out.Writes, d.Writes}, {out.LifeUp, d.LifeUp}, {out.LifeDown, d.LifeDown}, {out.Cards, d.Cards},
		} {
			for k := range pair.src {
				pair.dst[k] = true
			}
		}
		for k := range d.Emits {
			out.Emits[k] = true
		}
		out.Monarch = out.Monarch || d.Monarch
	}
	return out, true
}

// A Condition type with no If case would be an intervening "if" whose
// read nothing resolves on a real board.
func TestEveryConditionHasAConformanceCase(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "footprint_if.go", nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatal(err)
	}
	covered := map[string]bool{}
	var walk func(a Applier)
	walk = func(a Applier) {
		if s, ok := a.(If); ok {
			covered[strings.TrimPrefix(fmt.Sprintf("%T", s.Cond), "effects.")] = true
			for _, th := range s.Then {
				walk(th)
			}
		}
	}
	for _, c := range footprintConformanceCases {
		walk(c.step)
	}
	found := 0
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Recv == nil || fn.Name.Name != "condition" || len(fn.Recv.List) != 1 {
			continue
		}
		id, ok := fn.Recv.List[0].Type.(*ast.Ident)
		if !ok {
			continue
		}
		found++
		if !covered[id.Name] {
			t.Errorf("Condition %s has no If case in footprintConformanceCases", id.Name)
		}
	}
	if found == 0 {
		t.Fatal("found no condition() methods: the scan is broken")
	}
}

// touchesOnlyWrites reports whether a player's card zones changed only
// by losing or gaining objects the declaration writes (an exiled
// target that was in no card zone of theirs changes none of them, so
// this is for a written object that moved through one).
func touchesOnlyWrites(b, a confPlayer, decl game.FootprintDeclaration) bool {
	diff := func(x, y []uuid.UUID) bool {
		for _, id := range x {
			if !slices.Contains(y, id) && !decl.Writes[id] {
				return false
			}
		}
		return true
	}
	return diff(b.library, a.library) && diff(a.library, b.library) &&
		diff(b.hand, a.hand) && diff(a.hand, b.hand) &&
		diff(b.graveyard, a.graveyard) && diff(a.graveyard, b.graveyard)
}

package testsbr

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// cast_creature_type_test.go — #6245: Selective Snare's "Return X target
// creatures of the creature type of your choice to their owner's hand",
// on a test oracle with one target.
//
//	601.2c The player announces their choice of an appropriate object or
//	       player for each target the spell requires.
//	608.2b … The spell or ability … checks whether the targets are still
//	       legal …

const testSnareOracle = "test-6245-snare"

func registerSnare(t *testing.T) {
	t.Helper()
	effectstest.RegisterForTest(t, effects.Spec{
		OracleID:                  testSnareOracle,
		Name:                      "Test Snare",
		Completeness:              effects.CompletenessFull,
		ChoosesCreatureTypeAsCast: true,
		Targets:                   effects.OfTheChosenType(effects.TargetCreature("target creature of the creature type of your choice")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (effects.BounceToHand{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

type snareTable struct {
	g           *game.Game
	me          *game.Player
	elf, goblin uuid.UUID
}

func newSnareTable(t *testing.T) snareTable {
	t.Helper()
	registerSnare(t)
	g := effectstest.NewCatalogGame(t)
	effectstest.ToMainForCost(t, g)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%4]
	elf := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		Power: 1, Toughness: 1, Owner: opp.ID, Controller: opp.ID})
	goblin := effectstest.PushBattlefieldCardWithTimestamp(g, game.Card{InstanceID: uuid.New(), Name: "Goblin Guide", TypeLine: "Creature — Goblin Scout",
		Power: 2, Toughness: 2, Owner: opp.ID, Controller: opp.ID})
	return snareTable{g: g, me: me, elf: elf, goblin: goblin}
}

func (tb snareTable) cast(t *testing.T, chosen string, target uuid.UUID) (uuid.UUID, error) {
	t.Helper()
	c := game.Card{InstanceID: uuid.New(), Name: "Test Snare", TypeLine: "Instant", OracleID: testSnareOracle, Owner: tb.me.ID, Controller: tb.me.ID}
	tb.me.Hand.PushTop(c)
	return c.InstanceID, tb.g.CastSpell(tb.me.ID, c.InstanceID, game.CastSpellParams{
		ChosenCreatureType: chosen,
		Targets:            []game.TargetRef{{Kind: game.TargetCard, ID: target}},
		Strict:             true,
	})
}

// Naming Elf: only an Elf can be chosen; the Elf goes home.
func TestCastCreatureTypeLimitsTheTargets(t *testing.T) {
	tb := newSnareTable(t)
	if _, err := tb.cast(t, "Elf", tb.goblin); err == nil {
		t.Fatal("a Goblin was accepted as a target of the creature type Elf")
	}
	if _, err := tb.cast(t, "", tb.elf); err == nil {
		t.Fatal("the spell was cast without naming a creature type")
	}
	if _, err := tb.cast(t, "Notatype", tb.elf); err == nil {
		t.Fatal("a word that is no creature type was accepted (CR 205.3m)")
	}
	snare, err := tb.cast(t, "elf", tb.elf)
	if err != nil {
		t.Fatalf("cast naming Elf at the Elf: %v", err)
	}
	if got := stackCardTribe(tb.g, snare); got != "Elf" {
		t.Fatalf("the spell on the stack has chosen type %q, want Elf", got)
	}
	effectstest.PassPriorityAroundTable(t, tb.g)
	if effects.OnBattlefield(tb.g, tb.elf) {
		t.Fatal("the Elf is still on the battlefield")
	}
	if !effects.OnBattlefield(tb.g, tb.goblin) {
		t.Fatal("the Goblin left too")
	}
	if c, ok := tb.g.LookupCardForEffect(snare); ok && c.NamedTribe != "" {
		t.Fatalf("the card kept the chosen type %q after leaving the stack (CR 400.7)", c.NamedTribe)
	}
}

// CR 608.2b: a target that stops being of the chosen type before the
// spell resolves is illegal, and the spell does nothing to it.
func TestCastCreatureTypeIsRecheckedOnResolution(t *testing.T) {
	tb := newSnareTable(t)
	if _, err := tb.cast(t, "Elf", tb.elf); err != nil {
		t.Fatalf("cast: %v", err)
	}
	tb.g.WithWriteLock(func() {
		for i := range tb.g.Battlefield.Cards {
			if tb.g.Battlefield.Cards[i].InstanceID == tb.elf {
				tb.g.Battlefield.Cards[i].TypeLine = "Creature — Human Druid"
			}
		}
	})
	tb.g.BumpLayerVersionForTest()
	effectstest.PassPriorityAroundTable(t, tb.g)
	if !effects.OnBattlefield(tb.g, tb.elf) {
		t.Fatal("a creature no longer of the chosen type was returned")
	}
}

// Another spell may not name a type.
func TestCastCreatureTypeIsRefusedOnOtherSpells(t *testing.T) {
	tb := newSnareTable(t)
	c := game.Card{InstanceID: uuid.New(), Name: "Plain Spell", TypeLine: "Sorcery", Owner: tb.me.ID, Controller: tb.me.ID}
	tb.me.Hand.PushTop(c)
	if err := tb.g.CastSpell(tb.me.ID, c.InstanceID, game.CastSpellParams{ChosenCreatureType: "Elf", Strict: true}); err == nil {
		t.Fatal("a spell that names no creature type accepted one")
	}
}

// The view offers each type with its cards, and the enumerator casts it
// naming a type, at a target of that type.
func TestCastCreatureTypeViewAndMoves(t *testing.T) {
	tb := newSnareTable(t)
	c := game.Card{InstanceID: uuid.New(), Name: "Test Snare", TypeLine: "Instant", OracleID: testSnareOracle, Owner: tb.me.ID, Controller: tb.me.ID,
		KnownBy: map[uuid.UUID]bool{tb.me.ID: true}}
	tb.g.WithWriteLock(func() { tb.me.Hand.PushTop(c) })

	var offered map[string][]string
	for _, seat := range protocol.ViewOfGameFor(tb.g, tb.me.ID.String()).Seats {
		for _, hc := range seat.Hand.Cards {
			if hc.InstanceID == c.InstanceID.String() {
				offered = map[string][]string{}
				for _, o := range hc.CastCreatureTypes {
					offered[o.Type] = o.Cards
				}
			}
		}
	}
	if len(offered["Elf"]) != 1 || offered["Elf"][0] != tb.elf.String() || len(offered["Goblin"]) != 1 || offered["Goblin"][0] != tb.goblin.String() {
		t.Fatalf("cast_creature_types = %v, want Elf: the Elf and Goblin: the Goblin", offered)
	}

	moves := 0
	for _, m := range legal.EnumerateFor(tb.g, tb.me.ID) {
		if m.Type != legal.TypeCastSpell {
			continue
		}
		var p struct {
			InstanceID string `json:"instance_id"`
			Chosen     string `json:"chosen_creature_type"`
			Targets    []struct {
				ID string `json:"id"`
			} `json:"targets"`
		}
		if json.Unmarshal(m.Params, &p) != nil || p.InstanceID != c.InstanceID.String() {
			continue
		}
		moves++
		want := map[string]string{"Elf": tb.elf.String(), "Druid": tb.elf.String(), "Goblin": tb.goblin.String(), "Scout": tb.goblin.String()}[p.Chosen]
		if want == "" || len(p.Targets) != 1 || p.Targets[0].ID != want {
			t.Fatalf("move %s names type %q at %v", m.Label, p.Chosen, p.Targets)
		}
	}
	if moves == 0 {
		t.Fatal("the enumerator offered no cast")
	}
}

func stackCardTribe(g *game.Game, id uuid.UUID) string {
	for _, c := range g.Stack.Cards {
		if c.InstanceID == id {
			return c.NamedTribe
		}
	}
	return ""
}

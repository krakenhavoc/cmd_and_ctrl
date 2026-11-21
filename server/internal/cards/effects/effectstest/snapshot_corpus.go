package effectstest

import (
	"encoding/json"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// newCorpusGame is a started, mulligans-closed two-seat game: a
// commander and eleven Forests each, seeded so the deal is fixed.
func NewCorpusGame(t *testing.T) *game.Game {
	t.Helper()
	g := game.NewGame()
	for i, name := range []string{"P1", "P2"} {
		deck := []game.Card{{
			InstanceID:  uuid.New(),
			Name:        fmt.Sprintf("Test Commander %d", i+1),
			TypeLine:    "Legendary Creature — Test",
			Power:       3,
			Toughness:   3,
			IsCommander: true,
		}}
		for range 11 {
			deck = append(deck, game.Card{InstanceID: uuid.New(), Name: "Forest", TypeLine: "Basic Land — Forest"})
		}
		if _, err := g.AddPlayer(name, deck); err != nil {
			t.Fatalf("AddPlayer: %v", err)
		}
	}
	if err := g.StartWithSource(rand.NewPCG(522, 6)); err != nil {
		t.Fatalf("StartWithSource: %v", err)
	}
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	return g
}

const CorpusCounterspellOracle = "cc187110-1148-4090-bbb8-e205694a39f5"

// corpusSettleTrigger passes priority until a triggered item from
// `source` is on the stack, and returns it.
func CorpusSettleTrigger(t *testing.T, g *game.Game, source uuid.UUID) *game.StackItem {
	t.Helper()
	for i := 0; i < 8; i++ {
		if it := TriggerOnStack(g, source); it != nil {
			return it
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatalf("no trigger from %s reached the stack", source)
	return nil
}

// corpusRequireTriggeredStamp fails unless the item names a triggered
// catalog row whose ref starts with `refPrefix`.
func CorpusRequireTriggeredStamp(t *testing.T, it *game.StackItem, refPrefix string) {
	t.Helper()
	if it == nil || it.Body != game.CatalogTriggeredBodyKey || it.Params.Ability == nil ||
		it.Params.Ability.Slot != game.AbilitySlotTriggered || !strings.HasPrefix(it.Params.Ability.Ref, refPrefix) {
		t.Fatalf("setup: the trigger is not stamped with a %s… triggered ref: %+v", refPrefix, it)
	}
}

// corpusTargetsFromTrigger is a real Gixian Puppeteer's dies trigger —
// "return ANOTHER target creature card …", whose clause TargetsFrom
// builds from the trigger's own source — with its target chosen,
// waiting on the stack. Restore calls TargetsFrom again.
func CorpusTargetsFromTrigger(t *testing.T) *game.Game {
	g := NewCorpusGame(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	me := seat.ID
	puppeteer := B12Push(g, me, "Gixian Puppeteer", "Creature — Phyrexian Warlock", B40GixianPuppeteerOracle, 2, 3)
	elf := uuid.New()
	seat.Graveyard.PushTop(game.Card{
		InstanceID: elf, Name: "Llanowar Elves", TypeLine: "Creature — Elf Druid",
		ManaCost: "{G}", Power: 1, Toughness: 1, Owner: me, Controller: me,
	})
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(puppeteer) })
	PickTriggerTarget(t, g, me, elf)
	CorpusRequireTriggeredStamp(t, TriggerOnStack(g, puppeteer), "own:")
	return g
}

// corpusStormAfterCounter is a real Grapeshot's storm trigger waiting
// on the stack after Counterspell countered the Grapeshot: the trigger
// is data, and the spell it copies is in lastKnownStack.
func CorpusStormAfterCounter(t *testing.T) *game.Game {
	g := NewCorpusGame(t)
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)].ID
	CastCatalogSpell(t, g, "Lightning Bolt", "Instant", LightningBoltOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	PassPriorityAroundTable(t, g)
	shot := CastCatalogSpell(t, g, "Grapeshot", "Sorcery", GrapeshotOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp}})
	CorpusRequireTriggeredStamp(t, CorpusSettleTrigger(t, g, shot), "own:")
	counter := CastCatalogSpell(t, g, "Counterspell", "Instant", CorpusCounterspellOracle,
		[]game.TargetRef{{Kind: game.TargetCard, ID: shot}})
	for i := 0; i < 8 && g.Stack.Contains(counter); i++ {
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	snap := g.CaptureSnapshot()
	if g.Stack.Contains(shot) || TriggerOnStack(g, shot) == nil || len(snap.LastKnownStack) != 1 {
		t.Fatalf("setup: grapeshot on the stack %v, storm trigger %v, %d last-known spells — want the storm trigger over a countered Grapeshot",
			g.Stack.Contains(shot), TriggerOnStack(g, shot) != nil, len(snap.LastKnownStack))
	}
	return g
}

func DecodeRestorePoint(t *testing.T, raw []byte) *game.GameSnapshot {
	t.Helper()
	var s game.GameSnapshot
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	return &s
}

func MustMarshal(t *testing.T, v any) []byte {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

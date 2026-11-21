package effectstest

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// prowess_test.go — #706, CR 702.108. Prowess is a canonical keyword
// token the engine turns into one trigger per instance
// (game/prowess.go); these tests hold the rule's five edges — what
// triggers it, what doesn't, how instances add up, when it ends, and
// where the token can come from — plus one per proof card.

const (
	OracleMonasteryMentor   = "3979067a-9c68-443d-a85f-d9f07be880b9"
	OracleSokka             = "6b68acc2-b9d5-495b-8054-c04bae1349f1"
	OracleElementalEruption = "8629fdef-dbbe-4a97-a7e1-6b1646f4ca0b"
	OracleCoriSteelCutter   = "46b1a169-bc32-47bf-b379-fa0c3d13878f"
)

// pushProwessCreature is a creature with no catalog entry whose
// printed keywords the deck importer stamped — the Monastery
// Swiftspear case, and the proof that prowess needs no card file.
func PushProwessCreature(g *game.Game, controller uuid.UUID, name string, power, toughness int, keywords ...string) uuid.UUID {
	if len(keywords) == 0 {
		keywords = []string{game.KeywordProwess}
	}
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(),
		Name:       name,
		TypeLine:   "Creature — Human Monk",
		Power:      power,
		Toughness:  toughness,
		Keywords:   keywords,
		Owner:      controller,
		Controller: controller,
	})
}

// settleProwess passes priority until nothing is on or headed for the
// stack, answering any CR 603.3b trigger-order prompt in the order
// offered. A prowess creature and another trigger from a different
// source on the same cast is two different triggers, and the player
// orders them.
func SettleProwess(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if ch := TriggerOrderPrompt(g); ch != nil {
			AnswerTriggerOrderInOfferedOrder(t, g)
			continue
		}
		if StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			if errors.Is(err, game.ErrChoicePending) {
				if TriggerOrderPrompt(g) != nil {
					continue
				}
				// An optional trigger asking its "you may" as it
				// resolves (CR 603.5) is the caller's to answer.
				if LatestChoiceOfKind(g, game.PendingChoiceTriggerPrompt) != nil {
					return
				}
				t.Fatalf("a prompt other than trigger ordering is open: %+v", g.PendingChoices)
			}
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
}

// castNoncreature casts a non-catalog instant from the active seat —
// the plainest noncreature spell there is — and returns its ID with the
// cast's triggers already harvested.
func CastNoncreature(t *testing.T, g *game.Game) uuid.UUID {
	t.Helper()
	return CastCatalogSpell(t, g, "Filler", "Instant", "", nil)
}

func ProwessItemsFrom(g *game.Game, source uuid.UUID) int {
	n := 0
	for _, item := range g.StackMeta {
		if item != nil && item.Kind == game.StackItemTriggered && item.SourceCardID == source && item.Label == "Prowess — +1/+1 until end of turn" {
			n++
		}
	}
	for _, item := range g.PendingTriggers {
		if item != nil && item.SourceCardID == source && item.Label == "Prowess — +1/+1 until end of turn" {
			n++
		}
	}
	return n
}

func ProwessCountOf(t *testing.T, g *game.Game, id uuid.UUID) int {
	t.Helper()
	n := -1
	g.ReadSnapshot(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == id {
				n = game.ProwessCount(&g.Battlefield.Cards[i])
			}
		}
	})
	if n < 0 {
		t.Fatalf("card %s not on the battlefield", id)
	}
	return n
}

func ProwessFirstNamed(t *testing.T, g *game.Game, name string, controller uuid.UUID) uuid.UUID {
	t.Helper()
	var id uuid.UUID
	g.ReadSnapshot(func() {
		for _, c := range g.Battlefield.Cards {
			if c.Name == name && c.Controller == controller {
				id = c.InstanceID
				return
			}
		}
	})
	if id == uuid.Nil {
		t.Fatalf("no %q on the battlefield under %s", name, controller)
	}
	return id
}

func EffectiveAbilitiesContain(t *testing.T, g *game.Game, id uuid.UUID, kw string) bool {
	t.Helper()
	for _, a := range EffectiveAbilities(t, g, id) {
		if a == kw {
			return true
		}
	}
	return false
}

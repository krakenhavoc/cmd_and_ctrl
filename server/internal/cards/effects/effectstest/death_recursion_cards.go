package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// death_recursion_cards_test.go — the All Will Be Black death, discard
// and recursion cards (#2952): Bone Miser, Carnival of Souls,
// Deathgreeter, Endless Cockroaches, Glistening Oil, Necroskitter,
// Nether Traitor, Sifter of Skulls and Vat of Rebirth.

const (
	DrcBoneMiser        = "3984e5f6-8643-4104-aa9c-18d474e07801"
	DrcCarnivalOfSouls  = "95b10ca7-7360-4da5-bd93-686ae3051833"
	DrcDeathgreeter     = "b7ad11dc-c962-4034-b15a-a080c6fff75f"
	DrcCockroaches      = "9f31dbb1-c350-46f8-bd1d-9f23a073d2f1"
	DrcGlisteningOil    = "645a4dee-a4e0-4850-a781-522aa718827f"
	DrcNecroskitter     = "536f7c92-2f1e-49e7-98de-47382a9488f1"
	DrcNetherTraitor    = "6f16c3ac-a9b8-47e6-b18b-ae37c74d44a0"
	DrcSifterOfSkulls   = "688e5e19-8f28-45c4-ace4-a38144f494f1"
	DrcVatOfRebirth     = "10c31317-71e8-42e0-85e0-3e64bd0c3dd3"
	DrcRestInPeace      = "087f9ad7-e74f-40e2-8102-1ed2925d0418"
	DrcDamnationOracle  = "d57a8f0b-7989-4db5-8756-6f2690097252"
	DrcLibraryOfLeng    = "867def48-4be8-4056-bcf1-d6b00450b9a3"
	DrcCreatureTypeLine = "Creature — Bear"
)

// drcSettle passes priority until the stack is empty or a prompt other
// than a trigger-order one is open, ordering simultaneous triggers as
// offered.
func DrcSettle(t *testing.T, g *game.Game, chooser uuid.UUID) {
	t.Helper()
	for i := 0; i < 64; i++ {
		if p := TriggerOrderPromptFor(g, chooser); p != nil {
			if err := g.ResolveTriggerOrder(p.ID, chooser, append([]uuid.UUID(nil), p.TriggerOrderIDs...)); err != nil {
				t.Fatalf("ResolveTriggerOrder: %v", err)
			}
			continue
		}
		if len(g.PendingChoices) > 0 || StackFullyEmpty(g) {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("did not settle")
}

func DrcBear(g *game.Game, owner uuid.UUID) uuid.UUID {
	return B12Creature(g, owner, "Bear", DrcCreatureTypeLine, 2, 2)
}

func DrcManaOfColor(p *game.Player, color string) int {
	n := 0
	for _, tok := range p.ManaPool.Units() {
		if tok.Color == color {
			n++
		}
	}
	return n
}

// --- Bone Miser ------------------------------------------------------

func DrcDiscard(t *testing.T, g *game.Game, who *game.Player, name, typeLine string) uuid.UUID {
	t.Helper()
	card := HandCardFull(who, name, typeLine, "{1}", "", nil)
	LengDiscard(t, g, who, card)
	return card
}

// --- Necroskitter -----------------------------------------------------

func DrcWithMinusOne(g *game.Game, owner, controller uuid.UUID, counters int) uuid.UUID {
	c := game.Card{
		InstanceID: uuid.New(), Name: "Victim", TypeLine: DrcCreatureTypeLine,
		Power: 2, Toughness: 4, Owner: owner, Controller: controller,
	}
	if counters > 0 {
		c.Counters = map[string]int{game.CounterMinusOne: counters}
	}
	return PushBattlefieldCardWithTimestamp(g, c)
}

// --- Nether Traitor ---------------------------------------------------

func DrcTraitorInGraveyard(g *game.Game, owner *game.Player) uuid.UUID {
	id := uuid.New()
	g.WithWriteLock(func() {
		owner.Graveyard.PushTop(game.Card{
			InstanceID: id, Name: "Nether Traitor", TypeLine: "Creature — Spirit", OracleID: DrcNetherTraitor,
			Power: 1, Toughness: 1, Owner: owner.ID, Controller: owner.ID, ManaCost: "{B}{B}",
		})
	})
	return id
}

// --- Vat of Rebirth ---------------------------------------------------

func DrcOilOn(t *testing.T, g *game.Game, vat uuid.UUID) int {
	t.Helper()
	c := BattlefieldCardCopy(t, g, vat)
	return c.Counters["oil"]
}

// pushGraveyardCard puts a card of the given type line into p's graveyard.
func DrcGraveyardCard(p *game.Player, name, typeLine string) uuid.UUID {
	id := uuid.New()
	p.Graveyard.PushTop(game.Card{
		InstanceID: id, Name: name, TypeLine: typeLine, Power: 2, Toughness: 2,
		Owner: p.ID, Controller: p.ID,
	})
	return id
}

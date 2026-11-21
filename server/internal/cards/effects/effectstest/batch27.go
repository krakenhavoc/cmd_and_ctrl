package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// batch27_test.go — card-level coverage for the card-coverage
// roadmap's batch 27 (#389, `edhrec_rank` 2839–2939): the "no new
// machinery" group. One test per observable behaviour, driven
// through a real cast, activation, attack, damage, land play or step
// change. Helpers from the earlier batch test files are reused by
// name; new ones are b27-prefixed.

const (
	B27RenataOracle                = "6761b077-7a89-42c9-93ab-675fe4231564"
	B27UrzasBaubleOracle           = "17dbbca3-ac1c-4d4e-9618-3e66ac3ccd24"
	B27GingerbreadCabinOracle      = "fa98c367-0312-49c6-abef-72e5ead4cc7d"
	B27EarthquakeOracle            = "9a40614b-50a3-422c-849e-53c8b7d3d204"
	B27RevengeOfRavensOracle       = "dd96f145-73eb-4a6d-bcfd-5d6313fac1f9"
	B27ArchetypeOfAggressionOracle = "263408e6-b315-4af5-8cb8-3fd1aa88e48c"
	B27HellToPayOracle             = "7dc7d90d-5979-4c4a-9183-392e0e0b878a"
	B27FinaleOfGloryOracle         = "68273453-1059-4eab-8c27-5d96f998f3b1"
	B27HeadlessRiderOracle         = "d4fdacd7-3101-44e2-a880-dde7326137a4"
	B27OneWithTheMachineOracle     = "757eadac-dd29-4a7f-8683-5bc823168326"
	B27ChiefOfTheFoundryOracle     = "4fdfa41a-75a5-4b36-8c9e-083e26d137e2"
	B27OswaldFiddlebenderOracle    = "dbc9ea19-cf68-41d3-88a8-b5ab8df75c5a"
	B27MagewrightsStoneOracle      = "9e500094-c430-4c10-89cc-97f4f18b08bc"
	B27WolverineRidersOracle       = "09070db6-01f6-4a8a-b167-a7825e6959f4"
	B27LegionLoyaltyOracle         = "24e39d26-6a2b-461e-b194-bda575bcf898"
	B27DuplicantOracle             = "ea86abfa-6cab-4ef0-8463-34136fc25b59"
	B27FarmerCottonOracle          = "11f1d2b7-0c2b-40df-bf90-3c55b23449af"
	B27MemorialToFollyOracle       = "2bc38f14-0314-4351-8138-e2b8bf041404"
	B27OrcishLumberjackOracle      = "383e3fe4-8558-4561-8632-6eadb5d5963c"
	B27RiteOfTheDragoncallerOracle = "5097f4e6-50af-4641-909f-db44abf0ce32"
	B27FlumphOracle                = "bc328acc-8521-4581-88c3-99a7cb2c1cdb"
	B27TashasHideousLaughterOracle = "e352f5b9-6406-4914-bc79-f24608be6bc9"
	B27FoundationBreakerOracle     = "8b7a3613-f1bd-4262-9b11-b631167c2c2d"
	B27TheGooseMotherOracle        = "de595f1b-3f7d-45e0-a31b-ed23e5d1ee48"
	B27CuriousAltisaurOracle       = "80686908-abb5-4728-a5f6-71baca27f467"
	B27SanctumOfUginOracle         = "72cb5dcd-9b24-435c-921a-3766108374c4"
	B27DragonMageOracle            = "eab71e4e-27c3-4c41-b95f-259c1d14b97a"
)

// b27Lives reads every seat's life total.
func B27Lives(g *game.Game) []int {
	out := make([]int, 0, len(g.Seats))
	for _, p := range g.Seats {
		out = append(out, p.Life)
	}
	return out
}

// b27Push seeds a permanent with a mana cost, colours and P/T on the
// battlefield with a layer timestamp.
func B27Push(g *game.Game, owner uuid.UUID, name, typeLine, oracle, manaCost string, power, toughness int, colors ...string) uuid.UUID {
	return PushBattlefieldCardWithTimestamp(g, game.Card{
		InstanceID: uuid.New(), Name: name, TypeLine: typeLine, OracleID: oracle, ManaCost: manaCost,
		Colors: colors, Power: power, Toughness: toughness, Owner: owner, Controller: owner,
	})
}

// b27Kill destroys a permanent through the effect API and leaves
// whatever it triggered pending.
func B27Kill(g *game.Game, id uuid.UUID) {
	g.WithWriteLock(func() { _ = g.DestroyPermanentForEffect(id) })
}

// b27Damage deals noncombat damage to a creature through the effect
// API and leaves whatever it triggered pending.
func B27Damage(g *game.Game, source, target uuid.UUID, n int) {
	g.WithWriteLock(func() { _ = g.DealDamageToCreatureForEffect(source, target, n) })
}

// b27CountNamed counts battlefield permanents `controller` controls
// with the given name.
func B27CountNamed(g *game.Game, controller uuid.UUID, name string) int {
	n := 0
	for _, c := range g.Battlefield.Cards {
		if c.Controller == controller && c.Name == name {
			n++
		}
	}
	return n
}

// b27Tapped reads a battlefield card's tapped state.
func B27Tapped(t *testing.T, g *game.Game, id uuid.UUID) bool {
	t.Helper()
	c, ok := BattlefieldCard(g, id)
	if !ok {
		t.Fatalf("%s is not on the battlefield", id)
	}
	return c.Tapped
}

// b27Settle answers every prompt addressed to `chooser` until the
// stack is empty: trigger-order prompts in the offered order,
// yes/no trigger prompts with yes, and pick_target prompts with
// `pick` (a card) — then passes priority. For the cards whose entry
// queues an evoke sacrifice beside an optional targeted trigger.
func B27Settle(t *testing.T, g *game.Game, chooser, pick uuid.UUID) {
	t.Helper()
	for i := 0; i < 40; i++ {
		answered := false
		for _, c := range g.PendingChoices {
			if c == nil || c.Chooser != chooser {
				continue
			}
			switch c.Kind {
			case game.PendingChoiceTriggerOrder:
				AnswerTriggerOrderInOfferedOrder(t, g)
			case game.PendingChoiceTriggerPrompt:
				if err := g.ResolveTriggerPrompt(c.ID, chooser, true); err != nil {
					t.Fatalf("ResolveTriggerPrompt: %v", err)
				}
			case game.PendingChoicePickTarget:
				if err := g.ResolvePickTarget(c.ID, chooser, game.TargetRef{Kind: game.TargetCard, ID: pick}); err != nil {
					t.Fatalf("ResolvePickTarget: %v", err)
				}
			default:
				continue
			}
			answered = true
			break
		}
		if answered {
			continue
		}
		if StackFullyEmpty(g) && len(g.PendingChoices) == 0 {
			return
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("the stack never settled")
}

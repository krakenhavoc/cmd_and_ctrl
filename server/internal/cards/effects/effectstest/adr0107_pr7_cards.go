package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0107_pr7_cards_test.go — ADR 0107 PR 7 (#1860): the cards the
// next-damage shield unblocks, each against the damage it must stop and
// the damage it must not.

const (
	Pr7CoPRed            = "df2738fe-9cd1-4347-8808-105fcfde1190"
	Pr7CoPShadow         = "07c42f46-800e-4d2a-b844-c125ab93190d"
	Pr7RoPLands          = "605dd1b5-8946-4bd0-b31b-e99bf63a1bc4"
	Pr7GreaterRealm      = "b03eb0c4-89a4-420d-8a23-1a868a07d9cf"
	Pr7StoryCircle       = "7071aee8-b5ca-4be5-9ba0-2df7e3af303b"
	Pr7CircleOfSolace    = "14ebedd1-7a2d-413b-9b76-8610511292a4"
	Pr7KithkinArmor      = "35e2730e-b3aa-4080-aae8-7b3ad1e59848"
	Pr7SamiteBlessing    = "a524ae69-2594-467d-af0a-27b9984d4300"
	Pr7Mercenaries       = "9d96671c-7a98-48ec-b479-bc451e18264d"
	Pr7PilgrimOfJustice  = "959b672e-d4e4-49a1-874a-5148f040fadd"
	Pr7MartyrsCause      = "72d1c789-e79f-4f15-80d0-981d4120b085"
	Pr7BoneMask          = "ddfc7dac-f921-4529-935c-6c588890c522"
	Pr7DeflectingPalm    = "dc5dffc8-fac5-4956-bac4-1ad2cc16f6be"
	Pr7HonorablePassage  = "66805334-1015-4b41-ba9b-116de94b0744"
	Pr7Shadowbane        = "9e9ec33b-b923-40e5-81a8-e1e78df63ec7"
	Pr7AweStrike         = "994bcde3-4a76-403d-a6e1-88609a13a99c"
	Pr7DazzlingReflect   = "44db1ff5-7082-4da4-8f4f-4a2d2e3d7666"
	Pr7NewWayForward     = "429368e6-3de1-4a44-a062-86cbbb73e243"
	Pr7InterventionPact  = "9f9882f5-2338-4881-b418-b15348a462b4"
	Pr7ReverseDamage     = "eaaf7c30-f463-4115-a40e-7dc717063413"
	pr7Invulnerability   = "2c964157-d414-4d4e-822b-12ce61e882ff"
	Pr7HaazdaShieldMate  = "49b689a3-9197-4a04-a62f-218b245d6e23"
	Pr7PrismaticCircle   = "022dae2e-7fc3-486e-9168-59652d9ab21c"
	Pr7ChoArrimAlchemist = "1e1d4d81-4a45-41b5-a781-ae2227b4f4c0"
)

// pr7Creature puts a creature of `colors` on the battlefield for `owner`.
func Pr7Creature(g *game.Game, owner uuid.UUID, name string, power int, colors ...string) uuid.UUID {
	return ApaPush(g, owner, owner, game.Card{Name: name, TypeLine: "Creature — Test", Power: power, Toughness: 4, Colors: colors})
}

// pr7Activate has `who` activate row `index` of `source` in permissive
// mode and passes priority until a choice is open or the stack is empty.
func Pr7Activate(t *testing.T, g *game.Game, who, source uuid.UUID, index int, params game.ActivateAbilityParams) {
	t.Helper()
	if err := g.ActivateCatalogAbility(who, source, index, params); err != nil {
		t.Fatalf("activate row %d: %v", index, err)
	}
	PassPriorityAroundTable(t, g)
}

// pr7SourcePrompt is the open choose_source prompt.
func Pr7SourcePrompt(t *testing.T, g *game.Game) *game.PendingChoice {
	t.Helper()
	for _, c := range g.PendingChoices {
		if c != nil && c.Kind == game.PendingChoiceChooseSource {
			return c
		}
	}
	t.Fatal("no choose_source prompt open")
	return nil
}

// pr7Choose answers the open choose_source prompt with `source`.
func Pr7Choose(t *testing.T, g *game.Game, chooser, source uuid.UUID) {
	t.Helper()
	c := Pr7SourcePrompt(t, g)
	if err := g.ResolveChooseSource(c.ID, chooser, []uuid.UUID{source}); err != nil {
		t.Fatalf("choose source: %v", err)
	}
	PassPriorityAroundTable(t, g)
}

func Pr7Offers(c *game.PendingChoice, id uuid.UUID) bool {
	for _, x := range c.ChooseCards {
		if x == id {
			return true
		}
	}
	return false
}

// pr7Hit deals `n` damage from `source` to the player `to`. Each call is
// one damage instruction and so one instance of damage from a source
// (CR 615.8, ADR 0108 PR 0); the shields' "prevented this way" follow-ups
// run as it ends (CR 615.5).
//
// It then runs the priority boundary, ADR 0107's catch-all flush point.
func Pr7Hit(t *testing.T, g *game.Game, source, to uuid.UUID, n int) {
	t.Helper()
	g.WithWriteLock(func() {
		if err := g.DealDamageToPlayerForEffect(source, to, n); err != nil {
			t.Fatal(err)
		}
	})
	g.RunStateChecksForTest()
}

func Pr7Counters(g *game.Game, id uuid.UUID) int {
	return FindBattlefieldCardForTest(g, id).Counters[game.CounterPlusOne]
}

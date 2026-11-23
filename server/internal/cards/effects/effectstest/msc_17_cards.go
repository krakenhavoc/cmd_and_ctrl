package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// msc_17_cards_test.go — Marvel Super Heroes Commander, seventeenth batch
// (#4455). Each test plays the real catalog entry and asserts what its
// text says.

const (
	Msc17DaughterOracle   = "63e0f03f-16da-4c09-ad03-8945f660942d"
	Msc17DoormanOracle    = "01ea2130-e212-4532-b68b-2f940d53ef6e"
	Msc17AnnieOracle      = "4c3ad545-b375-44a7-a24e-58baacc4e4b6"
	Msc17SymbiosisOracle  = "19504e19-2886-4d28-a813-b798a9a446a2"
	Msc17DragonManOracle  = "c03ca2e6-fcce-4018-952f-1c451ab92cfd"
	Msc17DoraOracle       = "406f1982-5d0c-4dba-9010-28b1c6a3602c"
	Msc17ClosetOracle     = "072d05c4-e0d7-47af-b2a9-16b90810fd30"
	Msc17FirstFamOracle   = "e119c709-bf9d-451e-827b-9787d455029a"
	Msc17GenesisOracle    = "06f42f49-32f0-413b-8437-b8efa4d647c2"
	Msc17DoomfallOracle   = "ba2c2a65-256a-46db-bf2b-b02b7d114a42"
	Msc17HatutOracle      = "92ac307e-6531-4465-8870-07988458f5d4"
	Msc17BatrocOracle     = "ebbea7a7-c990-43fa-8f67-a5da17fbbb10"
	Msc17CaptainMarvelOra = "6d86d783-7116-4426-8c0f-718c2d904887"
	Msc17PigeonOracle     = "ac8cae63-270c-4f73-b29b-50f8f2395fd9"
)

// msc17Cast puts a card in the player's hand (they must be the active
// seat) and casts it from a main phase, returning its id and the cast
// error. The spell is left on the stack.
func Msc17Cast(t *testing.T, g *game.Game, p *game.Player, c game.Card, params game.CastSpellParams) (uuid.UUID, error) {
	t.Helper()
	c.InstanceID = uuid.New()
	c.Owner, c.Controller = p.ID, p.ID
	AdvanceToMain(t, g)
	p.Hand.PushTop(c)
	return c.InstanceID, g.CastSpell(p.ID, c.InstanceID, params)
}

// msc17Play casts a card and resolves everything it puts on the stack.
func Msc17Play(t *testing.T, g *game.Game, p *game.Player, c game.Card) uuid.UUID {
	t.Helper()
	id, err := Msc17Cast(t, g, p, c, game.CastSpellParams{})
	if err != nil {
		t.Fatalf("cast %s: %v", c.Name, err)
	}
	PassPriorityAroundTable(t, g)
	return id
}

// msc17UntilATriggerIsOnTheStack passes priority until a triggered ability
// sits on the stack, leaving it unresolved.
func Msc17UntilATriggerIsOnTheStack(t *testing.T, g *game.Game) {
	t.Helper()
	for i := 0; i < 16; i++ {
		for _, it := range g.StackMeta {
			if it != nil && it.Kind == game.StackItemTriggered {
				return
			}
		}
		if err := g.PassPriority(); err != nil {
			t.Fatalf("PassPriority: %v", err)
		}
	}
	t.Fatal("no triggered ability reached the stack")
}

func Msc17TargetPlayer(id uuid.UUID) game.TargetRef {
	return game.TargetRef{Kind: game.TargetPlayer, ID: id, Mode: 0}
}

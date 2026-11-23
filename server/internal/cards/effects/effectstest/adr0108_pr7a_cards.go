package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// adr0108_pr7a_cards_test.go — ADR 0108 Delivery PR 7 (#1904): the
// recipient and "dealt to and dealt by" families of the not-one-use
// shield, each against the damage it must stop and the damage it must
// not.

const (
	Pr7aIndestructibleAura = "e10e8d56-bba6-412d-970e-c24969f32b5b"
	Pr7aDjerusResolve      = "c48ce123-bb07-4221-b3b3-0a458e101ca1"
	Pr7aBraceForImpact     = "f250fb94-e519-4bfa-921e-eddeeafcf76c"
	Pr7aRedeem             = "c2032f88-a000-4ed2-a7e5-61d29723d45e"
	Pr7aEndure             = "f4cd32f0-d6aa-4497-b8db-a0bf4c3c31de"
	Pr7aSafePassage        = "dfa459a1-b065-4488-88d3-4da388261b52"
	Pr7aFavoredHoplite     = "145a15f9-89a9-4eb7-9924-e768f17e7e68"
	Pr7aMazeOfIth          = "38a12bd7-4394-44a8-91a0-6a4ff7fa4f71"
	Pr7aEnergyArc          = "959bcf99-3c9a-4f99-98b2-13514d3dac16"
	Pr7aOko                = "f4bf8bd1-71b0-4eb1-976c-54677238b2a6"
	Pr7aKiora              = "58cc2097-f4b4-4695-9076-78666050cdb6"
	Pr7aAvacyn             = "6da0b6d2-3c5e-49ef-9264-076269c04744"
	Pr7aMorningtidesLight  = "a25bcf65-4417-45e9-a4b1-94e0ea78af63"
	Pr7aMutationalAdv      = "2daa5b89-e772-4a2b-ad52-bbbf148c7b2f"
	Pr7aOriss              = "8d6e0dab-400a-4761-8343-92c7eb7e8735"
	Pr7aMoonlightGeist     = "12a8c546-d0d1-4c3d-bfca-c1e59cef349a"
	Pr7aRiotControl        = "a189e20f-1072-4608-9e50-da2a9d7a7ee3"
	Pr7aUltimateHoly       = "1eae48aa-f01d-4398-a68e-3b5af0542ab1"
	Pr7aTakeTheBait        = "5d39cce9-afa1-422a-8e6c-8f50c710a681"
	Pr7aKurbis             = "991fab9c-8554-4e64-a933-6cc2ffa4edb1"
)

func Pr7aCard(id uuid.UUID) game.TargetRef { return game.TargetRef{Kind: game.TargetCard, ID: id} }

// pr7aShields counts the live preventFromSource records.
func Pr7aShields(g *game.Game) int {
	n := 0
	for _, e := range g.ScopedEffects {
		for _, m := range e.Mods {
			if m.Kind == game.ModPreventFromSource {
				n++
			}
		}
	}
	return n
}

// pr7aBlockedCombat has `attacker` (the active seat's) attack `defender`
// and be blocked by `blocker`, then stops in the declare blockers step
// with priority to act.
func Pr7aBlockedCombat(t *testing.T, g *game.Game, defender, attacker, blocker uuid.UUID) {
	t.Helper()
	AdvanceTo(t, g, game.StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, defender); err != nil {
		t.Fatal(err)
	}
	AdvanceTo(t, g, game.StepDeclareBlockers)
	if err := g.DeclareBlocker(blocker, attacker); err != nil {
		t.Fatal(err)
	}
}

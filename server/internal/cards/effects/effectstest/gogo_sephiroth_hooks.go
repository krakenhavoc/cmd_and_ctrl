package effectstest

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// gogo_sephiroth_hooks_test.go — #1574's two hooks, proved on the two
// Edea-deck cards that needed them.
//
//	"This ability can't be copied"   Gogo, Master of Mimicry against
//	                                 Lithoform Engine, Rings of
//	                                 Brighthearth and another Gogo
//	                                 (ADR 0043 Decision 20)
//	"As this transforms into …"      Sephiroth, One-Winged Angel's Super
//	                                 Nova, made whatever turned it over
//	                                 (ADR 0079 Decision 9)

const (
	GogoLabelPrefix      = "{X}{X}, {T}: Copy target"
	LithoformLabelPrefix = "{2}, {T}: Copy target activated"
	SephirothBackName    = "Sephiroth, One-Winged Angel"
)

// --- "As this creature transforms into …" ---------------------------

// hooksSephirothOnTheBattlefield seats Sephiroth front face up and
// declines its entry sacrifice. Returns its instance ID.
func HooksSephirothOnTheBattlefield(t *testing.T, g *game.Game, me *game.Player) uuid.UUID {
	t.Helper()
	AdvanceToMain(t, g)
	seph := ImportToBattlefield(t, g, SephirothRow(), me)
	PassPriorityAroundTable(t, g)
	if p := LatestChoiceOfKindFor(g, game.PendingChoiceOwnPermanents, me.ID); p != nil {
		if err := g.ResolveOwnPermanents(p.ID, me.ID, nil); err != nil {
			t.Fatalf("decline the entry sacrifice: %v", err)
		}
	}
	PassPriorityAroundTable(t, g)
	return seph
}

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cloudsteel Kirin — Artifact Creature — Equipment Kirin {2}{W}, 3/2:
//
//	"Flying
//	 Equipped creature has flying and "You can't lose the game and your
//	 opponents can't win the game."
//	 Reconfigure {5}"
//
// The quoted ability is Platinum Angel's, given to the equipped
// creature, so its "you" is that creature's controller (CR 109.5), who
// need not be the Kirin's. The Kirin declares the two gates itself
// (game.GameEndGate) with You naming the host's controller, and they
// hold only while it is attached to a creature on the battlefield
// (cloudsteelKirinHost). Flying is an ordinary layer-6 grant.
//
// A host that has lost all its abilities does not have the quoted one,
// so the gates stop. Reconfigure is reconfigure.go (#2639).
func init() {
	Register(Spec{
		OracleID:        "a34a71eb-26a4-4e37-8d3c-a0352bea491b",
		Name:            "Cloudsteel Kirin",
		Completeness:    CompletenessCaveats,
		Caveats:         []string{"If the equipped creature lost all its abilities before the Kirin was attached, it still doesn't get the Kirin's \"can't lose\" ability."},
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{GrantToAttached("flying")},
		GameEndGates: []game.GameEndGate{
			{Scope: game.GateYou, CantLose: true, You: cloudsteelKirinHost},
			{Scope: game.GateOpponents, CantWin: true, You: cloudsteelKirinHost},
		},
		Activated: Reconfigure("{5}"),
	})
}

// cloudsteelKirinHost is the controller of the creature the Kirin is
// attached to, or nobody: unattached, attached to something that is
// not a creature on the battlefield, or attached to a creature that
// has lost all its abilities.
func cloudsteelKirinHost(g *game.Game, source game.Card) uuid.UUID {
	if source.AttachedTo.Kind != game.TargetCard {
		return uuid.Nil
	}
	host, ok := g.LookupCardForEffect(source.AttachedTo.ID)
	if !ok || !onBattlefield(g, host.InstanceID) || !host.IsCreature() || host.HasLostAllAbilities() {
		return uuid.Nil
	}
	return host.Controller
}

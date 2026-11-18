package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ironsoul Enforcer — Artifact Creature — Human Samurai {4}{W}, 4/4
// (Neon Dynasty Commander):
//
//	"Whenever this creature or a commander you control attacks alone,
//	 return target artifact card from your graveyard to the battlefield."
//
// The declaration is the controller's, it is the only attacker (CR 506.5),
// and it is this creature or a commander its controller controls. The
// artifact returns under its owner's control (Sheoldred's reanimation body,
// shared).
//
// No simplification.
func init() {
	const name = "Ironsoul Enforcer"
	Register(Spec{
		OracleID:     "a613b00e-d6e8-4470-a7dc-5eab5e48dce0",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventAttack, ironsoulEnforcerAttacksAlone,
					name+" — return target artifact card from your graveyard to the battlefield",
					sheoldredReanimate),
				TargetCardInGraveyard("target artifact card from your graveyard", YouOwn(), Artifact())),
		},
	})
}

func ironsoulEnforcerAttacksAlone(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
	if !attackDeclaredByYou(ev, source.Controller) || !game.AttackedAlone(g) {
		return false
	}
	if ev.CardID == source.InstanceID {
		return true
	}
	a, ok := g.LookupCardForEffect(ev.CardID)
	return ok && a.IsCommander && a.Controller == source.Controller
}

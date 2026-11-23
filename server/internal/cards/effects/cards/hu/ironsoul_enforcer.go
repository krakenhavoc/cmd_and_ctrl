package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "a613b00e-d6e8-4470-a7dc-5eab5e48dce0",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventAttack, ironsoulEnforcerAttacksAlone,
					name+" — return target artifact card from your graveyard to the battlefield",
					effects.SheoldredReanimate),
				effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.YouOwn(), effects.Artifact())),
		},
	})
}

func ironsoulEnforcerAttacksAlone(ev game.Event, source *game.Card, c game.Characteristic, g *game.Game) bool {
	if !effects.AttackDeclaredByYou(ev, source.Controller) || !game.AttackedAlone(g) {
		return false
	}
	if ev.CardID == source.InstanceID {
		return true
	}
	a, ok := g.LookupCardForEffect(ev.CardID)
	return ok && a.IsCommander && a.Controller == source.Controller
}

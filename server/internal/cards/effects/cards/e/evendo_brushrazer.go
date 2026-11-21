package cardse

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Evendo Brushrazer — Creature — Insect Warrior {2}{R}, 2/2 (Edge of
// Eternities Commander):
//
//	"Whenever you sacrifice a nontoken permanent, exile the top card of
//	 your library.
//	 During your turn, as long as you've sacrificed a nontoken permanent
//	 this turn, you may play cards exiled with this creature.
//	 {T}, Sacrifice a land: Add {R}{R}."
//
// The exile is a LINKED exile (ExileTarget.Linked, CR 607.2a), one per
// nontoken permanent sacrificed. The permission is Azula's shape: a gated
// standing permission over exile narrowed to cards exiled with this
// object (ExiledWithSource), in your turn only (TimingYourTurnOnly, the
// cards' own timing and costs in force), whose condition reads the turn's
// nontoken-sacrifice tally (TurnTally.SacrificedCardTypes under
// game.SacrificedNontokenKey, kept at the sacrifice). The mana ability's
// sacrifice is a cost, so it feeds the trigger and the tally itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "795bc0b0-4b0a-49d6-ac8e-a52a10a93be8",
		Name:         "Evendo Brushrazer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventSacrifice, effects.CmdrB9YouSacrificedNontoken,
				"Evendo Brushrazer — exile the top card of your library", evendoExileTop),
		},
		GatedCastPermissions: []game.CastPermissionGate{{
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.CmdrB9SacrificedNontokenThisTurn(g, controller)
			},
			Permission: game.CastPermission{
				Zone:   game.ZoneExile,
				Filter: game.PermissionFilter{ExiledWithSource: true},
				Timing: game.TimingYourTurnOnly,
				Label:  "During your turn, as long as you've sacrificed a nontoken permanent this turn, play cards exiled with this creature (Evendo Brushrazer)",
			},
		}},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, SacrificeOther: effects.B08SacrificeALand().SacrificeOther},
			Produced: "{R}{R}",
			Label:    "{T}, Sacrifice a land: Add {R}{R}.",
			Answers:  game.AnswerValue,
		}},
	})
}

// evendoExileTop exiles the top card of the controller's library, linked
// to this Brushrazer.
func evendoExileTop(g *game.Game, item *game.StackItem) error {
	p := g.PlayerByIDForEffect(item.Controller)
	if p == nil || p.Library == nil {
		return nil
	}
	top, err := p.Library.Top()
	if err != nil {
		return nil
	}
	return effects.ExileTarget{Target: top.InstanceID, Linked: true}.Apply(effects.NewContext(g, item))
}

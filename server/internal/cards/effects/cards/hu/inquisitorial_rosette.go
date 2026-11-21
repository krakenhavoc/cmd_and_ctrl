package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Inquisitorial Rosette — Artifact — Equipment {2} (Warhammer 40,000
// Commander, S82 GAP batch 186, #5948):
//
//	"Inquisition Agents — Whenever equipped creature attacks, create a
//	 2/2 white Astartes Warrior creature token with vigilance that's
//	 attacking. Then attacking creatures gain menace until end of turn.
//	 Equip {3}"
//
// The token is never declared as an attacker (CR 508.4); its controller
// chooses what it attacks, the equipped creature's defender first
// (CreateTokensAttackingYourChoice). "Then attacking creatures" is every
// attacking creature, whoever controls it, the new token included: it is
// the token's tail, which runs once the token has landed, also when there
// was nothing for it to attack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a3a25ae-61f5-4bcd-ae3d-4b31feaa83d8",
		Name:         "Inquisitorial Rosette",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.EquipAbility("{3}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.AttachedCreatureAttacked(ev, source)
			}, "Inquisitorial Rosette — create a 2/2 white Astartes Warrior token with vigilance that's attacking, then attacking creatures gain menace until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.CreateTokensAttackingYourChoice{
						Template: effects.TokenCard("2/2 white Astartes Warrior with vigilance"),
						N:        1,
						Prefer:   effects.TriggerAttackTarget(item),
						Label:    "Inquisitorial Rosette",
						Tail:     game.KeyedTail{Ref: inquisitorialRosetteMenace},
					}.Apply(ctx)
				}),
		},
	})
}

// inquisitorialRosetteMenace is "Then attacking creatures gain menace until
// end of turn", run once the token is in. An on-disk key: never renamed,
// never reused.
var inquisitorialRosetteMenace = effects.AttackingTokensThen("inquisitorial-rosette/menace", func(ctx *effects.Context, _ []uuid.UUID, _ game.KeyedTail) error {
	return effects.GrantKeywordUntilEOT{Match: effects.IsAttackingCreature(), Keywords: []string{"menace"}, Label: "Inquisitorial Rosette — menace"}.Apply(ctx)
})

package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aegis of the Legion — Artifact — Equipment {R}{W} (Ravnica: Clue
// Edition):
//
//	"Equipped creature gets +1/+1 and has mentor. (Whenever it attacks,
//	 put a +1/+1 counter on target attacking creature with lesser power.)
//	 Whenever equipped creature mentors a creature, put a shield counter
//	 on that creature. (If that creature would be dealt damage or
//	 destroyed, remove a shield counter from it instead.)
//	 Equip {3}"
//
// The granted mentor is the shared Mentor trigger (CR 702.134a), which
// records game.EventMentored as it resolves (#5932, CR 702.134c). The
// Equipment's own trigger watches that event for the creature it is
// attached to, and puts the shield counter on the mentored creature only
// while it is still the same object (CR 400.7).
//
// No simplification.
const aegisOfTheLegionMentor = "aegis-of-the-legion/mentor"

func init() {
	effects.Register(effects.Spec{
		OracleID:     "99eb05eb-11ef-456f-9132-166a0691dd04",
		Name:         "Aegis of the Legion",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key:       aegisOfTheLegionMentor,
			Triggered: []game.TriggeredAbility{effects.Mentor("Aegis of the Legion — mentor")},
			Text:      "Mentor (Whenever this creature attacks, put a +1/+1 counter on target attacking creature with lesser power.)",
		}},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantAbilitiesToAttached(aegisOfTheLegionMentor),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverEquippedCreatureMentors("Aegis of the Legion — put a shield counter on that creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if info, ok := ctx.TriggeringPermanent(); !ok || info.Left {
						return nil
					}
					return effects.AddCounter{Target: ctx.Trigger().Event.CardID, Kind: game.CounterShield, N: 1}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}

package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rabid Attack — Instant {1}{B} (Secrets of Strixhaven, #4456):
//
//	"Until end of turn, any number of target creatures you control each get
//	 +1/+0 and gain "When this creature dies, draw a card.""
//
// Any number of targets (zero allowed). Each target that is still legal as
// the spell resolves gets the pump and the granted dies trigger, both
// ending in this turn's cleanup step (CR 514.2). The trigger is the
// creature's own: its controller draws, from last-known information as it
// dies (CR 603.10a), and a creature that leaves and returns is a new object
// without it (CR 400.7). The grant is Feign Death's duration grant.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4db02b34-efec-437c-ab28-aa318b8879ad",
		Name:         "Rabid Attack",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("any number of target creatures you control", effects.YouControl()).WithCount(0, 0),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)),
		Grants: []effects.AbilityGrant{{
			Key: rabidAttackDiesDraw,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisDies("Rabid Attack — draw a card", effects.Do(effects.DrawCards{N: 1})),
			},
			Text: "When this creature dies, draw a card.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard || !effects.OnBattlefield(ctx.Game, t.ID) {
					continue
				}
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 1, Label: "Rabid Attack — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.GrantAbilitiesFor{
					Target: t.ID,
					Keys:   []string{rabidAttackDiesDraw},
					Label:  "Rabid Attack — it draws a card when it dies",
				}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}

const rabidAttackDiesDraw = "rabid-attack/dies-draw"

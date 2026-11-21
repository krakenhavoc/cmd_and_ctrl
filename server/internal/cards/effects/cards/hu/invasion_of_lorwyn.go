package cardshu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invasion of Lorwyn // Winnowing Forces — March of the Machine (#5477).
//
// Front, Battle — Siege 5, {4}{B}{G}: "When this Siege enters, destroy target
// non-Elf creature an opponent controls with power X or less, where X is the
// number of lands you control." Back, Winnowing Forces, Creature — Elf
// Warrior */*: "Winnowing Forces's power and toughness are each equal to the
// number of lands you control."
//
// X is counted as the target is chosen and again as the ability resolves
// (CR 608.2b); "Elf" is a word of the Siege's text (CR 612.2).
//
// No simplification.
const InvasionOfLorwynOracleID = "c877a430-75bc-45ac-b11d-2365d11ed341"

func init() {
	effects.Register(effects.Spec{
		OracleID:     InvasionOfLorwynOracleID,
		Name:         "Invasion of Lorwyn",
		Completeness: effects.CompletenessFull,
		Battle:       &effects.BattleSpec{Defense: 5, Subtype: effects.BattleSubtypeSiege},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Invasion of Lorwyn — destroy target non-Elf creature an opponent controls with power X or less",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if t := effects.FirstLegalBattlefieldTarget(ctx); t != uuid.Nil {
						return effects.DestroyTarget{Target: t}.Apply(ctx)
					}
					return nil
				}),
				effects.TargetCreature("target non-Elf creature an opponent controls with power X or less, where X is the number of lands you control",
					effects.OpponentControls(), effects.Not(effects.Subtype("Elf")), powerAtMostLandsYouControl())),
			effects.DefeatedTrigger("Invasion of Lorwyn — defeated: exile it, then cast Winnowing Forces", effects.SiegeDefeated()),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     InvasionOfLorwynOracleID + "#1",
		Name:         "Winnowing Forces",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.S82MomB17CountPT("Winnowing Forces's power and toughness are each equal to the number of lands you control.",
				func(g *game.Game, source *game.Card) int { return effects.LandsControlledBy(g, source) }),
		},
	})
}

// powerAtMostLandsYouControl is "power X or less, where X is the number of
// lands you control", "you" being the player choosing the target.
func powerAtMostLandsYouControl() effects.CardPredicate {
	return func(g *game.Game, caster uuid.UUID, c game.Card) bool {
		n := 0
		for _, p := range g.BattlefieldCardsForEffect() {
			if p.Controller == caster && p.IsLand() {
				n++
			}
		}
		return c.CurrentPower() <= n
	}
}

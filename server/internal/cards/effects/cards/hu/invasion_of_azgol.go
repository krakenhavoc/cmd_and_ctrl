package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Invasion of Azgol // Ashen Reaper — March of the Machine (#5477).
//
// Front, Battle — Siege 4, {B}{R}: "When this Siege enters, target player
// sacrifices a creature or planeswalker of their choice and loses 1 life."
// Back, Ashen Reaper, Creature — Zombie Elemental 2/1: "Menace. At the
// beginning of your end step, put a +1/+1 counter on this creature if a
// permanent was put into a graveyard from the battlefield this turn."
//
// The life is lost whether or not there was anything to sacrifice. The end
// step "if" is an intervening if (CR 603.4): asked as the step begins and
// again on resolution.
//
// No simplification.
const InvasionOfAzgolOracleID = "1abcdddd-0022-4dfe-8c15-eb1fa86de614"

func init() {
	effects.Register(effects.Spec{
		OracleID:     InvasionOfAzgolOracleID,
		Name:         "Invasion of Azgol",
		Completeness: effects.CompletenessFull,
		Battle:       &effects.BattleSpec{Defense: 4, Subtype: effects.BattleSubtypeSiege},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Invasion of Azgol — target player sacrifices a creature or planeswalker and loses 1 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					victim := effects.TargetedPlayer(ctx)
					if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
						return nil
					}
					var candidates = effects.PermanentsControlledByMatching(g, victim, effects.Or(effects.Creature(), effects.Planeswalker()))
					return effects.SacrificeChoice{
						Player:     victim,
						Candidates: candidates,
						Question:   "Invasion of Azgol — sacrifice a creature or planeswalker",
						ThenTail:   game.KeyedTail{Ref: effects.ThenChangeLife, Player: victim, Source: item.SourceCardID, N: -1},
					}.Apply(ctx)
				}), effects.TargetPlayer("target player")),
			effects.DefeatedTrigger("Invasion of Azgol — defeated: exile it, then cast Ashen Reaper", effects.SiegeDefeated()),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        InvasionOfAzgolOracleID + "#1",
		Name:            "Ashen Reaper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, func(_ game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.SocB9PermanentsPutIntoGraveyardsThisTurn(g) > 0
			}), "Ashen Reaper — put a +1/+1 counter on this creature (a permanent was put into a graveyard from the battlefield this turn)",
				func(g *game.Game, item *game.StackItem) error {
					if effects.SocB9PermanentsPutIntoGraveyardsThisTurn(g) == 0 {
						return nil
					}
					return effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})(g, item)
				}),
		},
	})
}

package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shuttlecraft Galileo — Legendary Artifact — Spacecraft {3}{W}, 3/3 (Star
// Trek Commander, #4974):
//
//	"Landfall — Whenever a land you control enters, create a 1/1 red
//	 Officer creature token.
//	 Station (Tap another creature you control: Put charge counters
//	 equal to its power on this Spacecraft. Station only as a sorcery.
//	 It's an artifact creature at 6+.)
//	 6+ | Flying
//	 Whenever Shuttlecraft Galileo attacks, up to seven target attacking
//	 creatures gain flying until end of turn."
//
// Landfall is above the bar, so it works from the moment the Spacecraft
// lands. The station body is The Seriema's (Station, SpacecraftAt,
// ThresholdKeywords), and the attack trigger is part of the same striated
// 6+ block (CR 721.2a), so it exists only from six counters on, as H.M.S.
// Bounty's combat-damage trigger does. Its targets are any zero to seven
// attacking creatures, whoever controls them, the Galileo included; each
// one still legal at resolution gains flying (CR 608.2b).
//
// No simplification.
func init() {
	attack := effects.Targeting(
		effects.WheneverThisAttacks("Shuttlecraft Galileo — up to seven target attacking creatures gain flying until end of turn",
			func(g *game.Game, item *game.StackItem) error {
				return effects.GrantEachLegalTargetUntilEOT(effects.NewContext(g, item), "flying", "Shuttlecraft Galileo — flying")
			}),
		effects.TargetCreature("up to seven target attacking creatures", effects.AttackingCreature()).WithCount(0, 7))
	attack.ActiveWhen = effects.AtChargeCounters(6)
	effects.Register(effects.Spec{
		OracleID:     "12e4e6b6-e684-49df-9373-d1e057664272",
		Name:         "Shuttlecraft Galileo",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Activated:    []effects.ActivatedAbility{effects.Station()},
		Static: append(
			effects.SpacecraftAt(6, 3, 3),
			effects.ThresholdKeywords(6, "flying"),
		),
		Triggered: []game.TriggeredAbility{
			effects.Landfall("Shuttlecraft Galileo — create a 1/1 red Officer creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Officer"), N: 1})),
			attack,
		},
	})
}

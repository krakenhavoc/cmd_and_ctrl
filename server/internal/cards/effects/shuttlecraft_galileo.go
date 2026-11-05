package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	attack := Targeting(
		WheneverThisAttacks("Shuttlecraft Galileo — up to seven target attacking creatures gain flying until end of turn",
			func(g *game.Game, item *game.StackItem) error {
				return grantEachLegalTargetUntilEOT(NewContext(g, item), "flying", "Shuttlecraft Galileo — flying")
			}),
		TargetCreature("up to seven target attacking creatures", AttackingCreature()).WithCount(0, 7))
	attack.ActiveWhen = AtChargeCounters(6)
	Register(Spec{
		OracleID:     "12e4e6b6-e684-49df-9373-d1e057664272",
		Name:         "Shuttlecraft Galileo",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Activated:    []ActivatedAbility{Station()},
		Static: append(
			SpacecraftAt(6, 3, 3),
			ThresholdKeywords(6, "flying"),
		),
		Triggered: []game.TriggeredAbility{
			Landfall("Shuttlecraft Galileo — create a 1/1 red Officer creature token",
				Do(CreateToken{Template: TokenCard("1/1 red Officer"), N: 1})),
			attack,
		},
	})
}

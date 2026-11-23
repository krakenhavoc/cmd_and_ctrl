package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Invasion of Regatha — Battle — Siege, defense 5, for {2}{R} (March of
// the Machine, #5477):
//
//	"(As a Siege enters, choose an opponent to protect it. You and
//	  others can attack it. When it's defeated, exile it, then cast it
//	  transformed.)
//	 When this Siege enters, it deals 4 damage to another target battle
//	 or opponent and 1 damage to up to one target creature."
//
// The reminder text is engine behaviour keyed on the card type
// (game/battle.go). This Spec carries the damage and the defeated
// trigger that hands over Disciples of the Inferno. The first target is
// any battle but this Siege, or any opponent; damage to a battle removes
// that many defense counters (CR 120.3h). The two damages are one
// damage instruction, and either target being gone does not stop the
// other (CR 608.2b).
//
// No simplification.
func init() {
	const name = "Invasion of Regatha"
	enters := WhenThisEnters(name+" — 4 damage to another target battle or opponent and 1 damage to up to one target creature",
		func(g *game.Game, item *game.StackItem) error {
			return DamageEachClauseTarget(4, 1)(item, NewContext(g, item))
		})
	enters.Targets = Clauses(targetBattleOrOpponent(), TargetCreature("up to one target creature").WithCount(0, 1))
	Register(Spec{
		OracleID:     InvasionOfRegathaOracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Battle: &BattleSpec{
			Defense: 5,
			Subtype: BattleSubtypeSiege,
		},
		Triggered: []game.TriggeredAbility{
			enters,
			DefeatedTrigger(name+" — defeated: exile it, then cast Disciples of the Inferno", SiegeDefeated()),
		},
	})
}

// targetBattleOrOpponent is "another target battle or opponent": an
// opponent of the caster, or a battle other than the source.
func targetBattleOrOpponent() *game.TargetSpec {
	spec := &game.TargetSpec{
		Mode:    "any",
		Label:   "another target battle or opponent",
		Players: true,
		Zones:   []game.ZoneKind{game.ZoneBattlefield},
		CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool {
			return c.IsBattle()
		},
		PlayerOK: func(g *game.Game, caster uuid.UUID, p *game.Player) bool {
			return Opponent()(g, caster, p)
		},
		Min: 1, Max: 1,
	}
	return Another(spec)
}

// invasionOfRegathaOracleID is shared with the back face's spec, which
// registers under this ID plus "#1".
const InvasionOfRegathaOracleID = "4f0bcfe5-7e52-4249-a26a-482619716f18"

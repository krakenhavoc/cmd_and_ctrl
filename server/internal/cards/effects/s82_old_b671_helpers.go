package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// untappedLandQuery is "an untapped land": a land, tapped or not, is read by its
// effective characteristics.
func untappedLandQuery() game.PermanentQuery {
	return game.PermanentQuery{Types: []string{"land"}, Untapped: true}
}

// OldB671DefendingPlayerControlsAnUntappedLand is "as long as defending player
// controls an untapped land" (Hazy Homunculus), read of the attacker as blockers
// are declared, a CantBeBlockedWhile condition: the player it attacks controls an
// untapped land (CR 508.5 for the defending player of a planeswalker or battle).
func OldB671DefendingPlayerControlsAnUntappedLand(g *game.Game, _ uuid.UUID, attacker game.Card) bool {
	defender := DefendingPlayerOfAttacker(g, attacker)
	if defender == uuid.Nil {
		return false
	}
	return g.CountMatchingForEffect(defender, []game.PermanentQuery{untappedLandQuery()}) > 0
}

// OldB671BrawlerSpec is the whole of Veteran Brawlers and Branded Brawlers, which
// print the same two lines on different bodies: "This creature can't attack if
// defending player controls an untapped land. This creature can't block if you
// control an untapped land." The attack half is CantAttackIfDefendingPlayerControls
// (asked of each target's defending player at declaration, CR 508.1c); the block
// half is a CantBlockUnless on its controller controlling no untapped land, read
// as blockers are declared (CR 509.1b).
func OldB671BrawlerSpec(oracleID, name string) Spec {
	return Spec{
		OracleID:     oracleID,
		Name:         name,
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{CantAttackIfDefendingPlayerControls(untappedLandQuery())},
		BlockRules: []game.BlockRule{CantBlockUnless(func(g *game.Game, you uuid.UUID) bool {
			return g.CountMatchingForEffect(you, []game.PermanentQuery{untappedLandQuery()}) == 0
		}, "it can't block if its controller controls an untapped land")},
	}
}

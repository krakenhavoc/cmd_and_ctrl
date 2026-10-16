package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weathered Sentinels — Artifact Creature — Wall {3}, 2/5:
//
//	"Defender, reach, vigilance, trample
//	 This creature can attack players who attacked you during their last
//	 turn as though it didn't have defender.
//	 Whenever this creature attacks, it gets +3/+3 and gains
//	 indestructible until end of turn."
//
// The waiver is ADR 0106's 2026-10-16 amendment (#1936): a DefenderWaiver
// the engine asks of each attack target, against ADR 0108 §6's record of
// whom each player attacked during their last turn. In Commander it may
// attack exactly the opponents who attacked you (not a planeswalker or
// battle of theirs: those are not players), and nobody else; the
// declaration verbs, the legal-move enumerator, the client's per-attacker
// targets and the bot all read the same check. It keeps defender
// throughout (CR 609.4).
//
// The attack trigger is the ordinary until-end-of-turn pump and keyword
// grant on itself, skipped if it has left the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0903b941-82f1-4d3e-9717-29ab7d0c8cef",
		Name:            "Weathered Sentinels",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"defender", "reach", "vigilance", "trample"},
		Static:          []game.StaticAbility{CanAttackPlayersWhoAttackedYouAsThoughNoDefender()},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Weathered Sentinels — it gets +3/+3 and gains indestructible until end of turn",
				weatheredSentinelsAttack),
		},
	})
}

func weatheredSentinelsAttack(g *game.Game, item *game.StackItem) error {
	if !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (BoostUntilEOT{Target: item.SourceCardID, Power: 3, Toughness: 3, Label: "Weathered Sentinels — +3/+3"}).Apply(ctx); err != nil {
		return err
	}
	return GrantKeywordUntilEOT{
		Target:   item.SourceCardID,
		Keywords: []string{"indestructible"},
		Label:    "Weathered Sentinels — indestructible",
	}.Apply(ctx)
}

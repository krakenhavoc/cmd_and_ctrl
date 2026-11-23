package cardsh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hammers of Moradin — Creature — Dwarf Cleric {2}{W}, 3/3 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Myriad (Whenever this creature attacks, for each opponent other than
//	 defending player, you may create a token copy that's tapped and
//	 attacking that player or a planeswalker they control. Exile the tokens
//	 at end of combat.)
//	 Whenever this creature attacks, for each opponent, tap up to one target
//	 creature that player controls."
//
// Myriad is the shared trigger (myriad.go, CR 702.116), a separate ability
// from the tap trigger, so both go on the stack. The tap trigger is Kang
// Dynasty's per-opponent "up to one" clauses (PerOpponentClauses): one
// optional target for each opponent, and each pick still legal as it
// resolves is tapped (CR 608.2b). Myriad's token copies enter attacking
// without being declared, so they fire no attack trigger (CR 508.4).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97e260c9-3e78-402b-8d97-af4b515b278b",
		Name:         "Hammers of Moradin",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GrantedMyriad("Hammers of Moradin"),
			effects.Targeting(
				effects.WheneverThisAttacks("Hammers of Moradin — for each opponent, tap up to one target creature that player controls",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard || t.ID == uuid.Nil {
								continue
							}
							if err := (effects.TapTarget{Target: t.ID}).Apply(ctx); err != nil {
								return err
							}
						}
						return nil
					}),
				effects.PerOpponentClauses(game.MaxPlayers-1, "creature", effects.Creature())),
		},
	})
}

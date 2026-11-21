package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Prudent Fateseer // Peer Review — Creature — Dwarf Wizard
// {1}{W/U}{W/U}, 1/4 // Sorcery {2}{W/U} (preparation card, CR 722):
//
//	"This creature enters prepared.
//	 Whenever you scry or surveil, creatures you control get +1/+0 until
//	 end of turn. This ability triggers only once each turn."
//
//	Peer Review — "Create a 2/2 colorless Wizard Soldier creature token
//	 named Cadet. Surveil 1."
//
// Scry and surveil are two event kinds on one ability. The once-a-turn
// limit counts abilities already put on the stack this turn (a
// countered one still counts, as printed), so Peer Review's own surveil
// pumps the team if it is the first of the turn.
//
// No simplification.
func init() {
	const id = "63f67baa-3f37-41b8-9ee6-3c37f828917d"
	const label = "Prudent Fateseer — creatures you control get +1/+0 until end of turn"
	effects.Register(effects.Spec{
		OracleID:     id,
		Name:         "Prudent Fateseer",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersPrepared()},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventScry, game.EventSurveil},
				func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
					return effects.ByYou(ev, source, lki, g) && !effects.B11TriggeredThisTurn(g, source.InstanceID, label)
				},
				label,
				effects.Do(effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1, Label: label})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     id + "#1",
		Name:         "Peer Review",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.PeerReviewResolve,
	})
}

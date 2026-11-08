package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Private Eye — Creature — Homunculus Detective {1}{W}{U}, 3/3 (Murders
// at Karlov Manor, #4469):
//
//	"Other Detectives you control get +1/+1.
//	 Whenever you draw your second card each turn, target Detective
//	 can't be blocked this turn."
//
// The anthem is the shared TribalAnthem with "other" and "you control"
// (Arahbo's). "Your second card each turn" is Detective of the Month's
// tally read (b41DrewYourSecondCardThisTurn), so it fires once a turn,
// on any player's turn. The target is any Detective, yours or not, and
// the evasion is Escape Tunnel's turn-long CantBeBlocked restriction on
// the creature still targeted as the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "20750317-061b-4b65-9fc8-43477fcd7a21",
		Name:         "Private Eye",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Detective"}, Others: true, YoursOnly: true}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return b41DrewYourSecondCardThisTurn(ev, source, g)
				}, "Private Eye — target Detective can't be blocked this turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							return RestrictUntilEOT{
								Target:       t.ID,
								Restrictions: game.CantBeBlocked,
								Label:        "Private Eye — can't be blocked",
							}.Apply(ctx)
						}
						return nil
					}),
				TargetCreature("target Detective", HasSubtype("Detective"))),
		},
	})
}

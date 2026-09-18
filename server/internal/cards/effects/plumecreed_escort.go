package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Plumecreed Escort — Creature — Bird Scout {1}{U}, 2/1 (EDHREC rank
// 4328):
//
//	"Flash
//	 Flying
//	 When this creature enters, target creature you control gains
//	 hexproof until end of turn."
//
// Two mana for a flash flier that also answers the removal spell that
// made you cast it. The whole card is the timing: flash plus an entry
// trigger means it is a protection spell that leaves a 2/1 behind, and
// the 2/1 is why it is played over a Ranger's Guile.
//
// The Escort can target ITSELF — "target creature you control" has no
// "another", and the Escort is on the battlefield by the time its own
// entry trigger goes on the stack. Flashing it in with nothing else on
// board is a legal and occasionally correct play.
//
// Hexproof is a canonical engine keyword read by the targeting gate,
// so the grant is real for the rest of the turn: it stops opponents
// targeting the creature and, unlike shroud, leaves your own spells
// free to. The affected creature is pinned at resolution (CR 611.2c),
// so one flickered in response comes back a new object without it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "04669a6f-6299-465b-a26e-6cae37cdc081",
		Name:            "Plumecreed Escort",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Plumecreed Escort — target creature you control gains hexproof",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						for _, t := range ctx.LegalTargets() {
							if t.Kind != game.TargetCard {
								continue
							}
							return GrantKeywordUntilEOT{
								Target:   t.ID,
								Keywords: []string{"hexproof"},
								Label:    "Plumecreed Escort — hexproof until end of turn",
							}.Apply(ctx)
						}
						return nil
					}),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

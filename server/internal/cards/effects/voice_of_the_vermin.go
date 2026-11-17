package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voice of the Vermin — Creature — Human Citizen {3}{G}, 2/2 (Streets
// of New Capenna):
//
//	"This creature enters with a shield counter on it. (If it would be
//	 dealt damage or destroyed, remove a shield counter from it
//	 instead.)
//	 Whenever this creature attacks, target creature you control has
//	 base power and toughness 4/4 until end of turn."
//
// Layer 7b sets the base; +1/+1 counters and pumps still apply on top
// (CR 613.4b, 613.4c). The Voice may pick itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "73829ae3-c8de-4e6d-9607-4b2e980ff144",
		Name:         "Voice of the Vermin",
		Completeness: CompletenessFull,
		Replacements: snc6Entry("Voice of the Vermin"),
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Voice of the Vermin — target creature you control has base power and toughness 4/4 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return untilEndOfTurn(ctx, id, nil, "Voice of the Vermin — base power and toughness 4/4", game.SetBasePTMods(4, 4)...)
				}),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}

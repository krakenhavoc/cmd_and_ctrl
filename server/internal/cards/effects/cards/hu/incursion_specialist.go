package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Incursion Specialist — Creature — Human Wizard {1}{U}, 1/3 (Gatecrash, S82
// OLD #5565):
//
//	"Whenever you cast your second spell each turn, this creature gets +2/+0
//	 until end of turn and can't be blocked this turn."
//
// Bloodsky Berserker's second-spell trigger (the per-turn cast tally already
// counts the spell that fired it), with Basim Ibn Ishaq's unblockable rider. The
// rider is only worth anything before blockers are declared (CR 509.1b), which
// is what the restriction judges. A Specialist that has left does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3c334eb3-e3fd-45b0-a36d-4681c9e72ac5",
		Name:         "Incursion Specialist",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.YouCastYourSecondSpellEachTurn,
				"Incursion Specialist — it gets +2/+0 until end of turn and can't be blocked this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if !effects.OnBattlefield(g, item.SourceCardID) || g.AbilitySourceIsNewObjectForEffect(item) {
						return nil
					}
					if err := (effects.BoostUntilEOT{Target: item.SourceCardID, Power: 2, Label: "Incursion Specialist — +2/+0"}).Apply(ctx); err != nil {
						return err
					}
					return effects.RestrictUntilEOT{Target: item.SourceCardID, Restrictions: game.CantBeBlocked, Label: "Incursion Specialist — can't be blocked this turn"}.Apply(ctx)
				}),
		},
	})
}

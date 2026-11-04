package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Poisoner's Apprentice — Creature — Orc Warlock {2}{B}, 2/2 (Secrets of
// Strixhaven, #4456):
//
//	"Infusion — When this creature enters, target creature an opponent
//	 controls gets -4/-4 until end of turn if you gained life this turn."
//
// Infusion is an ability word (CR 207.2c) and does nothing itself. The "if"
// is in the effect, not after "When" (CR 603.4 is about an "if" right after
// the trigger event), so the trigger always goes on the stack and always
// picks a target; the life-gain check is made as it resolves, against the
// turn's tally of life gained. If the creature is no longer a legal target
// the ability does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4b0e5626-0282-42e5-abef-a079b26a8e5f",
		Name:         "Poisoner's Apprentice",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Poisoner's Apprentice — target creature an opponent controls gets -4/-4 until end of turn if you gained life this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if b15LifeGainedThisTurn(g, ctx.Controller()) < 1 {
						return nil
					}
					id, ok := b16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return BoostUntilEOT{
						Target:    id,
						Power:     -4,
						Toughness: -4,
						Label:     "Poisoner's Apprentice — -4/-4 until end of turn",
					}.Apply(ctx)
				}),
				TargetCreature("target creature an opponent controls", OpponentControls())),
				ForTargets(HarmUntilEOT(0, -4, -4))),
		},
	})
}

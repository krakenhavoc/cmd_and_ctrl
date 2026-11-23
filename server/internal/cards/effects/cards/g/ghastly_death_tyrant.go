package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghastly Death Tyrant — Creature — Beholder Skeleton {4}{B}{B}, 6/5
// (Commander Legends: Battle for Baldur's Gate, #5948):
//
//	"When this creature enters, choose one —
//	 • Disintegration Ray — Destroy target enchantment an opponent controls.
//	   You lose life equal to its mana value.
//	 • Death Ray — Creatures you control gain deathtouch until end of turn."
//
// Baleful Beholder's modal enters trigger. The first mode's target is chosen
// when the trigger goes on the stack with the mode (CR 700.2b), so a trigger
// with no enchantment to hit can still choose the second. The mana value is
// read off the enchantment as the ability resolves, before it is destroyed
// (an {X} counts as 0, CR 202.3e), and the life is lost whether or not the
// destruction happened (indestructible, regenerated): they are two sentences.
// If the enchantment is no longer a legal target the whole mode does nothing
// (CR 608.2b). Death Ray affects the creatures you control as it resolves
// (CR 611.2c).
//
// No simplification.
func init() {
	tyrant := effects.WhenThisEnters("Ghastly Death Tyrant — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	tyrant.Modes = effects.ChooseOne(
		effects.ModeDoing("Disintegration Ray — Destroy target enchantment an opponent controls. You lose life equal to its mana value.",
			effects.TargetPermanent("target enchantment an opponent controls", effects.Enchantment(), effects.OpponentControls()),
			func(item *game.StackItem, ctx *effects.Context, occ int) error {
				t, ok := effects.ModeTarget(ctx, occ)
				if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					return nil
				}
				mv := 0
				if c, found := ctx.Game.LookupCardForEffect(t.ID); found {
					mv = c.ManaValue()
				}
				return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{t.ID}, ctx.TailN(ghastlyDeathTyrantLoseLife, mv))
			}),
		effects.ModeDoing("Death Ray — Creatures you control gain deathtouch until end of turn.", nil,
			func(_ *game.StackItem, ctx *effects.Context, _ int) error {
				return effects.GrantKeywordUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Keywords: []string{"deathtouch"},
					Label: "Ghastly Death Tyrant — Death Ray: deathtouch until end of turn"}.Apply(ctx)
			}),
	)
	effects.Register(effects.Spec{
		OracleID:     "ccd248ec-1029-47cd-b6b4-876505efd9dd",
		Name:         "Ghastly Death Tyrant",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{tyrant},
	})
}

// ghastlyDeathTyrantLoseLife is "you lose life equal to its mana value", run
// once the destruction has settled (it may have stopped for a replacement's
// prompt) and whether or not the enchantment was destroyed: the amount was
// read off the enchantment before. An on-disk key: never renamed, never reused.
var ghastlyDeathTyrantLoseLife = effects.CardTail("then/ghastly-death-tyrant-lose-life", func(ctx *effects.Context, o game.TailOutcome) error {
	if o.Tail.N <= 0 {
		return nil
	}
	return ctx.Game.ChangePlayerLifeForEffect(o.Tail.Source, o.Tail.Player, -o.Tail.N)
})

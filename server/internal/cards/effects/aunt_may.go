package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aunt May — Legendary Creature — Human Citizen {W}, 0/2 (Marvel Super
// Heroes Commander, #4455):
//
//	"Whenever another creature you control enters, you gain 1 life. If
//	 it's a Spider, put a +1/+1 counter on it."
//
// The life is the Dazzling Angel trigger; the counter goes on the
// creature that entered, read from the trigger's object, and only if it
// is still on the battlefield and a Spider as the ability resolves.
func init() {
	Register(Spec{
		OracleID:     "bdb769b5-c861-4ec1-ad6b-e2a56ca05fd1",
		Name:         "Aunt May",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Aunt May — you gain 1 life; if it's a Spider, put a +1/+1 counter on it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (GainLife{Amount: 1}).Apply(ctx); err != nil {
						return err
					}
					info, ok := ctx.TriggeringPermanent()
					if !ok || info.Left {
						return nil
					}
					ref := ctx.Trigger().Object.Ref()
					c, ok := g.LookupCardForEffect(ref.ID)
					if !ok || !c.HasSubtypeWord("Spider") {
						return nil
					}
					return AddCounter{Target: ref.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
				}),
		},
	})
}

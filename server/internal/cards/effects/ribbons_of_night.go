package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ribbons of Night — Sorcery {4}{B} (EDHREC rank 29633):
//
//	"Ribbons of Night deals 4 damage to target creature and you gain
//	 4 life. If {U} was spent to cast this spell, draw a card."
//
// The same "if {C} was spent" family as Gruul Scrapper, read from the
// other side of the wall: this one is a SPELL asking about its own
// payment while it resolves, so the record is still on the stack item
// and `ctx.ManaSpent()` is the whole of it. Gruul Scrapper is the
// permanent-side version and has to go through Card.Provenance; the
// two return the same value type (`game.ManaSpent`) on purpose, which
// is the point of #1212's one vocabulary.
//
// The {U} is paid out of the generic half, exactly as Gruul Scrapper's
// {R} is: a Dimir player taps an Island for one of the four and draws,
// a mono-black one does not. Since #3853 the card declares WantsColors,
// so an auto-paid cast taps the Island for the generic half rather than
// the Swamp beside it (an ordering, not a filter: no Island, no draw).
//
// Printed order, which is also the only order: the damage and the life
// happen whatever was spent, and the draw is the rider. A target that
// has left the battlefield in response takes no damage and the
// controller still gains the life, because "and you gain 4 life" is
// not conditional on the damage — CR 608.2b's partial-fizzle rule
// applied to a spell with one target.
func init() {
	Register(Spec{
		OracleID:       "f074c0d0-8455-484d-ae75-820fdfbc4740",
		Name:           "Ribbons of Night",
		ReadsManaSpent: game.ReadsOwnPayment,
		Completeness:   CompletenessFull,
		WantsColors:    []string{"U"},
		Targets:        TargetCreature("target creature"),
		Purpose:        ForTargets(DamageToTarget(0, 4)),
		// The life and the draw wait for a prompt the damage asks (a CR 616 order; CR 608.2c,
		// #4155), and the draw for one the life asks.
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return ctx.Game.DamageInstanceForEffect(func() error {
					for _, ref := range ctx.LegalTargets() {
						if err := (DealDamage{
							Source: ctx.Item.SourceCardID,
							Target: ref.ID,
							Amount: 4,
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				})
			}),
			Instruction(func(ctx *Context) error {
				return GainLife{Player: ctx.Item.Controller, Amount: 4}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if ctx.ManaSpent().Count("U") > 0 {
					return DrawCards{Player: ctx.Item.Controller, N: 1}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}

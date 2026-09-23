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
// a mono-black one does not. No wish is declared for the same reason —
// a colour is not a mana SOURCE, and the solver has no switch that
// concentrates a payment into one (ADR 0068 §5).
//
// Printed order, which is also the only order: the damage and the life
// happen whatever was spent, and the draw is the rider. A target that
// has left the battlefield in response takes no damage and the
// controller still gains the life, because "and you gain 4 life" is
// not conditional on the damage — CR 608.2b's partial-fizzle rule
// applied to a spell with one target.
func init() {
	Register(Spec{
		OracleID:     "f074c0d0-8455-484d-ae75-820fdfbc4740",
		Name:         "Ribbons of Night",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't see which mana you spent, so Ribbons of Night never draws the card."},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, ref := range ctx.LegalTargets() {
				if err := (DealDamage{
					Source: item.SourceCardID,
					Target: ref.ID,
					Amount: 4,
				}).Apply(ctx); err != nil {
					return err
				}
			}
			if err := (GainLife{Player: item.Controller, Amount: 4}).Apply(ctx); err != nil {
				return err
			}
			if ctx.ManaSpent().Count("U") > 0 {
				return (DrawCards{Player: item.Controller, N: 1}).Apply(ctx)
			}
			return nil
		},
	})
}

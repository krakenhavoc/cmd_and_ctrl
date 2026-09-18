package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Airbending Lesson — Instant — Lesson {2}{W} (EDHREC rank 4100):
//
//	"Airbend target nonland permanent. (Exile it. While it's exiled,
//	 its owner may cast it for {2} rather than its mana cost.)
//	 Draw a card."
//
// White's instant-speed answer to anything that is not a land, with
// the card replaced. Three mana to exile a commander, a
// game-ending enchantment or an indestructible artifact is a rate
// white does not normally get; the refund is the price, and against
// a big permanent it is a real tempo swing — the owner has to spend
// a whole turn and {2} putting it back, and the airbent card loses
// every counter and Aura it was wearing (CR 400.7, it returns a new
// object).
//
// AIRBEND IS THE SHARED KEYWORD ACTION, not a bespoke exile: the
// permission goes to the card's OWNER, not to the caster, which is
// what keeps this removal-with-a-refund rather than theft. The grant
// is unbounded ("while it's exiled") and cast-only, so an airbent
// land would be stranded — which is precisely why the printed target
// clause says NONLAND permanent and why this file spells that
// restriction into the target spec rather than leaving it to the
// primitive.
//
// "TARGET NONLAND PERMANENT" is wider than every other airbend card
// in the catalog, which all name creatures: this one also takes
// artifacts, enchantments, planeswalkers and battles. It can target
// your own, which is occasionally the right play — airbending your
// own creature in response to an exile effect gets it back for {2}
// instead of never.
//
// THE DRAW IS UNCONDITIONAL AND HAPPENS SECOND. If the target became
// illegal in response the spell is countered by the rules (CR 608.2b)
// and there is no draw; if the target is still legal, the exile and
// then the draw both happen. The Lesson subtype does nothing on its
// own (there is no Learn in the catalog), so it is type-line data
// only.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e101b744-9175-4aa5-bf80-d6944359e538",
		Name:         "Airbending Lesson",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard || !onBattlefield(ctx.Game, t.ID) {
					continue
				}
				if err := (Airbend{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				break
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}

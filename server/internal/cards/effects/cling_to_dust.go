package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cling to Dust — Instant {B}:
//
//	"Exile target card from a graveyard. If it was a creature card,
//	 you gain 3 life. Otherwise, you draw a card.
//	 Escape—{3}{B}, Exile five other cards from your graveyard."
//
// Graveyard hate that never runs out, and the escape card that best
// shows the cost and the effect pulling in opposite directions: the
// spell exiles ONE card from ANY graveyard, and escaping it exiles
// five more from YOUR OWN. The second cast is therefore a real
// decision rather than a free repeat.
//
// The "if it was a creature card" clause reads the type BEFORE the
// exile — same shape as Scavenging Ooze — because after the move the
// card is in exile and the question is about what it was.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "9c44e556-4c7a-48bf-a108-c584f69c2cfa",
		Name:             "Cling to Dust",
		Completeness:     CompletenessFull,
		Targets:          TargetCardInGraveyard("target card from a graveyard"),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Escape("{3}{B}", 5)},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					continue
				}
				wasCreature := c.IsCreature()
				if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if wasCreature {
					return GainLife{Player: item.Controller, Amount: 3}.Apply(ctx)
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}

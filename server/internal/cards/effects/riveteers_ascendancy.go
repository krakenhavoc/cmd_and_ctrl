package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Riveteers Ascendancy — Enchantment {B}{R}{G} (Streets of New Capenna,
// S82):
//
//	"Whenever you sacrifice a creature, you may return target creature card
//	 with lesser mana value from your graveyard to the battlefield tapped.
//	 Do this only once each turn."
//
// Jackdaw Savior's "lesser" (TargetsFrom): lesser than the mana value of the
// creature that was sacrificed, read off the trigger's last-known snapshot
// (CR 603.10), so the target is chosen as the trigger goes on the stack
// and the graveyard card must be cheaper; with nothing cheaper the trigger
// is not put on the stack (CR 603.3d). "You may" is the trigger's resolution
// question. "Do this only once each turn" is Ancient Cornucopia's record:
// the turn is marked only when the return really happens, so a "no" (or a
// target that has gone, which fizzles the ability) leaves a later sacrifice
// free to do it, and a second trigger already on the stack does nothing
// once the first has returned a card.
//
// No simplification.
func init() {
	const label = "Riveteers Ascendancy — return target creature card with lesser mana value from your graveyard to the battlefield tapped"
	Register(Spec{
		OracleID:     "71f3168d-9eeb-4570-8437-a3ffc39667e9",
		Name:         "Riveteers Ascendancy",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{Optional(game.TriggeredAbility{
			Watches: []game.EventKind{game.EventSacrifice},
			Key:     label,
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if source == nil || ev.Actor != source.Controller || g.ReplacementAppliedThisTurnForEffect(source, label) {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature()
			},
			TargetsFrom: riveteersAscendancyClause,
			Effect: func(g *game.Game, item *game.StackItem) error {
				src, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || g.ReplacementAppliedThisTurnForEffect(&src, label) {
					return nil
				}
				ctx := NewContext(g, item)
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				g.MarkReplacementAppliedThisTurnForEffect(&src, label)
				return ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: ctx.Controller(), Tapped: true}.Apply(ctx)
			},
		}, "Riveteers Ascendancy — return a creature card with lesser mana value from your graveyard to the battlefield tapped?")},
	})
}

// riveteersAscendancyClause is "target creature card with lesser mana value
// from your graveyard", the value being the sacrificed creature's.
func riveteersAscendancyClause(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
	ceiling := 0
	if tc.Object != nil {
		ceiling = tc.Object.ManaValue
	}
	return TargetCardInGraveyard(
		"target creature card with lesser mana value from your graveyard",
		YouOwn(), Creature(), ManaValueLE(ceiling-1),
	)
}

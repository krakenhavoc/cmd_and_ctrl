package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Puca's Covenant — Enchantment {2}{G}, (Lorwyn Eclipsed Commander, #5368):
//
//	"Whenever a creature you control with a counter on it dies, you may
//	 return another target permanent card with mana value less than or equal
//	 to the number of counters on that creature from your graveyard to your
//	 hand. Do this only once each turn."
//
// The trigger is a creature you controlled as it died that last had a counter
// of any kind (CR 603.10a). Its target is built per trigger (TargetsFrom) from
// the number of counters the creature had, read off the triggering event's
// last-known snapshot (ObjectSnapshot.Counters): a permanent card, not the creature
// itself, in your graveyard, with mana value at most that number. The "you
// may" is asked as it resolves (Optional), and "only once each turn" limits
// the DOING (b28DoneThisTurn), as Baron Strucker's does, so a declined "may"
// leaves the turn's one return for a later creature.
//
// No simplification.
func init() {
	const label = "Puca's Covenant — you may return another target permanent card from your graveyard to your hand"
	Register(Spec{
		OracleID:     "9b832f87-0bd5-47bd-ab9c-718bd9a6c441",
		Name:         "Puca's Covenant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Optional(game.TriggeredAbility{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
					return ACreatureYouControlDied(ev, source, ch, g) && cmdr2B25CreatureDiedWithACounter(ev, g) &&
						!b28DoneThisTurn(g, source.InstanceID, cmdr2B25PucaDone)
				},
				Key: label,
				TargetsFrom: func(tc game.TriggerContext, _ *game.Card, _ *game.Game) *game.TargetSpec {
					n := 0
					if tc.Object != nil {
						n = cmdr2B25CounterTotal(tc.Object.Counters)
					}
					return TargetCardInGraveyard("another target permanent card with mana value "+itoaB25(n)+" or less from your graveyard",
						YouOwn(), Permanent(), ManaValueLE(n), OtherThan(tc.Event.CardID))
				},
				Effect: func(g *game.Game, item *game.StackItem) error {
					if b28DoneThisTurn(g, item.SourceCardID, cmdr2B25PucaDone) {
						return nil
					}
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						b28NoteDoneThisTurn(g, item.SourceCardID, cmdr2B25PucaDone)
						return ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}.Apply(ctx)
					}
					return nil
				},
			}, "Puca's Covenant — return the target card to your hand?"),
		},
	})
}

const cmdr2B25PucaDone = "Puca's Covenant — a card was returned this turn"

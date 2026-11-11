package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rapid Augmenter — Creature — Otter Artificer {1}{U}{R}, 1/3 (Bloomburrow
// Commander):
//
//	"Haste
//	 Whenever another creature you control with base power 1 enters, it gains
//	 haste until end of turn.
//	 Whenever another creature you control enters, if it wasn't cast, put a
//	 +1/+1 counter on this creature and this creature can't be blocked this
//	 turn."
//
// Two enter triggers over creatures its controller controls. Base power is read
// after layer 7b as the creature enters (CR 208.4b). "Wasn't cast" is the
// entry's own history (cmdr2B21EnteredWithoutBeingCast): any entry that did not
// come from the stack, a token, a reanimated or fetched creature or a flicker
// included. Both ask once per entering creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "098a43e3-341d-4b27-8c17-ef5990e21b0a",
		Name:            "Rapid Augmenter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := cmdr2B21AnotherCreatureYouControlEntered(ev, source, g)
				return ok && c.BasePower() == 1
			}, "Rapid Augmenter — the creature gains haste until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return GrantKeywordUntilEOT{Target: item.Trigger.Event.CardID, Keywords: []string{"haste"},
						Label: "Rapid Augmenter — haste until end of turn"}.Apply(NewContext(g, item))
				}),
			On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := cmdr2B21AnotherCreatureYouControlEntered(ev, source, g)
				return ok && cmdr2B21EnteredWithoutBeingCast(g, ev.CardID)
			}, "Rapid Augmenter — put a +1/+1 counter on this creature and it can't be blocked this turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if !onBattlefield(g, ctx.Source()) {
						return nil
					}
					if err := (AddCounter{Target: ctx.Source(), Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
					return RestrictUntilEOT{Target: ctx.Source(), Restrictions: game.CantBeBlocked,
						Label: "Rapid Augmenter — can't be blocked this turn"}.Apply(ctx)
				}),
		},
	})
}

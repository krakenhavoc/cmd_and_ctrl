package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Oft-Nabbed Goat — Creature — Goat {1}{B}, 0/5 (Lorwyn Eclipsed Commander,
// #5368):
//
//	"{1}: Draw a card. Gain control of this creature and put a -1/-1
//	 counter on it. Only your opponents may activate this ability and only
//	 as a sorcery.
//	 When this creature dies, if it had one or more -1/-1 counters on it,
//	 its owner draws that many cards and each other player loses that much
//	 life."
//
// Only the controller's opponents may activate it (OpponentsOnly), at sorcery
// speed. The activator, who is the ability's controller, draws, gains control
// of the Goat with no stated duration and puts a -1/-1 counter on it, in that
// order (the control change is pinned to the Goat as it is now, CR 400.7).
// The dies trigger is an intervening "if" (CR 603.4): the counters it last
// had are read at the event (CR 603.10a), and carried as the amount, so a
// Goat that died with none does nothing. Its owner, not its controller, draws.
// Each other player is every player but the owner, the Goat's controller
// included.
//
// No simplification.
func init() {
	const diesLabel = "Oft-Nabbed Goat — its owner draws a card for each -1/-1 counter on it and each other player loses that much life"
	Register(Spec{
		OracleID:     "63193ad4-bbd2-4cc3-98e2-9f6b985e8ae7",
		Name:         "Oft-Nabbed Goat",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:         "{1}: Draw a card. Gain control of this creature and put a -1/-1 counter on it. Only your opponents may activate this ability and only as a sorcery.",
			Cost:          ManaCost("{1}"),
			OpponentsOnly: true,
			SorcerySpeed:  true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
					return err
				}
				if err := (GainControl{
					Target:     ctx.Source(),
					Controller: item.Controller,
					Duration:   game.IndefiniteDuration(),
					Label:      "Oft-Nabbed Goat — gain control (no stated duration)",
				}).Apply(ctx); err != nil {
					return err
				}
				return AddCounter{Target: ctx.Source(), Kind: game.CounterMinusOne, N: 1}.Apply(ctx)
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, ch game.Characteristic, g *game.Game) bool {
				return ThisDied(ev, source, ch, g) && g.LastKnownCountersForEffect(ev.CardID)[game.CounterMinusOne] > 0
			},
			Key: diesLabel,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, diesLabel)
				item.Params.Amount = g.LastKnownCountersForEffect(ev.CardID)[game.CounterMinusOne]
				if c, ok := g.LookupCardForEffect(ev.CardID); ok {
					item.Params.Player = c.Owner
				}
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				owner, n := item.Params.Player, item.Params.Amount
				if n <= 0 || owner == uuid.Nil {
					return nil
				}
				if err := (DrawCards{Player: owner, N: n}).Apply(ctx); err != nil {
					return err
				}
				for _, p := range g.Seats {
					if p.ID == owner || g.PlayerByIDForEffect(p.ID) == nil {
						continue
					}
					if err := g.ChangePlayerLifeForEffect(ctx.Source(), p.ID, -n); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

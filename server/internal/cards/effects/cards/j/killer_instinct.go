package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Killer Instinct — Enchantment {4}{R}{G} (Guildpact, S82 OLD batch 573,
// #5599):
//
//	"At the beginning of your upkeep, reveal the top card of your library. If
//	 it's a creature card, put it onto the battlefield. That creature gains
//	 haste until end of turn. Sacrifice it at the beginning of the next end
//	 step."
//
// The top card is revealed (a card that is not a creature stays on top). A
// creature card is put onto the battlefield without being cast, then gets
// haste until end of turn and a delayed trigger that sacrifices that very
// permanent at the next end step (CR 603.7); a permanent that has left by
// then is not sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bbf90463-280a-413e-b8a3-4577dcf11e54",
		Name:         "Killer Instinct",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourUpkeep("Killer Instinct — reveal the top card of your library; if it's a creature card, put it onto the battlefield with haste and sacrifice it at the next end step",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					revealed := g.RevealTopOfLibraryForEffect(ctx.Controller(), ctx.Source(), 1, "Killer Instinct — reveal the top card of your library")
					if len(revealed) != 1 {
						return nil
					}
					c, ok := g.LookupCardForEffect(revealed[0])
					if !ok || !c.IsCreature() {
						return nil
					}
					entered, err := g.PutFromLibraryOntoBattlefieldForEffect(revealed[0], game.LibraryEntryOptions{Controller: ctx.Controller()})
					if err != nil || entered == uuid.Nil {
						return err
					}
					if err := (effects.GrantKeywordUntilEOT{Target: entered, Keywords: []string{"haste"},
						Label: "Killer Instinct — the creature gains haste until end of turn"}).Apply(ctx); err != nil {
						return err
					}
					return effects.ScheduleDelayedTrigger{
						At:         game.StepEnd,
						Controller: ctx.Controller(),
						Label:      "Killer Instinct — sacrifice the creature",
						Cards:      []uuid.UUID{entered},
						Body:       effects.SacrificeListedCardsBody,
					}.Apply(ctx)
				}),
		},
	})
}

package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meren of Clan Nel Toth — Legendary Creature — Human Shaman {2}{B}{G}, 3/4:
//
//	"Whenever another creature you control dies, you get an experience
//	 counter.
//	 At the beginning of your end step, choose target creature card in your
//	 graveyard. If that card's mana value is less than or equal to the
//	 number of experience counters you have, return it to the battlefield.
//	 Otherwise, put it into your hand."
//
// The first ability counts only creatures other than Meren (Aang, A Lot to
// Learn's "another creature you control dies"). The end-step trigger
// targets a creature card in your graveyard when it goes on the stack
// (CR 603.3d: no creature card, no trigger) and reads your experience
// counters and the card's mana value as it resolves: the first Meren
// trigger of a game, with no counters, sends a mana value 2 card to hand
// and a mana value 0 card to the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0310dfdb-0498-488d-9f1a-279f8239e024",
		Name:         "Meren of Clan Nel Toth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AnotherCreatureYouControlDied(ev, source, g)
			}, "Meren of Clan Nel Toth — you get an experience counter", effects.YouGetAnExperienceCounter),
			effects.Targeting(
				effects.AtYourEndStep("Meren of Clan Nel Toth — return the chosen creature card to the battlefield if its mana value is at most your experience counters, otherwise to your hand",
					merenEndStep),
				effects.TargetCardInGraveyard("target creature card in your graveyard", effects.YouOwn(), effects.Creature())),
		},
	})
}

// merenEndStep returns the chosen card to the battlefield when its mana
// value is within the controller's experience counters, else to hand.
func merenEndStep(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := g.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		dest := game.ZoneHand
		if c.ManaValue() <= effects.ExperienceCounters(g, item.Controller) {
			dest = game.ZoneBattlefield
		}
		return effects.ReturnFromGraveyard{Target: t.ID, Dest: dest}.Apply(ctx)
	}
	return nil
}

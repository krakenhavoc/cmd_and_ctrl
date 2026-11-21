package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Puppeteer Clique — Creature — Faerie Wizard {3}{B}{B}, 3/2:
//
//	"Flying
//	 When this creature enters, put target creature card from an
//	 opponent's graveyard onto the battlefield under your control. It
//	 gains haste. At the beginning of your next end step, exile it.
//	 Persist"
//
// "It gains haste" has no duration, so it lasts as long as the creature
// stays (CR 611.2a). The exile is a delayed trigger at the beginning of
// YOUR next end step (CR 603.7). If an entry replacement sent the card
// elsewhere, nothing gains haste and nothing is exiled. A persisted
// Clique reanimates a second creature. Flying and persist are
// PrintedKeywords (#2075).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "00e0b103-892c-49b3-836d-867fff197bbd",
		Name:            "Puppeteer Clique",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordPersist},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Puppeteer Clique — put a creature card from an opponent's graveyard onto the battlefield under your control",
				puppeteerCliqueReanimate),
				effects.TargetCardInGraveyard("target creature card from an opponent's graveyard", effects.Creature(), effects.Not(effects.YouOwn()))),
		},
	})
}

func puppeteerCliqueReanimate(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	legal := ctx.LegalTargets()
	if len(legal) == 0 {
		return nil
	}
	id := legal[0].ID
	if err := (effects.ReturnFromGraveyard{Target: id, Dest: game.ZoneBattlefield, Controller: item.Controller}).Apply(ctx); err != nil {
		return err
	}
	if !effects.OnBattlefield(g, id) {
		return nil
	}
	if err := (effects.ScopedEffectFor{
		Target:   id,
		Mods:     []game.Mod{game.AddKeywordsMod("haste")},
		Duration: game.IndefiniteDuration(),
		Label:    "Puppeteer Clique — it gains haste",
	}).Apply(ctx); err != nil {
		return err
	}
	return effects.ScheduleDelayedTrigger{
		Label:              "Puppeteer Clique — exile it",
		Cards:              []uuid.UUID{id},
		Body:               effects.ExileListedCardsBody,
		ControllerTurnOnly: true,
	}.Apply(ctx)
}

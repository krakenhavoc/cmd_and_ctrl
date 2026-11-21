package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Syrix, Carrier of the Flame — Legendary Creature — Phoenix {2}{B}{R},
// 3/3 (Streets of New Capenna):
//
//	"Flying, haste
//	 At the beginning of each end step, if a creature card left your
//	 graveyard this turn, target Phoenix you control deals damage equal to
//	 its power to any target.
//	 Whenever another Phoenix you control dies, you may cast this card from
//	 your graveyard."
//
// The end-step trigger fires at every player's end step and is an
// intervening "if" (CR 603.4): it must already be true as the step begins
// and is checked again as the ability resolves. "A creature card left your
// graveyard" reads this turn's zone moves (a cast from the graveyard counts,
// cardThatLeftAGraveyard) and looks at the card as it now is. Two target
// clauses: the Phoenix you control (slot 0), then any target (slot 1); the
// Phoenix deals the damage, with its power as it resolves, and nothing
// happens if it is no longer a legal target (CR 608.2b).
//
// The graveyard ability watches from the graveyard (Triggered.Zones), where
// Syrix is when it matters. Another Phoenix dying (looked at as it last
// existed, CR 603.10a) lets you cast Syrix as the ability resolves, ignoring
// timing (CR 608.2g), for its mana cost; decline and it stays in the
// graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "ec97bb0e-0dc8-4b9b-8762-2f1649cc9044",
		Name:            "Syrix, Carrier of the Flame",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.InterveningIf(syrixCreatureLeftGraveyardCond,
					effects.On(game.EventBeginEndStep, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return creatureCardLeftYourGraveyardThisTurn(g, source.Controller)
					}, "Syrix, Carrier of the Flame — target Phoenix you control deals damage equal to its power to any target",
						syrixPhoenixDamage)),
				effects.TargetCreature("target Phoenix you control", effects.YouControl(), effects.OfCreatureType("Phoenix")).Then(effects.TargetAny())),
			{
				Watches: []game.EventKind{game.EventLTB},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					if ev.CardID == source.InstanceID {
						return false
					}
					dead, ok := effects.DiedCreature(ev, g)
					return ok && effects.LeftUnderControlOf(ev, dead) == source.Controller && effects.LeftAsSubtype(ev, dead, source.TextSubtype("Phoenix"))
				},
				Key:    "Syrix, Carrier of the Flame — you may cast this card from your graveyard",
				Zones:  []game.ZoneKind{game.ZoneGraveyard},
				Effect: syrixOfferCast,
			},
		},
	})
}

// creatureCardLeftYourGraveyardThisTurn is cardLeftYourGraveyardThisTurn for
// a creature card.
func creatureCardLeftYourGraveyardThisTurn(g *game.Game, player uuid.UUID) bool {
	for _, ev := range g.EventsThisTurn() {
		if c, ok := effects.CardThatLeftAGraveyard(ev, g); ok && c.Owner == player && c.IsCreature() {
			return true
		}
	}
	return false
}

var syrixCreatureLeftGraveyardCond = effects.CondBoard{
	Name: "a creature card left your graveyard this turn",
	Test: func(ctx *effects.Context) bool {
		return creatureCardLeftYourGraveyardThisTurn(ctx.Game, ctx.Controller())
	},
}

func syrixPhoenixDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	phoenix, ok := ctx.ClauseTarget(0)
	if !ok || phoenix.Kind != game.TargetCard {
		return nil
	}
	victim, ok := ctx.ClauseTarget(1)
	if !ok {
		return nil
	}
	power := effects.B31CurrentPowerOf(g, phoenix.ID)
	return effects.DealDamage{Source: phoenix.ID, Target: victim.ID, Amount: power}.Apply(ctx)
}

// syrixOfferCast offers the cast of Syrix from its owner's graveyard while
// the ability resolves: a flash-timed permission at the card's mana cost,
// offered through the resolution-cast prompt, whose decline leaves the card
// where it is (OfferGraveyardCast without Emet-Selch's exile rider).
func syrixOfferCast(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	id := ctx.Source()
	z := g.FindCardZoneForEffect(id)
	c, ok := g.LookupCardForEffect(id)
	if !ok || z == nil || z.Kind != game.ZoneGraveyard || c.Owner != ctx.Controller() {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      ctx.Controller(),
		Zone:        game.ZoneGraveyard,
		Duration:    g.UntilEndOfTurnDuration(),
		AltCostKey:  game.AltCostKeyGrantedCost,
		Timing:      game.TimingFlash,
		CastOnly:    true,
		LapseOnPass: game.LapseStaysInHand,
		Source:      id,
		SourceName:  "Syrix, Carrier of the Flame",
		Label:       "Syrix, Carrier of the Flame — cast it from your graveyard",
	}, []game.Card{c})
	effects.OfferResolutionCastOver(g, ctx.Controller(), id, "Syrix, Carrier of the Flame", game.MayCastKeywordPaidCast, effects.DeclineLeaveInGraveyard, id)
	return nil
}

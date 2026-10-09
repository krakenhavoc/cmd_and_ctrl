package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crabomination — Creature — Crab Demon {4}{B}{B}, 5/5:
//
//	"Emerge from artifact {5}{B}{B} (You may cast this spell by
//	 sacrificing an artifact and paying the emerge cost reduced by that
//	 artifact's mana value.)
//	 When this creature enters, target opponent exiles the top card of
//	 their library, a card at random from their graveyard, and a card at
//	 random from their hand. You may cast a spell from among cards
//	 exiled this way without paying its mana cost."
//
// "Emerge from artifact" is CR 702.119b's variant (EmergeFrom): the
// sacrificed permanent is any artifact, a Treasure (mana value 0) as
// readily as an artifact creature, and its mana value comes off the
// emerge cost's generic part (ADR 0135 §4).
//
// The enters trigger exiles the three cards as one simultaneous event,
// each random pick drawn from the game's keyed stream so an undo replays
// it, and an empty zone simply contributes nothing. Of the cards that
// reach exile (CR 400.7: a commander its owner moves on to the command
// zone is no longer one of them), the Crabomination's controller may cast
// one spell without paying its mana cost.
//
// The cast is a GRANT, the posture every resolution-time "you may cast"
// takes (cascade, discover, Emergent Ultimatum): one permission over the
// exiled cards, priced {0} (so an {X} is 0, CR 107.3b), spells only (a
// land among them can't be played), at instant speed (CR 608.2g), for one
// spell (CastsLeft), closing on its holder's next priority pass with the
// cards left in exile.
//
// No simplification beyond that posture.
func init() {
	enter := WhenThisEnters("Crabomination — target opponent exiles three cards; you may cast a spell from among them", crabominationExile)
	enter.Targets = TargetPlayer("target opponent", Opponent())
	Register(Spec{
		OracleID:         "5c26c30b-bc4a-4d20-8406-ef9022256e0d",
		Name:             "Crabomination",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{EmergeFrom("artifact", "{5}{B}{B}", Artifact())},
		Triggered:        []game.TriggeredAbility{enter},
	})
}

// crabominationName labels the permission and the log.
const crabominationName = "Crabomination"

// crabominationExile exiles the target opponent's top card, a random
// graveyard card and a random hand card, then grants the free cast over
// whichever of them reached exile.
func crabominationExile(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var opp uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			opp = t.ID
			break
		}
	}
	p := ctx.PlayerByID(opp)
	if p == nil || p.Eliminated {
		return nil
	}
	var ids []uuid.UUID
	if p.Library != nil {
		if top, err := p.Library.Top(); err == nil {
			ids = append(ids, top.InstanceID)
		}
	}
	ids = append(ids, g.ChooseAtRandomForEffect(randomDraw(ctx), graveyardIDs(g, opp, nil), 1)...)
	if p.Hand != nil {
		hand := make([]uuid.UUID, 0, len(p.Hand.Cards))
		for _, c := range p.Hand.Cards {
			hand = append(hand, c.InstanceID)
		}
		ids = append(ids, g.ChooseAtRandomForEffect(randomDraw(ctx), hand, 1)...)
	}
	if len(ids) == 0 {
		return nil
	}
	controller, source := item.Controller, item.SourceCardID
	return g.ExileCardsThenForEffect(ids, func(g *game.Game, exiled []uuid.UUID) error {
		cards := make([]game.Card, 0, len(exiled))
		for _, id := range exiled {
			if c, ok := g.LookupCardForEffect(id); ok {
				cards = append(cards, c)
			}
		}
		if len(cards) == 0 {
			return nil
		}
		g.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:      controller,
			Zone:        game.ZoneExile,
			Duration:    g.UntilEndOfTurnDuration(),
			Cost:        "{0}",
			Timing:      game.TimingFlash,
			CastOnly:    true,
			CastsLeft:   1,
			LapseOnPass: game.LapseStaysInExile,
			Source:      source,
			SourceName:  crabominationName,
			Label:       crabominationName + " — cast a spell from among the exiled cards without paying its mana cost",
		}, cards)
		return nil
	})
}

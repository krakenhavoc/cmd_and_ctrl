package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const kyloxsVoltstriderName = "Kylox's Voltstrider"

// Kylox's Voltstrider — Artifact — Vehicle {1}{U}{R} (Murders at Karlov
// Manor, #4469):
//
//	"Collect evidence 6: This Vehicle becomes an artifact creature until end
//	 of turn.
//	 Whenever this Vehicle attacks, you may cast an instant or sorcery spell
//	 from among cards exiled with it. If that spell would be put into a
//	 graveyard, put it on the bottom of its owner's library instead.
//	 Crew 2"
//
// Collect evidence 6 is the activated cost (CollectEvidenceN, #5193) with
// ExileCost.LinkToSource, so each card it exiles is "exiled with" this
// Vehicle (CR 607.2a: Card.ExiledWith names the object). The attack
// trigger reads those cards as it resolves, offers a cast of ONE instant or
// sorcery among them during the resolution (CR 608.2g; a permission over
// several cards is spent by one cast), paying its mana cost as usual, and the
// spell goes to the bottom of its owner's library rather than a graveyard
// (a cast follow-up, SetSpellBottomInsteadOfGraveyardForEffect). The link is to the object: a Vehicle that left
// and came back has none.
//
// No simplification.
func init() {
	evidence := CollectEvidenceN(6)
	evidence.ExileCards.LinkToSource = true
	Register(Spec{
		OracleID:     "e7b746c8-1b32-42ed-8328-4e16274209d8",
		Name:         kyloxsVoltstriderName,
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "Collect evidence 6: This Vehicle becomes an artifact creature until end of turn.",
				Cost:    evidence,
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return BecomeCreatureUntilEOT{Label: kyloxsVoltstriderName + " — becomes an artifact creature"}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "Crew 2",
				Cost:    CrewCost(2),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect(kyloxsVoltstriderName),
			},
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Kylox's Voltstrider — you may cast an instant or sorcery spell from among cards exiled with it",
				kyloxsVoltstriderCast),
		},
	})
}

// kyloxsVoltstriderBottom is "if that spell would be put into a graveyard, put
// it on the bottom of its owner's library instead", run when the spell is cast
// through the permission (game/cast_bottom_follow_up.go): the mark is on that
// spell and not on the card's other casts. An on-disk key.
var kyloxsVoltstriderBottom = game.RegisterCastFollowUp("kyloxs-voltstrider-bottom-instead-of-graveyard",
	func(g *game.Game, f game.CastFollowUp) error {
		g.SetSpellBottomInsteadOfGraveyardForEffect(f.Spell)
		return nil
	})

// kyloxsVoltstriderCast is the attack trigger's resolution: one permission
// over the instant and sorcery cards linked to this Vehicle, offered now.
func kyloxsVoltstriderCast(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	var cards []game.Card
	for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(ref)) {
		if !c.IsInstant() && !c.IsSorcery() {
			continue
		}
		cards = append(cards, *c)
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      ctx.Controller(),
		Zone:        game.ZoneExile,
		Timing:      game.TimingFlash,
		CastOnly:    true,
		Duration:    g.UntilEndOfTurnDuration(),
		LapseOnPass: game.LapseStaysInExile,
		FollowUp:    kyloxsVoltstriderBottom,
		Source:      ctx.Source(),
		SourceName:  kyloxsVoltstriderName,
		Label:       kyloxsVoltstriderName + " — cast an instant or sorcery spell from among cards exiled with it",
	}, cards)
	offerIDs := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		offerIDs = append(offerIDs, c.InstanceID)
	}
	offerResolutionCastOver(g, ctx.Controller(), ctx.Source(), kyloxsVoltstriderName,
		game.MayCastKeywordPaidCast, declineLeaveInExile, offerIDs...)
	return nil
}

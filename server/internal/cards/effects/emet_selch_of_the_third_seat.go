package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Emet-Selch of the Third Seat — Legendary Creature — Elder Wizard
// {2}{U}{B}, 3/4 (#3785, the Y'shtola deck request #3657):
//
//	"Spells you cast from your graveyard cost {2} less to cast.
//	 Whenever one or more opponents lose life, you may cast target
//	 instant or sorcery card from your graveyard. If that spell would
//	 be put into your graveyard, exile it instead. Do this only once
//	 each turn."
//
// Three printed clauses, three existing shapes, and one small seam:
//
//   - The discount is a cost modifier keyed on the zone the cast comes
//     FROM (CastFromGraveyard, a new constructor over
//     CostQuery.FromZone). It prices every graveyard cast the controller
//     makes, Emet's own included, flashback and escape casts too.
//   - "One or more opponents lose life" is a batch trigger (CR 603.2c):
//     OncePerBatch over the two events a loss rides (EventChangeLife for
//     a drain or a paid cost, EventDealDamage for damage), read through
//     b04OpponentLostLife, so two creatures connecting in one combat
//     damage step, or three opponents losing life to one Gray Merchant,
//     are ONE trigger. Two separate resolutions are two batches.
//   - "You may cast target ... from your graveyard" is
//     OfferGraveyardCast: a flash-timed, pass-closed cast permission
//     over the targeted card (CR 608.2g), paid at the card's MANA cost
//     ("cast", not "without paying"), whose ExileOnResolution is the
//     "exile it instead" replacement. The target is chosen as the trigger
//     goes on the stack and re-checked as it resolves (CR 608.2b).
//
// "Do this only once each turn" limits the DOING, as Krile Baldesion's
// and G'raha Tia's do: a cast the controller declines, or an offer that
// lapses because they pass priority, leaves the turn's one use for a
// later batch. The use is written by the offer's cast follow-up
// (emetSelchCastTally, game.CastPermission.FollowUp), which runs only
// once the spell is on the stack through THIS permission, so the grant
// cannot write it early and a declined offer cannot spend it. It is read
// when the ability triggers (no trigger goes on the stack once the use is
// spent) and again as it resolves (the second of two triggers stacked
// before either resolved does nothing). The tally is per Emet-Selch
// object (Game.NoteDoneThisTurn), so a second Emet-Selch has a use of its
// own, which is what "this ability" means for each.
//
// "If that spell would be put into your graveyard, exile it instead" is
// the offer's ExileInsteadOfGraveyard (#3863): the spell is exiled if it
// resolves, is countered or fizzles, and an instant or sorcery bounced
// to a hand or shuffled into a library goes there, as printed.
//
// No simplification.
const (
	emetSelchOracleID = "634475d9-a1d7-4146-9a37-645d3d162af1"
	emetSelchLabel    = "Emet-Selch of the Third Seat — you may cast target instant or sorcery card from your graveyard"
	// emetSelchDoneLabel is the tally the once-each-turn limit reads.
	emetSelchDoneLabel = "Emet-Selch of the Third Seat — graveyard cast (done)"
)

// emetSelchCastTally is the offer's "do this" having been done: the
// cast was made, so the turn's use is spent.
var emetSelchCastTally = game.RegisterCastFollowUp("emet-selch-of-the-third-seat-once-each-turn",
	func(g *game.Game, f game.CastFollowUp) error {
		g.NoteDoneThisTurn(f.Source, emetSelchDoneLabel)
		return nil
	})

func init() {
	Register(Spec{
		OracleID:     emetSelchOracleID,
		Name:         "Emet-Selch of the Third Seat",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(2, "Spells you cast from your graveyard cost {2} less to cast.",
				YourSpell(), CastFromGraveyard()),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				OncePerBatch(OnAny([]game.EventKind{game.EventChangeLife, game.EventDealDamage},
					emetSelchOpponentLostLife, emetSelchLabel, emetSelchOffer)),
				TargetCardInGraveyard("target instant or sorcery card in your graveyard",
					YouOwn(), Or(Instant(), Sorcery()))),
		},
	})
}

// emetSelchOpponentLostLife: an opponent lost life, and the turn's one
// use is not spent.
func emetSelchOpponentLostLife(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	if g.ResolvedThisTurn(source.InstanceID, emetSelchDoneLabel) > 0 {
		return false
	}
	_, ok := b04OpponentLostLife(ev, source.Controller, g)
	return ok
}

// emetSelchOffer opens the cast on the still-legal target, unless the
// turn's use was spent while the trigger waited.
func emetSelchOffer(g *game.Game, item *game.StackItem) error {
	if g.ResolvedThisTurn(item.SourceCardID, emetSelchDoneLabel) > 0 {
		return nil
	}
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return OfferGraveyardCast{
			Target:     t.ID,
			SourceName: "Emet-Selch of the Third Seat",
			FollowUp:   emetSelchCastTally,
		}.Apply(ctx)
	}
	return nil
}

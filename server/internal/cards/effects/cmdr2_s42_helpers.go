package effects

import (
	"strconv"

	"github.com/google/uuid"

	"fmt"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2S42AnyColorRefundBody is Plasm Capture's delayed refund: add
// <Amount> mana in any combination of colors. The amount rides as the
// trigger's params (ADR 0041 phase 3), bounded like Mana Drain's.
var cmdr2S42AnyColorRefundBody = game.DelayedBody("plasm-capture/refund",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if p.Amount <= 0 {
			return nil
		}
		if p.Amount > maxManaDrainRefund {
			return fmt.Errorf("plasm-capture/refund: implausible amount %d", p.Amount)
		}
		return AddMana{Amount: ManaN(p.Amount, game.AllColors...)}.Apply(NewContext(g, item))
	})

// cmdr2S42FreeGraveyardCast is the free cast out of a graveyard at the heart
// of this batch: the Arcane Heist grant (a {0}, flash-timed, cast-only,
// pass-closed permission, CR 608.2g) stamped over `ids` as ONE permission, so
// the first cast spends it ("cast a spell"). `spell` is the clause the spell
// must meet (judged on the face being cast), `exile` is "if that spell would
// be put into a graveyard, exile it instead". The offer is the
// resolution_cast prompt (#4250); an uncast card stays where it is.
func cmdr2S42FreeGraveyardCast(ctx *Context, name string, spell freeCastSpell, exile bool, ids []uuid.UUID) {
	cmdr2S42FreeGraveyardCasts(ctx, name, spell, exile, true, ids)
}

// cmdr2S42FreeGraveyardCasts is cmdr2S42FreeGraveyardCast with the choice of
// one shared permission (`one`: "cast a spell", the first cast spends it) or
// one permission per card ("cast each", "for each opponent, cast ..."), so a
// cast of one leaves the others open.
func cmdr2S42FreeGraveyardCasts(ctx *Context, name string, spell freeCastSpell, exile, one bool, ids []uuid.UUID) {
	g := ctx.Game
	perm := spell.on(game.CastPermission{
		Player:                  ctx.Controller(),
		Zone:                    game.ZoneGraveyard,
		Duration:                g.UntilEndOfTurnDuration(),
		Cost:                    "{0}",
		AltCostKey:              game.AltCostKeyGrantedCost,
		Timing:                  game.TimingFlash,
		CastOnly:                true,
		ExileInsteadOfGraveyard: exile,
		LapseOnPass:             game.LapseStaysInHand,
		Source:                  ctx.Source(),
		SourceName:              name,
		Label:                   name + " — cast it without paying its mana cost",
	})
	var cards []uuid.UUID
	for _, id := range ids {
		if z := g.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneGraveyard {
			if c, ok := g.LookupCardForEffect(id); ok && perm.OffersASpell(c) {
				cards = append(cards, id)
			}
		}
	}
	if len(cards) == 0 {
		return
	}
	offer := game.ResolutionCastOffer{
		Chooser:      ctx.Controller(),
		Source:       ctx.Source(),
		Reason:       name,
		Cards:        cards,
		Permission:   &perm,
		Keyword:      game.MayCastKeywordFreeCast,
		AcceptLabel:  "Cast it free",
		DeclineLabel: declineLeaveInGraveyard,
	}
	if !one {
		offer.Permission = nil
		for _, id := range cards {
			c, _ := g.LookupCardForEffect(id)
			g.GrantCastPermissionToCardsForEffect(perm, []game.Card{c})
		}
	}
	g.OfferResolutionCastForEffect(offer)
}

// cmdr2S42HaloForagerCastBody is Halo Forager's reflexive trigger: "you may
// cast target instant or sorcery card with mana value X from a graveyard
// without paying its mana cost. If that spell would be put into a graveyard,
// exile it instead." X rides Params.Amount, fixed when the trigger was made
// (CR 603.12); the clause is rebuilt from it on a restore.
var cmdr2S42HaloForagerCastBody = game.ReflexiveBody("halo-forager/cast",
	func(g *game.Game, item *game.StackItem, _ game.EffectParams) error {
		ctx := NewContext(g, item)
		var ids []uuid.UUID
		for _, t := range ctx.LegalTargets() {
			if t.Kind == game.TargetCard {
				ids = append(ids, t.ID)
			}
		}
		cmdr2S42FreeGraveyardCast(ctx, "Halo Forager", freeInstantOrSorcery(), true, ids)
		return nil
	},
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetCardInGraveyard("target instant or sorcery card with mana value "+strconv.Itoa(p.Amount)+" from a graveyard",
			Or(Instant(), Sorcery()), ManaValueLE(p.Amount), ManaValueGE(p.Amount))
	})

// cmdr2S42AdversaryBody is Bloodthirsty Adversary's reflexive trigger. The
// number of payments rides Params.Amount: it is both the counters and the
// cap of the "up to that many" targets.
var cmdr2S42AdversaryBody = game.ReflexiveBody("bloodthirsty-adversary/copy", cmdr2S42AdversaryResolve,
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetCardInGraveyard("up to "+strconv.Itoa(p.Amount)+" target instant and/or sorcery cards with mana value 3 or less from your graveyard",
			Or(Instant(), Sorcery()), YouOwn(), ManaValueLE(3)).WithCount(0, p.Amount)
	})

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// One Last Job — Sorcery {2}{W} (Outlaws of Thunder Junction):
//
//	"Spree (Choose one or more additional costs.)
//	 + {2} — Return target creature card from your graveyard to the
//	   battlefield.
//	 + {1} — Return target Mount or Vehicle card from your graveyard to the
//	   battlefield.
//	 + {1} — Return target Aura or Equipment card from your graveyard to the
//	   battlefield attached to a creature you control."
//
// The bullets run in printed order, so a creature the first one returns can be
// the host the third picks. The third bullet's creature is NOT a target: it is
// chosen as the spell resolves (a keyed pick, so a table waiting on it is a
// restore point), among the creatures you control that the card could be
// attached to (canBeAttachedTo, CR 303.4j). With nothing to attach to an Aura
// stays in the graveyard, and an Equipment returns unattached (the card's
// ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55cba59c-932c-4295-87dc-9179fbdfbd53",
		Name:         "One Last Job",
		Completeness: CompletenessFull,
		Modes: Spree(
			SpreeModeDoing("Return target creature card from your graveyard to the battlefield.", "{2}",
				TargetCardInGraveyard("target creature card in your graveyard", YouOwn(), Creature()),
				oneLastJobReturn),
			SpreeModeDoing("Return target Mount or Vehicle card from your graveyard to the battlefield.", "{1}",
				TargetCardInGraveyard("target Mount or Vehicle card in your graveyard", YouOwn(), Or(OfSubtype("Mount"), OfSubtype("Vehicle"))),
				oneLastJobReturn),
			SpreeModeDoing("Return target Aura or Equipment card from your graveyard to the battlefield attached to a creature you control.", "{1}",
				TargetCardInGraveyard("target Aura or Equipment card in your graveyard", YouOwn(), Or(OfSubtype("Aura"), OfSubtype("Equipment"))),
				oneLastJobAttached),
		),
	})
}

func oneLastJobReturn(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	return ReturnFromGraveyardTogether{Targets: []uuid.UUID{t.ID}, Controller: ctx.Controller()}.Apply(ctx)
}

func oneLastJobAttached(_ *game.StackItem, ctx *Context, occ int) error {
	t, ok := ModeTarget(ctx, occ)
	if !ok {
		return nil
	}
	g := ctx.Game
	if z := g.FindCardZoneForEffect(t.ID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	card, ok := g.LookupCardForEffect(t.ID)
	if !ok {
		return nil
	}
	var hosts []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == ctx.Controller() && c.IsCreature() && canBeAttachedTo(g, ctx.Controller(), card, c.InstanceID) {
			hosts = append(hosts, c.InstanceID)
		}
	}
	if len(hosts) == 0 {
		if card.HasSubtype("Equipment") {
			return ReturnFromGraveyardTogether{Targets: []uuid.UUID{t.ID}, Controller: ctx.Controller()}.Apply(ctx)
		}
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "One Last Job — choose a creature you control to attach " + card.Name + " to",
		Cards:    hosts,
		Min:      1,
		Max:      1,
		ThenKey:  oneLastJobHost,
		LiveZone: game.ZoneBattlefield,
		Carry:    []uuid.UUID{t.ID},
	})
	return nil
}

// oneLastJobHost is the creature chosen: the Aura or Equipment returns
// attached to it. The card must still be in the graveyard, and the creature
// still a legal host (CR 608.2b). An on-disk key: never renamed, never reused.
var oneLastJobHost = game.RegisterChooseCardsThen("choose-cards/one-last-job-host", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) == 0 || len(r.Carry) == 0 {
		return nil
	}
	z := g.FindCardZoneForEffect(r.Carry[0])
	if z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	card, ok := g.LookupCardForEffect(r.Carry[0])
	if !ok || !canBeAttachedTo(g, r.Chooser, card, r.Picked[0]) {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect(
		[]game.BatchEntry{{CardID: r.Carry[0], From: game.ZoneGraveyard}},
		game.ZoneEntryOptions{Controller: r.Chooser, AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: r.Picked[0]}},
		game.KeyedTail{})
})

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mairsil, the Pretender — Legendary Creature — Human Wizard {1}{U}{B}{R},
// 4/4:
//
//	"When Mairsil enters, you may exile an artifact or creature card
//	 from your hand or graveyard and put a cage counter on it.
//	 Mairsil has all activated abilities of all cards you own in exile
//	 with cage counters on them. You may activate each of those
//	 abilities only once each turn."
//
// The enters trigger is a choice made as it resolves, not a target: its
// controller picks at most one artifact or creature card from their hand
// and graveyard (picking none is the "may"), which is exiled and gets a
// cage counter once it has landed in exile.
//
// The borrow is not an "exiled with" link: every card you own in exile
// with a cage counter lends, whoever caged it (ExileLenders), so an
// earlier Mairsil's cages still count for a new one. "Only once each
// turn" is the borrow's EachOncePerTurn tally per (Mairsil object, lent
// row); two identical lent rows are one ability, as the ruling says. A
// card that arrives in exile from a hand re-reads the layers
// (DependsOnExile).
//
// No simplification.
func init() {
	const clause = "Mairsil has all activated abilities of all cards you own in exile with cage counters on them. You may activate each of those abilities only once each turn."
	borrow := HasActivatedAbilitiesOf(game.AbilityBorrow{
		Label: clause,
		Lenders: ExileLenders(func(g *game.Game, source *game.Card, c game.Card) bool {
			return c.Owner == source.Controller && LenderWithCounter(mairsilCageCounter)(g, source, c)
		}),
		EachOncePerTurn: true,
	})
	borrow.DependsOnExile = true
	Register(Spec{
		OracleID:     "ed5fbae0-2dca-44a5-84dc-62674b3a92a6",
		Name:         "Mairsil, the Pretender",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{borrow},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Mairsil — you may exile an artifact or creature card from your hand or graveyard with a cage counter", mairsilCage),
		},
	})
}

const mairsilCageCounter = "cage"

// mairsilCage is the enters trigger: offer the controller's artifact and
// creature cards in hand and graveyard, at most one, and cage the pick.
func mairsilCage(g *game.Game, item *game.StackItem) error {
	you := item.Controller
	p := g.PlayerByIDForEffect(you)
	if p == nil {
		return nil
	}
	var pool []uuid.UUID
	for _, z := range []*game.Zone{p.Hand, p.Graveyard} {
		if z == nil {
			continue
		}
		for _, c := range z.Cards {
			if c.IsArtifact() || c.IsCreature() {
				pool = append(pool, c.InstanceID)
			}
		}
	}
	if len(pool) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  you,
		Source:   item.SourceCardID,
		Question: "Mairsil — exile an artifact or creature card from your hand or graveyard with a cage counter?",
		Cards:    pool,
		Min:      0,
		Max:      1,
		ThenKey:  mairsilCagePick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// mairsilCagePick is the card chosen to cage, keyed (#4253): the asking item
// rides the prompt as data.
var mairsilCagePick = askedPick("choose-cards/mairsil-cage", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	// Still where it was offered from: a card that moved since
	// is a new object the trigger does not know (CR 400.7).
	z := ctx.Game.FindCardZoneForEffect(r.Picked[0])
	if z == nil || z.Owner != ctx.Controller() || (z.Kind != game.ZoneHand && z.Kind != game.ZoneGraveyard) {
		return nil
	}
	return exileCardWithCounter(ctx, r.Picked[0], mairsilCageCounter)
})

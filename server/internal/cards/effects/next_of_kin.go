package effects

import (
	"slices"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Next of Kin — Enchantment — Aura {2}{G} (Streets of New Capenna):
//
//	"Enchant creature
//	 When enchanted creature dies, you may put a creature card you own
//	 with lesser mana value from your hand or from the command zone onto
//	 the battlefield. If you do, return this card to the battlefield
//	 attached to that creature at the beginning of the next end step."
//
// The trigger reads the dead creature's mana value as it last existed
// (CR 603.10a, PermanentInfo.ManaValue) and carries it to the pick, which
// offers the creature cards you OWN in your hand and your command zone
// with a lesser mana value (a commander in the command zone is such a
// card). The pick is re-checked as it is answered: a card that left its
// zone, or is no longer cheap enough, is not put onto the battlefield. A
// card put onto the battlefield is not cast (no commander tax). "If you
// do" is the card having entered: only then is Gift of Immortality's
// delayed trigger scheduled (giftOfImmortalityReturnAuraBody): at the
// beginning of the next end step the Aura returns from the graveyard
// attached to that creature, while it is still the same object on the
// battlefield.
//
// "You may" is the pick's floor of zero.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b079914-1302-40b5-8060-7c3e5b6e9eae",
		Name:         "Next of Kin",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Triggered: []game.TriggeredAbility{
			WhenEnchantedCreatureDies("Next of Kin — you may put a creature card you own with lesser mana value from your hand or from the command zone onto the battlefield",
				nextOfKinAsk),
		},
	})
}

// nextOfKinCandidates are the creature cards `player` owns, in their hand
// and their command zone, with a mana value below `limit`.
func nextOfKinCandidates(g *game.Game, player uuid.UUID, limit int) []uuid.UUID {
	p := g.PlayerByIDForEffect(player)
	if p == nil {
		return nil
	}
	var ids []uuid.UUID
	for _, zone := range []*game.Zone{p.Hand, p.Command} {
		if zone == nil {
			continue
		}
		for _, c := range zone.Cards {
			if c.Owner != player || !c.IsCreature() || c.IsToken() {
				continue
			}
			if mv, ok := g.ManaValueForEffect(c); ok && mv < limit {
				ids = append(ids, c.InstanceID)
			}
		}
	}
	return ids
}

func nextOfKinAsk(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	info, ok := ctx.TriggeringPermanent()
	if !ok {
		return nil
	}
	candidates := nextOfKinCandidates(g, ctx.Controller(), info.ManaValue)
	if len(candidates) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:   ctx.Controller(),
		Source:    ctx.Source(),
		Question:  "Next of Kin — put a creature card you own with lesser mana value onto the battlefield",
		Cards:     candidates,
		Min:       0,
		Max:       1,
		ThenKey:   nextOfKinPicked,
		CarryNums: []int{info.ManaValue},
		Asked:     game.NewAskedItem(ctx.Item),
	})
	return nil
}

// nextOfKinPicked puts the chosen card onto the battlefield and schedules
// the Aura's return. An on-disk key.
var nextOfKinPicked = askedPickRule("choose-cards/next-of-kin-creature", nil,
	func(ctx *Context, r game.CardsChosen) error {
		if len(r.Picked) == 0 || len(r.Nums) == 0 {
			return nil
		}
		id := r.Picked[0]
		if !slices.Contains(nextOfKinCandidates(ctx.Game, ctx.Controller(), r.Nums[0]), id) {
			return nil
		}
		var entered uuid.UUID
		var err error
		if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneCommand {
			entered, err = ctx.Game.PutFromCommandZoneOntoBattlefieldForEffect(id, game.ZoneEntryOptions{Controller: ctx.Controller()})
		} else {
			entered, err = ctx.Game.PutFromHandOntoBattlefieldForEffect(id, game.HandEntryOptions{Controller: ctx.Controller()})
		}
		if err != nil {
			return err
		}
		if entered == uuid.Nil {
			entered = id
		}
		c, ok := ctx.Game.LookupCardForEffect(entered)
		if !ok || !onBattlefield(ctx.Game, entered) {
			return nil
		}
		return ScheduleDelayedTrigger{
			Label:  "Next of Kin — return this card to the battlefield attached to that creature",
			Cards:  []uuid.UUID{ctx.Source()},
			Body:   giftOfImmortalityReturnAuraBody,
			Params: game.EffectParams{Object: game.ObjectRef{ID: entered, Epoch: c.ObjectEpoch}},
		}.Apply(ctx)
	})

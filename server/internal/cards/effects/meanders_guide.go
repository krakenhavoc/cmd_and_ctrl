package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Meanders Guide — Creature — Merfolk Scout {2}{W}, 3/2 (Lorwyn Eclipsed,
// #4458):
//
//	"Whenever this creature attacks, you may tap another untapped Merfolk
//	 you control. When you do, return target creature card with mana value
//	 3 or less from your graveyard to the battlefield."
//
// Myr Battlesphere's offer (a pick of at most one untapped Merfolk,
// re-checked as it is tapped, since the board moves under a prompt) and a
// CR 603.12 reflexive trigger: it exists only once a Merfolk was really
// tapped, and its target is chosen as it goes on the stack, so a creature
// card that reached the graveyard in response is a legal pick. "Merfolk" is
// a word of the card's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8229954b-e1ef-4503-8443-2356e2a87663",
		Name:         "Meanders Guide",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Meanders Guide — you may tap another untapped Merfolk you control; when you do, return target creature card with mana value 3 or less from your graveyard to the battlefield",
				meandersGuideAttack),
		},
	})
}

func meandersGuideAttack(g *game.Game, item *game.StackItem) error {
	word := g.TextReaderForEffect(item).TextSubtype("Merfolk")
	var merfolk []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.InstanceID != item.SourceCardID && c.IsCreature() && !c.Tapped && c.HasSubtype(word) {
			merfolk = append(merfolk, c.InstanceID)
		}
	}
	if len(merfolk) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Meanders Guide — you may tap another untapped Merfolk you control",
		Cards:    merfolk,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  meandersGuidePick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// meandersGuidePick taps the chosen Merfolk if it is still an untapped
// Merfolk, and if it was tapped, creates the reflexive trigger. Keyed.
var meandersGuidePick = askedPick("choose-cards/meanders-guide-tap", func(ctx *Context, r game.CardsChosen) error {
	word := ctx.TextSubtype("Merfolk")
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != ctx.Controller() ||
			c.InstanceID == ctx.Item.SourceCardID || !c.HasSubtype(word) {
			continue
		}
		if err := (TapTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
		return ReflexiveTrigger{
			Label: "Meanders Guide — return target creature card with mana value 3 or less from your graveyard to the battlefield",
			Body:  meandersGuideReturnBody,
		}.Apply(ctx)
	}
	return nil
})

// meandersGuideReturnBody is "when you do, return target creature card with
// mana value 3 or less from your graveyard to the battlefield".
var meandersGuideReturnBody = game.ReflexiveBody("meanders-guide/return", simpleBody(
	func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		ids := legalTargetCardIDs(ctx)
		if len(ids) == 0 {
			return nil
		}
		return ReturnFromGraveyard{Target: ids[0], Dest: game.ZoneBattlefield, Controller: ctx.Controller()}.Apply(ctx)
	}),
	constTargets(func() *game.TargetSpec {
		return TargetCardInGraveyard("target creature card with mana value 3 or less from your graveyard",
			Creature(), ManaValueLE(3), YouOwn())
	}))

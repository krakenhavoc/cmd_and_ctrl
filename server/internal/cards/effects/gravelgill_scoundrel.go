package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gravelgill Scoundrel — Creature — Merfolk Rogue {1}{U}, 1/3 (Lorwyn
// Eclipsed, #4458):
//
//	"Vigilance
//	 Whenever this creature attacks, you may tap another untapped creature
//	 you control. If you do, this creature can't be blocked this turn."
//
// Meanders Guide's offer (a pick of at most one untapped creature other than
// the Scoundrel, re-checked as it is tapped since the board moves under a
// prompt). Tapping it is "if you do": only a creature that was really
// tapped makes the Scoundrel unblockable, for the turn, and only while it is
// still the permanent that attacked. The tap is not a cost, so a creature
// that entered this turn may be tapped.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6d7f7e6e-5076-4af7-afe3-3e500835c4aa",
		Name:            "Gravelgill Scoundrel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Gravelgill Scoundrel — you may tap another untapped creature you control; if you do, this creature can't be blocked this turn",
				gravelgillScoundrelAttack),
		},
	})
}

func gravelgillScoundrelAttack(g *game.Game, item *game.StackItem) error {
	var others []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.InstanceID != item.SourceCardID && c.IsCreature() && !c.Tapped {
			others = append(others, c.InstanceID)
		}
	}
	if len(others) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Gravelgill Scoundrel — you may tap another untapped creature you control (it can't be blocked this turn)",
		Cards:    others,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  gravelgillScoundrelPick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// gravelgillScoundrelPick taps the chosen creature if it is still an
// untapped creature of its controller's other than the Scoundrel, and if it
// was tapped, makes the Scoundrel unblockable this turn. Keyed.
var gravelgillScoundrelPick = askedPick("choose-cards/gravelgill-scoundrel-tap", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != ctx.Controller() || c.InstanceID == ctx.Item.SourceCardID {
			continue
		}
		if err := (TapTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
		if !sourceIsStillThisPermanent(ctx.Game, ctx.Item) {
			return nil
		}
		return RestrictUntilEOT{Target: ctx.Item.SourceCardID, Restrictions: game.CantBeBlocked,
			Label: "Gravelgill Scoundrel — can't be blocked this turn"}.Apply(ctx)
	}
	return nil
})

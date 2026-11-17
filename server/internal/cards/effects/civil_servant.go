package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Civil Servant — Creature — Cat Citizen {G}{W}, 2/3 (Streets of New
// Capenna, #5491):
//
//	"Whenever this creature attacks, you may tap another untapped Citizen
//	 you control. If you do, this creature gets +1/+0 and gains lifelink
//	 until end of turn."
//
// Meanders Guide's offer (a pick of at most one untapped other Citizen,
// re-checked as it is tapped, since the board moves under a prompt), then
// the "if you do" pump applies only when a Citizen was really tapped
// (CR 608.2c). "Citizen" is a word of the card's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "232e16f4-32ae-4f42-a377-6bdebc9a06c7",
		Name:         "Civil Servant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Civil Servant — you may tap another untapped Citizen you control; if you do, this creature gets +1/+0 and gains lifelink until end of turn",
				civilServantAttack),
		},
	})
}

func civilServantAttack(g *game.Game, item *game.StackItem) error {
	word := g.TextReaderForEffect(item).TextSubtype("Citizen")
	var citizens []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.InstanceID != item.SourceCardID && c.IsCreature() && !c.Tapped && c.HasSubtype(word) {
			citizens = append(citizens, c.InstanceID)
		}
	}
	if len(citizens) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Civil Servant — you may tap another untapped Citizen you control",
		Cards:    citizens,
		Min:      0,
		Max:      1,
		LiveZone: game.ZoneBattlefield,
		ThenKey:  civilServantPick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// civilServantPick taps the chosen Citizen if it is still an untapped other
// Citizen, and if it was tapped, pumps the Servant. Keyed.
var civilServantPick = askedPick("choose-cards/civil-servant-tap", func(ctx *Context, r game.CardsChosen) error {
	word := ctx.TextSubtype("Citizen")
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != ctx.Controller() ||
			c.InstanceID == ctx.Item.SourceCardID || !c.HasSubtype(word) {
			continue
		}
		if err := (TapTarget{Target: id}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
		self := ctx.Item.SourceCardID
		if err := (BoostUntilEOT{Target: self, Power: 1, Label: "Civil Servant — +1/+0"}).Apply(ctx); err != nil {
			return err
		}
		return GrantKeywordUntilEOT{Target: self, Keywords: []string{"lifelink"}, Label: "Civil Servant — lifelink"}.Apply(ctx)
	}
	return nil
})

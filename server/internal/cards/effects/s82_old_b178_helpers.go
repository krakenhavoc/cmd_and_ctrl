package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// OldB178ConsumeGain is Consume's tail: the caster gains life equal to the power
// of the creature the target player sacrificed, as it last existed. The carry is
// the caster, then the victim. It is on disk: never renamed.
var OldB178ConsumeGain = game.RegisterPromptRunThen("prompt-run/old-b178-consume-gain-power", func(g *game.Game, r game.RunDone) error {
	if len(r.Carry) < 2 {
		return nil
	}
	me, victim := r.Carry[0], r.Carry[1]
	ids := r.Sacrifices().By(victim)
	if len(ids) == 0 {
		return nil
	}
	power := DepartedCreaturePower(g, ids[0])
	if power <= 0 {
		return nil
	}
	return GainLife{Player: me, Amount: power}.Apply(KeyedContext(g, me, r.Source))
})

// OldB178EdictGreatestPowerGainPower is "Target player sacrifices a creature with the
// greatest power among creatures they control. You gain life equal to its power."
// (Consume): the target player picks among their creatures of greatest power, and the
// life waits for the sacrifice.
func OldB178EdictGreatestPowerGainPower(ctx *Context, name string) error {
	victim, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	return ctx.Game.PlayerSacrificesThenKeyedForEffect(ctx.Source(), victim,
		SacrificeSpec("a creature with the greatest power among creatures you control", GreatestPowerYouControl()),
		name+" — sacrifice a creature with the greatest power", 1,
		OldB178ConsumeGain, []uuid.UUID{ctx.Controller(), victim})
}

// OldB178TargetExilesOwnGraveyardCard is "Target player exiles a card from their
// graveyard" (Scrabbling Claws): the TARGET player chooses, from their own
// graveyard, as the ability resolves. The prompt is keyed (on disk, never renamed).
func OldB178TargetExilesOwnGraveyardCard(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var player uuid.UUID
	for _, t := range ctx.LegalTargets() {
		if t.Kind == game.TargetPlayer {
			player = t.ID
			break
		}
	}
	if player == uuid.Nil {
		return nil
	}
	cards := GraveyardCardsOfPlayer(g, player)
	if len(cards) == 0 {
		return nil
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     ctx.Source(),
		Question:   "Scrabbling Claws — exile a card from your graveyard",
		Cards:      cards,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneGraveyard,
		ThenKey:    oldB178ClawsExiled,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

var oldB178ClawsExiled = AskedPick("choose-cards/old-b178-scrabbling-claws-exile", func(ctx *Context, r game.CardsChosen) error {
	for _, id := range r.Picked {
		if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
})

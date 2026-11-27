package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b622_helpers.go — shared pieces of Judgment batch 622 (#5611),
// threshold.

// OldB622ThresholdSelfDiesLoseLife is the "has 'When this creature dies, you
// lose N life'" half of Treacherous Vampire and Treacherous Werewolf: the
// grant bundle (one dies trigger) and the layer 6 static that gives it to the
// creature itself while there are seven or more cards in your graveyard. The
// grant is read as it stood on the battlefield (CR 603.10a): a creature that
// died at threshold has the trigger even though it is now in the graveyard.
func OldB622ThresholdSelfDiesLoseLife(name, key string, n int) (AbilityGrant, game.StaticAbility) {
	text := "When this creature dies, you lose " + NumberWord(n) + " life."
	grant := AbilityGrant{
		Key: key,
		Triggered: []game.TriggeredAbility{
			WhenThisDies(name+" — you lose "+NumberWord(n)+" life", Do(YouLoseLife(n))),
		},
		Text: text,
	}
	static := Labelled(GrantAbilities(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && FraThresholdMet(g, source.Controller)
	}, key), "Threshold — As long as there are seven or more cards in your graveyard, this creature has \""+text+"\"")
	return grant, static
}

// OldB622SacrificeUnlessYouExileFromYourGraveyard is Treacherous Vampire's
// "sacrifice it unless you exile a card from your graveyard": the controller
// is asked, picks the card, and a "no" (or a graveyard with nothing to exile,
// CR 118.3) sacrifices the creature. A source that left in response, or left
// and came back (CR 400.7), asks nothing.
func OldB622SacrificeUnlessYouExileFromYourGraveyard(name string) Effect {
	words, action := ExileFromGraveyardPayment(1).Prompt()
	return func(g *game.Game, item *game.StackItem) error {
		if !OnBattlefield(g, item.SourceCardID) || SourceIsNewObject(g, item) {
			return nil
		}
		a := *action
		return UpkeepPayUnless{
			Chooser:  item.Controller,
			Cost:     words,
			Action:   &a,
			Question: name + " — exile a card from your graveyard, or sacrifice " + name + "?",
			Else:     UpkeepSacrificeSource,
		}.Apply(NewContext(g, item))
	}
}

// OldB622OtherCreaturesYouControlAtThreshold is the scope of Silver Seraph's
// threshold anthem: the source's other creatures, while its controller has
// seven or more cards in their graveyard.
func OldB622OtherCreaturesYouControlAtThreshold(target *game.Card, g *game.Game, source *game.Card) bool {
	return OtherCreaturesYouControl(target, g, source) && FraThresholdMet(g, source.Controller)
}

// OldB622GreenOrWhiteCreature is Masked Gorgon's "green creatures and white
// creatures", whoever controls them.
func OldB622GreenOrWhiteCreature(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && (target.HasColor(source.TextColor("G")) || target.HasColor(source.TextColor("W")))
}

// OldB622AvenWarcraftThreshold is the second sentence of Aven Warcraft:
// "Threshold — If there are seven or more cards in your graveyard, choose a
// color. Creatures you control also gain protection from the chosen color
// until end of turn." The creatures are the ones you control as it resolves.
var OldB622AvenWarcraftThreshold = Instruction(func(ctx *Context) error {
	g := ctx.Game
	if !FraThresholdMet(g, ctx.Controller()) {
		return nil
	}
	var ids []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.Controller == ctx.Controller() {
			ids = append(ids, c.InstanceID)
		}
	}
	if len(ids) == 0 {
		return nil
	}
	then := ctx.ItemTail(RadianceProtectionFromChosenColor)
	then.Carry = ids
	then.Words = []string{"Aven Warcraft"}
	ChooseColorTail(game.ColorForProtection, g, ctx.Controller(), ctx.Source(), "Aven Warcraft — choose a color", then)
	return nil
})

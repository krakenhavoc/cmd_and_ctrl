package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b714_helpers.go — shared shapes of S82 OLD batch 714 (Portal Three
// Kingdoms, #5622): Marshaling the Troops, Wu Spy and Wei Assassins.

// MarshalingTheTroopsResolve is "Tap any number of untapped creatures you
// control. You gain 4 life for each creature tapped this way." The creatures
// are chosen as the spell resolves, any number, none included; each pick is
// read again before it is tapped, so a creature that left, was tapped or was
// stolen in the meantime is not "tapped this way" and gains no life.
func MarshalingTheTroopsResolve(item *game.StackItem, ctx *Context) error {
	var candidates []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == item.Controller && c.IsCreature() && !c.Tapped {
			candidates = append(candidates, c.InstanceID)
		}
	}
	if len(candidates) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  item.Controller,
		Source:   item.SourceCardID,
		Question: "Marshaling the Troops — tap any number of untapped creatures you control",
		Cards:    candidates,
		Min:      0,
		Max:      0, // "any number"
		LiveZone: game.ZoneBattlefield,
		ThenKey:  marshalingTheTroopsPick,
		Asked:    game.NewAskedItem(item),
	})
	return nil
}

// marshalingTheTroopsPick taps the picks that are still untapped creatures
// the spell's controller controls and gains 4 life for each one tapped. An
// on-disk key: never renamed, never reused.
var marshalingTheTroopsPick = AskedPick("choose-cards/s82-old-b714-marshaling-the-troops-tap", func(ctx *Context, r game.CardsChosen) error {
	tapped := 0
	for _, id := range r.Picked {
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || c.Tapped || !c.IsCreature() || c.Controller != ctx.Item.Controller {
			continue
		}
		if err := (TapTarget{Target: id}).Apply(ctx.AsGroupMember()); err != nil {
			return err
		}
		tapped++
	}
	if tapped == 0 {
		return nil
	}
	return GainLife{Amount: 4 * tapped}.Apply(ctx)
})

// S82OldB714WuSpyLook is Wu Spy's trigger: "Look at the top two cards of
// target player's library. Put one of them into their graveyard." Only the
// controller looks (CR 701.20) and must pick exactly one (a library of one
// card gives no choice); the other stays where it was, on top.
func S82OldB714WuSpyLook(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	owner, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	looked := g.LookAtTopOfPlayersLibraryForEffect(ctx.Controller(), owner, 2)
	switch len(looked) {
	case 0:
		return nil
	case 1:
		return g.PutCardsIntoGraveyardKeyedForEffect(looked, game.KeyedTail{})
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    ctx.Controller(),
		FromPlayer: owner,
		Source:     ctx.Source(),
		Question:   "Wu Spy — put one of them into their graveyard",
		Cards:      looked,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneLibrary,
		ThenKey:    s82OldB714WuSpyPicked,
		Carry:      looked,
	})
	return nil
}

// s82OldB714WuSpyPicked is the card chosen to go to the graveyard. An
// on-disk key: never renamed, never reused.
var s82OldB714WuSpyPicked = game.RegisterChooseCardsThen("choose-cards/s82-old-b714-wu-spy-graveyard", func(g *game.Game, r game.CardsChosen) error {
	ids := CardsStillInALibrary(g, r.Picked)
	if len(ids) == 0 {
		return nil
	}
	return g.PutCardsIntoGraveyardKeyedForEffect(ids, game.KeyedTail{})
})

// S82OldB714WeiAssassinsChoose is Wei Assassins' trigger: "Target opponent
// chooses a creature they control. Destroy that creature." The choice is
// the TARGET opponent's, made at resolution and not targeting; one creature
// is a forced choice and needs no prompt, none does nothing. The creature
// is destroyed, not sacrificed, so indestructible and regeneration apply.
func S82OldB714WeiAssassinsChoose(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	player, ok := FirstLegalPlayerTarget(ctx)
	if !ok {
		return nil
	}
	var candidates []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == player && c.IsCreature() {
			candidates = append(candidates, c.InstanceID)
		}
	}
	switch len(candidates) {
	case 0:
		return nil
	case 1:
		return (DestroyTarget{Target: candidates[0]}).Apply(ctx.AsGroupMember())
	}
	g.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:    player,
		FromPlayer: player,
		Source:     ctx.Source(),
		Question:   "Wei Assassins — choose a creature you control; it is destroyed",
		Cards:      candidates,
		Min:        1,
		Max:        1,
		LiveZone:   game.ZoneBattlefield,
		ThenKey:    s82OldB714WeiAssassinsPicked,
		Asked:      game.NewAskedItem(item),
	})
	return nil
}

// s82OldB714WeiAssassinsPicked destroys the creature the opponent chose, if
// it is still a creature that player controls. An on-disk key: never
// renamed, never reused.
var s82OldB714WeiAssassinsPicked = AskedPick("choose-cards/s82-old-b714-wei-assassins-destroy", func(ctx *Context, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(r.Picked[0])
	if !ok || !c.IsCreature() || c.Controller != r.Chooser {
		return nil
	}
	return (DestroyTarget{Target: r.Picked[0]}).Apply(ctx.AsGroupMember())
})

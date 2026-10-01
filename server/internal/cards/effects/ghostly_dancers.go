package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ghostly Dancers — Creature — Spirit {3}{W}{W}, 2/5:
//
//	"Flying
//	 When this creature enters, return an enchantment card from your
//	 graveyard to your hand or unlock a locked door of a Room you
//	 control.
//	 Eerie — Whenever an enchantment you control enters and whenever
//	 you fully unlock a Room, create a 3/1 white Spirit creature token
//	 with flying."
//
// The enter trigger does not target: the controller chooses the branch
// as it resolves, and then the card or the door. Only a branch that can
// be taken is offered (an enchantment card in your graveyard; a Room you
// control with a locked door), and with one branch there is nothing to
// ask. The door pick is UnlockALockedDoorOfARoomYouControl, so a
// fully-unlocking door triggers eerie (this card's own, in particular)
// as it should. Eerie is one ability with two conditions (Eerie,
// rooms.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c53d958a-f660-4d2e-87cd-87702f973b3a",
		Name:            "Ghostly Dancers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ghostly Dancers — return an enchantment card from your graveyard to your hand or unlock a locked door of a Room you control", ghostlyDancersEnters),
			Eerie("Ghostly Dancers — create a 3/1 white Spirit with flying (eerie)",
				Do(CreateToken{Template: TokenCard("3/1 white Spirit with flying"), N: 1})),
		},
	})
}

func ghostlyDancersEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	me := ctx.Controller()
	var enchantments []uuid.UUID
	if p := ctx.PlayerByID(me); p != nil && p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			if c.IsEnchantment() {
				enchantments = append(enchantments, c.InstanceID)
			}
		}
	}
	hasLockedDoor := false
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == me && game.HasSharedTypeLine(c) && !c.Unlocked.Full() {
			hasLockedDoor = true
			break
		}
	}
	returnCard := func(ctx *Context) error { return ghostlyDancersReturn(ctx, enchantments) }
	unlock := func(ctx *Context) error { return UnlockALockedDoorOfARoomYouControl{Player: me}.Apply(ctx) }
	switch {
	case len(enchantments) > 0 && hasLockedDoor:
		return PickOption{
			Player:   me,
			Question: "Ghostly Dancers — choose one",
			Options: []game.ChoiceOption{
				{Label: "Return an enchantment card from your graveyard to your hand"},
				{Label: "Unlock a locked door of a Room you control"},
			},
			Then: func(ctx *Context, index int) error {
				switch index {
				case 0:
					return returnCard(ctx)
				case 1:
					return unlock(ctx)
				}
				return nil
			},
		}.Apply(ctx)
	case len(enchantments) > 0:
		return returnCard(ctx)
	case hasLockedDoor:
		return unlock(ctx)
	}
	return nil
}

// ghostlyDancersReturn puts the chosen enchantment card from the
// controller's graveyard into their hand; one candidate is not asked.
func ghostlyDancersReturn(ctx *Context, candidates []uuid.UUID) error {
	if len(candidates) == 1 {
		return ReturnFromGraveyard{Target: candidates[0], Dest: game.ZoneHand}.Apply(ctx)
	}
	return takeOneIntoYourHand(ctx, candidates, "Return an enchantment card from your graveyard to your hand", 1)
}

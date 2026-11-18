package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valduk, Keeper of the Flame — {2}{R} Legendary Creature — Human Shaman,
// 3/2:
//
//	"At the beginning of combat on your turn, for each Aura and Equipment
//	 attached to Valduk, create a 3/1 red Elemental creature token with
//	 trample and haste. Exile those tokens at the beginning of the next end
//	 step."
//
// The count is read as the trigger resolves: the Auras and Equipment
// attached to Valduk now, or as he last existed if he has left (CR 608.2h,
// msc16EquipmentAttachedToTheSource's reading, widened to Auras). One
// creation, then the keyed tail hands the tokens that landed to the end
// step exile (CR 603.7), as Manaform Hellkite's does. No tokens, no exile.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "102b0a65-f2e0-4ef3-a9dc-4a85557b852e",
		Name:         "Valduk, Keeper of the Flame",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtBeginningOfYourCombat("Valduk, Keeper of the Flame — for each Aura and Equipment attached to him, create a 3/1 red Elemental with trample and haste; exile those tokens at the beginning of the next end step",
				valdukElementals),
		},
	})
}

// rprB03AurasAndEquipmentAttachedToTheSource counts the Auras and Equipment
// attached to the ability's source: live while it is on the battlefield,
// else those attached as it last existed (CR 608.2h).
func rprB03AurasAndEquipmentAttachedToTheSource(ctx *Context) int {
	n := 0
	if info, ok := ctx.SourcePermanent(); ok && info.Left {
		for _, ref := range info.Attachments {
			if c, found := ctx.Game.LookupCardForEffect(ref.ID); found && (c.IsAura() || c.HasSubtype("Equipment")) {
				n++
			}
		}
		return n
	}
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if (c.IsAura() || c.HasSubtype("Equipment")) && c.IsAttachedTo(ctx.Source()) {
			n++
		}
	}
	return n
}

func valdukElementals(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	n := rprB03AurasAndEquipmentAttachedToTheSource(ctx)
	if n <= 0 {
		return nil
	}
	return g.CreateTokensKeyedForEffect(game.TokenCreation{
		Controller: ctx.Controller(),
		Source:     ctx.Source(),
		Groups:     []game.TokenGroup{{Template: ctx.TextToken(TokenCard("3/1 red Elemental with trample and haste")), Count: n}},
	}, ctx.ItemTail(valdukExileTheElementals))
}

// valdukExileTheElementals is "exile those tokens at the beginning of the
// next end step". An on-disk key.
var valdukExileTheElementals = socB8ExileTokensAtNextEndStep("valduk-keeper-of-the-flame/exile-the-elementals",
	"Valduk, Keeper of the Flame — exile the Elementals")

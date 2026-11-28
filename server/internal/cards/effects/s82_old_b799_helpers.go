package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b799_helpers.go — S82 OLD batch 799 (Mirage, #5633).

// OldB799EtherWell is "put target creature on top of its owner's library. If
// that creature is red, you may put it on the bottom of its owner's library
// instead": a red target is Jace's top-or-bottom placement (the caster
// chooses, CR 401.4), any other goes on top. The colour is read as the spell
// resolves.
func OldB799EtherWell(item *game.StackItem, ctx *Context) error {
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		c, ok := ctx.Game.LookupCardForEffect(t.ID)
		if !ok {
			return nil
		}
		if !c.HasColorWord("R") {
			return PutTargetOnTopOfOwnersLibrary(item, ctx)
		}
		return PutInLibraryInAnyOrder{
			Cards:     []uuid.UUID{t.ID},
			From:      game.ZoneBattlefield,
			Placement: game.LibraryPlaceTopOrBottom,
			Label:     "Ether Well — put it on the top or the bottom of its owner's library",
		}.Apply(ctx)
	}
	return nil
}

// stuporDiscardChosen is Stupor's second half, "then discards a card": the
// opponent chooses it, after the random discard has settled. The tail's
// Player is that opponent. An on-disk key.
var stuporDiscardChosen = CardTail("then/s82-old-b799-stupor-discard-a-card", func(ctx *Context, o game.TailOutcome) error {
	return DiscardOne(ctx.Game, ctx.Item, o.Tail.Player)
})

// OldB799Stupor is "target opponent discards a card at random, then discards
// a card": the random discard first (a keyed tail told it has settled), then
// the opponent's own choice from what is left. A hand emptied by the first
// discards nothing more (CR 701.9b).
func OldB799Stupor(_ *game.StackItem, ctx *Context) error {
	opp := TargetedPlayerOf(ctx)
	if opp == uuid.Nil {
		return nil
	}
	tail := ctx.ItemTail(stuporDiscardChosen)
	tail.Player = opp
	return ctx.Game.DiscardRandomKeyedForEffect(opp, 1, tail)
}

// prismaticLaceColors is Prismatic Lace's second half: the target permanent
// becomes exactly the chosen colors, with no duration (CR 611.2a, 613.1e),
// replacing its own. The chosen letters are the tail's Word. An on-disk key.
var prismaticLaceColors = CardTail("then/s82-old-b799-prismatic-lace-becomes-colors", func(ctx *Context, o game.TailOutcome) error {
	colors := game.ColorSetLetters(o.Tail.Word)
	if len(colors) == 0 {
		return nil
	}
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return ScopedEffectFor{
			Target:   t.ID,
			Mods:     []game.Mod{game.SetColorsMod(colors...)},
			Duration: game.IndefiniteDuration(),
			Label:    "Prismatic Lace — becomes " + game.ColorSetName(o.Tail.Word),
		}.Apply(ctx)
	}
	return nil
})

// OldB799PrismaticLace asks for the colors of the caster's choice, then makes
// the target permanent those colors. A target already gone makes the spell do
// nothing and asks nothing (CR 608.2b).
func OldB799PrismaticLace(item *game.StackItem, ctx *Context) error {
	if len(ctx.LegalTargets()) == 0 {
		return nil
	}
	ChooseColorsTail(game.ColorForBenefit, ctx.Game, item.Controller, item.SourceCardID,
		"Prismatic Lace — choose the color or colors the permanent becomes", ctx.ItemTail(prismaticLaceColors))
	return nil
}

// OldB799PrismaticBoon is "choose a color. X target creatures gain protection
// from the chosen color until end of turn": Bathe in Light's color choice and
// grant, over the creatures targeted rather than a shared-color set. The
// creatures still legal as the spell resolves are the ones that gain it.
func OldB799PrismaticBoon(item *game.StackItem, ctx *Context) error {
	ids := LegalTargetCardIDs(ctx)
	if len(ids) == 0 {
		return nil
	}
	then := ctx.ItemTail(RadianceProtectionFromChosenColor)
	then.Carry = ids
	then.Words = []string{"Prismatic Boon"}
	ChooseColorTail(game.ColorForProtection, ctx.Game, item.Controller, ctx.Source(), "Prismatic Boon — choose a color", then)
	return nil
}

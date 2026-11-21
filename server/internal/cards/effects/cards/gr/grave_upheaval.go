package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Grave Upheaval — Sorcery {4}{B}{R} (Commander 2016, #5537):
//
//	"Put target creature card from a graveyard onto the battlefield under
//	 your control. It gains haste.
//	 Basic landcycling {2} ({2}, Discard this card: Search your library
//	 for a basic land card, reveal it, put it into your hand, then
//	 shuffle.)"
//
// From the Catacombs' target (a creature card in any graveyard) with
// Kami of Industry's keyed return: the card enters under the caster's
// control and, once it is on the battlefield, gains haste with no end
// (the card names no duration), pinned to that object so it ends if the
// creature leaves. A card that left the graveyard in response makes the
// spell fizzle. Basic landcycling is the shared cycling row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5b43523-809a-4baf-b76d-9f59b82846b6",
		Name:         "Grave Upheaval",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature card from a graveyard", effects.Creature()),
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		OnResolve:    graveUpheavalResolve,
		Activated:    []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}

func graveUpheavalResolve(_ *game.StackItem, ctx *effects.Context) error {
	id, ok := effects.B16FirstLegalTargetCard(ctx)
	if !ok {
		return nil
	}
	if z := ctx.Game.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	return ctx.Game.ReturnFromGraveyardToBattlefieldKeyedForEffect(id, ctx.Controller(), false, ctx.ItemTail(graveUpheavalHaste))
}

// graveUpheavalHaste is "It gains haste." An on-disk key.
var graveUpheavalHaste = effects.EnteredTail("grave-upheaval/haste", func(ctx *effects.Context, entered uuid.UUID, _ game.TailOutcome) error {
	if entered == uuid.Nil || !effects.OnBattlefield(ctx.Game, entered) {
		return nil
	}
	return effects.ScopedEffectFor{
		Target:   entered,
		Mods:     []game.Mod{game.AddKeywordsMod("haste")},
		Duration: ctx.Game.PinnedTo(game.IndefiniteDuration(), entered),
		Label:    "Grave Upheaval — it gains haste",
	}.Apply(ctx)
})

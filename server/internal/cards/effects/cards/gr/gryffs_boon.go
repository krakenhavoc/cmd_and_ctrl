package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gryff's Boon — Enchantment — Aura {W}:
//
//	"Enchant creature
//	 Enchanted creature gets +1/+0 and has flying.
//	 {3}{W}: Return this card from your graveyard to the battlefield attached
//	 to target creature. Activate only as a sorcery."
//
// The graveyard ability targets a creature as it is activated. On
// resolution the Aura returns only if it is still in the graveyard and the
// creature is still a legal host (CR 303.4j: otherwise it stays put); it
// enters attached, never unattached (CR 303.4f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f62f3f4e-d506-43a9-91d8-bbc86e0a9167",
		Name:         "Gryff's Boon",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Targets:      effects.EnchantCreature(),
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 0),
			effects.GrantToAttached("flying"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:        "{3}{W}: Return this card from your graveyard to the battlefield attached to target creature. Activate only as a sorcery.",
			Cost:         effects.ManaCost("{3}{W}"),
			Zones:        []game.ZoneKind{game.ZoneGraveyard},
			SorcerySpeed: true,
			Targets:      effects.TargetCreature("target creature"),
			Effect:       gryffsBoonReturn,
		}},
	})
}

func gryffsBoonReturn(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	var t game.TargetRef
	found := false
	for _, lt := range ctx.LegalTargets() {
		if lt.Kind == game.TargetCard {
			t, found = lt, true
			break
		}
	}
	if !found {
		return nil
	}
	if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
		return nil
	}
	card, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !effects.CanBeAttachedTo(g, item.Controller, card, t.ID) {
		return nil
	}
	return effects.ReturnFromGraveyardTogether{
		Targets:    []uuid.UUID{item.SourceCardID},
		AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: t.ID},
	}.Apply(ctx)
}

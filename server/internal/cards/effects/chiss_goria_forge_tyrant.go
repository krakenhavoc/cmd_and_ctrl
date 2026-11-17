package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Chiss-Goria, Forge Tyrant — Legendary Creature — Dragon {6}{R}{R}{R},
// 5/4 (Phyrexia: All Will Be One Commander):
//
//	"Affinity for artifacts (This spell costs {1} less to cast for each
//	 artifact you control.)
//	 Flying, haste
//	 Whenever Chiss-Goria attacks, exile the top five cards of your library.
//	 You may cast an artifact spell from among them this turn. If you do, it
//	 has affinity for artifacts."
//
// The attack trigger exiles the five together, then opens ONE permission
// over the artifact cards among them (CastsLeft 1: "an artifact spell",
// CR 611.2c locks the set) until end of turn; a card that is not an artifact
// is stranded in exile. The cast is at the card's own timing and paid for.
// "It has affinity for artifacts" is the permission's AffinityFor: the spell
// cast through it costs {1} less for each artifact its caster controls as it
// is cast (CR 702.41a, CR 601.2f), on top of any other reduction.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3f4d1ae1-d5c6-4afb-94c8-8e0ae4891501",
		Name:            "Chiss-Goria, Forge Tyrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Chiss-Goria, Forge Tyrant — exile the top five cards of your library; you may cast an artifact spell from among them this turn, and it has affinity for artifacts",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					return g.ExileTopKeyedForEffect(ctx.Controller(), ctx.Controller(), 5, ctx.Tail(chissGoriaGrant))
				}),
		},
	})
}

// chissGoriaGrant is the attack trigger's tail: one cast, of an artifact spell
// from among the exiled cards, this turn, with affinity for artifacts. An
// on-disk key: never renamed, never reused.
var chissGoriaGrant = CardTail("chiss-goria-forge-tyrant/grant-cast", func(ctx *Context, o game.TailOutcome) error {
	g := ctx.Game
	var cards []game.Card
	for _, id := range o.IDs {
		if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
			continue
		}
		if c, ok := g.LookupCardForEffect(id); ok && c.IsArtifact() {
			cards = append(cards, c)
		}
	}
	if len(cards) == 0 {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:      ctx.Controller(),
		Zone:        game.ZoneExile,
		Duration:    g.UntilEndOfTurnDuration(),
		CastOnly:    true,
		CastsLeft:   1,
		AffinityFor: "Artifact",
		Filter:      game.PermissionFilter{CardType: "artifact"},
		Source:      ctx.Source(),
		SourceName:  "Chiss-Goria, Forge Tyrant",
		Label:       "Chiss-Goria, Forge Tyrant — cast an artifact spell from among the exiled cards (it has affinity for artifacts)",
	}, cards)
	return nil
})

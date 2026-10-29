package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Thranduil's Decree — Instant {4}{U}{U} (The Hobbit):
//
//	"Counter target spell. If a permanent spell is countered this way,
//	 exile it instead of putting it into its owner's graveyard. You may
//	 cast that card without paying its mana cost for as long as it remains
//	 exiled."
//
// Force of Negation's counter-to-exile (CounterTarget.Dest) for a permanent
// spell, and an ordinary counter for anything else. The cast permission is
// written only if the card actually reached exile: a spell that cannot be
// countered stays on the stack and is not "countered this way", and a
// commander whose owner took the command zone instead is not in exile
// either. The permission is the Decree controller's, at {0} (so X is 0),
// cast only, at the card's own timing (a creature still needs a main phase
// of an empty stack), for as long as that exile object stays in exile
// (CR 611.2b, WhileInZone); the exiled card is a new object (CR 400.7), so
// a card that leaves exile and returns is not castable on this grant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d7cba934-02ad-4677-bb4d-50808b01b4f9",
		Name:         "Thranduil's Decree",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			stackID := item.Targets[0].ID
			spell, ok := ctx.Game.LookupCardForEffect(stackID)
			if !ok || ctx.Game.StackItemForEffect(stackID) == nil {
				return nil
			}
			if !spell.IsPermanent() {
				return CounterTarget{StackID: stackID}.Apply(ctx)
			}
			if err := (CounterTarget{StackID: stackID, Dest: game.ZoneRef{Kind: game.ZoneExile}}).Apply(ctx); err != nil {
				return err
			}
			return thranduilsDecreeGrant(ctx, stackID)
		},
	})
}

// thranduilsDecreeGrant is "you may cast that card without paying its mana
// cost for as long as it remains exiled", over the card if it is in exile.
func thranduilsDecreeGrant(ctx *Context, cardID uuid.UUID) error {
	g := ctx.Game
	z := g.FindCardZoneForEffect(cardID)
	c, ok := g.LookupCardForEffect(cardID)
	if !ok || z == nil || z.Kind != game.ZoneExile || c.IsToken() {
		return nil
	}
	g.GrantCastPermissionToCardsForEffect(game.CastPermission{
		Player:     ctx.Controller(),
		Zone:       game.ZoneExile,
		Duration:   game.WhileInZoneDuration(),
		Cost:       "{0}",
		AltCostKey: game.AltCostKeyGrantedCost,
		CastOnly:   true,
		Source:     ctx.Source(),
		SourceName: "Thranduil's Decree",
		Label:      "Thranduil's Decree — cast " + c.Name + " without paying its mana cost",
	}, []game.Card{c})
	return nil
}

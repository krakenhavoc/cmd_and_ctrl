package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b453_helpers.go — Commander 2011 (S82 older sets, batch 453, #5574):
// the shared shapes of Nemesis Trap, the two wrath Demons and Spurnmage
// Advocate.

// S82OldB453AWhiteCreatureIsAttacking is Nemesis Trap's alternative-cost gate,
// "if a white creature is attacking": any creature attacking, anyone's, whose
// layered colours include white.
func S82OldB453AWhiteCreatureIsAttacking(g *game.Game, controller uuid.UUID) bool {
	white := OfColor("W")
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.AttackingTarget != uuid.Nil && white(g, controller, c) {
			return true
		}
	}
	return false
}

// S82OldB453NemesisTrap is "Exile target attacking creature. Create a token
// that's a copy of that creature. Exile it at the beginning of the next end
// step." The token is a copy of the creature as it last existed on the
// battlefield (its copiable values), made under the Trap's controller only if
// the exile happened.
func S82OldB453NemesisTrap(item *game.StackItem, ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	return ExileTarget{Target: t.ID, ThenTail: ctx.ItemTail(s82OldB453NemesisTrapCopy)}.Apply(ctx)
}

var s82OldB453NemesisTrapCopy = CardTail("nemesis-trap/copy", func(ctx *Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return TrekB40CreateCopyExiledAtNextEndStep(ctx, ctx.Controller(), o.IDs[0], func(*game.Card) {},
		"Nemesis Trap — exile the token")
})

// S82OldB453DestroyThenTapOthers is Dread Cacodemon's "destroy all creatures
// your opponents control, then tap all other creatures you control", run only
// when the Cacodemon was cast from hand.
func S82OldB453DestroyThenTapOthers(g *game.Game, item *game.StackItem) error {
	if !B30CastFromHand(g, item.SourceCardID) {
		return nil
	}
	ctx := NewContext(g, item)
	if err := (DestroyAllMatching{Match: And(Creature(), OpponentControls())}).Apply(ctx); err != nil {
		return err
	}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && c.Controller == item.Controller && c.InstanceID != item.SourceCardID {
			if err := (TapTarget{Target: c.InstanceID}).Apply(ctx.AsGroupMember()); err != nil {
				return err
			}
		}
	}
	return nil
}

// S82OldB453DestroyNonartifactNonblack is Reiver Demon's "destroy all
// nonartifact, nonblack creatures. They can't be regenerated.", run only when
// the Demon was cast from hand.
func S82OldB453DestroyNonartifactNonblack(g *game.Game, item *game.StackItem) error {
	if !B30CastFromHand(g, item.SourceCardID) {
		return nil
	}
	return DestroyAllMatching{
		Match:             And(Creature(), Not(Artifact()), NonBlack()),
		CantBeRegenerated: true,
	}.Apply(NewContext(g, item))
}

// S82OldB453SpurnmageAdvocate is "Return two target cards from an opponent's
// graveyard to their hand. Destroy target attacking creature." Each target
// is checked on its own as the ability resolves (CR 608.2b).
func S82OldB453SpurnmageAdvocate(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.ClauseTargets(0) {
		if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
			if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneHand}).Apply(ctx); err != nil {
				return err
			}
		}
	}
	if t, ok := ctx.ClauseTarget(1); ok && t.Kind == game.TargetCard {
		return DestroyTarget{Target: t.ID}.Apply(ctx)
	}
	return nil
}

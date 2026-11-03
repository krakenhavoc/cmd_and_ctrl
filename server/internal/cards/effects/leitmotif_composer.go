package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Leitmotif Composer — Creature — Human Bard {2}{U}, 2/2 (Secrets of
// Strixhaven Commander, #4834):
//
//	"Whenever this creature deals combat damage to a player, draw a card.
//	 Whenever you cast an instant or sorcery spell with mana value 5 or
//	 greater, create a token that's a copy of this creature.
//	 {2}{U}: Creatures named Leitmotif Composer can't be blocked this
//	 turn."
//
// The spell's mana value is read on the stack, X included (CR 202.3e).
// The copy is of the Composer's copiable values as the ability resolves,
// or as it last existed if it has left (CR 707.2, CR 608.2h), so it
// works off a token Composer that died in response, and copies whatever
// the Composer is copying (the ruling); the copy has the abilities too.
//
// "Creatures named Leitmotif Composer" is the set of creatures with that
// name as the ability resolves, every controller's, locked in then (CR
// 611.2c): one that arrives later in the turn can be blocked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61816f98-9d1a-4b16-9194-cb7249880710",
		Name:         "Leitmotif Composer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Leitmotif Composer — draw a card", Do(DrawCards{N: 1})),
			WheneverYouCast(And(Or(Instant(), Sorcery()), ManaValueGE(5)),
				"Leitmotif Composer — create a token that's a copy of this creature",
				leitmotifComposerCopy),
		},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Creatures named Leitmotif Composer can't be blocked this turn.",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{2}{U}"),
			Effect:  leitmotifComposerUnblockable,
		}},
	})
}

// leitmotifComposerCopy is the cast trigger's token copy of the source
// object, live or as it last existed.
func leitmotifComposerCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ref, ok := ctx.SourceRef()
	if !ok {
		return nil
	}
	tmpl, ok := g.TokenCopyTemplateForObjectForEffect(ref)
	if !ok {
		return nil
	}
	return createTokenCopies(ctx, ctx.Controller(), tmpl, 1)
}

// leitmotifComposerUnblockable makes every creature named Leitmotif
// Composer on the battlefield unblockable this turn.
func leitmotifComposerUnblockable(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	var named []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.IsCreature() && game.HasName(c, "Leitmotif Composer") {
			named = append(named, c.InstanceID)
		}
	}
	for _, id := range named {
		if err := (RestrictUntilEOT{Target: id, Restrictions: game.CantBeBlocked,
			Label: "Leitmotif Composer — can't be blocked"}).Apply(ctx.asGroupMember()); err != nil {
			return err
		}
	}
	return nil
}

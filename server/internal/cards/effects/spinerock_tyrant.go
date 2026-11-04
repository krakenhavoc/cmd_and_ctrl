package effects

import (
	"errors"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spinerock Tyrant — Creature — Dragon {3}{R}{R}, 6/6 (Lorwyn Eclipsed,
// #4458, #4693):
//
//	"Flying
//	 Wither (This deals damage to creatures in the form of -1/-1
//	 counters.)
//	 Whenever you cast an instant or sorcery spell with a single target,
//	 you may copy it. If you do, those spells gain wither. You may choose
//	 new targets for the copy."
//
// Flying and wither ride PrintedKeywords (ADR 0056).
//
// The trigger fires on a cast whose spell chose exactly one target
// (HasASingleTarget, Bolt Bend's reading: a spell that named one object
// for two instances of "target" has two). The "you may" is asked as the
// trigger resolves (CR 603.5). The trigger names the spell without
// targeting it, so a spell countered in response is still copied from
// last-known information (CR 608.2h), as Spider-Verse's is; its
// controller may choose new targets for the copy (CR 707.10c).
//
// "Those spells gain wither" is an ability-adding effect of the
// resolving trigger (CR 611.2, 613.1f) on two objects, given through the
// stack step of the layer pass (ADR 0107 §3): the original by
// GrantKeywordsToSpellForEffect, a ScopedEffect pinned to it, and the
// copy by the copy-making call's copyGrant once the copy exists. Each
// lasts while that object is on the stack (CR 400.7). Neither is a
// copiable value (CR 707.2), so a copy some other effect makes of either
// spell does not have wither. The damage tail reads the spell's keywords
// off the stack (CR 702.80c), so its damage to a creature is -1/-1
// counters (CR 702.80a, 120.3d) and its damage to a player is ordinary
// life loss (CR 120.3a). An original that has left the stack gains
// nothing; the copy still does. A spell that can't be copied (See
// Double) is not copied, so neither spell gains wither ("if you do").
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e35d4c62-5211-4785-b683-a83719195d87",
		Name:            "Spinerock Tyrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordWither},
		Triggered: []game.TriggeredAbility{
			Optional(On(game.EventCast, YouCast(And(Or(Instant(), Sorcery()), HasASingleTarget())),
				spinerockTyrantLabel, spinerockTyrantCopy),
				"Spinerock Tyrant — copy that spell? If you do, both spells gain wither."),
		},
	})
}

const spinerockTyrantLabel = "Spinerock Tyrant — copy that spell; those spells gain wither"

// spinerockTyrantCopy copies the spell the cast named, then gives the
// original wither while it is still that spell on the stack; the copy
// gains it as it is put on the stack, after any new targets are chosen.
func spinerockTyrantCopy(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	spell := item.Trigger.Event.CardID
	if c, ok := g.LookupCardForEffect(spell); ok && game.CatalogCantBeCopied != nil && game.CatalogCantBeCopied(game.CatalogKey(c)) {
		return nil
	}
	wither := []string{game.KeywordWither}
	err := g.CopyLastKnownSpellGainingKeywordsForEffect(item.SourceCardID, spell, item.Controller, true, wither,
		"Spinerock Tyrant — the copy gains wither")
	if errors.Is(err, game.ErrCardNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	g.GrantKeywordsToSpellForEffect(item.SourceCardID, spell, wither, game.IndefiniteDuration(),
		"Spinerock Tyrant — that spell gains wither")
	return nil
}

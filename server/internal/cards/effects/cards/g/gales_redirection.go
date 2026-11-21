package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gale's Redirection — Instant {3}{U}{U} (Commander Legends: Battle for
// Baldur's Gate, S82 GAP #5948):
//
//	"Exile target spell, then roll a d20 and add that spell's mana value.
//	 1—14 | You may cast the exiled card for as long as it remains exiled, and
//	 you may spend mana as though it were mana of any color to cast that spell.
//	 15+ | You may cast the exiled card without paying its mana cost for as
//	 long as it remains exiled."
//
// An exile, not a counter (ExileTargetSpell: a spell that can't be countered
// is still exiled). The spell's mana value, X counted as it was cast (CR
// 202.3e), is read before it moves and rides the exile's continuation, which
// runs only when the spell really reached exile (a commander sent to the
// command zone is not there to cast). The roll adds it as the instruction's
// modifier (CR 706.2), so "whenever you roll a natural 20" would still see
// the die; the table is read on the sum (CR 706.3a). The permission is the
// caster's, not the owner's: a standing grant over the exiled card, ending
// when it leaves exile (Kheru Spellsnatcher's shape): the printed cost with
// any-color mana on 1—14, a {0} cast on 15+. A sum below 1 is no row.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8676f007-bf5f-4312-968c-59f4549f7d51",
		Name:         "Gale's Redirection",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			id := item.Targets[0].ID
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok || ctx.Game.StackItemForEffect(id) == nil {
				return nil
			}
			mv, _ := ctx.Game.ManaValueForEffect(c)
			t := ctx.ItemTail(galesRedirectionExiled)
			t.N = mv
			return effects.ExileTargetSpell{StackID: id, ThenTail: t}.Apply(ctx)
		},
	})
}

// galesRedirectionExiled rolls the d20 once the spell is in exile, adding the
// mana value read before it left the stack. An on-disk key.
var galesRedirectionExiled = effects.CardTail("then/gales-redirection-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	if len(o.IDs) == 0 {
		return nil
	}
	return effects.RollDice{Sides: 20, N: 1, Modifier: o.Tail.N, Then: galesRedirectionRolled, TailCarry: append([]uuid.UUID(nil), o.IDs...)}.Apply(ctx)
})

// galesRedirectionRolled reads the table on the sum. An on-disk key.
var galesRedirectionRolled = effects.CardTail("dice/gales-redirection-rolled", func(ctx *effects.Context, o game.TailOutcome) error {
	sum := effects.DiceResult(o)
	if sum < 1 || len(o.Tail.Carry) == 0 {
		return nil
	}
	g := ctx.Game
	card, ok := g.LookupCardForEffect(o.Tail.Carry[0])
	if !ok {
		return nil
	}
	if z := g.FindCardZoneForEffect(card.InstanceID); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	perm := game.CastPermission{
		Player:     ctx.Controller(),
		Zone:       game.ZoneExile,
		CastOnly:   true,
		Duration:   game.WhileInZoneDuration(),
		Source:     ctx.Source(),
		SourceName: "Gale's Redirection",
	}
	if sum >= 15 {
		perm.Cost = "{0}"
		perm.Label = "Gale's Redirection — you may cast it without paying its mana cost for as long as it remains exiled"
	} else {
		perm.AnyColor = true
		perm.Label = "Gale's Redirection — you may cast it for as long as it remains exiled, spending mana as though it were mana of any color"
	}
	g.GrantCastPermissionToCardsForEffect(perm, []game.Card{card})
	return nil
})

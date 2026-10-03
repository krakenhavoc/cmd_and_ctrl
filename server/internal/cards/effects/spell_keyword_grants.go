package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// spell_keyword_grants.go — giving a SPELL a keyword (ADR 0107 §3,
// #1854). Append-only, like every mechanic-named helper file.
//
// A spell gains an ability two ways, and both are CR 613.1f layer-6
// effects the stack step of the layer pass applies
// (game/spell_keywords.go):
//
//	SpellsYouControlHave   a battlefield static over spells — Cast
//	                       Through Time's "Instant and sorcery spells
//	                       you control have rebound". Never locked in
//	                       (CR 611.3a).
//	ThatSpellGains         a resolving ability's effect on one spell —
//	                       Taigam's and Ojer Pakpatiq's "that spell
//	                       gains rebound". Data (a ScopedEffect pinned
//	                       to the spell), ending with the object
//	                       (CR 400.7).
//
// The rebound itself needs nothing more: the resolution reads it with
// game.HasKeyword, which sees a granted keyword like a printed one.

// SpellsYouControlHave is "<these> spells you control have <keywords>":
// a static over the spells on the stack whose controller is the
// static's. `spell` narrows which spells (Or(Instant(), Sorcery()) for
// Cast Through Time); nil is every spell. The predicate reads the
// spell's card as it sits on the stack.
func SpellsYouControlHave(spell CardPredicate, keywords ...string) game.StaticAbility {
	return game.StaticAbility{
		Layer:         game.Layer6Ability,
		AffectsSpells: true,
		AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
			if target.Controller != source.Controller {
				return false
			}
			return spell == nil || spell(g, source.Controller, *target)
		},
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			for _, kw := range keywords {
				c.Abilities = game.AppendKeywordAbility(c.Abilities, kw)
			}
		},
	}
}

// ThatSpellGains is "that spell gains <keywords>" for a cast trigger:
// the spell the triggering EventCast named, while it is still that
// object on the stack. A spell countered or resolved before the trigger
// resolves is given nothing: the card in its new zone is a new object
// (CR 400.7), and the spell the ability named no longer exists.
//
// A permanent spell hands what it gained to the permanent it becomes
// (CR 400.7a) — Lux Artillery's "it gains sunburst" is read as the
// permanent enters — and UntilEndOfTurn carries "until end of turn"
// with it (ADR 0109 §11 decision 3): Tyvar Kell's emblem gives an Elf
// spell haste until end of turn, and the creature it becomes keeps the
// haste for the rest of that turn and no longer.
type ThatSpellGains struct {
	Keywords       []string
	UntilEndOfTurn bool
}

func (a ThatSpellGains) Apply(ctx *Context) error {
	spell := ctx.Trigger().Event.CardID
	if spell == uuid.Nil || len(a.Keywords) == 0 {
		return nil
	}
	label := "that spell gains"
	if ctx.Item != nil && ctx.Item.Label != "" {
		label = ctx.Item.Label
	}
	d := game.IndefiniteDuration()
	if a.UntilEndOfTurn {
		d = ctx.Game.UntilEndOfTurnDuration()
	}
	ctx.Game.GrantKeywordsToSpellForEffect(ctx.Source(), spell, a.Keywords, d, label)
	return nil
}

// thatSpellYouJustCastGainsSunburst is the body of "when you next cast
// an artifact spell this turn, that spell gains sunburst" (Solar Array):
// the spell the delayed trigger fired on rides the item's payload, and
// it gains sunburst for as long as it is that object on the stack and
// then as the permanent it becomes (CR 400.7a). A spell countered in
// response gains nothing.
func thatSpellYouJustCastGainsSunburst(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	cast := ctx.PayloadCards()
	if len(cast) == 0 {
		return nil
	}
	g.GrantKeywordsToSpellForEffect(item.SourceCardID, cast[0], []string{game.KeywordSunburst},
		game.IndefiniteDuration(), item.Label)
	return nil
}

// YouCastFromYourHand is "whenever you cast a <spell> spell from your
// hand": YouCast, narrowed to a cast whose EventCast left the caster's
// hand. A cast from exile, a graveyard or the command zone is a cast,
// but not from your hand.
func YouCastFromYourHand(spell CardPredicate) When {
	cast := YouCast(spell)
	return func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
		return ev.OldZone == game.ZoneHand && cast(ev, source, lki, g)
	}
}

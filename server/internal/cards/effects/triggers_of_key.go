package effects

import (
	"strings"
	"sync"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// triggers_of_key.go — the catalog half of game/triggers_of_key.go
// (#5683): a token created "with … all triggered abilities of the exiled
// cards" (The Book of Vile Darkness's Vecna).

// derivedTriggerDefs memoises the definitions lookupDef builds for the
// derived keys. A definition handed out must be pointer-stable (the
// composite-key memo compares parts by pointer) and is read under many
// games' locks at once, so it is built once per key. The catalog is
// complete before the first lookup: the keys are only ever written by a
// resolving ability.
var derivedTriggerDefs sync.Map

// derivedTriggerDef answers a game.TriggersOfKey or game.KeywordTriggerKey
// part, or nil for any other key and for a card with no triggered
// ability.
func derivedTriggerDef(key string) *game.CardDef {
	if !strings.HasPrefix(key, game.GrantKeyPrefix) {
		return nil
	}
	if d, ok := derivedTriggerDefs.Load(key); ok {
		return d.(*game.CardDef)
	}
	var d *game.CardDef
	if lender, ok := game.TriggersOfKeyLender(key); ok {
		text := "All triggered abilities of an exiled card."
		oracleID, _, _ := strings.Cut(lender, "#")
		if spec, found := registry[oracleID]; found && spec.Name != "" {
			text = "All triggered abilities of " + spec.Name + "."
		}
		d = game.TriggersOnlyDef(defs[lender], text)
	} else {
		d = game.KeywordTriggerDef(key)
	}
	if d == nil {
		return nil
	}
	actual, _ := derivedTriggerDefs.LoadOrStore(key, d)
	return actual.(*game.CardDef)
}

// CreateTokenWithTriggeredAbilitiesOf is "create <token> with … all
// triggered abilities of <cards>" (CR 111.3: the creating effect defines
// the token's text). Each lender is read as it is now, where it is
// (game.TriggeredTextOf), and what it has becomes part of the token's
// own copiable text before the token enters, so a lent "when this
// enters" triggers for the token's own arrival.
func CreateTokenWithTriggeredAbilitiesOf(ctx *Context, template game.Card, lenders []game.Card) error {
	for _, l := range lenders {
		parts, keywords := game.TriggeredTextOf(l)
		template.GrantedAbilities = append(template.GrantedAbilities, parts...)
		template.Keywords = append(template.Keywords, keywords...)
	}
	return CreateToken{Controller: ctx.Controller(), Template: template, N: 1}.Apply(ctx)
}

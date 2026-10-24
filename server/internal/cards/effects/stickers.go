package effects

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// stickers.go — the catalog's half of stickers and ticket counters
// (#2138, ADR 0208; the engine half is game/sticker.go).
//
//   - GetTickets is "you get {TK}": ticket counters on the controller (CR
//     107.17), through the one player-counter door, so a trigger on
//     EventPlayerCounterPlaced sees them.
//   - PutASticker is "you may put a sticker on <an object you own>" (CR
//     123.3): the engine asks which object, which sticker, and for a name
//     sticker where it goes, each a keyed restore point
//     (game.PutStickerForEffect). The ticket cost is paid as it goes on.
//   - Spec.StickerSheet is a sticker sheet's entry: what each of its
//     ability stickers gives the object it is on, as keywords or a grant
//     bundle in Spec.Grants (ADR 0093). Its name, power and toughness, and
//     art stickers need no entry; the rules say what they do (CR 123.6,
//     123.8, 123.9).

// GetTickets is "you get N {TK}": N ticket counters on the controller of
// the resolving spell or ability. Zero or less gets nothing.
type GetTickets struct {
	N int
}

// Apply puts the counters on.
func (e GetTickets) Apply(ctx *Context) error {
	if e.N <= 0 {
		return nil
	}
	you := ctx.Controller()
	return ctx.Game.AddPlayerCounterByForEffect(you, you, game.CounterTicket, e.N)
}

// TicketSymbols is n ticket symbols, "{TK}{TK}", for a stack label.
func TicketSymbols(n int) string {
	if n <= 0 {
		return ""
	}
	return game.RepeatMana("{TK}", n)
}

// StickerOn names the objects a "put a sticker on …" instruction allows.
type StickerOn int

const (
	// StickerOnNonlandPermanentYouOwn is "a nonland permanent you own".
	StickerOnNonlandPermanentYouOwn StickerOn = iota
	// StickerOnCreatureYouOwn is "a creature you own".
	StickerOnCreatureYouOwn
	// StickerOnCreatureYouOwnThatEnteredThisTurn is "a creature you own
	// that entered this turn" (Park Bleater).
	StickerOnCreatureYouOwnThatEnteredThisTurn
)

// PutASticker is "you may put a sticker on <On>", optionally of one kind
// ("a power and toughness sticker"). Every printed instruction says "you
// may", so the engine's first question can be answered with nothing.
type PutASticker struct {
	On    StickerOn
	Kinds []game.StickerKind
}

// Apply asks the controller.
func (e PutASticker) Apply(ctx *Context) error {
	g := ctx.Game
	you := ctx.Controller()
	g.RecomputeLayersIfStaleLocked()
	var objects []uuid.UUID
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Owner != you {
			continue
		}
		switch e.On {
		case StickerOnNonlandPermanentYouOwn:
			if c.IsLand() {
				continue
			}
		case StickerOnCreatureYouOwn:
			if !c.IsCreature() {
				continue
			}
		case StickerOnCreatureYouOwnThatEnteredThisTurn:
			if !c.IsCreature() || !g.EnteredThisTurn(c.InstanceID) {
				continue
			}
		}
		objects = append(objects, c.InstanceID)
	}
	g.PutStickerForEffect(game.StickerRequest{
		Player:   you,
		Source:   ctx.Source(),
		Objects:  objects,
		Kinds:    append([]game.StickerKind(nil), e.Kinds...),
		Question: e.question(),
	})
	return nil
}

func (e PutASticker) question() string {
	kind := "a sticker"
	if len(e.Kinds) == 1 {
		switch e.Kinds[0] {
		case game.StickerName:
			kind = "a name sticker"
		case game.StickerAbility:
			kind = "an ability sticker"
		case game.StickerPT:
			kind = "a power and toughness sticker"
		case game.StickerArt:
			kind = "an art sticker"
		}
	}
	on := "a nonland permanent you own"
	switch e.On {
	case StickerOnCreatureYouOwn:
		on = "a creature you own"
	case StickerOnCreatureYouOwnThatEnteredThisTurn:
		on = "a creature you own that entered this turn"
	}
	return "You may put " + kind + " on " + on + ". Choose it, or none."
}

// GetTicketsThenSticker is the line most sticker cards print: "you get N
// {TK}, then you may put a sticker on a nonland permanent you own".
func GetTicketsThenSticker(n int) Effect {
	return Do(GetTickets{N: n}, PutASticker{})
}

// WhenThisEntersGetTicketsThenSticker is "When this creature enters, you
// get N {TK}, then you may put a sticker on a nonland permanent you own."
// `name` is the card's name, for the stack label.
func WhenThisEntersGetTicketsThenSticker(name string, n int) game.TriggeredAbility {
	return WhenThisEnters(name+" — you get "+TicketSymbols(n)+", then you may put a sticker on a nonland permanent you own",
		GetTicketsThenSticker(n))
}

// YouControlAStickeredPermanent is "as long as you control a stickered
// permanent" (CR 123.4), for a static's AppliesTo.
func YouControlAStickeredPermanent(g *game.Game, you uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == you && c.HasStickers() {
			return true
		}
	}
	return false
}

// Stickered passes for a "stickered" object (CR 123.4): one with any
// sticker on it now.
func Stickered() CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.HasStickers() }
}

// WithSticker passes for an object with a sticker of `kind` on it ("with
// a name sticker on it").
func WithSticker(kind game.StickerKind) CardPredicate {
	return func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return game.StickerCountOfKind(&c, kind) > 0 }
}

// YouPlacedASticker is the When for "whenever you place a sticker" / "put
// a sticker on …": the source's controller put one on (EventStickerPlaced).
func YouPlacedASticker(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Actor == source.Controller
}

// WheneverYouPlaceASticker is "Whenever you place a sticker, <effect>".
// The effect reads which sticker with PlacedStickerKind.
func WheneverYouPlaceASticker(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventStickerPlaced, YouPlacedASticker, label, effect)
}

// PlacedStickerKind is the kind of sticker the triggering placement put
// on, and the object it went on.
func PlacedStickerKind(item *game.StackItem) (game.StickerKind, uuid.UUID) {
	if item == nil || item.Trigger == nil {
		return "", uuid.Nil
	}
	kind, _ := game.StickerEventKindText(item.Trigger.Event)
	return kind, item.Trigger.Event.CardID
}

// CreateTokensForPlacedSticker is "create N <Template> tokens. If it's an
// art sticker, instead create ArtN" in a "whenever you place a sticker"
// trigger (Goblin Airbrusher): the count reads the triggering placement,
// the item's own constant.
type CreateTokensForPlacedSticker struct {
	Template game.Card
	N, ArtN  int
}

// Apply creates the tokens.
func (s CreateTokensForPlacedSticker) Apply(ctx *Context) error {
	n := s.N
	if kind, _ := PlacedStickerKind(ctx.Item); kind == game.StickerArt {
		n = s.ArtN
	}
	return CreateToken{Template: s.Template, N: n}.Apply(ctx)
}

// The step is a token creation whichever count the event picks, as
// CreateXByXToken's is whatever size its X picks.
func (s CreateTokensForPlacedSticker) footprint() (game.FootprintStep, bool) {
	key, manual, keywords := game.FootprintTokenOf(s.Template)
	return game.FootprintStep{
		Kind: game.FootprintCreateToken, N: s.N, TokenKey: key, TokenManual: manual, TokenKeywords: keywords,
	}, s.N > 0 && s.ArtN > 0
}

// StickerSheetSpec is a sticker sheet's catalog entry: what each of its
// ability stickers gives, in printed order (game.StickerPrint.Ordinal).
type StickerSheetSpec struct {
	Abilities []StickerAbilitySpec
}

// StickerAbilitySpec is one ability sticker: keyword tokens, or the key of
// a bundle in the sheet's Spec.Grants for an ability that is not a
// keyword (a triggered or activated one), granted to the object the
// sticker is on with that object as "this creature" (ADR 0093).
type StickerAbilitySpec struct {
	Keywords []string
	Grant    string
}

// stickerSheetSpecs indexes Spec.StickerSheet by the sheet's oracle ID.
// Written by Register at init, read-only afterwards.
var stickerSheetSpecs = map[string]*StickerSheetSpec{}

// registerStickerSheet is Register's half for a sheet: each ability
// sticker gives something, and a grant names a bundle the sheet declares.
func registerStickerSheet(spec Spec) {
	sheet := spec.StickerSheet
	if sheet == nil {
		return
	}
	if len(sheet.Abilities) == 0 {
		panic(fmt.Sprintf("effects.Register: sticker sheet %q declares no ability stickers", spec.Name))
	}
	for i, a := range sheet.Abilities {
		if len(a.Keywords) == 0 && a.Grant == "" {
			panic(fmt.Sprintf("effects.Register: sticker sheet %q ability sticker %d gives nothing", spec.Name, i))
		}
		if a.Grant == "" {
			continue
		}
		found := false
		for _, gr := range spec.Grants {
			if gr.Key == a.Grant {
				found = true
				break
			}
		}
		if !found {
			panic(fmt.Sprintf("effects.Register: sticker sheet %q ability sticker %d grants %q, which it does not declare in Grants", spec.Name, i, a.Grant))
		}
	}
	stickerSheetSpecs[spec.OracleID] = sheet
}

func init() {
	game.CatalogStickerAbility = func(sheet string, ordinal int) (game.StickerAbilityDef, bool) {
		s, ok := stickerSheetSpecs[sheet]
		if !ok || ordinal < 0 || ordinal >= len(s.Abilities) {
			return game.StickerAbilityDef{}, false
		}
		a := s.Abilities[ordinal]
		return game.StickerAbilityDef{Keywords: append([]string(nil), a.Keywords...), Grant: a.Grant}, true
	}
}

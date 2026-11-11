package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// cmdr2_s22_helpers.go — shared bodies of the "draw-matters, hand size
// and flash grants" Commander batch (#5368).

// cmdr2S22HandHas reports whether `player` is still in the game with
// exactly n cards in hand (Triskaidekaphile's intervening "if").
func cmdr2S22HandHas(g *game.Game, player uuid.UUID, n int) bool {
	p := g.PlayerByIDForEffect(player)
	return p != nil && !p.Eliminated && cardsInHand(g, player) == n
}

// cmdr2S22WinIfHandHas is "you win the game" behind that intervening
// "if", checked again as the trigger resolves (CR 603.4).
func cmdr2S22WinIfHandHas(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if !cmdr2S22HandHas(g, item.Controller, n) {
			return nil
		}
		return WinTheGame{Player: item.Controller}.Apply(NewContext(g, item))
	}
}

// cmdr2S22TargetPlayerDrawsThenYouDraw is "target player draws N cards
// and you draw M cards", the target first as printed. A target that
// became illegal draws nothing; the controller still draws.
func cmdr2S22TargetPlayerDrawsThenYouDraw(theirs, yours int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetPlayer {
				continue
			}
			if err := (DrawCards{Player: t.ID, N: theirs}).Apply(ctx); err != nil {
				return err
			}
			break
		}
		return DrawCards{N: yours}.Apply(ctx)
	}
}

// cmdr2S22HexproofAndTapThis is "This creature gains hexproof until end
// of turn. Tap it." as an activated ability's effect, both halves the
// resolving ability's own.
func cmdr2S22HexproofAndTapThis(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		if err := (GrantKeywordUntilEOT{
			Target:   item.SourceCardID,
			Keywords: []string{"hexproof"},
			Label:    label,
		}).Apply(ctx); err != nil {
			return err
		}
		return TapTarget{Target: item.SourceCardID}.Apply(ctx)
	}
}

// cmdr2B08ToughnessIsCardsInYourHand is the toughness half of "power and
// toughness are each equal to the number of cards in your hand" (CR 604.3,
// layer 7a); powerIsCardsInYourHand is the power half.
func cmdr2B08ToughnessIsCardsInYourHand() game.StaticAbility {
	return game.StaticAbility{
		Layer:             game.Layer7PT,
		SubLayer:          game.SubLayer7A_CDA,
		DependsOnHandSize: true,
		AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
			return target.InstanceID == source.InstanceID
		},
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			n := 0
			if p := g.PlayerByIDForEffect(source.Controller); p != nil && p.Hand != nil {
				n = p.Hand.Size()
			}
			c.Toughness = n
		},
	}
}

// cmdr2S22DrawThenMayReturnALand is Kefnet the Mindful's "Draw a card,
// then you may return a land you control to its owner's hand": the pick
// is made after the draw (CR 608.2c), so a land that was just drawn is
// not a candidate (it is in hand) but one the draw let you play is.
func cmdr2S22DrawThenMayReturnALand(g *game.Game, item *game.StackItem) error {
	return Do(DrawCards{N: 1}, Instruction(func(ctx *Context) error {
		return ChoosePermanents{
			Question: "Kefnet the Mindful — you may return a land you control to its owner's hand",
			Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
				var out []uuid.UUID
				for _, id := range PermanentsControlledBy(g, of) {
					if c, ok := g.LookupCardForEffect(id); ok && c.IsLand() {
						out = append(out, id)
					}
				}
				return out, 0, 1
			},
			Keyed: cmdr2S22LandReturned,
		}.Apply(ctx)
	}))(g, item)
}

// cmdr2S22LandReturned is the chosen land returning to its owner's hand,
// keyed (#4290). An on-disk key.
var cmdr2S22LandReturned = RegisterPermanentsPick("prompt-run/cmdr2-s22-kefnet-return-land", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !onBattlefield(ctx.Game, id) {
				continue
			}
			if err := (BounceToHand{Target: id}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	},
})

// cmdr2S22DrewYourFifthCardThisTurn is "whenever you draw your fifth
// card each turn": the turn's draw tally read after the draw, so it
// fires on the fifth draw and on no other (b41DrewYourSecondCardThisTurn's
// shape).
func cmdr2S22DrewYourFifthCardThisTurn(ev game.Event, source *game.Card, g *game.Game) bool {
	if ev.Kind != game.EventDrawCard || ev.Actor != source.Controller {
		return false
	}
	return g.TurnTallyFor(source.Controller).CardsDrawn == 5
}

// cmdr2S22AlandraAndDrakesGetCardsInHand is "Alandra and Drakes you
// control each get +X/+X until end of turn, where X is the number of
// cards in your hand": X is read as the trigger resolves and the set of
// creatures is locked then (CR 611.2c).
func cmdr2S22AlandraAndDrakesGetCardsInHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := cardsInHand(g, item.Controller)
	self := item.SourceCardID
	return untilEndOfTurn(ctx, uuid.Nil,
		And(YouControl(), Or(Subtype("Drake"), func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.InstanceID == self })),
		"Alandra, Sky Dreamer — Alandra and Drakes you control each get +X/+X until end of turn",
		game.ModifyPTMod(x, x))
}

// cmdr2S22CreaturesBecomeXByXForCardsInHand is "Until end of turn,
// creatures you control have base power and toughness X/X, where X is the
// number of cards in your hand" (Jolrael, Mwonvuli Recluse), Mirror
// Entity's layer 7b set with X read as the ability resolves.
func cmdr2S22CreaturesBecomeXByXForCardsInHand(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := cardsInHand(g, item.Controller)
	return untilEndOfTurn(ctx, uuid.Nil, And(Creature(), YouControl()),
		"Jolrael, Mwonvuli Recluse — base power and toughness X/X",
		game.SetBasePowerMod(x), game.SetBaseToughnessMod(x))
}

// cmdr2S22PlusXForCardsDrawnThisTurn is Knowledge Is Power's anthem:
// "Creatures you control get +X/+X, where X is the number of cards you've
// drawn this turn". Layer 7c over the controller's creatures, the tally
// read on every pass; DependsOnEvents{EventDrawCard} drops the layer
// cache on a draw (Duelist of the Mind's hint), since a draw moves
// nothing on the battlefield.
func cmdr2S22PlusXForCardsDrawnThisTurn() game.StaticAbility {
	return game.StaticAbility{
		Label:           "Creatures you control get +X/+X, where X is the number of cards you've drawn this turn.",
		Layer:           game.Layer7PT,
		SubLayer:        game.SubLayer7C_Modify,
		AppliesTo:       cmdr2S22YourCreatures,
		DependsOnEvents: []game.EventKind{game.EventDrawCard},
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			x := g.TurnTallyFor(source.Controller).CardsDrawn
			c.Power += x
			c.Toughness += x
		},
	}
}

// cmdr2S22YourCreatures is "creatures you control" for a static's
// AppliesTo: the source's controller's creatures, the source included.
func cmdr2S22YourCreatures(target *game.Card, _ *game.Game, source *game.Card) bool {
	return target.IsCreature() && target.Controller == source.Controller
}

// cmdr2S22CastWhilePowerParity is Kianne, Corrupted Memory's "as long as
// this creature's power is even (odd), you may cast <kind> spells as
// though they had flash": CastAsThoughFlash narrowed to a kind of spell
// and gated on the source's live power parity.
func cmdr2S22CastWhilePowerParity(filter game.PermissionFilter, even bool, label string) game.CastTimingRule {
	rule := CastKindAsThoughFlash(filter, label)
	rule.SourcePowerEven = even
	rule.SourcePowerOdd = !even
	return rule
}

// cmdr2S22CastDuringOpponentsEndStep is "During each opponent's end
// step, you may cast spells as though they had flash" (Final-Word
// Phantom).
func cmdr2S22CastDuringOpponentsEndStep(label string) game.CastTimingRule {
	rule := CastAsThoughFlash()
	rule.Label = label
	rule.DuringOpponentsEndStep = true
	return rule
}

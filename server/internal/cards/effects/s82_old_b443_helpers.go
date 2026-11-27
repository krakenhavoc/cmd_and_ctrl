package effects

import (
	"slices"
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dark Ascension batch 443 helpers.

// OldB443LoyalCatharReturnBody is the delayed return of Loyal Cathar: "return
// it to the battlefield transformed under your control at the beginning of
// the next end step". The card is the graveyard object it became as it died
// (CR 400.7): one that has left the graveyard since stays where it went. An
// on-disk key: never renamed, never reused.
var OldB443LoyalCatharReturnBody = game.SimpleDelayedBody(
	"delayed/old-b443-return-this-card-transformed", func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range item.Targets {
			if t.Kind != game.TargetCard || t.ID == uuid.Nil || !inGraveyardOneMoveAfter(g, t.ID, item.Params.Object.Epoch) {
				continue
			}
			if err := (ReturnFromGraveyard{
				Target:      t.ID,
				Dest:        game.ZoneBattlefield,
				Controller:  item.Controller,
				Transformed: true,
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})

// OldB443ReturnThisCardTransformedAtNextEndStep is Loyal Cathar's dies
// trigger: schedule the delayed return of this card (as the graveyard object
// it became) at the beginning of the next end step.
func OldB443ReturnThisCardTransformedAtNextEndStep(label string) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		ref, ok := ctx.SourceRef()
		if !ok || !SourceCardInGraveyardAfterLeaving(ctx) {
			return nil
		}
		return ScheduleDelayedTrigger{
			At:     game.StepEnd,
			Label:  label,
			Cards:  []uuid.UUID{ref.ID},
			Body:   OldB443LoyalCatharReturnBody,
			Params: game.EffectParams{Object: ref},
		}.Apply(ctx)
	}
}

// OldB443SpiritInAdditionExcept is Séance's copy exception "except it's a
// Spirit in addition to its other types" (CR 707.9b): the subtype becomes part
// of the token's copiable values, so a copy of the token keeps it.
func OldB443SpiritInAdditionExcept(t *game.Card) {
	super, types, sub := game.ParseTypeLine(t.TypeLine)
	if !slices.Contains(sub, "Spirit") {
		sub = append(sub, "Spirit")
	}
	head := strings.Join(append(append([]string{}, super...), types...), " ")
	t.TypeLine = head + " — " + strings.Join(sub, " ")
}

// OldB443SeanceTail is Séance's "if you do": the card reached exile, so a
// token that's a copy of it, a Spirit in addition to its other types, is
// created under the Séance's controller and exiled at the beginning of the
// next end step. An on-disk key: never renamed, never reused.
var OldB443SeanceTail = CardTail("then/old-b443-seance-copy", func(ctx *Context, o game.TailOutcome) error {
	if !OneWentThisWay(o) {
		return nil
	}
	return TrekB40CreateCopyExiledAtNextEndStep(ctx, o.Tail.Player, o.IDs[0], OldB443SpiritInAdditionExcept,
		"Séance — exile the token")
})

// OldB443ExileTargetThenSeanceCopy is Séance's resolution: exile the target
// creature card from your graveyard; if you do, the copy.
func OldB443ExileTargetThenSeanceCopy(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := FirstLegalTarget(ctx)
	if !ok {
		return nil
	}
	return ExileTarget{Target: t.ID, ThenTail: ctx.ItemTail(OldB443SeanceTail)}.Apply(ctx)
}

// OldB443LostInTheWoods is Lost in the Woods' resolution: reveal the top
// card of your library; if it's a Forest card, remove the attacking creature
// from combat; then put the revealed card on the bottom of your library.
func OldB443LostInTheWoods(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	me := item.Controller
	top := g.RevealTopOfLibraryForEffect(me, item.SourceCardID, 1, "Lost in the Woods — reveal the top card of your library")
	if len(top) == 0 {
		return nil
	}
	if c, ok := g.LookupCardForEffect(top[0]); ok && c.HasSubtypeWord("Forest") {
		if err := NothingIfGone(g.RemoveFromCombatForEffect(item.Trigger.Event.CardID)); err != nil {
			return err
		}
	}
	return g.PutOnBottomInRandomOrderForEffect(me, game.ZoneLibrary, top)
}

// OldB443NotYourTurn is "during turns other than yours" for a static read as
// the layers are rebuilt.
func OldB443NotYourTurn(g *game.Game, controller uuid.UUID) bool {
	return !IsActivePlayer(g, controller)
}

// OldB443GargoyleOnOthersTurns is Warden of the Wall's "During turns other
// than yours, this artifact is a 2/3 Gargoyle artifact creature with flying":
// a type change (layer 4), a P/T set (layer 7b) and a keyword (layer 6), each
// holding only on turns of other players.
func OldB443GargoyleOnOthersTurns() []game.StaticAbility {
	others := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return SelfOnly(target, g, source) && OldB443NotYourTurn(g, source.Controller)
	}
	return []game.StaticAbility{
		{
			Label:     "During turns other than yours, this artifact is a Gargoyle artifact creature.",
			Layer:     game.Layer4Type,
			AppliesTo: others,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !slices.Contains(c.Types, "Creature") {
					c.Types = append(c.Types, "Creature")
				}
				if !slices.Contains(c.Subtypes, "Gargoyle") {
					c.Subtypes = append(c.Subtypes, "Gargoyle")
				}
			},
		},
		{
			Label:     "During turns other than yours, this artifact is a 2/3 creature.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7B_Set,
			AppliesTo: others,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power, c.Toughness = 2, 3
			},
		},
		{
			Label:     "During turns other than yours, this artifact has flying.",
			Layer:     game.Layer6Ability,
			AppliesTo: others,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				AppendKeywordsTo(c, []string{"flying"})
			},
			Keywords: []string{"flying"},
		},
	}
}

// oldB443CardTypes are the card types a "shares a card type" clause reads
// (CR 205.2a), lowercase.
var OldB443CardTypes = []string{"artifact", "battle", "creature", "enchantment", "instant", "kindred", "land", "planeswalker", "sorcery"}

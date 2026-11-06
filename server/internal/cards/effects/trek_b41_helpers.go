package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// trek_b41_helpers.go — shared bodies for Star Trek Commander batch 41
// (#4974).

// trekB41PumpAndGrantEachTarget is "one or two target creatures each get
// +power/+toughness and gain <keywords> until end of turn": every target
// still legal as the ability resolves gets both (CR 608.2b), and an
// illegal one gets neither.
func trekB41PumpAndGrantEachTarget(power, toughness int, label string, keywords ...string) func(g *game.Game, item *game.StackItem) error {
	return func(g *game.Game, item *game.StackItem) error {
		ctx := NewContext(g, item)
		for _, t := range ctx.LegalTargets() {
			if t.Kind != game.TargetCard {
				continue
			}
			if err := (BoostUntilEOT{Target: t.ID, Power: power, Toughness: toughness, Label: label}).Apply(ctx); err != nil {
				return err
			}
			if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: keywords, Label: label}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}

// trekB41SeekOut is "Seek out — Whenever you draw your second card each
// turn, <effect>": Kid Loki's trigger condition (b41DrewYourSecondCardThisTurn,
// which fires once a turn, on the second draw). Seek out is an ability
// word with no rules meaning.
func trekB41SeekOut(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventDrawCard, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
		return b41DrewYourSecondCardThisTurn(ev, source, g)
	}, label, effect)
}

// trekB41ReturnTheObjectFromGraveyardToItsOwnerBody is "return that card
// to the battlefield under its owner's control at the beginning of the
// next end step" (The Guardian of Forever): the card <Object> names goes
// from a graveyard to the battlefield under its owner's control, if it is
// still that object in a graveyard (CR 400.7). Avacyn, Angel of Horror's
// body with the owner, not the trigger's controller, as the new
// controller. An on-disk key: never renamed, never reused.
var trekB41ReturnTheObjectFromGraveyardToItsOwnerBody = game.DelayedBody("graveyard/return-the-object-under-its-owners-control",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		if !objectStillInAGraveyard(g, p.Object) {
			return nil
		}
		return ReturnFromGraveyard{Target: p.Object.ID, Dest: game.ZoneBattlefield}.Apply(NewContext(g, item))
	})

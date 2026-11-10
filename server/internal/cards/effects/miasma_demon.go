package effects

import (
	"strconv"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Miasma Demon — Creature — Demon {4}{B}{B}, 5/4 (Duskmourn, #4466):
//
//	"Flying
//	 When this creature enters, you may discard any number of cards. When
//	 you do, up to that many target creatures each get -2/-2 until end of
//	 turn."
//
// "Any number" is a real choice from nothing to the whole hand (Steal the
// Show's up-to discard), asked as the trigger resolves. The second sentence
// is a reflexive trigger (CR 603.12, Seasoned Cryomancer's shape) created
// only when at least one card really left the hand; it carries how many in
// Params.Amount and chooses its "up to that many target creatures" as it
// goes on the stack, not when the parent was announced (the ruling). Each
// target still legal as it resolves gets -2/-2 until end of turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b1df6475-e1e8-4ef3-a7b2-919d5f121c57",
		Name:            "Miasma Demon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Miasma Demon — you may discard any number of cards; when you do, up to that many target creatures each get -2/-2 until end of turn",
				miasmaDemonDiscard),
		},
	})
}

// miasmaDemonShrinkBody is the reflexive trigger's registration: "up to
// that many target creatures", the count on Params.Amount, each still-legal
// one getting -2/-2 until end of turn.
var miasmaDemonShrinkBody = game.ReflexiveBody("miasma-demon/shrink", simpleBody(miasmaDemonShrink),
	func(_ uuid.UUID, p game.EffectParams) *game.TargetSpec {
		return TargetCreature("up to "+strconv.Itoa(p.Amount)+" target creatures").WithCount(0, p.Amount)
	})

// miasmaDemonShrink gives each target still legal -2/-2 until end of turn.
func miasmaDemonShrink(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		if err := (BoostUntilEOT{Target: t.ID, Power: -2, Toughness: -2, Label: "Miasma Demon — -2/-2"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

// miasmaDemonDiscard is the enters trigger's body: the optional discard of
// any number of cards.
func miasmaDemonDiscard(g *game.Game, item *game.StackItem) error {
	n := b14HandSize(g, item.Controller)
	if n == 0 {
		return nil
	}
	return g.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
		Player:   item.Controller,
		Source:   item.SourceCardID,
		N:        n,
		UpTo:     true,
		Question: "Miasma Demon — you may discard any number of cards; up to that many target creatures each get -2/-2 until end of turn",
	}, miasmaDemonDiscarded, nil)
}

// miasmaDemonDiscarded is the discard answered, keyed: the reflexive trigger
// exists when at least one card was discarded, and counts them. An on-disk
// key.
var miasmaDemonDiscarded = game.RegisterDiscardThen("choose-cards/miasma-demon-discarded", func(g *game.Game, r game.DiscardsDone) error {
	n := len(r.Discarded.By(r.Player))
	if n == 0 {
		return nil
	}
	return ReflexiveTrigger{
		Label:  "Miasma Demon — up to that many target creatures each get -2/-2 until end of turn",
		Body:   miasmaDemonShrinkBody,
		Params: game.EffectParams{Amount: n},
	}.Apply(keyedContext(g, r.Player, r.Source))
})

package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const torrentSculptorOracleID = "8a9fe2dd-50a7-4010-96ba-df4dec265e16"

// Torrent Sculptor // Flamethrower Sonata — a modal double-faced card, one
// file for both faces (the back under the "#1" key of the MDFC face
// keyspace).
//
// Front, Creature — Merfolk Wizard {2}{U}{U}, 2/2:
//
//	"Ward {2}
//	 When this creature enters, exile an instant or sorcery card from your
//	 graveyard. Put a number of +1/+1 counters on this creature equal to
//	 half that card's mana value, rounded up."
//
// Back, Sorcery {1}{R}:
//
//	"Discard a card, then draw a card. When you discard an instant or
//	 sorcery card this way, Flamethrower Sonata deals damage equal to that
//	 card's mana value to target creature or planeswalker you don't
//	 control."
//
// The front's pick is a keyed prompt; with no instant or sorcery card in the
// graveyard nothing is exiled and no counters go on. The back discards first
// and draws next, and the reflexive "when you do" (CR 603.12) is made when
// the instant or sorcery card is discarded, so its target is chosen as the
// spell finishes resolving and the damage is the discarded card's mana
// value.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        torrentSculptorOracleID,
		Name:            "Torrent Sculptor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"ward 2"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Torrent Sculptor — exile an instant or sorcery card from your graveyard; put counters equal to half its mana value, rounded up",
				effects.S82OldB40PickInstantOrSorcery("Torrent Sculptor")),
		},
	})
	effects.Register(effects.Spec{
		OracleID:     torrentSculptorOracleID + "#1",
		Name:         "Flamethrower Sonata",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
				Player:   item.Controller,
				Source:   item.SourceCardID,
				N:        1,
				Question: "Flamethrower Sonata — discard a card",
			}, s82OldB40SonataDiscarded, nil)
		},
	})
}

var (
	// The discard's answer, keyed: the reflexive trigger is made at the
	// discard, then the draw follows. An on-disk key.
	s82OldB40SonataDiscarded = game.RegisterDiscardThen("choose-cards/s82-old-b40-flamethrower-sonata-discarded", func(g *game.Game, r game.DiscardsDone) error {
		parent := &game.StackItem{Controller: r.Player, SourceCardID: r.Source, Label: "Flamethrower Sonata"}
		ctx := effects.NewContext(g, parent)
		for _, id := range r.Discarded.Cards() {
			c, ok := g.LookupCardForEffect(id)
			if !ok || !(c.IsInstant() || c.IsSorcery()) {
				continue
			}
			mv, _ := g.ManaValueForEffect(c)
			t := effects.WhenYouDo("Flamethrower Sonata — damage equal to the discarded card's mana value to target creature or planeswalker you don't control", s82OldB40SonataDamageBody)
			t.Params = game.EffectParams{Amount: mv}
			if err := t.Apply(ctx); err != nil {
				return err
			}
		}
		return effects.DrawCards{N: 1}.Apply(ctx)
	})

	// "Flamethrower Sonata deals damage equal to that card's mana value to
	// target creature or planeswalker you don't control": the amount rides
	// the body's Params. An on-disk key.
	s82OldB40SonataDamageBody = game.ReflexiveBody("flamethrower-sonata/damage", effects.BreechesBlastEffect,
		effects.ConstTargets(func() *game.TargetSpec {
			return effects.TargetPermanent("target creature or planeswalker you don't control", effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls())
		}))
)

package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kogla and Yidaro — Legendary Creature — Ape Dinosaur Turtle {2}{R}{R}{G}{G},
// 7/7 (March of the Machine Commander):
//
//	"When Kogla and Yidaro enters, choose one —
//	 • It gains trample and haste until end of turn.
//	 • It fights target creature you don't control.
//	 {2}{R}{G}, Discard this card: Destroy up to one target artifact or
//	 enchantment. Shuffle this card into your library from your graveyard, then
//	 draw a card."
//
// The enters trigger is a modal trigger: the fight bullet's target is chosen
// as the trigger goes on the stack and the keyword bullet has none. "It" is
// the source, read by object, so a Kogla that has left fights with its
// last-known power (CR 608.2h).
//
// The activated ability works from the hand (Harvester of Misery's shape, the
// discard is a cost). It has "up to one" target, so it resolves with none; the
// card is shuffled in only while it is still in the graveyard (CR 400.7), and
// the draw follows the shuffle either way.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Kogla and Yidaro — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = ChooseOne(
		ModeDoing("It gains trample and haste until end of turn.", nil,
			func(item *game.StackItem, ctx *Context, _ int) error {
				return GrantKeywordUntilEOT{Target: item.SourceCardID, Keywords: []string{"trample", "haste"},
					Label: "Kogla and Yidaro — trample and haste"}.Apply(ctx)
			}),
		ModeDoing("It fights target creature you don't control.",
			Harmful(TargetCreature("target creature you don't control", OpponentControls())),
			ThisFightsTheModesTarget),
	)
	Register(Spec{
		OracleID:     "fe8c2b7e-f80d-4acb-a976-edcf436a78a9",
		Name:         "Kogla and Yidaro",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
		Activated: []ActivatedAbility{{
			Label:   "{2}{R}{G}, Discard this card: Destroy up to one target artifact or enchantment. Shuffle this card into your library from your graveyard, then draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{2}{R}{G}"), DiscardThis()),
			Zones:   []game.ZoneKind{game.ZoneHand},
			Targets: Harmful(TargetPermanent("up to one target artifact or enchantment",
				Or(Artifact(), Enchantment())).WithCount(0, 1)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return koglaShuffleThenDraw(g, item)
			},
		}},
	})
}

// koglaShuffleThenDraw is "shuffle this card into your library from your
// graveyard, then draw a card".
func koglaShuffleThenDraw(g *game.Game, item *game.StackItem) error {
	id := item.SourceCardID
	owner := item.Controller
	if c, ok := g.LookupCardForEffect(id); ok && c.Owner != uuid.Nil {
		owner = c.Owner
	}
	tail := game.KeyedTail{Ref: koglaShuffleDrawTail, Player: item.Controller, Carry: []uuid.UUID{owner}}
	if !thisCardStillInTheGraveyard(g, item) {
		return tail.Run(g, nil)
	}
	return g.TuckCardsToLibraryKeyedForEffect([]uuid.UUID{id}, game.TuckOptions{}, tail)
}

// koglaShuffleDrawTail shuffles the owner's library and the controller draws.
// An on-disk key: never renamed, never reused.
var koglaShuffleDrawTail = game.RegisterTail("then/kogla-and-yidaro-shuffle-draw", func(g *game.Game, o game.TailOutcome) error {
	if len(o.Tail.Carry) > 0 {
		if err := g.ShuffleLibraryForEffect(o.Tail.Carry[0]); err != nil {
			return err
		}
	}
	return g.DrawNForEffect(o.Tail.Player, 1)
})

package cardsbu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cabal Therapist — Creature — Horror {B}, 1/1 (Modern Horizons):
//
//	"Menace
//	 At the beginning of your first main phase, you may sacrifice a creature.
//	 When you do, choose a nonland card name, then target player reveals their
//	 hand and discards all cards with that name."
//
// Two printed triggers: the first-main-phase one, whose "you may sacrifice a
// creature" is asked as it resolves (MaySacrificeOne); and the reflexive one it
// creates only if a creature really was sacrificed (CR 603.12), whose target
// player is chosen as that trigger goes on the stack. The name is asked as the
// reflexive trigger resolves, for any card name (CR 201.2: free text), through
// the keyed card-name prompt; the player reveals their hand and discards every
// card with that name (CR 201.2b: every face is checked), all at once and not by
// their choice. A land is never "a nonland card name", so a land named does
// nothing: nothing there can be discarded for it.
//
// No simplification.
const cabalTherapistLabel = "Cabal Therapist — you may sacrifice a creature; when you do, choose a nonland card name, then target player reveals their hand and discards all cards with that name"

var cabalTherapistSacrificed = effects.RegisterSacrificedOne("prompt-run/s82-old-b166-cabal-therapist-sacrifice", func(ctx *effects.Context, _ uuid.UUID) error {
	return effects.WhenYouDo("Cabal Therapist — choose a nonland card name; target player reveals their hand and discards all cards with that name", cabalTherapistBody).Apply(ctx)
})

var cabalTherapistBody = game.ReflexiveBody("s82-old-b166/cabal-therapist-name", effects.SimpleBody(func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	t, ok := effects.FirstLegalTarget(ctx)
	if !ok || t.Kind != game.TargetPlayer {
		return nil
	}
	g.QueueCardNameChoiceKeyedForEffect(ctx.Controller(), ctx.Source(),
		"Cabal Therapist — choose a nonland card name",
		effects.WithCarry(ctx.ItemTail(cabalTherapistNamed), t.ID))
	return nil
}), effects.ConstTargets(func() *game.TargetSpec { return effects.TargetPlayer("target player") }))

// cabalTherapistNamed is the reveal and the discard once the name (the tail's
// Word) is known; the player is the tail's Carry. An on-disk key.
var cabalTherapistNamed = effects.CardTail("then/s82-old-b166-cabal-therapist-named", func(ctx *effects.Context, o game.TailOutcome) error {
	victim, name := effects.TailCarry0(o), o.Tail.Word
	if victim == uuid.Nil {
		return nil
	}
	p := ctx.Game.PlayerByIDForEffect(victim)
	if p == nil {
		return nil
	}
	ctx.Game.RevealHandForEffect(victim)
	var named []uuid.UUID
	for _, c := range p.Hand.Cards {
		if !c.IsLand() && game.CardNameMatches(c, name) {
			named = append(named, c.InstanceID)
		}
	}
	if len(named) == 0 {
		return nil
	}
	return ctx.Game.DiscardCardsKeyedForEffect(victim, named, ctx.Source(), game.KeyedTail{})
})

func init() {
	effects.Register(effects.Spec{
		OracleID:        "9b7870df-6bca-499f-bc6e-e57f2ddfe640",
		Name:            "Cabal Therapist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered: []game.TriggeredAbility{
			effects.AtYourPrecombatMain(cabalTherapistLabel, func(g *game.Game, item *game.StackItem) error {
				return effects.MaySacrificeOne(g, item, "Cabal Therapist — sacrifice a creature?", effects.Creature(), false, cabalTherapistSacrificed)
			}),
		},
	})
}

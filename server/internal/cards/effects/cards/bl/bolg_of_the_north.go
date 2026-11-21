package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bolg of the North — Legendary Creature — Goblin Soldier {3}{B}{R}, 5/5
// (The Hobbit, HOB):
//
//	"When Bolg enters, you may sacrifice another creature. When you do,
//	 Bolg deals damage equal to that creature's power to another target
//	 creature. If excess damage was dealt this way, amass Goblins X, where
//	 X is that excess damage."
//
// Ziatora's shape, which this card follows: the enters trigger asks "you may
// sacrifice another creature" as it resolves (maySacrificeOne, CR 608.2c),
// and a creature that really went raises the printed "when you do" as a
// CR 603.12 reflexive trigger. That trigger goes on the stack above the
// parent and chooses its own "another target creature" as it does (so the
// target is picked after the sacrifice, and Bolg himself is not a legal one),
// carrying the sacrificed creature as its payload.
//
// On resolution Bolg deals damage equal to the sacrificed creature's power as
// it last stood (CR 608.2h, so an anthem or a counter counts) to the chosen
// creature, and the excess (CR 120.10: what landed beyond the damage the
// creature still needed, read after the damage so a prevention shield lowers
// it) is how many +1/+1 counters the Goblin amass puts on an Army. With no
// excess nothing is amassed; the amass is the card's own sentence, so it
// happens even when Bolg has left the battlefield by then, and a target that
// is gone leaves nothing dealt and nothing amassed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "88522a0f-5377-4522-97f4-4148bef954af",
		Name:         "Bolg of the North",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Bolg of the North — you may sacrifice another creature", bolgOfTheNorthSacrifice),
		},
	})
}

// bolgOfTheNorthSacrifice is the enters trigger's body: the "may" prompt.
func bolgOfTheNorthSacrifice(g *game.Game, item *game.StackItem) error {
	return effects.MaySacrificeOne(g, item, "Bolg of the North — sacrifice another creature to deal damage equal to its power to another target creature?", effects.Creature(), true, bolgOfTheNorthSacrificed)
}

// bolgOfTheNorthSacrificed raises the reflexive trigger once a creature has
// really gone, keyed (#4290).
var bolgOfTheNorthSacrificed = effects.RegisterSacrificedOne("prompt-run/bolg-of-the-north-sacrifice", func(ctx *effects.Context, sacrificed uuid.UUID) error {
	return effects.ReflexiveTrigger{
		Label: "Bolg of the North — damage equal to the sacrificed creature's power to another target creature",
		Cards: []uuid.UUID{sacrificed},
		Body:  bolgOfTheNorthDamageBody,
	}.Apply(ctx)
})

// bolgOfTheNorthDamageBody is the reflexive trigger's registration: the
// sacrificed creature is the payload, the victim the trigger's own clause,
// which is any creature but Bolg.
var bolgOfTheNorthDamageBody = game.ReflexiveBody("bolg-of-the-north/damage", effects.SimpleBody(bolgOfTheNorthDamage),
	func(sourceID uuid.UUID, _ game.EffectParams) *game.TargetSpec {
		return effects.TargetCreature("another target creature", effects.OtherThan(sourceID))
	})

// bolgOfTheNorthDamage deals the damage, then amasses Goblins by the excess.
func bolgOfTheNorthDamage(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	sacrificed := ctx.PayloadCards()
	if len(sacrificed) == 0 {
		return nil
	}
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	excess, err := effects.B27DealDamageWithExcess(ctx, targets[0].ID, effects.B29PowerAsItLastStood(g, sacrificed[0]))
	if err != nil || excess <= 0 {
		return err
	}
	return effects.Amass{Subtype: "Goblin", N: excess}.Apply(ctx)
}

package cardsd

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Dack Fayden — Legendary Planeswalker — Dack {1}{U}{R}, loyalty 3:
//
//	"+1: Target player draws two cards, then discards two cards.
//	 −2: Gain control of target artifact.
//	 −6: You get an emblem with "Whenever you cast a spell that targets one or
//	 more permanents, gain control of those permanents.""
//
// The +1 is Cephalid Broker's loot for the target player: both draws, then the
// discard prompt over the hand they made. The −2 takes the artifact for good
// (no stated duration, CR 611.2b). The emblem's trigger fires as the spell is
// cast (CR 601.2i), when its targets are already chosen (CR 601.2c): it
// reads the spell's targets on the stack, the ones that are permanents, and
// takes control of each that is still on the battlefield as it resolves; the
// spell resolves after it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9d4b1e44-3c66-4b1b-8a1d-308aa4f841f8",
		Name:            "Dack Fayden",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &effects.EmblemSpec{
			Label: "Dack Fayden emblem",
			Text:  "Whenever you cast a spell that targets one or more permanents, gain control of those permanents.",
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventCast, dackFaydenEmblemCast,
					"Dack Fayden emblem — gain control of the permanents the spell targets", dackFaydenEmblemSteal),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Target player draws two cards, then discards two cards.",
				Purpose: effects.ForTargets(game.TargetPurpose{Slot: 0, Draws: 2, Discards: 2}),
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetPlayer("target player"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					p, ok := effects.FirstLegalPlayerTarget(effects.NewContext(g, item))
					if !ok {
						return nil
					}
					return effects.DrawThenDiscard(g, p, item.SourceCardID, 2, 2, "")
				},
			},
			{
				Label:   "−2: Gain control of target artifact.",
				Cost:    effects.LoyaltyCost(-2),
				Targets: effects.TargetPermanent("target artifact", effects.Artifact()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					return effects.GainControl{Target: id, Controller: item.Controller, Duration: game.IndefiniteDuration(),
						Label: "Dack Fayden — gain control (no stated duration)"}.Apply(ctx)
				},
			},
			{
				Label: "−6: You get an emblem with \"Whenever you cast a spell that targets one or more permanents, " +
					"gain control of those permanents.\"",
				Cost: effects.LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.CreateEmblem{}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}

// dackFaydenSpellPermanentTargets is the battlefield permanents a spell on the
// stack targets.
func dackFaydenSpellPermanentTargets(g *game.Game, spell uuid.UUID) []uuid.UUID {
	item := g.StackItemForEffect(spell)
	if item == nil || item.Kind != game.StackItemSpell {
		return nil
	}
	var ids []uuid.UUID
	for _, t := range item.Targets {
		if t.Kind == game.TargetCard && t.ID != uuid.Nil && g.Battlefield.Contains(t.ID) {
			ids = append(ids, t.ID)
		}
	}
	return ids
}

// dackFaydenEmblemCast is "whenever you cast a spell that targets one or more
// permanents": the emblem's owner cast it and one of its targets is a permanent.
func dackFaydenEmblemCast(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return source != nil && ev.Actor == source.Controller && len(dackFaydenSpellPermanentTargets(g, ev.CardID)) > 0
}

// dackFaydenEmblemSteal takes control of those permanents, through the
// trigger's event (the spell).
func dackFaydenEmblemSteal(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	for _, id := range dackFaydenSpellPermanentTargets(g, item.Trigger.Event.CardID) {
		if err := (effects.GainControl{Target: id, Controller: item.Controller, Duration: game.IndefiniteDuration(),
			Label: "Dack Fayden emblem — gain control (no stated duration)"}).Apply(ctx); err != nil {
			return err
		}
	}
	return nil
}

package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jace, Vryn's Prodigy // Jace, Telepath Unbound — Magic Origins (S82 older
// sets), a transforming legendary creature whose back face is a planeswalker.
//
// Front face, Legendary Creature — Human Wizard {1}{U}, 0/2:
//
//	"{T}: Draw a card, then discard a card. If there are five or more cards in
//	 your graveyard, exile Jace, then return him to the battlefield transformed
//	 under his owner's control."
//
// Back face, Legendary Planeswalker — Jace, loyalty 5:
//
//	"+1: Up to one target creature gets -2/-0 until your next turn.
//	 −3: You may cast target instant or sorcery card from your graveyard this
//	     turn. If that spell would be put into your graveyard, exile it instead.
//	 −9: You get an emblem with "Whenever you cast a spell, target opponent mills
//	     five cards.""
//
// The back face registers under "<oracle_id>#1". The loot waits for its draw
// (a draw can pause on a prompt) and then asks for the discard; the graveyard is
// counted once the discard is done, and the flip is skipped for a Jace that is
// no longer the permanent that activated the ability (CR 400.7). The +1 is a
// layer 7c modifier lasting until the controller's next turn (CR 611.2b). The −3
// is Vohar's grant: a one-card permission for this turn, at the card's mana
// cost and normal timing, exiled instead of going to the graveyard. The emblem
// is a cast trigger with a target opponent chosen as it goes on the stack.
//
// No simplification.
func init() {
	const oracle = "594f6881-c059-46f8-aa4e-7151d502de73"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Jace, Vryn's Prodigy",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: Draw a card, then discard a card. If there are five or more cards in your graveyard, exile Jace, then return him to the battlefield transformed under his owner's control.",
			Purpose: game.Purpose{Draws: 1, Discards: 1, Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect:  effects.OldB322JaceLoot,
		}},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Jace, Telepath Unbound",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &effects.EmblemSpec{
			Label: "Jace, Telepath Unbound emblem",
			Text:  "Whenever you cast a spell, target opponent mills five cards.",
			Triggered: []game.TriggeredAbility{
				effects.Targeting(
					effects.WheneverYouCast(nil, "Jace, Telepath Unbound emblem — target opponent mills five cards", effects.TargetPlayerMillsEffect(5)),
					effects.TargetPlayer("target opponent", effects.Opponent())),
			},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Up to one target creature gets -2/-0 until your next turn.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						return effects.ScopedEffectFor{
							Target:   t.ID,
							Mods:     []game.Mod{game.ModifyPTMod(-2, 0)},
							Duration: effects.DurationUntilYourNextTurn(ctx, ctx.Controller()),
							Label:    "Jace, Telepath Unbound — -2/-0 until your next turn",
						}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "−3: You may cast target instant or sorcery card from your graveyard this turn. If that spell would be put into your graveyard, exile it instead.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCardInGraveyard("target instant or sorcery card from your graveyard", effects.Or(effects.Instant(), effects.Sorcery()), effects.YouOwn()),
				Effect:  effects.OldB322MayCastFromGraveyard("Jace, Telepath Unbound"),
			},
			{
				Label:  "−9: You get an emblem with \"Whenever you cast a spell, target opponent mills five cards.\"",
				Cost:   effects.LoyaltyCost(-9),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

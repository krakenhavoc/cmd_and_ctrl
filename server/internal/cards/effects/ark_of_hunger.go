package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ark of Hunger — Artifact {2}{R}{W} (Secrets of Strixhaven):
//
//	"Whenever one or more cards leave your graveyard, this artifact deals
//	 1 damage to each opponent and you gain 1 life.
//	 {T}: Mill a card. You may play that card this turn."
//
// The trigger is a batch trigger over cards of the controller's that leave
// a graveyard, by a zone move or by being cast from it
// (socB20CardLeftYourGraveyard): cards that leave as one event trigger it
// once. Its body is Arbaaz Mir's "damage to each opponent, you gain life".
//
// The mill is a MillToZone whose keyed tail hands the milled card a play
// permission out of the graveyard until end of turn (a land is played, a
// spell cast). A card a replacement sent elsewhere (Rest in Peace) was not
// milled, is not in the tail's list, and gets no permission.
func init() {
	Register(Spec{
		OracleID:     "2c654772-c786-4966-a46e-9556bbf31d7c",
		Name:         "Ark of Hunger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny([]game.EventKind{game.EventZoneMove, game.EventCast}, socB20CardLeftYourGraveyard,
				"Ark of Hunger — 1 damage to each opponent and you gain 1 life",
				b34DamageEachOpponentAndGainLife(1, 1))),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Mill a card. You may play that card this turn.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return MillToZone{
					Player:   item.Controller,
					N:        1,
					ThenTail: ctx.ItemTail(arkOfHungerMilled),
				}.Apply(ctx)
			},
		}},
	})
}

// arkOfHungerMilled is "you may play that card this turn" over the card
// that landed in the graveyard. An on-disk key.
var arkOfHungerMilled = CardTail("ark-of-hunger/play-the-milled-card", func(ctx *Context, o game.TailOutcome) error {
	for _, id := range o.IDs {
		z := ctx.Game.FindCardZoneForEffect(id)
		c, ok := ctx.Game.LookupCardForEffect(id)
		if !ok || z == nil || z.Kind != game.ZoneGraveyard {
			continue
		}
		ctx.Game.GrantCastPermissionToCardsForEffect(game.CastPermission{
			Player:     ctx.Controller(),
			Zone:       game.ZoneGraveyard,
			Duration:   ctx.Game.UntilEndOfTurnDuration(),
			Source:     ctx.Source(),
			SourceName: "Ark of Hunger",
			Label:      "Ark of Hunger — play the milled card this turn",
		}, []game.Card{c})
	}
	return nil
})

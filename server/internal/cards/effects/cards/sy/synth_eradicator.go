package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Synth Eradicator — Artifact Creature — Synth Soldier {2}{R}, 3/3:
//
//	"Haste
//	 Whenever this creature attacks, exile the top card of your library.
//	 You may get {E}{E} (two energy counters). If you don't, you may play
//	 that card this turn.
//	 {T}, Pay {E}{E}{E}: This creature deals 3 damage to any target."
//
// The attack trigger exiles the top card face up through the shared
// exit primitive (a commander is offered the command zone, CR 903.9a),
// then asks the one question the card prints: take the two energy, or
// don't — and only on "don't" is the card granted the play permission
// for this turn (CR 611.2c, one card). An empty library exiles nothing
// and the choice is still asked, since "you may get {E}{E}" does not
// depend on it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c68f09d8-a1eb-449b-8bf5-4251d7a14337",
		Name:            "Synth Eradicator",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Synth Eradicator — exile the top card of your library", synthEradicatorAttack),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Pay {E}{E}{E}: This creature deals 3 damage to any target.",
			Cost:    effects.Plus(effects.TapCost(), effects.PayEnergy(3)),
			Targets: effects.TargetAny(),
			Purpose: effects.ForTargets(effects.DamageToTarget(0, 3)),
			Effect:  effects.SourceDealsDamageToEachLegalTarget(3),
		}},
	})
}

// synthEradicatorAttack exiles the top card, then asks.
func synthEradicatorAttack(g *game.Game, item *game.StackItem) error {
	controller := item.Controller
	p := g.PlayerByIDForEffect(controller)
	if p == nil {
		return nil
	}
	ctx := effects.NewContext(g, item)
	if p.Library == nil || len(p.Library.Cards) == 0 {
		return synthEradicatorAsk(ctx, uuid.Nil)
	}
	top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
	return g.ExileCardKeyedForEffect(top, ctx.ItemTail(synthEradicatorExiled))
}

// synthEradicatorAsk is the choice over the card exiled, if any.
func synthEradicatorAsk(ctx *effects.Context, exiled uuid.UUID) error {
	return effects.MayChoice{
		Question: "Synth Eradicator — get {E}{E}? If you don't, you may play the exiled card this turn.",
		YesLabel: "Get {E}{E}",
		NoLabel:  "Play the exiled card this turn",
		Then:     synthEradicatorAnswer,
		Carry:    []uuid.UUID{exiled},
	}.Apply(ctx)
}

// synthEradicatorExiled asks once the exile has settled, told the card
// when it reached exile, keyed (#4253 slice 10a).
var synthEradicatorExiled = effects.CardTail("then/synth-eradicator-ask", func(ctx *effects.Context, o game.TailOutcome) error {
	exiled := uuid.Nil
	if effects.OneWentThisWay(o) {
		exiled = o.IDs[0]
	}
	return synthEradicatorAsk(ctx, exiled)
})

// synthEradicatorAnswer is the choice after the exile: get {E}{E}, or the
// exiled card (the carried ID, if any) may be played this turn. Keyed
// (#4291).
var synthEradicatorAnswer = effects.MayYesNo("confirm/synth-eradicator-energy-or-play",
	func(ctx *effects.Context, _ game.ConfirmAnswer) error { return effects.GetEnergy{N: 2}.Apply(ctx) },
	func(ctx *effects.Context, r game.ConfirmAnswer) error {
		if exiled := effects.CarryAt(r, 0); exiled != uuid.Nil {
			ctx.Game.GrantCastPermissionOverCardForEffect(exiled, game.CastPermission{Player: r.Chooser})
		}
		return nil
	})

package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "c68f09d8-a1eb-449b-8bf5-4251d7a14337",
		Name:            "Synth Eradicator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Synth Eradicator — exile the top card of your library", synthEradicatorAttack),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Pay {E}{E}{E}: This creature deals 3 damage to any target.",
			Cost:    Plus(TapCost(), PayEnergy(3)),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 3)),
			Effect:  sourceDealsDamageToEachLegalTarget(3),
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
	ctx := NewContext(g, item)
	if p.Library == nil || len(p.Library.Cards) == 0 {
		return synthEradicatorAsk(ctx, uuid.Nil)
	}
	top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
	return g.ExileCardKeyedForEffect(top, ctx.ItemTail(synthEradicatorExiled))
}

// synthEradicatorAsk is the choice over the card exiled, if any.
func synthEradicatorAsk(ctx *Context, exiled uuid.UUID) error {
	return MayChoice{
		Question: "Synth Eradicator — get {E}{E}? If you don't, you may play the exiled card this turn.",
		YesLabel: "Get {E}{E}",
		NoLabel:  "Play the exiled card this turn",
		Then:     synthEradicatorAnswer,
		Carry:    []uuid.UUID{exiled},
	}.Apply(ctx)
}

// synthEradicatorExiled asks once the exile has settled, told the card
// when it reached exile, keyed (#4253 slice 10a).
var synthEradicatorExiled = CardTail("then/synth-eradicator-ask", func(ctx *Context, o game.TailOutcome) error {
	exiled := uuid.Nil
	if oneWentThisWay(o) {
		exiled = o.IDs[0]
	}
	return synthEradicatorAsk(ctx, exiled)
})

// synthEradicatorAnswer is the choice after the exile: get {E}{E}, or the
// exiled card (the carried ID, if any) may be played this turn. Keyed
// (#4291).
var synthEradicatorAnswer = MayYesNo("confirm/synth-eradicator-energy-or-play",
	func(ctx *Context, _ game.ConfirmAnswer) error { return GetEnergy{N: 2}.Apply(ctx) },
	func(ctx *Context, r game.ConfirmAnswer) error {
		if exiled := carryAt(r, 0); exiled != uuid.Nil {
			ctx.Game.GrantCastPermissionOverCardForEffect(exiled, game.CastPermission{Player: r.Chooser})
		}
		return nil
	})

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
	ask := func(g *game.Game, exiled uuid.UUID) error {
		return MayChoice{
			Question: "Synth Eradicator — get {E}{E}? If you don't, you may play the exiled card this turn.",
			YesLabel: "Get {E}{E}",
			NoLabel:  "Play the exiled card this turn",
			OnYes: func(ctx *Context) error {
				return GetEnergy{N: 2}.Apply(ctx)
			},
			OnNo: func(ctx *Context) error {
				if exiled != uuid.Nil {
					ctx.Game.GrantCastPermissionOverCardForEffect(exiled, game.CastPermission{Player: controller})
				}
				return nil
			},
		}.Apply(NewContext(g, item))
	}
	if p.Library == nil || len(p.Library.Cards) == 0 {
		return ask(g, uuid.Nil)
	}
	top := p.Library.Cards[len(p.Library.Cards)-1].InstanceID
	return g.ExileCardThenForEffect(top, func(g *game.Game, landed bool) error {
		if !landed {
			return ask(g, uuid.Nil)
		}
		return ask(g, top)
	})
}

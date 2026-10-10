package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Codie, Ravenous Codex — Legendary Artifact Creature — Book Construct
// {3}, 1/4:
//
//	"Whenever you cast a prepared spell, copy it. You may choose new
//	 targets for the copy.
//	 {W}{U}{B}{R}{G}, {T}: Each creature you control becomes prepared.
//	 (Only creatures with prepare spells can become prepared.)"
//
// A "prepared spell" is the copy of a prepare spell cast out of exile
// (CR 722.3c), see castAPreparedSpell. The copy Codie makes is a
// copy of a spell and is not cast, so it neither unprepares anything nor
// triggers Codie again. The second ability is BecomePrepared on every
// creature the activator controls as it resolves; a creature with no
// prepare spell, or one that is already prepared, stays as it is.
//
// No simplification.
func init() {
	const label = "Codie, Ravenous Codex — copy that prepared spell"
	Register(Spec{
		OracleID:     "a17e414c-8de0-47fd-bcbc-540c7fec7642",
		Name:         "Codie, Ravenous Codex",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventCast && ev.Actor == source.Controller && castAPreparedSpell(g, ev.CardID)
			}, label, func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				return CopySpell{
					StackID:          ctx.Trigger().Event.CardID,
					Controller:       item.Controller,
					ChooseNewTargets: true,
				}.Apply(ctx)
			}),
		},
		Activated: []ActivatedAbility{{
			Label:   "{W}{U}{B}{R}{G}, {T}: Each creature you control becomes prepared.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    Plus(ManaCost("{W}{U}{B}{R}{G}"), TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				for _, c := range g.BattlefieldCardsForEffect() {
					if c.Controller != item.Controller || !c.IsCreature() {
						continue
					}
					if err := (BecomePrepared{Target: c.InstanceID}).Apply(ctx.asGroupMember()); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

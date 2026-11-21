package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fishing Pole — Artifact — Equipment {1} (Foundations, #4465):
//
//	"Equipped creature has "{1}, {T}, Tap Fishing Pole: Put a bait counter
//	 on Fishing Pole."
//	 Whenever equipped creature becomes untapped, remove a bait counter from
//	 this Equipment. If you do, create a 1/1 blue Fish creature token.
//	 Equip {2}"
//
// The granted ability is The Dominion Bracelet's kind of bundle, with a cost
// that taps the GRANTOR (TapTheGrantor, game.AbilityCost.TapGrantor): the
// creature's own {T} taps the creature, and the Equipment must be untapped
// to be tapped as well (CR 118.3). Its counter goes on the specific
// Equipment that granted the row (Context.Grantor), never on another
// permanent named Fishing Pole (the printed ruling).
//
// The trigger is the Equipment's own, watching the per-permanent untap event
// (Mesmeric Orb) of the creature it is attached to. The removal is the
// "if you do": with no bait counter on the Equipment, no Fish is made.
//
// No simplification.
func init() {
	const grant = "fishing-pole/bait"
	const label = "Fishing Pole — remove a bait counter from this Equipment; if you do, create a 1/1 blue Fish creature token"
	effects.Register(effects.Spec{
		OracleID:     "2c1d3e9e-213e-492a-ad5a-91139d70bc0f",
		Name:         "Fishing Pole",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: grant,
			Activated: []effects.ActivatedAbility{{
				Label:   "{1}, {T}, Tap Fishing Pole: Put a bait counter on Fishing Pole.",
				Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.TapTheGrantor()),
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					pole := ctx.Grantor()
					if !effects.OnBattlefield(g, pole) {
						return nil
					}
					return effects.AddCounter{Target: pole, Kind: "bait", N: 1}.Apply(ctx)
				},
			}},
			Text: "{1}, {T}, Tap Fishing Pole: Put a bait counter on Fishing Pole.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(grant)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventUntapCard, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventUntapCard && source.IsAttachedTo(ev.CardID)
			}, label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				pole, ok := g.LookupCardForEffect(item.SourceCardID)
				if !ok || !effects.OnBattlefield(g, item.SourceCardID) || pole.Counters["bait"] <= 0 {
					return nil
				}
				if err := (effects.AddCounter{Target: item.SourceCardID, Kind: "bait", N: -1}).Apply(ctx); err != nil {
					return err
				}
				return effects.CreateToken{Template: effects.TokenCard("1/1 blue Fish"), N: 1}.Apply(ctx)
			}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}

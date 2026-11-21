package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sigarda's Vanguard — Creature — Angel {4}{W}, 3/3 (Midnight Hunt
// Commander, S82 MID batch 13):
//
//	"Flash
//	 Flying
//	 Whenever this creature enters or attacks, choose any number of
//	 creatures with different powers. Those creatures gain double strike
//	 until end of turn."
//
// One trigger watching two events. The choice is made as the ability
// resolves (CR 608.2c) over every creature on the battlefield, whoever
// controls it, through the layers; a keyed choose_cards prompt with the
// set rule "different powers" (midB13DifferentPowers) refuses two creatures
// of one power, and none is a legal answer. Each chosen creature that is
// still on the battlefield gains double strike until end of turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "147d8a1f-fe82-4078-a281-29a2eed4c750",
		Name:            "Sigarda's Vanguard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.OnAny([]game.EventKind{game.EventETB, game.EventAttack},
				func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return effects.B21SelfEnteredOrAttacked(ev, source)
				},
				"Sigarda's Vanguard — choose any number of creatures with different powers; they gain double strike until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					creatures := effects.MidB13CreaturesOnBattlefield(g)
					if len(creatures) == 0 {
						return nil
					}
					effects.MidB13ChooseDifferentPowers(ctx, creatures, 0, len(creatures),
						"Sigarda's Vanguard — choose any number of creatures with different powers; they gain double strike until end of turn",
						sigardasVanguardPick)
					return nil
				}),
		},
	})
}

// sigardasVanguardPick gives each chosen creature double strike until end of
// turn. An on-disk key.
var sigardasVanguardPick = effects.AskedPickRule("choose-cards/sigardas-vanguard-double-strike", effects.MidB13DifferentPowers,
	func(ctx *effects.Context, r game.CardsChosen) error {
		for _, id := range r.Picked {
			if !effects.OnBattlefield(ctx.Game, id) {
				continue
			}
			if err := (effects.ScopedEffectFor{
				Target:   id,
				Mods:     []game.Mod{game.AddKeywordsMod("double strike")},
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Sigarda's Vanguard — double strike until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	})

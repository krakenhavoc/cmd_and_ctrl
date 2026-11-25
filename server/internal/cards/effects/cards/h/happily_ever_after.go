package cardsh

import (
	"strings"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Happily Ever After — Enchantment {2}{W} (Secret Lair Drop):
//
//	"When this enchantment enters, each player gains 5 life and draws a card.
//	 At the beginning of your upkeep, if there are five colors among
//	 permanents you control, there are six or more card types among permanents
//	 you control and/or cards in your graveyard, and your life total is greater
//	 than or equal to your starting life total, you win the game."
//
// The entry gains and draws for each player in turn order from the active
// player. The win is one intervening "if" (CR 603.4) of three conditions, all
// checked as the upkeep begins and again as the trigger resolves. Colours are
// counted among the permanents (their effective colours, so a five-colour
// commander alone is enough); card types are the distinct types among those
// permanents (through the layers) and the printed types of the cards in your
// graveyard, so a type present in both counts once; the starting life total is
// the game's setting (40 at a Commander table).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7a4cc57b-9f85-446c-9ca0-98c72b9bb732",
		Name:         "Happily Ever After",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Happily Ever After — each player gains 5 life and draws a card",
				effects.Do(effects.Instruction(func(ctx *effects.Context) error {
					for _, p := range effects.SeatsFromActive(ctx.Game) {
						if err := (effects.GainLife{Player: p, Amount: 5}).Apply(ctx); err != nil {
							return err
						}
						if err := (effects.DrawCards{Player: p, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return nil
				}))),
			effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor == source.Controller && happilyEverAfterHolds(g, source.Controller)
			}, "Happily Ever After — you win the game",
				effects.Do(effects.If{
					Cond: effects.CondBoard{Name: "five colors, six card types and at least your starting life", Test: func(ctx *effects.Context) bool {
						return happilyEverAfterHolds(ctx.Game, ctx.Controller())
					}},
					Then: []effects.Applier{effects.WinTheGame{}},
				})),
		},
	})
}

func happilyEverAfterHolds(g *game.Game, you uuid.UUID) bool {
	p := g.PlayerByIDForEffect(you)
	if p == nil || p.Eliminated || p.Life < g.Settings.StartingLife {
		return false
	}
	colors := map[string]bool{}
	types := map[string]bool{}
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller != you {
			continue
		}
		for _, color := range []string{"W", "U", "B", "R", "G"} {
			if c.HasColor(color) {
				colors[color] = true
			}
		}
		for _, t := range []string{"artifact", "battle", "creature", "enchantment", "instant", "kindred", "land", "planeswalker", "sorcery"} {
			if c.HasCardType(t) {
				types[t] = true
			}
		}
	}
	if len(colors) < 5 {
		return false
	}
	if p.Graveyard != nil {
		for _, c := range p.Graveyard.Cards {
			_, ts, _ := game.ParseTypeLine(c.TypeLine)
			for _, t := range ts {
				types[strings.ToLower(t)] = true
			}
		}
	}
	return len(types) >= 6
}

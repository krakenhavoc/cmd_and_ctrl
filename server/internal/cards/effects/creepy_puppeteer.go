package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Creepy Puppeteer — Creature — Human Rogue {3}{R}, 4/3 (Innistrad: Crimson
// Vow, #5495):
//
//	"Haste
//	 Whenever this creature attacks, if you attacked with exactly one other
//	 creature this combat, you may have that creature's base power and
//	 toughness become 4/3 until end of turn."
//
// One attack trigger with an intervening "if" (CR 603.4), checked as the
// attack is declared and again as it resolves, and a "you may" asked as it
// resolves (Optional). "Attacked with" is a fact about the declaration, so
// it is read from the turn's attack records (s82VowB06OtherAttackersThisCombat),
// not from who is still attacking: the other creature having died or left
// combat in response does not make the count wrong, and a creature put onto
// the battlefield attacking was not declared and is not counted. The one
// other creature, if it is still on the battlefield, has its base power and
// toughness set to 4/3 until end of turn (layer 7b), so counters and pumps
// still apply on top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7bb3c3b8-25b7-4c48-aa41-dcdcebbc3d5e",
		Name:            "Creepy Puppeteer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{
			InterveningIf(creepyPuppeteerExactlyOneOther, Optional(
				On(game.EventAttack, AllOf(ThisAttacked,
					func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						// Every attacker is staged before the declaration is
						// announced (battalion), but the turn's attack records
						// are written event by event, so as it triggers the
						// count is the staged attackers.
						return attackersYouControlOtherThan(g, source.Controller, source.InstanceID) == 1
					}),
					"Creepy Puppeteer — you may have that creature's base power and toughness become 4/3 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						ctx := NewContext(g, item)
						others := s82VowB06OtherAttackersThisCombat(g, ctx.Controller(), item.SourceCardID)
						if len(others) != 1 {
							return nil
						}
						other, ok := g.LookupCardForEffect(others[0])
						if !ok || !other.IsCreature() || !g.Battlefield.Contains(other.InstanceID) {
							return nil
						}
						return untilEndOfTurn(ctx, other.InstanceID, nil,
							"Creepy Puppeteer — base power and toughness 4/3", game.SetBasePTMods(4, 3)...)
					}),
				"Creepy Puppeteer — have that creature's base power and toughness become 4/3 until end of turn?")),
		},
	})
}

var creepyPuppeteerExactlyOneOther = CondBoard{
	Name: "you attacked with exactly one other creature this combat",
	Test: func(ctx *Context) bool {
		return len(s82VowB06OtherAttackersThisCombat(ctx.Game, ctx.Controller(), ctx.Item.SourceCardID)) == 1
	},
}

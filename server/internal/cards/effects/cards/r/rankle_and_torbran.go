package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Rankle and Torbran — Legendary Creature — Faerie Dwarf {1}{B}{B}{R}{R}, 3/4
// (March of the Machine Commander):
//
//	"Flying, first strike, haste
//	 Whenever Rankle and Torbran deals combat damage to a player or battle,
//	 choose any number —
//	 • Each player creates a Treasure token.
//	 • Each player sacrifices a creature of their choice.
//	 • If a source would deal damage to a player or battle this turn, it deals
//	   that much damage plus 2 instead."
//
// Rankle, Master of Pranks's "choose any number" trigger, fired by damage to a
// player or a battle (Bilbo, Unexpected Adventurer). The bullets resolve in
// printed order, so the Treasures exist before anyone sacrifices. The third is
// a turn-long additive replacement (Taii Wakeen, Perfect Shot's record, CR
// 614.1a) for any source, combat or not, whose recipient is a player or a
// battle (game.DamageRecipientsPlayersAndBattles); it is registered by the
// resolving ability so it outlives Rankle and Torbran, and beside another
// damage modifier the affected player orders them (CR 616.1).
//
// No simplification.
func init() {
	t := effects.On(game.EventDealDamage, effects.ThisDealtCombatDamageToAPlayerOrBattle,
		"Rankle and Torbran — choose any number", func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = effects.ChooseN("Choose any number", 0, 3,
		effects.ModeSteps("Each player creates a Treasure token.", nil,
			effects.Instruction(func(ctx *effects.Context) error {
				for _, id := range effects.TablePlayers(ctx) {
					if err := (effects.CreateToken{Controller: id, Template: effects.TreasureToken(), N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})),
		effects.ModeSteps("Each player sacrifices a creature of their choice.", nil,
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.JmpB1EachPlayerSacrificesN(1, "Rankle and Torbran — choose a creature to sacrifice")(ctx.Game, ctx.Item)
			})),
		effects.ModeSteps("If a source would deal damage to a player or battle this turn, it deals that much damage plus 2 instead.", nil,
			effects.Instruction(func(ctx *effects.Context) error {
				ctx.Game.AddDamageForEffect(game.DamageAddition{
					EffectSource: ctx.Source(),
					Controller:   ctx.Controller(),
					Amount:       2,
					Sources:      game.DamageSourcesAny,
					Recipients:   game.DamageRecipientsPlayersAndBattles,
					Label:        "Rankle and Torbran: damage to a player or battle plus 2",
				})
				return nil
			})),
	)
	effects.Register(effects.Spec{
		OracleID:        "80d918b2-3a28-49ad-a485-496658bd7ac3",
		Name:            "Rankle and Torbran",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "haste"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered:       []game.TriggeredAbility{t},
	})
}

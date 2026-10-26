package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	geekLotusWarriorPump = "geek-lotus-warrior/pump"
	geekLotusWarriorPing = "geek-lotus-warrior/creature-enters-damage"
)

// Geek Lotus Warrior — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Geek", "Lotus", "Warrior"; three art stickers.
//	{TK}{TK} — {2}: This creature gets +2/+0 until end of turn.
//	{TK}{TK}{TK}{TK} — Whenever a creature enters under your control, this permanent deals 2 damage to target player.
//	{TK}{TK} — 4/1
//	{TK}{TK}{TK} — 3/6
//
// The first sticker gives a pump ability on the permanent itself. The second gives a trigger on each creature entering under your control (the creature itself included) that deals 2 damage to a target player from the permanent it is on.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c0a448ee-e5f9-4e57-85b0-f6d401018170",
		Name:         "Geek Lotus Warrior",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: geekLotusWarriorPump,
				Activated: []ActivatedAbility{{
					Label:  "{2}: This creature gets +2/+0 until end of turn.",
					Cost:   ManaCost("{2}"),
					Effect: thisGetsUntilEndOfTurn(2, 0, "Geek Lotus Warrior sticker — +2/+0"),
				}},
				Text: "{2}: This creature gets +2/+0 until end of turn.",
			},
			{
				Key: geekLotusWarriorPing,
				Triggered: []game.TriggeredAbility{
					Targeting(
						On(game.EventETB, CreatureEnteredUnderYourControl,
							"Geek Lotus Warrior sticker — this permanent deals 2 damage to target player",
							func(g *game.Game, item *game.StackItem) error {
								ctx := NewContext(g, item)
								if len(item.Targets) == 0 || !ctx.IsTargetLegal(item.Targets[0]) {
									return nil
								}
								return DealDamage{Source: item.SourceCardID, Target: item.Targets[0].ID, Amount: 2}.Apply(ctx)
							}),
						TargetPlayer("target player")),
				},
				Text: "Whenever a creature enters under your control, this permanent deals 2 damage to target player.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: geekLotusWarriorPump},
			{Grant: geekLotusWarriorPing},
		}},
	})
}

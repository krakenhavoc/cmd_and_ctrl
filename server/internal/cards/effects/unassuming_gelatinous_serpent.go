package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	unassumingGelatinousSerpentReturn = "unassuming-gelatinous-serpent/dies-return"
	unassumingGelatinousSerpentMill   = "unassuming-gelatinous-serpent/damage-mill"
)

// Unassuming Gelatinous Serpent — Stickers (an Unfinity sticker sheet, CR 123.2; ADR 0208, #4307):
//
//	Name stickers "Unassuming", "Gelatinous", "Serpent"; three art stickers.
//	{TK}{TK} — When this permanent dies, return target noncreature, nonland card from your graveyard to your hand.
//	{TK}{TK}{TK}{TK} — Whenever this creature deals combat damage to a player, that player mills twice that many cards.
//	{TK}{TK} — 2/3
//	{TK}{TK}{TK} — 7/2
//
// The first sticker gives a dies trigger that returns a target noncreature, nonland card from your graveyard to your hand. The second gives a combat-damage trigger: the damaged player mills twice the damage dealt.
//
// The name, power and toughness, and art stickers are the rules' (CR 123.6, 123.8, 123.9).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ad4d181-21ce-47f3-9c00-2aa8c307e7fe",
		Name:         "Unassuming Gelatinous Serpent",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{
			{
				Key: unassumingGelatinousSerpentReturn,
				Triggered: []game.TriggeredAbility{
					Targeting(
						WhenThisDies("Unassuming Gelatinous Serpent sticker — return target noncreature, nonland card from your graveyard to your hand",
							func(g *game.Game, item *game.StackItem) error {
								ctx := NewContext(g, item)
								id, ok := stickerFirstLegalTarget(ctx)
								if !ok {
									return nil
								}
								return ReturnFromGraveyard{Target: id, Dest: game.ZoneHand}.Apply(ctx)
							}),
						TargetCardInGraveyard("target noncreature, nonland card in your graveyard", YouOwn(), Not(Creature()), Nonland())),
				},
				Text: "When this permanent dies, return target noncreature, nonland card from your graveyard to your hand.",
			},
			{
				Key: unassumingGelatinousSerpentMill,
				Triggered: []game.TriggeredAbility{
					WheneverThisDealsCombatDamageToAPlayer("Unassuming Gelatinous Serpent sticker — that player mills twice that many cards",
						func(g *game.Game, item *game.StackItem) error {
							ev := item.Trigger.Event
							if ev.Amount <= 0 || g.PlayerByIDForEffect(ev.Target) == nil {
								return nil
							}
							return MillCards{Player: ev.Target, N: 2 * ev.Amount}.Apply(NewContext(g, item))
						}),
				},
				Text: "Whenever this creature deals combat damage to a player, that player mills twice that many cards.",
			},
		},
		StickerSheet: &StickerSheetSpec{Abilities: []StickerAbilitySpec{
			{Grant: unassumingGelatinousSerpentReturn},
			{Grant: unassumingGelatinousSerpentMill},
		}},
	})
}

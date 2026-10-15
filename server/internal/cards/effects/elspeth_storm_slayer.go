package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Elspeth, Storm Slayer — Legendary Planeswalker — Elspeth {3}{W}{W},
// loyalty 5 (EDHREC top 1,000, #3654):
//
//	"If one or more tokens would be created under your control, twice
//	 that many of those tokens are created instead.
//	 +1: Create a 1/1 white Soldier creature token.
//	 0: Put a +1/+1 counter on each creature you control. Those
//	 creatures gain flying until your next turn.
//	 −3: Destroy target creature an opponent controls with mana value 3
//	 or greater."
//
// The static is Doubling Season's token half (TokensDoubled), so her own
// +1 makes two Soldiers. The 0 snapshots your creatures, counters each
// (through the counter pipeline, so Hardened Scales applies) and gives
// the same set flying until your next turn; a creature that arrives
// later gets neither. The −3 is a targeted destroy gated on mana value
// at announce and again at resolution.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f78af825-023a-42e9-8374-5c52303a1417",
		Name:            "Elspeth, Storm Slayer",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		Replacements:    []game.ReplacementEffect{TokensDoubled("Elspeth, Storm Slayer: double tokens")},
		Activated: []ActivatedAbility{
			{
				Label:  "+1: Create a 1/1 white Soldier creature token.",
				Cost:   LoyaltyCost(1),
				Effect: Do(CreateToken{Template: TokenCard("1/1 white Soldier"), N: 1}),
			},
			{
				Label: "0: Put a +1/+1 counter on each creature you control. Those creatures gain flying until your next turn.",
				Cost:  LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() {
							ids = append(ids, c.InstanceID)
						}
					}
					for _, id := range ids {
						if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					for _, id := range ids {
						if !onBattlefield(g, id) {
							continue
						}
						if err := (ScopedEffectFor{
							Target:   id,
							Mods:     []game.Mod{game.AddKeywordsMod("flying")},
							Duration: DurationUntilYourNextTurn(ctx, item.Controller),
							Label:    "Elspeth, Storm Slayer — flying until your next turn",
						}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label:   "−3: Destroy target creature an opponent controls with mana value 3 or greater.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetCreature("target creature an opponent controls with mana value 3 or greater", OpponentControls(), ManaValueGE(3)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return DestroyTarget{Target: t.ID}.Apply(ctx)
					}
					return nil
				},
			},
		},
	})
}

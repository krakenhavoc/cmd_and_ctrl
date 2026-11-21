package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Serra the Benevolent — Legendary Planeswalker — Serra {2}{W}{W}, loyalty 4:
//
//	"+2: Creatures you control with flying get +1/+1 until end of turn.
//	 −3: Create a 4/4 white Angel creature token with flying and vigilance.
//	 −6: You get an emblem with "If you control a creature, damage that would
//	 reduce your life total to less than 1 reduces it to 1 instead.""
//
// The +2 affects the flyers you control as it resolves (CR 611.2c). The
// emblem is a replacement on the life window (CR 614.1, CR 120.3a): a life
// event of its owner that is the life damage costs, that would leave the total
// at 0 or less, while they control a creature, leaves it at 1 instead
// (LifeTotalBecomesInsteadForEffect). A loss that is not damage is untouched.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8b0e2382-dd91-42e6-aba3-5e8a15418a49",
		Name:            "Serra the Benevolent",
		Completeness:    CompletenessFull,
		StartingLoyalty: 4,
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		Emblem: &EmblemSpec{
			Label:        "Serra the Benevolent emblem",
			Text:         "If you control a creature, damage that would reduce your life total to less than 1 reduces it to 1 instead.",
			Replacements: []game.ReplacementEffect{serraTheBenevolentEmblemLife()},
		},
		Activated: []ActivatedAbility{
			{
				Label: "+2: Creatures you control with flying get +1/+1 until end of turn.",
				Cost:  LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsCreature() && game.HasKeyword(&c, "flying") {
							ids = append(ids, c.InstanceID)
						}
					}
					for _, id := range ids {
						if err := (BoostUntilEOT{Target: id, Power: 1, Toughness: 1,
							Label: "Serra the Benevolent — +1/+1 until end of turn"}).Apply(ctx.asGroupMember()); err != nil {
							return err
						}
					}
					return nil
				},
			},
			{
				Label:   "−3: Create a 4/4 white Angel creature token with flying and vigilance.",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Cost:    LoyaltyCost(-3),
				Effect:  Do(CreateToken{Template: TokenCard("4/4 white Angel with flying and vigilance"), N: 1}),
			},
			{
				Label: "−6: You get an emblem with \"If you control a creature, damage that would reduce your life total to " +
					"less than 1 reduces it to 1 instead.\"",
				Cost: LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// serraTheBenevolentEmblemLife is the emblem's one ability.
func serraTheBenevolentEmblemLife() game.ReplacementEffect {
	return game.ReplacementEffect{
		Watches: []game.EventKind{game.EventChangeLife},
		AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, src *game.Card) bool {
			if src == nil || src.Controller == uuid.Nil || ev.LifePlayer != src.Controller || !ev.LifeFromDamage {
				return false
			}
			if !g.LifeWouldBeReducedToZeroOrLessForEffect(ev) {
				return false
			}
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == src.Controller && c.IsCreature() {
					return true
				}
			}
			return false
		},
		Replace: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) error {
			g.LifeTotalBecomesInsteadForEffect(ev, 1)
			return nil
		},
		Controller: func(_ *game.ReplacementEvent, _ *game.Game, src *game.Card) uuid.UUID {
			return src.Controller
		},
		Label: "Serra the Benevolent emblem: damage reduces your life total to 1 instead of less than 1",
	}
}

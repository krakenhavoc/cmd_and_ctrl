package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

const (
	arlinnKordOracleID = "6bb7d0df-7a9f-4e17-8c31-7628c4b12356"
	arlinnKordPingKey  = "arlinn-embraced-by-the-moon/tap-for-damage"
)

// Arlinn Kord // Arlinn, Embraced by the Moon — Legendary Planeswalker —
// Arlinn {2}{R}{G}, loyalty 3 (Innistrad Remastered), a double-faced
// planeswalker:
//
//	Arlinn Kord
//	 "+1: Until end of turn, up to one target creature gets +2/+2 and gains
//	  vigilance and haste.
//	  0: Create a 2/2 green Wolf creature token. Transform Arlinn Kord."
//	Arlinn, Embraced by the Moon
//	 "+1: Creatures you control get +1/+1 and gain trample until end of turn.
//	  −1: Arlinn deals 3 damage to any target. Transform Arlinn.
//	  −6: You get an emblem with "Creatures you control have haste and '{T}:
//	  This creature deals damage equal to its power to any target.'""
//
// The two faces are two registrations (Garruk Relentless's shape). "Transform
// Arlinn" is the ability's last instruction and does nothing if the
// planeswalker has already been turned over since the ability went on the
// stack (CR 701.28f). The −6 emblem grants a layer-6 keyword and a tap
// ability bundle (Burning Anger's damage body) to the creatures its owner
// controls, as long as the emblem exists (CR 114.2).
func init() {
	Register(Spec{
		OracleID:        arlinnKordOracleID,
		Name:            "Arlinn Kord",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Emblem:          arlinnKordEmblem(),
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Until end of turn, up to one target creature gets +2/+2 and gains vigilance and haste.",
				Cost:    LoyaltyCost(1),
				Purpose: ForTargets(BenefitUntilEOT(0, 2, 2, "vigilance", "haste")),
				Targets: TargetCreature("up to one target creature").WithCount(0, 1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					targets := ctx.LegalTargets()
					if len(targets) == 0 {
						return nil
					}
					if err := (BoostUntilEOT{Target: targets[0].ID, Power: 2, Toughness: 2, Label: "Arlinn Kord — +2/+2"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Target: targets[0].ID, Keywords: []string{"vigilance", "haste"}, Label: "Arlinn Kord — vigilance and haste"}.Apply(ctx)
				},
			},
			{
				Label:   "0: Create a 2/2 green Wolf creature token. Transform Arlinn Kord.",
				Cost:    LoyaltyCost(0),
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1},
				Effect:  Do(CreateToken{Template: TokenCard("2/2 green Wolf"), N: 1}, TransformThis{}),
			},
		},
	})

	Register(Spec{
		OracleID:     arlinnKordOracleID + "#1",
		Name:         "Arlinn, Embraced by the Moon",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: arlinnKordPingKey,
			Activated: []ActivatedAbility{{
				Label:   "{T}: This creature deals damage equal to its power to any target.",
				Cost:    TapCost(),
				Targets: TargetAny(),
				Effect:  sourceDealsDamageEqualToItsPowerToTarget,
			}},
			Text: "{T}: This creature deals damage equal to its power to any target.",
		}},
		Emblem: arlinnKordEmblem(),
		Activated: []ActivatedAbility{
			{
				Label: "+1: Creatures you control get +1/+1 and gain trample until end of turn.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (BoostUntilEOT{Match: And(Creature(), YouControl()), Power: 1, Toughness: 1, Label: "Arlinn — +1/+1"}).Apply(ctx); err != nil {
						return err
					}
					return GrantKeywordUntilEOT{Match: And(Creature(), YouControl()), Keywords: []string{"trample"}, Label: "Arlinn — trample"}.Apply(ctx)
				},
			},
			{
				Label:   "−1: Arlinn deals 3 damage to any target. Transform Arlinn.",
				Cost:    LoyaltyCost(-1),
				Purpose: ForTargets(DamageToTarget(0, 3)),
				Targets: TargetAny(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if targets := ctx.LegalTargets(); len(targets) > 0 {
						if err := (DealDamage{Source: item.SourceCardID, Target: targets[0].ID, Amount: 3}).Apply(ctx); err != nil {
							return err
						}
					}
					return TransformThis{}.Apply(ctx)
				},
			},
			{
				Label: "−6: You get an emblem with \"Creatures you control have haste and '{T}: This creature deals damage equal to its power to any target.'\"",
				Cost:  LoyaltyCost(-6),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateEmblem{}.Apply(NewContext(g, item))
				},
			},
		},
	})
}

// arlinnKordEmblem is the emblem of the −6, declared on both faces. The
// emblem is filed under the key of the card that makes it, and Arlinn at six
// loyalty pays the −6 down to zero, so by the time the ability resolves she
// is in the graveyard and has turned back to her front face (CR 712.8a): the
// front registration is the one the engine finds then, the back one covers a
// source still on the battlefield.
func arlinnKordEmblem() *EmblemSpec {
	return &EmblemSpec{
		Label: "Arlinn, Embraced by the Moon emblem",
		Text:  "Creatures you control have haste and \"{T}: This creature deals damage equal to its power to any target.\"",
		Static: []game.StaticAbility{
			{
				Label:     "Creatures you control have haste.",
				Layer:     game.Layer6Ability,
				AppliesTo: CreaturesYouControl,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Abilities = game.AppendKeywordAbility(c.Abilities, "haste")
				},
			},
			GrantAbilities(CreaturesYouControl, arlinnKordPingKey),
		},
	}
}

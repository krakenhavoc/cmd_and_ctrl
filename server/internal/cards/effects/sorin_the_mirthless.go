package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sorin the Mirthless — Legendary Planeswalker — Sorin {2}{B}{B},
// starting loyalty 4 (#4077, deck request Sorin's Birthday Party):
//
//	"+1: Look at the top card of your library. You may reveal that
//	     card and put it into your hand. If you do, you lose life
//	     equal to its mana value.
//	 −2: Create a 2/3 black Vampire creature token with flying and
//	     lifelink.
//	 −7: Sorin deals 13 damage to any target. You gain 13 life."
//
// The +1 is the optional look-and-take (TakeFromLibraryToHand): a LOOK,
// so only Sorin's controller sees the card; taking it reveals it
// (CR 701.20) before it moves, so the table learns what it cost. Nothing
// is drawn, so a draw replacement or a "whenever you draw" payoff does
// not see it. "If you do" is the Taken list: a declined look, an empty
// library and a commander that took its CR 903.9b offer instead of the
// hand all leave nothing taken and no life paid. The mana value is read
// off the card in the hand (CR 202.3e: {X} is zero anywhere but the
// stack, so it is the same number it was on the library), and a land
// costs nothing.
//
// The −7 is the widest target clause (any target), and the life gain is
// a separate sentence: a prevented 13 still gains the 13.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fbffaf9f-40e7-4113-a037-533b733cebce",
		Name:         "Sorin the Mirthless",
		Completeness: CompletenessFull,
		// Printed loyalty reaches the card through deck import
		// (ADR 0032 §1); this is the fallback for tokens, fixtures
		// and the dev spawner.
		StartingLoyalty: 4,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Look at the top card of your library. You may reveal that card and put it into your hand. If you do, you lose life equal to its mana value.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player:   player,
						Cards:    g.LookAtTopOfLibraryForEffect(player, 1),
						Max:      1,
						Optional: true,
						Label:    "Sorin the Mirthless — reveal that card and put it into your hand",
						Key:      sorinMirthlessTake,
					}.Apply(ctx)
				},
			},
			{
				Label: "−2: Create a 2/3 black Vampire creature token with flying and lifelink.",
				Cost:  LoyaltyCost(-2),
				Effect: Do(CreateToken{
					Template: TokenCard("2/3 black Vampire with flying and lifelink"),
					N:        1,
				}),
			},
			{
				Label:   "−7: Sorin deals 13 damage to any target. You gain 13 life.",
				Cost:    LoyaltyCost(-7),
				Targets: TargetAny(),
				Purpose: ForTargets(DamageToTarget(0, 13)),
				// The life gain waits for a prompt the damage asks (CR 608.2c, #4155).
				Effect: Do(
					// CR 608.2b: with its only target gone the ability does
					// nothing at all, life gain included (#4082). A prevented
					// damage still gains the life.
					Instruction(func(ctx *Context) error {
						legal := ctx.LegalTargets()
						if len(legal) == 0 {
							return nil
						}
						return DealDamage{Source: ctx.Item.SourceCardID, Target: legal[0].ID, Amount: 13}.Apply(ctx)
					}),
					Instruction(func(ctx *Context) error {
						if len(ctx.LegalTargets()) == 0 {
							return nil
						}
						return GainLife{Amount: 13}.Apply(ctx)
					}),
				),
			},
		},
	})
}

// sorinMirthlessPayForTheCard is the +1's "if you do, you lose life
// equal to its mana value": it acts on what was taken, so it is a no-op
// for a declined look.
func sorinMirthlessPayForTheCard(g *game.Game, res TakeFromLibraryResult) error {
	if len(res.Taken) == 0 {
		return nil
	}
	c, ok := g.LookupCardForEffect(res.Taken[0])
	if !ok {
		return nil
	}
	mv := c.ManaValue()
	if mv <= 0 {
		return nil
	}
	return g.ChangePlayerLifeForEffect(res.Source, res.Player, -mv)
}

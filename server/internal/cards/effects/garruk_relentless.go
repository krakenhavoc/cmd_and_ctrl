package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Garruk Relentless // Garruk, the Veil-Cursed — Legendary Planeswalker
// {3}{G}, loyalty 3, transforming:
//
//	Front: "When Garruk has two or fewer loyalty counters on him,
//	        transform him.
//	        0: Garruk deals 3 damage to target creature. That creature
//	           deals damage equal to its power to him.
//	        0: Create a 2/2 green Wolf creature token."
//	Back:  "+1: Create a 1/1 black Wolf creature token with deathtouch.
//	        −1: Sacrifice a creature. If you do, search your library for
//	            a creature card, reveal it, put it into your hand, then
//	            shuffle.
//	        −3: Creatures you control gain trample and get +X/+X until
//	            end of turn, where X is the number of creature cards in
//	            your graveyard."
//
// ADR 0107 §1 (#1858):
//
//   - The transform is a CR 603.8 state trigger on the front face. It
//     triggers whenever Garruk is at two or fewer — after his own fight,
//     or after damage from an attacker — and Garruk stays the same object
//     as he turns over (CR 712.18), loyalty counters and all. CR 701.27f:
//     the trigger transforms him only if he has not transformed since it
//     was put on the stack, so a Garruk already back face up is left alone.
//   - The fight: Garruk's 3 damage is dealt first, and then the creature
//     deals damage equal to its power to him. State-based actions are not
//     checked in between (CR 704.3), so a creature dealt lethal damage is
//     still there and still hits back; one that left the battlefield some
//     other way hits back with its power as it last existed (CR 608.2h).
//     A Garruk that has left is not dealt the damage.
//   - The −1's search is "if you do" (CR 701.21a): no sacrifice, no search.
//   - The −3's X is counted as it resolves; the creatures it affects are
//     the ones you control then (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        garrukRelentlessOracleID,
		Name:            "Garruk Relentless",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Triggered: []game.TriggeredAbility{
			WhenState("Garruk Relentless — transform him",
				func(_ *game.Game, source *game.Card, _ uuid.UUID) bool {
					return source.Counters[game.CounterLoyalty] <= 2
				}, garrukTransform),
		},
		Activated: []ActivatedAbility{
			{
				Label:   "0: Garruk deals 3 damage to target creature. That creature deals damage equal to its power to him.",
				Cost:    LoyaltyCost(0),
				Targets: TargetCreature("target creature"),
				Effect:  garrukFight,
			},
			{
				Label: "0: Create a 2/2 green Wolf creature token.",
				Cost:  LoyaltyCost(0),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Template: TokenCard("2/2 green Wolf"), N: 1}.Apply(NewContext(g, item))
				},
			},
		},
	})

	Register(Spec{
		OracleID:     garrukRelentlessOracleID + "#1",
		Name:         "Garruk, the Veil-Cursed",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label: "+1: Create a 1/1 black Wolf creature token with deathtouch.",
				Cost:  LoyaltyCost(1),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Template: TokenCard("1/1 black Wolf with deathtouch"), N: 1}.Apply(NewContext(g, item))
				},
			},
			{
				Label:  "−1: Sacrifice a creature. If you do, search your library for a creature card, reveal it, put it into your hand, then shuffle.",
				Cost:   LoyaltyCost(-1),
				Effect: garrukSacrificeThenTutor,
			},
			{
				Label:  "−3: Creatures you control gain trample and get +X/+X until end of turn, where X is the number of creature cards in your graveyard.",
				Cost:   LoyaltyCost(-3),
				Effect: garrukOverrun,
			},
		},
	})
}

const garrukRelentlessOracleID = "7cec9021-6f25-4fd8-b40e-adf4ffd3a7b8"

// garrukTransform turns Garruk over unless he already has (CR 701.27f).
func garrukTransform(g *game.Game, item *game.StackItem) error {
	c, ok := g.LookupCardForEffect(item.SourceCardID)
	if !ok || !onBattlefield(g, item.SourceCardID) || c.ActiveFace != 0 {
		return nil
	}
	return TransformThis{}.Apply(NewContext(g, item))
}

// garrukFight is the front face's first 0.
func garrukFight(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	creature := targets[0].ID
	ref, ok := g.PermanentRefForEffect(creature)
	if !ok {
		return nil
	}
	if err := (DealDamage{Source: item.SourceCardID, Target: creature, Amount: 3}).Apply(ctx); err != nil {
		return err
	}
	info, ok := g.PermanentForEffect(ref)
	if !ok || info.Power <= 0 || !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
		return nil
	}
	return DealDamage{SourceObject: &ref, Target: item.SourceCardID, Amount: info.Power}.Apply(ctx)
}

// garrukSacrificeThenTutor is the back face's −1.
func garrukSacrificeThenTutor(g *game.Game, item *game.StackItem) error {
	return g.PlayerSacrificesThenForEffect(item.SourceCardID, item.Controller,
		sacrificeSpec("a creature you control", Creature(), YouControl()),
		"Garruk, the Veil-Cursed — sacrifice a creature", 1,
		func(g *game.Game, sacrificed game.PromptedSacrifices) error {
			if !sacrificed.Sacrificed(item.Controller) {
				return nil
			}
			return SearchLibrary{
				Player:    item.Controller,
				Predicate: func(c game.Card) bool { return c.IsCreature() },
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Garruk, the Veil-Cursed — a creature card",
			}.Apply(NewContext(g, item))
		})
}

// garrukOverrun is the back face's −3.
func garrukOverrun(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := b11CreatureCardsInGraveyard(g, item.Controller)
	yours := And(Creature(), YouControl())
	if x > 0 {
		if err := (BoostUntilEOT{Match: yours, Power: x, Toughness: x, Label: "Garruk, the Veil-Cursed — +X/+X"}).Apply(ctx); err != nil {
			return err
		}
	}
	return GrantKeywordUntilEOT{Match: yours, Keywords: []string{"trample"}, Label: "Garruk, the Veil-Cursed — trample"}.Apply(ctx)
}

package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hero of Bladehold — Creature — Human Knight {2}{W}{W}, 3/4:
//
//	"Battle cry (Whenever this creature attacks, each other attacking
//	 creature gets +1/+0 until end of turn.)
//	 Whenever this creature attacks, create two 1/1 white Soldier
//	 creature tokens that are tapped and attacking."
//
// Two triggers off the one attack, as the printed card has: the battle
// cry and the token maker. They are separate abilities so the
// controller orders them (CR 603.3b); the order is the card's whole
// skill, because the tokens only get the +1/+0 when the battle cry
// resolves AFTER they arrive.
//
// Battle cry has no keyword token in the engine's table, so it is
// written out as the trigger it is rather than declared in
// PrintedKeywords, which would put a badge on a rule nothing reads.
// "Each other attacking creature" is every attacker but this one; in
// a four-player game all of them are the active player's.
//
// The tokens are put onto the battlefield tapped and attacking
// (CreateTokensAttackingYourChoice): never declared, so they fire no
// "whenever a creature attacks" trigger (CR 508.4) and can be
// blocked. The controller chooses what each attacks, a planeswalker or
// battle included, with what Hero attacked asked first. No
// simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "454a8902-8120-4373-96ee-bbf352b04e8d",
		Name:         "Hero of Bladehold",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisAttacks("Hero of Bladehold — battle cry: each other attacking creature gets +1/+0",
				func(g *game.Game, item *game.StackItem) error {
					return effects.BoostUntilEOT{
						Match: effects.And(effects.AttackingCreature(), effects.OtherThan(item.SourceCardID)),
						Power: 1,
						Label: "Hero of Bladehold — battle cry",
					}.Apply(effects.NewContext(g, item))
				}),
			effects.WheneverThisAttacks("Hero of Bladehold — two tapped and attacking 1/1 Soldiers",
				func(g *game.Game, item *game.StackItem) error {
					return effects.CreateTokensAttackingYourChoice{
						Template: effects.TokenCard("1/1 white Soldier"),
						N:        2,
						Tapped:   true,
						Prefer:   effects.TriggerAttackTarget(item),
						Label:    "Hero of Bladehold",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

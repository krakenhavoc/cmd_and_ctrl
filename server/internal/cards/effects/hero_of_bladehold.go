package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// The tokens are put onto the battlefield tapped and attacking the
// player Hero attacked (CreateTokensAttackingForEffect, Brimaz's
// path): never declared, so they fire no "whenever a creature
// attacks" trigger (CR 508.4) and can be blocked. When Hero attacks a
// planeswalker or battle the tokens arrive tapped and not attacking,
// since that path takes a player; weaker than printed, never stronger.
func init() {
	Register(Spec{
		OracleID:     "454a8902-8120-4373-96ee-bbf352b04e8d",
		Name:         "Hero of Bladehold",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"When Hero attacks a planeswalker or battle, the two Soldiers arrive tapped but not attacking."},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Hero of Bladehold — battle cry: each other attacking creature gets +1/+0",
				func(g *game.Game, item *game.StackItem) error {
					return BoostUntilEOT{
						Match: And(AttackingCreature(), OtherThan(item.SourceCardID)),
						Power: 1,
						Label: "Hero of Bladehold — battle cry",
					}.Apply(NewContext(g, item))
				}),
			WheneverThisAttacks("Hero of Bladehold — two tapped and attacking 1/1 Soldiers",
				func(g *game.Game, item *game.StackItem) error {
					tmpl := TokenCard("1/1 white Soldier")
					tmpl.Tapped = true
					return g.CreateTokensAttackingForEffect(item.Controller, tmpl, 2, item.Trigger.Event.Target)
				}),
		},
	})
}

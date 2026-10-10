package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Guild Artisan — Legendary Enchantment — Background {1}{R} (EDHREC
// rank 3588):
//
//	"Commander creatures you own have "Whenever this creature attacks
//	 a player, if no opponent has more life than that player, you
//	 create two Treasure tokens." (They're artifacts with "{T},
//	 Sacrifice this token: Add one mana of any color.")"
//
// The Treasure Background. Since #2874 the ability is really GRANTED
// (ADR 0093), Agent of the Iron Throne's posture: a layer-6 grant of
// the trigger to every commander creature the Background's controller
// owns, under anyone's control. It used to live on the Background,
// gated on a commander you also controlled, so a commander an opponent
// had stolen lost it; now the thief's attack makes the thief's
// Treasures, as printed. The intervening if — no opponent of the
// creature's controller has more life than the attacked player — is
// read as the attack is declared and again as the trigger resolves (CR
// 603.4). "Attacks a player": an attack on a planeswalker or a battle
// does not fire it. Two commander creatures each have it. The
// Treasures are the real token, with its own mana ability.
//
// No simplification.
const guildArtisanGrant = "guild-artisan/treasures"

func init() {
	Register(Spec{
		OracleID:     "aceac269-5100-428a-a4a3-bac57031e30b",
		Name:         "Guild Artisan",
		Completeness: CompletenessFull,
		Grants: []AbilityGrant{{
			Key: guildArtisanGrant,
			Triggered: []game.TriggeredAbility{
				wheneverThisAttacksAPlayerNoOpponentRicher("Guild Artisan — create two Treasures", func(g *game.Game, item *game.StackItem) error {
					return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: 2}.Apply(NewContext(g, item))
				}),
			},
			Text: "Whenever this creature attacks a player, if no opponent has more life than that player, you create two Treasure tokens.",
		}},
		Static: []game.StaticAbility{grantToCommanderCreaturesYouOwn(guildArtisanGrant)},
	})
}

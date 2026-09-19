package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Olivia, Opulent Outlaw — Legendary Creature — Vampire Assassin
// {1}{R}{W}{B}, 3/3 (EDHREC rank 3372):
//
//	"Flying, lifelink
//	 Whenever one or more outlaws you control deal combat damage to a
//	 player, create a Treasure token. (Assassins, Mercenaries, Pirates,
//	 Rogues, and Warlocks are outlaws.)
//	 {3}, Sacrifice two Treasures: Put two +1/+1 counters on each
//	 creature you control. Activate only as a sorcery."
//
// The Mardu outlaw commander, herself an Assassin.
//
//   - The Treasure trigger is Grazilaxx's combat-damage watcher
//     narrowed to the five outlaw creature types (read post-layer, so
//     a changeling counts): the engine emits one damage event per
//     creature, and "one or more … to a player" is one Treasure per
//     PLAYER the outlaws connect with in that damage step (CR
//     603.2c, #784) — two outlaws on one opponent make one Treasure,
//     two opponents hit at once make two.
//   - The anthem's cost is a mana component plus a sacrifice clause
//     with a count of two (#747, SacrificeN) at sorcery speed. The
//     counters go on each creature the controller controls as the
//     ability resolves, as one placement of two counters per creature
//     (b11PutCountersOnEachCreatureYouControl), so a counter
//     replacement such as Hardened Scales applies once, not twice.
//
// No simplification. The once-per-damage-step retreat this card
// shipped with is gone: #784 gave the "one or more" guard its second
// dimension, so the trigger is once per player, as printed.
func init() {
	Register(Spec{
		OracleID:        "4e5e0b22-ee5a-4c19-be95-2f9bd50641bc",
		Name:            "Olivia, Opulent Outlaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(outlaw(),
				"Olivia, Opulent Outlaw — create a Treasure", Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:        "{3}, Sacrifice two Treasures: Put two +1/+1 counters on each creature you control. Activate only as a sorcery.",
			Cost:         Plus(ManaCost("{3}"), SacrificeN(2, "two Treasures", isTreasure)),
			SorcerySpeed: true,
			Effect: func(g *game.Game, item *game.StackItem) error {
				return b11PutCountersOnEachCreatureYouControl(g, item, 2)
			},
		}},
	})
}

// outlaw is Outlaws of Thunder Junction's batching term (CR 700.12):
// a creature that is an Assassin, a Mercenary, a Pirate, a Rogue or a
// Warlock. Reads the effective subtypes.
func outlaw() CardPredicate {
	return AnySubtype("Assassin", "Mercenary", "Pirate", "Rogue", "Warlock")
}

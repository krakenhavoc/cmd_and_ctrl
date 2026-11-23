package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Icingdeath, Frost Tyrant — Legendary Creature — Dragon {2}{W}{W}, 4/3
// (Adventures in the Forgotten Realms Commander):
//
//	"Flying, vigilance
//	 When Icingdeath, Frost Tyrant dies, create Icingdeath, Frost Tongue, a
//	 legendary white Equipment artifact token with "Equipped creature gets
//	 +2/+0," "Whenever equipped creature attacks, tap target creature
//	 defending player controls," and equip {2}."
//
// The token is a catalog template of its own (token_catalog.go), so its three
// abilities survive a snapshot as the key alone. The attack trigger is
// Captain America's Shield's: the target is read off the attacker's own
// attack, so a planeswalker attack finds its defender.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1dc2c213-72f6-4112-96a8-bf6ad553c707",
		Name:            "Icingdeath, Frost Tyrant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Icingdeath, Frost Tyrant — create Icingdeath, Frost Tongue",
				Do(CreateToken{Template: IcingdeathFrostTongueToken(), N: 1})),
		},
	})
}

// IcingdeathFrostTongueToken is the legendary white Equipment token as printed.
func IcingdeathFrostTongueToken() game.Card {
	return TokenFromCatalog(printedIcingdeathFrostTongueToken)
}

func printedIcingdeathFrostTongueToken() TokenTemplate {
	return TokenTemplate{
		Slug: "icingdeath-frost-tongue",
		Card: game.Card{
			Name:     "Icingdeath, Frost Tongue",
			TypeLine: "Token Legendary Artifact — Equipment",
			Colors:   []string{"W"},
		},
		Text:   "Equipped creature gets +2/+0.\nWhenever equipped creature attacks, tap target creature defending player controls.\nEquip {2}",
		Static: []game.StaticAbility{PumpAttached(2, 0)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return AttachedCreatureAttacked(ev, source)
			},
			TargetsFrom: MshB2TargetCreatureTheAttackerDefends,
			Key:         "Icingdeath, Frost Tongue — tap target creature defending player controls",
			Effect:      B36TapChosenCreature,
		}},
		Activated: ActivatedShapes([]ActivatedAbility{EquipAbility("{2}")}),
	}
}

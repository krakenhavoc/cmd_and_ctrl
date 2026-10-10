package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Airtight Alibi — Enchantment — Aura {2}{G}:
//
//	"Flash
//	 Enchant creature
//	 When this Aura enters, untap enchanted creature. It gains hexproof
//	 until end of turn. If it's suspected, it's no longer suspected.
//	 Enchanted creature gets +2/+2 and can't become suspected."
//
// "Can't become suspected" is the CantBecomeSuspected restriction
// (#2733, ADR 0071 amendment 2026-10-10), which Game.SuspectForEffect
// reads: a suspect instruction aimed at the enchanted creature does
// nothing, and a card that offers a choice of creatures to suspect
// (Frantic Scapegoat) leaves it out. Like Pacifism's restriction it is
// the Aura's, so it lasts exactly as long as the Aura is attached and a
// "loses all abilities" on the creature does not end it. It does not
// clear a suspicion the creature already has; the enters trigger does
// that.
//
// The enters trigger names "enchanted creature" as it resolves, or as
// the Aura last was if the Aura has gone (CR 608.2h), the way
// Convenient Target's does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "cb666ea8-55e3-4e27-ae6d-806677dfc17a",
		Name:            "Airtight Alibi",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(2, 2),
			RestrictAttached(game.CantBecomeSuspected),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Airtight Alibi — untap enchanted creature, hexproof, no longer suspected", airtightAlibiEnters),
		},
	})
}

// airtightAlibiEnters untaps the enchanted creature, gives it hexproof
// until end of turn, and ends its suspicion if it has one.
func airtightAlibiEnters(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	host, ok := enchantedCreatureAsItResolves(ctx)
	if !ok {
		return nil
	}
	if err := (UntapTarget{Target: host}).Apply(ctx); err != nil {
		return err
	}
	if err := (GrantKeywordUntilEOT{
		Target:   host,
		Keywords: []string{"hexproof"},
		Label:    "Airtight Alibi — hexproof until end of turn",
	}).Apply(ctx); err != nil {
		return err
	}
	return Unsuspect{Target: host}.Apply(ctx)
}

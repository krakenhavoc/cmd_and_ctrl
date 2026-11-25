package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// source_targeting_bar.go — the card side of #6196: "can't be the target
// of" a spell or ability WITH A QUALITY (game/source_targeting_bar.go).
//
//	Bartel Runeaxe, Tetsuo Umezawa:
//	    Static: []game.StaticAbility{CantBeTargetedByAuraSpells(text)}
//
//	Raiding Party ("white spells or abilities from white sources"):
//	    Static: []game.StaticAbility{CantBeTargetedByColor("W", text)}
//
//	Anti-Magic Aura ("enchanted creature can't be the target of spells"):
//	    Static: []game.StaticAbility{AttachedCantBeTargetedBySpells(text)}
//
//	Artifact Ward ("enchanted creature can't be the target of abilities
//	from artifact sources"):
//	    Static: []game.StaticAbility{AttachedCantBeTargetedByAbilitiesFrom("artifact", text)}
//
//	Lurker ("can't be the target of spells unless it attacked or blocked
//	this turn"):
//	    Static: []game.StaticAbility{CantBeTargetedBySpellsUnlessItAttackedOrBlocked(text)}
//
// The quality matcher is protection's (CR 702.16b), and the bar is ONLY
// a targeting bar: none of these cards prevents damage, stops a block or
// forbids enchanting, which protection would. Nothing here grants
// hexproof either, so a "may be targeted as though it didn't have
// hexproof" waiver does not lift it.

// CantBeTargetedByAuraSpells is "~ can't be the target of Aura spells"
// (Bartel Runeaxe, Tetsuo Umezawa). An Aura SPELL only: the Aura on the
// battlefield has no ability that targets, and an instant that happens
// to be white or artifact is untouched.
func CantBeTargetedByAuraSpells(label string) game.StaticAbility {
	return selfTargetingBar(game.SourceTargetingBar{
		Label:   label,
		Quality: game.ProtectionQuality{Kind: game.ProtectionQualitySubtype, Value: "Aura", Printed: "Aura"},
		Spells:  true,
	})
}

// CantBeTargetedByColor is "~ can't be the target of <colour> spells or
// abilities from <colour> sources" (Raiding Party's white). `color` is
// the one-letter wire colour.
func CantBeTargetedByColor(color, label string) game.StaticAbility {
	return selfTargetingBar(game.SourceTargetingBar{
		Label:     label,
		Quality:   game.ProtectionQuality{Kind: game.ProtectionQualityColor, Value: color, Printed: game.ColorName(color)},
		Spells:    true,
		Abilities: true,
	})
}

// CantBeTargetedBySpellsUnlessItAttackedOrBlocked is Lurker's "~ can't
// be the target of spells unless it attacked or blocked this turn". The
// condition is read live off the turn's combat records, so it lifts as
// the creature is declared and stays lifted after combat.
func CantBeTargetedBySpellsUnlessItAttackedOrBlocked(label string) game.StaticAbility {
	return selfTargetingBar(game.SourceTargetingBar{
		Label:                           label,
		Quality:                         game.ProtectionQuality{Kind: game.ProtectionQualityEverything, Printed: "everything"},
		Spells:                          true,
		UnlessAttackedOrBlockedThisTurn: true,
	})
}

// AttachedCantBeTargetedBySpells is "Enchanted creature can't be the
// target of spells" (Anti-Magic Aura). Abilities still target it, which
// is what the card says.
func AttachedCantBeTargetedBySpells(label string) game.StaticAbility {
	return attachedTargetingBar(game.SourceTargetingBar{
		Label:   label,
		Quality: game.ProtectionQuality{Kind: game.ProtectionQualityEverything, Printed: "everything"},
		Spells:  true,
	})
}

// AttachedCantBeTargetedByAbilitiesFrom is "Enchanted creature can't be
// the target of abilities from <type> sources" (Artifact Ward's
// artifact). `cardType` is the lowercase singular card type. Spells of
// that type still target it: the clause says abilities.
func AttachedCantBeTargetedByAbilitiesFrom(cardType, label string) game.StaticAbility {
	return attachedTargetingBar(game.SourceTargetingBar{
		Label:     label,
		Quality:   game.ProtectionQuality{Kind: game.ProtectionQualityCardType, Value: cardType, Printed: cardType},
		Abilities: true,
	})
}

// selfTargetingBar writes the bar onto the permanent whose ability it
// is. Every permanent type, not creatures only: Raiding Party is an
// enchantment barring spells aimed at itself.
func selfTargetingBar(bar game.SourceTargetingBar) game.StaticAbility {
	return game.StaticAbility{
		Label:     bar.Label,
		Layer:     game.Layer6Ability,
		AppliesTo: SelfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.TargetingBars = append(c.TargetingBars, bar)
		},
	}
}

// attachedTargetingBar writes the bar onto the permanent the Aura or
// Equipment is attached to, re-read on every pass, so moving the Aura
// moves the bar with it.
func attachedTargetingBar(bar game.SourceTargetingBar) game.StaticAbility {
	return game.StaticAbility{
		Label:     bar.Label,
		Layer:     game.Layer6Ability,
		AppliesTo: AttachedToSource,
		Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
			c.TargetingBars = append(c.TargetingBars, bar)
		},
	}
}

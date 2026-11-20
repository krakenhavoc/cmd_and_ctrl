package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// mrd_b5_helpers.go — statics and rows Mirrodin batch 5 shares (#5606).
// Append-only: add a builder, never change what one means.

// mrdB5PowerPerArtifact is "This creature gets +1/+0 for each artifact you
// control." (Nim Lasher, Nim Shrieker): Guidelight Synergist's layer 7c
// count, read through the layers on every recompute, so an artifact that
// arrives or leaves changes the creature's power at once and an animated
// artifact land counts.
func mrdB5PowerPerArtifact() game.StaticAbility {
	return game.StaticAbility{
		Label:     "This creature gets +1/+0 for each artifact you control.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			c.Power += artifactsControlledBy(g, source)
		},
	}
}

// mrdB5PerEquipmentOnThis is "This creature gets +n/+n for each Equipment
// attached to it." (Loxodon Punisher, Myr Adapter): a layer 7c count of
// the Equipment attached to the creature itself, whoever controls them
// (CR 301.5).
func mrdB5PerEquipmentOnThis(n int) game.StaticAbility {
	return game.StaticAbility{
		Label:     "This creature gets " + signedPT(n, n) + " for each Equipment attached to it.",
		Layer:     game.Layer7PT,
		SubLayer:  game.SubLayer7C_Modify,
		AppliesTo: selfOnly,
		Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
			k := n * s82AfrB08EquipmentAttachedTo(g, source.InstanceID)
			c.Power += k
			c.Toughness += k
		},
	}
}

// mrdB5WhileEquippedPlusOne is "As long as this creature is equipped, it
// gets +1/+1 and has <keyword>." (Leonin Den-Guard, Skyhunter Cub), Armed
// Assailant's shape: a layer 7c boost and a layer 6 grant on the creature
// itself, both read live, so they end the moment the last Equipment
// leaves it.
func mrdB5WhileEquippedPlusOne(keyword string) []game.StaticAbility {
	whileEquipped := func(target *game.Card, g *game.Game, source *game.Card) bool {
		return selfOnly(target, g, source) && isEquippedCreature(g, *target)
	}
	pump := b16Anthem(whileEquipped, 1, 1)
	pump.Label = "As long as this creature is equipped, it gets +1/+1."
	grant := KeywordGrant(whileEquipped, keyword)
	grant.Label = "As long as this creature is equipped, it has " + keyword + "."
	return []game.StaticAbility{pump, grant}
}

// mrdB5SpellbombDraw is the Spellbombs' second row, "{1}, Sacrifice this
// artifact: Draw a card." (Pyrite, Necrogen and Sunbeam Spellbomb).
func mrdB5SpellbombDraw() ActivatedAbility {
	return ActivatedAbility{
		Label:   "{1}, Sacrifice this artifact: Draw a card.",
		Purpose: game.Purpose{Answers: game.AnswerValue},
		Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
		Effect:  drawOneCardEffect,
	}
}

// mrdB5SacrificeAnArtifactFor is "Sacrifice an artifact: This creature
// gets +p/+t and gains <keyword> until end of turn." (Megatog, Krark-Clan
// Grunt): Atog's cost, with the boost and the grant both landing on the
// source as the ability resolves (thisCreatureUntilEOT), and nothing if it
// has left the battlefield.
func mrdB5SacrificeAnArtifactFor(name string, power, toughness int, keyword string) ActivatedAbility {
	pt := signedPT(power, toughness)
	return ActivatedAbility{
		Label: "Sacrifice an artifact: This creature gets " + pt + " and gains " + keyword + " until end of turn.",
		Purpose: game.Purpose{
			Answers: game.AnswerPump | game.AnswerCombatGrant,
			Pump:    &game.Pump{Power: power, Toughness: toughness},
		},
		Cost:   SacrificeN(1, "an artifact", Artifact()),
		Effect: thisCreatureUntilEOT(name+" — "+pt+" and "+keyword, power, toughness, keyword),
	}
}

// mrdB5TowerRow is a Tower's "{8}, {T}: <effect>" row.
func mrdB5TowerRow(label string, purpose game.Purpose, effect Effect) ActivatedAbility {
	return ActivatedAbility{
		Label:   "{8}, {T}: " + label,
		Purpose: purpose,
		Cost:    Plus(ManaCost("{8}"), TapCost()),
		Effect:  effect,
	}
}

// mrdB5SlithGrows is the Sliths' "Whenever this creature deals combat
// damage to a player, put a +1/+1 counter on it." Hexgold Slith's trigger:
// the counter goes on the creature only while it is still the object that
// dealt the damage (CR 400.7).
func mrdB5SlithGrows(name string) game.TriggeredAbility {
	return WheneverThisDealsCombatDamageToAPlayer(name+" — put a +1/+1 counter on it", thisStillHere(plusOneCountersOnThis(1)))
}

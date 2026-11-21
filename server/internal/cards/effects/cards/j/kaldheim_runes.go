package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// kaldheim_runes.go — the five Kaldheim Runes, one table because they
// are one card in five colours:
//
//	"Enchant permanent
//	 When this Aura enters, draw a card.
//	 As long as enchanted permanent is a creature, it [gets +P/+T and]
//	 has <keyword>.
//	 As long as enchanted permanent is an Equipment, it has "Equipped
//	 creature [gets +P/+T and] has <keyword>.""
//
// The creature half is an ordinary layer-6 keyword (and layer-7c pump)
// on the enchanted permanent, gated on its being a creature. The
// Equipment half is a GRANT (ADR 0093): the Equipment has the quoted
// static, so "equipped creature" is whatever the Equipment is attached
// to, and an Equipment that loses its abilities stops passing the
// keyword on while the Rune stays attached. The keyword half of that
// static is a granted layer-6 static, which the layer pass could not
// apply before #3784 (ADR 0093 amendment 2026-10-15); the +P/+T half is
// a granted layer-7c one (#2562).
//
// Both conditions read the enchanted permanent's types after layer 4, so
// an Equipment creature (a reconfigured or animated one) gets both
// halves, as printed. "Enchant permanent" keeps the Rune on a host that
// stops being a creature (CR 704.5m).
//
// No simplification.

type kaldheimRune struct {
	Name, OracleID, key string
	Keyword             string
	Power, Toughness    int
	text                string
}

var KaldheimRunes = []kaldheimRune{
	{Name: "Rune of Flight", OracleID: "829a4110-9209-45e9-b1a5-25e005711118", key: "rune-of-flight/equipped",
		Keyword: "flying", text: "Equipped creature has flying."},
	{Name: "Rune of Might", OracleID: "7826ce52-2fcc-422e-817e-6040de70f5f1", key: "rune-of-might/equipped",
		Keyword: "trample", Power: 1, Toughness: 1, text: "Equipped creature gets +1/+1 and has trample."},
	{Name: "Rune of Mortality", OracleID: "b827f8d2-07da-4358-9496-a68f8ccdbfe5", key: "rune-of-mortality/equipped",
		Keyword: "deathtouch", text: "Equipped creature has deathtouch."},
	{Name: "Rune of Speed", OracleID: "94d57b98-0f96-43c6-bf04-441339c89063", key: "rune-of-speed/equipped",
		Keyword: "haste", Power: 1, text: "Equipped creature gets +1/+0 and has haste."},
	{Name: "Rune of Sustenance", OracleID: "03ea5632-7d90-4390-8df7-686f63f6b1a3", key: "rune-of-sustenance/equipped",
		Keyword: "lifelink", text: "Equipped creature has lifelink."},
}

// enchantedCreature is "as long as enchanted permanent is a creature".
func enchantedCreature(target *game.Card, g *game.Game, source *game.Card) bool {
	return effects.AttachedToSource(target, g, source) && target.IsCreature()
}

// enchantedEquipment is "as long as enchanted permanent is an Equipment".
func enchantedEquipment(target *game.Card, g *game.Game, source *game.Card) bool {
	return effects.AttachedToSource(target, g, source) && target.HasSubtype("Equipment")
}

func init() {
	for _, r := range KaldheimRunes {
		var granted, own []game.StaticAbility
		if r.Power != 0 || r.Toughness != 0 {
			granted = append(granted, effects.PumpAttached(r.Power, r.Toughness))
			own = append(own, effects.B16Anthem(enchantedCreature, r.Power, r.Toughness))
		}
		granted = append(granted, effects.GrantToAttached(r.Keyword))
		own = append(own,
			effects.B16GrantKeywords(enchantedCreature, r.Keyword),
			effects.GrantAbilities(enchantedEquipment, r.key))
		effects.Register(effects.Spec{
			OracleID:     r.OracleID,
			Name:         r.Name,
			Completeness: effects.CompletenessFull,
			Targets:      effects.EnchantPermanent(),
			Purpose:      game.Purpose{Aura: game.AuraBenefit},
			Grants: []effects.AbilityGrant{{
				Key:    r.key,
				Static: granted,
				Text:   r.text,
			}},
			Static: own,
			Triggered: []game.TriggeredAbility{
				effects.WhenThisEnters(r.Name+" — draw a card", effects.Do(effects.DrawCards{N: 1})),
			},
		})
	}
}

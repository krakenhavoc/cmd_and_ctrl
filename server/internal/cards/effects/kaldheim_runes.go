package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	name, oracleID, key string
	keyword             string
	power, toughness    int
	text                string
}

var kaldheimRunes = []kaldheimRune{
	{name: "Rune of Flight", oracleID: "829a4110-9209-45e9-b1a5-25e005711118", key: "rune-of-flight/equipped",
		keyword: "flying", text: "Equipped creature has flying."},
	{name: "Rune of Might", oracleID: "7826ce52-2fcc-422e-817e-6040de70f5f1", key: "rune-of-might/equipped",
		keyword: "trample", power: 1, toughness: 1, text: "Equipped creature gets +1/+1 and has trample."},
	{name: "Rune of Mortality", oracleID: "b827f8d2-07da-4358-9496-a68f8ccdbfe5", key: "rune-of-mortality/equipped",
		keyword: "deathtouch", text: "Equipped creature has deathtouch."},
	{name: "Rune of Speed", oracleID: "94d57b98-0f96-43c6-bf04-441339c89063", key: "rune-of-speed/equipped",
		keyword: "haste", power: 1, text: "Equipped creature gets +1/+0 and has haste."},
	{name: "Rune of Sustenance", oracleID: "03ea5632-7d90-4390-8df7-686f63f6b1a3", key: "rune-of-sustenance/equipped",
		keyword: "lifelink", text: "Equipped creature has lifelink."},
}

// enchantedCreature is "as long as enchanted permanent is a creature".
func enchantedCreature(target *game.Card, g *game.Game, source *game.Card) bool {
	return AttachedToSource(target, g, source) && target.IsCreature()
}

// enchantedEquipment is "as long as enchanted permanent is an Equipment".
func enchantedEquipment(target *game.Card, g *game.Game, source *game.Card) bool {
	return AttachedToSource(target, g, source) && target.HasSubtype("Equipment")
}

func init() {
	for _, r := range kaldheimRunes {
		var granted, own []game.StaticAbility
		if r.power != 0 || r.toughness != 0 {
			granted = append(granted, PumpAttached(r.power, r.toughness))
			own = append(own, b16Anthem(enchantedCreature, r.power, r.toughness))
		}
		granted = append(granted, GrantToAttached(r.keyword))
		own = append(own,
			b16GrantKeywords(enchantedCreature, r.keyword),
			GrantAbilities(enchantedEquipment, r.key))
		Register(Spec{
			OracleID:     r.oracleID,
			Name:         r.name,
			Completeness: CompletenessFull,
			Targets:      EnchantPermanent(),
			Purpose:      game.Purpose{Aura: game.AuraBenefit},
			Grants: []AbilityGrant{{
				Key:    r.key,
				Static: granted,
				Text:   r.text,
			}},
			Static: own,
			Triggered: []game.TriggeredAbility{
				WhenThisEnters(r.name+" — draw a card", Do(DrawCards{N: 1})),
			},
		})
	}
}

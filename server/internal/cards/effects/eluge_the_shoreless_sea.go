package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eluge, the Shoreless Sea — Legendary Creature — Elemental Fish {1}{U}{U}{U},
// */* (Bloomburrow, #4467):
//
//	"Eluge's power and toughness are each equal to the number of Islands you
//	 control.
//	 Whenever Eluge enters or attacks, put a flood counter on target land.
//	 It's an Island in addition to its other types for as long as it has a
//	 flood counter on it.
//	 The first instant or sorcery spell you cast each turn costs {U} (or {1})
//	 less to cast for each land you control with a flood counter on it."
//
// The first is a layer 7a characteristic-defining ability reading the
// layered view (a flooded land is an Island by then). The second is
// Xolatoyac's trigger (FloodTargetLand). The third is a cost reduction
// whose unit is {U}: CR 118.7b and 118.7c are the engine's reduction rule,
// so a spell with {U} loses the blue first and then generic mana, and one
// with no {U} loses generic ("costs {U} (or {1}) less", the rulings). It
// applies to the first instant or sorcery spell its controller casts each
// turn (the cast tally is bumped after a spell is priced, so none so far
// means this is the first), whether or not any land is flooded.
func init() {
	reduction := CostsLessEach(elugeFloodedLands, "The first instant or sorcery spell you cast each turn costs {U} (or {1}) less to cast for each land you control with a flood counter on it.",
		YourSpell(), elugeFirstInstantOrSorcery)
	unit, err := game.ParseCost("{U}")
	if err != nil {
		panic(err)
	}
	reduction.Unit = &unit
	Register(Spec{
		OracleID:      "027989e7-20c4-4773-b6b4-20337cc5fe8c",
		Name:          "Eluge, the Shoreless Sea",
		Completeness:  CompletenessFull,
		CostModifiers: []game.CostModifier{reduction},
		Static: []game.StaticAbility{{
			Label:    "Power and toughness are each equal to the number of Islands you control.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7A_CDA,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				island := source.TextSubtype("Island")
				n := 0
				for i := range g.Battlefield.Cards {
					b := &g.Battlefield.Cards[i]
					if b.Controller == source.Controller && b.HasSubtype(island) {
						n++
					}
				}
				c.Power, c.Toughness = n, n
			},
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEntersOrAttacks("Eluge, the Shoreless Sea — put a flood counter on target land",
				func(g *game.Game, item *game.StackItem) error {
					return FloodTargetLand(NewContext(g, item), "Eluge, the Shoreless Sea")
				}), TargetPermanent("target land", Land())),
		},
	})
}

// elugeFloodedLands is the count of lands the modifier's source's
// controller controls with a flood counter on them.
func elugeFloodedLands(q game.CostQuery) int {
	if q.Game == nil {
		return 0
	}
	return countFloodedLands(q.Game, q.Source.Controller)
}

func countFloodedLands(g *game.Game, player uuid.UUID) int {
	n := 0
	if g.Battlefield == nil {
		return 0
	}
	for _, c := range g.Battlefield.Cards {
		if c.Controller == player && c.IsLand() && c.Counters["flood"] > 0 {
			n++
		}
	}
	return n
}

// elugeFirstInstantOrSorcery passes on an instant or sorcery spell when its
// caster has cast no other instant or sorcery spell this turn.
func elugeFirstInstantOrSorcery(q game.CostQuery) bool {
	return q.Game != nil && (q.Card.IsInstant() || q.Card.IsSorcery()) &&
		q.Game.CastTallyFor(q.Controller).InstantSorcery == 0
}

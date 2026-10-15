package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lyse Hext — Legendary Creature — Human Rebel Monk {1}{W}{U}, 2/2:
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)
//	 Noncreature spells you cast cost {1} less to cast.
//	 As long as you've cast two or more noncreature spells this turn,
//	 Lyse Hext has double strike."
//
// Prowess is the canonical keyword. The tax cut is Geist of Saint
// Thalia's. The double strike is Stoic Sphinx's shape: a layer-6 grant
// to the source gated on the per-turn cast tally (Noncreature, bumped
// before EventCast fires), flagged DependsOnSpellsCast so the layers are
// recomputed the moment the second spell is cast.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "47bce63e-8edb-43c1-b273-4b86da5c616b",
		Name:            "Lyse Hext",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{game.KeywordProwess},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Noncreature spells you cast cost {1} less to cast.",
				YourSpell(), NoncreatureSpell()),
		},
		Static: []game.StaticAbility{lyseHextDoubleStrike()},
	})
}

func lyseHextDoubleStrike() game.StaticAbility {
	ab := KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
		return target.InstanceID == source.InstanceID && g.CastTallyFor(source.Controller).Noncreature >= 2
	}, "double strike")
	ab.DependsOnSpellsCast = true
	return ab
}

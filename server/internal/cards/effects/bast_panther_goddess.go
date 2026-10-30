package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bast, Panther Goddess — {2}{G}{W} Legendary Creature — Cat God 4/4
// (Marvel Super Heroes Commander, #4455):
//
//	"Reach, trample, indestructible
//	 Bast can't attack or block unless you control three or more
//	 creatures.
//	 Whenever you attack, target attacking creature gets +X/+X until end
//	 of turn, where X is the number of creatures you control."
//
// The restriction is Hazoret, Godseeker's layer 6 CantAttack|CantBlock
// self static, on while its controller controls fewer than three
// creatures (Bast herself counts). The trigger is Raffine's: once per
// attack declaration, aimed at an attacking creature, with X counted as
// it resolves (CR 611.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e3233686-7703-460c-80ee-8cd37b3014ff",
		Name:            "Bast, Panther Goddess",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample", "indestructible"},
		Static: []game.StaticAbility{{
			Label: "Bast can't attack or block unless you control three or more creatures.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return selfOnly(target, g, source) && b04CreaturesControlled(g, source.Controller) < 3
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(Targeting(
				On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return attackDeclaredByYou(ev, source.Controller)
				}, "Bast — target attacking creature gets +X/+X until end of turn, where X is the number of creatures you control",
					bastPumpAttacker),
				TargetCreature("target attacking creature", AttackingCreature()))),
		},
	})
}

func bastPumpAttacker(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	x := b04CreaturesControlled(g, item.Controller)
	for _, t := range ctx.LegalTargets() {
		if t.Kind != game.TargetCard {
			continue
		}
		return BoostUntilEOT{Target: t.ID, Power: x, Toughness: x, Label: "Bast — +X/+X"}.Apply(ctx)
	}
	return nil
}

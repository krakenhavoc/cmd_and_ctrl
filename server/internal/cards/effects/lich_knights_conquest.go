package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lich-Knights' Conquest — Sorcery {4}{B} (EDHREC rank 2596):
//
//	"Sacrifice any number of artifacts, enchantments, and/or tokens.
//	 Return that many creature cards from your graveyard to the
//	 battlefield."
//
// Treasures and Clues into a mass reanimation. The two choices the
// card asks for — which permanents to sacrifice, which creature
// cards to return — are made in the opposite order from the printed
// one, because that is the pair of prompts the engine has: the
// creature cards are the spell's TARGET clause ("any number of
// target creature cards in your graveyard", picked at cast), and the
// sacrifices are that many sacrifice prompts at resolution, each the
// caster's own pick among their artifacts, enchantments and tokens
// (the b17PlayerSacrificesN shape, one permanent per prompt). The
// prompts are queued BEFORE the creatures return, so their option
// lists were computed without the returning cards — an artifact
// creature that just came back is never offered as the price of its
// own return. Fewer eligible permanents than chosen cards returns
// only as many as can be paid for; a card that left the graveyard
// in response is skipped (CR 608.2b).
//
// DECLARED SIMPLIFICATION, weaker than printed: the order of the
// choices. Choosing the creatures at cast means opponents see the
// picks and can respond to them, and — the printed card's best
// trick — a nontoken artifact or enchantment creature sacrificed to
// the spell cannot be one of the cards it returns, since it was not
// in the graveyard when the targets were chosen. Never stronger: the
// count is capped by what the caster can actually sacrifice.
func init() {
	Register(Spec{
		OracleID:     "b0e533dd-baf2-482a-b9e1-872eb5e47439",
		Name:         "Lich-Knights' Conquest",
		Completeness: CompletenessCaveats,
		Caveats: []string{
			"You pick the creature cards to return as you cast the spell and the permanents to sacrifice as it resolves — so a creature sacrificed to it can't be one of the cards returned, and opponents can respond to the picks.",
		},
		Targets: TargetCardInGraveyard("any number of target creature cards in your graveyard", YouOwn(), Creature()).WithCount(1, 0),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			chosen := 0
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					chosen++
				}
			}
			eligible := len(MatchingBattlefield(ctx, And(YouControl(), Or(Artifact(), Enchantment(), IsTokenPredicate()))))
			n := chosen
			if eligible < n {
				n = eligible
			}
			if n <= 0 {
				return nil
			}
			for i := 0; i < n; i++ {
				if ctx.Game.PlayerSacrificesForEffect(ctx.Source(), ctx.Controller(),
					sacrificeSpec("an artifact, enchantment, or token", Or(Artifact(), Enchantment(), IsTokenPredicate())),
					"Lich-Knights' Conquest — sacrifice an artifact, enchantment, or token") == 0 {
					n = i
					break
				}
			}
			return b24ReturnGraveyardTargetsToBattlefield(ctx, n)
		},
	})
}
